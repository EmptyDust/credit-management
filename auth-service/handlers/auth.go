package handlers

import (
	"context"
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"credit-management/auth-service/models"
	"credit-management/auth-service/utils"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
	redis     *utils.RedisClient
}

func NewAuthHandler(db *gorm.DB, jwtSecret string, redis *utils.RedisClient) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
		redis:     redis,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	// 检查至少提供了UID或用户名
	if req.UID == "" && req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "必须提供 uid 或 username", "data": nil})
		return
	}

	var user models.User
	query := h.db
	if req.UID != "" {
		// 使用UID查询
		query = query.Where("identity_number = ?", req.UID)
	} else {
		// 使用用户名查询
		query = query.Where("username = ?", req.Username)
	}

	if err := query.First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误", "data": nil})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误", "data": nil})
		return
	}

	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账户未激活", "data": nil})
		return
	}

	now := time.Now()
	h.db.Model(&user).Update("last_login_at", &now)

	// 生成JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成token失败", "data": nil})
		return
	}

	// 生成refresh token
	refreshToken, err := h.generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成refresh token失败", "data": nil})
		return
	}

	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        "",
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		LastLoginAt:  nil,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	if user.Phone != nil {
		userResponse.Phone = *user.Phone
	}
	if user.LastLoginAt != nil {
		userResponse.LastLoginAt = user.LastLoginAt
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token":         token,
			"refresh_token": refreshToken,
			"user":          userResponse,
			"message":       "登录成功",
		},
	})
}

// ValidateToken 验证JWT token（增强版，包含黑名单检查）
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.TokenValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	// 检查token是否在黑名单中
	ctx := context.Background()
	if blacklisted, err := h.redis.IsBlacklisted(ctx, req.Token); err != nil {
		log.Printf("检查token黑名单失败: %v", err)
	} else if blacklisted {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "token已被撤销",
			},
		})
		return
	}

	// 解析token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "无效的token",
			},
		})
		return
	}

	// 获取用户信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "无效的token claims",
			},
		})
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "token中的用户ID无效",
			},
		})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "用户不存在",
			},
		})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": models.TokenValidationResponse{
				Valid:   false,
				Message: "账户未激活",
			},
		})
		return
	}

	// 构建用户响应
	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        "",
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		LastLoginAt:  nil,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	// 安全地设置指针字段
	if user.Phone != nil {
		userResponse.Phone = *user.Phone
	}
	if user.LastLoginAt != nil {
		userResponse.LastLoginAt = user.LastLoginAt
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.TokenValidationResponse{
			Valid:   true,
			User:    userResponse,
			Message: "Token有效",
		},
	})
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	// 解析refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的refresh token", "data": nil})
		return
	}

	// 获取用户信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token claims", "data": nil})
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的用户ID", "data": nil})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在", "data": nil})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账号未激活", "data": nil})
		return
	}

	// 生成新的token
	newToken, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成新的token失败", "data": nil})
		return
	}

	// 生成新的refresh token
	newRefreshToken, err := h.generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成新的refresh token失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.RefreshTokenResponse{
			Token:        newToken,
			RefreshToken: newRefreshToken,
			Message:      "Token刷新成功",
		},
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少认证令牌", "data": nil})
		return
	}

	// 检查Bearer前缀
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "认证令牌格式错误", "data": nil})
		return
	}

	tokenString := authHeader[7:]

	// 解析JWT token获取过期时间
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的认证令牌", "data": nil})
		return
	}

	// 获取token的过期时间
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的token claims", "data": nil})
		return
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "token中缺少过期时间", "data": nil})
		return
	}

	// 计算剩余时间
	expTime := time.Unix(int64(exp), 0)
	remainingTime := time.Until(expTime)

	if remainingTime > 0 {
		// 将token添加到黑名单
		ctx := context.Background()
		if err := h.redis.AddToBlacklist(ctx, tokenString, remainingTime); err != nil {
			log.Printf("添加token到黑名单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登出失败", "data": nil})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"message": "登出成功"}})
}

// ValidatePermission 验证用户权限
func (h *AuthHandler) ValidatePermission(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证令牌", "data": nil})
		return
	}

	// 检查Bearer前缀
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌格式错误", "data": nil})
		return
	}

	tokenString := authHeader[7:]

	// 解析JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证令牌", "data": nil})
		return
	}

	// 获取用户信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token claims", "data": nil})
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token中的用户ID无效", "data": nil})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在", "data": nil})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账户未激活", "data": nil})
		return
	}

	// 返回用户权限信息
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":        user.UserID,
			"username":  user.Username,
			"user_type": user.UserType,
			"status":    user.Status,
			"real_name": user.RealName,
		},
	})
}

// ValidateTokenWithClaims 验证token并返回claims信息（供其他服务调用）
func (h *AuthHandler) ValidateTokenWithClaims(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证令牌", "data": nil})
		return
	}

	// 检查Bearer前缀
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌格式错误", "data": nil})
		return
	}

	tokenString := authHeader[7:]

	// 解析JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证令牌", "data": nil})
		return
	}

	// 获取claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token claims", "data": nil})
		return
	}

	// 验证用户是否存在且状态正常
	userID, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token中的用户ID无效", "data": nil})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在", "data": nil})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账户未激活", "data": nil})
		return
	}

	// 返回验证成功和claims信息
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.TokenValidationWithClaimsResponse{
			Valid:   true,
			Claims:  claims,
			Message: "token验证成功",
		},
	})
}

// generateToken 生成JWT token
func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":        user.UserID,
		"username":  user.Username,
		"user_type": user.UserType,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *AuthHandler) generateRefreshToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":   user.UserID,
		"type": "refresh",
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func InitializeAdminUser(db *gorm.DB) error {
	var userCount int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&userCount)
	if userCount == 0 {
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
		log.Println("Admin user created successfully")
	} else {
		log.Println("Admin user already exists")
	}

	return nil
}
