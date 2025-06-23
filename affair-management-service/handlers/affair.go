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
	var req models.AffairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	affair := models.Affair{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		MaxCredits:  req.MaxCredits,
		Status:      "active",
	}

	if err := h.db.Create(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "事项创建成功",
		"affair_id": affair.ID,
	})
}

// GetAffair 获取单个事项
func (h *AffairHandler) GetAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	var affair models.Affair
	err = h.db.First(&affair, id).Error
	if err != nil {
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

	var req models.AffairUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
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

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["affair_name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.MaxCredits > 0 {
		updates["max_credits"] = req.MaxCredits
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.db.Model(&affair).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "事项更新成功"})
}

// DeleteAffair 删除事项
func (h *AffairHandler) DeleteAffair(c *gin.Context) {
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

	if err := h.db.Delete(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "事项删除成功"})
}

// GetAllAffairs 获取所有事项
func (h *AffairHandler) GetAllAffairs(c *gin.Context) {
	var affairs []models.Affair
	err := h.db.Find(&affairs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}

// GetAffairsByCategory 根据类别获取事项
func (h *AffairHandler) GetAffairsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "类别不能为空"})
		return
	}

	var affairs []models.Affair
	err := h.db.Where("category = ?", category).Find(&affairs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}

// GetActiveAffairs 获取活跃事项
func (h *AffairHandler) GetActiveAffairs(c *gin.Context) {
	var affairs []models.Affair
	err := h.db.Where("status = ?", "active").Find(&affairs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}

// AddStudentToAffair 为学生添加事项
func (h *AffairHandler) AddStudentToAffair(c *gin.Context) {
	var req models.AffairStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	affairStudent := models.AffairStudent{
		AffairID:          req.AffairID,
		StudentID:         req.StudentID,
		IsMainResponsible: req.IsMainResponsible,
	}

	if err := h.db.Create(&affairStudent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加学生到事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "学生已添加到事项"})
}

// RemoveStudentFromAffair 从事项中移除学生
func (h *AffairHandler) RemoveStudentFromAffair(c *gin.Context) {
	affairID, err := strconv.Atoi(c.Param("affairID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学生ID不能为空"})
		return
	}

	var affairStudent models.AffairStudent
	if err := h.db.Where("affair_id = ? AND student_id = ?", affairID, studentID).First(&affairStudent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不在该事项中"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询关系失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&affairStudent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "学生已从事项中移除"})
}

// GetStudentsByAffair 获取事项下的所有学生
func (h *AffairHandler) GetStudentsByAffair(c *gin.Context) {
	affairID, err := strconv.Atoi(c.Param("affairID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	var affairStudents []models.AffairStudent
	err = h.db.Where("affair_id = ?", affairID).
		Preload("Student").
		Find(&affairStudents).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairStudents)
}

// GetAffairsByStudent 获取学生参与的所有事项
func (h *AffairHandler) GetAffairsByStudent(c *gin.Context) {
	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学生ID不能为空"})
		return
	}

	var affairStudents []models.AffairStudent
	err := h.db.Where("student_id = ?", studentID).
		Preload("Affair").
		Find(&affairStudents).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairStudents)
}
