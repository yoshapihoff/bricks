package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
)

type PasswordResetToken struct {
	Token     uuid.UUID `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	Find(ctx context.Context, token uuid.UUID) (*PasswordResetToken, error)
	Delete(ctx context.Context, token uuid.UUID) error
	ClearFromOld(ctx context.Context, olderThan time.Time) error
	CreateTables(ctx context.Context) error
}

type PasswordResetTokenService interface {
	Create(ctx context.Context, userEmail string) (*PasswordResetToken, error)
	ReceiveUserIdByToken(ctx context.Context, token uuid.UUID, expiration time.Duration) (uuid.UUID, error)
	ClearFromOld(ctx context.Context, olderThan time.Time) error
}
