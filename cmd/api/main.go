package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/interfaces/http/health"
)

func main() {
	cfg := config.Load()

	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	health.RegisterRoutes(r)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("api server starting on %s (mode=%s)", addr, cfg.Mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("api server failed: %v", err)
	}
}
