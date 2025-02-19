package response

import "time"

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  *Error      `json:"error,omitempty"`
	Meta   Meta        `json:"meta"`
}

type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Meta struct {
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type PaginatedResponse struct {
	Response
	Pagination Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalItems   int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
}

func NewResponse(data interface{}) Response {
	return Response{
		Status: "success",
		Data:   data,
		Meta: Meta{
			Timestamp: time.Now(),
			Version:   "1.0.0",
		},
	}
}

func NewErrorResponse(code, message string, details interface{}) Response {
	return Response{
		Status: "error",
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: Meta{
			Timestamp: time.Now(),
			Version:   "1.0.0",
		},
	}
} 