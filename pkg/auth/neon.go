package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apierrors "github.com/videoforge/backend/pkg/errors"
)

// NeonAuthClient wraps Neon Auth API calls for serverless authentication.
// It supports OAuth providers (Google, GitHub), email/password authentication,
// JWT tokens, and session management.
type NeonAuthClient struct {
	apiKey       string
	projectID    string
	branchID     string
	cookieSecret string
	httpClient  *http.Client
	baseURL      string
}

// NeonAuthConfig contains configuration for Neon Auth client.
// All fields are required and should be set via environment variables.
type NeonAuthConfig struct {
	ProjectID    string `envconfig:"NEON_PROJECT_ID" required:"true"`
	BranchID     string `envconfig:"NEON_BRANCH_ID" required:"true"`
	APIKey       string `envconfig:"NEON_API_KEY" required:"true"`
	CookieSecret string `envconfig:"NEON_AUTH_COOKIE_SECRET" required:"true"`
}

// NewNeonAuthClient creates a new Neon Auth client with the given configuration.
// It validates that all required configuration values are provided.
func NewNeonAuthClient(config NeonAuthConfig) *NeonAuthClient {
	return &NeonAuthClient{
		apiKey:       config.APIKey,
		projectID:    config.ProjectID,
		branchID:     config.BranchID,
		cookieSecret: config.CookieSecret,
		baseURL:      "https://console.neon.tech/api/v2",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProviderType represents the type of authentication provider.
type ProviderType string

const (
	ProviderGoogle ProviderType = "google"
	ProviderGitHub ProviderType = "github"
	ProviderEmail  ProviderType = "email"
)

// ProvidersRequest represents a request to configure auth providers.
type ProvidersRequest struct {
	Providers []ProviderType `json:"providers"`
}

// ProvidersResponse represents the response from configuring auth providers.
type ProvidersResponse struct {
	Providers []ProviderConfig `json:"providers"`
}

// ProviderConfig represents the configuration for a single auth provider.
type ProviderConfig struct {
	Type        ProviderType `json:"type"`
	Enabled    bool        `json:"enabled"`
	ClientID   string     `json:"client_id,omitempty"`
	ClientSecret string  `json:"client_secret,omitempty"`
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Provider  string `json:"provider,omitempty"`
}

// RegisterResponse represents a user registration response.
type RegisterResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email       string `json:"email"`
	Password   string `json:"password"`
	Provider   string `json:"provider,omitempty"`
	OauthToken string `json:"oauth_token,omitempty"`
}

// LoginResponse represents a user login response.
type LoginResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// User represents a Neon Auth user.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Session represents a user session.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionResponse represents a session check response.
type SessionResponse struct {
	Session Session `json:"session"`
	User    User   `json:"user"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id,omitempty"`
}

// JWTClaims represents JWT token claims.
type JWTClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
}

// buildError creates an error with the appropriate status code.
func buildError(message string, status int) error {
	return apierrors.New(message, status)
}

// ConfigureProviders configures the authentication providers for a project.
// This endpoint is used to enable OAuth providers like Google and GitHub,
// as well as email/password authentication.
func (c *NeonAuthClient) ConfigureProviders(providers []ProviderType) (*ProvidersResponse, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/providers",
		c.baseURL, c.projectID, c.branchID)

	reqBody := ProvidersRequest{
		Providers: providers,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, buildError("failed to marshal request", http.StatusInternalServerError)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, buildError("failed to configure providers", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, buildError("failed to configure providers", resp.StatusCode)
	}

	var result ProvidersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

// Register registers a new user with Neon Auth.
// It supports email/password registration and can be extended for OAuth flows.
func (c *NeonAuthClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/register",
		c.baseURL, c.projectID, c.branchID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, buildError("failed to marshal request", http.StatusInternalServerError)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, buildError("registration failed", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, buildError("user already exists", http.StatusConflict)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, buildError("registration failed", resp.StatusCode)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

// Login authenticates a user with Neon Auth.
// It supports email/password login and OAuth token exchange.
func (c *NeonAuthClient) Login(req LoginRequest) (*LoginResponse, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/login",
		c.baseURL, c.projectID, c.branchID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, buildError("failed to marshal request", http.StatusInternalServerError)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, buildError("login failed", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, buildError("invalid credentials", http.StatusUnauthorized)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, buildError("login failed", resp.StatusCode)
	}

	var result LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

// GetSession retrieves the current session for a user.
// The session token is typically passed as a cookie or Bearer token.
func (c *NeonAuthClient) GetSession(sessionToken string) (*SessionResponse, error) {
	if sessionToken == "" {
		return nil, buildError("session token required", http.StatusBadRequest)
	}

	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/session",
		c.baseURL, c.projectID, c.branchID)

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, buildError("failed to get session", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, buildError("invalid or expired session", http.StatusUnauthorized)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, buildError("failed to get session", resp.StatusCode)
	}

	var result SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

// Logout invalidates a user session.
// This invalidates the refresh token and/or session ID.
func (c *NeonAuthClient) Logout(req LogoutRequest) error {
	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/logout",
		c.baseURL, c.projectID, c.branchID)

	body, err := json.Marshal(req)
	if err != nil {
		return buildError("failed to marshal request", http.StatusInternalServerError)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return buildError("logout failed", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return buildError("logout failed", resp.StatusCode)
	}

	return nil
}

// RefreshToken refreshes an access token using a refresh token.
func (c *NeonAuthClient) RefreshToken(refreshToken string) (*LoginResponse, error) {
	if refreshToken == "" {
		return nil, buildError("refresh token required", http.StatusBadRequest)
	}

	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/refresh",
		c.baseURL, c.projectID, c.branchID)

	type refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	body, err := json.Marshal(refreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, buildError("failed to marshal request", http.StatusInternalServerError)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, buildError("token refresh failed", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, buildError("invalid or expired refresh token", http.StatusUnauthorized)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, buildError("token refresh failed", resp.StatusCode)
	}

	var result LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

// ValidateToken validates an access token and returns the claims.
func (c *NeonAuthClient) ValidateToken(accessToken string) (*JWTClaims, error) {
	if accessToken == "" {
		return nil, buildError("access token required", http.StatusBadRequest)
	}

	endpoint := fmt.Sprintf("%s/projects/%s/branches/%s/auth/validate",
		c.baseURL, c.projectID, c.branchID)

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, buildError("failed to create request", http.StatusInternalServerError)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, buildError("token validation failed", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, buildError("invalid or expired token", http.StatusUnauthorized)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, buildError("token validation failed", resp.StatusCode)
	}

	var claims JWTClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, buildError("failed to decode response", http.StatusInternalServerError)
	}

	return &claims, nil
}

// IsConfigured returns true if the Neon Auth client has valid configuration.
// This can be used to check if Neon Auth is available before attempting to use it.
func (c *NeonAuthClient) IsConfigured() bool {
	return c.apiKey != "" && c.projectID != "" && c.branchID != "" && c.cookieSecret != ""
}