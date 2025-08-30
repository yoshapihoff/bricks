package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/domain"
)

type passwordResetTokenRepository struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(db *sql.DB) domain.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (p *passwordResetTokenRepository) Create(ctx context.Context, passwordResetToken *domain.PasswordResetToken) error {
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

func (p *passwordResetTokenRepository) ClearFromOld(ctx context.Context, olderThan time.Time) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE created_at < $1
	`

	_, err := p.db.ExecContext(ctx, query, olderThan)
	return err
}

func (p *passwordResetTokenRepository) Find(ctx context.Context, token uuid.UUID) (*domain.PasswordResetToken, error) {
	query := `
		SELECT token, user_id, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var passwordResetToken domain.PasswordResetToken
	err := p.db.QueryRowContext(ctx, query, token).Scan(
		&passwordResetToken.Token,
		&passwordResetToken.UserID,
		&passwordResetToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPasswordResetTokenNotFound
		}
		return nil, err
	}

	return &passwordResetToken, nil
}

func (p *passwordResetTokenRepository) Delete(ctx context.Context, token uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE token = $1`

	result, err := p.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrPasswordResetTokenNotFound
	}

	return nil
}

func (p *passwordResetTokenRepository) CreateTables(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			created_at TIMESTAMP NOT NULL,
		);
	`

	_, err := p.db.ExecContext(ctx, query)
	return err
}
