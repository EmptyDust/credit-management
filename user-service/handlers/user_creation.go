package handlers

import (
	"net/http"

	"credit-management/user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register 用户注册（仅限学生）
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 只允许注册学生
	if req.UserType != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能注册学生用户", "data": nil})
		return
	}

	// 验证请求数据
	if err := h.validateUserRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查用户名是否已存在（包括软删除的用户）
	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	// 检查邮箱是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	// 检查手机号是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
		}
		return
	}

	// 检查学号是否已存在（如果提供了学号，包括软删除的用户）
	if req.StudentID != "" {
		if err := h.db.Unscoped().Where("student_id = ?", req.StudentID).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被删除的用户使用，请使用其他学号", "data": nil})
			} else {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
			}
			return
		}
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
		UserType: "student",
		Status:   "active",
	}

	// 设置学生特有字段
	if req.StudentID != "" {
		user.StudentID = &req.StudentID
	}
	if req.College != "" {
		user.College = &req.College
	}
	if req.Major != "" {
		user.Major = &req.Major
	}
	if req.Class != "" {
		user.Class = &req.Class
	}
	if req.Grade != "" {
		user.Grade = &req.Grade
	}

	if err := h.db.Create(&user).Error; err != nil {
		// 记录详细错误到日志，但不返回给客户端
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	// 返回用户信息
	userResponse := h.convertToUserResponse(user)
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
	claimsMap, ok := claims.(jwt.MapClaims)
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

	// 验证请求数据
	if err := h.validateUserRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查用户名是否已存在（包括软删除的用户）
	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	// 检查邮箱是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	// 检查手机号是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
		}
		return
	}

	// 检查学号是否已存在（如果提供了学号，包括软删除的用户）
	if req.StudentID != "" {
		if err := h.db.Unscoped().Where("student_id = ?", req.StudentID).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被删除的用户使用，请使用其他学号", "data": nil})
			} else {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
			}
			return
		}
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
		UserType: "teacher",
		Status:   "active",
	}

	// 设置教师特有字段
	if req.Department != "" {
		user.Department = &req.Department
	}
	if req.Title != "" {
		user.Title = &req.Title
	}
	if req.Specialty != "" {
		user.Specialty = &req.Specialty
	}

	if err := h.db.Create(&user).Error; err != nil {
		// 记录详细错误到日志，但不返回给客户端
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	// 返回用户信息
	userResponse := h.convertToUserResponse(user)
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
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
		return
	}

	// 调试日志
	userType, exists := claimsMap["user_type"]
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
		return
	}

	if userType != "admin" {
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

	// 验证请求数据
	if err := h.validateUserRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查用户名是否已存在（包括软删除的用户）
	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	// 检查邮箱是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	// 检查手机号是否已存在（包括软删除的用户）
	if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
		}
		return
	}

	// 检查学号是否已存在（如果提供了学号，包括软删除的用户）
	if req.StudentID != "" {
		if err := h.db.Unscoped().Where("student_id = ?", req.StudentID).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被删除的用户使用，请使用其他学号", "data": nil})
			} else {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
			}
			return
		}
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
		UserType: "student",
		Status:   "active",
	}

	// 设置学生特有字段
	if req.StudentID != "" {
		user.StudentID = &req.StudentID
	}
	if req.College != "" {
		user.College = &req.College
	}
	if req.Major != "" {
		user.Major = &req.Major
	}
	if req.Class != "" {
		user.Class = &req.Class
	}
	if req.Grade != "" {
		user.Grade = &req.Grade
	}

	if err := h.db.Create(&user).Error; err != nil {
		// 记录详细错误到日志，但不返回给客户端
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	// 返回用户信息
	userResponse := h.convertToUserResponse(user)
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "学生创建成功",
			"user":    userResponse,
		},
	})
}
