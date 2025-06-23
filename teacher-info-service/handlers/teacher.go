package handlers

import (
	"net/http"
	"teacher-info-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeacherHandler struct {
	DB *gorm.DB
}

func NewTeacherHandler(db *gorm.DB) *TeacherHandler {
	return &TeacherHandler{DB: db}
}

// CreateTeacher 创建教师信息
func (h *TeacherHandler) CreateTeacher(c *gin.Context) {
	var teacher models.Teacher
	if err := c.ShouldBindJSON(&teacher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, teacher)
}

// GetTeachers 获取教师列表
func (h *TeacherHandler) GetTeachers(c *gin.Context) {
	var teachers []models.Teacher
	if err := h.DB.Find(&teachers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// GetTeacher 获取单个教师信息
func (h *TeacherHandler) GetTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := h.DB.First(&teacher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}

	c.JSON(http.StatusOK, teacher)
}

// GetTeacherByUserID 根据用户ID获取教师信息
func (h *TeacherHandler) GetTeacherByUserID(c *gin.Context) {
	userID := c.Param("userID")
	var teacher models.Teacher

	if err := h.DB.Where("user_id = ?", userID).First(&teacher).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}

	c.JSON(http.StatusOK, teacher)
}

// UpdateTeacher 更新教师信息
func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := h.DB.First(&teacher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}

	var updateData models.Teacher
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&teacher).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teacher)
}

// DeleteTeacher 删除教师信息
func (h *TeacherHandler) DeleteTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := h.DB.First(&teacher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}

	if err := h.DB.Delete(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teacher deleted successfully"})
}
