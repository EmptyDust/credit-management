package models

type Teacher struct {
    ID        uint   `gorm:"primaryKey"`
    Username  string `gorm:"unique;not null"`
    Name      string `gorm:"not null"`
    Contact   string
} 