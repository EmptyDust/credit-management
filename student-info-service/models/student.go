package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Student 学生模型
type Student struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"uniqueIndex;not null"` // 关联用户表的 user_id
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	StudentID *string        `json:"student_id" gorm:"column:student_id"`
	Name      string         `json:"name" gorm:"column:name;not null"`
	College   string         `json:"college" gorm:"column:college"`
	Major     string         `json:"major" gorm:"column:major"`
	Class     string         `json:"class" gorm:"column:class"`
	Contact   string         `json:"contact" gorm:"column:contact"`
	Email     string         `json:"email" gorm:"column:email"`
	Grade     string         `json:"grade" gorm:"column:grade"` // 年级
	Status    string         `json:"status" gorm:"column:status;default:'active'"` // active, inactive, graduated
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (s *Student) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Student) TableName() string {
	return "students"
}

// StudentRequest 学生创建请求
type StudentRequest struct {
	UserID    string `json:"user_id" binding:"required"` // 关联用户表的 user_id
	Username  string `json:"username" binding:"required"`
	StudentID string `json:"student_id"`
	Name      string `json:"name" binding:"required"`
	College   string `json:"college"`
	Major     string `json:"major"`
	Class     string `json:"class"`
	Contact   string `json:"contact"`
	Email     string `json:"email"`
	Grade     string `json:"grade"`
}

// StudentUpdateRequest 学生更新请求
type StudentUpdateRequest struct {
	Name    string `json:"name"`
	College string `json:"college"`
	Major   string `json:"major"`
	Class   string `json:"class"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
	Grade   string `json:"grade"`
	Status  string `json:"status"`
} 