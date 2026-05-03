package dto

// RegisterReq is the body for POST /api/v1/user/new
type RegisterReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginReq is the body for POST /api/user/login
// Mode must be "password" (only mode supported in P3).
type LoginReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Mode     string `json:"mode"`
}

type LoginResp struct {
	Token  string `json:"token"`
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

type UserResp struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type UpdateUserReq struct {
	ID        int     `json:"id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}
