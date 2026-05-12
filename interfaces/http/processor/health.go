package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/infrastructure/media"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type HealthHandler struct {
	db   *gorm.DB
	mq   *mq.Client
	pool *media.WorkerPool
}

func NewHealthHandler(db *gorm.DB, mqClient *mq.Client, pool *media.WorkerPool) *HealthHandler {
	return &HealthHandler{db: db, mq: mqClient, pool: pool}
}

func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", h.Liveness)
	r.GET("/readyz", h.Readiness)
	r.GET("/api/walrus/status", h.Status)
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": "db connection failed"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": "db ping failed"})
		return
	}

	if h.mq != nil {
		if err := h.mq.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": "mq ping failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"queue_pending": h.pool.Pending(),
		"queue_len":     h.pool.QueueLen(),
		"queue_cap":     h.pool.QueueCap(),
	})
}
