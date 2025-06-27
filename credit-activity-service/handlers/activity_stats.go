package handlers

import (
	"net/http"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
)

func (h *ActivityHandler) GetActivityStats(c *gin.Context) {
	var stats models.ActivityStats

	// 统计各种状态的活动数量
	h.db.Model(&models.CreditActivity{}).Count(&stats.TotalActivities)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusDraft).Count(&stats.DraftCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusPendingReview).Count(&stats.PendingCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusApproved).Count(&stats.ApprovedCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusRejected).Count(&stats.RejectedCount)

	// 统计参与者总数
	h.db.Model(&models.ActivityParticipant{}).Count(&stats.TotalParticipants)

	// 统计总学分
	h.db.Model(&models.ActivityParticipant{}).Select("COALESCE(SUM(credits), 0)").Scan(&stats.TotalCredits)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

func (h *ActivityHandler) GetActivityCategories(c *gin.Context) {
	categories := models.GetActivityCategories()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"categories":  categories,
			"count":       len(categories),
			"description": "活动类别列表",
		},
	})
}

func (h *ActivityHandler) GetActivityTemplates(c *gin.Context) {
	templates := []gin.H{
		{
			"name":        "创新创业实践活动",
			"category":    models.CategoryInnovation,
			"title":       "创新创业实践活动",
			"description": "参与创新创业项目，提升创新能力和实践技能",
		},
		{
			"name":        "学科竞赛",
			"category":    models.CategoryCompetition,
			"title":       "学科竞赛活动",
			"description": "参加各类学科竞赛，提升专业能力和竞争意识",
		},
		{
			"name":        "大学生创业项目",
			"category":    models.CategoryEntrepreneurship,
			"title":       "大学生创业项目",
			"description": "参与大学生创业项目，培养创业精神和实践能力",
		},
		{
			"name":        "创业实践项目",
			"category":    models.CategoryPractice,
			"title":       "创业实践项目",
			"description": "参与创业实践项目，积累创业经验和实践技能",
		},
		{
			"name":        "论文专利",
			"category":    models.CategoryPaperPatent,
			"title":       "论文专利活动",
			"description": "发表论文或申请专利，提升学术研究能力",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    templates,
	})
}

func (h *ActivityHandler) GetActivityReport(c *gin.Context) {
	reportType := c.DefaultQuery("type", "monthly")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var start, end time.Time
	var err error

	// 解析日期范围
	if startDate != "" && endDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误",
				"data":    nil,
			})
			return
		}
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误",
				"data":    nil,
			})
			return
		}
	} else {
		end = time.Now()
		start = end.AddDate(0, 0, -30)
	}

	var report interface{}

	switch reportType {
	case "monthly":
		report = h.generateMonthlyReport(start, end)
	case "category":
		report = h.generateCategoryReport(start, end)
	case "status":
		report = h.generateStatusReport(start, end)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的报表类型",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取报表成功",
		"data":    report,
	})
}

// generateMonthlyReport 生成月度报表
func (h *ActivityHandler) generateMonthlyReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	// 按月份统计活动数量
	rows, err := h.db.Raw(`
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as total_activities,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_activities,
			COUNT(CASE WHEN status = 'pending_review' THEN 1 END) as pending_activities,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_activities
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var month time.Time
			var total, approved, pending, rejected int64
			rows.Scan(&month, &total, &approved, &pending, &rejected)
			result = append(result, map[string]interface{}{
				"month":               month.Format("2006-01"),
				"total_activities":    total,
				"approved_activities": approved,
				"pending_activities":  pending,
				"rejected_activities": rejected,
			})
		}
	}

	return map[string]interface{}{
		"type":       "monthly",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
	}
}

// generateCategoryReport 生成分类报表
func (h *ActivityHandler) generateCategoryReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			category,
			COUNT(*) as total_activities,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_activities,
			AVG(EXTRACT(EPOCH FROM (end_date - start_date))/86400) as avg_duration_days
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY category
		ORDER BY total_activities DESC
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var category string
			var total, approved int64
			var avgDuration float64
			rows.Scan(&category, &total, &approved, &avgDuration)
			result = append(result, map[string]interface{}{
				"category":            category,
				"total_activities":    total,
				"approved_activities": approved,
				"avg_duration_days":   avgDuration,
			})
		}
	}

	return map[string]interface{}{
		"type":       "category",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
	}
}

// generateStatusReport 生成状态报表
func (h *ActivityHandler) generateStatusReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			status,
			COUNT(*) as count,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as recent_count
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY status
		ORDER BY count DESC
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count, recentCount int64
			rows.Scan(&status, &count, &recentCount)
			result = append(result, map[string]interface{}{
				"status":       status,
				"count":        count,
				"recent_count": recentCount,
			})
		}
	}

	return map[string]interface{}{
		"type":       "status",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
	}
}
