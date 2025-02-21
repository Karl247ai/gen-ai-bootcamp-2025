package models

import (
    "time"
    "encoding/json"
)

// Word represents a vocabulary item
// @Description Word vocabulary item
type Word struct {
    ID        int64  `json:"id" example:"1"`
    Japanese  string `json:"japanese" example:"猫" binding:"required"`
    Romaji    string `json:"romaji" example:"neko" binding:"required"`
    English   string `json:"english" example:"cat" binding:"required"`
    Parts     json.RawMessage `json:"parts,omitempty"`
    CreatedAt string `json:"created_at" example:"2024-02-21T15:04:05Z07:00"`
    UpdatedAt string `json:"updated_at" example:"2024-02-21T15:04:05Z07:00"`
}