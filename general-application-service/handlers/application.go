package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/general-application-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	db *gorm.DB
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

// CreateApplication 创建申请
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req models.ApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建申请记录
	application := models.Application{
		AffairID:       req.AffairID,
		StudentID:      req.StudentID,
		SubmitTime:     time.Now(),
		Status:         "pending",
		AppliedCredits: req.AppliedCredits,
	}

	if err := tx.Create(&application).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建申请失败: " + err.Error()})
		return
	}

	// 创建证明材料
	for _, material := range req.ProofMaterials {
		proofMaterial := models.ProofMaterial{
			ApplicationID: application.ID,
			AffairID:      material.AffairID,
			Content:       material.Content,
			FileName:      material.FileName,
			FileSize:      material.FileSize,
			FileType:      material.FileType,
		}
		if err := tx.Create(&proofMaterial).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建证明材料失败: " + err.Error()})
			return
		}
	}

	// 根据事项类型创建对应的学分申请细分记录
	if req.InnovationPractice != nil {
		innovationCredit := models.InnovationPracticeCredit{
			ApplicationID: application.ID,
			Company:       req.InnovationPractice.Company,
			ProjectID:     req.InnovationPractice.ProjectID,
			IssuingOrg:    req.InnovationPractice.IssuingOrg,
			Date:          req.InnovationPractice.Date,
			TotalHours:    req.InnovationPractice.TotalHours,
		}
		if err := tx.Create(&innovationCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建创新创业实践活动学分记录失败: " + err.Error()})
			return
		}
	}

	if req.DisciplineCompetition != nil {
		competitionCredit := models.DisciplineCompetitionCredit{
			ApplicationID:   application.ID,
			Level:           req.DisciplineCompetition.Level,
			CompetitionName: req.DisciplineCompetition.CompetitionName,
			AwardLevel:      req.DisciplineCompetition.AwardLevel,
			Ranking:         req.DisciplineCompetition.Ranking,
		}
		if err := tx.Create(&competitionCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学科竞赛学分记录失败: " + err.Error()})
			return
		}
	}

	if req.StudentEntrepreneurship != nil {
		entrepreneurshipCredit := models.StudentEntrepreneurshipCredit{
			ApplicationID:  application.ID,
			ProjectName:    req.StudentEntrepreneurship.ProjectName,
			ProjectLevel:   req.StudentEntrepreneurship.ProjectLevel,
			ProjectRanking: req.StudentEntrepreneurship.ProjectRanking,
		}
		if err := tx.Create(&entrepreneurshipCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建大学生创业项目学分记录失败: " + err.Error()})
			return
		}
	}

	if req.EntrepreneurshipPractice != nil {
		practiceCredit := models.EntrepreneurshipPracticeCredit{
			ApplicationID: application.ID,
			CompanyName:   req.EntrepreneurshipPractice.CompanyName,
			LegalPerson:   req.EntrepreneurshipPractice.LegalPerson,
			ShareRatio:    req.EntrepreneurshipPractice.ShareRatio,
		}
		if err := tx.Create(&practiceCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建创业实践项目学分记录失败: " + err.Error()})
			return
		}
	}

	if req.PaperPatent != nil {
		paperPatentCredit := models.PaperPatentCredit{
			ApplicationID: application.ID,
			Title:         req.PaperPatent.Title,
			Category:      req.PaperPatent.Category,
			Ranking:       req.PaperPatent.Ranking,
		}
		if err := tx.Create(&paperPatentCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建论文专利学分记录失败: " + err.Error()})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "申请创建成功",
		"application_id": application.ID,
	})
}

// GetApplication 获取单个申请详情
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	var application models.Application
	err = h.db.Preload("Affair").
		Preload("Student").
		Preload("Reviewer").
		Preload("ProofMaterials").
		Preload("InnovationPracticeCredit").
		Preload("DisciplineCompetitionCredit").
		Preload("StudentEntrepreneurshipCredit").
		Preload("EntrepreneurshipPracticeCredit").
		Preload("PaperPatentCredit").
		First(&application, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, application)
}

