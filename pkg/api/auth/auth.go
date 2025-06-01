// Package auth provides authentication mechanisms for API testing.
//
// This package implements various authentication methods commonly used in APIs,
// including OAuth, API keys, JWT, and basic authentication. It provides a unified
// interface for handling different authentication schemes.
package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// AuthType represents the type of authentication
type AuthType int

const (
	// AuthNone represents no authentication
	AuthNone AuthType = iota
	// AuthBasic represents HTTP Basic authentication
	AuthBasic
	// AuthBearer represents Bearer token authentication
	AuthBearer
	// AuthAPIKey represents API key authentication
	AuthAPIKey
	// AuthOAuth represents OAuth authentication
	AuthOAuth
)

// AuthProvider is an interface for different authentication mechanisms
type AuthProvider interface {
	// AddAuth adds authentication to the given HTTP request
	AddAuth(req *http.Request) error
	// GetAuthType returns the type of authentication
	GetAuthType() AuthType
	// GetDescription returns a human-readable description of the authentication
	GetDescription() string
}

// BasicAuth implements HTTP Basic authentication
type BasicAuth struct {
	Username string
	Password string
}

// NewBasicAuth creates a new BasicAuth provider
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

// AddAuth adds Basic authentication to the given HTTP request
func (a *BasicAuth) AddAuth(req *http.Request) error {
	auth := a.Username + ":" + a.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

// GetAuthType returns the type of authentication
func (a *BasicAuth) GetAuthType() AuthType {
	return AuthBasic
}

// GetDescription returns a human-readable description of the authentication
func (a *BasicAuth) GetDescription() string {
	return fmt.Sprintf("Basic Auth (username: %s)", a.Username)
}

// BearerAuth implements Bearer token authentication
type BearerAuth struct {
	Token string
}

// NewBearerAuth creates a new BearerAuth provider
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		Token: token,
	}
}

// AddAuth adds Bearer token authentication to the given HTTP request
func (a *BearerAuth) AddAuth(req *http.Request) error {
	if a.Token == "" {
		return api.NewAPIError("Bearer token is empty", 0)
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

// GetAuthType returns the type of authentication
func (a *BearerAuth) GetAuthType() AuthType {
	return AuthBearer
}

// GetDescription returns a human-readable description of the authentication
func (a *BearerAuth) GetDescription() string {
	// Mask the token for security
	tokenPreview := ""
	if len(a.Token) > 8 {
		tokenPreview = a.Token[:4] + "..." + a.Token[len(a.Token)-4:]
	} else if len(a.Token) > 0 {
		tokenPreview = "***"
	}
	return fmt.Sprintf("Bearer Token (%s)", tokenPreview)
}

// APIKeyAuth implements API key authentication
type APIKeyAuth struct {
	Key      string
	Name     string
	Location string // header, query, or cookie
}

// NewAPIKeyAuth creates a new APIKeyAuth provider
func NewAPIKeyAuth(key, name, location string) *APIKeyAuth {
	// Default to header if location is not specified
	if location == "" {
		location = "header"
	}

	// Normalize location
	location = strings.ToLower(location)

	return &APIKeyAuth{
		Key:      key,
		Name:     name,
		Location: location,
	}
}

// AddAuth adds API key authentication to the given HTTP request
func (a *APIKeyAuth) AddAuth(req *http.Request) error {
	if a.Key == "" {
		return api.NewAPIError("API key is empty", 0)
	}

	switch a.Location {
	case "header":
		req.Header.Set(a.Name, a.Key)
	case "query":
		q := req.URL.Query()
		q.Add(a.Name, a.Key)
		req.URL.RawQuery = q.Encode()
	case "cookie":
		req.Header.Add("Cookie", a.Name+"="+a.Key)
	default:
		return api.NewAPIError("Invalid API key location: "+a.Location, 0)
	}

	return nil
}

// GetAuthType returns the type of authentication
func (a *APIKeyAuth) GetAuthType() AuthType {
	return AuthAPIKey
}

// GetDescription returns a human-readable description of the authentication
func (a *APIKeyAuth) GetDescription() string {
	// Mask the key for security
	keyPreview := ""
	if len(a.Key) > 8 {
		keyPreview = a.Key[:4] + "..." + a.Key[len(a.Key)-4:]
	} else if len(a.Key) > 0 {
		keyPreview = "***"
	}
	return fmt.Sprintf("API Key (%s in %s: %s)", a.Name, a.Location, keyPreview)
}

// OAuthGrantType represents the type of OAuth grant
type OAuthGrantType string

const (
	// GrantTypeClientCredentials represents the client credentials grant type
	GrantTypeClientCredentials OAuthGrantType = "client_credentials"
	// GrantTypePassword represents the password grant type
	GrantTypePassword OAuthGrantType = "password"
	// GrantTypeAuthCode represents the authorization code grant type
	GrantTypeAuthCode OAuthGrantType = "authorization_code"
	// GrantTypeImplicit represents the implicit grant type
	GrantTypeImplicit OAuthGrantType = "implicit"
)

// OAuthToken represents an OAuth access token response
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresAt    time.Time
}

// OAuthAuth implements OAuth 2.0 authentication
type OAuthAuth struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	GrantType    OAuthGrantType
	Username     string // Used for password grant
	Password     string // Used for password grant
	Scope        string
	Token        *OAuthToken
	Client       *http.Client
}

