package models

type TodoCreateRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=100"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low medium high"`
	DueDate     *string `json:"due_date,omitempty" binding:"omitempty"`
}

type TodoUpdateRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=100"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low medium high"`
	DueDate     *string `json:"due_date,omitempty"`
}
