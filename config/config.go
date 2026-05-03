package config

import (
	"os"
	"strconv"
	"time"
)

type TOSConfig struct {
	Host      string
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
}

type JWTConfig struct {
	Secret  string
	Timeout time.Duration
}

type Config struct {
	ServerPort  int
	DBDsn       string
	MQHost      string
	MQPort      int
	RabbitMQURL string
	Mode        string
	TOS         TOSConfig
	JWT         JWTConfig
}

func Load() *Config {
	return &Config{
		ServerPort:  envInt("SERVER_PORT", 8080),
		DBDsn:       os.Getenv("DATABASE_URL"),
		MQHost:      envStr("RABBITMQ_HOST", "localhost"),
		MQPort:      envInt("RABBITMQ_PORT", 5672),
		RabbitMQURL: envStr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Mode:        envStr("DIFFGRAM_SYSTEM_MODE", "sandbox"),
		TOS: TOSConfig{
			Host:      envStr("TOS_HOST", "tos-cn-beijing.volces.com"),
			AccessKey: os.Getenv("TOS_ACCESS_KEY"),
			SecretKey: os.Getenv("TOS_SECRET_KEY"),
			Region:    envStr("TOS_REGION", "cn-beijing"),
			Bucket:    os.Getenv("TOS_BUCKET"),
		},
		JWT: JWTConfig{
			Secret:  envStr("JWT_SECRET", "change-me-in-production"),
			Timeout: envDuration("JWT_TIMEOUT", 24*time.Hour),
		},
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
