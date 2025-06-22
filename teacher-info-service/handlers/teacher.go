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

func (h *TeacherHandler) Register(c *gin.Context) {
	var req models.Teacher
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func (h *TeacherHandler) GetTeacher(c *gin.Context) {
	username := c.Param("username")
	var t models.Teacher
	if err := h.DB.Where("username = ?", username).First(&t).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	username := c.Param("username")
	var req models.Teacher
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.Teacher{}).Where("username = ?", username).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
} 