package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Pinger interface {
	Ping() error
}

func RegisterWorkerHealth(r gin.IRouter, db *gorm.DB, mqPing Pinger) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if sqlDB, err := db.DB(); err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db not ready"})
			return
		}
		if err := mqPing.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "mq not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
