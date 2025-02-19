package models

import "time"

type WordGroup struct {
	ID        int       `json:"id" db:"id"`
	WordID    int       `json:"word_id" db:"word_id"`
	GroupID   int       `json:"group_id" db:"group_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
} 