package dto

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	Token     uuid.UUID `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
