package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"credit-management/auth-service/models"
)

type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req models.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
	}

	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建角色失败", "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	})
}

// GetRoles 获取所有角色
func (h *PermissionHandler) GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取角色失败", "data": nil})
		return
	}

	var responses []models.RoleResponse
	for _, role := range roles {
		responses = append(responses, models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": responses})
}

// GetRole 获取指定角色
func (h *PermissionHandler) GetRole(c *gin.Context) {
	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	var role models.Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误: " + err.Error(), "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	})
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	var role models.Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误: " + err.Error(), "data": nil})
		}
		return
	}

	// 更新字段
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	role.IsSystem = req.IsSystem

	if err := h.db.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新角色失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	})
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	var role models.Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误: " + err.Error(), "data": nil})
		}
		return
	}

	if role.IsSystem {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能删除系统角色", "data": nil})
		return
	}

	if err := h.db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除角色失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req models.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	permission := models.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": models.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		},
	})
}

// GetPermissions 获取所有权限
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := h.db.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取权限失败", "data": nil})
		return
	}

	var responses []models.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, models.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": responses})
}

// GetPermission 获取指定权限
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "权限ID不能为空", "data": nil})	
		return
	}

	var permission models.Permission
	if err := h.db.Where("id = ?", id).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "权限不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误: " + err.Error(), "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": models.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		},
	})
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "权限ID不能为空", "data": nil})
		return
	}

	var permission models.Permission
	if err := h.db.Where("id = ?", id).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "权限不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误: " + err.Error(), "data": nil})
		}
		return
	}

	if err := h.db.Delete(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// AssignRole 分配角色给用户
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	userRole := models.UserRole{
		UserID: userID,
		RoleID: req.RoleID,
	}

	if err := h.db.Create(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "分配角色失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// RemoveRole 移除用户角色
func (h *PermissionHandler) RemoveRole(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	if err := h.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除角色失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// AssignPermission 分配权限给用户
func (h *PermissionHandler) AssignPermission(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	var req models.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	userPermission := models.UserPermission{
		UserID:       userID,
		PermissionID: req.PermissionID,
	}

	if err := h.db.Create(&userPermission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "分配权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// RemovePermission 移除用户权限
func (h *PermissionHandler) RemovePermission(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	permissionID := c.Param("permissionID")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "权限ID不能为空", "data": nil})
		return
	}

	if err := h.db.Where("user_id = ? AND permission_id = ?", userID, permissionID).Delete(&models.UserPermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// AssignPermissionToRole 分配权限给角色
func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	var req models.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	rolePermission := models.RolePermission{
		RoleID:       roleID,
		PermissionID: req.PermissionID,
	}

	if err := h.db.Create(&rolePermission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "分配权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// RemovePermissionFromRole 移除角色权限
func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID := c.Param("roleID")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "角色ID不能为空", "data": nil})
		return
	}

	permissionID := c.Param("permissionID")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "权限ID不能为空", "data": nil})
		return
	}

	if err := h.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&models.RolePermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除权限失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// GetUserRoles 获取用户角色
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	var userRoles []models.UserRole
	if err := h.db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户角色失败", "data": nil})
		return
	}

	var roles []models.Role
	for _, userRole := range userRoles {
		var role models.Role
		if err := h.db.Where("id = ?", userRole.RoleID).First(&role).Error; err == nil {
			roles = append(roles, role)
		}
	}

	var responses []models.RoleResponse
	for _, role := range roles {
		responses = append(responses, models.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": responses})
}

// GetUserPermissions 获取用户权限
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	var userPermissions []models.UserPermission
	if err := h.db.Where("user_id = ?", userID).Find(&userPermissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户权限失败", "data": nil})
		return
	}

	var permissions []models.Permission
	for _, userPermission := range userPermissions {
		var permission models.Permission
		if err := h.db.Where("id = ?", userPermission.PermissionID).First(&permission).Error; err == nil {
			permissions = append(permissions, permission)
		}
	}

	var responses []models.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, models.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": responses})
}

// InitializePermissions 初始化权限
func (h *PermissionHandler) InitializePermissions(c *gin.Context) {
	// 创建基础权限
	permissions := []models.Permission{
		{Name: "用户管理", Description: "管理用户信息", Resource: "user", Action: "manage"},
		{Name: "权限管理", Description: "管理权限和角色", Resource: "permission", Action: "manage"},
		{Name: "申请管理", Description: "管理申请信息", Resource: "application", Action: "manage"},
		{Name: "事务管理", Description: "管理事务信息", Resource: "affair", Action: "manage"},
		{Name: "文件管理", Description: "管理文件上传下载", Resource: "file", Action: "manage"},
		{Name: "通知管理", Description: "管理通知信息", Resource: "notification", Action: "manage"},
	}

	// 创建权限并保存ID
	permissionMap := make(map[string]string)
	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := h.db.Where("resource = ? AND action = ?", permission.Resource, permission.Action).First(&existingPermission).Error; err != nil {
			if err := h.db.Create(&permission).Error; err == nil {
				permissionMap[permission.Resource+"_"+permission.Action] = permission.ID
			}
		} else {
			permissionMap[permission.Resource+"_"+permission.Action] = existingPermission.ID
		}
	}

	// 创建基础角色
	roles := []models.Role{
		{Name: "admin", Description: "系统管理员", IsSystem: true},
		{Name: "teacher", Description: "教师", IsSystem: true},
		{Name: "student", Description: "学生", IsSystem: true},
	}

	// 创建角色并保存ID
	roleMap := make(map[string]string)
	for _, role := range roles {
		var existingRole models.Role
		if err := h.db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err := h.db.Create(&role).Error; err == nil {
				roleMap[role.Name] = role.ID
			}
		} else {
			roleMap[role.Name] = existingRole.ID
		}
	}

	// 为admin角色分配所有权限
	if adminRoleID, exists := roleMap["admin"]; exists {
		for _, permissionID := range permissionMap {
			// 检查是否已经分配
			var existingRolePermission models.RolePermission
			if err := h.db.Where("role_id = ? AND permission_id = ?", adminRoleID, permissionID).First(&existingRolePermission).Error; err != nil {
				// 如果不存在，则分配权限
				rolePermission := models.RolePermission{
					RoleID:       adminRoleID,
					PermissionID: permissionID,
				}
				h.db.Create(&rolePermission)
			}
		}
	}

	// 只有在HTTP请求上下文中才返回JSON响应
	if c != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
	}
}