// UpdateApplication 更新申请
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	var req models.ApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查申请是否存在
	var existingApplication models.Application
	if err := h.db.First(&existingApplication, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		}
		return
	}

	// 只能更新待审核状态的申请
	if existingApplication.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能更新待审核状态的申请"})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新申请基本信息
	updates := map[string]interface{}{
		"affair_id":       req.AffairID,
		"applied_credits": req.AppliedCredits,
	}
	if err := tx.Model(&existingApplication).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新申请失败: " + err.Error()})
		return
	}

	// 删除旧的证明材料
	if err := tx.Where("application_id = ?", id).Delete(&models.ProofMaterial{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧证明材料失败: " + err.Error()})
		return
	}

	// 创建新的证明材料
	for _, material := range req.ProofMaterials {
		proofMaterial := models.ProofMaterial{
			ApplicationID: id,
			AffairID:      material.AffairID,
			Content:       material.Content,
			FileName:      material.FileName,
			FileSize:      material.FileSize,
			FileType:      material.FileType,
		}
		if err := tx.Create(&proofMaterial).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建证明材料失败: " + err.Error()})
			return
		}
	}

	// 删除旧的学分申请细分记录
	tx.Where("application_id = ?", id).Delete(&models.InnovationPracticeCredit{})
	tx.Where("application_id = ?", id).Delete(&models.DisciplineCompetitionCredit{})
	tx.Where("application_id = ?", id).Delete(&models.StudentEntrepreneurshipCredit{})
	tx.Where("application_id = ?", id).Delete(&models.EntrepreneurshipPracticeCredit{})
	tx.Where("application_id = ?", id).Delete(&models.PaperPatentCredit{})

	// 创建新的学分申请细分记录
	if req.InnovationPractice != nil {
		innovationCredit := models.InnovationPracticeCredit{
			ApplicationID: id,
			Company:       req.InnovationPractice.Company,
			ProjectID:     req.InnovationPractice.ProjectID,
			IssuingOrg:    req.InnovationPractice.IssuingOrg,
			Date:          req.InnovationPractice.Date,
			TotalHours:    req.InnovationPractice.TotalHours,
		}
		if err := tx.Create(&innovationCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建创新创业实践活动学分记录失败: " + err.Error()})
			return
		}
	}

	if req.DisciplineCompetition != nil {
		competitionCredit := models.DisciplineCompetitionCredit{
			ApplicationID:   id,
			Level:           req.DisciplineCompetition.Level,
			CompetitionName: req.DisciplineCompetition.CompetitionName,
			AwardLevel:      req.DisciplineCompetition.AwardLevel,
			Ranking:         req.DisciplineCompetition.Ranking,
		}
		if err := tx.Create(&competitionCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学科竞赛学分记录失败: " + err.Error()})
			return
		}
	}

	if req.StudentEntrepreneurship != nil {
		entrepreneurshipCredit := models.StudentEntrepreneurshipCredit{
			ApplicationID:  id,
			ProjectName:    req.StudentEntrepreneurship.ProjectName,
			ProjectLevel:   req.StudentEntrepreneurship.ProjectLevel,
			ProjectRanking: req.StudentEntrepreneurship.ProjectRanking,
		}
		if err := tx.Create(&entrepreneurshipCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建大学生创业项目学分记录失败: " + err.Error()})
			return
		}
	}

	if req.EntrepreneurshipPractice != nil {
		practiceCredit := models.EntrepreneurshipPracticeCredit{
			ApplicationID: id,
			CompanyName:   req.EntrepreneurshipPractice.CompanyName,
			LegalPerson:   req.EntrepreneurshipPractice.LegalPerson,
			ShareRatio:    req.EntrepreneurshipPractice.ShareRatio,
		}
		if err := tx.Create(&practiceCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建创业实践项目学分记录失败: " + err.Error()})
			return
		}
	}

	if req.PaperPatent != nil {
		paperPatentCredit := models.PaperPatentCredit{
			ApplicationID: id,
			Title:         req.PaperPatent.Title,
			Category:      req.PaperPatent.Category,
			Ranking:       req.PaperPatent.Ranking,
		}
		if err := tx.Create(&paperPatentCredit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建论文专利学分记录失败: " + err.Error()})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请更新成功"})
}

