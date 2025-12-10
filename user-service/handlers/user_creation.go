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

	// 处理班级/部门信息：
	// 如果提供了学部 + 专业 + 班级名称，则验证层级关系并反查 department_id
	collegeName := req.College
	if collegeName == "" {
		collegeName = req.Department
	}
	if collegeName != "" && req.Major != "" && req.Class != "" {
		type deptRow struct {
			ID string
		}

		// 首先验证学部是否存在
		var collegeDept deptRow
		if err := h.db.Raw(`
			SELECT id
			FROM departments
			WHERE dept_type = 'college' AND name = ?
			LIMIT 1
		`, collegeName).Scan(&collegeDept).Error; err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
		if collegeDept.ID == "" {
			utils.SendBadRequest(c, "未找到对应的学部（学部="+collegeName+"）")
			return
		}

		// 验证专业是否属于该学部
		var majorDept deptRow
		if err := h.db.Raw(`
			SELECT id
			FROM departments
			WHERE dept_type = 'major' AND name = ? AND parent_id = ?
			LIMIT 1
		`, req.Major, collegeDept.ID).Scan(&majorDept).Error; err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
		if majorDept.ID == "" {
			utils.SendBadRequest(c, "未找到对应的专业，或专业不属于该学部（学部="+collegeName+", 专业="+req.Major+"）")
			return
		}

		// 验证班级是否属于该专业
		var classDept deptRow
		if err := h.db.Raw(`
			SELECT id
			FROM departments
			WHERE dept_type = 'class' AND name = ? AND parent_id = ?
			LIMIT 1
		`, req.Class, majorDept.ID).Scan(&classDept).Error; err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
		if classDept.ID == "" {
			utils.SendBadRequest(c, "未找到对应的班级，或班级不属于该专业（专业="+req.Major+", 班级="+req.Class+"）")
			return
		}
		req.DepartmentID = classDept.ID
	}

	// 最终仍然没有 department_id，则认为未指定班级
	if req.DepartmentID == "" {
		utils.SendBadRequest(c, "学生必须指定班级")
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

	studentID := req.StudentID
	user := models.User{
		StudentID:    &studentID,
		Username:     req.Username,
		Password:     string(hashedPassword),
		Email:        req.Email,
		Phone:        &req.Phone,
		RealName:     req.RealName,
		UserType:     "student",
		Status:       "active",
		DepartmentID: &req.DepartmentID,
		Grade:        &req.Grade,
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

	// 教师必须隶属某个学部/部门：
	// 1. 如果前端直接提供了 department_id，则优先使用
	// 2. 否则根据传入的学部名称（department/college）反查学部 department_id
	if req.DepartmentID == "" && (req.Department != "" || req.College != "") {
		type deptRow struct {
			ID string
		}
		var collegeDept deptRow
		name := req.Department
		if name == "" {
			name = req.College
		}
		if err := h.db.Raw(`
			SELECT id
			FROM departments
			WHERE dept_type = 'college' AND name = ?
			LIMIT 1
		`, name).Scan(&collegeDept).Error; err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
		if collegeDept.ID == "" {
			utils.SendBadRequest(c, "未找到对应的学部，请检查学部名称是否匹配")
			return
		}
		req.DepartmentID = collegeDept.ID
	}

	if req.DepartmentID == "" {
		utils.SendBadRequest(c, "教师必须指定学部")
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

	teacherID := req.TeacherID
	user := models.User{
		TeacherID:    &teacherID,
		Username:     req.Username,
		Password:     string(hashedPassword),
		Email:        req.Email,
		Phone:        &req.Phone,
		RealName:     req.RealName,
		UserType:     "teacher",
		Status:       "active",
		DepartmentID: &req.DepartmentID,
		Title:        &req.Title,
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

	// 学生必须提供学号
	if req.StudentID == "" {
		utils.SendBadRequest(c, "学生必须提供学号")
		return
	}
	if err := validator.ValidateStudentID(req.StudentID); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证年级格式（如果提供）
	if req.Grade != "" {
		if err := validator.ValidateGrade(req.Grade); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
	}

	if req.UserType != "student" && req.UserType != "teacher" {
		utils.SendBadRequest(c, "用户类型必须是student或teacher")
		return
	}

	// 学生创建时要求指定班级：
	// 1. 如果直接提供了 department_id，则优先使用
	// 2. 否则尝试根据学部 + 专业 + 班级名称反查班级 department_id，并验证层级关系
	if req.UserType == "student" {
		if req.DepartmentID == "" && req.Major != "" && req.Class != "" {
			collegeName := req.College
			if collegeName == "" {
				collegeName = req.Department
			}

			type deptRow struct {
				ID string
			}

			if collegeName != "" {
				// 首先验证学部是否存在
				var collegeDept deptRow
				if err := h.db.Raw(`
					SELECT id
					FROM departments
					WHERE dept_type = 'college' AND name = ?
					LIMIT 1
				`, collegeName).Scan(&collegeDept).Error; err != nil {
					utils.SendInternalServerError(c, err)
					return
				}
				if collegeDept.ID == "" {
					utils.SendBadRequest(c, "未找到对应的学部（学部="+collegeName+"）")
					return
				}

				// 验证专业是否属于该学部
				var majorDept deptRow
				if err := h.db.Raw(`
					SELECT id
					FROM departments
					WHERE dept_type = 'major' AND name = ? AND parent_id = ?
					LIMIT 1
				`, req.Major, collegeDept.ID).Scan(&majorDept).Error; err != nil {
					utils.SendInternalServerError(c, err)
					return
				}
				if majorDept.ID == "" {
					utils.SendBadRequest(c, "未找到对应的专业，或专业不属于该学部（学部="+collegeName+", 专业="+req.Major+"）")
					return
				}

				// 验证班级是否属于该专业
				var classDept deptRow
				if err := h.db.Raw(`
					SELECT id
					FROM departments
					WHERE dept_type = 'class' AND name = ? AND parent_id = ?
					LIMIT 1
				`, req.Class, majorDept.ID).Scan(&classDept).Error; err != nil {
					utils.SendInternalServerError(c, err)
					return
				}
				if classDept.ID == "" {
					utils.SendBadRequest(c, "未找到对应的班级，或班级不属于该专业（专业="+req.Major+", 班级="+req.Class+"）")
					return
				}
				req.DepartmentID = classDept.ID
			} else {
				// 没有提供学部名称，仅根据专业+班级查找（保持向后兼容，但不验证学部）
				var classDept deptRow
				if err := h.db.Raw(`
					SELECT c.id
					FROM departments c
					JOIN departments m ON c.parent_id = m.id
					WHERE c.dept_type = 'class' AND m.dept_type = 'major' AND m.name = ? AND c.name = ?
					LIMIT 1
				`, req.Major, req.Class).Scan(&classDept).Error; err != nil {
					utils.SendInternalServerError(c, err)
					return
				}
				if classDept.ID == "" {
					utils.SendBadRequest(c, "未找到对应的班级，请检查学部/专业/班级是否匹配")
					return
				}
				req.DepartmentID = classDept.ID
			}
		}

		// 最终仍然没有 department_id，则认为未指定班级
		if req.DepartmentID == "" {
			utils.SendBadRequest(c, "学生必须指定班级")
			return
		}
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

	if err := h.checkStudentIDUniqueness(req.StudentID); err != nil {
		utils.SendConflict(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	studentID := req.StudentID
	user := models.User{
		StudentID: &studentID,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		RealName:  req.RealName,
		UserType:  "student",
		Status:    "active",
	}

	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.DepartmentID != "" {
		user.DepartmentID = &req.DepartmentID
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
	if studentID == "" {
		return nil
	}
	var existingUser models.User
	if err := h.db.Unscoped().Where("student_id = ?", studentID).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("学号已被删除的用户使用，请使用其他学号")
		} else {
			return fmt.Errorf("学号已存在")
		}
	}
	return nil
}

// 辅助函数：检查工号唯一性
func (h *UserHandler) checkTeacherIDUniqueness(teacherID string) error {
	if teacherID == "" {
		return nil
	}
	var existingUser models.User
	if err := h.db.Unscoped().Where("teacher_id = ?", teacherID).First(&existingUser).Error; err == nil {
		if existingUser.DeletedAt.Valid {
			return fmt.Errorf("工号已被删除的用户使用，请使用其他工号")
		} else {
			return fmt.Errorf("工号已存在")
		}
	}
	return nil
}

func (h *UserHandler) ImportUsers(c *gin.Context) {
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

	validator := utils.NewValidator()

	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		key := strings.TrimSpace(header)
		if key != "" {
			headerMap[key] = i
		}
	}

	// 以当前导出的表头为准：
	// 学生：学号, 姓名, 邮箱, 手机号, 学部, 专业, 班级, 年级, 状态
	// 教师：工号, 姓名, 邮箱, 手机号, 学部, 专业, 班级, 职称, 状态
	var requiredColumns []string
	switch userType {
	case "student":
		requiredColumns = []string{"学号", "姓名", "邮箱", "学部", "专业", "班级", "年级"}
	case "teacher":
		requiredColumns = []string{"工号", "姓名", "邮箱", "学部", "专业", "班级", "职称"}
	default:
		utils.SendBadRequestWithData(c, "仅支持导入学生或教师数据", gin.H{
			"errors": []string{"仅支持导入学生或教师数据"},
		})
		return
	}

	for _, col := range requiredColumns {
		if _, exists := headerMap[col]; !exists {
			msg := fmt.Sprintf("缺少必需的列: %s", col)
			utils.SendBadRequestWithData(c, "数据验证失败", gin.H{
				"errors": []string{msg},
			})
			return
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

		getVal := func(col string) string {
			idx, ok := headerMap[col]
			if !ok || idx >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[idx])
		}

		var user models.UserRequest
		user.UserType = userType
		user.Email = getVal("邮箱")
		user.Phone = getVal("手机号")
		user.RealName = getVal("姓名")

		if userType == "student" {
			studentID := getVal("学号")
			collegeName := getVal("学部")
			majorName := getVal("专业")
			className := getVal("班级")
			grade := getVal("年级")

			if studentID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 学生必须提供学号", rowNum))
				continue
			}
			if collegeName == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 学生必须提供学部名称", rowNum))
				continue
			}
			if majorName == "" || className == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 学生必须提供专业和班级名称", rowNum))
				continue
			}

			// 使用学号作为用户名，密码使用默认密码
			defaultPassword := utils.GenerateDefaultPassword()
			user.StudentID = studentID
			user.Username = studentID
			user.Password = defaultPassword
			user.Grade = grade

			// 验证学部 -> 专业 -> 班级的层级关系
			type deptRow struct {
				ID string
			}

			// 首先验证学部是否存在
			var collegeDept deptRow
			if err := h.db.Raw(`
				SELECT id
				FROM departments
				WHERE dept_type = 'college' AND name = ?
				LIMIT 1
			`, collegeName).Scan(&collegeDept).Error; err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: 查询学部失败: %v", rowNum, err))
				continue
			}
			if collegeDept.ID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 未找到对应的学部（学部=%s）", rowNum, collegeName))
				continue
			}

			// 验证专业是否属于该学部
			var majorDept deptRow
			if err := h.db.Raw(`
				SELECT id
				FROM departments
				WHERE dept_type = 'major' AND name = ? AND parent_id = ?
				LIMIT 1
			`, majorName, collegeDept.ID).Scan(&majorDept).Error; err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: 查询专业失败: %v", rowNum, err))
				continue
			}
			if majorDept.ID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 未找到对应的专业，或专业不属于该学部（学部=%s, 专业=%s）", rowNum, collegeName, majorName))
				continue
			}

			// 验证班级是否属于该专业
			var classDept deptRow
			if err := h.db.Raw(`
				SELECT id
				FROM departments
				WHERE dept_type = 'class' AND name = ? AND parent_id = ?
				LIMIT 1
			`, className, majorDept.ID).Scan(&classDept).Error; err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: 查询班级失败: %v", rowNum, err))
				continue
			}
			if classDept.ID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 未找到对应的班级，或班级不属于该专业（专业=%s, 班级=%s）", rowNum, majorName, className))
				continue
			}
			user.DepartmentID = classDept.ID
		} else if userType == "teacher" {
			teacherID := getVal("工号")
			collegeName := getVal("学部")
			majorName := getVal("专业") // 当前未用于反查，仅保留
			className := getVal("班级") // 当前未用于反查
			title := getVal("职称")

			if teacherID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 教师必须提供工号", rowNum))
				continue
			}
			if collegeName == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 教师必须提供学部名称", rowNum))
				continue
			}

			user.TeacherID = teacherID
			user.Username = teacherID
			user.Password = utils.GenerateDefaultPassword()
			user.Title = title

			type deptRow struct {
				ID string
			}
			var collegeDept deptRow
			if err := h.db.Raw(`
				SELECT id
				FROM departments
				WHERE dept_type = 'college' AND name = ?
				LIMIT 1
			`, collegeName).Scan(&collegeDept).Error; err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: 查询学部失败: %v", rowNum, err))
				continue
			}
			if collegeDept.ID == "" {
				errors = append(errors, fmt.Sprintf("第%d行: 未找到对应的学部（学部=%s）", rowNum, collegeName))
				continue
			}
			user.DepartmentID = collegeDept.ID
			_ = majorName
			_ = className
		}

		// 处理状态列：允许为空（默认 active），否则按导出的英文枚举校验
		statusVal := getVal("状态")
		if statusVal == "" {
			statusVal = "active"
		} else {
			if err := validator.ValidateStatus(statusVal); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		}
		user.Status = statusVal

		// 与注册/更新保持一致的格式校验
		if err := validator.ValidateEmail(user.Email); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}
		if user.Phone != "" {
			if err := validator.ValidatePhone(user.Phone); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		}
		if err := validator.ValidateUsername(user.Username); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}
		if err := validator.ValidatePassword(user.Password); err != nil {
			// 密码格式错误也视为普通业务错误：记录但不中断后续行检查
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}

		// 学生/教师特定字段校验
		if userType == "student" {
			if err := validator.ValidateStudentID(user.StudentID); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
			if err := validator.ValidateGrade(user.Grade); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		} else if userType == "teacher" {
			if err := validator.ValidateTeacherID(user.TeacherID); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		}

		// 与注册/更新保持一致的唯一性检查
		if err := h.checkUsernameUniqueness(user.Username); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}
		if err := h.checkEmailUniqueness(user.Email); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}
		if user.Phone != "" {
			if err := h.checkPhoneUniqueness(user.Phone); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		}
		if userType == "student" {
			if err := h.checkStudentIDUniqueness(user.StudentID); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
		} else if userType == "teacher" {
			if err := h.checkTeacherIDUniqueness(user.TeacherID); err != nil {
				errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
				continue
			}
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
			Username: userReq.Username,
			Password: string(hashedPassword),
			Email:    userReq.Email,
			RealName: userReq.RealName,
			UserType: userReq.UserType,
			Status:   userReq.Status,
		}
		if userReq.Phone != "" {
			user.Phone = &userReq.Phone
		}
		if userReq.DepartmentID != "" {
			user.DepartmentID = &userReq.DepartmentID
		}

		if userType == "student" {
			studentID := userReq.StudentID
			user.StudentID = &studentID
			if userReq.Grade != "" {
				user.Grade = &userReq.Grade
			}
		} else {
			teacherID := userReq.TeacherID
			user.TeacherID = &teacherID
			if userReq.Title != "" {
				user.Title = &userReq.Title
			}
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
		// 与导出/导入表头保持一致
		template = [][]string{
			{"学号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "年级", "状态"},
			{"20240001", "张三", "student001@example.com", "13800138001", "计算机科学与技术学部", "计算机科学与技术", "202405C1", "2024", "active"},
		}
	} else {
		template = [][]string{
			{"工号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "职称", "状态"},
			{"T001", "王老师", "teacher001@example.com", "13800138003", "计算机科学与技术学部", "", "", "副教授", "active"},
		}
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_template.csv", userType))

	// 写入 UTF-8 BOM，避免在 Excel 中出现中文乱码
	if _, err := c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

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
		headers = []string{"学号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "年级", "状态"}
		examples = [][]string{
			{"20240001", "张三", "student001@example.com", "13800138001", "计算机科学与技术学部", "计算机科学与技术", "202405C1", "2024", "active"},
		}
	} else {
		headers = []string{"工号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "职称", "状态"}
		examples = [][]string{
			{"T001", "王老师", "teacher001@example.com", "13800138003", "计算机科学与技术学部", "", "", "副教授", "active"},
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
