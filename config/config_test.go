package config_test

import (
	"testing"
	"time"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DIFFGRAM_SYSTEM_MODE", "")
	t.Setenv("RABBITMQ_HOST", "")
	t.Setenv("RABBITMQ_PORT", "")

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

func TestConfig_TOSDefaults(t *testing.T) {
	t.Setenv("TOS_HOST", "")
	t.Setenv("TOS_REGION", "")
	cfg := config.Load()
	assert.Equal(t, "tos-cn-beijing.volces.com", cfg.TOS.Host)
	assert.Equal(t, "cn-beijing", cfg.TOS.Region)
}

func TestConfig_JWTDefaults(t *testing.T) {
	t.Setenv("JWT_TIMEOUT", "")
	cfg := config.Load()
	assert.Equal(t, 24*time.Hour, cfg.JWT.Timeout)
}

func TestConfig_RabbitMQDefault(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "")
	cfg := config.Load()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
}
