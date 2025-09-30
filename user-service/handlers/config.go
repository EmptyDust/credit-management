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

// GetOptions 读取配置文件并返回前端所需的下拉选项
func GetOptions(c *gin.Context) {
	configPath := os.Getenv("OPTIONS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/options.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to read options config", "error": err.Error()})
		return
	}

	var options OptionsResponse
	if err := json.Unmarshal(data, &options); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "invalid options config", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": options})
}
