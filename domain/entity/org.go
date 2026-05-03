package entity

import "time"

type Org struct {
	Base
	SecurityDisable      *bool      `gorm:"column:security_disable"`
	VerifiedByDiffgram   *bool      `gorm:"column:verified_by_diffgram"`
	Name                 *string    `gorm:"column:name"`
	APIAddressValid      *bool      `gorm:"column:api_address_valid"`
	APITrainerOrg        *bool      `gorm:"column:api_trainer_org"`
	AddressPrimaryID     *int       `gorm:"column:address_primary_id"`
	PrimaryUserID        *int       `gorm:"column:primary_user_id"`
	MemberCreatedID      *int       `gorm:"column:member_created_id"`
	MemberUpdatedID      *int       `gorm:"column:member_updated_id"`
	TimeCreated          *time.Time `gorm:"column:time_created"`
	TimeUpdated          *time.Time `gorm:"column:time_updated"`
	RemoteAddressCreated *string    `gorm:"column:remote_address_created"`
}
