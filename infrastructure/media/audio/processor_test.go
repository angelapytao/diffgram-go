package audio

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
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
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.AudioFile{}, &entity.File{}))
	return db
}

func TestProcessor_Process_WithFFprobe(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available for test audio generation")
	}

	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{
		FFprobePath: "ffprobe",
		TempDir:     t.TempDir(),
	}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "test.wav")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", audioPath)
	require.NoError(t, cmd.Run())

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "audio",
		Status:           "processing",
		OriginalFilename: "test.wav",
		BlobPath:         audioPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	var af entity.AudioFile
	require.NoError(t, db.First(&af).Error)
	require.Greater(t, af.Duration, 0.0)
	require.Greater(t, af.SampleRate, 0)
	require.Greater(t, af.Channels, 0)

	require.Len(t, store.uploaded, 1)
}
