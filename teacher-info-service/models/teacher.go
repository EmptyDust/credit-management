package models

import (
	"time"
)

// Teacher 教师模型
type Teacher struct {
	Username   string    `json:"username" gorm:"primaryKey;column:username"`
	Name       string    `json:"name" gorm:"column:name;not null"`
	Contact    string    `json:"contact" gorm:"column:contact"`
	Email      string    `json:"email" gorm:"column:email"`
	Department string    `json:"department" gorm:"column:department"`          // 所属院系
	Title      string    `json:"title" gorm:"column:title"`                    // 职称
	Specialty  string    `json:"specialty" gorm:"column:specialty"`            // 专业领域
	Status     string    `json:"status" gorm:"column:status;default:'active'"` // active, inactive, retired
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:Username"`
}

// TableName 指定表名
func (Teacher) TableName() string {
	return "teachers"
}

// 关联模型定义
type User struct {
	Username     string    `json:"username" gorm:"primaryKey;column:username"`
	UserType     string    `json:"user_type" gorm:"column:user_type;not null"`
	RegisterTime time.Time `json:"register_time" gorm:"column:register_time;autoCreateTime"`
}

func (User) TableName() string {
	return "users"
}

// TeacherRequest 教师创建请求
type TeacherRequest struct {
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
