package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fourth04/onekey/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getResourceByID(id string) (model.Resource, error) {
	var resource model.Resource
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("resource_name = ?", id).First(&resource)
	} else {
		db = model.DB.First(&resource, idNum)
	}
	if err := db.Error; err != nil {
		return resource, err
	}
	return resource, nil
}

func getResourcesByID(id string) ([]model.Resource, error) {
	var resources []model.Resource
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("resource_name like ?", "%"+id+"%").Find(&resources)
	} else {
		db = model.DB.Find(&resources, idNum)
	}
	if err := db.Error; err != nil {
		return resources, err
	}
	if len(resources) == 0 {
		return resources, errors.New("record not found")
	}
	return resources, nil
}

func PostResource(c *gin.Context) {
	var resource model.Resource
	if err := c.ShouldBind(&resource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldResource model.Resource
	if !model.DB.Where("resource_name = ?", resource.ResourceName).First(&oldResource).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var permissionType model.PermissionType
	if model.DB.Find(&permissionType, resource.PermissionTypeID).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PermissionTypeID not found"})
		return
	}

	// INSERT INTO resources (name) VALUES (resource.Name);
	if err := model.DB.Create(&resource).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resource})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"resource_name\": \"/api/test\", \"resource_label\": \"测试\"}" http://localhost:8080/api/resources
}

func GetResources(c *gin.Context) {
	// Get id resource
	id := c.DefaultQuery("id", "")

	// SELECT * FROM resources
	resources, err := getResourcesByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, resources)
	return

	// curl -i http://localhost:8080/api/v1/resources

}

func GetResource(c *gin.Context) {
	// Get id resources
	id := c.Params.ByName("id")

	// SELECT * FROM resources WHERE id = 1;
	resource, err := getResourceByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
	return

	// curl -i http://localhost:8080/api/v1/resources/1
}

func UpdateResource(c *gin.Context) {
	// Get id resources
	id := c.Params.ByName("id")

	// SELECT * FROM resources WHERE id = 1;
	oldResource, err := getResourceByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newResource model.Resource
	if err := c.ShouldBind(&newResource); err != nil {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if newResource.PermissionTypeID != oldResource.PermissionTypeID {
		// Check if the associated record exists
		var permissionType model.PermissionType
		if model.DB.Find(&permissionType, newResource.PermissionTypeID).RecordNotFound() {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PermissionTypeID not found"})
			return
		}
	}

	newResource.ID = oldResource.ID
	newResource.CreatedAt = oldResource.CreatedAt
	if err := model.DB.Save(&newResource).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": newResource})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/resources/1
}

func DeleteResource(c *gin.Context) {
	// Get id resources
	id := c.Params.ByName("id")

	// SELECT * FROM resources WHERE id = 1;
	resource, err := getResourceByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM resources WHERE id = resource.ID
	if err := model.DB.Delete(&resource).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/resources/1
}
