package authz

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleStudent, RoleTeacher, RoleAdmin:
		return true
	default:
		return false
	}
}

func CanReview(role string) bool {
	return role == RoleTeacher || role == RoleAdmin
}

func CanTransferActivityOwner(role string) bool {
	return role == RoleTeacher || role == RoleAdmin
}

func IsAdmin(role string) bool {
	return role == RoleAdmin
}
