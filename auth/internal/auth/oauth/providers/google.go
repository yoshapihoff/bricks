package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/interfaces"
	"github.com/yoshapihoff/bricks/auth/internal/auth/oauth/oauthtypes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Ensure googleProvider implements interfaces.Provider
var _ interfaces.Provider = (*googleProvider)(nil)

type googleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) interfaces.Provider {
	if scopes == nil {
		scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}
	}

	return &googleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *googleProvider) Name() string {
	return "google"
}

func (g *googleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g *googleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (g *googleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*oauthtypes.UserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var gui googleUserInfo
	if err := json.Unmarshal(body, &gui); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	return &oauthtypes.UserInfo{
		ID:       gui.ID,
		Email:    gui.Email,
		Name:     gui.Name,
		Picture:  gui.Picture,
		Username: "", // Google doesn't provide a username field
	}, nil
}
