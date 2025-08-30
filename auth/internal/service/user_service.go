package service

import (
	"context"
	"errors"
	"net/mail"

	"github.com/google/uuid"
	"github.com/yoshapihoff/bricks/auth/internal/auth"
	"github.com/yoshapihoff/bricks/auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already in use")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	userRepo domain.UserRepository
	jwtSvc   *auth.JWTService
}

func NewUserService(userRepo domain.UserRepository, jwtSvc *auth.JWTService) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
	}
}

func (s *UserService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Note: Token generation is now handled in the handler layer
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	return s.jwtSvc.GenerateToken(user.ID, user.Email)
}

func (s *UserService) LoginByID(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Generate JWT token
	return s.jwtSvc.GenerateToken(user.ID, user.Email)
}

func (s *UserService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.jwtSvc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *UserService) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.userRepo.UpdateEmail(ctx, userID, email)
}

func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePasswordHash(ctx, userID, string(hashedPassword))
}

// GetUserByEmail retrieves a user by their email address
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}
