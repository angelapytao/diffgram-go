package facade

import (
	"github.com/gin-gonic/gin"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/middleware"
)

// RegisterAPIRoutes wires all application routes onto the Gin engine.
func RegisterAPIRoutes(r *gin.Engine, tokenSvc domainservice.TokenService) {
	RegisterUserRoutes(r)

	authed := r.Group("/", middleware.Auth(tokenSvc))
	RegisterProjectRoutes(authed)
}
