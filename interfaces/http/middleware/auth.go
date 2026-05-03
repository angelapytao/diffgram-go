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
		tokenStr := tokenFromRequest(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": util.ErrUnauthorized.Code,
				"msg":  util.ErrUnauthorized.Msg,
			})
			return
		}
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

func tokenFromRequest(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if cookie, err := c.Cookie("diffgram_jwt"); err == nil && cookie != "" {
		return cookie
	}
	return ""
}
