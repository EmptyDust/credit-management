package handlers

import (
	"net/http"
	"regexp"
	"time"

	"credit-management/user-service/models"

	"github.com/gin-gonic/gin"
)

type SearchUsersRequest struct {
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100"`
	Query      string `form:"query"`
	UserType   string `form:"user_type" binding:"required,oneof=student teacher"`
	College    string `form:"college"`
	Major      string `form:"major"`
	Class      string `form:"class"`
	Grade      string `form:"grade"`
	Department string `form:"department"`
	Title      string `form:"title"`
	Status     string `form:"status"`
}

func isUUID(str string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(str)
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	var req SearchUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	currentUserRole := getCurrentUserRole(c)
	if currentUserRole == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问", "data": nil})
		return
	}

	// 根据用户类型和权限确定使用的视图
	var viewName string
	switch req.UserType {
	case "student":
		switch currentUserRole {
		case "admin":
			viewName = "student_complete_info"
		case "teacher":
			viewName = "student_detail_info"
		case "student":
			viewName = "student_basic_info"
		default:
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
			return
		}
	case "teacher":
		switch currentUserRole {
		case "admin":
			viewName = "teacher_complete_info"
		case "teacher":
			viewName = "teacher_basic_info"
		case "student":
			viewName = "teacher_basic_info"
		default:
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户类型", "data": nil})
		return
	}

	query := h.db.Table(viewName)

	if req.Query != "" {
		if isUUID(req.Query) {
			query = query.Where("user_id = ?", req.Query)
		} else {
			searchQuery := "%" + req.Query + "%"
			query = query.Where(
				"username LIKE ? OR real_name LIKE ? OR student_id LIKE ?",
				searchQuery, searchQuery, searchQuery,
			)
		}
	}

	// 根据视图类型添加不同的筛选条件
	switch req.UserType {
	case "student":
		if req.College != "" {
			query = query.Where("college = ?", req.College)
		}
		if req.Major != "" {
			query = query.Where("major = ?", req.Major)
		}
		if req.Class != "" {
			query = query.Where("class = ?", req.Class)
		}
		if req.Grade != "" {
			query = query.Where("grade = ?", req.Grade)
		}
		if req.Status != "" && (currentUserRole == "admin" || currentUserRole == "teacher") {
			query = query.Where("status = ?", req.Status)
		}
	case "teacher":
		if req.Department != "" {
			query = query.Where("department = ?", req.Department)
		}
		if req.Title != "" {
			query = query.Where("title = ?", req.Title)
		}
		if req.Status != "" && currentUserRole == "admin" {
			query = query.Where("status = ?", req.Status)
		}
	}

	var total int64
	query.Count(&total)

	// 获取用户列表
	offset := (req.Page - 1) * req.PageSize
	var users []map[string]interface{}
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索用户失败", "data": nil})
		return
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	response := models.ViewBasedSearchResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int(totalPages),
		ViewType:   viewName,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

func (h *UserHandler) GetUserStats(c *gin.Context) {
	var stats models.UserStats

	h.db.Model(&models.User{}).Count(&stats.TotalUsers)

	h.db.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)

	h.db.Model(&models.User{}).Where("status = ?", "suspended").Count(&stats.SuspendedUsers)

	h.db.Model(&models.User{}).Where("user_type = ?", "student").Count(&stats.StudentUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "teacher").Count(&stats.TeacherUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "admin").Count(&stats.AdminUsers)

	today := time.Now().Truncate(24 * time.Hour)
	h.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday)

	weekStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday()))
	h.db.Model(&models.User{}).Where("created_at >= ?", weekStart).Count(&stats.NewUsersWeek)

	monthStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	h.db.Model(&models.User{}).Where("created_at >= ?", monthStart).Count(&stats.NewUsersMonth)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

func (h *UserHandler) GetStudentStats(c *gin.Context) {
	var stats models.StudentStats

	h.db.Model(&models.User{}).Where("user_type = ?", "student").Count(&stats.TotalStudents)

	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "student", "active").Count(&stats.ActiveStudents)

	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "student", "graduated").Count(&stats.GraduatedStudents)

	stats.StudentsByCollege = make(map[string]int64)
	var collegeStats []struct {
		College string
		Count   int64
	}
	h.db.Model(&models.User{}).
		Select("college, count(*) as count").
		Where("user_type = ? AND college IS NOT NULL", "student").
		Group("college").
		Find(&collegeStats)

	for _, stat := range collegeStats {
		stats.StudentsByCollege[stat.College] = stat.Count
	}

	stats.StudentsByMajor = make(map[string]int64)
	var majorStats []struct {
		Major string
		Count int64
	}
	h.db.Model(&models.User{}).
		Select("major, count(*) as count").
		Where("user_type = ? AND major IS NOT NULL", "student").
		Group("major").
		Find(&majorStats)

	for _, stat := range majorStats {
		stats.StudentsByMajor[stat.Major] = stat.Count
	}

	stats.StudentsByGrade = make(map[string]int64)
	var gradeStats []struct {
		Grade string
		Count int64
	}
	h.db.Model(&models.User{}).
		Select("grade, count(*) as count").
		Where("user_type = ? AND grade IS NOT NULL", "student").
		Group("grade").
		Find(&gradeStats)

	for _, stat := range gradeStats {
		stats.StudentsByGrade[stat.Grade] = stat.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

func (h *UserHandler) GetTeacherStats(c *gin.Context) {
	var stats models.TeacherStats

	h.db.Model(&models.User{}).Where("user_type = ?", "teacher").Count(&stats.TotalTeachers)

	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "teacher", "active").Count(&stats.ActiveTeachers)

	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "teacher", "retired").Count(&stats.RetiredTeachers)

	stats.TeachersByDepartment = make(map[string]int64)
	var deptStats []struct {
		Department string
		Count      int64
	}
	h.db.Model(&models.User{}).
		Select("department, count(*) as count").
		Where("user_type = ? AND department IS NOT NULL", "teacher").
		Group("department").
		Find(&deptStats)

	for _, stat := range deptStats {
		stats.TeachersByDepartment[stat.Department] = stat.Count
	}

	stats.TeachersByTitle = make(map[string]int64)
	var titleStats []struct {
		Title string
		Count int64
	}
	h.db.Model(&models.User{}).
		Select("title, count(*) as count").
		Where("user_type = ? AND title IS NOT NULL", "teacher").
		Group("title").
		Find(&titleStats)

	for _, stat := range titleStats {
		stats.TeachersByTitle[stat.Title] = stat.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}
