package domain

import "time"

type User struct {
	ID                  string
	Username            string
	PasswordHash        string
	Email               string
	Phone               string
	RealName            string
	UserType            string
	Status              string
	AvatarFileID        string
	DepartmentID        string
	StudentID           string
	TeacherID           string
	Grade               string
	Title               string
	RefreshTokenVersion int64
	LastLoginAt         *time.Time
}

type ProfilePatch struct {
	Email        string
	Phone        string
	RealName     string
	Grade        string
	Title        string
	DepartmentID string
}
