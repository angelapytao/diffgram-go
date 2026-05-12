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

type ProcessorConfig struct {
	HTTPPort           int
	RPCPort            int
	WorkerPoolSize     int
	MaxUploadSize      int64
	TempDir            string
	FFmpegPath         string
	FFprobePath        string
	GDALInfoPath       string
	GDALTranslatePath  string
	VideoSplitDuration int
	VideoFPSDefault    int
	ThumbLargeSize     int
	ThumbSmallSize     int
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
	Processor   ProcessorConfig
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
		Processor: ProcessorConfig{
			HTTPPort:           envInt("PROCESSOR_HTTP_PORT", 8082),
			RPCPort:            envInt("PROCESSOR_RPC_PORT", 8083),
			WorkerPoolSize:     envInt("PROCESSOR_WORKER_POOL_SIZE", 0),
			MaxUploadSize:      envInt64("PROCESSOR_MAX_UPLOAD_SIZE", 250*1024*1024),
			TempDir:            envStr("PROCESSOR_TEMP_DIR", os.TempDir()),
			FFmpegPath:         envStr("FFMPEG_PATH", "ffmpeg"),
			FFprobePath:        envStr("FFPROBE_PATH", "ffprobe"),
			GDALInfoPath:       envStr("GDALINFO_PATH", "gdalinfo"),
			GDALTranslatePath:  envStr("GDAL_TRANSLATE_PATH", "gdal_translate"),
			VideoSplitDuration: envInt("VIDEO_SPLIT_DURATION", 60),
			VideoFPSDefault:    envInt("VIDEO_FPS_DEFAULT", 5),
			ThumbLargeSize:     envInt("THUMB_LARGE_SIZE", 800),
			ThumbSmallSize:     envInt("THUMB_SMALL_SIZE", 200),
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

func envInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
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
