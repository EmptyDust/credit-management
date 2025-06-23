package models

import "time"

type Affair struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Description string
	Type        string `gorm:"not null"`                  // 事项类型
	Status      string `gorm:"not null;default:'active'"` // active, inactive
	StartDate   time.Time
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
