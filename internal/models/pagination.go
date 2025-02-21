package models

type PaginationMeta struct {
    CurrentPage  int   `json:"current_page"`
    TotalPages   int   `json:"total_pages"`
    TotalItems   int64 `json:"total_items"`
    ItemsPerPage int   `json:"items_per_page"`
}

type PaginatedResponse struct {
    Data       interface{}    `json:"data"`
    Pagination PaginationMeta `json:"pagination"`
}