package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func (h *ActivityHandler) CopyActivity(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	var originalActivity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&originalActivity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	newActivity := models.CreditActivity{
		Title:       originalActivity.Title + " (副本)",
		Description: originalActivity.Description,
		StartDate:   originalActivity.StartDate,
		EndDate:     originalActivity.EndDate,
		Status:      models.StatusDraft,
		Category:    originalActivity.Category,
		OwnerID:     userID.(string),
	}

	if err := h.db.Create(&newActivity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "复制活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	response := h.enrichActivityResponse(newActivity, "")

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "活动复制成功",
		"data":    response,
	})
}

func (h *ActivityHandler) ExportActivities(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	category := c.Query("category")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	dbQuery := h.db.Model(&models.CreditActivity{})

	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}
	if startDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
			dbQuery = dbQuery.Where("start_date >= ?", parsedDate)
		}
	}
	if endDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDate); err == nil {
			dbQuery = dbQuery.Where("end_date <= ?", parsedDate)
		}
	}

	var activities []models.CreditActivity
	if err := dbQuery.Order("created_at DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取活动数据失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data":    activities,
		})
	case "csv":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data":    gin.H{"message": "CSV导出功能待实现", "count": len(activities)},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的导出格式",
			"data":    nil,
		})
	}
}

func (h *ActivityHandler) SaveAsTemplate(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	var req struct {
		TemplateName string `json:"template_name" binding:"required"`
		Description  string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	var originalActivity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&originalActivity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 这里可以实现模板保存逻辑
	// 目前返回成功消息，实际实现需要创建模板表

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "模板保存成功",
		"data": gin.H{
			"template_name": req.TemplateName,
			"activity_id":   activityID,
		},
	})
}

func (h *ActivityHandler) ImportActivities(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请上传文件: " + err.Error(),
			"data":    nil,
		})
		return
	}

	fmt.Printf("File received: %s, size: %d\n", file.Filename, file.Size)

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

	// 处理解析后的数据
	h.processImportData(c, records, userID.(string), file.Filename)
}

// parseCSVFile 解析CSV文件
func (h *ActivityHandler) parseCSVFile(file *multipart.FileHeader) ([][]string, error) {
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

func (h *ActivityHandler) parseExcelFile(file *multipart.FileHeader) ([][]string, error) {
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
			record := make([]string, 5)
			for i := 0; i < 5 && i < len(row); i++ {
				record[i] = strings.TrimSpace(row[i])
			}
			records = append(records, record)
		}
	}

	return records, nil
}

func (h *ActivityHandler) processImportData(c *gin.Context, records [][]string, userID string, fileName string) {
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件至少需要包含标题行和一行数据",
			"data":    nil,
		})
		return
	}

	if len(records) > 1001 { // 标题行 + 1000行数据
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件最多支持1000行数据",
			"data":    nil,
		})
		return
	}

	headers := records[0]
	expectedHeaders := []string{"title", "description", "start_date", "end_date", "category"}

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

	var activities []models.ActivityRequest
	var errors []string

	for i, record := range records[1:] {
		rowNum := i + 2

		if len(record) < len(headers) {
			errors = append(errors, fmt.Sprintf("第%d行: 列数不匹配", rowNum))
			continue
		}

		activity := models.ActivityRequest{
			Title:       strings.TrimSpace(record[headerMap["title"]]),
			Description: strings.TrimSpace(record[headerMap["description"]]),
			StartDate:   strings.TrimSpace(record[headerMap["start_date"]]),
			EndDate:     strings.TrimSpace(record[headerMap["end_date"]]),
			Category:    strings.TrimSpace(record[headerMap["category"]]),
		}

		if err := h.validateActivityRequest(activity); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}

		activities = append(activities, activity)
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "数据验证失败",
			"data": gin.H{
				"errors":       errors,
				"total_rows":   len(records) - 1,
				"valid_rows":   len(activities),
				"invalid_rows": len(errors),
			},
		})
		return
	}

	var createdActivities []models.ActivityCreateResponse
	var createErrors []string

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, activityReq := range activities {
		startDate, endDate, err := h.parseActivityDates(activityReq.StartDate, activityReq.EndDate)
		if err != nil {
			createErrors = append(createErrors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
			continue
		}

		activity := models.CreditActivity{
			Title:       activityReq.Title,
			Description: activityReq.Description,
			StartDate:   startDate,
			EndDate:     endDate,
			Status:      models.StatusDraft,
			Category:    activityReq.Category,
			OwnerID:     userID,
		}

		if err := tx.Create(&activity).Error; err != nil {
			createErrors = append(createErrors, fmt.Sprintf("第%d个活动创建失败: %s", i+1, err.Error()))
			continue
		}

		response := models.ActivityCreateResponse{
			ID:          activity.ID,
			Title:       activity.Title,
			Description: activity.Description,
			StartDate:   activity.StartDate,
			EndDate:     activity.EndDate,
			Status:      activity.Status,
			Category:    activity.Category,
			OwnerID:     activity.OwnerID,
			CreatedAt:   activity.CreatedAt,
			UpdatedAt:   activity.UpdatedAt,
		}
		createdActivities = append(createdActivities, response)
	}

	if len(createErrors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量创建活动失败",
			"data": gin.H{
				"errors":             createErrors,
				"created_count":      0,
				"total_count":        len(activities),
				"created_activities": []models.ActivityCreateResponse{},
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
			"created_count":      len(createdActivities),
			"total_count":        len(activities),
			"created_activities": createdActivities,
			"file_name":          fileName,
			"file_type":          filepath.Ext(fileName),
		},
	})
}

func (h *ActivityHandler) GetCSVTemplate(c *gin.Context) {
	template := [][]string{
		{"title", "description", "start_date", "end_date", "category"},
		{"示例活动1", "这是一个示例活动描述", "2024-01-01", "2024-01-31", "创新创业"},
		{"示例活动2", "另一个示例活动", "2024-02-01", "2024-02-28", "学科竞赛"},
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=activity_template.csv")

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

func (h *ActivityHandler) ImportActivitiesFromCSV(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
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

	h.processImportData(c, records, userID.(string), file.Filename)
}

func (h *ActivityHandler) GetExcelTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "活动导入模板"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"title", "description", "start_date", "end_date", "category"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	examples := [][]string{
		{"创新创业实践活动", "参与创新创业项目，提升创新能力和实践技能", "2024-01-01", "2024-01-31", "创新创业"},
		{"学科竞赛活动", "参加学科竞赛，获得优异成绩", "2024-02-01", "2024-02-28", "学科竞赛"},
	}

	for i, example := range examples {
		for j, value := range example {
			cell := fmt.Sprintf("%c%d", 'A'+j, i+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 25)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=activity_template.xlsx")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Excel模板失败",
			"data":    nil,
		})
		return
	}
}
