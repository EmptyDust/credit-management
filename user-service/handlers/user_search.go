package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/user-service/models"

	"github.com/gin-gonic/gin"
)

// SearchUsers 搜索用户（基于角色的权限控制）
func (h *UserHandler) SearchUsers(c *gin.Context) {
	// 获取当前用户角色
	currentUserRole := getCurrentUserRole(c)
	if currentUserRole == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
		return
	}

	var req models.SearchRequest

	// 从查询参数获取搜索条件
	req.Query = c.Query("query")
	req.UserType = c.Query("user_type")
	req.College = c.Query("college")
	req.Major = c.Query("major")
	req.Class = c.Query("class")
	req.Grade = c.Query("grade")
	req.Department = c.Query("department")
	req.Title = c.Query("title")
	req.Status = c.Query("status")

	// 从路径参数获取搜索条件（兼容旧的路由参数）
	if req.College == "" {
		req.College = c.Param("college")
	}
	if req.Major == "" {
		req.Major = c.Param("major")
	}
	if req.Class == "" {
		req.Class = c.Param("class")
	}
	if req.Status == "" {
		req.Status = c.Param("status")
	}
	if req.Department == "" {
		req.Department = c.Param("department")
	}
	if req.Title == "" {
		req.Title = c.Param("title")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	req.Page = page
	req.PageSize = pageSize

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 根据用户角色限制搜索范围
	switch currentUserRole {
	case "student":
		// 学生只能搜索学生和教师的基本信息
		if req.UserType == "" {
			// 如果没有指定用户类型，默认搜索学生
			req.UserType = "student"
		} else if req.UserType != "student" && req.UserType != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足，只能搜索学生和教师信息", "data": nil})
			return
		}
	case "teacher":
		// 教师可以搜索学生详细信息和其他教师基本信息
		if req.UserType == "" {
			// 如果没有指定用户类型，默认搜索学生
			req.UserType = "student"
		}
	case "admin":
		// 管理员可以搜索所有用户的所有信息
		// 不限制搜索范围
	default:
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
		return
	}

	// 构建查询
	query := h.db.Model(&models.User{})

	// 添加搜索条件
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		query = query.Where(
			"username LIKE ? OR email LIKE ? OR real_name LIKE ? OR phone LIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery,
		)
	}

	if req.UserType != "" {
		query = query.Where("user_type = ?", req.UserType)
	}

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

	if req.Department != "" {
		query = query.Where("department = ?", req.Department)
	}

	if req.Title != "" {
		query = query.Where("title = ?", req.Title)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取用户列表
	offset := (req.Page - 1) * req.PageSize
	var users []models.User
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索用户失败", "data": nil})
		return
	}

	// 根据角色转换响应
	var responses []interface{}
	for _, user := range users {
		response := h.convertToRoleBasedResponse(user, currentUserRole, false)
		responses = append(responses, response)
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	response := models.RoleBasedSearchResponse{
		Users:      responses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int(totalPages),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	var stats models.UserStats

	// 总用户数
	h.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// 活跃用户数
	h.db.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)

	// 暂停用户数
	h.db.Model(&models.User{}).Where("status = ?", "suspended").Count(&stats.SuspendedUsers)

	// 各类型用户数
	h.db.Model(&models.User{}).Where("user_type = ?", "student").Count(&stats.StudentUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "teacher").Count(&stats.TeacherUsers)
	h.db.Model(&models.User{}).Where("user_type = ?", "admin").Count(&stats.AdminUsers)

	// 今日新增用户
	today := time.Now().Truncate(24 * time.Hour)
	h.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday)

	// 本周新增用户
	weekStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday()))
	h.db.Model(&models.User{}).Where("created_at >= ?", weekStart).Count(&stats.NewUsersWeek)

	// 本月新增用户
	monthStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	h.db.Model(&models.User{}).Where("created_at >= ?", monthStart).Count(&stats.NewUsersMonth)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetStudentStats 获取学生统计信息
func (h *UserHandler) GetStudentStats(c *gin.Context) {
	var stats models.StudentStats

	// 总学生数
	h.db.Model(&models.User{}).Where("user_type = ?", "student").Count(&stats.TotalStudents)

	// 活跃学生数
	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "student", "active").Count(&stats.ActiveStudents)

	// 毕业学生数
	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "student", "graduated").Count(&stats.GraduatedStudents)

	// 按学院统计
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

	// 按专业统计
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

	// 按年级统计
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

// GetTeacherStats 获取教师统计信息
func (h *UserHandler) GetTeacherStats(c *gin.Context) {
	var stats models.TeacherStats

	// 总教师数
	h.db.Model(&models.User{}).Where("user_type = ?", "teacher").Count(&stats.TotalTeachers)

	// 活跃教师数
	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "teacher", "active").Count(&stats.ActiveTeachers)

	// 退休教师数
	h.db.Model(&models.User{}).Where("user_type = ? AND status = ?", "teacher", "retired").Count(&stats.RetiredTeachers)

	// 按院系统计
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

	// 按职称统计
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
