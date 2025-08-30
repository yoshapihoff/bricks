package oauthtypes

// UserInfo contains the standardized user information returned by OAuth providers
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture,omitempty"`
	Username string `json:"username,omitempty"` // For providers like GitHub that use login/username
}
