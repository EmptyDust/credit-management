package handlers

import (
	"net/http"
	"strconv"

	"credit-management/user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetUser 获取用户信息（基于角色的权限控制）
func (h *UserHandler) GetUser(c *gin.Context) {
	// 获取当前用户角色
	currentUserRole := getCurrentUserRole(c)
	if currentUserRole == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
		return
	}

	userID := c.Param("id")

	// 如果没有提供用户ID，则获取当前用户信息
	if userID == "" {
		userID = getCurrentUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			return
		}
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		}
		return
	}

	// 检查是否为查看自己的信息
	currentUserID := getCurrentUserID(c)
	isOwnProfile := (userID == currentUserID)

	// 根据角色转换响应
	response := h.convertToRoleBasedResponse(user, currentUserRole, isOwnProfile)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	// 如果没有提供用户ID，则更新当前用户信息
	if userID == "" {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			return
		}
		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			return
		}
		userID = claimsMap["user_id"].(string)
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		}
		return
	}

	// 更新基本信息
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser models.User
		if err := h.db.Where("email = ? AND user_id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
			return
		}
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Avatar != "" {
		user.Avatar = &req.Avatar
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	// 更新学生特有字段
	if req.StudentID != nil {
		// 检查学号是否已被其他用户使用
		var existingUser models.User
		if err := h.db.Where("student_id = ? AND user_id != ?", *req.StudentID, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
			return
		}
		user.StudentID = req.StudentID
	}
	if req.College != nil {
		user.College = req.College
	}
	if req.Major != nil {
		user.Major = req.Major
	}
	if req.Class != nil {
		user.Class = req.Class
	}
	if req.Grade != nil {
		user.Grade = req.Grade
	}

	// 更新教师特有字段
	if req.Department != nil {
		user.Department = req.Department
	}
	if req.Title != nil {
		user.Title = req.Title
	}
	if req.Specialty != nil {
		user.Specialty = req.Specialty
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户信息失败", "data": nil})
		return
	}

	userResponse := h.convertToUserResponse(user)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    userResponse,
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		}
		return
	}

	// 软删除用户
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除用户失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"message": "用户删除成功"},
	})
}

// GetAllUsers 获取所有用户（管理员功能）
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var users []models.User
	var total int64

	// 获取总数
	h.db.Model(&models.User{}).Count(&total)

	// 获取用户列表
	if err := h.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败", "data": nil})
		return
	}

	// 转换为响应格式
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, h.convertToUserResponse(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"users":       userResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetStudents 获取所有学生（基于角色的权限控制）
func (h *UserHandler) GetStudents(c *gin.Context) {
	// 设置用户类型为student
	c.Params = append(c.Params, gin.Param{Key: "userType", Value: "student"})
	h.GetUsersByType(c)
}

// GetTeachers 获取所有教师（基于角色的权限控制）
func (h *UserHandler) GetTeachers(c *gin.Context) {
	// 设置用户类型为teacher
	c.Params = append(c.Params, gin.Param{Key: "userType", Value: "teacher"})
	h.GetUsersByType(c)
}

// GetUsersByType 根据用户类型获取用户列表（基于角色的权限控制）
func (h *UserHandler) GetUsersByType(c *gin.Context) {
	// 获取当前用户角色
	currentUserRole := getCurrentUserRole(c)
	if currentUserRole == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
		return
	}

	userType := c.Param("userType")
	if userType == "" {
		userType = c.Query("user_type")
	}

	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户类型不能为空", "data": nil})
		return
	}

	// 根据用户角色限制访问范围
	switch currentUserRole {
	case "student":
		// 学生只能查看学生和教师的基本信息
		if userType != "student" && userType != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足，只能查看学生和教师信息", "data": nil})
			return
		}
	case "teacher":
		// 教师可以查看学生详细信息和其他教师基本信息
		if userType != "student" && userType != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足，只能查看学生和教师信息", "data": nil})
			return
		}
	case "admin":
		// 管理员可以查看所有用户的所有信息
		// 不限制访问范围
	default:
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var users []models.User
	var total int64

	// 获取总数
	h.db.Model(&models.User{}).Where("user_type = ?", userType).Count(&total)

	// 获取用户列表
	if err := h.db.Where("user_type = ?", userType).Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败", "data": nil})
		return
	}

	// 根据角色转换响应
	var responses []interface{}
	for _, user := range users {
		response := h.convertToRoleBasedResponse(user, currentUserRole, false)
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"users":       responses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// BatchDeleteUsers 批量删除用户
func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 检查用户是否存在
	var users []models.User
	if err := h.db.Where("user_id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		return
	}

	if len(users) != len(req.UserIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部分用户不存在", "data": nil})
		return
	}

	// 批量软删除用户
	if err := h.db.Delete(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "批量删除用户失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"deleted_count": len(users)},
	})
}

// BatchUpdateUserStatus 批量更新用户状态
func (h *UserHandler) BatchUpdateUserStatus(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
		Status  string   `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 批量更新用户状态
	if err := h.db.Model(&models.User{}).Where("user_id IN ?", req.UserIDs).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "批量更新用户状态失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"updated_count": len(req.UserIDs), "status": req.Status},
	})
}

// ResetPassword 重置用户密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 验证密码强度
	if err := h.validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.db.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		}
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码重置失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"message": "密码重置成功"},
	})
}

// ChangePassword 用户修改自己的密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
		return
	}

	// 获取用户信息
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "原密码错误", "data": nil})
		return
	}

	// 验证新密码强度
	if err := h.validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码修改失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"message": "密码修改成功"},
	})
}

// GetUserActivity 获取用户活动记录
func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		userID = getCurrentUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 这里可以添加用户活动记录的查询逻辑
	// 例如：登录记录、操作日志等
	// 目前返回空结果，后续可以扩展

	response := gin.H{
		"user_id":   userID,
		"activities": []interface{}{},
		"total":      0,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// ExportUsers 导出用户数据
func (h *UserHandler) ExportUsers(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	userType := c.Query("user_type")
	status := c.Query("status")

	// 构建查询
	query := h.db.Model(&models.User{})

	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户数据失败", "data": nil})
		return
	}

	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    users,
		})
	case "csv":
		// 这里可以实现CSV导出功能
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    gin.H{"message": "CSV导出功能待实现", "count": len(users)},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的导出格式", "data": nil})
	}
}
