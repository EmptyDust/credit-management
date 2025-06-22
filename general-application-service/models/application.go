package models

import "time"

type Application struct {
    ID            uint      `gorm:"primaryKey"`
    AffairID      uint      // 事项编号
    StudentID     string    // 学生学号
    ApplyTime     time.Time // 申请时间
    Status        string    // 状态
    ReviewerID    *string   // 审核者ID
    ReviewOpinion string    // 审核意见
    AppliedCredit float64   // 申请学分
    FinalCredit   float64   // 认定学分
    CreatedAt     time.Time
    UpdatedAt     time.Time
} 