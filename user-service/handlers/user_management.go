package handlers

import (
	"fmt"
	"strconv"

	"credit-management/user-service/models"
	"credit-management/user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := utils.GetCurrentUserID(c)
	if userID == "" {
		userID = currentUserID
	}

	var userType string
	err := h.db.Table("users").Select("user_type").Where("uuid = ?", userID).Scan(&userType).Error
	if err != nil || userType == "" {
		utils.SendNotFound(c, "用户不存在")
		return
	}

	if userID == currentUserID {
		var result map[string]interface{}
		var viewName string
		switch userType {
		case "student":
			viewName = "student_complete_info"
		case "teacher":
			viewName = "teacher_complete_info"
		default:
			// 管理员等其他类型直接查users表所有字段
			h.db.Table("users").Where("uuid = ?", userID).Find(&result)
			// 移除敏感字段（如密码哈希），避免暴露给前端
			sanitizeUserResult(result)
			utils.SendSuccessResponse(c, result)
			return
		}
		h.db.Table(viewName).Where("uuid = ?", userID).Find(&result)
		// 视图中如果包含密码等敏感字段，同样进行清理
		sanitizeUserResult(result)
		utils.SendSuccessResponse(c, result)
		return
	}

	var user models.User
	if err := h.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	currentUserRole := utils.GetCurrentUserRole(c)
	if !utils.CanViewUserDetails(currentUserRole, user.UserType) {
		utils.SendForbidden(c, "权限不足")
		return
	}

	response := h.convertToUserResponse(user)
	utils.SendSuccessResponse(c, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		// 从上下文中获取当前用户ID（由认证中间件设置）
		userID = utils.GetCurrentUserID(c)
		if userID == "" {
			utils.SendUnauthorized(c)
			return
		}
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	validator := utils.NewValidator()

	var user models.User
	if err := h.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 验证邮箱唯一性
	if req.Email != "" {
		var existingUser models.User
		if err := h.db.Where("email = ? AND uuid != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			utils.SendConflict(c, "邮箱已被使用")
			return
		}
		user.Email = req.Email
	}

	// 更新用户信息
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Avatar != "" {
		user.Avatar = &req.Avatar
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Grade != nil {
		// Only update grade if it's a valid 4-digit string or null
		if *req.Grade == "" || (len(*req.Grade) == 4 && isNumeric(*req.Grade)) {
			user.Grade = req.Grade
		}
	}
	if req.Title != nil {
		user.Title = req.Title
	}

	// 处理部门 / 班级变更：
	// 1. 优先使用显式传入的 department_id
	// 2. 教师：根据传入的学部名称（department/college）反查学部 department_id
	// 3. 学生：根据传入的专业 + 班级名称反查班级 department_id
	if req.DepartmentID != "" {
		// 直接使用前端提供的部门ID
		user.DepartmentID = &req.DepartmentID
	} else if user.UserType == "teacher" && (req.Department != "" || req.College != "") {
		// 教师：根据学部名称反查学部 department_id
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
		user.DepartmentID = &collegeDept.ID
	} else if req.Major != "" && req.Class != "" {
		// 学生：根据“专业名称 + 班级代码”反查班级对应的 department_id
		type deptRow struct {
			ID string
		}
		var classDept deptRow
		// 班级节点的 name 是班级代码（例如 202405C1），父节点是专业名称（例如 计算机科学与技术）
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
		user.DepartmentID = &classDept.ID
	}

	if req.StudentID != nil {
		if user.UserType != "student" {
			utils.SendBadRequest(c, "只有学生用户可以设置学号")
			return
		}
		if err := validator.ValidateStudentID(*req.StudentID); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
		var existing models.User
		if err := h.db.Where("student_id = ? AND uuid != ?", *req.StudentID, userID).First(&existing).Error; err == nil {
			utils.SendConflict(c, "学号已被使用")
			return
		}
		user.StudentID = req.StudentID
	}

	if req.TeacherID != nil {
		if user.UserType != "teacher" {
			utils.SendBadRequest(c, "只有教师用户可以设置工号")
			return
		}
		if err := validator.ValidateTeacherID(*req.TeacherID); err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}
		var existing models.User
		if err := h.db.Where("teacher_id = ? AND uuid != ?", *req.TeacherID, userID).First(&existing).Error; err == nil {
			utils.SendConflict(c, "工号已被使用")
			return
		}
		user.TeacherID = req.TeacherID
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	userResponse := h.convertToUserResponse(user)
	utils.SendSuccessResponse(c, userResponse)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.SendBadRequest(c, "用户ID不能为空")
		return
	}

	var user models.User
	if err := h.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "用户删除成功"})
}

