package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const version = "0.1.0"

func RegisterRoutes(r gin.IRouter) {
	r.GET("/api/status", statusHandler)
	r.GET("/healthz", healthzHandler)
	r.GET("/readyz", readyzHandler)
}

func statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": version,
	})
}

func healthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func readyzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
