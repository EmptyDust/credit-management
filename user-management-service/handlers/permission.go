package handlers

import (
	"net/http"
	"strconv"

	"credit-management/user-management-service/models"
	"credit-management/user-management-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db                *gorm.DB
	permissionManager *utils.PermissionManager
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	permissionManager := utils.NewPermissionManager(db)

	return &PermissionHandler{
		db:                db,
		permissionManager: permissionManager,
	}
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req models.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查角色名是否已存在
	var existingRole models.Role
	if err := h.db.Where("name = ?", req.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "角色名已存在"})
		return
	}

	role, err := h.permissionManager.CreateRole(req.Name, req.Description, req.IsSystem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

// GetRoles 获取角色列表
func (h *PermissionHandler) GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询角色失败: " + err.Error()})
		return
	}

	var roleResponses []models.RoleResponse
	for _, role := range roles {
		roleResponse := models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}
		roleResponses = append(roleResponses, roleResponse)
	}

	c.JSON(http.StatusOK, roleResponses)
}

// GetRole 获取角色详情
func (h *PermissionHandler) GetRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询角色失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询角色失败: " + err.Error()})
		}
		return
	}

	// 系统角色不允许修改名称
	if role.IsSystem && req.Name != "" && req.Name != role.Name {
		c.JSON(http.StatusForbidden, gin.H{"error": "系统角色不允许修改名称"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["is_system"] = req.IsSystem

	if err := h.db.Model(&role).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色更新成功"})
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询角色失败: " + err.Error()})
		}
		return
	}

	// 系统角色不允许删除
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "系统角色不允许删除"})
		return
	}

	// 检查是否有用户使用该角色
	var userCount int64
	h.db.Model(&models.UserRole{}).Where("role_id = ?", roleID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "该角色下还有用户，无法删除"})
		return
	}

	if err := h.db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色删除成功"})
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req models.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查权限名是否已存在
	var existingPermission models.Permission
	if err := h.db.Where("name = ?", req.Name).First(&existingPermission).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "权限名已存在"})
		return
	}

	permission, err := h.permissionManager.CreatePermission(req.Name, req.Description, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		Resource:    permission.Resource,
		Action:      permission.Action,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	})
}

// GetPermissions 获取权限列表
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := h.db.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败: " + err.Error()})
		return
	}

	var permissionResponses []models.PermissionResponse
	for _, permission := range permissions {
		permissionResponse := models.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		}
		permissionResponses = append(permissionResponses, permissionResponse)
	}

	c.JSON(http.StatusOK, permissionResponses)
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	permissionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	var permission models.Permission
	if err := h.db.Preload("Users").First(&permission, permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, permission)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	permissionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	var permission models.Permission
	if err := h.db.First(&permission, permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限删除成功"})
}

// AssignRole 分配角色给用户
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 检查角色是否存在
	var role models.Role
	if err := h.db.First(&role, req.RoleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询角色失败: " + err.Error()})
		}
		return
	}

	if err := h.permissionManager.AssignRole(req.UserID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色分配成功"})
}

// RemoveRole 移除用户角色
func (h *PermissionHandler) RemoveRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	roleID, err := strconv.ParseUint(c.Param("roleID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	if err := h.permissionManager.RemoveRole(uint(userID), uint(roleID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色移除成功"})
}

// AssignPermission 分配权限给用户
func (h *PermissionHandler) AssignPermission(c *gin.Context) {
	var req models.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败: " + err.Error()})
		}
		return
	}

	// 检查权限是否存在
	var permission models.Permission
	if err := h.db.First(&permission, req.PermissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败: " + err.Error()})
		}
		return
	}

	if err := h.permissionManager.AssignPermission(req.UserID, req.PermissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限分配成功"})
}

// RemovePermission 移除用户权限
func (h *PermissionHandler) RemovePermission(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permissionID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	if err := h.permissionManager.RemovePermission(uint(userID), uint(permissionID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限移除成功"})
}

// GetUserRoles 获取用户角色
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	roles, err := h.permissionManager.GetUserRoles(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// GetUserPermissions 获取用户权限
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	permissions, err := h.permissionManager.GetUserPermissions(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// AssignPermissionToRole 给角色分配权限
func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permissionID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	if err := h.permissionManager.AssignPermissionToRole(uint(roleID), uint(permissionID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限分配成功"})
}

// RemovePermissionFromRole 移除角色权限
func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permissionID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	if err := h.permissionManager.RemovePermissionFromRole(uint(roleID), uint(permissionID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除权限失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限移除成功"})
}

// InitializePermissions 初始化默认权限和角色
func (h *PermissionHandler) InitializePermissions(c *gin.Context) {
	// 初始化默认权限
	if err := h.permissionManager.InitializeDefaultPermissions(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化权限失败: " + err.Error()})
		return
	}

	// 初始化默认角色
	if err := h.permissionManager.InitializeDefaultRoles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限和角色初始化成功"})
}
