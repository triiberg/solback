package models

import "time"

type Log struct {
	ID       string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID  *string   `gorm:"type:uuid" json:"event_id,omitempty"`
	Datetime time.Time `gorm:"column:datetime;not null" json:"datetime"`
	Action   string    `gorm:"type:text;not null" json:"action"`
	Outcome  string    `gorm:"type:text;not null" json:"outcome"`
	Message  *string   `gorm:"type:text" json:"message,omitempty"`
}