// NewOAuthClientCredentials creates a new OAuthAuth provider with client credentials grant
func NewOAuthClientCredentials(clientID, clientSecret, tokenURL, scope string) *OAuthAuth {
	return &OAuthAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		GrantType:    GrantTypeClientCredentials,
		Scope:        scope,
		Client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// NewOAuthPassword creates a new OAuthAuth provider with password grant
func NewOAuthPassword(clientID, clientSecret, username, password, tokenURL, scope string) *OAuthAuth {
	return &OAuthAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		TokenURL:     tokenURL,
		GrantType:    GrantTypePassword,
		Scope:        scope,
		Client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// fetchToken obtains a new OAuth token
func (a *OAuthAuth) fetchToken() error {
	data := url.Values{}
	data.Set("grant_type", string(a.GrantType))

	if a.Scope != "" {
		data.Set("scope", a.Scope)
	}

	switch a.GrantType {
	case GrantTypeClientCredentials:
		// Client credentials flow
	case GrantTypePassword:
		// Resource owner password credentials flow
		if a.Username == "" || a.Password == "" {
			return api.NewAPIError("Username and password are required for password grant", 0)
		}
		data.Set("username", a.Username)
		data.Set("password", a.Password)
	default:
		return api.NewAPIError("Unsupported grant type: "+string(a.GrantType), 0)
	}

	req, err := http.NewRequest("POST", a.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return api.NewAPIError("Failed to create token request: "+err.Error(), 0)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add client authentication
	if a.ClientID != "" && a.ClientSecret != "" {
		req.SetBasicAuth(a.ClientID, a.ClientSecret)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return api.NewAPIError("Failed to fetch token: "+err.Error(), 0)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.NewAPIError("Failed to read token response: "+err.Error(), 0)
	}

	if resp.StatusCode != http.StatusOK {
		return api.NewAPIError(fmt.Sprintf("Token request failed with status %d: %s", resp.StatusCode, string(body)), resp.StatusCode)
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return api.NewAPIError("Failed to parse token response: "+err.Error(), 0)
	}

	// Calculate token expiration time
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	a.Token = &token
	return nil
}

// isTokenExpired checks if the current token is expired
func (a *OAuthAuth) isTokenExpired() bool {
	if a.Token == nil {
		return true
	}

	// Consider token expired if it expires in less than 30 seconds
	return a.Token.ExpiresAt.Before(time.Now().Add(30 * time.Second))
}

// AddAuth adds OAuth authentication to the given HTTP request
func (a *OAuthAuth) AddAuth(req *http.Request) error {
	// Fetch token if we don't have one or if it's expired
	if a.isTokenExpired() {
		if err := a.fetchToken(); err != nil {
			return err
		}
	}

	if a.Token == nil || a.Token.AccessToken == "" {
		return api.NewAPIError("No valid OAuth token available", 0)
	}

	// Add the token to the request
	tokenType := a.Token.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	req.Header.Set("Authorization", tokenType+" "+a.Token.AccessToken)
	return nil
}

// GetAuthType returns the type of authentication
func (a *OAuthAuth) GetAuthType() AuthType {
	return AuthOAuth
}

// GetDescription returns a human-readable description of the authentication
func (a *OAuthAuth) GetDescription() string {
	grantTypeStr := string(a.GrantType)

	// Mask client secret for security
	clientSecretPreview := ""
	if len(a.ClientSecret) > 8 {
		clientSecretPreview = a.ClientSecret[:4] + "..." + a.ClientSecret[len(a.ClientSecret)-4:]
	} else if len(a.ClientSecret) > 0 {
		clientSecretPreview = "***"
	}

	tokenPreview := ""
	if a.Token != nil && len(a.Token.AccessToken) > 8 {
		tokenPreview = a.Token.AccessToken[:4] + "..." + a.Token.AccessToken[len(a.Token.AccessToken)-4:]
	} else if a.Token != nil && len(a.Token.AccessToken) > 0 {
		tokenPreview = "***"
	}

	return fmt.Sprintf("OAuth (%s, client: %s, secret: %s, token: %s)", grantTypeStr, a.ClientID, clientSecretPreview, tokenPreview)
}
