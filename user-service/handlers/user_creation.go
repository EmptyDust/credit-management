package handlers

import (
	"credit-management/user-service/models"
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) Register(c *gin.Context) {
	var req models.StudentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	if !models.ValidatePasswordComplexity(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码必须包含大小写字母和数字，且长度至少8位", "data": nil})
		return
	}

	if !models.ValidatePhoneFormat(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确", "data": nil})
		return
	}

	if !models.ValidateStudentIDFormat(req.StudentID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "学号格式不正确", "data": nil})
		return
	}

	if !models.ValidateGradeFormat(req.Grade) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年级格式不正确", "data": nil})
		return
	}

	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("student_id = ?", req.StudentID).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被删除的用户使用，请使用其他学号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	user := models.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Phone:     &req.Phone,
		RealName:  req.RealName,
		UserType:  "student",
		Status:    "active",
		StudentID: &req.StudentID,
		College:   &req.College,
		Major:     &req.Major,
		Class:     &req.Class,
		Grade:     &req.Grade,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	userResponse := h.convertToUserResponse(user)
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "学生注册成功",
			"user":    userResponse,
		},
	})
}

func (h *UserHandler) CreateTeacher(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证，无法操作", "data": nil})
		return
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok || claimsMap["user_type"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有管理员可以创建教师", "data": nil})
		return
	}

	var req models.TeacherRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	if !models.ValidatePasswordComplexity(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码必须包含大小写字母和数字，且长度至少8位", "data": nil})
		return
	}

	if !models.ValidatePhoneFormat(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确", "data": nil})
		return
	}

	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	user := models.User{
		Username:   req.Username,
		Password:   string(hashedPassword),
		Email:      req.Email,
		Phone:      &req.Phone,
		RealName:   req.RealName,
		UserType:   "teacher",
		Status:     "active",
		Department: &req.Department,
		Title:      &req.Title,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	userResponse := h.convertToUserResponse(user)
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "教师创建成功",
			"user":    userResponse,
		},
	})
}

// CreateStudent 管理员创建学生
func (h *UserHandler) CreateStudent(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证，无法操作", "data": nil})
		return
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok || claimsMap["user_type"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有管理员可以创建学生", "data": nil})
		return
	}

	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error(), "data": nil})
		return
	}

	if req.UserType != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能创建学生用户", "data": nil})
		return
	}

	if !models.ValidatePasswordComplexity(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码必须包含大小写字母和数字，且长度至少8位", "data": nil})
		return
	}

	if req.Phone != "" && !models.ValidatePhoneFormat(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确", "data": nil})
		return
	}

	if req.StudentID != "" && !models.ValidateStudentIDFormat(req.StudentID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "学号格式不正确", "data": nil})
		return
	}

	if req.Grade != "" && !models.ValidateGradeFormat(req.Grade) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年级格式不正确", "data": nil})
		return
	}

	if err := h.validateUserRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已被删除的用户使用，请选择其他用户名", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在", "data": nil})
		}
		return
	}

	if err := h.db.Unscoped().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被删除的用户使用，请使用其他邮箱", "data": nil})
		} else {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被使用", "data": nil})
		}
		return
	}

	if req.Phone != "" {
		if err := h.db.Unscoped().Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被删除的用户使用，请使用其他手机号", "data": nil})
			} else {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "手机号已被使用", "data": nil})
			}
			return
		}
	}

	if req.StudentID != "" {
		if err := h.db.Unscoped().Where("student_id = ?", req.StudentID).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被删除的用户使用，请使用其他学号", "data": nil})
			} else {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "学号已被使用", "data": nil})
			}
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败", "data": nil})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		RealName: req.RealName,
		UserType: "student",
		Status:   "active",
	}

	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.StudentID != "" {
		user.StudentID = &req.StudentID
	}
	if req.College != "" {
		user.College = &req.College
	}
	if req.Major != "" {
		user.Major = &req.Major
	}
	if req.Class != "" {
		user.Class = &req.Class
	}
	if req.Grade != "" {
		user.Grade = &req.Grade
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败，请稍后重试", "data": nil})
		return
	}

	userResponse := h.convertToUserResponse(user)
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "学生创建成功",
			"user":    userResponse,
		},
	})
}

