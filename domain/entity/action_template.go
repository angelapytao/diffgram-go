package entity

import (
	"time"

	"gorm.io/datatypes"
)

// ActionTemplate is a configured action that fires when an event matches its event_type.
type ActionTemplate struct {
	Base
	ProjectID   *int           `gorm:"column:project_id;index"`
	EventType   *string        `gorm:"column:event_type;size:100;index"`
	RunnerName  *string        `gorm:"column:runner_name;size:100"`
	ConfigData  datatypes.JSON `gorm:"column:config_data;type:json"`
	PublicName  *string        `gorm:"column:public_name"`
	Kind        *string        `gorm:"column:kind"`
	IsAvailable *bool          `gorm:"column:is_available"`
	IsGlobal    *bool          `gorm:"column:is_global"`
	TimeCreated *time.Time     `gorm:"column:time_created"`
	TimeUpdated *time.Time     `gorm:"column:time_updated"`
}

func (ActionTemplate) TableName() string { return "action_template" }
