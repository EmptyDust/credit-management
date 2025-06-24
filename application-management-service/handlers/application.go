package handlers

import (
	"errors"
	"net/http"
	"time"

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
	ID   string `json:"id" gorm:"primaryKey;type:uuid;column:affair_id"`
	Name string `json:"name" gorm:"unique;not null;column:affair_name"`
}

func (Affair) TableName() string {
	return "affairs"
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{DB: db}
}

// BatchCreateApplications 批量创建申请（事务创建时调用）
func (h *ApplicationHandler) BatchCreateApplications(c *gin.Context) {
	var req models.BatchCreateApplicationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// 验证事务是否存在
	var affair Affair
	if err := h.DB.Where("affair_id = ?", req.AffairID).First(&affair).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Specified affair does not exist"})
		return
	}

	// 批量创建申请
	var applications []models.Application
	for _, studentID := range req.Participants {
		app := models.Application{
			AffairID:      req.AffairID,
			StudentNumber: studentID,
			Status:        "unsubmitted", // Initial status is unsubmitted
		}
		applications = append(applications, app)
	}

	if err := h.DB.Create(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch create applications: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Applications created successfully",
		"count":   len(applications),
		"applications": applications,
	})
}

// CreateApplication 创建申请
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// 1. Check if the affair type exists and get its name
	var affair Affair
	if err := h.DB.Where("affair_id = ?", req.AffairID).First(&affair).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Specified affair does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query affair"})
		return
	}

	// 2. Create the base application object
	app := models.Application{
		AffairID:      req.AffairID,
		StudentNumber: req.StudentNumber,
		Status:        "pending",
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
		case "student_entrepreneurship":
			var detail models.StudentEntrepreneurshipProjectCredit
			mapstructure.Decode(req.Details, &detail)
			detail.ApplicationID = app.ID
			detailErr = tx.Create(&detail).Error
		case "entrepreneurship_practice":
			var detail models.EntrepreneurshipPracticeCredit
			mapstructure.Decode(req.Details, &detail)
			detail.ApplicationID = app.ID
			detailErr = tx.Create(&detail).Error
		case "paper_patent":
			var detail models.PaperPatentCredit
			mapstructure.Decode(req.Details, &detail)
			detail.ApplicationID = app.ID
			detailErr = tx.Create(&detail).Error
		default:
			// If the affair has no specific detail table, we do nothing.
			// Or return an error if details are expected.
			// For now, we assume it's fine.
		}

		return detailErr
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// GetApplicationDetail 获取申请完整详情
func (h *ApplicationHandler) GetApplicationDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application ID cannot be empty"})
		return
	}

	// 获取基础申请信息
	var application models.Application
	if err := h.DB.Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query application: " + err.Error()})
		}
		return
	}

	// 获取事务信息
	var affair Affair
	h.DB.Where("affair_id = ?", application.AffairID).First(&affair)

	// 构建详情响应
	detail := models.ApplicationDetail{
		Application: application,
	}

	// 根据事务类型查询对应详情
	switch affair.Name {
	case "innovation_practice":
		var innovationDetail models.InnovationPracticeCredit
		if err := h.DB.Where("application_id = ?", id).First(&innovationDetail).Error; err == nil {
			detail.InnovationPracticeCredit = &innovationDetail
		}
	case "discipline_competition":
		var competitionDetail models.DisciplineCompetitionCredit
		if err := h.DB.Where("application_id = ?", id).First(&competitionDetail).Error; err == nil {
			detail.DisciplineCompetitionCredit = &competitionDetail
		}
	case "student_entrepreneurship":
		var entrepreneurshipDetail models.StudentEntrepreneurshipProjectCredit
		if err := h.DB.Where("application_id = ?", id).First(&entrepreneurshipDetail).Error; err == nil {
			detail.StudentEntrepreneurshipProjectCredit = &entrepreneurshipDetail
		}
	case "entrepreneurship_practice":
		var practiceDetail models.EntrepreneurshipPracticeCredit
		if err := h.DB.Where("application_id = ?", id).First(&practiceDetail).Error; err == nil {
			detail.EntrepreneurshipPracticeCredit = &practiceDetail
		}
	case "paper_patent":
		var paperDetail models.PaperPatentCredit
		if err := h.DB.Where("application_id = ?", id).First(&paperDetail).Error; err == nil {
			detail.PaperPatentCredit = &paperDetail
		}
	}

	c.JSON(http.StatusOK, detail)
}

