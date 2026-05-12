package video

import (
	"context"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
)

type mockStorage struct {
	uploaded map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{uploaded: make(map[string][]byte)}
}

func (m *mockStorage) Put(_ context.Context, _, key string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.uploaded[key] = data
	return nil
}

func (m *mockStorage) Get(_ context.Context, _, _ string) (io.ReadCloser, error) { return nil, nil }
func (m *mockStorage) Delete(_ context.Context, _, _ string) error               { return nil }
func (m *mockStorage) PresignGet(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "", nil
}
func (m *mockStorage) PresignPut(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "", nil
}

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.Video{}, &entity.File{}))
	return db
}

func TestParseFPS(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"30/1", 30.0},
		{"30000/1001", 29.97002997002997},
		{"25/1", 25.0},
		{"24", 24.0},
	}
	for _, tt := range tests {
		got := parseFPS(tt.input)
		require.InDelta(t, tt.expected, got, 0.01, "input: %s", tt.input)
	}
}

func TestProcessor_Process_WithFFprobe(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available for test video generation")
	}

	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{
		FFprobePath:     "ffprobe",
		VideoFPSDefault: 5,
		TempDir:         t.TempDir(),
	}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	// Create a minimal test video
	tmpDir := t.TempDir()
	videoPath := tmpDir + "/test.mp4"
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=red:s=64x64:d=1", "-c:v", "libx264", "-t", "1", videoPath)
	require.NoError(t, cmd.Run())

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "video",
		Status:           "processing",
		OriginalFilename: "test.mp4",
		BlobPath:         videoPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	var video entity.Video
	require.NoError(t, db.First(&video).Error)
	require.Equal(t, 64, video.Width)
	require.Equal(t, 64, video.Height)
	require.Greater(t, video.Duration, 0.0)
	require.Equal(t, "complete", video.Status)
}
