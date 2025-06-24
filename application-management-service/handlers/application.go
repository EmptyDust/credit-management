package handlers

import (
	"errors"
	"net/http"
	"strconv"
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
	ID   int    `json:"id" gorm:"primaryKey;column:affair_id"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
		return
	}

	// 验证事务是否存在
	var affair Affair
	if err := h.DB.First(&affair, req.AffairID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "指定的事务不存在"})
		return
	}

	// 批量创建申请
	var applications []models.Application
	for _, studentID := range req.Participants {
		app := models.Application{
			AffairID:      req.AffairID,
			StudentNumber: studentID,
			Status:        "未提交", // 初始状态为未提交
		}
		applications = append(applications, app)
	}

	if err := h.DB.Create(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量创建申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "批量创建申请成功",
		"count":   len(applications),
		"applications": applications,
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// GetApplicationDetail 获取申请完整详情
func (h *ApplicationHandler) GetApplicationDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	// 获取基础申请信息
	var application models.Application
	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	// 获取事务信息
	var affair Affair
	h.DB.First(&affair, application.AffairID)

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
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	// 权限校验：只能编辑自己的申请
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	var application models.Application
	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	if application.StudentNumber != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能编辑自己的申请"})
		return
	}

	if application.Status != "未提交" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能编辑未提交的申请"})
		return
	}

	var req models.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
		return
	}

	// 更新申请基础信息
	application.AppliedCredits = req.AppliedCredits
	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新申请失败"})
		return
	}

	// 获取事务信息以确定详情类型
	var affair Affair
	h.DB.First(&affair, application.AffairID)

	// 更新详情信息
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		switch affair.Name {
		case "innovation_practice":
			var detail models.InnovationPracticeCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				// 如果不存在则创建
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = uint(id)
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
				detail.ApplicationID = uint(id)
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "student_entrepreneurship":
			var detail models.StudentEntrepreneurshipProjectCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = uint(id)
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "entrepreneurship_practice":
			var detail models.EntrepreneurshipPracticeCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = uint(id)
				return tx.Create(&detail).Error
			} else {
				mapstructure.Decode(req.Details, &detail)
				return tx.Save(&detail).Error
			}
		case "paper_patent":
			var detail models.PaperPatentCredit
			if err := tx.Where("application_id = ?", id).First(&detail).Error; err != nil {
				mapstructure.Decode(req.Details, &detail)
				detail.ApplicationID = uint(id)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新申请详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请详情更新成功"})
}

// SubmitApplication 提交申请（状态变为待审核）
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	// 权限校验
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	var application models.Application
	if err := h.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	if application.StudentNumber != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能提交自己的申请"})
		return
	}

	if application.Status != "未提交" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能提交未提交的申请"})
		return
	}

	// 更新状态为待审核
	application.Status = "待审核"
	application.SubmissionTime = time.Now()
	
	if err := h.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交申请失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请提交成功"})
}

// UpdateApplicationStatus 更新申请状态（教师审核）
func (h *ApplicationHandler) UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	reviewerID := c.GetHeader("X-User-Id")

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

	// 只有待审核的申请才能被审核
	if application.Status != "待审核" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能审核待审核的申请"})
		return
	}

	application.Status = req.Status
	application.ReviewComment = req.ReviewComment
	application.ApprovedCredits = req.ApprovedCredits
	application.ReviewTime = &time.Time{}
	*application.ReviewTime = time.Now()
	
	if reviewerID != "" {
		if reviewerIDUint, err := strconv.ParseUint(reviewerID, 10, 32); err == nil {
			application.ReviewerID = uint(reviewerIDUint)
		}
	}

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

// GetAllApplications 获取所有申请
func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	var applications []models.Application
	
	// Get query parameters for filtering
	status := c.Query("status")
	affairID := c.Query("affair_id")
	
	query := h.DB
	
	// Apply filters if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if affairID != "" {
		query = query.Where("affair_id = ?", affairID)
	}
	
	// Get applications (no Preload)
	if err := query.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取申请列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"total":        len(applications),
	})
}
