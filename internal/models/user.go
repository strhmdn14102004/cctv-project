// internal/models/user.go
package models

import "time"

type User struct {
	ID               int        `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	Password         string     `json:"-"`
	Name             string     `json:"name"`
	PhotoURL         *string    `json:"photoUrl"`
	Role             string     `json:"role"`
	AccountStatus    string     `json:"accountStatus"`
	LastLogin        *time.Time `json:"-"`
	SessionToken     *string    `json:"-"`
	DeviceID         *string    `json:"-"`
	ResetRequested   bool       `json:"-"`
	ResetToken       *string    `json:"-"`
	ResetTokenExpiry *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type UserResponse struct {
	ID            int     `json:"id"`
	Username      string  `json:"username"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	PhotoURL      *string `json:"photoUrl"`
	Role          string  `json:"role"`
	AccountStatus string  `json:"accountStatus"`
}

type ResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPassword struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
