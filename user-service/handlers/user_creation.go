package handlers

import (
	"credit-management/user-service/models"
	"credit-management/user-service/utils"
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) Register(c *gin.Context) {
	var req models.StudentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	validator := utils.NewValidator()

	// 验证密码复杂度
	if err := validator.ValidatePassword(req.Password); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证手机号格式
	if err := validator.ValidatePhone(req.Phone); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证学号格式
	if err := validator.ValidateStudentID(req.StudentID); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证年级格式
	if err := validator.ValidateGrade(req.Grade); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 检查用户名唯一性
	if err := h.checkUsernameUniqueness(req.Username); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查邮箱唯一性
	if err := h.checkEmailUniqueness(req.Email); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查手机号唯一性
	if err := h.checkPhoneUniqueness(req.Phone); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查学号唯一性
	if err := h.checkStudentIDUniqueness(req.StudentID); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
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
		utils.SendInternalServerError(c, err)
		return
	}

	userResponse := h.convertToUserResponse(user)
	utils.SendCreatedResponse(c, "学生注册成功", gin.H{
		"message": "学生注册成功",
		"user":    userResponse,
	})
}

func (h *UserHandler) CreateTeacher(c *gin.Context) {
	claims, exists := utils.GetUserClaims(c)
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	if !utils.IsAdmin(claims["user_type"].(string)) {
		utils.SendForbidden(c, "只有管理员可以创建教师")
		return
	}

	var req models.TeacherRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	validator := utils.NewValidator()

	// 验证密码复杂度
	if err := validator.ValidatePassword(req.Password); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证手机号格式
	if err := validator.ValidatePhone(req.Phone); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 检查用户名唯一性
	if err := h.checkUsernameUniqueness(req.Username); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查邮箱唯一性
	if err := h.checkEmailUniqueness(req.Email); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查手机号唯一性
	if err := h.checkPhoneUniqueness(req.Phone); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
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
		utils.SendInternalServerError(c, err)
		return
	}

	userResponse := h.convertToUserResponse(user)
	utils.SendCreatedResponse(c, "教师创建成功", gin.H{
		"message": "教师创建成功",
		"user":    userResponse,
	})
}

// CreateStudent 管理员创建学生
func (h *UserHandler) CreateStudent(c *gin.Context) {
	claims, exists := utils.GetUserClaims(c)
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	if !utils.IsAdmin(claims["user_type"].(string)) {
		utils.SendForbidden(c, "只有管理员可以创建学生")
		return
	}

	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.UserType != "student" {
		utils.SendBadRequest(c, "只能创建学生用户")
		return
	}

	validator := utils.NewValidator()

	// 验证密码复杂度
	if err := validator.ValidatePassword(req.Password); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证手机号格式（如果提供）
	if req.Phone != "" {
		if err := validator.ValidatePhone(req.Phone); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
	}

	// 验证学号格式（如果提供）
	if req.StudentID != "" {
		if err := validator.ValidateStudentID(req.StudentID); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
	}

	// 验证年级格式（如果提供）
	if req.Grade != "" {
		if err := validator.ValidateGrade(req.Grade); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
	}

	// 验证用户请求
	if err := h.validateUserRequest(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 检查用户名唯一性
	if err := h.checkUsernameUniqueness(req.Username); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查邮箱唯一性
	if err := h.checkEmailUniqueness(req.Email); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	// 检查手机号唯一性（如果提供）
	if req.Phone != "" {
		if err := h.checkPhoneUniqueness(req.Phone); err != nil {
			utils.SendConflict(c, err.Error())
			return
		}
	}

	// 检查学号唯一性（如果提供）
	if req.StudentID != "" {
		if err := h.checkStudentIDUniqueness(req.StudentID); err != nil {
			utils.SendConflict(c, err.Error())
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
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
		utils.SendInternalServerError(c, err)
		return
	}

	userResponse := h.convertToUserResponse(user)
	utils.SendCreatedResponse(c, "学生创建成功", gin.H{
		"message": "学生创建成功",
		"user":    userResponse,
	})
}

// 辅助函数：检查用户名唯一性
func (h *UserHandler) checkUsernameUniqueness(username string) error {
	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ?", username).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("用户名已被删除的用户使用，请选择其他用户名")
		} else {
			return fmt.Errorf("用户名已存在")
		}
	}
	return nil
}

// 辅助函数：检查邮箱唯一性
func (h *UserHandler) checkEmailUniqueness(email string) error {
	var existingUser models.User
	if err := h.db.Unscoped().Where("email = ?", email).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("邮箱已被删除的用户使用，请使用其他邮箱")
		} else {
			return fmt.Errorf("邮箱已被使用")
		}
	}
	return nil
}

// 辅助函数：检查手机号唯一性
func (h *UserHandler) checkPhoneUniqueness(phone string) error {
	var existingUser models.User
	if err := h.db.Unscoped().Where("phone = ?", phone).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("手机号已被删除的用户使用，请使用其他手机号")
		} else {
			return fmt.Errorf("手机号已被使用")
		}
	}
	return nil
}

