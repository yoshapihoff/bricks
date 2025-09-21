package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/dto"
)

type UserRepository interface {
	Create(ctx context.Context, user *dto.User) error
	FindByEmail(ctx context.Context, email string) (*dto.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*dto.User, error)
	UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateTables(ctx context.Context) error
}

type DefaultUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *DefaultUserRepository {
	return &DefaultUserRepository{db: db}
}

func (r *DefaultUserRepository) Create(ctx context.Context, user *dto.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *DefaultUserRepository) FindByEmail(ctx context.Context, email string) (*dto.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user dto.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *DefaultUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user dto.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *DefaultUserRepository) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	updatedAt := time.Now()

	query := `
		UPDATE users
		SET email = $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		email,
		updatedAt,
	).Scan(&updatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *DefaultUserRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	updatedAt := time.Now()

	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		passwordHash,
		updatedAt,
	).Scan(&updatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *DefaultUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// CreateTables creates the necessary database tables
func (r *DefaultUserRepository) CreateTables(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}
