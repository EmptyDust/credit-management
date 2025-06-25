package handlers

import (
	"fmt"
	"net/http"

	"credit-management/teacher-info-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeacherHandler struct {
	db *gorm.DB
}

func NewTeacherHandler(db *gorm.DB) *TeacherHandler {
	return &TeacherHandler{db: db}
}

// CreateTeacher 创建教师
func (h *TeacherHandler) CreateTeacher(c *gin.Context) {
	var req models.TeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 检查用户名是否已存在
	var existingTeacher models.Teacher
	if err := h.db.Where("username = ?", req.Username).First(&existingTeacher).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		return
	}

	// 创建教师
	teacher := models.Teacher{
		UserID:     req.UserID,
		Username:   req.Username,
		Name:       req.Name,
		Contact:    req.Contact,
		Email:      req.Email,
		Department: req.Department,
		Title:      req.Title,
		Specialty:  req.Specialty,
		Status:     "active",
	}

	if err := h.db.Create(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": teacher,
	})
}

// GetTeacher 获取教师信息
func (h *TeacherHandler) GetTeacher(c *gin.Context) {
	teacherID := c.Param("id")
	if teacherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "教师ID不能为空", "data": nil})
		return
	}

	// 获取当前用户信息
	userType, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	var teacher models.Teacher
	err := h.db.Where("id = ?", teacherID).First(&teacher).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "教师不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 根据用户类型返回不同的数据
	if userType == "admin" {
		// 管理员可以看到所有信息
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teacher})
	} else if userType == "teacher" {
		// 检查是否是查看自己的信息
		if teacher.UserID == currentUserID {
			// 教师查看自己的信息，可以看到全部信息
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teacher})
		} else {
			// 教师查看其他教师信息，只能看到基本信息
			basicTeacher := gin.H{
				"id":         teacher.ID,
				"name":       teacher.Name,
				"department": teacher.Department,
				"title":      teacher.Title,
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": basicTeacher})
		}
	} else if userType == "student" {
		// 学生只能看到基本信息
		basicTeacher := gin.H{
			"id":         teacher.ID,
			"name":       teacher.Name,
			"department": teacher.Department,
			"title":      teacher.Title,
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": basicTeacher})
	} else {
		// 其他用户类型无权访问
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
	}
}

// UpdateTeacher 更新教师信息
func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	teacherID := c.Param("id")
	if teacherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "教师ID不能为空", "data": nil})
		return
	}

	var req models.TeacherUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	var teacher models.Teacher
	if err := h.db.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "教师不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Contact != "" {
		updates["contact"] = req.Contact
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Department != "" {
		updates["department"] = req.Department
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Specialty != "" {
		updates["specialty"] = req.Specialty
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.db.Model(&teacher).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teacher})
}

// DeleteTeacher 删除教师
func (h *TeacherHandler) DeleteTeacher(c *gin.Context) {
	teacherID := c.Param("id")
	if teacherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "教师ID不能为空", "data": nil})
		return
	}

	var teacher models.Teacher
	if err := h.db.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "教师不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		}
		return
	}

	if err := h.db.Delete(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// GetAllTeachers 获取所有教师
func (h *TeacherHandler) GetAllTeachers(c *gin.Context) {
	// 获取当前用户类型
	userType, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	var teachers []models.Teacher
	err := h.db.Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		return
	}

	// 根据用户类型返回不同的数据
	if userType == "admin" {
		// 管理员可以看到所有信息
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teachers})
	} else if userType == "teacher" || userType == "student" {
		// 学生和教师只能看到基本信息（UUID、姓名、部门、职称）
		var basicTeachers []gin.H
		for _, teacher := range teachers {
			basicTeachers = append(basicTeachers, gin.H{
				"id":         teacher.ID,
				"name":       teacher.Name,
				"department": teacher.Department,
				"title":      teacher.Title,
			})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": basicTeachers})
	} else {
		// 其他用户类型无权访问
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
	}
}

// GetTeachersByDepartment 根据院系获取教师
func (h *TeacherHandler) GetTeachersByDepartment(c *gin.Context) {
	department := c.Param("department")
	if department == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "院系不能为空", "data": nil})
		return
	}

	var teachers []models.Teacher
	err := h.db.Where("department = ?", department).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"teachers": teachers, "count": len(teachers)}})
}

// GetTeachersByTitle 根据职称获取教师
func (h *TeacherHandler) GetTeachersByTitle(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "职称不能为空", "data": nil})
		return
	}

	var teachers []models.Teacher
	err := h.db.Where("title = ?", title).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teachers})
}

// GetTeachersByStatus 根据状态获取教师
func (h *TeacherHandler) GetTeachersByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态不能为空", "data": nil})
		return
	}

	var teachers []models.Teacher
	err := h.db.Where("status = ?", status).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teachers})
}

// SearchTeachers 搜索教师
func (h *TeacherHandler) SearchTeachers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "搜索关键词不能为空", "data": nil})
		return
	}

	var teachers []models.Teacher
	err := h.db.Where("name LIKE ? OR department LIKE ? OR specialty LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teachers})
}

// GetActiveTeachers 获取活跃教师
func (h *TeacherHandler) GetActiveTeachers(c *gin.Context) {
	var teachers []models.Teacher
	err := h.db.Where("status = ?", "active").Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teachers})
}

// GetTeacherByUsername 根据用户名获取教师信息
func (h *TeacherHandler) GetTeacherByUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名不能为空", "data": nil})
		return
	}

	var teacher models.Teacher
	err := h.db.Where("username = ?", username).First(&teacher).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "教师不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []models.Teacher{teacher}})
}

// DeleteTeacherByUserID 根据用户ID删除教师档案
func (h *TeacherHandler) DeleteTeacherByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	// 先查找教师是否存在
	var teacher models.Teacher
	if err := h.db.Where("user_id = ?", userID).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该用户对应的教师信息不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询教师失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 软删除教师档案
	if err := h.db.Delete(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除教师档案失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0, 
		"message": "success", 
		"data": gin.H{
			"message": fmt.Sprintf("教师档案删除成功，用户ID: %s", userID),
		},
	})
} 