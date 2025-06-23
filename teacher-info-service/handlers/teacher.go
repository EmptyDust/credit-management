package handlers

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingTeacher models.Teacher
	if err := h.db.Where("username = ?", req.Username).First(&existingTeacher).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 创建教师
	teacher := models.Teacher{
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "教师创建成功",
		"teacher": teacher,
	})
}

// GetTeacher 获取教师信息
func (h *TeacherHandler) GetTeacher(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var teacher models.Teacher
	err := h.db.Where("username = ?", username).First(&teacher).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, teacher)
}

// UpdateTeacher 更新教师信息
func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var req models.TeacherUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var teacher models.Teacher
	if err := h.db.Where("username = ?", username).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "教师信息更新成功"})
}

// DeleteTeacher 删除教师
func (h *TeacherHandler) DeleteTeacher(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var teacher models.Teacher
	if err := h.db.Where("username = ?", username).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "教师删除成功"})
}

// GetAllTeachers 获取所有教师
func (h *TeacherHandler) GetAllTeachers(c *gin.Context) {
	var teachers []models.Teacher
	err := h.db.Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// GetTeachersByDepartment 根据院系获取教师
func (h *TeacherHandler) GetTeachersByDepartment(c *gin.Context) {
	department := c.Param("department")
	if department == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "院系不能为空"})
		return
	}

	var teachers []models.Teacher
	err := h.db.Where("department = ?", department).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// GetTeachersByTitle 根据职称获取教师
func (h *TeacherHandler) GetTeachersByTitle(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "职称不能为空"})
		return
	}

	var teachers []models.Teacher
	err := h.db.Preload("User").Where("title = ?", title).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// GetTeachersByStatus 根据状态获取教师
func (h *TeacherHandler) GetTeachersByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态不能为空"})
		return
	}

	var teachers []models.Teacher
	err := h.db.Preload("User").Where("status = ?", status).Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// SearchTeachers 搜索教师
func (h *TeacherHandler) SearchTeachers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	var teachers []models.Teacher
	err := h.db.Preload("User").
		Where("name LIKE ? OR department LIKE ? OR title LIKE ? OR specialty LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// GetActiveTeachers 获取活跃教师
func (h *TeacherHandler) GetActiveTeachers(c *gin.Context) {
	var teachers []models.Teacher
	err := h.db.Preload("User").Where("status = ?", "active").Find(&teachers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询教师失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}
