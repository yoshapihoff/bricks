package service

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/auth"
	"github.com/yoshapihoff/bricks/auth/internal/dto"
	"github.com/yoshapihoff/bricks/auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrWeakPassword    = errors.New("password is too weak")
)

type UserService interface {
	Register(ctx context.Context, email, password, name string) (*dto.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*dto.User, error)
	LoginByID(ctx context.Context, userID uuid.UUID) (string, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.User, error)
	UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*dto.User, error)
}

type DefaultUserService struct {
	userRepo repository.UserRepository
	jwtSvc   auth.JWTService
}

func NewUserService(userRepo repository.UserRepository, jwtSvc auth.JWTService) *DefaultUserService {
	return &DefaultUserService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
	}
}

func (s *DefaultUserService) Register(ctx context.Context, email, password, name string) (*dto.User, error) {
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &dto.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *DefaultUserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidEmail
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidPassword
	}

	return s.jwtSvc.GenerateToken(user.ID, user.Email)
}

func (s *DefaultUserService) LoginByID(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return s.jwtSvc.GenerateToken(user.ID, user.Email)
}

func (s *DefaultUserService) ValidateToken(ctx context.Context, tokenString string) (*dto.User, error) {
	claims, err := s.jwtSvc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *DefaultUserService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *DefaultUserService) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return s.userRepo.UpdateEmail(ctx, userID, email)
}

func (s *DefaultUserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePasswordHash(ctx, userID, string(hashedPassword))
}

func (s *DefaultUserService) GetUserByEmail(ctx context.Context, email string) (*dto.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
