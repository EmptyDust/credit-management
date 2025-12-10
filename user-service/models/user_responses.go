package models

// ViewBasedSearchResponse 基于视图的搜索响应
type ViewBasedSearchResponse struct {
	Users      []map[string]interface{} `json:"users"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
	ViewType   string                   `json:"view_type"`
}

// UserStats 用户统计信息
type UserStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	SuspendedUsers    int64 `json:"suspended_users"`
	StudentUsers      int64 `json:"student_users"`
	TeacherUsers      int64 `json:"teacher_users"`
	AdminUsers        int64 `json:"admin_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersWeek      int64 `json:"new_users_week"`
	NewUsersMonth     int64 `json:"new_users_month"`
	NewUsersLastMonth int64 `json:"new_users_last_month"`
}

// StudentStats 学生统计信息
type StudentStats struct {
	TotalStudents     int64            `json:"total_students"`
	ActiveStudents    int64            `json:"active_students"`
	GraduatedStudents int64            `json:"graduated_students"`
	StudentsByCollege map[string]int64 `json:"students_by_college"`
	StudentsByMajor   map[string]int64 `json:"students_by_major"`
	StudentsByGrade   map[string]int64 `json:"students_by_grade"`
}

// TeacherStats 教师统计信息
type TeacherStats struct {
	TotalTeachers        int64            `json:"total_teachers"`
	ActiveTeachers       int64            `json:"active_teachers"`
	RetiredTeachers      int64            `json:"retired_teachers"`
	TeachersByDepartment map[string]int64 `json:"teachers_by_department"`
	TeachersByTitle      map[string]int64 `json:"teachers_by_title"`
}
