package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Teacher 教师模型
type Teacher struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string         `json:"user_id" gorm:"uniqueIndex;not null"` // 关联用户表的 user_id
	Username   string         `json:"username" gorm:"uniqueIndex;not null"`
	Name       string         `json:"name" gorm:"column:name;not null"`
	Contact    string         `json:"contact" gorm:"column:contact"`
	Email      string         `json:"email" gorm:"column:email"`
	Department string         `json:"department" gorm:"column:department"`          // 所属院系
	Title      string         `json:"title" gorm:"column:title"`                    // 职称
	Specialty  string         `json:"specialty" gorm:"column:specialty"`            // 专业领域
	Status     string         `json:"status" gorm:"column:status;default:'active'"` // active, inactive, retired
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (t *Teacher) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Teacher) TableName() string {
	return "teachers"
}

// TeacherRequest 教师创建请求
type TeacherRequest struct {
	UserID     string `json:"user_id" binding:"required"` // 关联用户表的 user_id
	Username   string `json:"username" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Contact    string `json:"contact"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Specialty  string `json:"specialty"`
}

// TeacherUpdateRequest 教师更新请求
type TeacherUpdateRequest struct {
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Specialty  string `json:"specialty"`
	Status     string `json:"status"`
}
