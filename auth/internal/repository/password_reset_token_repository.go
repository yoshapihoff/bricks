package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/dto"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *dto.PasswordResetToken) error
	Find(ctx context.Context, token uuid.UUID) (*dto.PasswordResetToken, error)
	Delete(ctx context.Context, token uuid.UUID) error
	ClearFromOld(ctx context.Context, olderThan time.Time) error
	CreateTables(ctx context.Context) error
}

type DefaultPasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(db *sql.DB) *DefaultPasswordResetTokenRepository {
	return &DefaultPasswordResetTokenRepository{db: db}
}

func (p *DefaultPasswordResetTokenRepository) Create(ctx context.Context, passwordResetToken *dto.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (token, user_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING token, user_id, created_at
	`

	err := p.db.QueryRowContext(
		ctx,
		query,
		uuid.New(),
		passwordResetToken.UserID,
		time.Now(),
	).Scan(
		&passwordResetToken.Token,
		&passwordResetToken.UserID,
		&passwordResetToken.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *DefaultPasswordResetTokenRepository) ClearFromOld(ctx context.Context, olderThan time.Time) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE created_at < $1
	`
	_, err := p.db.ExecContext(ctx, query, olderThan)
	return err
}

func (p *DefaultPasswordResetTokenRepository) Find(ctx context.Context, token uuid.UUID) (*dto.PasswordResetToken, error) {
	query := `
		SELECT token, user_id, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var passwordResetToken dto.PasswordResetToken
	err := p.db.QueryRowContext(ctx, query, token).Scan(
		&passwordResetToken.Token,
		&passwordResetToken.UserID,
		&passwordResetToken.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &passwordResetToken, nil
}

func (p *DefaultPasswordResetTokenRepository) Delete(ctx context.Context, token uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE token = $1`

	_, err := p.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	return nil
}

func (p *DefaultPasswordResetTokenRepository) CreateTables(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`

	_, err := p.db.ExecContext(ctx, query)
	return err
}
