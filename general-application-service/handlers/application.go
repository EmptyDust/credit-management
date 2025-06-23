package handlers

import (
	"general-application-service/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	DB *gorm.DB
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{DB: db}
}

// CreateApplication 创建申请
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var application models.Application
	if err := c.ShouldBindJSON(&application); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, application)
}

// GetApplications 获取申请列表
func (h *ApplicationHandler) GetApplications(c *gin.Context) {
	var applications []models.Application
	if err := h.DB.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// GetApplication 获取单个申请
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	var application models.Application

	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, application)
}

// GetApplicationsByUser 根据用户ID获取申请列表
func (h *ApplicationHandler) GetApplicationsByUser(c *gin.Context) {
	userID := c.Param("userID")
	var applications []models.Application

	if err := h.DB.Where("user_id = ?", userID).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// GetApplicationsByStudent 根据学生ID获取申请列表
func (h *ApplicationHandler) GetApplicationsByStudent(c *gin.Context) {
	studentID := c.Param("studentID")
	var applications []models.Application

	if err := h.DB.Where("student_id = ?", studentID).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// UpdateApplication 更新申请
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	var application models.Application

	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	var updateData models.Application
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&application).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// ReviewApplication 审核申请
func (h *ApplicationHandler) ReviewApplication(c *gin.Context) {
	id := c.Param("id")
	var application models.Application

	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	var reviewData struct {
		Status        string  `json:"status"`
		Credits       float64 `json:"credits"`
		ReviewComment string  `json:"review_comment"`
		ReviewerID    uint    `json:"reviewer_id"`
	}

	if err := c.ShouldBindJSON(&reviewData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	updateData := map[string]interface{}{
		"status":         reviewData.Status,
		"credits":        reviewData.Credits,
		"review_comment": reviewData.ReviewComment,
		"reviewer_id":    reviewData.ReviewerID,
		"review_date":    &now,
	}

	if err := h.DB.Model(&application).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// DeleteApplication 删除申请
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	var application models.Application

	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if err := h.DB.Delete(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application deleted successfully"})
}

func (h *ApplicationHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.Application{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *ApplicationHandler) UpdateFinalCredit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		FinalCredit float64 `json:"final_credit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.Application{}).Where("id = ?", id).Update("final_credit", req.FinalCredit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
