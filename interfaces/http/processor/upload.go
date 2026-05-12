package processor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
)

type UploadHandler struct {
	db      *gorm.DB
	pool    *media.WorkerPool
	storage domainservice.StorageProvider
	cfg     config.ProcessorConfig
	bucket  string
}

func NewUploadHandler(db *gorm.DB, pool *media.WorkerPool, storage domainservice.StorageProvider, cfg config.ProcessorConfig, bucket string) *UploadHandler {
	return &UploadHandler{db: db, pool: pool, storage: storage, cfg: cfg, bucket: bucket}
}

func (h *UploadHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/input/upload", h.SimpleUpload)
	rg.POST("/input/upload/start", h.StartResumable)
	rg.POST("/input/upload/complete", h.CompleteResumable)
	rg.POST("/input/packet", h.Packet)
}

func (h *UploadHandler) SimpleUpload(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer func() { _ = file.Close() }()

	if header.Size > h.cfg.MaxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
		return
	}

	mediaType := c.PostForm("media_type")
	if mediaType == "" {
		mediaType = "image"
	}

	input := &entity.Input{
		ProjectID:        projectID,
		MediaType:        mediaType,
		Status:           "pending",
		OriginalFilename: header.Filename,
		Extension:        filepath.Ext(header.Filename),
	}
	if err := h.db.WithContext(c.Request.Context()).Create(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create input"})
		return
	}

	tmpDir := filepath.Join(h.cfg.TempDir, fmt.Sprintf("input_%d", input.ID))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp dir"})
		return
	}

	localPath := filepath.Join(tmpDir, header.Filename)
	dst, err := os.Create(localPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	_ = dst.Close()

	input.BlobPath = localPath
	if err := h.db.WithContext(c.Request.Context()).Save(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update input"})
		return
	}

	if err := h.pool.Submit(input); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "processing queue full"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"input_id": input.ID,
		"status":   "pending",
	})
}

func (h *UploadHandler) StartResumable(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var req struct {
		Filename  string `json:"filename" binding:"required"`
		Size      int64  `json:"size" binding:"required"`
		MediaType string `json:"media_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MediaType == "" {
		req.MediaType = "image"
	}

	input := &entity.Input{
		ProjectID:        projectID,
		MediaType:        req.MediaType,
		Status:           "pending",
		OriginalFilename: req.Filename,
		Extension:        filepath.Ext(req.Filename),
	}
	if err := h.db.WithContext(c.Request.Context()).Create(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create input"})
		return
	}

	blobKey := fmt.Sprintf("projects/%d/uploads/%d/%s", projectID, input.ID, req.Filename)
	uploadURL, err := h.storage.PresignPut(c.Request.Context(), h.bucket, blobKey, 1*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL"})
		return
	}

	input.BlobPath = blobKey
	if err := h.db.WithContext(c.Request.Context()).Save(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update input"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"input_id":   input.ID,
		"upload_url": uploadURL,
		"chunk_size": 5 * 1024 * 1024,
	})
}

func (h *UploadHandler) CompleteResumable(c *gin.Context) {
	var req struct {
		InputID int64 `json:"input_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input entity.Input
	if err := h.db.WithContext(c.Request.Context()).First(&input, req.InputID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "input not found"})
		return
	}

	if err := h.pool.Submit(&input); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "processing queue full"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"input_id": input.ID,
		"status":   "processing",
	})
}

func (h *UploadHandler) Packet(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var req struct {
		MediaType string `json:"media_type" binding:"required"`
		URL       string `json:"url" binding:"required"`
		Filename  string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Filename == "" {
		req.Filename = filepath.Base(req.URL)
	}

	input := &entity.Input{
		ProjectID:        projectID,
		MediaType:        req.MediaType,
		Type:             "from_url",
		Status:           "pending",
		OriginalFilename: req.Filename,
		Extension:        filepath.Ext(req.Filename),
		BlobPath:         req.URL,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create input"})
		return
	}

	if err := h.pool.Submit(input); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "processing queue full"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"input_id": input.ID,
		"status":   "pending",
	})
}
