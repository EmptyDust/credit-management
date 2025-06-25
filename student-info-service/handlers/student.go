package handlers

import (
	"fmt"
	"net/http"

	"credit-management/student-info-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudentHandler struct {
	db *gorm.DB
}

func NewStudentHandler(db *gorm.DB) *StudentHandler {
	return &StudentHandler{db: db}
}

// CreateStudent 创建学生
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var req models.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	// 检查用户名是否已存在
	var existingStudent models.Student
	if err := h.db.Where("username = ?", req.Username).First(&existingStudent).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		return
	}

	// 创建学生
	student := models.Student{
		Username:  req.Username,
		Name:      req.Name,
		College:   req.College,
		Major:     req.Major,
		Class:     req.Class,
		Contact:   req.Contact,
		Email:     req.Email,
		Grade:     req.Grade,
		Status:    "active",
	}
	if req.StudentID != "" {
		student.StudentID = &req.StudentID
	} else {
		student.StudentID = nil
	}

	if err := h.db.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": student,
	})
}

// GetStudentByID 根据UUID获取学生信息
func (h *StudentHandler) GetStudentByID(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "学生ID不能为空", "data": nil})
		return
	}

	// 获取当前用户信息
	userType, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	var student models.Student
	err := h.db.Where("id = ?", studentID).First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "学生不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 根据用户类型返回不同的数据
	if userType == "admin" {
		// 管理员可以看到所有信息
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": student})
	} else if userType == "teacher" {
		// 教师可以看到所有学生信息（除敏感信息）
		teacherViewStudent := gin.H{
			"id":         student.ID,
			"username":   student.Username,
			"name":       student.Name,
			"student_id": student.StudentID,
			"college":    student.College,
			"major":      student.Major,
			"class":      student.Class,
			"contact":    student.Contact,
			"email":      student.Email,
			"grade":      student.Grade,
			"status":     student.Status,
			"created_at": student.CreatedAt,
			"updated_at": student.UpdatedAt,
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teacherViewStudent})
	} else if userType == "student" {
		// 检查是否是查看自己的信息
		if student.UserID == currentUserID {
			// 学生查看自己的信息，可以看到全部信息
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": student})
		} else {
			// 学生查看其他学生信息，只能看到基本信息
			basicStudent := gin.H{
				"id":         student.ID,
				"name":       student.Name,
				"student_id": student.StudentID,
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": basicStudent})
		}
	} else {
		// 其他用户类型无权访问
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
	}
}

// UpdateStudentByID updates a student's information by UUID
func (h *StudentHandler) UpdateStudentByID(c *gin.Context) {
	studentID := c.Param("id")
	var req models.StudentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	var student models.Student
	if err := h.db.Where("id = ?", studentID).First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "学生不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		}
		return
	}

	if err := h.db.Model(&student).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新学生失败: " + err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": student})
}

// DeleteStudentByID deletes a student by UUID
func (h *StudentHandler) DeleteStudentByID(c *gin.Context) {
	studentID := c.Param("id")
	if err := h.db.Where("id = ?", studentID).Delete(&models.Student{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除学生失败: " + err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// GetAllStudents 获取所有学生
func (h *StudentHandler) GetAllStudents(c *gin.Context) {
	// 获取当前用户类型
	userType, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	// 根据用户类型返回不同的数据
	if userType == "admin" {
		// 管理员可以看到所有信息
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
	} else if userType == "teacher" {
		// 教师可以看到所有学生信息（除敏感信息）
		var teacherViewStudents []gin.H
		for _, student := range students {
			teacherViewStudents = append(teacherViewStudents, gin.H{
				"id":         student.ID,
				"username":   student.Username,
				"name":       student.Name,
				"student_id": student.StudentID,
				"college":    student.College,
				"major":      student.Major,
				"class":      student.Class,
				"contact":    student.Contact,
				"email":      student.Email,
				"grade":      student.Grade,
				"status":     student.Status,
				"created_at": student.CreatedAt,
				"updated_at": student.UpdatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": teacherViewStudents})
	} else if userType == "student" {
		// 学生只能看到基本信息（UUID、姓名、学号）
		var basicStudents []gin.H
		for _, student := range students {
			basicStudents = append(basicStudents, gin.H{
				"id":         student.ID,
				"name":       student.Name,
				"student_id": student.StudentID,
			})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": basicStudents})
	} else {
		// 其他用户类型无权访问
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
	}
}

// GetStudentsByCollege 根据学院获取学生
func (h *StudentHandler) GetStudentsByCollege(c *gin.Context) {
	college := c.Param("college")
	if college == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "学院不能为空", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("college = ?", college).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"students": students, "count": len(students)}})
}

// GetStudentsByMajor 根据专业获取学生
func (h *StudentHandler) GetStudentsByMajor(c *gin.Context) {
	major := c.Param("major")
	if major == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "专业不能为空", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("major = ?", major).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
}

// GetStudentsByClass 根据班级获取学生
func (h *StudentHandler) GetStudentsByClass(c *gin.Context) {
	class := c.Param("class")
	if class == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "班级不能为空", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("class = ?", class).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
}

// GetStudentsByStatus 根据状态获取学生
func (h *StudentHandler) GetStudentsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态不能为空", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("status = ?", status).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
}

// SearchStudents 搜索学生
func (h *StudentHandler) SearchStudents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "搜索关键词不能为空", "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("name LIKE ? OR user_id LIKE ? OR college LIKE ? OR major LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
}

// GetStudentByUsername 根据用户名获取学生信息
func (h *StudentHandler) GetStudentByUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名不能为空", "data": nil})
		return
	}

	var student models.Student
	err := h.db.Where("username = ?", username).First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "学生不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []models.Student{student}})
}

// GetStudentByUserID 根据用户ID获取学生信息
func (h *StudentHandler) GetStudentByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	var student models.Student
	err := h.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该用户对应的学生信息不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": student})
}

// GetStudentsByUserIDs 批量获取学生信息
func (h *StudentHandler) GetStudentsByUserIDs(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	var students []models.Student
	err := h.db.Where("user_id IN ?", req.UserIDs).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": students})
}

// DeleteStudentByUserID 根据用户ID删除学生档案
func (h *StudentHandler) DeleteStudentByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID不能为空", "data": nil})
		return
	}

	// 先查找学生是否存在
	var student models.Student
	if err := h.db.Where("user_id = ?", userID).First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该用户对应的学生信息不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询学生失败: " + err.Error(), "data": nil})
		}
		return
	}

	// 软删除学生档案
	if err := h.db.Delete(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除学生档案失败: " + err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0, 
		"message": "success", 
		"data": gin.H{
			"message": fmt.Sprintf("学生档案删除成功，用户ID: %s", userID),
		},
	})
}
