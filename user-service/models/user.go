package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 统一用户模型（合并用户、学生、教师信息）
type User struct {
	// 基础用户信息
	UserID       string         `json:"user_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null"`
	Password     string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	Phone        string         `json:"phone"`
	RealName     string         `json:"real_name" gorm:"not null"`
	UserType     string         `json:"user_type" gorm:"not null"` // student, teacher, admin
	Status       string         `json:"status" gorm:"not null;default:active"`
	Avatar       string         `json:"avatar"` // 头像文件路径
	LastLoginAt  *time.Time     `json:"last_login_at"`
	RegisterTime time.Time      `json:"register_time" gorm:"autoCreateTime"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 学生特有字段（可选）
	StudentID *string `json:"student_id,omitempty" gorm:"uniqueIndex"`
	College   *string `json:"college,omitempty"`
	Major     *string `json:"major,omitempty"`
	Class     *string `json:"class,omitempty"`
	Grade     *string `json:"grade,omitempty"`

	// 教师特有字段（可选）
	Department *string `json:"department,omitempty"`
	Title      *string `json:"title,omitempty"`
	Specialty  *string `json:"specialty,omitempty"`
}

// BeforeCreate 在创建前自动生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserID == "" {
		u.UserID = uuid.New().String()
	}
	return nil
}

// UserRequest 用户注册/创建请求
type UserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password  string `json:"password" binding:"required,min=8"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"omitempty,len=11"`
	RealName  string `json:"real_name" binding:"required,min=2,max=50"`
	UserType  string `json:"user_type" binding:"required,oneof=student teacher"`
	StudentID string `json:"student_id" binding:"omitempty,len=8"` // 可选的学生ID，用于管理员创建学生时指定学号

	// 学生特有字段
	College string `json:"college" binding:"omitempty,max=100"`
	Major   string `json:"major" binding:"omitempty,max=100"`
	Class   string `json:"class" binding:"omitempty,max=50"`
	Grade   string `json:"grade" binding:"omitempty,len=4"`

	// 教师特有字段
	Department string `json:"department" binding:"omitempty,max=100"`
	Title      string `json:"title" binding:"omitempty,max=50"`
	Specialty  string `json:"specialty" binding:"omitempty,max=200"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	UserType string `json:"user_type"`
	Status   string `json:"status"`
	Avatar   string `json:"avatar"`

	// 学生特有字段
	StudentID *string `json:"student_id"`
	College   *string `json:"college"`
	Major     *string `json:"major"`
	Class     *string `json:"class"`
	Grade     *string `json:"grade"`

	// 教师特有字段
	Department *string `json:"department"`
	Title      *string `json:"title"`
	Specialty  *string `json:"specialty"`
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
	Avatar       string     `json:"avatar"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	RegisterTime time.Time  `json:"register_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 学生特有字段
	StudentID *string `json:"student_id,omitempty"`
	College   *string `json:"college,omitempty"`
	Major     *string `json:"major,omitempty"`
	Class     *string `json:"class,omitempty"`
	Grade     *string `json:"grade,omitempty"`

	// 教师特有字段
	Department *string `json:"department,omitempty"`
	Title      *string `json:"title,omitempty"`
	Specialty  *string `json:"specialty,omitempty"`
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

// StudentStats 学生统计信息
type StudentStats struct {
	TotalStudents     int64            `json:"total_students"`
	ActiveStudents    int64            `json:"active_students"`
	GraduatedStudents int64            `json:"graduated_students"`
	StudentsByCollege map[string]int64 `json:"students_by_college"`
	StudentsByMajor   map[string]int64 `json:"students_by_major"`
	StudentsByGrade   map[string]int64 `json:"students_by_grade"`
}

// TeacherStats 教师统计信息
type TeacherStats struct {
	TotalTeachers        int64            `json:"total_teachers"`
	ActiveTeachers       int64            `json:"active_teachers"`
	RetiredTeachers      int64            `json:"retired_teachers"`
	TeachersByDepartment map[string]int64 `json:"teachers_by_department"`
	TeachersByTitle      map[string]int64 `json:"teachers_by_title"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query      string `json:"query" form:"query"`
	UserType   string `json:"user_type" form:"user_type"`
	College    string `json:"college" form:"college"`
	Major      string `json:"major" form:"major"`
	Class      string `json:"class" form:"class"`
	Grade      string `json:"grade" form:"grade"`
	Department string `json:"department" form:"department"`
	Title      string `json:"title" form:"title"`
	Status     string `json:"status" form:"status"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
