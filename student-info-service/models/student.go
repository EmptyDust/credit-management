package models

type Student struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"unique;not null"`
	Username  string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	StudentNo string `gorm:"unique;not null"`
	College   string
	Major     string
	Class     string
	Contact   string
} 