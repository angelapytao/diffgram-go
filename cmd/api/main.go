package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appservice "github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/config"
	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
	infratoken "github.com/angelapytao/diffgram-go/infrastructure/token"
	"github.com/angelapytao/diffgram-go/interfaces/http/facade"
	"github.com/angelapytao/diffgram-go/interfaces/http/health"
	"github.com/angelapytao/diffgram-go/interfaces/http/middleware"
)

func main() {
	cfg := config.Load()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}

	gormDB, err := diffdb.NewConnection(cfg.DBDsn)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	tokenSvc := infratoken.NewJWTService(cfg.JWT)
	appservice.Init(gormDB, tokenSvc, nil)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))

	health.RegisterRoutes(r)
	facade.RegisterAPIRoutes(r, tokenSvc)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.WithField("addr", addr).WithField("mode", cfg.Mode).Info("api server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("api server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down api server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("server forced to shutdown")
	}
	log.Info("server exited")
}
