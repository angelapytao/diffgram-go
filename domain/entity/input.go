package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Input struct {
	ID               int64          `gorm:"primaryKey;autoIncrement"`
	ProjectID        int64          `gorm:"not null;index"`
	MemberID         *int64
	MediaType        string         `gorm:"size:32;index"`
	Type             string         `gorm:"size:32"`
	Status           string         `gorm:"size:32;default:'pending';index"`
	StatusText       string         `gorm:"size:512"`
	OriginalFilename string         `gorm:"size:512"`
	Extension        string         `gorm:"size:32"`
	DirectoryID      *int64
	FileID           *int64
	ParentFileID     *int64
	BatchID          *int64
	BlobPath         string         `gorm:"size:1024"`
	PercentComplete  float64
	RetryCount       int
	UpdateLog        datatypes.JSON `gorm:"type:json"`
	TimeCreated      time.Time      `gorm:"autoCreateTime"`
	TimeUpdated      time.Time      `gorm:"autoUpdateTime"`
}

func (Input) TableName() string { return "input" }
