package entity

import "time"

type Member struct {
	Base
	Kind      *string `gorm:"column:kind"`
	UserID    *int    `gorm:"column:user_id"`
	AuthAPIID *int    `gorm:"column:auth_api_id"`
}

type AuthAPI struct {
	Base
	MemberID        *int       `gorm:"column:member_id"`
	CreatedTime     *time.Time `gorm:"column:created_time"`
	ClientID        *string    `gorm:"column:client_id"`
	ClientSecret    *string    `gorm:"column:client_secret"`
	ProjectStringID *string    `gorm:"column:project_string_id"`
	ProjectID       *int       `gorm:"column:project_id"`
	IsLive          *bool      `gorm:"column:is_live"`
	IsValid         *bool      `gorm:"column:is_valid"`
	PermissionLevel *string    `gorm:"column:permission_level"`
}

func (AuthAPI) TableName() string { return "auth_api" }
