package entity

import "time"

type AudioFile struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	OriginalFilename string    `gorm:"size:512"`
	BlobPath         string    `gorm:"size:1024"`
	Duration         float64
	SampleRate       int
	Channels         int
	TimeCreated      time.Time `gorm:"autoCreateTime"`
}

func (AudioFile) TableName() string { return "audio_file" }
