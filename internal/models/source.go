package models

type Source struct {
	ID      string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	URL     string  `gorm:"type:text;not null" json:"url"`
	Comment *string `gorm:"type:text" json:"comment,omitempty"`
}
