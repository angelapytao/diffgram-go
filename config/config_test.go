package config_test

import (
	"os"
	"testing"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DIFFGRAM_SYSTEM_MODE")
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")

	cfg := config.Load()

	assert.Equal(t, 8080, cfg.ServerPort)
	assert.Equal(t, "sandbox", cfg.Mode)
	assert.Equal(t, "localhost", cfg.MQHost)
	assert.Equal(t, 5672, cfg.MQPort)
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DIFFGRAM_SYSTEM_MODE", "production")
	t.Setenv("RABBITMQ_HOST", "rabbitmq.internal")
	t.Setenv("DATABASE_URL", "root:pass@tcp(localhost:3306)/diffgram")

	cfg := config.Load()

	assert.Equal(t, 9090, cfg.ServerPort)
	assert.Equal(t, "production", cfg.Mode)
	assert.Equal(t, "rabbitmq.internal", cfg.MQHost)
	assert.Equal(t, "root:pass@tcp(localhost:3306)/diffgram", cfg.DBDsn)
}
