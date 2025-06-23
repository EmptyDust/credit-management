package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
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

	// 关联关系
	UserFiles     []UserFile     `json:"user_files,omitempty" gorm:"foreignKey:UserID"`
	Permissions   []Permission   `json:"permissions,omitempty" gorm:"many2many:user_permissions;"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
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

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	ID           uint       `json:"id"`
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

// LoginResponse 登录响应
type LoginResponse struct {
	Token   string       `json:"token"`
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// Permission 权限模型
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	Resource    string         `json:"resource" gorm:"not null"` // 资源类型
	Action      string         `json:"action" gorm:"not null"`   // 操作类型
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Users []User `json:"users,omitempty" gorm:"many2many:user_permissions;"`
}

// UserFile 用户文件模型
type UserFile struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	FileName      string         `json:"file_name" gorm:"not null"`
	OriginalName  string         `json:"original_name" gorm:"not null"`
	FilePath      string         `json:"file_path" gorm:"not null"`
	FileSize      int64          `json:"file_size"`
	FileType      string         `json:"file_type"`
	MimeType      string         `json:"mime_type"`
	Category      string         `json:"category"` // 文件分类：avatar, document, certificate, etc.
	Description   string         `json:"description"`
	IsPublic      bool           `json:"is_public" gorm:"default:false"`
	DownloadCount int            `json:"download_count" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	User User `json:"user,omitempty"`
}

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// FileResponse 文件响应
type FileResponse struct {
	ID            uint      `json:"id"`
	FileName      string    `json:"file_name"`
	OriginalName  string    `json:"original_name"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `json:"file_type"`
	MimeType      string    `json:"mime_type"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	IsPublic      bool      `json:"is_public"`
	DownloadCount int       `json:"download_count"`
	DownloadURL   string    `json:"download_url"`
	PreviewURL    string    `json:"preview_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Notification 通知模型
type Notification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"` // info, success, warning, error
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	User User `json:"user,omitempty"`
}

// NotificationRequest 通知请求
type NotificationRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=info success warning error"`
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
