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
	db                *gorm.DB
	studentServiceURL string
	teacherServiceURL string
}

func NewUserHandler(db *gorm.DB, jwtSecret, studentServiceURL, teacherServiceURL string) *UserHandler {
	return &UserHandler{
		db:                db,
		studentServiceURL: studentServiceURL,
		teacherServiceURL: teacherServiceURL,
	}
}

// Register 用户注册（仅限学生）
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] 注册参数绑定失败: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 只允许注册学生
	if req.UserType != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能注册学生用户", "data": nil})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		return
	}

	// 检查邮箱是否已存在
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		UserType: "student", // 强制设置为学生
		Status:   "active",
	}
	// UserID将由BeforeCreate钩子自动生成

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败: " + err.Error(), "data": nil})
		return
	}

	// 创建学生档案，支持自定义学号
	studentID := req.StudentID
	if studentID == "" {
		studentID = user.UserID
	}
	if err := h.createStudentProfile(user, studentID); err != nil {
		h.db.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建学生档案失败: " + err.Error(), "data": nil})
		return
	}

	// 返回用户信息（不包含密码）
	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "学生注册成功",
			"user":    userResponse,
		},
	})
}

// CreateTeacher 管理员创建教师
func (h *UserHandler) CreateTeacher(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证，无法操作", "data": nil})
		return
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok || claimsMap["user_type"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有管理员可以创建教师", "data": nil})
		return
	}
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}
	if req.UserType != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能创建教师用户", "data": nil})
		return
	}
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		return
	}
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		UserType: "teacher",
		Status:   "active",
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败: " + err.Error(), "data": nil})
		return
	}
	if err := h.createTeacherProfile(user); err != nil {
		h.db.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建教师档案失败: " + err.Error(), "data": nil})
		return
	}
	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "教师创建成功",
			"user":    userResponse,
		},
	})
}

// CreateStudent 管理员创建学生
func (h *UserHandler) CreateStudent(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证，无法操作", "data": nil})
		return
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok || claimsMap["user_type"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有管理员可以创建学生", "data": nil})
		return
	}

	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}
	if req.UserType != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能创建学生用户", "data": nil})
		return
	}
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		return
	}
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		UserType: "student",
		Status:   "active",
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败: " + err.Error(), "data": nil})
		return
	}
	studentID := req.StudentID
	if studentID == "" {
		studentID = user.UserID
	}
	if err := h.createStudentProfile(user, studentID); err != nil {
		h.db.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建学生档案失败: " + err.Error(), "data": nil})
		return
	}
	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "学生创建成功",
			"user":    userResponse,
		},
	})
}

// createStudentProfile 创建学生档案（支持自定义学号）
func (h *UserHandler) createStudentProfile(user models.User, studentID string) error {
	studentReq := map[string]interface{}{
		"user_id":    user.UserID, // 必填，自动生成UUID
		"username":   user.Username,
		"student_id": studentID, // 支持自定义学号
		"name":       user.RealName,
		"email":      user.Email,
		"contact":    user.Phone,
		"status":     "active",
	}
	return h.callExternalService("POST", "http://api-gateway:8080/api/students", studentReq)
}

