package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	APIVersion   string
}

type OAuthConfig struct {
	Google      OAuthProviderConfig
	GitHub      OAuthProviderConfig
	VK          OAuthProviderConfig
	RedirectURL string
}

type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type KafkaConfig struct {
	KafkaUrl          string
	SchemaRegistryUrl string
}

type Config struct {
	DB                              DBConfig
	JWT                             JWTConfig
	OAuth                           OAuthConfig
	Server                          ServerConfig
	Kafka                           KafkaConfig
	AppPort                         string
	PasswordResetTokenExpiration    time.Duration
	ForgotPasswordEmailSendingTopic string
}

func Load() (*Config, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}
	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	passwordResetTokenExpiration, err := time.ParseDuration(getEnv("PASSWORD_RESET_TOKEN_EXPIRATION"))
	if err != nil {
		return nil, err
	}

	// Initialize OAuth configuration
	oauthConfig := OAuthConfig{
		Google: OAuthProviderConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID"),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET"),
		},
		GitHub: OAuthProviderConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID"),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET"),
		},
		VK: OAuthProviderConfig{
			ClientID:     getEnv("VK_CLIENT_ID"),
			ClientSecret: getEnv("VK_CLIENT_SECRET"),
			APIVersion:   getEnv("VK_API_VERSION"),
		},
		RedirectURL: getEnv("OAUTH_REDIRECT_URL"),
	}

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET"),
			Expiration: jwtExpiration,
		},
		Server: ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		Kafka: KafkaConfig{
			KafkaUrl:          getEnv("KAFKA_URL"),
			SchemaRegistryUrl: getEnv("SCHEMA_REGISTRY_URL"),
		},
		OAuth:                           oauthConfig,
		AppPort:                         getEnv("PORT"),
		PasswordResetTokenExpiration:    passwordResetTokenExpiration,
		ForgotPasswordEmailSendingTopic: getEnv("FORGOT_PASSWORD_EMAIL_SENDING_TOPIC"),
	}, nil
}

// GetDSN returns the database connection string
func (c *DBConfig) GetDSN() string {
	return "postgres://" +
		c.User + ":" +
		c.Password + "@" +
		c.Host + ":" +
		c.Port + "/" +
		c.Name + "?sslmode=" +
		c.SSLMode
}

// getEnv gets an environment variable or returns a default value
func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
