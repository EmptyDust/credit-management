package utils

import (
	"fmt"

	"credit-management/user-management-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionManager 权限管理器
type PermissionManager struct {
	db *gorm.DB
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager(db *gorm.DB) *PermissionManager {
	return &PermissionManager{db: db}
}

// CheckPermission 检查用户权限
func (pm *PermissionManager) CheckPermission(userID uint, resource, action string) (bool, error) {
	var count int64

	// 检查用户直接权限
	err := pm.db.Model(&models.UserPermission{}).
		Joins("JOIN permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ? AND permissions.resource = ? AND permissions.action = ?",
			userID, resource, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// 检查用户角色权限
	err = pm.db.Model(&models.RolePermission{}).
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.resource = ? AND permissions.action = ?",
			userID, resource, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasRole 检查用户是否有指定角色
func (pm *PermissionManager) HasRole(userID uint, roleName string) (bool, error) {
	var count int64

	err := pm.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, roleName).
		Count(&count).Error

	return count > 0, err
}

// GetUserPermissions 获取用户所有权限
func (pm *PermissionManager) GetUserPermissions(userID uint) ([]models.Permission, error) {
	var permissions []models.Permission

	// 获取直接权限
	err := pm.db.Model(&models.Permission{}).
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ?", userID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	// 获取角色权限
	var rolePermissions []models.Permission
	err = pm.db.Model(&models.Permission{}).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&rolePermissions).Error

	if err != nil {
		return nil, err
	}

	// 合并权限（去重）
	permissionMap := make(map[uint]models.Permission)
	for _, p := range permissions {
		permissionMap[p.ID] = p
	}
	for _, p := range rolePermissions {
		permissionMap[p.ID] = p
	}

	result := make([]models.Permission, 0, len(permissionMap))
	for _, p := range permissionMap {
		result = append(result, p)
	}

	return result, nil
}

// GetUserRoles 获取用户所有角色
func (pm *PermissionManager) GetUserRoles(userID uint) ([]models.Role, error) {
	var roles []models.Role

	err := pm.db.Model(&models.Role{}).
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error

	return roles, err
}

// AssignRole 分配角色给用户
func (pm *PermissionManager) AssignRole(userID, roleID uint) error {
	// 检查是否已存在
	var count int64
	err := pm.db.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("用户已拥有该角色")
	}

	userRole := models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	return pm.db.Create(&userRole).Error
}

// RemoveRole 移除用户角色
func (pm *PermissionManager) RemoveRole(userID, roleID uint) error {
	return pm.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error
}

// AssignPermission 分配权限给用户
func (pm *PermissionManager) AssignPermission(userID, permissionID uint) error {
	// 检查是否已存在
	var count int64
	err := pm.db.Model(&models.UserPermission{}).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("用户已拥有该权限")
	}

	userPermission := models.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
	}

	return pm.db.Create(&userPermission).Error
}

// RemovePermission 移除用户权限
func (pm *PermissionManager) RemovePermission(userID, permissionID uint) error {
	return pm.db.Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&models.UserPermission{}).Error
}

// CreatePermission 创建权限
func (pm *PermissionManager) CreatePermission(name, description, resource, action string) (*models.Permission, error) {
	permission := &models.Permission{
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
	}

	err := pm.db.Create(permission).Error
	return permission, err
}

// CreateRole 创建角色
func (pm *PermissionManager) CreateRole(name, description string, isSystem bool) (*models.Role, error) {
	role := &models.Role{
		Name:        name,
		Description: description,
		IsSystem:    isSystem,
	}

	err := pm.db.Create(role).Error
	return role, err
}

// AssignPermissionToRole 给角色分配权限
func (pm *PermissionManager) AssignPermissionToRole(roleID, permissionID uint) error {
	// 检查是否已存在
	var count int64
	err := pm.db.Model(&models.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("角色已拥有该权限")
	}

	rolePermission := models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	return pm.db.Create(&rolePermission).Error
}

