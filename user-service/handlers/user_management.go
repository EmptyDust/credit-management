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

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := getCurrentUserID(c)
	if userID == "" {
		userID = currentUserID
	}

	var userType string
	err := h.db.Table("users").Select("user_type").Where("user_id = ?", userID).Scan(&userType).Error
	if err != nil || userType == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		return
	}

	if userID == currentUserID {
		var result map[string]interface{}
		var viewName string
		switch userType {
		case "student":
			viewName = "student_complete_info"
		case "teacher":
			viewName = "teacher_complete_info"
		default:

			// 管理员等其他类型直接查users表所有字段
			h.db.Table("users").Where("user_id = ?", userID).Find(&result)
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
			return
		}
		h.db.Table(viewName).Where("user_id = ?", userID).Find(&result)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
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

	currentUserRole := getCurrentUserRole(c)
	if !canViewUserDetails(currentUserRole, user.UserType) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
		return
	}

	response := h.convertToUserResponse(user)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

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

	if req.Email != "" {
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

	if req.StudentID != nil {
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

	if req.Department != nil {
		user.Department = req.Department
	}
	if req.Title != nil {
		user.Title = req.Title
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

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
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

func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	var users []models.User
	if err := h.db.Where("user_id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		return
	}

	if len(users) != len(req.UserIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部分用户不存在", "data": nil})
		return
	}

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

func (h *UserHandler) BatchUpdateUserStatus(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
		Status  string   `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

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

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	if !models.ValidatePasswordComplexity(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码必须包含大小写字母和数字，且长度至少8位", "data": nil})
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

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

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "data": nil})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "原密码错误", "data": nil})
		return
	}

	if !models.ValidatePasswordComplexity(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码必须包含大小写字母和数字，且长度至少8位", "data": nil})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

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
		"user_id":     userID,
		"activities":  []interface{}{},
		"total":       0,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *UserHandler) ExportUsers(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	userType := c.Query("user_type")
	status := c.Query("status")

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
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    gin.H{"message": "CSV导出功能待实现", "count": len(users)},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的导出格式", "data": nil})
	}
}
