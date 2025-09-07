package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yoshapihoff/bricks/auth/internal/auth"
	"github.com/yoshapihoff/bricks/auth/internal/domain"
	"github.com/yoshapihoff/bricks/auth/internal/kafka/producers"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type OAuthStartResponse struct {
	URL string `json:"url"`
}

type OAuthCallbackResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

type AuthHandler struct {
	userService                  domain.UserService
	jwtSvc                       *auth.JWTService
	passwordResetTokenSvc        domain.PasswordResetTokenService
	passwordResetTokenExpiration time.Duration
	forgotPasswordEmailProducer  *producers.ForgotPasswordEmailProducer
}

func NewAuthHandler(
	userService domain.UserService,
	jwtSvc *auth.JWTService,
	passwordResetTokenSvc domain.PasswordResetTokenService,
	passwordResetTokenExpiration time.Duration,
	forgotPasswordEmailProducer *producers.ForgotPasswordEmailProducer,
) *AuthHandler {
	return &AuthHandler{
		userService:                  userService,
		jwtSvc:                       jwtSvc,
		passwordResetTokenSvc:        passwordResetTokenSvc,
		passwordResetTokenExpiration: passwordResetTokenExpiration,
		forgotPasswordEmailProducer:  forgotPasswordEmailProducer,
	}
}

func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	authRouter := router.PathPrefix("/auth").Subrouter()

	// Public routes
	authRouter.HandleFunc("/register", h.handleRegister).Methods("POST")
	authRouter.HandleFunc("/login", h.handleLogin).Methods("POST")
	authRouter.HandleFunc("/forgot-password", h.handleForgotPassword).Methods("POST")
	authRouter.HandleFunc("/receive-password-reset-token/{token}", h.handleReceivePasswordResetToken).Methods("GET")

	// Protected routes
	protected := authRouter.PathPrefix("/me").Subrouter()
	protected.Use(h.authMiddleware)
	protected.HandleFunc("", h.handleGetProfile).Methods("GET")
	protected.HandleFunc("/password", h.handleChangePassword).Methods("PUT")
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Register the user
	user, err := h.userService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		handleError(w, err)
		return
	}

	// Generate JWT token
	token, err := h.jwtSvc.GenerateToken(user.ID, user.Email)
	if err != nil {
		handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, &LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Generate password reset token
	token, err := h.passwordResetTokenSvc.Create(r.Context(), req.Email)
	if err != nil {
		handleError(w, err)
		return
	}

	// Send reset password email
	if _, err := h.forgotPasswordEmailProducer.ProduceForgotPasswordEmail(req.Email, token.Token.String()); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) handleReceivePasswordResetToken(w http.ResponseWriter, r *http.Request) {
	forgotPasswordToken := mux.Vars(r)["token"]
	tokenUUID, err := uuid.Parse(forgotPasswordToken)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	userId, err := h.passwordResetTokenSvc.ReceiveUserIdByToken(r.Context(), tokenUUID, h.passwordResetTokenExpiration)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	token, err := h.userService.LoginByID(r.Context(), userId)
	if err != nil {
		handleError(w, err)
		return
	}

	user, err := h.userService.ValidateToken(r.Context(), token)
	if err != nil {
		handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, &LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleError(w, err)
		return
	}

	user, err := h.userService.ValidateToken(r.Context(), token)
	if err != nil {
		handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, &LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]
		if tokenString == "" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		user, err := h.userService.ValidateToken(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), "userID", user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrEmailExists):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	case errors.Is(err, domain.ErrWeakPassword):
		status = http.StatusBadRequest
	}

	http.Error(w, err.Error(), status)
}
