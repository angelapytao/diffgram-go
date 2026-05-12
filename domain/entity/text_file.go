package entity

import "time"

type TextFile struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	OriginalFilename string    `gorm:"size:512"`
	BlobPath         string    `gorm:"size:1024"`
	TokenCount       int
	TimeCreated      time.Time `gorm:"autoCreateTime"`
}

func (TextFile) TableName() string { return "text_file" }
