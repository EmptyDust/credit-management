package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型（用户管理服务专用）
type User struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null"`
	Password     string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Email        string         `json:"email" gorm:"uniqueIndex"`
	Phone        string         `json:"phone"`
	RealName     string         `json:"real_name"`
	UserType     string         `json:"user_type" gorm:"not null"`      // student, teacher, admin
	Role         string         `json:"role" gorm:"default:'user'"`     // user, moderator, admin
	Status       string         `json:"status" gorm:"default:'active'"` // active, inactive, suspended
	Avatar       string         `json:"avatar"`                         // 头像文件路径
	LastLoginAt  *time.Time     `json:"last_login_at"`
	RegisterTime time.Time      `json:"register_time" gorm:"autoCreateTime"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserRequest 用户注册请求
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name" binding:"required"`
	UserType string `json:"user_type" binding:"required,oneof=student teacher admin"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	UserType string `json:"user_type"`
	Status   string `json:"status"`
	Role     string `json:"role"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	RealName     string     `json:"real_name"`
	UserType     string     `json:"user_type"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	Avatar       string     `json:"avatar"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	RegisterTime time.Time  `json:"register_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserStats 用户统计信息
type UserStats struct {
	TotalUsers     int64 `json:"total_users"`
	ActiveUsers    int64 `json:"active_users"`
	SuspendedUsers int64 `json:"suspended_users"`
	StudentUsers   int64 `json:"student_users"`
	TeacherUsers   int64 `json:"teacher_users"`
	AdminUsers     int64 `json:"admin_users"`
	NewUsersToday  int64 `json:"new_users_today"`
	NewUsersWeek   int64 `json:"new_users_week"`
	NewUsersMonth  int64 `json:"new_users_month"`
}
