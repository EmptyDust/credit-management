package handlers

import (
	"net/http"
	"strconv"

	"credit-management/affair-management-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AffairHandler struct {
	db *gorm.DB
}

func NewAffairHandler(db *gorm.DB) *AffairHandler {
	return &AffairHandler{db: db}
}

// CreateAffair 创建事项
func (h *AffairHandler) CreateAffair(c *gin.Context) {
	var affair models.Affair
	if err := c.ShouldBindJSON(&affair); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.db.Create(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, affair)
}

// GetAffair 获取单个事项
func (h *AffairHandler) GetAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	var affair models.Affair
	if err := h.db.First(&affair, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "事项不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, affair)
}

// UpdateAffair 更新事项
func (h *AffairHandler) UpdateAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	var affair models.Affair
	if err := h.db.First(&affair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "事项不存在"})
		return
	}

	if err := c.ShouldBindJSON(&affair); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.db.Save(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affair)
}

// DeleteAffair 删除事项
func (h *AffairHandler) DeleteAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	if err := h.db.Delete(&models.Affair{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "事项删除成功"})
}

// GetAllAffairs 获取所有事项
func (h *AffairHandler) GetAllAffairs(c *gin.Context) {
	var affairs []models.Affair
	if err := h.db.Find(&affairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询所有事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}
