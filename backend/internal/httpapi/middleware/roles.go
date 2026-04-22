package middleware

import (
	"net/http"

	"credit-management/backend/internal/platform/authz"

	"github.com/gin-gonic/gin"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       "AUTH_TOKEN_INVALID",
				"message":    "access token is invalid",
				"request_id": c.GetString(RequestIDKey),
			})
			return
		}
		if !authz.IsValidRole(claims.UserType) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":       "ACCESS_FORBIDDEN",
				"message":    "role is invalid",
				"request_id": c.GetString(RequestIDKey),
			})
			return
		}
		if _, ok := allowed[claims.UserType]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":       "ACCESS_FORBIDDEN",
				"message":    "access forbidden",
				"request_id": c.GetString(RequestIDKey),
			})
			return
		}

		c.Next()
	}
}
