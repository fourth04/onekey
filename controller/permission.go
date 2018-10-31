package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fourth04/onekey/model"
	"github.com/fourth04/onekey/utils"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetPermissionByPermissionID(id string) (map[string]interface{}, error) {
	var permission model.Permission
	if err := model.DB.Preload("PermissionType").Preload("Operation").First(&permission, id).Error; err != nil {
		return nil, err
	}

	permissionMap := map[string]interface{}{
		"permission_type_name":  permission.PermissionType.PermissionTypeName,
		"permission_type_label": permission.PermissionType.PermissionTypeLabel,
		"operation_name":        permission.Operation.OperationName,
		"operation_label":       permission.Operation.OperationLabel,
		"created_at":            permission.CreatedAt,
		"updated_at":            permission.UpdatedAt,
		"id":                    permission.ID,
	}
	switch permission.PermissionType.PermissionTypeName {
	default:
		var resource model.Resource
		if err := model.DB.First(&resource, permission.ResourceID).Error; err != nil {
			return nil, err
		}
		permissionMap["resource_label"] = resource.ResourceLabel
		permissionMap["resource_name"] = resource.ResourceName
	}
	return permissionMap, nil
}

func GetPermissionsByRoleID(id string) (map[string]interface{}, error) {
	var role model.Role
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("role_name = ?", id).Preload("Permissions.PermissionType").Preload("Permissions.Operation").Find(&role)
	} else {
		db = model.DB.Preload("Permissions.PermissionType").Preload("Permissions.Operation").Find(&role, idNum)
	}
	if err := db.Error; err != nil {
		return nil, err
	}

	permissionMaps := make([]map[string]interface{}, len(role.Permissions))
	for ix, permission := range role.Permissions {
		tmpMap := map[string]interface{}{
			"permission_type_name":  permission.PermissionType.PermissionTypeName,
			"permission_type_label": permission.PermissionType.PermissionTypeLabel,
			"operation_name":        permission.Operation.OperationName,
			"operation_label":       permission.Operation.OperationLabel,
			"created_at":            permission.CreatedAt,
			"updated_at":            permission.UpdatedAt,
			"id":                    permission.ID,
		}
		permissionMaps[ix] = tmpMap
		switch permission.PermissionType.PermissionTypeName {
		default:
			var resource model.Resource
			if err := model.DB.First(&resource, permission.ResourceID).Error; err != nil {
				return nil, err
			}
			permissionMaps[ix]["resource_label"] = resource.ResourceLabel
			permissionMaps[ix]["resource_name"] = resource.ResourceName
		}
	}

	roleMap, err := utils.StructToMapByJson(role)
	if err != nil {
		return nil, err
	}
	roleMap["permissions"] = permissionMaps
	return roleMap, nil
}

func GetPermissionsByUserID(id string) (map[string]interface{}, error) {
	var user model.User
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Preload("Roles").Where("username = ?", id).First(&user)
	} else {
		db = model.DB.Preload("Roles").First(&user, idNum)
	}
	if err := db.Error; err != nil {
		return nil, err
	}

	roleMaps := make([]map[string]interface{}, len(user.Roles))
	for ix, role := range user.Roles {
		roleMap, err := GetPermissionsByRoleID(strconv.Itoa(int(role.ID)))
		if err != nil {
			return nil, err
		}
		roleMaps[ix] = roleMap
	}

	userMap, err := utils.StructToMapByJson(user)
	if err != nil {
		// Display JSON error
		return nil, err
	}
	userMap["roles"] = roleMaps
	delete(userMap, "password")
	delete(userMap, "salt")
	return userMap, nil
}

func GetUserClipedByUserID(id string) (map[string]interface{}, error) {
	userMap, err := GetPermissionsByUserID(id)
	if err != nil {
		return nil, err
	}

	roles := userMap["roles"].([]map[string]interface{})
	newRoles := make([]map[string]interface{}, len(roles))
	for ix, role := range roles {
		var permissionsFiltered []map[string]interface{}
		permissions := role["permissions"].([]map[string]interface{})
		for _, permission := range permissions {
			delete(permission, "id")
			delete(permission, "created_at")
			delete(permission, "updated_at")
			permissionTypeName := permission["permission_type_name"].(string)
			if permissionTypeName == "resource" {
				permissionsFiltered = append(permissionsFiltered, permission)
			}
		}
		role["permissions"] = permissionsFiltered
		delete(role, "id")
		delete(role, "created_at")
		delete(role, "updated_at")
		delete(role, "users")
		newRoles[ix] = role
	}
	userMap["roles"] = newRoles

	return userMap, nil
}

