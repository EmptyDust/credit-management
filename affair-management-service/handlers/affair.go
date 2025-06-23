package handlers

import (
	"net/http"

	"affair-management-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AffairHandler struct {
	DB *gorm.DB
}

func NewAffairHandler(db *gorm.DB) *AffairHandler {
	return &AffairHandler{DB: db}
}

// CreateAffair 创建事项
func (h *AffairHandler) CreateAffair(c *gin.Context) {
	var affair models.Affair
	if err := c.ShouldBindJSON(&affair); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, affair)
}

// GetAffairs 获取事项列表
func (h *AffairHandler) GetAffairs(c *gin.Context) {
	var affairs []models.Affair
	if err := h.DB.Find(&affairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}

// GetAffair 获取单个事项
func (h *AffairHandler) GetAffair(c *gin.Context) {
	id := c.Param("id")
	var affair models.Affair

	if err := h.DB.First(&affair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Affair not found"})
		return
	}

	c.JSON(http.StatusOK, affair)
}

// UpdateAffair 更新事项
func (h *AffairHandler) UpdateAffair(c *gin.Context) {
	id := c.Param("id")
	var affair models.Affair

	if err := h.DB.First(&affair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Affair not found"})
		return
	}

	var updateData models.Affair
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&affair).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, affair)
}

// DeleteAffair 删除事项
func (h *AffairHandler) DeleteAffair(c *gin.Context) {
	id := c.Param("id")
	var affair models.Affair

	if err := h.DB.First(&affair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Affair not found"})
		return
	}

	if err := h.DB.Delete(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Affair deleted successfully"})
}
