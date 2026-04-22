package middleware

import (
	"net/http"
	"strings"

	"credit-management/backend/internal/platform/authn"

	"github.com/gin-gonic/gin"
)

const (
	AuthClaimsKey  = "auth_claims"
	AccessTokenKey = "access_token"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := bearerToken(c.GetHeader("Authorization"))
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       "AUTH_TOKEN_REQUIRED",
				"message":    "authorization header is required",
				"request_id": c.GetString(RequestIDKey),
			})
			return
		}

		claims, err := authn.ParseAccessToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       "AUTH_TOKEN_INVALID",
				"message":    "access token is invalid",
				"request_id": c.GetString(RequestIDKey),
			})
			return
		}

		c.Set(AuthClaimsKey, claims)
		c.Set(AccessTokenKey, accessToken)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (authn.Claims, bool) {
	value, ok := c.Get(AuthClaimsKey)
	if !ok {
		return authn.Claims{}, false
	}

	claims, ok := value.(authn.Claims)
	return claims, ok
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
