package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type OptionsResponse struct {
	Colleges      []SelectOption            `json:"colleges"`
	Majors        map[string][]SelectOption `json:"majors"`
	Classes       map[string][]SelectOption `json:"classes"`
	Grades        []SelectOption            `json:"grades"`
	UserStatuses  []SelectOption            `json:"user_statuses"`
	TeacherTitles []SelectOption            `json:"teacher_titles"`
}

// DepartmentNode 描述 new_options.json 中的层级节点（学校 / 学部 / 专业 / 班级）
type DepartmentNode struct {
	Type     string                       `json:"type"`
	Children []map[string]DepartmentNode  `json:"children"`
}

// RawOptionsTree 对应 new_options.json 的整体结构
// 顶层是“上海电力大学”树形结构，外加年级、用户状态、教师职称等下拉选项
type RawOptionsTree struct {
	University    DepartmentNode `json:"上海电力大学"`
	Grades        []SelectOption `json:"grades"`
	UserStatuses  []SelectOption `json:"user_statuses"`
	TeacherTitles []SelectOption `json:"teacher_titles"`
}

// loadOptions 从配置文件中加载下拉选项配置
func loadOptions() (*OptionsResponse, error) {
	configPath := os.Getenv("OPTIONS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/options.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 先按新的树形结构读取
	var raw RawOptionsTree
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// 将树形结构转换为前端及其它逻辑沿用的扁平结构
	result := &OptionsResponse{
		Colleges:      []SelectOption{},
		Majors:        make(map[string][]SelectOption),
		Classes:       make(map[string][]SelectOption),
		Grades:        raw.Grades,
		UserStatuses:  raw.UserStatuses,
		TeacherTitles: raw.TeacherTitles,
	}

	// 顶层：大学 -> 学部/学部
	for _, collegeWrapper := range raw.University.Children {
		for collegeName, collegeNode := range collegeWrapper {
			// 记录“学部/学部”作为一个 college
			result.Colleges = append(result.Colleges, SelectOption{
				Value: collegeName,
				Label: collegeName,
			})

			// 下一级：专业
			for _, majorWrapper := range collegeNode.Children {
				for majorName, majorNode := range majorWrapper {
					result.Majors[collegeName] = append(result.Majors[collegeName], SelectOption{
						Value: majorName,
						Label: majorName,
					})

					// 再下一级：班级
					for _, classWrapper := range majorNode.Children {
						for className := range classWrapper {
							result.Classes[majorName] = append(result.Classes[majorName], SelectOption{
								Value: className,
								Label: className,
							})
						}
					}
				}
			}
		}
	}

	return result, nil
}

// GetOptions 读取配置文件并返回前端所需的下拉选项
func GetOptions(c *gin.Context) {
	options, err := loadOptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "failed to load options config",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": options})
}

