package entity

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	Base
	MemberID                         *int           `gorm:"column:member_id"`
	QosLastCachedValue               *float64       `gorm:"column:qos_last_cached_value"`
	FirstName                        *string        `gorm:"column:first_name;size:100"`
	LastName                         *string        `gorm:"column:last_name;size:100"`
	HowHearAboutUs                   *string        `gorm:"column:how_hear_about_us"`
	CompanyName                      *string        `gorm:"column:company_name"`
	SecurityDisableGlobal            *bool          `gorm:"column:security_disable_global"`
	SecurityEmailVerified            *bool          `gorm:"column:security_email_verified"`
	APIActions                       *bool          `gorm:"column:api_actions"`
	APIEnabledBuilderBrain           *bool          `gorm:"column:api_enabled_builder_brain"`
	APIEnabledBuilder                *bool          `gorm:"column:api_enabled_builder"`
	APIEnabledTrainer                *bool          `gorm:"column:api_enabled_trainer"`
	LastBuilderOrTrainerMode         *string        `gorm:"column:last_builder_or_trainer_mode"`
	PhoneNumber                      *string        `gorm:"column:phone_number"`
	City                             *string        `gorm:"column:city"`
	Username                         *string        `gorm:"column:username"`
	Email                            string         `gorm:"column:email;size:100;not null"`
	PasswordHash                     *string        `gorm:"column:password_hash"`
	PasswordAttemptCount             *int           `gorm:"column:password_attempt_count"`
	AutoCommit                       *bool          `gorm:"column:auto_commit"`
	OTPSecret                        *string        `gorm:"column:otp_secret"`
	OTPEnabled                       *bool          `gorm:"column:otp_enabled"`
	OTPBackup                        *string        `gorm:"column:otp_backup"`
	OTPCurrentSession                *string        `gorm:"column:otp_current_session"`
	OTPCurrentSessionExpiry          *int           `gorm:"column:otp_current_session_expiry"`
	ProfileImageID                   *int           `gorm:"column:profile_image_id"`
	ProfileImageURL                  *string        `gorm:"column:profile_image_url"`
	ProfileImageBlob                 *string        `gorm:"column:profile_image_blob"`
	ProfileImageExpiry               *int           `gorm:"column:profile_image_expiry"`
	ProfileImageThumbURL             *string        `gorm:"column:profile_image_thumb_url"`
	ProfileImageThumbBlob            *string        `gorm:"column:profile_image_thumb_blob"`
	CreatedTime                      *time.Time     `gorm:"column:created_time"`
	CreatedRemoteAddress             *string        `gorm:"column:created_remote_address"`
	LastTime                         *time.Time     `gorm:"column:last_time"`
	CurrentAIVersionNumber           *int           `gorm:"column:current_ai_version_number"`
	CurrentProjectStringID           *string        `gorm:"column:current_project_string_id;size:100"`
	ProjectCurrentID                 *int           `gorm:"column:project_current_id"`
	PermissionsProjects              *string        `gorm:"column:permissions_projects"`
	PermissionsGeneral               *string        `gorm:"column:permissions_general"`
	IsSuperAdmin                     *bool          `gorm:"column:is_super_admin"`
	LastTaskID                       *int           `gorm:"column:last_task_id"`
	AvailableForAnnotationAssignment *bool          `gorm:"column:available_for_annotation_assignment"`
	IsAnnotator                      *bool          `gorm:"column:is_annotator"`
	SignupCodeID                     *int           `gorm:"column:signup_code_id"`
	VerifyEmailCodeID                *int           `gorm:"column:verify_email_code_id"`
	FollowIngCount                   *int           `gorm:"column:follow_ing_count"`
	FollowErsCount                   *int           `gorm:"column:follow_ers_count"`
	SignupRole                       *string        `gorm:"column:signup_role"`
	SignupDemo                       *string        `gorm:"column:signup_demo"`
	SignupHowManyDataLabelers        *string        `gorm:"column:signup_how_many_data_labelers"`
	OccupationList                   datatypes.JSON `gorm:"column:occupation_list"`
	LinkedinProfileURL               *string        `gorm:"column:linkedin_profile_url"`
	DefaultPlanID                    *int           `gorm:"column:default_plan_id"`
	OIDCID                           *string        `gorm:"column:oidc_id"`
}

func (User) TableName() string { return "userbase" }

type Account struct {
	Base
	Nickname              *string    `gorm:"column:nickname"`
	ModeTrainerOrBuilder  *string    `gorm:"column:mode_trainer_or_builder"`
	AccountType           *string    `gorm:"column:account_type"`
	CreditLimit           *int       `gorm:"column:credit_limit"`
	PaymentMethodOnFile   *bool      `gorm:"column:payment_method_on_file"`
	SecurityDisable       *bool      `gorm:"column:security_disable"`
	TransactionPreviousID *int       `gorm:"column:transaction_previous_id"`
	AddressPrimaryID      *int       `gorm:"column:address_primary_id"`
	PrimaryUserID         *int       `gorm:"column:primary_user_id"`
	StripeID              *string    `gorm:"column:stripe_id"`
	MemberCreatedID       *int       `gorm:"column:member_created_id"`
	MemberUpdatedID       *int       `gorm:"column:member_updated_id"`
	TimeCreated           *time.Time `gorm:"column:time_created"`
	TimeUpdated           *time.Time `gorm:"column:time_updated"`
}
