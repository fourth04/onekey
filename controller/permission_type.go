package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fourth04/onekey/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getPermissionTypeByID(id string) (model.PermissionType, error) {
	var permissionType model.PermissionType
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("permission_type_name = ?", id).First(&permissionType)
	} else {
		db = model.DB.First(&permissionType, idNum)
	}
	if err := db.Error; err != nil {
		return permissionType, err
	}
	return permissionType, nil
}

func getPermissionTypesByID(id string) ([]model.PermissionType, error) {
	var permissionTypes []model.PermissionType
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("permission_type_name like ?", "%"+id+"%").Find(&permissionTypes)
	} else {
		db = model.DB.Find(&permissionTypes, idNum)
	}
	if err := db.Error; err != nil {
		return permissionTypes, err
	}
	if len(permissionTypes) == 0 {
		return permissionTypes, errors.New("record not found")
	}
	return permissionTypes, nil
}

type PermissionTypeLite struct {
	PermissionTypeName  string `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"permission_type_name" json:"permission_type_name"`
	PermissionTypeLabel string `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"permission_type_label" json:"permission_type_label"`
	OperationIDs        []uint `form:"operation_ids" json:"operation_ids"`
	ResourceIDs         []uint `form:"resource_ids" json:"resource_ids"`
}

func PostPermissionType(c *gin.Context) {
	var permissionTypeLite PermissionTypeLite
	if err := c.ShouldBind(&permissionTypeLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldPermissionType model.PermissionType
	if !model.DB.Where("permission_type_name = ?", permissionTypeLite.PermissionTypeName).First(&oldPermissionType).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var operations []model.Operation
	if err := model.DB.Find(&operations, permissionTypeLite.OperationIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range permissionTypeLite.OperationIDs {
		flagContained := false
		for _, operation := range operations {
			if operation.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("operation_id #%d not found", id)})
			return
		}
	}

	newPermissionType := model.PermissionType{
		PermissionTypeName:  permissionTypeLite.PermissionTypeName,
		PermissionTypeLabel: permissionTypeLite.PermissionTypeLabel,
		Operations:          operations,
	}

	switch {
	default:
		var resources []model.Resource
		if err := model.DB.Find(&resources, permissionTypeLite.ResourceIDs).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, id := range permissionTypeLite.ResourceIDs {
			flagContained := false
			for _, resource := range resources {
				if resource.ID == id {
					flagContained = true
					break
				}
			}
			if !flagContained {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("resource_id #%d not found", id)})
				return
			}
		}
		newPermissionType.Resources = resources
	}

	// INSERT INTO permission_types (name) VALUES (oldPermissionType.Name);
	if err := model.DB.Set("gorm:save_associations", false).Create(&newPermissionType).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": newPermissionType})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"oldPermissionType_name\": \"/api/test\", \"oldPermissionType_label\": \"测试\"}" http://localhost:8080/api/permission_types
}

func GetPermissionTypes(c *gin.Context) {
	// Get id permission_type
	id := c.DefaultQuery("id", "")

	// SELECT * FROM permission_types
	permissionTypes, err := getPermissionTypesByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, permissionTypes)
	return

	// curl -i http://localhost:8080/api/v1/permission_type
}

func GetPermissionType(c *gin.Context) {
	// Get id permission_types
	id := c.Params.ByName("id")

	// SELECT * FROM permission_types WHERE id = 1;
	permissionType, err := getPermissionTypeByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch permissionType.PermissionTypeName {
	default:
		if err := model.DB.Preload("Operations").Preload("Resources").First(&permissionType).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, permissionType)
	return

	// curl -i http://localhost:8080/api/v1/permission_types/1
}

func UpdatePermissionType(c *gin.Context) {
	// Get id permission_types
	id := c.Params.ByName("id")

	// SELECT * FROM permission_types WHERE id = 1;
	oldPermissionType, err := getPermissionTypeByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newPermissionType PermissionTypeLite
	if err := c.ShouldBind(&newPermissionType); err != nil {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the associated record exists
	var operations []model.Operation
	if err := model.DB.Find(&operations, newPermissionType.OperationIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range newPermissionType.OperationIDs {
		flagContained := false
		for _, operation := range operations {
			if operation.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("operation_id #%d not found", id)})
			return
		}
	}
	if err := model.DB.Model(&oldPermissionType).Association("Operations").Replace(operations).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	oldPermissionType.Operations = operations
	switch {
	default:
		var resources []model.Resource
		if err := model.DB.Find(&resources, newPermissionType.ResourceIDs).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, id := range newPermissionType.ResourceIDs {
			flagContained := false
			for _, resource := range resources {
				if resource.ID == id {
					flagContained = true
					break
				}
			}
			if !flagContained {
				// Display JSON error
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("resource_id #%d not found", id)})
				return
			}
		}
		if err := model.DB.Model(&oldPermissionType).Association("Resources").Replace(resources).Error; err != nil {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		oldPermissionType.Resources = resources
	}

	oldPermissionType.PermissionTypeName = newPermissionType.PermissionTypeName
	oldPermissionType.PermissionTypeLabel = newPermissionType.PermissionTypeLabel
	oldPermissionType.UpdatedAt = time.Now()
	if err := model.DB.Set("gorm:save_associations", false).Save(&oldPermissionType).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": newPermissionType})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/permission_types/1
}

func DeletePermissionType(c *gin.Context) {
	// Get id permission_types
	id := c.Params.ByName("id")

	// SELECT * FROM permission_types WHERE id = 1;
	permissionType, err := getPermissionTypeByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE the associatied records
	if err := model.DB.Model(&permissionType).Association("Operations").Clear().Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	if err := model.DB.Model(&permissionType).Association("Resources").Clear().Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM permissionTypes WHERE id = permissionType.ID
	if err := model.DB.Delete(&permissionType).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/permission_types/1
}
