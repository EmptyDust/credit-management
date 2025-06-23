package models

import "time"

type Application struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `gorm:"not null"`
	StudentID     uint   `gorm:"not null"`
	AffairID      uint   `gorm:"not null"`
	Title         string `gorm:"not null"`
	Description   string
	Type          string     `gorm:"not null"`                   // 申请类型
	Status        string     `gorm:"not null;default:'pending'"` // pending, approved, rejected
	Credits       float64    `gorm:"not null;default:0"`         // 认定学分
	AttachmentURL string     // 附件URL
	ReviewerID    *uint      // 审核人ID
	ReviewComment string     // 审核意见
	ReviewDate    *time.Time // 审核时间
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
