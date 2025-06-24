package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"

	"credit-management/auth-service/models"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	h.db.Model(&user).Update("last_login_at", &now)

	// 生成JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 生成refresh token
	refreshToken, err := h.generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 构建响应
	userResponse := models.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Role:         user.Role,
		Status:       user.Status,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          userResponse,
		"message":       "Login successful",
	})
}

// ValidateToken 验证JWT token
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.TokenValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 解析token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusOK, models.TokenValidationResponse{
			Valid:   false,
			Message: "Invalid token",
		})
		return
	}

	// 获取用户信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusOK, models.TokenValidationResponse{
			Valid:   false,
			Message: "Invalid token claims",
		})
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusOK, models.TokenValidationResponse{
			Valid:   false,
			Message: "Invalid user ID in token",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, models.TokenValidationResponse{
			Valid:   false,
			Message: "User not found",
		})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusOK, models.TokenValidationResponse{
			Valid:   false,
			Message: "Account is not active",
		})
		return
	}

	// 构建用户响应
	userResponse := models.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Role:         user.Role,
		Status:       user.Status,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.TokenValidationResponse{
		Valid:   true,
		User:    userResponse,
		Message: "Token is valid",
	})
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 解析refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 获取用户信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
		return
	}

	// 生成新的token
	newToken, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	// 生成新的refresh token
	newRefreshToken, err := h.generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
		return
	}

	c.JSON(http.StatusOK, models.RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		Message:      "Token refreshed successfully",
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 在实际应用中，这里可以将token加入黑名单
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// generateToken 生成JWT token
func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"username":  user.Username,
		"user_type": user.UserType,
		"role":      user.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// generateRefreshToken 生成refresh token
func (h *AuthHandler) generateRefreshToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// InitializeAdminUser 初始化默认管理员用户
func InitializeAdminUser(db *gorm.DB) error {
	// Check if admin user exists
	var userCount int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&userCount)
	if userCount == 0 {
		// Create admin user
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
		adminUser := models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@example.com",
			RealName: "Administrator",
			UserType: "admin",
			Status:   "active",
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
	}

	// Check if admin role exists
	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create admin role if it does not exist
			adminRole = models.Role{Name: "admin", Description: "Administrator role", IsSystem: true}
			if err := db.Create(&adminRole).Error; err != nil {
				log.Printf("failed to create admin role: %v", err)
				return err
			}
		} else {
			return err
		}
	}

	// Assign admin role to admin user
	var adminUser models.User
	db.Where("username = ?", "admin").First(&adminUser)
	var userRoleCount int64
	db.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).Count(&userRoleCount)
	if userRoleCount == 0 {
		userRole := models.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID}
		if err := db.Create(&userRole).Error; err != nil {
			return err
		}
	}

	return nil
}
