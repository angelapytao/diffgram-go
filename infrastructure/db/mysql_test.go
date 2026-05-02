package db_test

import (
	"context"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
)

func TestNewConnection_Ping(t *testing.T) {
	ctx := context.Background()

	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("diffgram_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "charset=utf8mb4&parseTime=True&loc=Local")
	require.NoError(t, err)

	gdb, err := diffdb.NewConnection(dsn)
	require.NoError(t, err)

	sqlDB, err := gdb.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
}
