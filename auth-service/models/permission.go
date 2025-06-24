package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"` // 是否为系统角色
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Users       []User       `json:"users,omitempty" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// BeforeCreate 在创建前自动生成UUID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// RoleRequest 角色请求
type RoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// RoleUpdateRequest 角色更新请求
type RoleUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	Resource    string         `json:"resource" gorm:"not null"` // 资源类型
	Action      string         `json:"action" gorm:"not null"`   // 操作类型
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Users []User `json:"users,omitempty" gorm:"many2many:user_permissions;"`
}

// BeforeCreate 在创建前自动生成UUID
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PermissionGroup 权限组模型
type PermissionGroup struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:permission_group_permissions;"`
}

// BeforeCreate 在创建前自动生成UUID
func (pg *PermissionGroup) BeforeCreate(tx *gorm.DB) error {
	if pg.ID == "" {
		pg.ID = uuid.New().String()
	}
	return nil
}

// PermissionGroupRequest 权限组请求
type PermissionGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description"`
}

// PermissionGroupResponse 权限组响应
type PermissionGroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionRequest 权限请求
type PermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID    string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	RoleID    string    `json:"role_id" gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time `json:"created_at"`
}

// UserPermission 用户权限关联表
type UserPermission struct {
	UserID       string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	PermissionID string    `json:"permission_id" gorm:"primaryKey;type:uuid"`
	CreatedAt    time.Time `json:"created_at"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       string    `json:"role_id" gorm:"primaryKey;type:uuid"`
	PermissionID string    `json:"permission_id" gorm:"primaryKey;type:uuid"`
	CreatedAt    time.Time `json:"created_at"`
}

// PermissionGroupPermission 权限组权限关联表
type PermissionGroupPermission struct {
	PermissionGroupID string    `json:"permission_group_id" gorm:"primaryKey;type:uuid"`
	PermissionID      string    `json:"permission_id" gorm:"primaryKey;type:uuid"`
	CreatedAt         time.Time `json:"created_at"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// AssignPermissionRequest 分配权限请求
type AssignPermissionRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// UserPermissionResponse 用户权限响应
type UserPermissionResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	PermissionID string `json:"permission_id"`
	Permission   string `json:"permission"`
	Resource     string `json:"resource"`
	Action       string `json:"action"`
}

// RolePermissionResponse 角色权限响应
type RolePermissionResponse struct {
	RoleID       string `json:"role_id"`
	RoleName     string `json:"role_name"`
	PermissionID string `json:"permission_id"`
	Permission   string `json:"permission"`
	Resource     string `json:"resource"`
	Action       string `json:"action"`
} 