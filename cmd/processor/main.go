package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/infrastructure/db"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
	"github.com/angelapytao/diffgram-go/infrastructure/media/audio"
	"github.com/angelapytao/diffgram-go/infrastructure/media/geotiff"
	"github.com/angelapytao/diffgram-go/infrastructure/media/image"
	"github.com/angelapytao/diffgram-go/infrastructure/media/sensorfusion"
	"github.com/angelapytao/diffgram-go/infrastructure/media/text"
	"github.com/angelapytao/diffgram-go/infrastructure/media/video"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	"github.com/angelapytao/diffgram-go/infrastructure/storage"
	httpproc "github.com/angelapytao/diffgram-go/interfaces/http/processor"
)

func main() {
	cfg := config.Load()

	// Database
	gormDB, err := db.NewConnection(cfg.DBDsn)
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}

	// Storage
	storageProvider, err := storage.NewTOSProvider(cfg.TOS)
	if err != nil {
		logrus.Fatalf("failed to create storage provider: %v", err)
	}

	// Message queue
	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		logrus.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer mqClient.Close()
	publisher, err := mq.NewPublisher(mqClient)
	if err != nil {
		logrus.Fatalf("failed to create publisher: %v", err)
	}

	bucket := cfg.TOS.Bucket

	// Media processors
	processors := map[string]media.MediaProcessor{
		"image":         image.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
		"video":         video.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
		"text":          text.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
		"audio":         audio.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
		"sensor_fusion": sensorfusion.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
		"geo_tiff":      geotiff.NewProcessor(storageProvider, gormDB, cfg.Processor, bucket),
	}

	// Pipeline and worker pool
	pipeline := media.NewPipeline(gormDB, publisher, processors)
	pool := media.NewWorkerPool(cfg.Processor.WorkerPoolSize, pipeline)

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start worker pool
	pool.Start(ctx)

	// HTTP server
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Health routes on engine root
	healthHandler := httpproc.NewHealthHandler(gormDB, mqClient, pool)
	healthHandler.RegisterRoutes(engine)

	// API routes
	projectGroup := engine.Group("/api/walrus/v1/project/:project_id")
	uploadHandler := httpproc.NewUploadHandler(gormDB, pool, storageProvider, cfg.Processor, bucket)
	uploadHandler.RegisterRoutes(projectGroup)
	inputHandler := httpproc.NewInputHandler(gormDB)
	inputHandler.RegisterRoutes(projectGroup)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Processor.HTTPPort),
		Handler: engine,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logrus.Infof("processor HTTP server listening on :%d", cfg.Processor.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		logrus.Info("shutting down HTTP server")
		return srv.Close()
	})

	if err := g.Wait(); err != nil {
		logrus.Fatalf("processor exited with error: %v", err)
	}
}
