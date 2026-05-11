package entity

import (
	"time"

	"gorm.io/datatypes"
)

// ActionRun represents one execution of an Action triggered by an event.
// Status flow: pending -> running -> {complete | failed}
type ActionRun struct {
	ID               int64          `gorm:"column:id;primarykey;autoIncrement"`
	ActionTemplateID *int           `gorm:"column:action_template_id;index"`
	ProjectID        *int           `gorm:"column:project_id;index"`
	RunnerName       string         `gorm:"column:runner_name;size:100;not null"`
	Status           string         `gorm:"column:status;size:32;not null;default:'pending';index"`
	ConfigData       datatypes.JSON `gorm:"column:config_data;type:json"`
	EventPayload     datatypes.JSON `gorm:"column:event_payload;type:json"`
	Output           datatypes.JSON `gorm:"column:output;type:json"`
	ErrorMessage     *string        `gorm:"column:error_message;type:text"`
	TimeCreated      *time.Time     `gorm:"column:time_created"`
	TimeUpdated      *time.Time     `gorm:"column:time_updated"`
}

func (ActionRun) TableName() string { return "action_run" }
