package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/yoshapihoff/bricks/auth/internal/auth"
	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth"
	"github.com/yoshapihoff/bricks/auth/internal/config"
	"github.com/yoshapihoff/bricks/auth/internal/db"
	httpHandler "github.com/yoshapihoff/bricks/auth/internal/handler/http"
	"github.com/yoshapihoff/bricks/auth/internal/kafka/producers"
	postgresRepo "github.com/yoshapihoff/bricks/auth/internal/repository/postgres"
	"github.com/yoshapihoff/bricks/auth/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	dbConn, err := db.Init(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Create repositories
	userRepo := postgresRepo.NewUserRepository(dbConn)

	// Initialize JWT service
	jwtSvc := auth.NewJWTService(auth.JWTConfig{
		Secret:     cfg.JWT.Secret,
		Expiration: cfg.JWT.Expiration,
	})

	// Initialize OAuth service
	oauthCfg := oauth.Config{
		Google: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
		},
		GitHub: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
		},
		VK: struct {
			ClientID     string
			ClientSecret string
			APIVersion   string
		}{
			ClientID:     cfg.OAuth.VK.ClientID,
			ClientSecret: cfg.OAuth.VK.ClientSecret,
			APIVersion:   cfg.OAuth.VK.APIVersion,
		},
		RedirectURL: cfg.OAuth.RedirectURL,
	}

	oauthSvc := oauth.NewService(oauthCfg)

	// Initialize services
	userSvc := service.NewUserService(userRepo, jwtSvc)
	passwordResetTokenSvc := service.NewPasswordResetTokenService(postgresRepo.NewPasswordResetTokenRepository(dbConn), userSvc)

	// Initialize forgot password email Kafka producer
	forgotPasswordEmailProducer, err := producers.NewForgotPasswordEmailProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize forgot password email Kafka producer: %v", err)
	}
	defer forgotPasswordEmailProducer.Close()

	// Create HTTP server
	r := mux.NewRouter()

	// Create handler
	handler := httpHandler.NewAuthHandler(
		userSvc,
		oauthSvc,
		jwtSvc,
		passwordResetTokenSvc,
		cfg.PasswordResetTokenExpiration,
		forgotPasswordEmailProducer,
	)
	handler.RegisterRoutes(r)

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server is running on http://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

	log.Println("Server stopped")
}
