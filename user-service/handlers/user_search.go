package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"credit-management/user-service/models"
	"credit-management/user-service/utils"

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
		utils.SendBadRequest(c, "请求参数错误")
		return
	}

	// 验证用户类型
	validator := utils.NewValidator()
	if err := validator.ValidateUserType(req.UserType); err != nil {
		utils.SendBadRequest(c, "无效的用户类型")
		return
	}

	// 验证分页参数
	page, pageSize, _ := validator.ValidatePagination(strconv.Itoa(req.Page), strconv.Itoa(req.PageSize))
	req.Page = page
	req.PageSize = pageSize

	currentUserRole := utils.GetCurrentUserRole(c)
	if currentUserRole == "" {
		utils.SendUnauthorized(c)
		return
	}

	// 权限检查
	if utils.IsStudent(currentUserRole) {
		if req.UserType == "teacher" {
			utils.SendForbidden(c, "权限不足")
			return
		}
	}

	// 确定视图名称
	var viewName string
	switch req.UserType {
	case "student":
		// if utils.IsAdmin(currentUserRole) {
		viewName = "student_complete_info"
		// } else if utils.IsTeacher(currentUserRole) {
		// 	viewName = "student_teacher_view"
		// } else {
		// 	viewName = "student_student_view"
		// }
	case "teacher":
		// if utils.IsAdmin(currentUserRole) {
		viewName = "teacher_admin_view"
		// } else {
		// 	utils.SendForbidden(c, "权限不足")
		// 	return
		// }
	default:
		utils.SendBadRequest(c, "无效的用户类型")
		return
	}

	// 构建查询
	query := h.db.Table(viewName)

	// 搜索条件
	if req.Query != "" {
		if isUUID(req.Query) {
			query = query.Where("id = ?", req.Query)
		} else {
			query = query.Where("(username ILIKE ? OR real_name ILIKE ? OR email ILIKE ?)",
				"%"+req.Query+"%", "%"+req.Query+"%", "%"+req.Query+"%")
		}
	}

	// 根据用户类型添加特定过滤条件
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
		if req.Status != "" && utils.IsTeacherOrAdmin(currentUserRole) {
			query = query.Where("status = ?", req.Status)
		}
	case "teacher":
		if req.Department != "" {
			query = query.Where("department = ?", req.Department)
		}
		if req.Title != "" {
			query = query.Where("title = ?", req.Title)
		}
		if req.Status != "" && utils.IsAdmin(currentUserRole) {
			query = query.Where("status = ?", req.Status)
		}
	}

	var total int64
	query.Count(&total)

	// 获取用户列表
	offset := (req.Page - 1) * req.PageSize
	var users []map[string]interface{}
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
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

	utils.SendSuccessResponse(c, response)
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

	utils.SendSuccessResponse(c, stats)
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

	utils.SendSuccessResponse(c, stats)
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