func (h *UserHandler) validateUserRequest(req *models.UserRequest) error {
	if req.UserType != "student" && req.UserType != "teacher" {
		return fmt.Errorf("用户类型必须是student或teacher")
	}

	if req.UserType == "student" {
		if req.College == "" {
			return fmt.Errorf("学生必须指定学院")
		}
		if req.Major == "" {
			return fmt.Errorf("学生必须指定专业")
		}
		if req.Class == "" {
			return fmt.Errorf("学生必须指定班级")
		}
	}

	if req.UserType == "teacher" {
		if req.Department == "" {
			return fmt.Errorf("教师必须指定部门")
		}
		if req.Title == "" {
			return fmt.Errorf("教师必须指定职称")
		}
	}

	return nil
}

func (h *UserHandler) ImportUsers(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType := c.PostForm("user_type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请指定用户类型 (student/teacher)",
			"data":    nil,
		})
		return
	}

	if userType != "student" && userType != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户类型只能是 student 或 teacher",
			"data":    nil,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请上传文件",
			"data":    nil,
		})
		return
	}

	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小不能超过10MB",
			"data":    nil,
		})
		return
	}

	fileName := strings.ToLower(file.Filename)
	var records [][]string
	var parseError error

	if strings.HasSuffix(fileName, ".csv") {
		records, parseError = h.parseCSVFile(file)
	} else if strings.HasSuffix(fileName, ".xlsx") || strings.HasSuffix(fileName, ".xls") {
		records, parseError = h.parseExcelFile(file)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只支持CSV、XLSX、XLS文件格式",
			"data":    nil,
		})
		return
	}

	if parseError != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件解析失败: " + parseError.Error(),
			"data":    nil,
		})
		return
	}

	h.processImportData(c, records, userType, file.Filename)
}

func (h *UserHandler) parseCSVFile(file *multipart.FileHeader) ([][]string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1 // 允许变长记录

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV文件格式错误: %v", err)
	}

	return records, nil
}

func (h *UserHandler) parseExcelFile(file *multipart.FileHeader) ([][]string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	tempFile, err := os.CreateTemp("", "excel_import_*.xlsx")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	fileBytes := make([]byte, file.Size)
	_, err = src.Read(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %v", err)
	}
	tempFile.Close()

	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	firstSheet := sheets[0]
	rows, err := f.GetRows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %v", err)
	}

	var records [][]string
	for _, row := range rows {
		if len(row) > 0 {
			record := make([]string, 10)
			for i := 0; i < 10 && i < len(row); i++ {
				record[i] = strings.TrimSpace(row[i])
			}
			records = append(records, record)
		}
	}

	return records, nil
}

