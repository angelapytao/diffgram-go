package service

import (
	"gorm.io/gorm"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
)

var tokenSvc domainservice.TokenService

// Init is the composition root. Called once from cmd/api/main.go.
// db may be nil in unit tests (repos pre-initialised via domain service Init).
func Init(db *gorm.DB, ts domainservice.TokenService, _ interface{}) {
	tokenSvc = ts
	if db != nil {
		domainservice.GetUserService().Init(infrarepo.NewUserRepository(db))
		domainservice.GetProjectService().Init(infrarepo.NewProjectRepository(db))
	}
}
