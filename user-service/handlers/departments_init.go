package handlers

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InitDepartments 在服务启动时根据配置初始化 departments 表（如果尚未存在对应记录）
func InitDepartments(db *gorm.DB) error {
	options, err := loadOptions()
	if err != nil {
		return err
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	type deptRow struct {
		ID   string
		Name string
	}

	// 1. 确保学校节点存在
	var school deptRow
	if err := tx.Raw(
		`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'school' LIMIT 1`,
		"上海电力大学",
	).Scan(&school).Error; err != nil && err != gorm.ErrRecordNotFound {
		_ = tx.Rollback()
		return err
	}

	if school.ID == "" {
		schoolID := uuid.New().String()
		if err := tx.Exec(
			`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'school', NULL)`,
			schoolID, "上海电力大学", "SUEP",
		).Error; err != nil {
			_ = tx.Rollback()
			return err
		}
		school.ID = schoolID
		school.Name = "上海电力大学"
	}

	// 2. 遍历“学部/学部”
	for _, collegeOpt := range options.Colleges {
		var college deptRow
		if err := tx.Raw(
			`SELECT id, name FROM departments WHERE name = ? AND dept_type = 'college' AND parent_id = ? LIMIT 1`,
			collegeOpt.Value, school.ID,
		).Scan(&college).Error; err != nil && err != gorm.ErrRecordNotFound {
			_ = tx.Rollback()
			return err
		}

		if college.ID == "" {
			collegeID := uuid.New().String()
			if err := tx.Exec(
				`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'college', ?)`,
				collegeID, collegeOpt.Value, collegeOpt.Value, school.ID,
			).Error; err != nil {
				_ = tx.Rollback()
				return err
			}
			college.ID = collegeID
			college.Name = collegeOpt.Value
		}

		// 3. 处理该“学部/学部”下的专业
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
				_ = tx.Rollback()
				return err
			}

			if major.ID == "" {
				majorID := uuid.New().String()
				if err := tx.Exec(
					`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'major', ?)`,
					majorID, majorOpt.Value, majorOpt.Value, college.ID,
				).Error; err != nil {
					_ = tx.Rollback()
					return err
				}
				major.ID = majorID
				major.Name = majorOpt.Value
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
					_ = tx.Rollback()
					return err
				}

				if class.ID == "" {
					classID := uuid.New().String()
					if err := tx.Exec(
						`INSERT INTO departments (id, name, code, dept_type, parent_id) VALUES (?, ?, ?, 'class', ?)`,
						classID, classOpt.Value, classOpt.Value, major.ID,
					).Error; err != nil {
						_ = tx.Rollback()
						return err
					}
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Println("部门数据已根据配置初始化/更新完成")
	return nil
}