// UpdateApplicationDetails 更新申请详情（学生编辑）
func (h *ApplicationHandler) UpdateApplicationDetails(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application ID cannot be empty"})
		return
	}

	// 权限校验：只能编辑自己的申请
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not provided"})
		return
	}

	var application models.Application
	if err := h.DB.Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query application: " + err.Error()})
		}
		return
	}

	if application.StudentNumber != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only edit your own applications"})
		return
	}

	if application.Status != "unsubmitted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only edit unsubmitted applications"})
		return
	}

	var req models.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// 更新申请基础信息
	application.AppliedCredits = req.AppliedCredits
	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	// 获取事务信息以确定详情类型
	var affair Affair
	h.DB.Where("affair_id = ?", application.AffairID).First(&affair)

	// 更新详情信息
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		switch affair.Name {
		case "innovation_practice":
			var detail models.InnovationPracticeCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				// 如果不存在则创建
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = id
				return tx.Create(&detail).Error
			} else {
				// 更新现有记录
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "discipline_competition":
			var detail models.DisciplineCompetitionCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = id
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "student_entrepreneurship":
			var detail models.StudentEntrepreneurshipProjectCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = id
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "entrepreneurship_practice":
			var detail models.EntrepreneurshipPracticeCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = id
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "paper_patent":
			var detail models.PaperPatentCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = id
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		default:
			return nil
		}
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application details: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application details updated successfully"})
}

// SubmitApplication 提交申请（状态变为待审核）
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application ID cannot be empty"})
		return
	}

	// 权限校验
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not provided"})
		return
	}

	var application models.Application
	if err := h.DB.Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query application: " + err.Error()})
		}
		return
	}

	if application.StudentNumber != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only submit your own applications"})
		return
	}

	if application.Status != "unsubmitted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only submit unsubmitted applications"})
		return
	}

	// 更新状态为待审核
	application.Status = "pending"
	application.SubmissionTime = time.Now()
	
	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application submitted successfully"})
}

// UpdateApplicationStatus 更新申请状态（教师审核）
func (h *ApplicationHandler) UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	reviewerID := c.GetHeader("X-User-Id")

	// 权限校验：只允许教师或管理员审核
	userType, _ := c.Get("user_type")
	if userType != "teacher" && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers or admins can review applications"})
		return
	}

	var req struct {
		Status          string  `json:"status" binding:"required"`
		ReviewComment   string  `json:"review_comment"`
		ApprovedCredits float64 `json:"approved_credits"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// 验证状态值
	validStatuses := []string{"approved", "rejected"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'approved' or 'rejected'"})
		return
	}

	var application models.Application
	if err := h.DB.Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query application: " + err.Error()})
		}
		return
	}

	// 只有待审核的申请才能被审核
	if application.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only review pending applications"})
		return
	}

	application.Status = req.Status
	application.ReviewComment = req.ReviewComment
	application.ApprovedCredits = req.ApprovedCredits
	application.ReviewTime = &time.Time{}
	*application.ReviewTime = time.Now()
	
	if reviewerID != "" {
		application.ReviewerID = reviewerID
	}

	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application status updated successfully"})
}

// GetUserApplications 获取指定用户的所有申请
func (h *ApplicationHandler) GetUserApplications(c *gin.Context) {
	studentNumber := c.Param("studentNumber")
	if studentNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student number"})
		return
	}

	var applications []models.Application
	if err := h.DB.Where("student_number = ?", studentNumber).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user applications"})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// GetAllApplications 获取所有申请
func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	var applications []models.Application
	if err := h.DB.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get all applications"})
		return
	}

	c.JSON(http.StatusOK, applications)
}
