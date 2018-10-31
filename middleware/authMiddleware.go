package middleware

import (
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/fourth04/onekey/config"
	"github.com/fourth04/onekey/controller"
	"github.com/fourth04/onekey/model"
	"github.com/fourth04/onekey/utils"
	"github.com/gin-gonic/gin"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var realm = config.Cfg["jwt_realm"].(string)
var key = config.Cfg["jwt_key"].(string)
var timeout, _ = time.ParseDuration(config.Cfg["jwt_timeout"].(string))
var maxRefresh, _ = time.ParseDuration(config.Cfg["jwt_max_refresh"].(string))
var identityKey = "id"

func unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func authenticatorByID(c *gin.Context) (interface{}, error) {
	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	username := loginVals.Username
	password := loginVals.Password

	var user model.User
	if model.DB.Where("username = ?", username).First(&user).RecordNotFound() {
		return nil, jwt.ErrFailedAuthentication
	}
	passwordEncrypted, err := utils.Encrypt(password, user.Salt)
	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	if passwordEncrypted != user.Password {
		return nil, jwt.ErrFailedAuthentication
	}
	return user.Desentitize(), nil
}

func payloadFuncByID(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		return jwt.MapClaims{
			identityKey: v["id"].(int),
		}
	}
	return jwt.MapClaims{}
}

var GinJWTMiddlewareByID = jwt.GinJWTMiddleware{
	Realm:         realm,
	Key:           []byte(key),
	Timeout:       timeout,
	MaxRefresh:    maxRefresh,
	IdentityKey:   identityKey,
	PayloadFunc:   payloadFuncByID,
	Authenticator: authenticatorByID,
	Unauthorized:  unauthorized,
	TokenLookup:   "header:Authorization",
	TokenHeadName: "Bearer",
	TimeFunc:      time.Now,
}

func authenticatorByUser(c *gin.Context) (interface{}, error) {
	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	username := loginVals.Username
	password := loginVals.Password

	var user model.User
	if model.DB.Where("username = ?", username).First(&user).RecordNotFound() {
		return nil, jwt.ErrFailedAuthentication
	}
	passwordEncrypted, err := utils.Encrypt(password, user.Salt)
	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	if passwordEncrypted != user.Password {
		return nil, jwt.ErrFailedAuthentication
	}

	userMap, err := controller.GetUserClipedByUserID(username)
	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}

	// return user.Desentitize(), nil
	// return user.Clip(), nil
	return userMap, nil

}

func payloadFuncByUser(data interface{}) jwt.MapClaims {
	rv := jwt.MapClaims{}
	value := data.(map[string]interface{})
	for k, v := range value {
		rv[k] = v
	}
	return rv
}

var GinJWTMiddlewareByUser = jwt.GinJWTMiddleware{
	Realm:         realm,
	Key:           []byte(key),
	Timeout:       timeout,
	MaxRefresh:    maxRefresh,
	IdentityKey:   identityKey,
	PayloadFunc:   payloadFuncByUser,
	Authenticator: authenticatorByUser,
	Unauthorized:  unauthorized,
	TokenLookup:   "header:Authorization",
	TokenHeadName: "Bearer",
	TimeFunc:      time.Now,
}

func identityHandlerByID(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	id := claims[identityKey]
	userMap, err := controller.GetPermissionsByUserID(strconv.Itoa(id.(int)))
	if err != nil {
		return map[string]interface{}{}
	}
	return userMap
}

func authorizatorByID(data interface{}, c *gin.Context) bool {
	user := data.(map[string]interface{})
	if tmpRoles, ok := user["roles"]; ok {
		roles := tmpRoles.([]interface{})
		for _, tmpRole := range roles {
			role := tmpRole.(string)
			if role == "admin" {
				return true
			}
		}
	}
	return false
}

var GinJWTMiddlewareAdminByID = jwt.GinJWTMiddleware{
	Realm:           realm,
	Key:             []byte(key),
	Timeout:         timeout,
	MaxRefresh:      maxRefresh,
	IdentityKey:     identityKey,
	PayloadFunc:     payloadFuncByID,
	Authenticator:   authenticatorByID,
	IdentityHandler: identityHandlerByID,
	Authorizator:    authorizatorByID,
	Unauthorized:    unauthorized,
	TokenLookup:     "header:Authorization",
	TokenHeadName:   "Bearer",
	TimeFunc:        time.Now,
}

func identityHandlerByUser(c *gin.Context) interface{} {
	user := jwt.ExtractClaims(c)
	return user
}

func authorizatorAdminByUser(data interface{}, c *gin.Context) bool {
	userMap := data.(jwt.MapClaims)
	roles := userMap["roles"].([]interface{})
	for _, role := range roles {
		roleMap := role.(map[string]interface{})
		if roleName := roleMap["role_name"].(string); roleName == "admin" {
			return true
		}
	}
	return false
}

func authorizatorDeleteByUser(data interface{}, c *gin.Context) bool {
	userMap := data.(jwt.MapClaims)
	roles := userMap["roles"].([]interface{})
	for _, role := range roles {
		roleMap := role.(map[string]interface{})
		permissions := roleMap["permissions"].([]map[string]interface{})
		for _, permission := range permissions {
			if permission["operation_name"] == "del" {
				return true
			}
		}
	}
	return false
}

var GinJWTMiddlewareAdminByUser = jwt.GinJWTMiddleware{
	Realm:           realm,
	Key:             []byte(key),
	Timeout:         timeout,
	MaxRefresh:      maxRefresh,
	IdentityKey:     identityKey,
	PayloadFunc:     payloadFuncByUser,
	Authenticator:   authenticatorByUser,
	IdentityHandler: identityHandlerByUser,
	Authorizator:    authorizatorAdminByUser,
	Unauthorized:    unauthorized,
	TokenLookup:     "header:Authorization",
	TokenHeadName:   "Bearer",
	TimeFunc:        time.Now,
}

var AuthUserMiddlewareByID, _ = jwt.New(&GinJWTMiddlewareByID)
var AuthAdminMiddlewareByID, _ = jwt.New(&GinJWTMiddlewareAdminByID)
var AuthUserMiddlewareByUser, _ = jwt.New(&GinJWTMiddlewareByUser)
var AuthAdminMiddlewareByUser, _ = jwt.New(&GinJWTMiddlewareAdminByUser)
