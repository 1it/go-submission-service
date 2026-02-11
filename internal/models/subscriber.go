package models

import "time"

type Subscriber struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"` // e.g., "pending", "invited", "failed"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
