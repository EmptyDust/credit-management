package models

import "time"

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    Role      string    `gorm:"not null;default:'student'"`
    CreatedAt time.Time
} 