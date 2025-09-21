package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/dto"
	"github.com/yoshapihoff/bricks/auth/internal/repository"
)

var (
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
)

type PasswordResetTokenService interface {
	Create(ctx context.Context, userEmail string) (*dto.PasswordResetToken, error)
	ReceiveUserIdByToken(ctx context.Context, token uuid.UUID, expiration time.Duration) (uuid.UUID, error)
	ClearFromOld(ctx context.Context, olderThan time.Time) error
}

type DefaultPasswordResetTokenService struct {
	repo        repository.PasswordResetTokenRepository
	userService UserService
}

func NewPasswordResetTokenService(repo repository.PasswordResetTokenRepository, userService UserService) *DefaultPasswordResetTokenService {
	return &DefaultPasswordResetTokenService{repo: repo, userService: userService}
}

func (p *DefaultPasswordResetTokenService) ReceiveUserIdByToken(ctx context.Context, token uuid.UUID, expiration time.Duration) (uuid.UUID, error) {
	passwordResetToken, err := p.repo.Find(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, ErrPasswordResetTokenNotFound
		}
		return uuid.UUID{}, err
	}
	if passwordResetToken.CreatedAt.Add(expiration).Before(time.Now()) {
		return uuid.UUID{}, ErrPasswordResetTokenExpired
	}
	return passwordResetToken.UserID, nil
}

func (p *DefaultPasswordResetTokenService) ClearFromOld(ctx context.Context, olderThan time.Time) error {
	return p.repo.ClearFromOld(ctx, olderThan)
}

func (p *DefaultPasswordResetTokenService) Create(ctx context.Context, userEmail string) (*dto.PasswordResetToken, error) {
	user, err := p.userService.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	passwordResetToken := &dto.PasswordResetToken{
		UserID: user.ID,
	}

	if err := p.repo.Create(ctx, passwordResetToken); err != nil {
		return nil, err
	}

	return passwordResetToken, nil
}
