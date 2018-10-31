package model

import (
	"log"

	"github.com/fourth04/onekey/config"
	"github.com/fourth04/onekey/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var DB *gorm.DB
var err error

func init() {
	dialect := config.Cfg["dialect"].(string)
	dbPath := config.Cfg["db_path"].(string)

	log.Println("正在初始化数据库连接...")
	DB, err = gorm.Open(dialect, dbPath)
	// Error
	if err != nil {
		utils.ErrHandleFatalln(err, "初始化数据库连接失败！")
	}
	log.Println("初始化数据库连接完成！")

	// Display SQL queries
	DB.LogMode(config.Cfg["db_log_mode"].(bool))

	// Creating the table
	DB.AutoMigrate(&User{}, &Role{}, &Permission{}, &PermissionType{}, &Operation{}, &Resource{}, &LiveChannel{}, &LiveChannelLog{}, &Area{})

	// DB.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&User{})
}
