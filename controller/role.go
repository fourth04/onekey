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

func getRoleByID(id string) (model.Role, error) {
	var role model.Role
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("role_name = ?", id).First(&role)
	} else {
		db = model.DB.First(&role, idNum)
	}
	if err := db.Error; err != nil {
		return role, err
	}
	return role, nil
}

func getRolesByID(id string) ([]model.Role, error) {
	var roles []model.Role
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("role_name like ?", "%"+id+"%").Find(&roles)
	} else {
		db = model.DB.Find(&roles, idNum)
	}
	if err := db.Error; err != nil {
		return roles, err
	}
	if len(roles) == 0 {
		return roles, errors.New("record not found")
	}
	return roles, nil
}

type RoleLite struct {
	RoleName      string `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"role_name" json:"role_name"`
	RoleLabel     string `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"role_label" json:"role_label"`
	PermissionIDs []uint `form:"permission_ids" json:"permission_ids"`
}

func PostRole(c *gin.Context) {
	var roleLite RoleLite
	if err := c.ShouldBind(&roleLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the record exists
	var oldRole model.Role
	if !model.DB.Where("role_name = ?", roleLite.RoleName).First(&oldRole).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var permissions []*model.Permission
	if err := model.DB.Find(&permissions, roleLite.PermissionIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range roleLite.PermissionIDs {
		flagContained := false
		for _, permission := range permissions {
			if permission.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("permission_id #%d not found", id)})
			return
		}
	}

	newRole := model.Role{
		RoleName:    roleLite.RoleName,
		RoleLabel:   roleLite.RoleLabel,
		Permissions: permissions,
	}

	if err := model.DB.Set("gorm:save_associations", false).Create(&newRole).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": newRole})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"role_name\": \"/api/test\", \"role_label\": \"测试\"}" http://localhost:8080/api/roles
}

func GetRoles(c *gin.Context) {
	// Get id role
	id := c.DefaultQuery("id", "")

	// SELECT * FROM roles
	roles, err := getRolesByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, roles)
	return

	// curl -i http://localhost:8080/api/v1/roles
}

func GetRole(c *gin.Context) {
	// Get id roles
	id := c.Params.ByName("id")

	// SELECT * FROM roles WHERE id = 1;
	roleMap, err := GetPermissionsByRoleID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roleMap)
	return

	// curl -i http://localhost:8080/api/v1/roles/1
}

func UpdateRole(c *gin.Context) {
	// Get id roles
	id := c.Params.ByName("id")

	// SELECT * FROM roles WHERE id = 1;
	oldRole, err := getRoleByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newRole RoleLite
	if err := c.ShouldBind(&newRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the associated record exists
	var permissions []*model.Permission
	if err := model.DB.Find(&permissions, newRole.PermissionIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range newRole.PermissionIDs {
		flagContained := false
		for _, permission := range permissions {
			if permission.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("permission_id #%d not found", id)})
			return
		}
	}
	oldRole.Permissions = permissions

	oldRole.RoleName = newRole.RoleName
	oldRole.RoleName = newRole.RoleLabel
	oldRole.UpdatedAt = time.Now()
	if err := model.DB.Save(&oldRole).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": oldRole})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/roles/1
}

func DeleteRole(c *gin.Context) {
	// Get id roles
	id := c.Params.ByName("id")

	// SELECT * FROM roles WHERE id = 1;
	role, err := getRoleByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE the associatied records
	if err := model.DB.Model(&role).Association("Permissions").Clear().Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM roles WHERE id = role.ID
	if err := model.DB.Delete(&role).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/roles/1
}