func (h *UserHandler) processImportData(c *gin.Context, records [][]string, userType string, fileName string) {
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件至少需要包含标题行和一行数据",
			"data":    nil,
		})
		return
	}

	if len(records) > 1001 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件最多支持1000行数据",
			"data":    nil,
		})
		return
	}

	var expectedHeaders []string
	if userType == "student" {
		expectedHeaders = []string{"username", "password", "email", "phone", "real_name", "student_id", "college", "major", "class", "grade"}
	} else {
		expectedHeaders = []string{"username", "password", "email", "phone", "real_name", "department", "title"}
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	missingHeaders := []string{}
	for _, expected := range expectedHeaders {
		if _, exists := headerMap[expected]; !exists {
			missingHeaders = append(missingHeaders, expected)
		}
	}

	if len(missingHeaders) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件缺少必需的列: " + strings.Join(missingHeaders, ", "),
			"data":    nil,
		})
		return
	}

	var users []models.UserRequest
	var errors []string

	for i, record := range records[1:] {
		rowNum := i + 2 // 从第2行开始计算

		// 检查记录长度
		if len(record) < len(headers) {
			errors = append(errors, fmt.Sprintf("第%d行: 列数不匹配", rowNum))
			continue
		}

		// 构建用户请求
		user := models.UserRequest{
			Username: strings.TrimSpace(record[headerMap["username"]]),
			Password: strings.TrimSpace(record[headerMap["password"]]),
			Email:    strings.TrimSpace(record[headerMap["email"]]),
			Phone:    strings.TrimSpace(record[headerMap["phone"]]),
			RealName: strings.TrimSpace(record[headerMap["real_name"]]),
			UserType: userType,
		}

		if userType == "student" {
			user.StudentID = strings.TrimSpace(record[headerMap["student_id"]])
			user.College = strings.TrimSpace(record[headerMap["college"]])
			user.Major = strings.TrimSpace(record[headerMap["major"]])
			user.Class = strings.TrimSpace(record[headerMap["class"]])
			user.Grade = strings.TrimSpace(record[headerMap["grade"]])
		} else {
			user.Department = strings.TrimSpace(record[headerMap["department"]])
			user.Title = strings.TrimSpace(record[headerMap["title"]])
		}

		if err := h.validateUserRequest(&user); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}

		users = append(users, user)
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "数据验证失败",
			"data": gin.H{
				"errors":       errors,
				"total_rows":   len(records) - 1,
				"valid_rows":   len(users),
				"invalid_rows": len(errors),
			},
		})
		return
	}

	var createdUsers []models.UserResponse
	var createErrors []string

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, userReq := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
		if err != nil {
			createErrors = append(createErrors, fmt.Sprintf("第%d个用户: 密码加密失败", i+1))
			continue
		}

		user := models.User{
			Username:     userReq.Username,
			Password:     string(hashedPassword),
			Email:        userReq.Email,
			Phone:        &userReq.Phone,
			RealName:     userReq.RealName,
			UserType:     userReq.UserType,
			Status:       "active",
			RegisterTime: time.Now(),
		}

		if userType == "student" {
			user.StudentID = &userReq.StudentID
			user.College = &userReq.College
			user.Major = &userReq.Major
			user.Class = &userReq.Class
			user.Grade = &userReq.Grade
		} else {
			user.Department = &userReq.Department
			user.Title = &userReq.Title
		}

		if err := tx.Create(&user).Error; err != nil {
			createErrors = append(createErrors, fmt.Sprintf("第%d个用户创建失败: %s", i+1, err.Error()))
			continue
		}

		response := h.convertToUserResponse(user)
		createdUsers = append(createdUsers, response)
	}

	// 如果有创建错误，回滚事务
	if len(createErrors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量创建用户失败",
			"data": gin.H{
				"errors":        createErrors,
				"created_count": 0,
				"total_count":   len(users),
				"created_users": []models.UserResponse{},
			},
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "批量导入成功",
		"data": gin.H{
			"created_count": len(createdUsers),
			"total_count":   len(users),
			"created_users": createdUsers,
			"user_type":     userType,
			"file_name":     fileName,
			"file_type":     filepath.Ext(fileName),
		},
	})
}

func (h *UserHandler) GetUserCSVTemplate(c *gin.Context) {
	userType := c.Query("user_type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请指定用户类型 (student/teacher)",
			"data":    nil,
		})
		return
	}

	if userType != "student" && userType != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户类型只能是 student 或 teacher",
			"data":    nil,
		})
		return
	}

	var template [][]string
	if userType == "student" {
		template = [][]string{
			{"username", "password", "email", "phone", "real_name", "student_id", "college", "major", "class", "grade"},
			{"student001", "Password123", "student001@example.com", "13800138001", "张三", "20240001", "计算机学院", "软件工程", "软工2401", "2024"},
			{"student002", "Password123", "student002@example.com", "13800138002", "李四", "20240002", "计算机学院", "软件工程", "软工2401", "2024"},
		}
	} else {
		template = [][]string{
			{"username", "password", "email", "phone", "real_name", "department", "title"},
			{"teacher001", "Password123", "teacher001@example.com", "13800138003", "王老师", "计算机学院", "副教授"},
			{"teacher002", "Password123", "teacher002@example.com", "13800138004", "赵老师", "计算机学院", "讲师"},
		}
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_template.csv", userType))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	for _, record := range template {
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "生成CSV模板失败",
				"data":    nil,
			})
			return
		}
	}
}

