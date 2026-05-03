package facade

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appservice "github.com/angelapytao/diffgram-go/application/service"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/dto"
	"github.com/angelapytao/diffgram-go/util"
)

func RegisterProjectRoutes(r gin.IRouter) {
	r.POST("/api/project/new", createProjectHandler)
	r.POST("/api/v1/project/list", listProjectsHandler)
	r.GET("/api/project/:project_string_id/view", viewProjectHandler)
}

func createProjectHandler(c *gin.Context) {
	claims := c.MustGet("claims").(*domainservice.Claims)

	var req dto.CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"log": errLog("input", err.Error())})
		return
	}
	resp, err := appservice.GetProjectApp().Create(c.Request.Context(), &req, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*util.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"log": errLog("project", appErr.Msg)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"log": errLog("server", "internal error")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": okLog(), "project": resp})
}

func listProjectsHandler(c *gin.Context) {
	claims := c.MustGet("claims").(*domainservice.Claims)

	projects, err := appservice.GetProjectApp().ListByUserPrimaryID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"log": errLog("server", "internal error")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": okLog(), "project_list": projects})
}

func viewProjectHandler(c *gin.Context) {
	sid := c.Param("project_string_id")

	proj, err := appservice.GetProjectApp().GetByStringID(c.Request.Context(), sid)
	if err != nil {
		if appErr, ok := err.(*util.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"log": errLog("project", appErr.Msg)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"log": errLog("server", "internal error")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": okLog(), "project": proj})
}
