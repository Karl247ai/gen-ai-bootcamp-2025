package models

// ErrorResponse represents an error response
type ErrorResponse struct {
    Error string `json:"error" example:"error message"`
}

// WordResponse represents a successful word operation response
type WordResponse struct {
    Data Word `json:"data"`
}