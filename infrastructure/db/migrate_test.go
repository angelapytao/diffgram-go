package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
)

func TestRunMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}

	ctx := context.Background()
	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("diffgram_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "charset=utf8mb4&parseTime=True&loc=Local")
	require.NoError(t, err)

	sqlDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	err = diffdb.RunMigrations(sqlDB, migrationsDir)
	require.NoError(t, err, "goose up must succeed")

	coreTables := []string{
		"userbase", "member", "auth_api", "org",
		"project", "working_dir", "project_settings",
		"label", "label_schema", "label_schema_link",
		"role", "role_member_object",
	}
	for _, tbl := range coreTables {
		var count int
		row := sqlDB.QueryRow(
			"SELECT COUNT(*) FROM information_schema.tables "+
				"WHERE table_schema = DATABASE() AND table_name = ?", tbl,
		)
		require.NoError(t, row.Scan(&count))
		assert.Equal(t, 1, count, "table %q should exist after migration", tbl)
	}
}