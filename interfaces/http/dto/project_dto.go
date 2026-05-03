package dto

// CreateProjectReq is the body for POST /api/project/new
type CreateProjectReq struct {
	Name            *string `json:"project_name"      binding:"required"`
	ProjectStringID *string `json:"project_string_id" binding:"required"`
	OrgID           *int    `json:"org_id"`
}

type ProjectResp struct {
	ID              int     `json:"id"`
	Name            *string `json:"name,omitempty"`
	ProjectStringID *string `json:"project_string_id,omitempty"`
}

type UpdateProjectReq struct {
	ID   int     `json:"id"`
	Name *string `json:"project_name"`
}