// ReviewApplication 审核申请
func (h *ApplicationHandler) ReviewApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	var req models.ApplicationReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查申请是否存在
	var application models.Application
	if err := h.db.First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		}
		return
	}

	// 只能审核待审核状态的申请
	if application.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能审核待审核状态的申请"})
		return
	}

	// 更新申请状态
	updates := map[string]interface{}{
		"status":           req.Status,
		"reviewer_id":      req.ReviewerID,
		"review_comment":   req.ReviewComment,
		"approved_credits": req.ApprovedCredits,
	}
	if err := h.db.Model(&application).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请审核成功"})
}

// DeleteApplication 删除申请
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的申请ID"})
		return
	}

	// 检查申请是否存在
	var application models.Application
	if err := h.db.First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		}
		return
	}

	// 只能删除待审核状态的申请
	if application.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能删除待审核状态的申请"})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除证明材料
	if err := tx.Where("application_id = ?", id).Delete(&models.ProofMaterial{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除证明材料失败: " + err.Error()})
		return
	}

	// 删除学分申请细分记录
	tx.Where("application_id = ?", id).Delete(&models.InnovationPracticeCredit{})
	tx.Where("application_id = ?", id).Delete(&models.DisciplineCompetitionCredit{})
	tx.Where("application_id = ?", id).Delete(&models.StudentEntrepreneurshipCredit{})
	tx.Where("application_id = ?", id).Delete(&models.EntrepreneurshipPracticeCredit{})
	tx.Where("application_id = ?", id).Delete(&models.PaperPatentCredit{})

	// 删除申请
	if err := tx.Delete(&application).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除申请失败: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "申请删除成功"})
}

// GetAllApplications 获取所有申请
func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	var applications []models.Application
	err := h.db.Preload("Affair").
		Preload("Student").
		Preload("Reviewer").
		Preload("ProofMaterials").
		Preload("InnovationPracticeCredit").
		Preload("DisciplineCompetitionCredit").
		Preload("StudentEntrepreneurshipCredit").
		Preload("EntrepreneurshipPracticeCredit").
		Preload("PaperPatentCredit").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// GetApplicationsByStudent 根据学生ID获取申请
func (h *ApplicationHandler) GetApplicationsByStudent(c *gin.Context) {
	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学生ID不能为空"})
		return
	}

	var applications []models.Application
	err := h.db.Where("student_id = ?", studentID).
		Preload("Affair").
		Preload("Student").
		Preload("Reviewer").
		Preload("ProofMaterials").
		Preload("InnovationPracticeCredit").
		Preload("DisciplineCompetitionCredit").
		Preload("StudentEntrepreneurshipCredit").
		Preload("EntrepreneurshipPracticeCredit").
		Preload("PaperPatentCredit").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}

// GetApplicationsByStatus 根据状态获取申请
func (h *ApplicationHandler) GetApplicationsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态不能为空"})
		return
	}

	var applications []models.Application
	err := h.db.Where("status = ?", status).
		Preload("Affair").
		Preload("Student").
		Preload("Reviewer").
		Preload("ProofMaterials").
		Preload("InnovationPracticeCredit").
		Preload("DisciplineCompetitionCredit").
		Preload("StudentEntrepreneurshipCredit").
		Preload("EntrepreneurshipPracticeCredit").
		Preload("PaperPatentCredit").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询申请失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, applications)
}
