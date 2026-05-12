package image

import (
	"context"
	"image"
	"image/color"
	"image/png"
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
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.Image{}, &entity.File{}))
	return db
}

func createTestImage(t *testing.T, dir string) string {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	path := filepath.Join(dir, "test.png")
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, img))
	require.NoError(t, f.Close())
	return path
}

func TestProcessor_Process(t *testing.T) {
	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{
		ThumbLargeSize: 800,
		ThumbSmallSize: 200,
		TempDir:        t.TempDir(),
	}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	tmpDir := t.TempDir()
	imgPath := createTestImage(t, tmpDir)

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "image",
		Status:           "processing",
		OriginalFilename: "test.png",
		BlobPath:         imgPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	// Verify storage uploads (original + 2 thumbnails)
	require.Len(t, store.uploaded, 3)

	// Verify DB records
	var imgRecord entity.Image
	require.NoError(t, db.First(&imgRecord).Error)
	require.Equal(t, 100, imgRecord.Width)
	require.Equal(t, 100, imgRecord.Height)

	var file entity.File
	require.NoError(t, db.First(&file).Error)
	require.Equal(t, "image", file.Type)
	require.Equal(t, &imgRecord.ID, file.ImageID)
}
