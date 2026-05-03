package entity

import "time"

type Project struct {
	Base
	Name                                *string    `gorm:"column:name;size:250"`
	IsPublic                            *bool      `gorm:"column:is_public"`
	Goal                                *string    `gorm:"column:goal"`
	DeletionPending                     *bool      `gorm:"column:deletion_pending"`
	DeletionID                          *int       `gorm:"column:deletion_id"`
	HighestAIVersionNumber              *int       `gorm:"column:highest_ai_version_number"`
	StarCount                           *int       `gorm:"column:star_count"`
	LatestIssueNumber                   *int       `gorm:"column:latest_issue_number"`
	CreditBalance                       *float64   `gorm:"column:credit_balance"`
	ProjectStringID                     *string    `gorm:"column:project_string_id;size:100"`
	UserPrimaryID                       *int       `gorm:"column:user_primary_id"`
	DirectoryDefaultID                  *int       `gorm:"column:directory_default_id"`
	DefaultReportDashboardID            *int       `gorm:"column:default_report_dashboard_id"`
	APIBillingEnabled                   *bool      `gorm:"column:api_billing_enabled"`
	AnnotationsFeedbackLoopTriggerCheck *bool      `gorm:"column:annotations_feedback_loop_trigger_check"`
	SettingsInputVideoFPS               *int       `gorm:"column:settings_input_video_fps"`
	Readme                              *string    `gorm:"column:readme"`
	MemberCreatedID                     *int       `gorm:"column:member_created_id"`
	MemberUpdatedID                     *int       `gorm:"column:member_updated_id"`
	TimeCreated                         *time.Time `gorm:"column:time_created"`
	TimeUpdated                         *time.Time `gorm:"column:time_updated"`
	LabelDict                           *string    `gorm:"column:label_dict"`
	CacheDict                           *string    `gorm:"column:cache_dict"`
	DefaultExternalMapID                *int       `gorm:"column:default_external_map_id"`
	OrgID                               *int       `gorm:"column:org_id"`
	AccountID                           *int       `gorm:"column:account_id"`
	PlanID                              *int       `gorm:"column:plan_id"`
}

func (Project) TableName() string { return "project" }

type WorkingDir struct {
	Base
	CreatedTime          *time.Time `gorm:"column:created_time"`
	Archived             *bool      `gorm:"column:archived"`
	HasChanges           *bool      `gorm:"column:has_changes"`
	Nickname             *string    `gorm:"column:nickname"`
	CountChanges         *int       `gorm:"column:count_changes"`
	UserID               *int       `gorm:"column:user_id"`
	FileLimitTime        *int       `gorm:"column:file_limit_time"`
	BranchID             *int       `gorm:"column:branch_id"`
	ProjectID            *int       `gorm:"column:project_id"`
	LabelFileColourMap   *string    `gorm:"column:label_file_colour_map"`
	JobsToSync           *string    `gorm:"column:jobs_to_sync"`
	Type                 *string    `gorm:"column:type"`
	DefaultExternalMapID *int       `gorm:"column:default_external_map_id"`
	AccessType           string     `gorm:"column:access_type;not null;default:'project'"`
}

func (WorkingDir) TableName() string { return "working_dir" }

type ProjectSettings struct {
	Base
	ProjectID           *int     `gorm:"column:project_id"`
	FanOn               *bool    `gorm:"column:fan_on"`
	FanTriggerInterval  *int     `gorm:"column:fan_trigger_interval"`
	FanInferenceSize    *int     `gorm:"column:fan_inference_size"`
	FanType             *string  `gorm:"column:fan_type"`
	FanInferenceMinimum *float64 `gorm:"column:fan_inference_minimum"`
	FanMethod           *string  `gorm:"column:fan_method"`
	FanSubMethod        *string  `gorm:"column:fan_sub_method"`
}

func (ProjectSettings) TableName() string { return "project_settings" }

type ProjectDirectoryList struct {
	WorkingDirID int     `gorm:"column:working_dir_id;primarykey"`
	ProjectID    int     `gorm:"column:project_id;primarykey"`
	Nickname     *string `gorm:"column:nickname"`
}

func (ProjectDirectoryList) TableName() string { return "project_directory_list" }
