package handlers

import (
	"strconv"

	"credit-management/user-service/models"
	"credit-management/user-service/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := utils.GetCurrentUserID(c)
	if userID == "" {
		userID = currentUserID
	}

	var userType string
	err := h.db.Table("users").Select("user_type").Where("user_id = ?", userID).Scan(&userType).Error
	if err != nil || userType == "" {
		utils.SendNotFound(c, "用户不存在")
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
			utils.SendSuccessResponse(c, result)
			return
		}
		h.db.Table(viewName).Where("user_id = ?", userID).Find(&result)
		utils.SendSuccessResponse(c, result)
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	currentUserRole := utils.GetCurrentUserRole(c)
	if !utils.CanViewUserDetails(currentUserRole, user.UserType) {
		utils.SendForbidden(c, "权限不足")
		return
	}

	response := h.convertToUserResponse(user)
	utils.SendSuccessResponse(c, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		claims, exists := utils.GetUserClaims(c)
		if !exists {
			utils.SendUnauthorized(c)
			return
		}
		userID = claims["user_id"].(string)
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 验证邮箱唯一性
	if req.Email != "" {
		var existingUser models.User
		if err := h.db.Where("email = ? AND user_id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			utils.SendConflict(c, "邮箱已被使用")
			return
		}
		user.Email = req.Email
	}

	// 验证学号唯一性
	if req.StudentID != nil {
		var existingUser models.User
		if err := h.db.Where("student_id = ? AND user_id != ?", *req.StudentID, userID).First(&existingUser).Error; err == nil {
			utils.SendConflict(c, "学号已被使用")
			return
		}
		user.StudentID = req.StudentID
	}

	// 更新用户信息
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
		utils.SendInternalServerError(c, err)
		return
	}

	userResponse := h.convertToUserResponse(user)
	utils.SendSuccessResponse(c, userResponse)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.SendBadRequest(c, "用户ID不能为空")
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "用户删除成功"})
}

func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	var users []models.User
	if err := h.db.Where("user_id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	if len(users) != len(req.UserIDs) {
		utils.SendBadRequest(c, "部分用户不存在")
		return
	}

	if err := h.db.Delete(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"deleted_count": len(users)})
}

func (h *UserHandler) BatchUpdateUserStatus(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
		Status  string   `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.db.Model(&models.User{}).Where("user_id IN ?", req.UserIDs).Update("status", req.Status).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"updated_count": len(req.UserIDs), "status": req.Status})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "密码重置成功"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := utils.GetCurrentUserID(c)
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		utils.SendBadRequest(c, "原密码错误")
		return
	}

	// 验证新密码复杂度
	validator := utils.NewValidator()
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "密码修改成功"})
}

func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		userID = utils.GetCurrentUserID(c)
		if userID == "" {
			utils.SendUnauthorized(c)
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	validator := utils.NewValidator()
	page, pageSize, _ = validator.ValidatePagination(strconv.Itoa(page), strconv.Itoa(pageSize))

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

	utils.SendSuccessResponse(c, response)
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
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "json":
		utils.SendSuccessResponse(c, users)
	case "csv":
		utils.SendSuccessResponse(c, gin.H{"message": "CSV导出功能待实现", "count": len(users)})
	default:
		utils.SendBadRequest(c, "不支持的导出格式")
	}
}

// convertToUserResponse 将User模型转换为UserResponse
func (h *UserHandler) convertToUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        utils.DerefString(user.Phone),
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       utils.DerefString(user.Avatar),
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		StudentID:    user.StudentID,
		College:      user.College,
		Major:        user.Major,
		Class:        user.Class,
		Grade:        user.Grade,
		Department:   user.Department,
		Title:        user.Title,
	}
}
