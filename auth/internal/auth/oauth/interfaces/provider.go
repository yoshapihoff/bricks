package interfaces

import (
	"context"

	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/oauthtypes"
	"golang.org/x/oauth2"
)

// Provider defines the interface that all OAuth providers must implement
type Provider interface {
	// GetAuthURL returns the URL to redirect the user to for authentication
	GetAuthURL(state string) string

	// Exchange exchanges an authorization code for a token
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)

	// GetUserInfo retrieves the user's information using the access token
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*oauthtypes.UserInfo, error)

	// Name returns the name of the provider (e.g., "google", "github")
	Name() string
}
