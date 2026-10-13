package models

import "time"

// 使用指针类型表示可选字段
type Todo struct {
	// 必填
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 非必填
	Description *string    `json:"description,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	DueDate     *string    `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TodoFilter struct {
	Status   *string `form:"status" binding:"omitempty,oneof=pending completed cancelled"`
	Priority *string `form:"priority" binding:"omitempty,oneof=low medium high"`
	Search   *string `form:"search"`
	Page     int     `form:"page,default=1" binding:"gte=1"`
	PageSize int     `form:"page_size,default=20" binding:"gte=1,lte=100"`
}
