package model

import (
	"encoding/json"
	"time"
)

type UserDesentitized struct {
	ID            uint      `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	Username      string    `gorm:"TYPE:VARCHAR(100);UNIQUE_INDEX;NOT NULL" form:"username" json:"username"`
	RateFormatted string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"rate_formatted" json:"rate_formatted"`
	Roles         []*Role   `gorm:"many2many:user_roles" form:"roles" json:"roles"`
	CreatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type User struct {
	ID            uint      `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	Username      string    `gorm:"TYPE:VARCHAR(100);UNIQUE_INDEX;NOT NULL" form:"username" json:"username"`
	Password      string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"password" json:"password"`
	Salt          string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"salt" json:"salt"`
	RateFormatted string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"rate_formatted" json:"rate_formatted"`
	Roles         []*Role   `gorm:"many2many:user_roles" form:"roles" json:"roles"`
	CreatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

func (u User) Desentitize() map[string]interface{} {
	js, _ := json.Marshal(u)
	var userDesensitized map[string]interface{}
	json.Unmarshal(js, &userDesensitized)

	delete(userDesensitized, "password")
	delete(userDesensitized, "salt")
	return userDesensitized
}

func (u User) Clip() map[string]interface{} {
	var roleNames []string
	for _, role := range u.Roles {
		roleNames = append(roleNames, role.RoleName)
	}
	userCliped := map[string]interface{}{
		"id":             u.ID,
		"rate_formatted": u.RateFormatted,
		"username":       u.Username,
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
		"roles":          roleNames,
	}
	return userCliped
}

type Role struct {
	ID          uint          `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	RoleName    string        `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"role_name" json:"role_name"`
	RoleLabel   string        `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"role_label" json:"role_label"`
	Users       []*User       `gorm:"many2many:user_roles" form:"users" json:"users"`
	Permissions []*Permission `gorm:"many2many:role_permissions" form:"permissions" json:"permissions"`
	CreatedAt   time.Time     `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type PermissionType struct {
	ID                  uint        `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	PermissionTypeName  string      `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"permission_type_name" json:"permission_type_name"`
	PermissionTypeLabel string      `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"permission_type_label" json:"permission_type_label"`
	Operations          []Operation `form:"operations" json:"operations"`
	Resources           []Resource  `form:"resources" json:"resources"`
	CreatedAt           time.Time   `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt           time.Time   `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type Operation struct {
	ID               uint           `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	OperationName    string         `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"operation_name" json:"operation_name"`
	OperationLabel   string         `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"operation_label" json:"operation_label"`
	PermissionTypeID uint           `form:"permission_type_id" json:"permission_type_id"`
	PermissionType   PermissionType `form:"permission_type" json:"permission_type"`
	CreatedAt        time.Time      `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type Resource struct {
	ID               uint           `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	ResourceName     string         `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"resource_name" json:"resource_name"`
	ResourceLabel    string         `gorm:"TYPE:VARCHAR(100);NOT NULL" binding:"required" form:"resource_label" json:"resource_label"`
	PermissionTypeID uint           `form:"permission_type_id" json:"permission_type_id"`
	PermissionType   PermissionType `form:"permission_type" json:"permission_type"`
	CreatedAt        time.Time      `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}

type Permission struct {
	ID               uint           `gorm:"PRIMARY_KEY;AUTO_INCREMENT" form:"id" json:"id"`
	PermissionTypeID uint           `gorm:"UNIQUE_INDEX:idx_permission;NOT NULL" binding:"required" form:"permission_type_id" json:"permission_type_id"`
	PermissionType   PermissionType `form:"permission_type" json:"permission_type"`

	OperationID uint      `gorm:"UNIQUE_INDEX:idx_permission;NOT NULL" binding:"required" form:"operation_id" json:"operation_id"`
	Operation   Operation `form:"operation" json:"operation"`
	ResourceID  uint      `gorm:"UNIQUE_INDEX:idx_permission;NOT NULL" binding:"required" form:"resource_id" json:"resource_id"`
	Roles       []*Role   `gorm:"many2many:role_permissions" form:"roles" json:"roles"`
	CreatedAt   time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP" form:"updated_at" json:"updated_at"`
}
