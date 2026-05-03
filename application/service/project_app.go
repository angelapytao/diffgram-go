package service

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/dto"
	"github.com/angelapytao/diffgram-go/util"
)

type projectApp struct{}

var pApp projectApp

func GetProjectApp() *projectApp { return &pApp }

func (a *projectApp) Create(ctx context.Context, req *dto.CreateProjectReq, userID int) (*dto.ProjectResp, error) {
	proj := &entity.Project{
		Name:            req.Name,
		ProjectStringID: req.ProjectStringID,
		UserPrimaryID:   &userID,
	}
	if err := domainservice.GetProjectService().Create(ctx, proj); err != nil {
		return nil, err
	}
	return &dto.ProjectResp{ID: proj.ID, Name: proj.Name, ProjectStringID: proj.ProjectStringID}, nil
}

func (a *projectApp) GetByID(ctx context.Context, id int) (*dto.ProjectResp, error) {
	proj, err := domainservice.GetProjectService().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, util.ErrNotFound
	}
	return &dto.ProjectResp{ID: proj.ID, Name: proj.Name, ProjectStringID: proj.ProjectStringID}, nil
}

func (a *projectApp) GetByStringID(ctx context.Context, sid string) (*dto.ProjectResp, error) {
	proj, err := domainservice.GetProjectService().GetByStringID(ctx, sid)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, util.ErrNotFound
	}
	return &dto.ProjectResp{ID: proj.ID, Name: proj.Name, ProjectStringID: proj.ProjectStringID}, nil
}

func (a *projectApp) ListByOrgID(ctx context.Context, orgID int) ([]*dto.ProjectResp, error) {
	projects, err := domainservice.GetProjectService().ListByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.ProjectResp, len(projects))
	for i, p := range projects {
		result[i] = &dto.ProjectResp{ID: p.ID, Name: p.Name, ProjectStringID: p.ProjectStringID}
	}
	return result, nil
}

func (a *projectApp) ListByUserPrimaryID(ctx context.Context, userID int) ([]*dto.ProjectResp, error) {
	projects, err := domainservice.GetProjectService().ListByUserPrimaryID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.ProjectResp, len(projects))
	for i, p := range projects {
		result[i] = &dto.ProjectResp{ID: p.ID, Name: p.Name, ProjectStringID: p.ProjectStringID}
	}
	return result, nil
}
