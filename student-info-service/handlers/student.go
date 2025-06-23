package handlers

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查学号是否已存在
	var existingStudent models.Student
	if err := h.db.Where("student_id = ?", req.StudentID).First(&existingStudent).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "学号已存在"})
		return
	}

	// 检查用户名是否已存在
	if err := h.db.Where("username = ?", req.Username).First(&existingStudent).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 创建学生
	student := models.Student{
		Username:  req.Username,
		StudentID: req.StudentID,
		Name:      req.Name,
		College:   req.College,
		Major:     req.Major,
		Class:     req.Class,
		Contact:   req.Contact,
		Email:     req.Email,
		Grade:     req.Grade,
		Status:    "active",
	}

	if err := h.db.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "学生创建成功",
		"student": student,
	})
}

// GetStudent 获取学生信息
func (h *StudentHandler) GetStudent(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var student models.Student
	err := h.db.Preload("User").Where("username = ?", username).First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, student)
}

// GetStudentByID 根据学号获取学生信息
func (h *StudentHandler) GetStudentByID(c *gin.Context) {
	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学号不能为空"})
		return
	}

	var student models.Student
	err := h.db.Preload("User").Where("student_id = ?", studentID).First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, student)
}

// UpdateStudent 更新学生信息
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var req models.StudentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var student models.Student
	if err := h.db.Where("username = ?", username).First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		}
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.College != "" {
		updates["college"] = req.College
	}
	if req.Major != "" {
		updates["major"] = req.Major
	}
	if req.Class != "" {
		updates["class"] = req.Class
	}
	if req.Contact != "" {
		updates["contact"] = req.Contact
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Grade != "" {
		updates["grade"] = req.Grade
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.db.Model(&student).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "学生信息更新成功"})
}

// DeleteStudent 删除学生
func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var student models.Student
	if err := h.db.Where("username = ?", username).First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "学生删除成功"})
}

// GetAllStudents 获取所有学生
func (h *StudentHandler) GetAllStudents(c *gin.Context) {
	var students []models.Student
	err := h.db.Preload("User").Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// GetStudentsByCollege 根据学院获取学生
func (h *StudentHandler) GetStudentsByCollege(c *gin.Context) {
	college := c.Param("college")
	if college == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学院不能为空"})
		return
	}

	var students []models.Student
	err := h.db.Preload("User").Where("college = ?", college).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// GetStudentsByMajor 根据专业获取学生
func (h *StudentHandler) GetStudentsByMajor(c *gin.Context) {
	major := c.Param("major")
	if major == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "专业不能为空"})
		return
	}

	var students []models.Student
	err := h.db.Preload("User").Where("major = ?", major).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// GetStudentsByClass 根据班级获取学生
func (h *StudentHandler) GetStudentsByClass(c *gin.Context) {
	class := c.Param("class")
	if class == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "班级不能为空"})
		return
	}

	var students []models.Student
	err := h.db.Preload("User").Where("class = ?", class).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// GetStudentsByStatus 根据状态获取学生
func (h *StudentHandler) GetStudentsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态不能为空"})
		return
	}

	var students []models.Student
	err := h.db.Preload("User").Where("status = ?", status).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// SearchStudents 搜索学生
func (h *StudentHandler) SearchStudents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	var students []models.Student
	err := h.db.Preload("User").
		Where("name LIKE ? OR student_id LIKE ? OR college LIKE ? OR major LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索学生失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}
