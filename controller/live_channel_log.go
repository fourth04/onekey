package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fourth04/onekey/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetLiveChannelLogByID(id string) (model.LiveChannelLog, error) {
	var liveChannelLog model.LiveChannelLog
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("operate_username = ? OR live_channel_name = ? OR live_channel_ip = ?", id, id, id).Preload("Areas.ParentArea").First(&liveChannelLog)
	} else {
		db = model.DB.Preload("Areas.ParentArea").First(&liveChannelLog, idNum)
	}
	if err := db.Error; err != nil {
		return liveChannelLog, err
	}
	return liveChannelLog, nil
}

func GetLiveChannelLogsByID(id string, areaID string, startTime string, endTime string) ([]model.LiveChannelLog, error) {
	var liveChannelLogs []model.LiveChannelLog
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("operate_username LIKE ? OR live_channel_name LIKE ? OR live_channel_ip LIKE ?", "%"+id+"%", "%"+id+"%", "%"+id+"%").Where("created_at >= ? AND created_at <= ?", startTime, endTime).Preload("Areas", "id = ?", areaID).Preload("Areas.ParentArea").Find(&liveChannelLogs)
	} else {
		db = model.DB.Where("created_at >= ? AND created_at <= ?", startTime, endTime).Preload("Areas", "id = ?", areaID).Preload("Areas.ParentArea").Find(&liveChannelLogs, idNum)
	}
	if err := db.Error; err != nil {
		return liveChannelLogs, err
	}
	if len(liveChannelLogs) == 0 {
		return liveChannelLogs, errors.New("record not found")
	}
	return liveChannelLogs, nil
}

func GetLiveChannelLogs(c *gin.Context) {
	// Get id liveChannelLog
	id := c.DefaultQuery("id", "")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	areaID := c.Query("area_id")

	// SELECT * FROM liveChannelLogs
	liveChannelLogs, err := GetLiveChannelLogsByID(id, areaID, startTime, endTime)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, liveChannelLogs)
	return

	// curl -i http://localhost:8080/api/v1/liveChannelLogs
}

func GetLiveChannelLog(c *gin.Context) {
	// Get id liveChannelLog
	id := c.Params.ByName("id")

	// SELECT * FROM liveChannelLog WHERE id = 1;
	liveChannelLog, err := GetLiveChannelLogByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, liveChannelLog)
	return

	// curl -i http://localhost:8080/api/v1/liveChannelLogs/1
}