// RemovePermissionFromRole 移除角色权限
func (pm *PermissionManager) RemovePermissionFromRole(roleID, permissionID uint) error {
	return pm.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

// InitializeDefaultPermissions 初始化默认权限
func (pm *PermissionManager) InitializeDefaultPermissions() error {
	// 用户管理权限
	permissions := []models.Permission{
		{Name: "用户查看", Description: "查看用户信息", Resource: "user", Action: "read"},
		{Name: "用户创建", Description: "创建新用户", Resource: "user", Action: "create"},
		{Name: "用户更新", Description: "更新用户信息", Resource: "user", Action: "update"},
		{Name: "用户删除", Description: "删除用户", Resource: "user", Action: "delete"},

		// 文件管理权限
		{Name: "文件上传", Description: "上传文件", Resource: "file", Action: "upload"},
		{Name: "文件下载", Description: "下载文件", Resource: "file", Action: "download"},
		{Name: "文件删除", Description: "删除文件", Resource: "file", Action: "delete"},
		{Name: "文件预览", Description: "预览文件", Resource: "file", Action: "preview"},

		// 权限管理权限
		{Name: "权限查看", Description: "查看权限信息", Resource: "permission", Action: "read"},
		{Name: "权限分配", Description: "分配权限", Resource: "permission", Action: "assign"},
		{Name: "权限撤销", Description: "撤销权限", Resource: "permission", Action: "revoke"},

		// 角色管理权限
		{Name: "角色查看", Description: "查看角色信息", Resource: "role", Action: "read"},
		{Name: "角色创建", Description: "创建角色", Resource: "role", Action: "create"},
		{Name: "角色更新", Description: "更新角色", Resource: "role", Action: "update"},
		{Name: "角色删除", Description: "删除角色", Resource: "role", Action: "delete"},

		// 通知管理权限
		{Name: "通知查看", Description: "查看通知", Resource: "notification", Action: "read"},
		{Name: "通知发送", Description: "发送通知", Resource: "notification", Action: "send"},
		{Name: "通知删除", Description: "删除通知", Resource: "notification", Action: "delete"},

		// 统计报表权限
		{Name: "统计查看", Description: "查看统计数据", Resource: "statistics", Action: "read"},
		{Name: "报表导出", Description: "导出报表", Resource: "statistics", Action: "export"},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		err := pm.db.Where("name = ?", permission.Name).First(&existingPermission).Error
		if err == gorm.ErrRecordNotFound {
			pm.db.Create(&permission)
		}
	}

	return nil
}

// InitializeDefaultRoles 初始化默认角色
func (pm *PermissionManager) InitializeDefaultRoles() error {
	// 创建默认角色
	roles := []models.Role{
		{Name: "admin", Description: "系统管理员", IsSystem: true},
		{Name: "moderator", Description: "内容审核员", IsSystem: true},
		{Name: "teacher", Description: "教师", IsSystem: true},
		{Name: "student", Description: "学生", IsSystem: true},
		{Name: "user", Description: "普通用户", IsSystem: true},
	}

	for _, role := range roles {
		var existingRole models.Role
		err := pm.db.Where("name = ?", role.Name).First(&existingRole).Error
		if err == gorm.ErrRecordNotFound {
			pm.db.Create(&role)
		}
	}

	return nil
}

// PermissionMiddleware 权限中间件
func PermissionMiddleware(pm *PermissionManager, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		hasPermission, err := pm.CheckPermission(userID.(uint), resource, action)
		if err != nil {
			c.JSON(500, gin.H{"error": "权限检查失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(403, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色中间件
func RoleMiddleware(pm *PermissionManager, roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		hasRole, err := pm.HasRole(userID.(uint), roleName)
		if err != nil {
			c.JSON(500, gin.H{"error": "角色检查失败"})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(403, gin.H{"error": "角色权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly 仅管理员中间件
func AdminOnly(pm *PermissionManager) gin.HandlerFunc {
	return RoleMiddleware(pm, "admin")
}

// TeacherOnly 仅教师中间件
func TeacherOnly(pm *PermissionManager) gin.HandlerFunc {
	return RoleMiddleware(pm, "teacher")
}

// StudentOnly 仅学生中间件
func StudentOnly(pm *PermissionManager) gin.HandlerFunc {
	return RoleMiddleware(pm, "student")
}
