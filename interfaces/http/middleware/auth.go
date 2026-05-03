package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/util"
)

func Auth(tokenSvc domainservice.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": util.ErrUnauthorized.Code,
				"msg":  util.ErrUnauthorized.Msg,
			})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.Verify(c.Request.Context(), tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": util.ErrUnauthorized.Code,
				"msg":  util.ErrUnauthorized.Msg,
			})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
