package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型（认证服务专用）
type User struct {
	UserID       string         `json:"user_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null"`
	Password     string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Email        string         `json:"email" gorm:"uniqueIndex"`
	Phone        string         `json:"phone"`
	RealName     string         `json:"real_name"`
	UserType     string         `json:"user_type" gorm:"not null"`      // student, teacher, admin
	Status       string         `json:"status" gorm:"default:'active'"` // active, inactive, suspended
	Avatar       string         `json:"avatar"`                         // 头像文件路径
	LastLoginAt  *time.Time     `json:"last_login_at"`
	RegisterTime time.Time      `json:"register_time" gorm:"autoCreateTime"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:user_permissions;"`
}

// BeforeCreate 在创建前自动生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserID == "" {
		u.UserID = uuid.New().String()
	}
	return nil
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	RealName     string     `json:"real_name"`
	UserType     string     `json:"user_type"`
	Status       string     `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	RegisterTime time.Time  `json:"register_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token   string       `json:"token"`
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// TokenValidationRequest Token验证请求
type TokenValidationRequest struct {
	Token string `json:"token" binding:"required"`
}

// TokenValidationResponse Token验证响应
type TokenValidationResponse struct {
	Valid   bool         `json:"valid"`
	User    UserResponse `json:"user,omitempty"`
	Message string       `json:"message"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Message      string `json:"message"`
}
