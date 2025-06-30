package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetCurrentUserRole 获取当前用户角色
func GetCurrentUserRole(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	userType, exists := claimsMap["user_type"]
	if !exists {
		return ""
	}
	return userType.(string)
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	userID, exists := claimsMap["user_id"]
	if !exists {
		return ""
	}
	return userID.(string)
}

// CanViewUserDetails 检查当前用户是否可以查看目标用户的详细信息
func CanViewUserDetails(currentUserRole, targetUserType string) bool {
	switch currentUserRole {
	case "admin":
		return true
	case "teacher":
		return targetUserType == "student"
	case "student":
		return false
	default:
		return false
	}
}

// IsAdmin 检查用户是否为管理员
func IsAdmin(userRole string) bool {
	return userRole == "admin"
}

// IsTeacher 检查用户是否为教师
func IsTeacher(userRole string) bool {
	return userRole == "teacher"
}

// IsStudent 检查用户是否为学生
func IsStudent(userRole string) bool {
	return userRole == "student"
}

// IsTeacherOrAdmin 检查用户是否为教师或管理员
func IsTeacherOrAdmin(userRole string) bool {
	return userRole == "teacher" || userRole == "admin"
}

// IsStudentTeacherOrAdmin 检查用户是否为学生、教师或管理员
func IsStudentTeacherOrAdmin(userRole string) bool {
	return userRole == "student" || userRole == "teacher" || userRole == "admin"
}

// GetUserClaims 获取用户声明信息
func GetUserClaims(c *gin.Context) (jwt.MapClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}
	return claimsMap, true
}

// GetUsername 获取当前用户名
func GetUsername(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	username, exists := claimsMap["username"]
	if !exists {
		return ""
	}
	return username.(string)
}
