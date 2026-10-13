package models

import "time"

type APIResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type PaginateResponse struct {
	Items      any `json:"items"`
	Total      int `json:"total"`
}
