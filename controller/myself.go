package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/fourth04/onekey/model"
	"github.com/fourth04/onekey/utils"
	"github.com/gin-gonic/gin"
)

func GetMyself(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id := int(claims["id"].(float64))

	// SELECT * FROM users WHERE id = 1;
	userMap, err := GetPermissionsByUserID(strconv.Itoa(id))
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userMap)
	return

	// curl -i http://localhost:8080/api/v1/users/1
}

type PasswordInfo struct {
	OldPassword string `binding:"required"  form:"old_password" json:"old_password"`
	NewPassword string `binding:"required"  form:"new_password" json:"new_password"`
}

func PostMyself(c *gin.Context) {
	// Get id users
	claims := jwt.ExtractClaims(c)
	fmt.Println(claims)
	id := int(claims["id"].(float64))

	// SELECT * FROM users WHERE id = 1;
	var oldUser model.User
	if model.DB.First(&oldUser, id).RecordNotFound() {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record not found"})
		return
	}

	var passwordInfo PasswordInfo
	if err := c.ShouldBind(&passwordInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if passwordInfo.OldPassword == passwordInfo.NewPassword {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Old Password is the same as new password"})
		return
	}
	oldPasswordEncrypted, err := utils.Encrypt(passwordInfo.OldPassword, oldUser.Salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if oldPasswordEncrypted != oldUser.Password {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Old Password wrong"})
		return
	}

	newSalt := utils.RandomString(10)
	newPasswordEncrypted, err := utils.Encrypt(passwordInfo.NewPassword, newSalt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	oldUser.Password = newPasswordEncrypted
	oldUser.Salt = newSalt
	oldUser.UpdatedAt = time.Now()

	if err := model.DB.Set("gorm:save_associations", false).Save(&oldUser).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Display modified data in JSON message "success"
	c.JSON(http.StatusOK, gin.H{"success": oldUser.Desentitize()})
	return

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/users/1
}
