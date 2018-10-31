package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fourth04/onekey/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetLiveChannelByID(id string) (model.LiveChannel, error) {
	var liveChannel model.LiveChannel
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("live_channel_name = ? OR outflow_ip = ?", id, id).Preload("Areas.ParentArea").First(&liveChannel)
	} else {
		db = model.DB.Preload("Areas.ParentArea").First(&liveChannel, idNum)
	}
	if err := db.Error; err != nil {
		return liveChannel, err
	}
	return liveChannel, nil
}

func GetLiveChannelsByID(id string) ([]model.LiveChannel, error) {
	var liveChannels []model.LiveChannel
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("live_channel_name like ? OR outflow_ip like ?", "%"+id+"%", "%"+id+"%").Preload("Areas.ParentArea").Find(&liveChannels)
	} else {
		db = model.DB.Preload("Areas.ParentArea").Find(&liveChannels, idNum)
	}
	if err := db.Error; err != nil {
		return liveChannels, err
	}
	if len(liveChannels) == 0 {
		return liveChannels, errors.New("record not found")
	}
	return liveChannels, nil
}

func PostLiveChannel(c *gin.Context) {
	var liveChannel model.LiveChannel
	if err := c.ShouldBind(&liveChannel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldLiveChannel model.LiveChannel
	if !model.DB.Where("live_channel_name = ?", liveChannel.LiveChannelName).First(&oldLiveChannel).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// INSERT INTO liveChannels (name) VALUES (liveChannel.Name);
	if err := model.DB.Create(&liveChannel).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": liveChannel})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"live_channel_name\": \"/api/test\", \"liveChannel_label\": \"测试\"}" http://localhost:8080/api/liveChannels
}

func GetLiveChannels(c *gin.Context) {
	// Get id liveChannel
	id := c.DefaultQuery("id", "")

	// SELECT * FROM liveChannels
	liveChannels, err := GetLiveChannelsByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, liveChannels)
	return

	// curl -i http://localhost:8080/api/v1/liveChannels
}

func GetLiveChannel(c *gin.Context) {
	// Get id liveChannel
	id := c.Params.ByName("id")

	// SELECT * FROM liveChannel WHERE id = 1;
	liveChannel, err := GetLiveChannelByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, liveChannel)
	return

	// curl -i http://localhost:8080/api/v1/liveChannels/1
}

func UpdateLiveChannel(c *gin.Context) {
	// Get id liveChannels
	id := c.Params.ByName("id")

	// SELECT * FROM liveChannels WHERE id = 1;
	oldLiveChannel, err := GetLiveChannelByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newLiveChannel model.LiveChannel
	if err := c.ShouldBind(&newLiveChannel); err != nil {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newLiveChannel.ID = oldLiveChannel.ID
	newLiveChannel.CreatedAt = oldLiveChannel.CreatedAt
	if err := model.DB.Save(&newLiveChannel).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": newLiveChannel})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/liveChannels/1
}

func DeleteLiveChannel(c *gin.Context) {
	// Get id liveChannels
	id := c.Params.ByName("id")

	// SELECT * FROM liveChannels WHERE id = 1;
	liveChannel, err := GetLiveChannelByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM liveChannels WHERE id = liveChannel.ID
	if err := model.DB.Delete(&liveChannel).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/liveChannels/1
}
