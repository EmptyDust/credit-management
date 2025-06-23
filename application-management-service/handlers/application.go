package handlers

import (
	"errors"
	"net/http"

	"credit-management/application-management-service/models"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	DB *gorm.DB
}

// Affair represents the affair model from the affair-management-service
// We define it here to avoid a direct dependency, assuming a shared database.
type Affair struct {
	ID   int    `json:"id" gorm:"primaryKey;column:affair_id"`
	Name string `json:"name" gorm:"unique;not null;column:affair_name"`
}

func (Affair) TableName() string {
	return "affairs"
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{DB: db}
}

// CreateApplication 创建申请
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
		return
	}

	// 1. Check if the affair type exists and get its name
	var affair Affair
	if err := h.DB.First(&affair, req.AffairID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "指定的事项不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败"})
		return
	}

	// 2. Create the base application object
	app := models.Application{
		AffairID:      req.AffairID,
		StudentNumber: req.StudentNumber,
		Status:        "待审核",
	}
	if val, ok := req.Details["applied_credits"].(float64); ok {
		app.AppliedCredits = val
	}

	// 3. Use a transaction to save the application and its details
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&app).Error; err != nil {
			return err
		}

		var detailErr error
		// Use the affair name to determine which detail table to use
		switch affair.Name {
		case "innovation_practice":
			var detail models.InnovationPracticeCredit
			mapstructure.Decode(req.Details, &detail)
			detail.ApplicationID = app.ID
			detailErr = tx.Create(&detail).Error
		case "discipline_competition":
			var detail models.DisciplineCompetitionCredit
			mapstructure.Decode(req.Details, &detail)
			detail.ApplicationID = app.ID
			detailErr = tx.Create(&detail).Error
		// ... add other cases here ...
		default:
			// If the affair has no specific detail table, we do nothing.
			// Or return an error if details are expected.
			// For now, we assume it's fine.
		}

		return detailErr
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// GetApplication 获取单个申请的详细信息
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")

	var application models.Application
	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	// Here you would typically join with the affair table to get the affair name
	// and then switch on the name to query the correct details table.
	// This part is left as an exercise as it requires a proper affair lookup.

	c.JSON(http.StatusOK, gin.H{
		"application": application,
		"details":     "details_lookup_not_implemented",
	})
}

// UpdateApplicationStatus 更新申请状态
func (h *ApplicationHandler) UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	// Reviewer ID would come from auth middleware
	// reviewerID, _ := c.Get("userID")

	var req struct {
		Status          string  `json:"status" binding:"required"`
		ReviewComment   string  `json:"review_comment"`
		ApprovedCredits float64 `json:"approved_credits"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
		return
	}

	var application models.Application
	if err := h.DB.First(&application, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	application.Status = req.Status
	application.ReviewComment = req.ReviewComment
	application.ApprovedCredits = req.ApprovedCredits
	// application.ReviewerID = reviewerID.(uint)

	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新申请状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请状态更新成功"})
}

// GetUserApplications 获取指定用户的所有申请
func (h *ApplicationHandler) GetUserApplications(c *gin.Context) {
	studentNumber := c.Param("studentNumber")
	if studentNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的学号"})
		return
	}

	var applications []models.Application
	if err := h.DB.Where("student_number = ?", studentNumber).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户申请列表失败"})
		return
	}

	c.JSON(http.StatusOK, applications)
}