func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UUIDs []string `json:"ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	var users []models.User
	if err := h.db.Where("uuid IN ?", req.UUIDs).Find(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	if len(users) != len(req.UUIDs) {
		utils.SendBadRequest(c, "部分用户不存在")
		return
	}

	if err := h.db.Delete(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"deleted_count": len(users)})
}

func (h *UserHandler) BatchUpdateUserStatus(c *gin.Context) {
	var req struct {
		UUIDs  []string `json:"ids" binding:"required,min=1"`
		Status string   `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.db.Model(&models.User{}).Where("uuid IN ?", req.UUIDs).Update("status", req.Status).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"updated_count": len(req.UUIDs), "status": req.Status})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		UUID        string `json:"id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	validator := utils.NewValidator()
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("uuid = ?", req.UUID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "用户不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "密码重置成功"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := utils.GetCurrentUserID(c)
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	var user models.User
	if err := h.db.Where("uuid = ?", userID).First(&user).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		utils.SendBadRequest(c, "原密码错误")
		return
	}

	// 验证新密码复杂度
	validator := utils.NewValidator()
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "密码修改成功"})
}

func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		userID = utils.GetCurrentUserID(c)
		if userID == "" {
			utils.SendUnauthorized(c)
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	validator := utils.NewValidator()
	page, pageSize, _ = validator.ValidatePagination(strconv.Itoa(page), strconv.Itoa(pageSize))

	// 这里可以添加用户活动记录的查询逻辑
	// 例如：登录记录、操作日志等
	// 目前返回空结果，后续可以扩展

	response := gin.H{
		"id":          userID,
		"activities":  []interface{}{},
		"total":       0,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": 0,
	}

	utils.SendSuccessResponse(c, response)
}

func (h *UserHandler) ExportUsers(c *gin.Context) {
	format := c.DefaultQuery("format", "xlsx")
	userType := c.Query("user_type")
	status := c.Query("status")

	query := h.db.Model(&models.User{})

	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "json":
		utils.SendSuccessResponse(c, users)
	case "xlsx":
		// 创建 Excel 文件
		f := excelize.NewFile()
		defer f.Close()

		sheetName := "用户列表"
		f.SetSheetName("Sheet1", sheetName)

		// 根据用户类型设置表头
		var headers []string
		switch userType {
		case "teacher":
			// 教师：部门一般是到“学部”一级
			headers = []string{"工号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "职称", "状态"}
		case "student":
			// 学生：部门为班级，向上关联专业和学部
			headers = []string{"学号", "姓名", "邮箱", "手机号", "学部", "专业", "班级", "年级", "状态"}
		default:
			// 其他用户：仅展示部门名称
			headers = []string{"UUID", "用户名", "姓名", "邮箱", "手机号", "用户类型", "学部", "专业", "班级", "状态"}
		}

		// 写入表头
		for i, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cell, header)
		}

		// 写入数据
		for rowIndex, user := range users {
			// Excel 行号从 2 开始（1 是表头）
			rowNumber := rowIndex + 2

			var rowValues []interface{}

			// 拆分部门为学部 / 专业 / 班级
			var collegeName, majorName, className string
			if user.DepartmentID != nil && *user.DepartmentID != "" {
				// 查询部门层级：可能是 class / major / college / 其他
				type deptRow struct {
					Name     string
					DeptType string
					ParentID *string
				}

				var dept deptRow
				if err := h.db.Raw(`
					SELECT name, dept_type::text AS dept_type, parent_id::text AS parent_id
					FROM departments
					WHERE id = ?
					LIMIT 1
				`, *user.DepartmentID).Scan(&dept).Error; err == nil && dept.Name != "" {
					switch dept.DeptType {
					case "class":
						className = dept.Name
						// 找到专业
						var major deptRow
						if err := h.db.Raw(`
							SELECT name, dept_type::text AS dept_type, parent_id::text AS parent_id
							FROM departments
							WHERE id = ?
							LIMIT 1
						`, dept.ParentID).Scan(&major).Error; err == nil && major.Name != "" {
							majorName = major.Name
							// 找到学部
							var college deptRow
							if err := h.db.Raw(`
								SELECT name, dept_type::text AS dept_type, parent_id::text AS parent_id
								FROM departments
								WHERE id = ?
								LIMIT 1
							`, major.ParentID).Scan(&college).Error; err == nil && college.Name != "" {
								collegeName = college.Name
							}
						}
					case "major":
						majorName = dept.Name
						// 找到学部
						var college deptRow
						if err := h.db.Raw(`
							SELECT name, dept_type::text AS dept_type, parent_id::text AS parent_id
							FROM departments
							WHERE id = ?
							LIMIT 1
						`, dept.ParentID).Scan(&college).Error; err == nil && college.Name != "" {
							collegeName = college.Name
						}
					case "college":
						collegeName = dept.Name
					default:
						// 其他类型部门先简单放在“学部”列
						collegeName = dept.Name
					}
				}
			}

			switch userType {
			case "teacher":
				rowValues = []interface{}{
					utils.DerefString(user.TeacherID),
					user.RealName,
					user.Email,
					utils.DerefString(user.Phone),
					collegeName,
					majorName,
					className,
					utils.DerefString(user.Title),
					user.Status,
				}
			case "student":
				rowValues = []interface{}{
					utils.DerefString(user.StudentID),
					user.RealName,
					user.Email,
					utils.DerefString(user.Phone),
					collegeName,
					majorName,
					className,
					utils.DerefString(user.Grade),
					user.Status,
				}
			default:
				rowValues = []interface{}{
					user.UUID,
					user.Username,
					user.RealName,
					user.Email,
					utils.DerefString(user.Phone),
					user.UserType,
					collegeName,
					majorName,
					className,
					user.Status,
				}
			}

			for colIndex, value := range rowValues {
				cell := fmt.Sprintf("%c%d", 'A'+colIndex, rowNumber)
				f.SetCellValue(sheetName, cell, value)
			}
		}

		// 设置响应头并输出文件
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

		filename := "users.xlsx"
		if userType == "teacher" {
			filename = "teachers.xlsx"
		} else if userType == "student" {
			filename = "students.xlsx"
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if err := f.Write(c.Writer); err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
	case "csv":
		utils.SendSuccessResponse(c, gin.H{"message": "CSV导出功能待实现", "count": len(users)})
	default:
		utils.SendBadRequest(c, "不支持的导出格式")
	}
}

// convertToUserResponse 将User模型转换为UserResponse
func (h *UserHandler) convertToUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		UUID:         user.UUID,
		StudentID:    user.StudentID,
		TeacherID:    user.TeacherID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        utils.DerefString(user.Phone),
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       utils.DerefString(user.Avatar),
		DepartmentID: utils.DerefString(user.DepartmentID),
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Grade:        user.Grade,
		Title:        user.Title,
	}
}
