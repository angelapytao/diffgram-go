package text

import (
	"context"
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
	require.NoError(t, db.AutoMigrate(&entity.Input{}, &entity.TextFile{}, &entity.File{}))
	return db
}

func TestProcessor_Process(t *testing.T) {
	db := setupDB(t)
	store := newMockStorage()
	cfg := config.ProcessorConfig{TempDir: t.TempDir()}

	proc := NewProcessor(store, db, cfg, "test-bucket")

	tmpDir := t.TempDir()
	textPath := filepath.Join(tmpDir, "sample.txt")
	require.NoError(t, os.WriteFile(textPath, []byte("hello world foo bar baz"), 0644))

	input := &entity.Input{
		ProjectID:        1,
		MediaType:        "text",
		Status:           "processing",
		OriginalFilename: "sample.txt",
		BlobPath:         textPath,
	}
	require.NoError(t, db.Create(input).Error)

	err := proc.Process(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, input.FileID)

	var tf entity.TextFile
	require.NoError(t, db.First(&tf).Error)
	require.Equal(t, 5, tf.TokenCount)
	require.NotEmpty(t, tf.BlobPath)

	require.Len(t, store.uploaded, 1)
}
