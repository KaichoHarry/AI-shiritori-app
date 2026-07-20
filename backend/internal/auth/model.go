package auth

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PasswordResetStatus string

const (
	PasswordResetStatusPending  PasswordResetStatus = "pending"
	PasswordResetStatusApproved PasswordResetStatus = "approved"
	PasswordResetStatusExpired  PasswordResetStatus = "expired"
)

type PasswordResetRequest struct {
	ID              string
	UserID          string
	Status          PasswordResetStatus
	AdminNotifiedAt time.Time
	ApprovedAt      *time.Time
	ExpiresAt       time.Time
	CreatedAt       time.Time
}
