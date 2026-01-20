package models

import "time"

type ProcessedFile struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ZipFilename string    `gorm:"type:text;not null;uniqueIndex" json:"zip_filename"`
	ProcessedAt time.Time `gorm:"not null" json:"processed_at"`
}
