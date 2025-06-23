package models

// Affair 事项模型
type Affair struct {
	ID   int    `json:"id" gorm:"primaryKey;column:affair_id"`
	Name string `json:"name" gorm:"unique;not null;column:affair_name"`
}

// TableName 指定表名
func (Affair) TableName() string {
	return "affairs"
}

// AffairStudent 事项-学生关系模型
type AffairStudent struct {
	AffairID  int    `json:"affair_id" gorm:"primaryKey;column:affair_id"`
	StudentID string `json:"student_id" gorm:"primaryKey;column:student_id"`
	IsPrimary bool   `json:"is_primary" gorm:"column:is_main_responsible"` // 是否主要负责人
}

// TableName 指定表名
func (AffairStudent) TableName() string {
	return "affair_students"
}
