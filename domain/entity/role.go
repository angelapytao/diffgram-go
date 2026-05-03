package entity

import "gorm.io/datatypes"

type Role struct {
	Base
	ProjectID       *int           `gorm:"column:project_id"`
	Name            *string        `gorm:"column:name"`
	PermissionsList datatypes.JSON `gorm:"column:permissions_list"`
}

type RoleMemberObject struct {
	Base
	MemberID        *int    `gorm:"column:member_id"`
	ObjectID        *int    `gorm:"column:object_id"`
	ObjectType      *string `gorm:"column:object_type"`
	DefaultRoleName *string `gorm:"column:default_role_name"`
	RoleID          *int    `gorm:"column:role_id"`
}

func (RoleMemberObject) TableName() string { return "role_member_object" }
