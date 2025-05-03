package sale

import "time"

type Sale struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Amount   float64   `json:"amount"`
	Status   string    `json:"status"`
	CreateAt time.Time `json:"created_at"`
	UpdateAt time.Time `json:"updated_at"`
	Version  int       `json:"version"`
}

type UpdateFields struct {
	Status *string `json:"status"`
}
