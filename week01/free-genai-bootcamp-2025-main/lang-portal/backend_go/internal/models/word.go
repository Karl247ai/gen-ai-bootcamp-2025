package models

import "time"

type Word struct {
	ID        int       `json:"id" db:"id"`
	Japanese  string    `json:"japanese" db:"japanese"`
	Romaji    string    `json:"romaji" db:"romaji"`
	English   string    `json:"english" db:"english"`
	Parts     string    `json:"parts,omitempty" db:"parts"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
} 