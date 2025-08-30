package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/domain"
)

type PasswordResetTokenService struct {
	repo        domain.PasswordResetTokenRepository
	userService domain.UserService
}

func (p *PasswordResetTokenService) ReceiveUserIdByToken(ctx context.Context, token uuid.UUID, expiration time.Duration) (uuid.UUID, error) {
	passwordResetToken, err := p.repo.Find(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrPasswordResetTokenNotFound) {
			return uuid.UUID{}, domain.ErrPasswordResetTokenNotFound
		}
		return uuid.UUID{}, err
	}
	if passwordResetToken.CreatedAt.Add(expiration).Before(time.Now()) {
		return uuid.UUID{}, domain.ErrPasswordResetTokenExpired
	}
	return passwordResetToken.UserID, nil
}

func (p *PasswordResetTokenService) ClearFromOld(ctx context.Context, olderThan time.Time) error {
	return p.repo.ClearFromOld(ctx, olderThan)
}

func (p *PasswordResetTokenService) Create(ctx context.Context, userEmail string) (*domain.PasswordResetToken, error) {
	user, err := p.userService.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	passwordResetToken := &domain.PasswordResetToken{
		UserID: user.ID,
	}

	if err := p.repo.Create(ctx, passwordResetToken); err != nil {
		return nil, err
	}

	return passwordResetToken, nil
}

func NewPasswordResetTokenService(repo domain.PasswordResetTokenRepository, userService domain.UserService) domain.PasswordResetTokenService {
	return &PasswordResetTokenService{repo: repo, userService: userService}
}
