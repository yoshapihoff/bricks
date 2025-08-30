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
)

// Ensure githubProvider implements interfaces.Provider
var _ interfaces.Provider = (*githubProvider)(nil)

type githubProvider struct {
	config *oauth2.Config
}

// NewGitHubProvider creates a new GitHub OAuth provider
func NewGitHubProvider(clientID, clientSecret, redirectURL string, scopes []string) interfaces.Provider {
	if scopes == nil {
		scopes = []string{"user:email"}
	}

	return &githubProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
	}
}

func (g *githubProvider) Name() string {
	return "github"
}

func (g *githubProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *githubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

type githubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (g *githubProvider) getUserEmails(client *http.Client) ([]struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create email request: %v", err)
	}
	token := &oauth2.Token{}
	token.SetAuthHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get email: status %d", resp.StatusCode)
	}

	var emails []struct {
		Email      string `json:"email"`
		Primary    bool   `json:"primary"`
		Verified   bool   `json:"verified"`
		Visibility string `json:"visibility"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to decode email response: %v", err)
	}

	return emails, nil
}

func (g *githubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*oauthtypes.UserInfo, error) {
	client := g.config.Client(ctx, token)

	// Get user's primary email
	emails, err := g.getUserEmails(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get user emails: %v", err)
	}

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
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

	var gui githubUserInfo
	if err := json.Unmarshal(body, &gui); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	// Use primary email if available, otherwise use the first one
	email := ""
	for _, e := range emails {
		if e.Primary && e.Verified {
			email = e.Email
			break
		} else if email == "" {
			email = e.Email
		}
	}

	return &oauthtypes.UserInfo{
		ID:       fmt.Sprint(gui.ID),
		Email:    email,
		Name:     gui.Name,
		Picture:  gui.AvatarURL,
		Username: gui.Login,
	}, nil
}
