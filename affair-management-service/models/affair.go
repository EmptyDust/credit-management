package models

import (
	"time"
)

// Affair 事项模型
type Affair struct {
	ID          int       `json:"id" gorm:"primaryKey;column:affair_id"`
	Name        string    `json:"name" gorm:"column:affair_name;not null"`
	Description string    `json:"description" gorm:"column:description"`
	Category    string    `json:"category" gorm:"column:category"`              // 事项类别
	MaxCredits  float64   `json:"max_credits" gorm:"column:max_credits"`        // 最大可申请学分
	Status      string    `json:"status" gorm:"column:status;default:'active'"` // active, inactive
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Affair) TableName() string {
	return "affairs"
}

// AffairStudent 事项-学生关系模型
type AffairStudent struct {
	AffairID          int       `json:"affair_id" gorm:"primaryKey;column:affair_id"`
	StudentID         string    `json:"student_id" gorm:"primaryKey;column:student_id"`
	IsMainResponsible bool      `json:"is_main_responsible" gorm:"column:is_main_responsible;default:false"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联关系
	Affair  Affair  `json:"affair" gorm:"foreignKey:AffairID"`
	Student Student `json:"student" gorm:"foreignKey:StudentID;references:StudentID"`
}

// TableName 指定表名
func (AffairStudent) TableName() string {
	return "affair_students"
}

// 关联模型定义
type Student struct {
	Username  string `json:"username" gorm:"primaryKey;column:username"`
	StudentID string `json:"student_id" gorm:"column:student_id;unique;not null"`
	Name      string `json:"name" gorm:"column:name;not null"`
	College   string `json:"college" gorm:"column:college"`
	Major     string `json:"major" gorm:"column:major"`
	Class     string `json:"class" gorm:"column:class"`
	Contact   string `json:"contact" gorm:"column:contact"`
}

func (Student) TableName() string {
	return "students"
}

// AffairRequest 事项请求结构
type AffairRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	MaxCredits  float64 `json:"max_credits"`
}

// AffairUpdateRequest 事项更新请求结构
type AffairUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	MaxCredits  float64 `json:"max_credits"`
	Status      string  `json:"status"`
}

// AffairStudentRequest 事项-学生关系请求结构
type AffairStudentRequest struct {
	AffairID          int    `json:"affair_id" binding:"required"`
	StudentID         string `json:"student_id" binding:"required"`
	IsMainResponsible bool   `json:"is_main_responsible"`
}
