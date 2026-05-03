package dto

type CreateProjectReq struct {
	Name            *string `json:"name"`
	ProjectStringID *string `json:"project_string_id"`
	OrgID           *int    `json:"org_id"`
}

type ProjectResp struct {
	ID              int     `json:"id"`
	Name            *string `json:"name,omitempty"`
	ProjectStringID *string `json:"project_string_id,omitempty"`
}

type UpdateProjectReq struct {
	ID   int     `json:"id"`
	Name *string `json:"name"`
}
