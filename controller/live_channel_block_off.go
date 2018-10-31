package controller

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/fourth04/onekey/config"
	"github.com/fourth04/onekey/model"
	"github.com/fourth04/onekey/utils"
	"github.com/gin-gonic/gin"
)

type IDs struct {
	ID          string `binding:"required" form:"id" json:"id"`
	OperateType string `binding:"required" form:"operate_type" json:"operate_type"`
	AreaIDs     []int  `binding:"required" form:"area_ids" json:"area_ids"`
}

func PostLiveChannelBlockOff(c *gin.Context) {
	var ids IDs
	if err := c.ShouldBind(&ids); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := ids.ID
	operateType := ids.OperateType
	areaIDs := ids.AreaIDs

	if (operateType != "block_off") && (operateType != "deblock") {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operate_type: " + operateType + " 未定义，请修改操作类型"})
		return
	}

	// SELECT * FROM liveChannels WHERE id = 1;
	liveChannel, err := GetLiveChannelByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the associated record exists
	var areas []*model.Area
	if err := model.DB.Find(&areas, areaIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range areaIDs {
		flagContained := false
		for _, area := range areas {
			if int(area.ID) == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("area_id #%d not found", id)})
			return
		}
	}

	var operateTypeLabel string
	switch operateType {
	case "block_off":
		// do the block_off
		operateTypeLabel = "封堵"

		if runtime.GOOS != "windows" {
			blockOffScriptFilepath := config.Cfg["block_off_script_filepath"].(string)
			command := blockOffScriptFilepath + " " + liveChannel.OutflowIP
			_, err := utils.ExecuteAndGetResultCombineError(command)
			if err != nil {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			/* if !strings.Contains(commandResult, "OK") {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Execute Return：" + commandResult})
				return
			} */
		}

		// replace blocked off areas
		if err := model.DB.Model(&liveChannel).Association("Areas").Replace(areas).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "deblock":
		// do deblock
		operateTypeLabel = "解封"

		if runtime.GOOS != "windows" {
			deblockScriptFilepath := config.Cfg["deblock_script_filepath"].(string)
			command := deblockScriptFilepath + " " + liveChannel.OutflowIP
			_, err := utils.ExecuteAndGetResultCombineError(command)
			if err != nil {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			/* if !strings.Contains(commandResult, "OK") {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Execute Return：" + commandResult})
				return
			} */
		}

		// replace blocked off areas
		var oldAreas []*model.Area
		var newAreas []*model.Area
		if err := model.DB.Model(&liveChannel).Related(&oldAreas, "Areas").Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, oldArea := range oldAreas {
			flagContained := false
			for _, area := range areas {
				if area.ID == oldArea.ID {
					flagContained = true
					break
				}
			}
			if !flagContained {
				newAreas = append(newAreas, oldArea)
			}
		}
		if err := model.DB.Model(&liveChannel).Association("Areas").Replace(newAreas).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	default:
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operateType: " + operateType + " 未定义，请修改操作类型"})
		return
	}

	// 插入操作日志
	claims := jwt.ExtractClaims(c)
	username := claims["username"].(string)
	liveChannelLog := model.LiveChannelLog{
		OperateUsername: username,
		OperateTime:     time.Now(),
		OperateType:     operateTypeLabel,
		OperateStatus:   "成功",
		LiveChannelName: liveChannel.LiveChannelName,
		LiveChannelIP:   liveChannel.OutflowIP,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := model.DB.Create(&liveChannelLog).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Ints(areaIDs)
	var areaIDsStr string
	for _, value := range areaIDs {
		areaIDsStr += strconv.Itoa(value)
	}
	if areaIDsStr == "23456789" {
		var tmpArea model.Area
		if err := model.DB.First(&tmpArea, 1).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = model.DB.Model(&liveChannelLog).Association("Areas").Append(tmpArea).Error
	} else {
		err = model.DB.Model(&liveChannelLog).Association("Areas").Append(areas).Error
	}
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": operateTypeLabel + "成功"})
}