// createTeacherProfile 创建教师档案
func (h *UserHandler) createTeacherProfile(user models.User) error {
	teacherReq := map[string]interface{}{
		"user_id":  user.UserID, // 必填，自动生成UUID
		"username": user.Username,
		"name":     user.RealName,
		"email":    user.Email,
		"contact":  user.Phone,
		"status":   "active",
	}
	return h.callExternalService("POST", "http://api-gateway:8080/api/teachers", teacherReq)
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
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			return
		}
		userID = jwtUserID.(string)
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 创建响应map
	response := map[string]interface{}{
		"id":            user.UserID,
		"username":      user.Username,
		"email":         user.Email,
		"phone":         user.Phone,
		"real_name":     user.RealName,
		"user_type":     user.UserType,
		"status":        user.Status,
		"avatar":        user.Avatar,
		"last_login_at": user.LastLoginAt,
		"register_time": user.RegisterTime,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	// 根据用户类型获取额外的学生或教师信息
	if user.UserType == "student" {
		studentInfo, err := h.getStudentInfo(user.Username)
		if err != nil {
			// 如果获取学生信息失败，记录错误但不影响基本用户信息的返回
			fmt.Printf("获取学生信息失败: %v\n", err)
		} else {
			// 将学生信息合并到响应中
			for key, value := range studentInfo {
				response[key] = value
			}
		}
	} else if user.UserType == "teacher" {
		teacherInfo, err := h.getTeacherInfo(user.Username)
		if err != nil {
			// 如果获取教师信息失败，记录错误但不影响基本用户信息的返回
			fmt.Printf("获取教师信息失败: %v\n", err)
		} else {
			// 将教师信息合并到响应中
			for key, value := range teacherInfo {
				response[key] = value
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": response})
}

// getStudentInfo 获取学生信息
func (h *UserHandler) getStudentInfo(username string) (map[string]interface{}, error) {
	// 通过API Gateway调用学生信息服务
	url := "http://api-gateway:8080/api/students/search?username=" + username

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用学生信息服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("学生信息服务返回错误状态码: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析学生信息响应失败: %v", err)
	}

	// 提取学生信息
	if students, ok := response["students"].([]interface{}); ok && len(students) > 0 {
		if student, ok := students[0].(map[string]interface{}); ok {
			return student, nil
		}
	}

	return nil, fmt.Errorf("未找到学生信息")
}

// getTeacherInfo 获取教师信息
func (h *UserHandler) getTeacherInfo(username string) (map[string]interface{}, error) {
	// 通过API Gateway调用教师信息服务
	url := "http://api-gateway:8080/api/teachers/search?username=" + username

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用教师信息服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("教师信息服务返回错误状态码: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析教师信息响应失败: %v", err)
	}

	// 提取教师信息
	if teachers, ok := response["teachers"].([]interface{}); ok && len(teachers) > 0 {
		if teacher, ok := teachers[0].(map[string]interface{}); ok {
			return teacher, nil
		}
	}

	return nil, fmt.Errorf("未找到教师信息")
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		jwtUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			return
		}
		userID = jwtUserID.(string)
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := h.db.Where("email = ? AND user_id != ?", req.Email, user.UserID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被其他用户使用", "data": nil})
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

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户失败: " + err.Error(), "data": nil})
		return
	}

	// 返回更新后的用户信息
	userResponse := models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": userResponse})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 权限检查：只有管理员可以删除用户
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证，无法操作", "data": nil})
		return
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok || claimsMap["user_type"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有管理员可以删除用户", "data": nil})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	// 检查是否尝试删除自己
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能删除自己的账户", "data": nil})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 检查是否为系统管理员（防止删除超级管理员）
	if user.UserType == "admin" {
		// 检查是否还有其他管理员
		var adminCount int64
		h.db.Model(&models.User{}).Where("user_type = ? AND user_id != ?", "admin", userID).Count(&adminCount)
		if adminCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能删除最后一个系统管理员", "data": nil})
			return
		}
	}

	// 开始事务处理
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 根据用户类型删除关联数据
	if user.UserType == "student" {
		// 删除学生档案
		if err := h.deleteStudentProfile(user.UserID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除学生档案失败: " + err.Error(), "data": nil})
			return
		}
	} else if user.UserType == "teacher" {
		// 删除教师档案
		if err := h.deleteTeacherProfile(user.UserID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除教师档案失败: " + err.Error(), "data": nil})
			return
		}
	}

	// 软删除用户（设置删除时间）
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除用户失败: " + err.Error(), "data": nil})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交事务失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": fmt.Sprintf("用户 %s 删除成功", user.Username),
			"deleted_user": gin.H{
				"user_id":   user.UserID,
				"username":  user.Username,
				"user_type": user.UserType,
			},
		},
	})
}

// deleteStudentProfile 删除学生档案
func (h *UserHandler) deleteStudentProfile(userID string) error {
	// 调用学生信息服务删除学生档案
	return h.callExternalService("DELETE", fmt.Sprintf("http://api-gateway:8080/api/students/user/%s", userID), nil)
}

// deleteTeacherProfile 删除教师档案
func (h *UserHandler) deleteTeacherProfile(userID string) error {
	// 调用教师信息服务删除教师档案
	return h.callExternalService("DELETE", fmt.Sprintf("http://api-gateway:8080/api/teachers/user/%s", userID), nil)
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
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error(), "data": nil})
		return
	}

	// 转换为响应格式（不包含密码）
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponse := models.UserResponse{
			UserID:       user.UserID,
			Username:     user.Username,
			Email:        user.Email,
			Phone:        user.Phone,
			RealName:     user.RealName,
			UserType:     user.UserType,
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
		"code":    0,
		"message": "success",
		"data": gin.H{
			"users":       userResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// GetUsersByType 根据用户类型获取用户
func (h *UserHandler) GetUsersByType(c *gin.Context) {
	userType := c.Param("userType")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户类型不能为空", "data": nil})
		return
	}

	var users []models.User
	if err := h.db.Where("user_type = ?", userType).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败: " + err.Error(), "data": nil})
		return
	}

	// 转换为响应格式（不包含密码）
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponse := models.UserResponse{
			UserID:       user.UserID,
			Username:     user.Username,
			Email:        user.Email,
			Phone:        user.Phone,
			RealName:     user.RealName,
			UserType:     user.UserType,
			Status:       user.Status,
			Avatar:       user.Avatar,
			LastLoginAt:  user.LastLoginAt,
			RegisterTime: user.RegisterTime,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": userResponses})
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

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": stats})
}
