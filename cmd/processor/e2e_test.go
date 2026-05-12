package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
	imageproc "github.com/angelapytao/diffgram-go/infrastructure/media/image"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	httpproc "github.com/angelapytao/diffgram-go/interfaces/http/processor"
)

// memStorage implements domainservice.StorageProvider in-memory for testing.
type memStorage struct {
	data map[string][]byte
}

func newMemStorage() *memStorage {
	return &memStorage{data: make(map[string][]byte)}
}

func (m *memStorage) Put(_ context.Context, _, key string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.data[key] = b
	return nil
}

func (m *memStorage) Get(_ context.Context, _, _ string) (io.ReadCloser, error) { return nil, nil }
func (m *memStorage) Delete(_ context.Context, _, _ string) error               { return nil }
func (m *memStorage) PresignGet(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://example.com/get", nil
}
func (m *memStorage) PresignPut(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://example.com/put", nil
}

func TestE2E_SimpleUpload_Image(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test: requires Docker")
	}

	ctx := context.Background()

	// SQLite in-memory DB — avoids Docker MySQL startup time
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.Image{}, &entity.File{}))

	// RabbitMQ container
	rmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.13-management")
	require.NoError(t, err)
	t.Cleanup(func() { _ = rmqContainer.Terminate(ctx) })

	rmqURL, err := rmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	// Declare the events exchange so publishing doesn't fail
	bootstrapConn, err := amqp.Dial(rmqURL)
	require.NoError(t, err)
	bootstrapCh, err := bootstrapConn.Channel()
	require.NoError(t, err)
	require.NoError(t, bootstrapCh.ExchangeDeclare("events", "topic", true, false, false, false, nil))
	_ = bootstrapCh.Close()
	_ = bootstrapConn.Close()

	mqClient, err := mq.NewClient(rmqURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mqClient.Close() })

	publisher, err := mq.NewPublisher(mqClient)
	require.NoError(t, err)
	t.Cleanup(func() { _ = publisher.Close() })

	// Setup pipeline with image processor only
	store := newMemStorage()
	cfg := config.ProcessorConfig{
		ThumbLargeSize: 800,
		ThumbSmallSize: 200,
		TempDir:        t.TempDir(),
		MaxUploadSize:  10 * 1024 * 1024,
	}

	imgProc := imageproc.NewProcessor(store, db, cfg, "test-bucket")
	processors := map[string]media.MediaProcessor{
		"image": imgProc,
	}

	pipeline := media.NewPipeline(db, publisher, processors)
	pool := media.NewWorkerPool(2, pipeline)

	poolCtx, poolCancel := context.WithCancel(ctx)
	defer poolCancel()
	pool.Start(poolCtx)

	// Setup HTTP
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uploadHandler := httpproc.NewUploadHandler(db, pool, store, cfg, "test-bucket")
	rg := r.Group("/api/walrus/v1/project/:project_id")
	uploadHandler.RegisterRoutes(rg)

	// Create a test PNG image (50x50 red square)
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	require.NoError(t, png.Encode(&imgBuf, img))

	// Upload via multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "red_square.png")
	require.NoError(t, err)
	_, err = io.Copy(part, &imgBuf)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("media_type", "image"))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/walrus/v1/project/1/input/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	inputID := int64(resp["input_id"].(float64))
	require.Greater(t, inputID, int64(0))

	// Wait for processing to complete
	var input entity.Input
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.First(&input, inputID).Error; err == nil && input.Status == "success" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	require.Equal(t, "success", input.Status)
	require.NotNil(t, input.FileID)

	// Verify Image record
	var imgRecord entity.Image
	require.NoError(t, db.First(&imgRecord).Error)
	require.Equal(t, 50, imgRecord.Width)
	require.Equal(t, 50, imgRecord.Height)
	require.NotEmpty(t, imgRecord.URLSignedBlobPath)
	require.NotEmpty(t, imgRecord.ThumbLargeBlobPath)
	require.NotEmpty(t, imgRecord.ThumbSmallBlobPath)

	// Verify File record
	var fileRecord entity.File
	require.NoError(t, db.First(&fileRecord).Error)
	require.Equal(t, "image", fileRecord.Type)
	require.Equal(t, int64(1), fileRecord.ProjectID)

	// Verify storage uploads (original + 2 thumbnails)
	require.GreaterOrEqual(t, len(store.data), 3)

	fmt.Printf("E2E passed: input_id=%d, image_id=%d, file_id=%d, blobs=%d\n",
		inputID, imgRecord.ID, fileRecord.ID, len(store.data))
}
