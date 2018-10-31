package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fourth04/onekey/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getOperationByID(id string) (model.Operation, error) {
	var operation model.Operation
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("operation_name = ?", id).First(&operation)
	} else {
		db = model.DB.First(&operation, idNum)
	}
	if err := db.Error; err != nil {
		return operation, err
	}
	return operation, nil
}

func getOperationsByID(id string) ([]model.Operation, error) {
	var operations []model.Operation
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("operation_name like ?", "%"+id+"%").Find(&operations)
	} else {
		db = model.DB.Find(&operations, idNum)
	}
	if err := db.Error; err != nil {
		return operations, err
	}
	if len(operations) == 0 {
		return operations, errors.New("record not found")
	}
	return operations, nil
}

func PostOperation(c *gin.Context) {
	var operation model.Operation
	if err := c.ShouldBind(&operation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldOperation model.Operation
	if !model.DB.Where("operation_name = ?", operation.OperationName).First(&oldOperation).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var permissionType model.PermissionType
	if model.DB.Find(&permissionType, operation.PermissionTypeID).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PermissionTypeID not found"})
		return
	}

	// INSERT INTO operations (name) VALUES (operation.Name);
	if err := model.DB.Create(&operation).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": operation})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"operation_name\": \"/api/test\", \"operation_label\": \"测试\"}" http://localhost:8080/api/operations
}

func GetOperations(c *gin.Context) {
	// Get id operation
	id := c.DefaultQuery("id", "")

	// SELECT * FROM operations
	operations, err := getOperationsByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, operations)
	return

	// curl -i http://localhost:8080/api/v1/operations
}

func GetOperation(c *gin.Context) {
	// Get id operations
	id := c.Params.ByName("id")

	// SELECT * FROM operations WHERE id = 1;
	operation, err := getOperationByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, operation)
	return

	// curl -i http://localhost:8080/api/v1/operations/1
}

func UpdateOperation(c *gin.Context) {
	// Get id operations
	id := c.Params.ByName("id")

	// SELECT * FROM operations WHERE id = 1;
	oldOperation, err := getOperationByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newOperation model.Operation
	if err := c.ShouldBind(&newOperation); err != nil {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if newOperation.PermissionTypeID != oldOperation.PermissionTypeID {
		// Check if the associated record exists
		var permissionType model.PermissionType
		if model.DB.Find(&permissionType, newOperation.PermissionTypeID).RecordNotFound() {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PermissionTypeID not found"})
			return
		}
	}

	newOperation.ID = oldOperation.ID
	newOperation.CreatedAt = oldOperation.CreatedAt
	if err := model.DB.Save(&newOperation).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": newOperation})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/operations/1
}

func DeleteOperation(c *gin.Context) {
	// Get id operations
	id := c.Params.ByName("id")

	// SELECT * FROM operations WHERE id = 1;
	operation, err := getOperationByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM operations WHERE id = operation.ID
	if err := model.DB.Delete(&operation).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/operations/1
}
