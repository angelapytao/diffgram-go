package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
)

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _, _ string, _ any) error { return nil }

type mockStorage struct{}

func (m *mockStorage) Put(_ context.Context, _, _ string, _ io.Reader) error { return nil }
func (m *mockStorage) Get(_ context.Context, _, _ string) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockStorage) Delete(_ context.Context, _, _ string) error { return nil }
func (m *mockStorage) PresignGet(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://example.com/get", nil
}
func (m *mockStorage) PresignPut(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://example.com/put", nil
}

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}))

	pub := &mockPublisher{}
	pipeline := media.NewPipeline(db, pub, map[string]media.MediaProcessor{})
	pool := media.NewWorkerPool(2, pipeline)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	pool.Start(ctx)

	cfg := config.ProcessorConfig{
		MaxUploadSize: 10 * 1024 * 1024,
		TempDir:       t.TempDir(),
	}

	handler := NewUploadHandler(db, pool, &mockStorage{}, cfg, "test-bucket")

	r := gin.New()
	rg := r.Group("/api/walrus/v1/project/:project_id")
	handler.RegisterRoutes(rg)
	return r, db
}

func TestSimpleUpload(t *testing.T) {
	r, db := setupTestRouter(t)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.png")
	require.NoError(t, err)
	_, _ = part.Write([]byte("fake image data"))
	_ = writer.WriteField("media_type", "image")
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/walrus/v1/project/1/input/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotZero(t, resp["input_id"])

	var input entity.Input
	require.NoError(t, db.First(&input).Error)
	require.Equal(t, "image", input.MediaType)
}

func TestStartResumable(t *testing.T) {
	r, _ := setupTestRouter(t)

	body, _ := json.Marshal(map[string]any{
		"filename":   "big_video.mp4",
		"size":       1024 * 1024 * 100,
		"media_type": "video",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/walrus/v1/project/1/input/upload/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotZero(t, resp["input_id"])
	require.NotEmpty(t, resp["upload_url"])
}
