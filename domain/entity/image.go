package entity

import "time"

type Image struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement"`
	OriginalFilename   string    `gorm:"size:512"`
	URLSignedBlobPath  string    `gorm:"size:1024"`
	ThumbLargeBlobPath string    `gorm:"size:1024"`
	ThumbSmallBlobPath string    `gorm:"size:1024"`
	Width              int
	Height             int
	TimeCreated        time.Time `gorm:"autoCreateTime"`
}

func (Image) TableName() string { return "image" }
