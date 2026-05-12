package rpc

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
)

// ProcessorServiceHandler implements the ProcessorService RPC methods.
// It will be wired to Kitex-generated server code once codegen is available.
type ProcessorServiceHandler struct {
	db   *gorm.DB
	pool *media.WorkerPool
}

func NewProcessorServiceHandler(db *gorm.DB, pool *media.WorkerPool) *ProcessorServiceHandler {
	return &ProcessorServiceHandler{db: db, pool: pool}
}

// --- Request/Response types (mirrors IDL, used until Kitex codegen) ---

type ProcessMediaRequest struct {
	ProjectID int64             `json:"project_id"`
	InputID   int64             `json:"input_id"`
	MediaType string            `json:"media_type"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ProcessMediaResponse struct {
	InputID      int64  `json:"input_id"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type GetInputRequest struct {
	InputID int64 `json:"input_id"`
}

type InputInfo struct {
	ID               int64   `json:"id"`
	ProjectID        int64   `json:"project_id"`
	MediaType        string  `json:"media_type"`
	Status           string  `json:"status"`
	PercentComplete  float64 `json:"percent_complete,omitempty"`
	OriginalFilename string  `json:"original_filename,omitempty"`
	FileID           *int64  `json:"file_id,omitempty"`
}

type ListInputsRequest struct {
	ProjectID    int64  `json:"project_id"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	StatusFilter string `json:"status_filter,omitempty"`
}

type ListInputsResponse struct {
	Inputs   []InputInfo `json:"inputs"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ProcessMedia submits an input to the worker pool for async processing.
func (h *ProcessorServiceHandler) ProcessMedia(ctx context.Context, req *ProcessMediaRequest) (*ProcessMediaResponse, error) {
	var input entity.Input
	if err := h.db.WithContext(ctx).First(&input, req.InputID).Error; err != nil {
		return &ProcessMediaResponse{
			InputID:      req.InputID,
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("input not found: %v", err),
		}, nil
	}

	if err := h.pool.Submit(&input); err != nil {
		return &ProcessMediaResponse{
			InputID:      req.InputID,
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, nil
	}

	return &ProcessMediaResponse{
		InputID: req.InputID,
		Status:  "accepted",
	}, nil
}

// GetInput retrieves a single input by ID.
func (h *ProcessorServiceHandler) GetInput(ctx context.Context, req *GetInputRequest) (*InputInfo, error) {
	var input entity.Input
	if err := h.db.WithContext(ctx).First(&input, req.InputID).Error; err != nil {
		return nil, fmt.Errorf("input not found: %w", err)
	}

	return &InputInfo{
		ID:               input.ID,
		ProjectID:        input.ProjectID,
		MediaType:        input.MediaType,
		Status:           input.Status,
		PercentComplete:  input.PercentComplete,
		OriginalFilename: input.OriginalFilename,
		FileID:           input.FileID,
	}, nil
}

// ListInputs returns a paginated list of inputs for a project.
func (h *ProcessorServiceHandler) ListInputs(ctx context.Context, req *ListInputsRequest) (*ListInputsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 25
	}

	var total int64
	query := h.db.WithContext(ctx).Model(&entity.Input{}).Where("project_id = ?", req.ProjectID)
	if req.StatusFilter != "" {
		query = query.Where("status = ?", req.StatusFilter)
	}
	query.Count(&total)

	var inputs []entity.Input
	query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order("id DESC").Find(&inputs)

	items := make([]InputInfo, 0, len(inputs))
	for _, inp := range inputs {
		items = append(items, InputInfo{
			ID:               inp.ID,
			ProjectID:        inp.ProjectID,
			MediaType:        inp.MediaType,
			Status:           inp.Status,
			PercentComplete:  inp.PercentComplete,
			OriginalFilename: inp.OriginalFilename,
			FileID:           inp.FileID,
		})
	}

	return &ListInputsResponse{
		Inputs:   items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
