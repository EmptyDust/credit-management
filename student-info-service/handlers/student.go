package handlers

import (
	"net/http"
	"student-info-service/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudentHandler struct {
	DB *gorm.DB
}

func (h *StudentHandler) Register(c *gin.Context) {
	var req models.Student
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学号或用户名已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func (h *StudentHandler) GetStudent(c *gin.Context) {
	studentNo := c.Param("studentNo")
	var stu models.Student
	if err := h.DB.Where("student_no = ?", studentNo).First(&stu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	c.JSON(http.StatusOK, stu)
}

func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	studentNo := c.Param("studentNo")
	var req models.Student
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.Student{}).Where("student_no = ?", studentNo).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
} 