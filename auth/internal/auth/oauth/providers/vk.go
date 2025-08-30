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

type vkProvider struct {
	config     *oauth2.Config
	apiVersion string
}

// NewVKProvider creates a new VK OAuth provider
func NewVKProvider(clientID, clientSecret, redirectURL, apiVersion string, scopes []string) interfaces.Provider {
	if scopes == nil {
		scopes = []string{"email"} // Default scope for basic user info and email
	}

	if apiVersion == "" {
		apiVersion = "5.199" // Default to a recent stable version
	}

	return &vkProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://oauth.vk.com/authorize",
				TokenURL: "https://oauth.vk.com/access_token",
			},
		},
		apiVersion: apiVersion,
	}
}

func (v *vkProvider) GetAuthURL(state string) string {
	// VK requires 'display' and 'v' parameters
	return v.config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("display", "page"),
		oauth2.SetAuthURLParam("v", v.apiVersion),
	)
}

func (v *vkProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return v.config.Exchange(ctx, code)
}

func (v *vkProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*oauthtypes.UserInfo, error) {
	// Get email from token (VK includes it in the token response)
	email, _ := token.Extra("email").(string)
	userID, _ := token.Extra("user_id").(float64)

	// If we don't have email in token, try to get it from token response
	if email == "" {
		email, _ = token.Extra("email").(string)
	}

	// Get user info from VK API
	userInfoURL := fmt.Sprintf(
		"https://api.vk.com/method/users.get?user_ids=%.0f&fields=photo_200,first_name,last_name&access_token=%s&v=%s",
		userID,
		token.AccessToken,
		v.apiVersion,
	)

	resp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse VK API response
	var result struct {
		Response []struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Photo     string `json:"photo_200"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	if len(result.Response) == 0 {
		return nil, fmt.Errorf("no user data in response")
	}

	user := result.Response[0]
	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	return &oauthtypes.UserInfo{
		ID:       fmt.Sprint(user.ID),
		Email:    email,
		Name:     name,
		Picture:  user.Photo,
		Username: fmt.Sprint(user.ID),
	}, nil
}

func (v *vkProvider) Name() string {
	return "vk"
}
