package sensorfusion

import (
	"context"
	"encoding/json"
	"io"
	"os"
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
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.File{}))
	return db
}

func TestProcessor_Process(t *testing.T) {
	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{TempDir: t.TempDir()}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	tmpDir := t.TempDir()

	// Create sub-files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "cloud.pcd"), []byte("pcd data"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "camera.png"), []byte("png data"), 0644))

	// Create manifest
	manifest := Manifest{
		Files: []ManifestFile{
			{Filename: "cloud.pcd", Type: "point_cloud"},
			{Filename: "camera.png", Type: "image"},
		},
	}
	manifestData, _ := json.Marshal(manifest)
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	require.NoError(t, os.WriteFile(manifestPath, manifestData, 0644))

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "sensor_fusion",
		Status:           "processing",
		OriginalFilename: "manifest.json",
		BlobPath:         manifestPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	// 2 files uploaded
	require.Len(t, store.uploaded, 2)

	// 2 File records created
	var files []entity.File
	require.NoError(t, db.Find(&files).Error)
	require.Len(t, files, 2)
	require.Equal(t, "point_cloud", files[0].Type)
	require.Equal(t, "image", files[1].Type)
}
