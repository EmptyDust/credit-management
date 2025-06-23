package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/user-management-service/models"
	"credit-management/user-management-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db                  *gorm.DB
	jwtSecret           []byte
	fileUploader        *utils.FileUploader
	permissionManager   *utils.PermissionManager
	notificationManager *utils.NotificationManager
}

func NewUserHandler(db *gorm.DB, jwtSecret string) *UserHandler {
	fileUploader := utils.NewFileUploader(nil)
	permissionManager := utils.NewPermissionManager(db)
	notificationManager := utils.NewNotificationManager(db)

	return &UserHandler{
		db:                  db,
		jwtSecret:           []byte(jwtSecret),
		fileUploader:        fileUploader,
		permissionManager:   permissionManager,
		notificationManager: notificationManager,
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

	// 根据用户类型分配默认角色
	var defaultRoleID uint
	switch req.UserType {
	case "student":
		defaultRoleID = 4 // student角色
	case "teacher":
		defaultRoleID = 3 // teacher角色
	case "admin":
		defaultRoleID = 1 // admin角色
	default:
		defaultRoleID = 5 // user角色
	}

	h.permissionManager.AssignRole(user.ID, defaultRoleID)

	// 发送欢迎通知
	h.notificationManager.SendTemplateNotification(user.ID, utils.UserRegisteredTemplate, map[string]interface{}{
		"username": user.Username,
	})

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

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "账户已被禁用，请联系管理员"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	h.db.Model(&user).Update("last_login_at", &now)

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"username":  user.Username,
		"user_type": user.UserType,
		"role":      user.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
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

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:   tokenString,
		User:    userResponse,
		Message: "登录成功",
	})
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
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
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 更新字段
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

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户更新成功"})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// GetAllUsers 获取所有用户
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

// ValidateToken 验证JWT token
func (h *UserHandler) ValidateToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少Authorization头"})
		return
	}

	// 移除"Bearer "前缀
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"user_id":   claims["user_id"],
		"username":  claims["username"],
		"user_type": claims["user_type"],
		"role":      claims["role"],
	})
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}

	// 上传文件
	fileInfo, err := h.fileUploader.UploadFile(file, "avatars")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	// 保存文件记录到数据库
	userFile := models.UserFile{
		UserID:       userID.(uint),
		FileName:     fileInfo.FileName,
		OriginalName: fileInfo.OriginalName,
		FilePath:     fileInfo.FilePath,
		FileSize:     fileInfo.FileSize,
		FileType:     fileInfo.FileType,
		MimeType:     fileInfo.MimeType,
		Category:     "avatar",
		Description:  "用户头像",
		IsPublic:     true,
	}

	if err := h.db.Create(&userFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件记录失败"})
		return
	}

	// 更新用户头像
	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("avatar", fileInfo.FilePath).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户头像失败"})
		return
	}

	// 发送通知
	h.notificationManager.SendTemplateNotification(userID.(uint), utils.FileUploadedTemplate, map[string]interface{}{
		"filename": fileInfo.OriginalName,
		"filesize": fileInfo.FileSize,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "头像上传成功",
		"file": models.FileResponse{
			ID:            userFile.ID,
			FileName:      userFile.FileName,
			OriginalName:  userFile.OriginalName,
			FileSize:      userFile.FileSize,
			FileType:      userFile.FileType,
			MimeType:      userFile.MimeType,
			Category:      userFile.Category,
			Description:   userFile.Description,
			IsPublic:      userFile.IsPublic,
			DownloadCount: userFile.DownloadCount,
			DownloadURL:   h.fileUploader.GetFileURL(userFile.FilePath),
			PreviewURL:    h.fileUploader.GetPreviewURL(userFile.FilePath),
			CreatedAt:     userFile.CreatedAt,
		},
	})
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