// 辅助函数：检查学号唯一性
func (h *UserHandler) checkStudentIDUniqueness(studentID string) error {
	var existingUser models.User
	if err := h.db.Unscoped().Where("student_id = ?", studentID).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("学号已被删除的用户使用，请使用其他学号")
		} else {
			return fmt.Errorf("学号已被使用")
		}
	}
	return nil
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
	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType := c.PostForm("user_type")
	if userType == "" {
		utils.SendBadRequest(c, "请指定用户类型 (student/teacher)")
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidateUserType(userType); err != nil {
		utils.SendBadRequest(c, "用户类型只能是 student 或 teacher")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendBadRequest(c, "请上传文件")
		return
	}

	// 验证文件大小
	maxFileSize := utils.GetMaxFileSize()
	if err := validator.ValidateFileSize(file.Size, maxFileSize); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	fileName := strings.ToLower(file.Filename)
	var records [][]string
	var parseError error

	// 验证文件类型
	allowedTypes := utils.GetAllowedFileTypes()
	if err := validator.ValidateFileType(fileName, allowedTypes); err != nil {
		utils.SendBadRequest(c, "只支持CSV、XLSX、XLS文件格式")
		return
	}

	if strings.HasSuffix(fileName, ".csv") {
		records, parseError = h.parseCSVFile(file)
	} else if strings.HasSuffix(fileName, ".xlsx") || strings.HasSuffix(fileName, ".xls") {
		records, parseError = h.parseExcelFile(file)
	} else {
		utils.SendBadRequest(c, "只支持CSV、XLSX、XLS文件格式")
		return
	}

	if parseError != nil {
		utils.SendBadRequest(c, "文件解析失败: "+parseError.Error())
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
		utils.SendBadRequest(c, "文件至少需要包含标题行和一行数据")
		return
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// 验证必需的列
	requiredColumns := []string{"username", "password", "email", "real_name"}
	for _, col := range requiredColumns {
		if _, exists := headerMap[col]; !exists {
			utils.SendBadRequest(c, fmt.Sprintf("缺少必需的列: %s", col))
			return
		}
	}

	// 根据用户类型验证特定列
	switch userType {
	case "student":
		studentColumns := []string{"student_id", "college", "major", "class", "grade"}
		for _, col := range studentColumns {
			if _, exists := headerMap[col]; !exists {
				utils.SendBadRequest(c, fmt.Sprintf("学生导入缺少必需的列: %s", col))
				return
			}
		}
	case "teacher":
		teacherColumns := []string{"department", "title"}
		for _, col := range teacherColumns {
			if _, exists := headerMap[col]; !exists {
				utils.SendBadRequest(c, fmt.Sprintf("教师导入缺少必需的列: %s", col))
				return
			}
		}
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
		utils.SendBadRequestWithData(c, "数据验证失败", gin.H{
			"errors":       errors,
			"total_rows":   len(records) - 1,
			"valid_rows":   len(users),
			"invalid_rows": len(errors),
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
		utils.SendBadRequestWithData(c, "批量创建用户失败", gin.H{
			"errors":        createErrors,
			"created_count": 0,
			"total_count":   len(users),
			"created_users": []models.UserResponse{},
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendCreatedResponse(c, "批量导入成功", gin.H{
		"created_count": len(createdUsers),
		"total_count":   len(users),
		"created_users": createdUsers,
		"user_type":     userType,
		"file_name":     fileName,
		"file_type":     filepath.Ext(fileName),
	})
}

func (h *UserHandler) GetUserCSVTemplate(c *gin.Context) {
	userType := c.Query("user_type")
	if userType == "" {
		utils.SendBadRequest(c, "请指定用户类型 (student/teacher)")
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidateUserType(userType); err != nil {
		utils.SendBadRequest(c, "用户类型只能是 student 或 teacher")
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
			utils.SendInternalServerError(c, err)
			return
		}
	}
}

func (h *UserHandler) GetUserExcelTemplate(c *gin.Context) {
	userType := c.Query("user_type")
	if userType == "" {
		utils.SendBadRequest(c, "请指定用户类型 (student/teacher)")
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidateUserType(userType); err != nil {
		utils.SendBadRequest(c, "用户类型只能是 student 或 teacher")
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

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_template.xlsx", userType))

	if err := f.Write(c.Writer); err != nil {
		utils.SendInternalServerError(c, err)
		return
	}
}

func (h *UserHandler) ImportUsersFromCSV(c *gin.Context) {
	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType := c.PostForm("user_type")
	if userType == "" {
		utils.SendBadRequest(c, "请指定用户类型 (student/teacher)")
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidateUserType(userType); err != nil {
		utils.SendBadRequest(c, "用户类型只能是 student 或 teacher")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendBadRequest(c, "请上传CSV文件")
		return
	}

	if !strings.HasSuffix(file.Filename, ".csv") {
		utils.SendBadRequest(c, "只支持CSV文件格式")
		return
	}

	// 验证文件大小
	maxFileSize := utils.GetMaxFileSize()
	if err := validator.ValidateFileSize(file.Size, maxFileSize); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	src, err := file.Open()
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		utils.SendBadRequest(c, "CSV文件格式错误: "+err.Error())
		return
	}

	h.processImportData(c, records, userType, file.Filename)
}
