package entity

import "time"

type Label struct {
	Base
	Name       *string `gorm:"column:name;size:200"`
	Colour     *string `gorm:"column:colour"`
	SoftDelete *bool   `gorm:"column:soft_delete"`
}

type LabelSchema struct {
	Base
	Name            string     `gorm:"column:name;not null"`
	ProjectID       *int       `gorm:"column:project_id"`
	Archived        *bool      `gorm:"column:archived"`
	MemberCreatedID *int       `gorm:"column:member_created_id"`
	MemberUpdatedID *int       `gorm:"column:member_updated_id"`
	TimeCreated     *time.Time `gorm:"column:time_created"`
	TimeUpdated     *time.Time `gorm:"column:time_updated"`
	IsDefault       *bool      `gorm:"column:is_default"`
}

func (LabelSchema) TableName() string { return "label_schema" }

type LabelSchemaLink struct {
	Base
	SchemaID                 *int       `gorm:"column:schema_id"`
	LabelFileID              *int       `gorm:"column:label_file_id"`
	InstanceTemplateID       *int       `gorm:"column:instance_template_id"`
	AttributeTemplateGroupID *int       `gorm:"column:attribute_template_group_id"`
	MemberCreatedID          *int       `gorm:"column:member_created_id"`
	MemberUpdatedID          *int       `gorm:"column:member_updated_id"`
	TimeCreated              *time.Time `gorm:"column:time_created"`
	TimeUpdated              *time.Time `gorm:"column:time_updated"`
}

func (LabelSchemaLink) TableName() string { return "label_schema_link" }
