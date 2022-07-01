package models

import "time"

// SerialNumberSet TODO
type SerialNumberSet struct {
	SetID     int64     `json:"setid"`
	MinSN     string    `json:"minsn"`
	MaxSN     string    `json:"maxsn"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
