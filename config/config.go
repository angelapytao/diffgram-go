package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort int
	DBDsn      string
	MQHost     string
	MQPort     int
	Mode       string
}

func Load() *Config {
	return &Config{
		ServerPort: envInt("SERVER_PORT", 8080),
		DBDsn:      os.Getenv("DATABASE_URL"),
		MQHost:     envStr("RABBITMQ_HOST", "localhost"),
		MQPort:     envInt("RABBITMQ_PORT", 5672),
		Mode:       envStr("DIFFGRAM_SYSTEM_MODE", "sandbox"),
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
