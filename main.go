package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fourth04/onekey/config"
	"github.com/fourth04/onekey/controller"
	"github.com/fourth04/onekey/middleware"
	"github.com/fourth04/onekey/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
)

func initLog() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   config.Cfg["running_log_filepath"].(string),
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})
	log.SetPrefix("[INFO]")
}

func init() {
	initLog()
}

func main() {
	defer model.DB.Close()

	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(&lumberjack.Logger{
		Filename:   config.Cfg["httpserver_log_filepath"].(string),
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"withCredentials", "Authorization", "Origin", "Content-Type", "Accept", "*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	router.Static("assets", config.Cfg["static_dir"].(string))
	router.Static("js", "dist/js")

	router.LoadHTMLFiles(config.Cfg["index_filepath"].(string))
	router.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, filepath.Base(config.Cfg["index_filepath"].(string)), gin.H{})
	})

	// rateLimiterMiddleware := middleware.RateLimiterMiddlewareByUser
	authUserMiddleware := middleware.AuthUserMiddlewareByUser
	authAdminMiddleware := middleware.AuthAdminMiddlewareByUser

	router.POST("api/auth/login", authUserMiddleware.LoginHandler)
	auth := router.Group("api/auth")
	auth.Use(authAdminMiddleware.MiddlewareFunc())
	{
		auth.GET("/refresh_token", authUserMiddleware.RefreshHandler)
		// auth.GET("/refresh_token", rateLimiterMiddleware.Middleware(), authUserMiddleware.RefreshHandler)
	}

	user := router.Group("api/users")
	user.Use(authAdminMiddleware.MiddlewareFunc())
	{
		user.POST("/", controller.PostUser)
		user.GET("/", controller.GetUsers)
		user.GET("/:id", controller.GetUser)
		user.PUT("/:id", controller.UpdateUser)
		user.DELETE("/:id", controller.DeleteUser)
	}

	myself := router.Group("api/myself")
	myself.Use(authUserMiddleware.MiddlewareFunc())
	{
		myself.GET("/", controller.GetMyself)
		myself.POST("/", controller.PostMyself)
	}

	role := router.Group("api/roles")
	role.Use(authAdminMiddleware.MiddlewareFunc())
	{
		role.POST("/", controller.PostRole)
		role.GET("/", controller.GetRoles)
		role.GET("/:id", controller.GetRole)
		role.PUT("/:id", controller.UpdateRole)
		role.DELETE("/:id", controller.DeleteRole)
	}

	permissionType := router.Group("api/permission_types")
	permissionType.Use(authAdminMiddleware.MiddlewareFunc())
	{
		permissionType.POST("/", controller.PostPermissionType)
		permissionType.GET("/", controller.GetPermissionTypes)
		permissionType.GET("/:id", controller.GetPermissionType)
		permissionType.PUT("/:id", controller.UpdatePermissionType)
		permissionType.DELETE("/:id", controller.DeletePermissionType)
	}

	operation := router.Group("api/operations")
	operation.Use(authAdminMiddleware.MiddlewareFunc())
	{
		operation.POST("/", controller.PostOperation)
		operation.GET("/", controller.GetOperations)
		operation.GET("/:id", controller.GetOperation)
		operation.PUT("/:id", controller.UpdateOperation)
		operation.DELETE("/:id", controller.DeleteOperation)
	}

	resource := router.Group("api/resources")
	resource.Use(authAdminMiddleware.MiddlewareFunc())
	{
		resource.POST("/", controller.PostResource)
		resource.GET("/", controller.GetResources)
		resource.GET("/:id", controller.GetResource)
		resource.PUT("/:id", controller.UpdateResource)
		resource.DELETE("/:id", controller.DeleteResource)
	}

	permission := router.Group("api/permissions")
	permission.Use(authAdminMiddleware.MiddlewareFunc())
	{
		permission.POST("/", controller.PostPermission)
		permission.GET("/", controller.GetPermissions)
		permission.GET("/:id", controller.GetPermission)
		permission.PUT("/:id", controller.UpdatePermission)
		permission.DELETE("/:id", controller.DeletePermission)
	}

	livechannel := router.Group("api/live_channels")
	livechannel.Use(authAdminMiddleware.MiddlewareFunc())
	{
		livechannel.POST("/", controller.PostLiveChannel)
		livechannel.PUT("/:id", controller.UpdateLiveChannel)
		livechannel.DELETE("/:id", controller.DeleteLiveChannel)
	}

	router.GET("api/live_channels/", authUserMiddleware.MiddlewareFunc(), controller.GetLiveChannels)
	router.GET("api/live_channels/:id", authUserMiddleware.MiddlewareFunc(), controller.GetLiveChannel)

	router.POST("api/live_channel_block_off/", authAdminMiddleware.MiddlewareFunc(), controller.PostLiveChannelBlockOff)
	router.GET("api/live_channel_logs/", authUserMiddleware.MiddlewareFunc(), controller.GetLiveChannelLogs)

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for s := range c {
			switch s {
			case os.Interrupt, os.Kill:
				log.Println("退出进程", s)
				os.Exit(0)
			default:
				log.Println("other", s)
			}
		}
	}()

	router.Run(config.Cfg["uri"].(string))
}
