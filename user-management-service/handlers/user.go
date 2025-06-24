package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"credit-management/user-management-service/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db                 *gorm.DB
	studentServiceURL  string
	teacherServiceURL  string
}

func NewUserHandler(db *gorm.DB, jwtSecret, studentServiceURL, teacherServiceURL string) *UserHandler {
	return &UserHandler{
		db:                 db,
		studentServiceURL:  studentServiceURL,
		teacherServiceURL:  teacherServiceURL,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 检查邮箱是否已存在
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被使用"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		UserType: req.UserType,
		Role:     "user",
		Status:   "active",
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
		return
	}

	// 根据用户类型创建对应的学生或教师记录
	if err := h.createUserProfile(user); err != nil {
		// 如果创建用户档案失败，删除已创建的用户
		h.db.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户档案失败: " + err.Error()})
		return
	}

	// 返回用户信息（不包含密码）
	userResponse := models.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Role:         user.Role,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户注册成功",
		"user":    userResponse,
	})
}

// createUserProfile 根据用户类型创建对应的学生或教师记录
func (h *UserHandler) createUserProfile(user models.User) error {
	switch user.UserType {
	case "student":
		return h.createStudentProfile(user)
	case "teacher":
		return h.createTeacherProfile(user)
	default:
		return fmt.Errorf("不支持的用户类型: %s", user.UserType)
	}
}

// createStudentProfile 创建学生档案
func (h *UserHandler) createStudentProfile(user models.User) error {
	// 构建学生创建请求
	studentReq := map[string]interface{}{
		"username":   user.Username,
		"student_id": user.ID, // 使用用户ID作为学号
		"name":       user.RealName,
		"email":      user.Email,
		"contact":    user.Phone,
		"status":     "active",
	}

	// 通过API Gateway调用学生信息服务创建学生记录
	return h.callExternalService("POST", "http://api-gateway:8000/api/students", studentReq)
}

// createTeacherProfile 创建教师档案
func (h *UserHandler) createTeacherProfile(user models.User) error {
	// 构建教师创建请求
	teacherReq := map[string]interface{}{
		"username": user.Username,
		"name":     user.RealName,
		"email":    user.Email,
		"contact":  user.Phone,
		"status":   "active",
	}

	// 通过API Gateway调用教师信息服务创建教师记录
	return h.callExternalService("POST", "http://api-gateway:8000/api/teachers", teacherReq)
}

// callExternalService 调用外部服务
func (h *UserHandler) callExternalService(method, url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("调用外部服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 新增日志
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("调用外部服务: %s %s\n请求体: %s\n响应状态: %d\n响应内容: %s\n", method, url, string(jsonData), resp.StatusCode, string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("外部服务返回错误状态码: %d, 响应内容: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	// 如果URL中没有id，则从JWT token中获取
	if userID == "" {
		jwtUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			return
		}
		userID = jwtUserID.(string)
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 返回用户信息（不包含密码）
	userResponse := models.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Role:         user.Role,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		jwtUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			return
		}
		userID = jwtUserID.(string)
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := h.db.Where("email = ? AND id != ?", req.Email, user.ID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被其他用户使用"})
			return
		}
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.UserType != "" {
		updates["user_type"] = req.UserType
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败: " + err.Error()})
			return
		}
	}

	// 重新查询用户信息
	h.db.First(&user, user.ID)

	// 返回更新后的用户信息
	userResponse := models.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Role:         user.Role,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户信息更新成功",
		"user":    userResponse,
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 软删除用户
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// GetAllUsers 获取所有用户（分页）
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userType := c.Query("user_type")
	status := c.Query("status")

	query := h.db.Model(&models.User{})

	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		return
	}

	// 转换为响应格式（不包含密码）
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponse := models.UserResponse{
			ID:           user.ID,
			Username:     user.Username,
			Email:        user.Email,
			Phone:        user.Phone,
			RealName:     user.RealName,
			UserType:     user.UserType,
			Role:         user.Role,
			Status:       user.Status,
			Avatar:       user.Avatar,
			LastLoginAt:  user.LastLoginAt,
			RegisterTime: user.RegisterTime,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       userResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}

// GetUsersByType 根据用户类型获取用户
func (h *UserHandler) GetUsersByType(c *gin.Context) {
	userType := c.Param("userType")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户类型不能为空"})
		return
	}

	var users []models.User
	if err := h.db.Where("user_type = ?", userType).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		return
	}

	// 转换为响应格式（不包含密码）
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponse := models.UserResponse{
			ID:           user.ID,
			Username:     user.Username,
			Email:        user.Email,
			Phone:        user.Phone,
			RealName:     user.RealName,
			UserType:     user.UserType,
			Role:         user.Role,
			Status:       user.Status,
			Avatar:       user.Avatar,
			LastLoginAt:  user.LastLoginAt,
			RegisterTime: user.RegisterTime,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, userResponses)
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats := models.UserStats{}

	// 总用户数
	h.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// 活跃用户数
	h.db.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)

	// 禁用用户数
	h.db.Model(&models.User{}).Where("status = ?", "suspended").Count(&stats.SuspendedUsers)

	// 各类型用户数
	h.db.Model(&models.User{}).Where("user_type = ?", "student").Count(&stats.StudentUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "teacher").Count(&stats.TeacherUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "admin").Count(&stats.AdminUsers)

	// 新增用户统计
	today := time.Now().Truncate(24 * time.Hour)
	h.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday)

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	h.db.Model(&models.User{}).Where("created_at >= ?", weekStart).Count(&stats.NewUsersWeek)

	monthStart := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	monthStart = monthStart.Truncate(24 * time.Hour)
	h.db.Model(&models.User{}).Where("created_at >= ?", monthStart).Count(&stats.NewUsersMonth)

	c.JSON(http.StatusOK, stats)
}
