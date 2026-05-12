package geotiff

import (
	"context"
	"io"
	"os"
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
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.File{}))
	return db
}

func TestProcessor_Process_NoGDAL(t *testing.T) {
	// Test that the processor returns an error when gdalinfo is unavailable
	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{
		GDALInfoPath:      "false", // will fail with exit code 1
		GDALTranslatePath: "false",
		TempDir:           t.TempDir(),
	}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	tmpDir := t.TempDir()
	tiffPath := filepath.Join(tmpDir, "test.tif")
	require.NoError(t, os.WriteFile(tiffPath, []byte("fake tiff data"), 0644))

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "geo_tiff",
		Status:           "processing",
		OriginalFilename: "test.tif",
		BlobPath:         tiffPath,
	}
	require.NoError(t, db.Create(input).Error)

	// This will fail at gdalinfo step since "false" returns exit code 1
	err := proc.Process(context.Background(), input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "gdalinfo")
}

func TestProcessor_Process_WithGDAL(t *testing.T) {
	if _, err := exec.LookPath("gdalinfo"); err != nil {
		t.Skip("gdalinfo not available")
	}

	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{
		GDALInfoPath:      "gdalinfo",
		GDALTranslatePath: "gdal_translate",
		TempDir:           t.TempDir(),
	}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	// We need a real GeoTIFF for this test - skip if we can't create one
	tmpDir := t.TempDir()
	tiffPath := filepath.Join(tmpDir, "test.tif")

	// Create a minimal TIFF using gdal_create if available
	if _, err := exec.LookPath("gdal_create"); err != nil {
		t.Skip("gdal_create not available for test fixture generation")
	}
	cmd := exec.Command("gdal_create", "-outsize", "4", "4", "-bands", "1", tiffPath)
	if err := cmd.Run(); err != nil {
		t.Skip("failed to create test GeoTIFF")
	}

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "geo_tiff",
		Status:           "processing",
		OriginalFilename: "test.tif",
		BlobPath:         tiffPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	// At minimum, original was uploaded
	require.GreaterOrEqual(t, len(store.uploaded), 1)

	var file entity.File
	require.NoError(t, db.First(&file).Error)
	require.Equal(t, "geo_tiff", file.Type)
}

func TestProcessor_MediaType(t *testing.T) {
	proc := &Processor{}
	require.Equal(t, "geo_tiff", proc.MediaType())
}