func (h *UserHandler) GetUserExcelTemplate(c *gin.Context) {
	userType := c.Query("user_type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请指定用户类型 (student/teacher)",
			"data":    nil,
		})
		return
	}

	if userType != "student" && userType != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户类型只能是 student 或 teacher",
			"data":    nil,
		})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "用户导入模板"
	f.SetSheetName("Sheet1", sheetName)

	var headers []string
	var examples [][]string

	if userType == "student" {
		headers = []string{"username", "password", "email", "phone", "real_name", "student_id", "college", "major", "class", "grade"}
		examples = [][]string{
			{"student001", "Password123", "student001@example.com", "13800138001", "张三", "20240001", "计算机学院", "软件工程", "软工2401", "2024"},
			{"student002", "Password123", "student002@example.com", "13800138002", "李四", "20240002", "计算机学院", "软件工程", "软工2401", "2024"},
		}
	} else {
		headers = []string{"username", "password", "email", "phone", "real_name", "department", "title"}
		examples = [][]string{
			{"teacher001", "Password123", "teacher001@example.com", "13800138003", "王老师", "计算机学院", "副教授"},
			{"teacher002", "Password123", "teacher002@example.com", "13800138004", "赵老师", "计算机学院", "讲师"},
		}
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, example := range examples {
		for j, value := range example {
			cell := fmt.Sprintf("%c%d", 'A'+j, i+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	if userType == "student" {
		f.SetColWidth(sheetName, "A", "A", 15) // username
		f.SetColWidth(sheetName, "B", "B", 15) // password
		f.SetColWidth(sheetName, "C", "C", 25) // email
		f.SetColWidth(sheetName, "D", "D", 15) // phone
		f.SetColWidth(sheetName, "E", "E", 15) // real_name
		f.SetColWidth(sheetName, "F", "F", 15) // student_id
		f.SetColWidth(sheetName, "G", "G", 20) // college
		f.SetColWidth(sheetName, "H", "H", 20) // major
		f.SetColWidth(sheetName, "I", "I", 15) // class
		f.SetColWidth(sheetName, "J", "J", 10) // grade
	} else {
		f.SetColWidth(sheetName, "A", "A", 15) // username
		f.SetColWidth(sheetName, "B", "B", 15) // password
		f.SetColWidth(sheetName, "C", "C", 25) // email
		f.SetColWidth(sheetName, "D", "D", 15) // phone
		f.SetColWidth(sheetName, "E", "E", 15) // real_name
		f.SetColWidth(sheetName, "F", "F", 20) // department
		f.SetColWidth(sheetName, "G", "G", 15) // title
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_template.xlsx", userType))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Excel模板失败",
			"data":    nil,
		})
		return
	}
}

func (h *UserHandler) ImportUsersFromCSV(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType := c.PostForm("user_type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请指定用户类型 (student/teacher)",
			"data":    nil,
		})
		return
	}

	if userType != "student" && userType != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户类型只能是 student 或 teacher",
			"data":    nil,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请上传CSV文件",
			"data":    nil,
		})
		return
	}

	if !strings.HasSuffix(file.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只支持CSV文件格式",
			"data":    nil,
		})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小不能超过5MB",
			"data":    nil,
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "打开文件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV文件格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.processImportData(c, records, userType, file.Filename)
}
