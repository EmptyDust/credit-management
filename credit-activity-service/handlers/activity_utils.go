package handlers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func (h *ActivityHandler) CopyActivity(c *gin.Context) {
	activityID := c.Param("id")
	if err := h.validator.ValidateUUID(activityID); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	// 使用数据库基类获取活动
	originalActivity, err := h.base.GetActivityByID(activityID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
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
		OwnerID:     userID,
	}

	if err := h.db.Create(&newActivity).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := h.enrichActivityResponse(newActivity, "")
	utils.SendCreatedResponse(c, "活动复制成功", response)
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
		if parsedDate, err := utils.ParseDate(startDate); err == nil {
			dbQuery = dbQuery.Where("start_date >= ?", parsedDate)
		}
	}
	if endDate != "" {
		if parsedDate, err := utils.ParseDate(endDate); err == nil {
			dbQuery = dbQuery.Where("end_date <= ?", parsedDate)
		}
	}

	var activities []models.CreditActivity
	if err := dbQuery.Order("created_at DESC").Find(&activities).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "json":
		utils.SendSuccessResponse(c, activities)
	case "csv":
		utils.SendSuccessResponse(c, gin.H{"message": "CSV导出功能待实现", "count": len(activities)})
	default:
		utils.SendBadRequest(c, "不支持的导出格式")
	}
}

func (h *ActivityHandler) SaveAsTemplate(c *gin.Context) {
	activityID := c.Param("id")
	if err := h.validator.ValidateUUID(activityID); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	var req struct {
		TemplateName string `json:"template_name" binding:"required"`
		Description  string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 使用数据库基类检查活动是否存在
	if err := h.base.CheckActivityExists(activityID); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 这里可以实现模板保存逻辑
	// 目前返回成功消息，实际实现需要创建模板表
	utils.SendSuccessResponse(c, gin.H{
		"template_name": req.TemplateName,
		"activity_id":   activityID,
	})
}

func (h *ActivityHandler) ImportActivities(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendBadRequest(c, "请选择要导入的文件")
		return
	}

	// 验证文件类型
	allowedTypes := []string{"csv", "xlsx", "xls"}
	if err := h.validator.ValidateFileType(file.Filename, allowedTypes); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证文件大小 (10MB)
	maxSize := int64(10 * 1024 * 1024)
	if err := h.validator.ValidateFileSize(file.Size, maxSize); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	var records [][]string
	fileExt := strings.ToLower(filepath.Ext(file.Filename))

	switch fileExt {
	case ".csv":
		records, err = h.parseCSVFile(file)
	case ".xlsx", ".xls":
		records, err = h.parseExcelFile(file)
	default:
		utils.SendBadRequest(c, "不支持的文件格式")
		return
	}

	if err != nil {
		utils.SendBadRequest(c, "文件解析失败: "+err.Error())
		return
	}

	// 处理解析后的数据
	h.processImportData(c, records, userID, file.Filename)
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
		utils.SendBadRequest(c, "文件至少需要包含标题行和一行数据")
		return
	}

	if len(records) > 1001 { // 标题行 + 1000行数据
		utils.SendBadRequest(c, "文件最多支持1000行数据")
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
		utils.SendBadRequest(c, "文件缺少必需的列: "+strings.Join(missingHeaders, ", "))
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
		utils.SendBadRequestWithData(c, "数据验证失败", gin.H{
			"errors":       errors,
			"total_rows":   len(records) - 1,
			"valid_rows":   len(activities),
			"invalid_rows": len(errors),
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
		startDate, endDate, err := utils.ParseDateRange(activityReq.StartDate, activityReq.EndDate)
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
			Category:    activity.Category,
			Status:      activity.Status,
			CreatedAt:   activity.CreatedAt,
		}

		createdActivities = append(createdActivities, response)
	}

	if len(createErrors) > 0 {
		tx.Rollback()
		utils.SendBadRequestWithData(c, "部分活动创建失败", gin.H{
			"errors":           createErrors,
			"total_activities": len(activities),
			"created":          len(createdActivities),
			"failed":           len(createErrors),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"message":            "批量导入成功",
		"total_activities":   len(activities),
		"created_activities": createdActivities,
		"file_name":          fileName,
	})
}

func (h *ActivityHandler) GetCSVTemplate(c *gin.Context) {
	headers := []string{"title", "description", "start_date", "end_date", "category"}
	sampleData := []string{"示例活动", "这是一个示例活动", "2024-01-01", "2024-12-31", "创新创业实践活动"}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=activity_template.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入标题行
	writer.Write(headers)
	// 写入示例数据
	writer.Write(sampleData)
}

func (h *ActivityHandler) ImportActivitiesFromCSV(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendBadRequest(c, "请选择要导入的CSV文件")
		return
	}

	// 验证文件类型
	if err := h.validator.ValidateFileType(file.Filename, []string{"csv"}); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 验证文件大小 (5MB)
	maxSize := int64(5 * 1024 * 1024)
	if err := h.validator.ValidateFileSize(file.Size, maxSize); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	records, err := h.parseCSVFile(file)
	if err != nil {
		utils.SendBadRequest(c, "CSV文件解析失败: "+err.Error())
		return
	}

	h.processImportData(c, records, userID, file.Filename)
}

func (h *ActivityHandler) GetExcelTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	// 设置标题行
	headers := []string{"title", "description", "start_date", "end_date", "category"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	// 设置示例数据
	sampleData := []string{"示例活动", "这是一个示例活动", "2024-01-01", "2024-12-31", "创新创业实践活动"}
	for i, data := range sampleData {
		cell := fmt.Sprintf("%c2", 'A'+i)
		f.SetCellValue("Sheet1", cell, data)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=activity_template.xlsx")

	if err := f.Write(c.Writer); err != nil {
		utils.SendInternalServerError(c, err)
		return
	}
}
