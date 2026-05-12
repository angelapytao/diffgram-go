package processor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

func setupInputTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}))

	handler := NewInputHandler(db)
	r := gin.New()
	rg := r.Group("/api/walrus/v1/project/:project_id")
	handler.RegisterRoutes(rg)
	return r, db
}

func TestInputList(t *testing.T) {
	r, db := setupInputTestRouter(t)

	// Create test inputs
	for i := 0; i < 3; i++ {
		db.Create(&entity.Input{ProjectID: 1, MediaType: "image", Status: "success"})
	}
	db.Create(&entity.Input{ProjectID: 1, MediaType: "video", Status: "failed"})
	db.Create(&entity.Input{ProjectID: 2, MediaType: "image", Status: "success"})

	req := httptest.NewRequest(http.MethodGet, "/api/walrus/v1/project/1/input/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	inputs := resp["inputs"].([]any)
	require.Len(t, inputs, 4) // all project 1 inputs

	meta := resp["meta"].(map[string]any)
	require.Equal(t, float64(4), meta["total"])
}

func TestInputList_StatusFilter(t *testing.T) {
	r, db := setupInputTestRouter(t)

	db.Create(&entity.Input{ProjectID: 1, MediaType: "image", Status: "success"})
	db.Create(&entity.Input{ProjectID: 1, MediaType: "image", Status: "failed"})

	req := httptest.NewRequest(http.MethodGet, "/api/walrus/v1/project/1/input/list?status=failed", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	inputs := resp["inputs"].([]any)
	require.Len(t, inputs, 1)
}

func TestInputGet(t *testing.T) {
	r, db := setupInputTestRouter(t)

	input := &entity.Input{ProjectID: 1, MediaType: "image", Status: "success", OriginalFilename: "test.png"}
	db.Create(input)

	req := httptest.NewRequest(http.MethodGet, "/api/walrus/v1/project/1/input/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, float64(1), resp["id"])
	require.Equal(t, "image", resp["media_type"])
}

func TestInputUpdate(t *testing.T) {
	r, db := setupInputTestRouter(t)

	input := &entity.Input{ProjectID: 1, MediaType: "image", Status: "pending"}
	db.Create(input)

	body := `{"status": "cancelled", "status_text": "user cancelled"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/walrus/v1/project/1/input/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var updated entity.Input
	db.First(&updated, 1)
	require.Equal(t, "cancelled", updated.Status)
	require.Equal(t, "user cancelled", updated.StatusText)
}
