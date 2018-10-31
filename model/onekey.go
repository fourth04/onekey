package model

import (
	"time"
)

type LiveChannel struct {
	ID                uint      `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	DeviceName        string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"device_name" json:"device_name"`
	LiveChannelName   string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"live_channel_name" json:"live_channel_name"`
	PrimaryInflowIP   string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"primary_inflow_ip" json:"primary_inflow_ip"`
	PrimaryInflowPort string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"primary_inflow_port" json:"primary_inflow_port"`
	StandbyInflowIP   string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"standby_inflow_ip" json:"standby_inflow_ip"`
	StandbyInflowPort string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"standby_inflow_port" json:"standby_inflow_port"`
	OutflowIP         string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"outflow_ip" json:"outflow_ip"`
	OutflowPort       string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"outflow_port" json:"outflow_port"`
	Areas             []*Area   `gorm:"many2many:area_live_channels;" form:"areas" json:"areas"`
	CreatedAt         time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type LiveChannelLog struct {
	ID              uint      `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	OperateUsername string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"operate_username" json:"operate_username"`
	OperateTime     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"operate_time" json:"operate_time"`
	OperateType     string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"operate_type" json:"operate_type"`
	OperateStatus   string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"operate_status" json:"operate_status"`
	LiveChannelName string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"live_channel_name" json:"live_channel_name"`
	LiveChannelIP   string    `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"live_channel_ip" json:"live_channel_ip"`
	Areas           []*Area   `gorm:"many2many:area_live_channel_logs;" form:"areas" json:"areas"`
	CreatedAt       time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type Area struct {
	ID               uint              `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	AreaName         string            `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"area_name" json:"area_name"`
	AreaLabel        string            `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"area_label" json:"area_label"`
	AreaAbbreviation string            `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"area_abbreviation" json:"area_abbreviation"`
	ParentID         uint              `form:"parent_id" json:"parent_id"`
	ParentArea       *Area             `gorm:"foreignkey:ID;association_foreignkey:ParentID" form:"parent_area" json:"parent_area"`
	LiveChannels     []*LiveChannel    `gorm:"many2many:area_live_channels;" form:"live_channels" json:"live_channels"`
	LiveChannelLogs  []*LiveChannelLog `gorm:"many2many:area_live_channel_logs;" form:"live_channel_logs" json:"live_channel_logs"`
	CreatedAt        time.Time         `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}
