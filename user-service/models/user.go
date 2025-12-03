package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 统一用户模型
type User struct {
	// 基础用户信息
	UUID         string         `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:gen_random_uuid()"`
	StudentID    *string        `json:"student_id,omitempty" gorm:"column:student_id;uniqueIndex"`
	TeacherID    *string        `json:"teacher_id,omitempty" gorm:"column:teacher_id;uniqueIndex"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null;size:20"`
	Password     string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Email        string         `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Phone        *string        `json:"phone" gorm:"uniqueIndex;size:11"` // 可为空
	RealName     string         `json:"real_name" gorm:"not null;size:50"`
	UserType     string         `json:"user_type" gorm:"not null;size:20"` // student, teacher, admin
	Status       string         `json:"status" gorm:"not null;default:active;size:20"`
	Avatar       *string        `json:"avatar"`                                                              // 头像文件路径，可为空
	DepartmentID *string        `json:"department_id,omitempty" gorm:"type:uuid;references:departments(id)"` // 关联部门表
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 学生特有字段（可选）
	Grade *string `json:"grade,omitempty" gorm:"size:4"`

	// 教师特有字段（可选）
	Title *string `json:"title,omitempty" gorm:"size:50"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}

// UserRequest 用户注册/创建请求
type UserRequest struct {
	StudentID    string `json:"student_id" binding:"omitempty,len=8,numeric"`
	TeacherID    string `json:"teacher_id" binding:"omitempty,min=1,max=18"`
	Username     string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password     string `json:"password" binding:"required,min=8"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"omitempty,len=11,startswith=1"`
	RealName     string `json:"real_name" binding:"required,min=2,max=50"`
	UserType     string `json:"user_type" binding:"required,oneof=student teacher"`
	DepartmentID string `json:"department_id" binding:"omitempty,uuid"` // 部门ID

	// 前端为了展示方便，会传递学部/专业/班级名称（字符串），后端据此反查 department_id
	// 教师/学生创建时与更新保持一致的字段命名
	Department string `json:"department" binding:"omitempty,max=100"`
	College    string `json:"college" binding:"omitempty,max=100"`
	Major      string `json:"major" binding:"omitempty,max=100"`
	Class      string `json:"class" binding:"omitempty,max=50"`

	// 学生特有字段
	Grade string `json:"grade" binding:"omitempty,len=4,numeric"`

	// 教师特有字段
	Title string `json:"title" binding:"omitempty,max=50"`
}

// StudentRegisterRequest 学生注册请求
type StudentRegisterRequest struct {
	StudentID    string `json:"student_id" binding:"required,len=8,numeric"`
	Username     string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password     string `json:"password" binding:"required,min=8"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required,len=11,startswith=1"`
	RealName     string `json:"real_name" binding:"required,min=2,max=50"`
	DepartmentID string `json:"department_id" binding:"omitempty"` // 班级ID（可选，优先于major/class）

	// 自助注册时，前端通过学部/专业/班级名称传递信息，后端负责反查 department_id
	Department string `json:"department" binding:"omitempty,max=100"`
	College    string `json:"college" binding:"omitempty,max=100"`
	Major      string `json:"major" binding:"omitempty,max=100"`
	Class      string `json:"class" binding:"omitempty,max=50"`

	Grade string `json:"grade" binding:"required,len=4,numeric"`
}

// TeacherRegisterRequest 教师注册请求
type TeacherRegisterRequest struct {
	TeacherID    string `json:"teacher_id" binding:"required,min=1,max=18"`
	Username     string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password     string `json:"password" binding:"required,min=8"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required,len=11,startswith=1"`
	RealName     string `json:"real_name" binding:"required,min=2,max=50"`
	DepartmentID string `json:"department_id" binding:"omitempty,uuid"` // 部门ID（可选，优先于department/college）

	// 前端可以通过学部名称传递部门信息，后端负责反查 department_id
	Department string `json:"department" binding:"omitempty,max=100"`
	College    string `json:"college" binding:"omitempty,max=100"`

	Title string `json:"title" binding:"required,max=50"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Email        string  `json:"email" binding:"omitempty,email"`
	Phone        string  `json:"phone" binding:"omitempty,len=11,startswith=1"`
	RealName     string  `json:"real_name" binding:"omitempty,min=2,max=50"`
	UserType     string  `json:"user_type" binding:"omitempty,oneof=student teacher admin"`
	Status       string  `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Avatar       string  `json:"avatar" binding:"omitempty"`
	DepartmentID string  `json:"department_id" binding:"omitempty,uuid"`
	StudentID    *string `json:"student_id" binding:"omitempty,len=8,numeric"`
	TeacherID    *string `json:"teacher_id" binding:"omitempty,min=1,max=18"`

	// 前端为了展示方便，会传递学部/专业/班级名称（字符串），后端据此反查 department_id
	// 教师更新时主要使用 Department/College 表示学部
	Department string `json:"department" binding:"omitempty,max=100"`
	College string `json:"college" binding:"omitempty,max=100"`
	Major   string `json:"major" binding:"omitempty,max=100"`
	Class   string `json:"class" binding:"omitempty,max=50"`

	// 学生特有字段
	Grade *string `json:"grade" binding:"omitempty,len=4,numeric"`

	// 教师特有字段
	Title *string `json:"title" binding:"omitempty,max=50"`
}

// UserResponse 用户响应
type UserResponse struct {
	UUID         string     `json:"uuid"`
	StudentID    *string    `json:"student_id,omitempty"`
	TeacherID    *string    `json:"teacher_id,omitempty"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	RealName     string     `json:"real_name"`
	UserType     string     `json:"user_type"`
	Status       string     `json:"status"`
	Avatar       string     `json:"avatar"`
	DepartmentID string     `json:"department_id,omitempty"` // 部门ID
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 学生特有字段
	Grade *string `json:"grade,omitempty"`

	// 教师特有字段
	Title *string `json:"title,omitempty"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query        string `json:"query" form:"query"`
	UserType     string `json:"user_type" form:"user_type"`
	DepartmentID string `json:"department_id" form:"department_id"` // 部门ID搜索
	Grade        string `json:"grade" form:"grade"`
	Title        string `json:"title" form:"title"`
	Status       string `json:"status" form:"status"`
	Page         int    `json:"page" form:"page"`
	PageSize     int    `json:"page_size" form:"page_size"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// 自定义验证函数
func ValidatePasswordComplexity(password string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	return len(password) >= 8 && hasUpper && hasLower && hasDigit
}

func ValidatePhoneFormat(phone string) bool {
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	if phone[1] < '3' || phone[1] > '9' {
		return false
	}
	for i := 2; i < 11; i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}
	return true
}

func ValidateGradeFormat(grade string) bool {
	if len(grade) != 4 {
		return false
	}
	for _, char := range grade {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
