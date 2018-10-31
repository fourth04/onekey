package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fourth04/onekey/model"
	"github.com/fourth04/onekey/utils"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getUserByID(id string) (model.User, error) {
	var user model.User
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Where("user_name = ?", id).First(&user)
	} else {
		db = model.DB.First(&user, idNum)
	}
	if err := db.Error; err != nil {
		return user, err
	}
	return user, nil
}

func getUsersByID(id string) ([]model.User, error) {
	var users []model.User
	var db *gorm.DB
	idNum, err := strconv.Atoi(id)
	if err != nil {
		db = model.DB.Preload("Roles").Where("username like ?", "%"+id+"%").Find(&users)
	} else {
		db = model.DB.Find(&users, idNum)
	}
	if err := db.Error; err != nil {
		return users, err
	}
	if len(users) == 0 {
		return users, errors.New("record not found")
	}
	return users, nil
}

type UserLite struct {
	Username      string `gorm:"TYPE:VARCHAR(100);UNIQUE_INDEX;NOT NULL" binding:"required" form:"username" json:"username"`
	Password      string `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"password" json:"password"`
	RateFormatted string `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"rate_formatted" json:"rate_formatted"`
	RoleIDs       []uint `form:"role_ids" json:"role_ids"`
}

func PostUser(c *gin.Context) {
	var userLite UserLite
	if err := c.ShouldBind(&userLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// PostUser UserLite.Password can not be empty
	if userLite.Password == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Key: 'UserLite.Password' Error:Field validation for 'Password' failed on the 'required' tag"})
		return
	}

	// Check if the record exists
	var oldUser model.User
	if !model.DB.Where("username = ?", userLite.Username).First(&oldUser).RecordNotFound() {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Record exists"})
		return
	}

	// Check if the associated record exists
	var roles []*model.Role
	if err := model.DB.Find(&roles, userLite.RoleIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range userLite.RoleIDs {
		flagContained := false
		for _, role := range roles {
			if role.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("role_id #%d not found", id)})
			return
		}
	}

	salt := utils.RandomString(10)
	password, err := utils.Encrypt(userLite.Password, salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newUser := model.User{
		Username:      userLite.Username,
		Salt:          salt,
		Password:      password,
		Roles:         roles,
		RateFormatted: userLite.RateFormatted,
	}

	if err := model.DB.Set("gorm:save_associations", false).Create(&newUser).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": newUser.Desentitize()})
	return

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"user_name\": \"/api/test\", \"user_label\": \"测试\"}" http://localhost:8080/api/users
}

func GetUsers(c *gin.Context) {
	// Get id user
	id := c.DefaultQuery("id", "")

	// SELECT * FROM users
	users, err := getUsersByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Desentitize
	usersDesensitized := make([]interface{}, len(users))
	for i, user := range users {
		usersDesensitized[i] = user.Desentitize()
	}

	// Display JSON result
	c.JSON(http.StatusOK, usersDesensitized)
	return

	// curl -i http://localhost:8080/api/v1/users
}

func GetUser(c *gin.Context) {
	// Get id users
	id := c.Params.ByName("id")

	// SELECT * FROM users WHERE id = 1;
	userMap, err := GetPermissionsByUserID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userMap)
	return

	// curl -i http://localhost:8080/api/v1/users/1
}

func UpdateUser(c *gin.Context) {
	// Get id users
	id := c.Params.ByName("id")

	// SELECT * FROM users WHERE id = 1;
	oldUser, err := getUserByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userLite UserLite
	if err := c.ShouldBind(&userLite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the associated record exists
	var roles []*model.Role
	if err := model.DB.Find(&roles, userLite.RoleIDs).Error; err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, id := range userLite.RoleIDs {
		flagContained := false
		for _, role := range roles {
			if role.ID == id {
				flagContained = true
				break
			}
		}
		if !flagContained {
			// Display JSON error
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("role_id #%d not found", id)})
			return
		}
	}

	var password, salt string
	// UpdateUser UserLite.Password can be empty, it means don't change the password
	if userLite.Password == "" {
		password = oldUser.Password
		salt = oldUser.Salt
	} else {
		salt = utils.RandomString(10)
		passwordEncrypted, err := utils.Encrypt(userLite.Password, salt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		password = passwordEncrypted
	}

	oldUser.Username = userLite.Username
	oldUser.Password = password
	oldUser.Salt = salt
	oldUser.RateFormatted = userLite.RateFormatted
	oldUser.Roles = roles
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

func DeleteUser(c *gin.Context) {
	// Get id users
	id := c.Params.ByName("id")

	// SELECT * FROM users WHERE id = 1;
	user, err := getUserByID(id)
	if err != nil {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DELETE the associatied records
	if err := model.DB.Model(&user).Association("Roles").Clear().Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// DELETE FROM users WHERE id = user.ID
	if err := model.DB.Delete(&user).Error; err != nil {
		// Display error
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": "Record #" + id + " deleted"})
	return

	// curl -i -X DELETE http://localhost:8080/api/v1/users/1
}
