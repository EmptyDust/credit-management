package models

import (
	"time"
)

// StudentBasicResponse 学生基本信息响应（学生可查看其他学生的基本信息）
type StudentBasicResponse struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	RealName     string    `json:"real_name"`
	StudentID    *string   `json:"student_id,omitempty"`
	College      *string   `json:"college,omitempty"`
	Major        *string   `json:"major,omitempty"`
	Class        *string   `json:"class,omitempty"`
	Grade        *string   `json:"grade,omitempty"`
	Status       string    `json:"status"`
	Avatar       string    `json:"avatar"`
	RegisterTime time.Time `json:"register_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TeacherBasicResponse 教师基本信息响应（学生和教师可查看教师的基本信息）
type TeacherBasicResponse struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	RealName     string    `json:"real_name"`
	Department   *string   `json:"department,omitempty"`
	Title        *string   `json:"title,omitempty"`
	Status       string    `json:"status"`
	Avatar       string    `json:"avatar"`
	RegisterTime time.Time `json:"register_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StudentDetailResponse 学生详细信息响应（教师可查看学生的详细信息）
type StudentDetailResponse struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	RealName     string     `json:"real_name"`
	StudentID    *string    `json:"student_id,omitempty"`
	College      *string    `json:"college,omitempty"`
	Major        *string    `json:"major,omitempty"`
	Class        *string    `json:"class,omitempty"`
	Grade        *string    `json:"grade,omitempty"`
	Status       string     `json:"status"`
	Avatar       string     `json:"avatar"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	RegisterTime time.Time  `json:"register_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TeacherDetailResponse 教师详细信息响应（管理员可查看教师的详细信息）
type TeacherDetailResponse struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	RealName     string     `json:"real_name"`
	Department   *string    `json:"department,omitempty"`
	Title        *string    `json:"title,omitempty"`
	Specialty    *string    `json:"specialty,omitempty"`
	Status       string     `json:"status"`
	Avatar       string     `json:"avatar"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	RegisterTime time.Time  `json:"register_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RoleBasedSearchResponse 基于角色的搜索响应
type RoleBasedSearchResponse struct {
	Users      []interface{} `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}
