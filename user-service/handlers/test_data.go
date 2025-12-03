package handlers

import (
	"net/http"
	"os"

	"credit-management/user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateDepartmentsFromOptions 根据 options.json 生成学部/专业/班级等部门测试数据
// 注意：仅用于测试环境，不建议在生产环境频繁调用
func (h *UserHandler) GenerateDepartmentsFromOptions(c *gin.Context) {
	// 通过环境变量控制测试期：
	// TEST_DATA_MODE=enabled 时代表处于测试期，此时接口无需鉴权即可使用；
	// 其他值或未设置时，接口禁止使用。
	if os.Getenv("TEST_DATA_MODE") != "enabled" {
		utils.SendForbidden(c, "测试数据接口仅在测试期可用")
		return
	}

	options, err := loadOptions()
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		utils.SendInternalServerError(c, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	type deptRow struct {
		ID   string
		Name string
	}

	createdColleges := 0
	createdMajors := 0
	createdClasses := 0

	// 1. 确保学校节点存在
	var school deptRow
	if err := tx.Raw(
		`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'school' LIMIT 1`,
		"上海电力大学",
	).Scan(&school).Error; err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		utils.SendInternalServerError(c, err)
		return
	}

	if school.ID == "" {
		schoolID := uuid.New().String()
		if err := tx.Exec(
			`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'school', NULL)`,
			schoolID, "上海电力大学", "SUEP",
		).Error; err != nil {
			tx.Rollback()
			utils.SendInternalServerError(c, err)
			return
		}
		school.ID = schoolID
		school.Name = "上海电力大学"
	}

	// 2. 遍历学部
	for _, collegeOpt := range options.Colleges {
		var college deptRow
		if err := tx.Raw(
			`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'college' AND parent_id = ? LIMIT 1`,
			collegeOpt.Value, school.ID,
		).Scan(&college).Error; err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			utils.SendInternalServerError(c, err)
			return
		}

		if college.ID == "" {
			collegeID := uuid.New().String()
			if err := tx.Exec(
				`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'college', ?)`,
				collegeID, collegeOpt.Value, collegeOpt.Value, school.ID,
			).Error; err != nil {
				tx.Rollback()
				utils.SendInternalServerError(c, err)
				return
			}
			college.ID = collegeID
			college.Name = collegeOpt.Value
			createdColleges++
		}

		// 3. 处理该学部下的专业
		majorOptions, ok := options.Majors[collegeOpt.Value]
		if !ok {
			continue
		}

		for _, majorOpt := range majorOptions {
			var major deptRow
			if err := tx.Raw(
				`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'major' AND parent_id = ? LIMIT 1`,
				majorOpt.Value, college.ID,
			).Scan(&major).Error; err != nil && err != gorm.ErrRecordNotFound {
				tx.Rollback()
				utils.SendInternalServerError(c, err)
				return
			}

			if major.ID == "" {
				majorID := uuid.New().String()
				if err := tx.Exec(
					`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'major', ?)`,
					majorID, majorOpt.Value, majorOpt.Value, college.ID,
				).Error; err != nil {
					tx.Rollback()
					utils.SendInternalServerError(c, err)
					return
				}
				major.ID = majorID
				major.Name = majorOpt.Value
				createdMajors++
			}

			// 4. 处理该专业下的班级
			classOptions, ok := options.Classes[majorOpt.Value]
			if !ok {
				continue
			}

			for _, classOpt := range classOptions {
				var class deptRow
				if err := tx.Raw(
					`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'class' AND parent_id = ? LIMIT 1`,
					classOpt.Value, major.ID,
				).Scan(&class).Error; err != nil && err != gorm.ErrRecordNotFound {
					tx.Rollback()
					utils.SendInternalServerError(c, err)
					return
				}

				if class.ID == "" {
					classID := uuid.New().String()
					if err := tx.Exec(
						`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'class', ?)`,
						classID, classOpt.Value, classOpt.Value, major.ID,
					).Error; err != nil {
						tx.Rollback()
						utils.SendInternalServerError(c, err)
						return
					}
					createdClasses++
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":              200,
		"message":           "部门测试数据生成完成",
		"created_colleges":  createdColleges,
		"created_majors":    createdMajors,
		"created_classes":   createdClasses,
		"school_department": school.Name,
	})
}


