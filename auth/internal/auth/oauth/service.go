package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/interfaces"
	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/oauthtypes"
	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/providers"
	"golang.org/x/oauth2"
)

// Service handles OAuth authentication flows
type Service struct {
	providers map[string]interfaces.Provider
}

// Config holds OAuth configuration
// This is a placeholder - you should use your existing config structure
// and update it to include the new OAuth providers

type Config struct {
	Google struct {
		ClientID     string
		ClientSecret string
	}
	GitHub struct {
		ClientID     string
		ClientSecret string
	}
	VK struct {
		ClientID     string
		ClientSecret string
		APIVersion   string
	}
	RedirectURL string
}

// NewService creates a new OAuth service with the given providers
func NewService(cfg Config) *Service {
	s := &Service{
		providers: make(map[string]interfaces.Provider),
	}

	// Initialize Google provider if configured
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		s.providers["google"] = providers.NewGoogleProvider(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			fmt.Sprintf(cfg.RedirectURL, "google"),
			nil, // use default scopes
		)
	}

	// Initialize GitHub provider if configured
	if cfg.GitHub.ClientID != "" && cfg.GitHub.ClientSecret != "" {
		s.providers["github"] = providers.NewGitHubProvider(
			cfg.GitHub.ClientID,
			cfg.GitHub.ClientSecret,
			fmt.Sprintf(cfg.RedirectURL, "github"),
			nil, // use default scopes
		)
	}

	// Initialize VK provider if configured
	if cfg.VK.ClientID != "" && cfg.VK.ClientSecret != "" {
		s.providers["vk"] = providers.NewVKProvider(
			cfg.VK.ClientID,
			cfg.VK.ClientSecret,
			fmt.Sprintf(cfg.RedirectURL, "vk"),
			cfg.VK.APIVersion,
			nil, // use default scopes (email)
		)
	}

	return s
}

// GetProvider returns the provider with the given name
func (s *Service) GetProvider(name string) (interfaces.Provider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("oauth provider %s not found", name)
	}
	return provider, nil
}

// GetAuthURL returns the authorization URL for the given provider and state
func (s *Service) GetAuthURL(provider, state string) (string, error) {
	p, err := s.GetProvider(provider)
	if err != nil {
		return "", err
	}

	// Generate a random state if not provided
	if state == "" {
		state, err = generateRandomString(32)
		if err != nil {
			return "", err
		}
	}

	return p.GetAuthURL(state), nil
}

// ExchangeCode exchanges an authorization code for a token
func (s *Service) ExchangeCode(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	p, err := s.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	token, err := p.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}

	return token, nil
}

// GetUserInfo retrieves the user's information from the OAuth provider
func (s *Service) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*oauthtypes.UserInfo, error) {
	p, err := s.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	userInfo, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	return userInfo, nil
}

// generateRandomString generates a random string of the given length
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetSupportedProviders returns a list of supported OAuth providers
func (s *Service) GetSupportedProviders() []string {
	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}
	return providers
}
