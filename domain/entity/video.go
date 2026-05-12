package entity

import "time"

type Video struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	OriginalFilename string    `gorm:"size:512"`
	FileBlobPath     string    `gorm:"size:1024"`
	FileSignedURL    string    `gorm:"size:2048"`
	FPS              float64
	FrameCount       int
	Duration         float64
	Width            int
	Height           int
	Status           string    `gorm:"size:32"`
	TimeCreated      time.Time `gorm:"autoCreateTime"`
}

func (Video) TableName() string { return "video" }
