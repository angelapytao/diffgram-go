package facade

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appservice "github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/dto"
	"github.com/angelapytao/diffgram-go/util"
)

func RegisterUserRoutes(r gin.IRouter) {
	r.POST("/api/v1/user/new", registerHandler)
	r.POST("/api/user/login", loginHandler)
}

func registerHandler(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"log": errLog("input", err.Error())})
		return
	}
	resp, err := appservice.GetUserApp().Register(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*util.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"log": errLog("email", appErr.Msg)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"log": errLog("server", "internal error")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": okLog(), "user": resp})
}

func loginHandler(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"log": errLog("input", err.Error())})
		return
	}
	if req.Mode != "password" {
		c.JSON(http.StatusBadRequest, gin.H{"log": errLog("mode", "only 'password' mode supported")})
		return
	}
	resp, err := appservice.GetUserApp().Login(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*util.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"log": errLog("password", appErr.Msg)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"log": errLog("server", "internal error")})
		return
	}
	c.SetCookie("diffgram_jwt", resp.Token, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"log":  okLog(),
		"user": dto.UserResp{ID: resp.UserID, Email: resp.Email},
	})
}
