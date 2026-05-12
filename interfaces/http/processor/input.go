package processor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type InputHandler struct {
	db *gorm.DB
}

func NewInputHandler(db *gorm.DB) *InputHandler {
	return &InputHandler{db: db}
}

func (h *InputHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/input/list", h.List)
	rg.GET("/input/:input_id", h.Get)
	rg.PATCH("/input/:input_id", h.Update)
}

func (h *InputHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}

	statusFilter := c.Query("status")

	var total int64
	query := h.db.WithContext(c.Request.Context()).Model(&entity.Input{}).Where("project_id = ?", projectID)
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	query.Count(&total)

	var inputs []entity.Input
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&inputs)

	type inputResp struct {
		ID               int64   `json:"id"`
		ProjectID        int64   `json:"project_id"`
		MediaType        string  `json:"media_type"`
		Status           string  `json:"status"`
		PercentComplete  float64 `json:"percent_complete"`
		OriginalFilename string  `json:"original_filename"`
		FileID           *int64  `json:"file_id"`
	}

	items := make([]inputResp, 0, len(inputs))
	for _, inp := range inputs {
		items = append(items, inputResp{
			ID:               inp.ID,
			ProjectID:        inp.ProjectID,
			MediaType:        inp.MediaType,
			Status:           inp.Status,
			PercentComplete:  inp.PercentComplete,
			OriginalFilename: inp.OriginalFilename,
			FileID:           inp.FileID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"inputs": items,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *InputHandler) Get(c *gin.Context) {
	inputID, err := strconv.ParseInt(c.Param("input_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input_id"})
		return
	}

	var input entity.Input
	if err := h.db.WithContext(c.Request.Context()).First(&input, inputID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "input not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                input.ID,
		"project_id":        input.ProjectID,
		"media_type":        input.MediaType,
		"status":            input.Status,
		"status_text":       input.StatusText,
		"percent_complete":  input.PercentComplete,
		"original_filename": input.OriginalFilename,
		"file_id":           input.FileID,
		"time_created":      input.TimeCreated,
		"time_updated":      input.TimeUpdated,
	})
}

func (h *InputHandler) Update(c *gin.Context) {
	inputID, err := strconv.ParseInt(c.Param("input_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input_id"})
		return
	}

	var input entity.Input
	if err := h.db.WithContext(c.Request.Context()).First(&input, inputID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "input not found"})
		return
	}

	var req struct {
		Status     *string `json:"status"`
		StatusText *string `json:"status_text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != nil {
		input.Status = *req.Status
	}
	if req.StatusText != nil {
		input.StatusText = *req.StatusText
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update input"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     input.ID,
		"status": input.Status,
	})
}
