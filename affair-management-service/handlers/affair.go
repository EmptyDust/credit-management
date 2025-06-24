package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"credit-management/affair-management-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AffairHandler struct {
	db *gorm.DB
}

func NewAffairHandler(db *gorm.DB) *AffairHandler {
	return &AffairHandler{db: db}
}

// CreateAffair 创建事项（支持参与同学、描述、附件）
func (h *AffairHandler) CreateAffair(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		CreatorID   string   `json:"creator_id"`
		Participants []string `json:"participants"` // 学号数组
		Attachments string   `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	affair := models.Affair{
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   req.CreatorID,
		Attachments: req.Attachments,
	}
	if err := h.db.Create(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建事项失败: " + err.Error()})
		return
	}

	// 批量插入参与同学
	for _, sid := range req.Participants {
		affairStudent := models.AffairStudent{
			AffairID:  affair.ID,
			StudentID: sid,
			IsPrimary: sid == req.CreatorID,
			Role:      func() string { if sid == req.CreatorID { return "primary" } else { return "member" } }(),
		}
		h.db.Create(&affairStudent)
	}

	// 调用 application-management-service 批量生成申请
	appServiceURL := getEnv("APPLICATION_SERVICE_URL", "http://application-management-service:8082")
	url := fmt.Sprintf("%s/api/applications/batch", appServiceURL)
	
	batchReq := map[string]interface{}{
		"affair_id":     affair.ID,
		"creator_id":    req.CreatorID,
		"participants":  req.Participants,
	}
	
	jsonData, err := json.Marshal(batchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化批量创建请求失败: " + err.Error()})
		return
	}
	
	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建批量申请请求失败: " + err.Error()})
		return
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	// 转发认证信息
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}
	if userID := c.GetHeader("X-User-Id"); userID != "" {
		httpReq.Header.Set("X-User-Id", userID)
	}
	
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		// 如果调用失败，记录错误但不影响事务创建
		fmt.Printf("Warning: Failed to create applications for affair %d: %v\n", affair.ID, err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			fmt.Printf("Warning: Application service returned status %d for affair %d\n", resp.StatusCode, affair.ID)
		}
	}

	c.JSON(http.StatusCreated, affair)
}

// GetAffair 返回详情（含参与者）
func (h *AffairHandler) GetAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}
	var affair models.Affair
	if err := h.db.First(&affair, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "事项不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事项失败: " + err.Error()})
		}
		return
	}
	var participants []models.AffairStudent
	h.db.Where("affair_id = ?", id).Find(&participants)
	c.JSON(http.StatusOK, gin.H{"affair": affair, "participants": participants})
}

// UpdateAffair 仅允许创建者编辑
func (h *AffairHandler) UpdateAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}
	var affair models.Affair
	if err := h.db.First(&affair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "事项不存在"})
		return
	}
	// 权限校验（假设从header获取user_id）
	userID := c.GetHeader("X-User-Id")
	if userID == "" || userID != affair.CreatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有创建者可以编辑该事务"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Attachments string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}
	affair.Name = req.Name
	affair.Description = req.Description
	affair.Attachments = req.Attachments
	if err := h.db.Save(&affair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新事项失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, affair)
}

// DeleteAffair 删除事项
func (h *AffairHandler) DeleteAffair(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}

	if err := h.db.Delete(&models.Affair{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "事项删除成功"})
}

// GetAllAffairs 获取所有事项
func (h *AffairHandler) GetAllAffairs(c *gin.Context) {
	var affairs []models.Affair
	if err := h.db.Find(&affairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询所有事项失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, affairs)
}

// GetAffairParticipants 获取事务参与者
func (h *AffairHandler) GetAffairParticipants(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的事项ID"})
		return
	}
	var participants []models.AffairStudent
	h.db.Where("affair_id = ?", id).Find(&participants)
	c.JSON(http.StatusOK, participants)
}

// GetAffairApplications 获取事务下所有申请（调用 application-management-service）
func (h *AffairHandler) GetAffairApplications(c *gin.Context) {
	id := c.Param("id")
	
	// 验证事务是否存在
	var affair models.Affair
	if err := h.db.First(&affair, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "事务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事务失败: " + err.Error()})
		}
		return
	}

	// 调用 application-management-service 获取该事务下的所有申请
	appServiceURL := getEnv("APPLICATION_SERVICE_URL", "http://application-management-service:8082")
	url := fmt.Sprintf("%s/api/applications?affair_id=%s", appServiceURL, id)
	
	// 转发认证头
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败: " + err.Error()})
		return
	}
	
	// 转发认证信息
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	if userID := c.GetHeader("X-User-Id"); userID != "" {
		req.Header.Set("X-User-Id", userID)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用申请服务失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败: " + err.Error()})
		return
	}
	
	// 转发响应状态码和内容
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := getEnvFromContext(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvFromContext 从上下文获取环境变量（简化实现）
func getEnvFromContext(key string) string {
	// 这里可以扩展为从配置或环境变量获取
	// 暂时返回空字符串，使用默认值
	return ""
}
