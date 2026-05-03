package service

import (
	"context"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/dto"
	"github.com/angelapytao/diffgram-go/util"
)

type userApp struct{}

var uApp userApp

func GetUserApp() *userApp { return &uApp }

func (a *userApp) Register(ctx context.Context, req *dto.RegisterReq) (*dto.UserResp, error) {
	user, err := domainservice.GetUserService().Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &dto.UserResp{ID: user.ID, Email: user.Email}, nil
}

func (a *userApp) Login(ctx context.Context, req *dto.LoginReq) (*dto.LoginResp, error) {
	user, err := domainservice.GetUserService().Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := tokenSvc.Issue(ctx, domainservice.Claims{UserID: user.ID, Email: user.Email})
	if err != nil {
		return nil, util.ErrInternal
	}
	return &dto.LoginResp{Token: token, UserID: user.ID, Email: user.Email}, nil
}

func (a *userApp) GetByID(ctx context.Context, id int) (*dto.UserResp, error) {
	user, err := domainservice.GetUserService().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, util.ErrNotFound
	}
	return &dto.UserResp{ID: user.ID, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}, nil
}
