package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getCurrentUserRole(c *gin.Context) string {
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

func getCurrentUserID(c *gin.Context) string {
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

func canViewUserDetails(currentUserRole, targetUserType string) bool {
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
