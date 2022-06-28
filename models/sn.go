package entity

import "time"

// SerialNumberSet TODO
type SerialNumberSet struct {
	SetID     int64     `json:"setid"`
	MinSN     string    `json:"minsn"`
	MaxSN     string    `json:"maxsn"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Article TODO
type Article struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title" validate:"required"`
	Content   string    `json:"content" validate:"required"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
