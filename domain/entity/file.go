package entity

import "time"

type File struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	ProjectID        int64     `gorm:"not null;index"`
	Type             string    `gorm:"size:32"`
	ImageID          *int64
	VideoID          *int64
	TextFileID       *int64
	AudioFileID      *int64
	OriginalFilename string    `gorm:"size:512"`
	State            string    `gorm:"size:32;default:'active'"`
	InputID          *int64
	TimeCreated      time.Time `gorm:"autoCreateTime"`
}

func (File) TableName() string { return "file" }
