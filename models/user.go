package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	PhotoURL  *string   `json:"photoUrl"`
	Role      string    `json:"role"` // "user" or "developer"
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserResponse struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	PhotoURL *string `json:"photoUrl"`
	Role     string  `json:"role"`
}
