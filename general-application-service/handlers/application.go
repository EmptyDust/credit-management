package handlers

import (
    "net/http"
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
    "general-application-service/models"
    "gorm.io/gorm"
)

type ApplicationHandler struct {
    DB *gorm.DB
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
    var app models.Application
    if err := c.ShouldBindJSON(&app); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    app.ApplyTime = time.Now()
    if err := h.DB.Create(&app).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, app)
}

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var app models.Application
    if err := h.DB.First(&app, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, app)
}

func (h *ApplicationHandler) UpdateStatus(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var req struct {
        Status string `json:"status"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.DB.Model(&models.Application{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *ApplicationHandler) UpdateFinalCredit(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var req struct {
        FinalCredit float64 `json:"final_credit"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.DB.Model(&models.Application{}).Where("id = ?", id).Update("final_credit", req.FinalCredit).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
} 