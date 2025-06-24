package models

// Affair 事项模型
type Affair struct {
	ID          int    `json:"id" gorm:"primaryKey;column:affair_id"`
	Name        string `json:"name" gorm:"unique;not null;column:affair_name"`
	Description string `json:"description" gorm:"column:description"`
	CreatorID   string `json:"creator_id" gorm:"column:creator_id"`
	Attachments string `json:"attachments" gorm:"type:text;column:attachments"` // JSON字符串
}

// TableName 指定表名
func (Affair) TableName() string {
	return "affairs"
}

// AffairStudent 事项-学生关系模型
type AffairStudent struct {
	AffairID   int    `json:"affair_id" gorm:"primaryKey;column:affair_id"`
	StudentID  string `json:"student_id" gorm:"primaryKey;column:student_id"`
	IsPrimary  bool   `json:"is_primary" gorm:"column:is_main_responsible"` // 是否主要负责人
	Role       string `json:"role" gorm:"column:role"` // 角色：primary/member
}

// TableName 指定表名
func (AffairStudent) TableName() string {
	return "affair_students"
}

// 可选：如需支持附件表
// type AffairAttachment struct {
// 	ID        int    `json:"id" gorm:"primaryKey"`
// 	AffairID  int    `json:"affair_id"`
// 	FileName  string `json:"file_name"`
// 	FileURL   string `json:"file_url"`
// }
// func (AffairAttachment) TableName() string { return "affair_attachments" }