type PermissionLite struct {
	PermissionTypeID uint `gorm:"NOT NULL" binding:"required" form:"permission_type_id" json:"permission_type_id"`
	OperationID      uint `gorm:"NOT NULL" binding:"required" form:"operation_id" json:"operation_id"`
	ResourceID       uint `gorm:"NOT NULL" binding:"required" form:"resource_id" json:"resource_id"`
}

func PostPermission(c *gin.Context) {
	var permissionLite PermissionLite
	if err := c.ShouldBind(&permissionLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldPermission model.Permission
	if !model.DB.Where("permission_type_id = ? AND operation_id = ? AND resource_id = ?", permissionLite.PermissionTypeID, permissionLite.OperationID, permissionLite.ResourceID).First(&oldPermission).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var permissionType model.PermissionType
	var operation model.Operation
	if model.DB.First(&permissionType, permissionLite.PermissionTypeID).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "permission_type_id not found"})
		return
	}
	if model.DB.First(&operation, permissionLite.OperationID).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation_id not found"})
		return
	}
	switch permissionType.PermissionTypeName {
	default:
		var resource model.Resource
		if err := model.DB.First(&resource, permissionLite.ResourceID).Error; err != nil {
			// Display error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "operation_id not found"})
			return
		}
	}

	newPermission := model.Permission{
		PermissionTypeID: permissionLite.PermissionTypeID,
		PermissionType:   permissionType,
		OperationID:      permissionLite.OperationID,
		Operation:        operation,
		ResourceID:       permissionLite.ResourceID,
	}

	if err := model.DB.Set("gorm:save_associations", false).Create(&newPermission).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": newPermission})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"permission_name\": \"/api/test\", \"permission_label\": \"测试\"}" http://localhost:8080/api/permissions
}

func GetPermissions(c *gin.Context) {
	// SELECT * FROM permissions
	var permissions []model.Permission
	if model.DB.Find(&permissions).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record not found"})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, permissions)
	return

	// curl -i http://localhost:8080/api/v1/permissions
}

func GetPermission(c *gin.Context) {
	// GET id permission
	id := c.Params.ByName("id")

	// SELECT * FROM permissions WHERE id = 1;
	permissionMap, err := GetPermissionByPermissionID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, permissionMap)
	return

	// curl -i http://localhost:8080/api/v1/permissions/1
}

func UpdatePermission(c *gin.Context) {
	// Get id permission
	id := c.Params.ByName("id")

	// SELECT * FROM permissions WHERE id = 1;
	var oldPermission model.Permission
	if model.DB.First(&oldPermission, id).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record not found"})
		return
	}

	var permissionLite PermissionLite
	if err := c.ShouldBind(&permissionLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the associated record exists
	var permissionType model.PermissionType
	var operation model.Operation
	if model.DB.First(&permissionType, permissionLite.PermissionTypeID).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "permission_type_id not found"})
		return
	}
	if model.DB.First(&operation, permissionLite.OperationID).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation_id not found"})
		return
	}
	switch permissionType.PermissionTypeName {
	default:
		var resource model.Resource
		if err := model.DB.First(&resource, permissionLite.ResourceID).Error; err != nil {
			// Display error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "operation_id not found"})
			return
		}
	}

	oldPermission.PermissionType = permissionType
	oldPermission.PermissionTypeID = permissionLite.PermissionTypeID
	oldPermission.OperationID = permissionLite.OperationID
	oldPermission.ResourceID = permissionLite.ResourceID
	oldPermission.UpdatedAt = time.Now()

	if err := model.DB.Set("gorm:save_associations", false).Save(&oldPermission).Error; err != nil {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": oldPermission})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/permissions/1
}

func DeletePermission(c *gin.Context) {
	// Get id permission
	id := c.Params.ByName("id")

	// SELECT * FROM permissions WHERE id = 1;
	var permission model.Permission
	if model.DB.First(&permission, id).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record not found"})
		return
	}

	// DELETE FROM permissions WHERE id = permission.ID
	if err := model.DB.Delete(&permission).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/permissions/1
}
