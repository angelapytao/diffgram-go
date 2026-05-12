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

type WorkerConfig struct {
	HTTPPort          int
	EventsPrefetch    int
	ActionsPrefetch   int
	JobsPrefetch      int
	SchedulerPrefetch int
}

type MLRunnerConfig struct {
	HTTPAddr string
}

type GCPConfig struct {
	ProjectID         string
	Region            string
	ServiceAccountB64 string
	VertexBaseURL     string
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
	Worker      WorkerConfig
	MLRunner    MLRunnerConfig
	GCP         GCPConfig
}

func Load() *Config {
	region := envStr("GCP_REGION", "us-central1")
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
		Worker: WorkerConfig{
			HTTPPort:          envInt("WORKER_HTTP_PORT", 8086),
			EventsPrefetch:    envInt("WORKER_EVENTS_PREFETCH", 5),
			ActionsPrefetch:   envInt("WORKER_ACTIONS_PREFETCH", 10),
			JobsPrefetch:      envInt("WORKER_JOBS_PREFETCH", 5),
			SchedulerPrefetch: envInt("WORKER_SCHEDULER_PREFETCH", 2),
		},
		MLRunner: MLRunnerConfig{
			HTTPAddr: envStr("ML_RUNNER_HTTP_ADDR", "http://localhost:8087"),
		},
		GCP: GCPConfig{
			ProjectID:         os.Getenv("GCP_PROJECT_ID"),
			Region:            region,
			ServiceAccountB64: os.Getenv("GCP_SERVICE_ACCOUNT_JSON"),
			VertexBaseURL:     envStr("VERTEX_BASE_URL", "https://"+region+"-aiplatform.googleapis.com"),
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
