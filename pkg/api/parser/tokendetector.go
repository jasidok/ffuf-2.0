// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// TokenType represents the type of token detected
type TokenType string

const (
	// TokenTypeAPIKey represents an API key
	TokenTypeAPIKey TokenType = "api_key"
	// TokenTypeJWT represents a JWT token
	TokenTypeJWT TokenType = "jwt"
	// TokenTypeOAuth represents an OAuth token
	TokenTypeOAuth TokenType = "oauth"
	// TokenTypeBearer represents a Bearer token
	TokenTypeBearer TokenType = "bearer"
	// TokenTypeBasic represents a Basic auth token
	TokenTypeBasic TokenType = "basic"
	// TokenTypeUnknown represents an unknown token type
	TokenTypeUnknown TokenType = "unknown"
)

// TokenLocation represents where the token was found
type TokenLocation string

const (
	// LocationHeader represents a token found in a header
	LocationHeader TokenLocation = "header"
	// LocationBody represents a token found in the response body
	LocationBody TokenLocation = "body"
	// LocationURL represents a token found in a URL
	LocationURL TokenLocation = "url"
)

// Token represents a detected API key or token
type Token struct {
	// Value is the actual token value
	Value string `json:"value"`
	// Type is the detected token type
	Type TokenType `json:"type"`
	// Location is where the token was found
	Location TokenLocation `json:"location"`
	// Path is the JSONPath where the token was found (for body tokens)
	Path string `json:"path,omitempty"`
	// Name is the name of the header or parameter (if applicable)
	Name string `json:"name,omitempty"`
	// Confidence is a score from 0-100 indicating confidence in the detection
	Confidence int `json:"confidence"`
}

// TokenDetector provides methods for detecting API keys and tokens
type TokenDetector struct {
	// Common patterns for different token types
	patterns map[TokenType]*regexp.Regexp
}

// NewTokenDetector creates a new TokenDetector
func NewTokenDetector() *TokenDetector {
	detector := &TokenDetector{
		patterns: make(map[TokenType]*regexp.Regexp),
	}

	// Initialize patterns for different token types
	detector.patterns[TokenTypeJWT] = regexp.MustCompile(`^eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}$`)
	detector.patterns[TokenTypeAPIKey] = regexp.MustCompile(`^[A-Za-z0-9_-]{20,}$`)
	detector.patterns[TokenTypeOAuth] = regexp.MustCompile(`^[a-zA-Z0-9]{20,}$`)
	detector.patterns[TokenTypeBearer] = regexp.MustCompile(`^Bearer\s+[a-zA-Z0-9_.-]+$`)
	detector.patterns[TokenTypeBasic] = regexp.MustCompile(`^Basic\s+[a-zA-Z0-9+/=]+$`)

	return detector
}

// DetectTokensInResponse detects API keys and tokens in an HTTP response
func (d *TokenDetector) DetectTokensInResponse(resp *ffuf.Response) ([]Token, error) {
	tokens := make([]Token, 0)

	// Check headers for tokens
	headerTokens, err := d.detectTokensInHeaders(resp.Headers)
	if err != nil {
		return nil, err
	}
	tokens = append(tokens, headerTokens...)

	// Check body for tokens if it's JSON
	if strings.Contains(resp.ContentType, "application/json") {
		bodyTokens, err := d.detectTokensInJSON(resp.Data)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, bodyTokens...)
	}

	return tokens, nil
}

// detectTokensInHeaders checks for tokens in HTTP headers
func (d *TokenDetector) detectTokensInHeaders(headers map[string][]string) ([]Token, error) {
	tokens := make([]Token, 0)

	// Common header names that might contain tokens
	tokenHeaders := []string{
		"Authorization",
		"X-API-Key",
		"API-Key",
		"X-Auth-Token",
		"Auth-Token",
		"Token",
		"Key",
		"Access-Token",
		"X-Access-Token",
	}

	for _, name := range tokenHeaders {
		if values, ok := headers[name]; ok {
			for _, value := range values {
				tokenType, confidence := d.detectTokenType(value)
				if tokenType != TokenTypeUnknown && confidence > 50 {
					tokens = append(tokens, Token{
						Value:      value,
						Type:       tokenType,
						Location:   LocationHeader,
						Name:       name,
						Confidence: confidence,
					})
				}
			}
		}
	}

	return tokens, nil
}

// detectTokensInJSON checks for tokens in JSON response bodies
func (d *TokenDetector) detectTokensInJSON(data []byte) ([]Token, error) {
	tokens := make([]Token, 0)

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
	}

	// Use JSONPath parser to navigate the JSON structure
	parser := NewJSONPathParserFromObject(jsonData)

	// Extract tokens from the JSON data
	d.extractTokensFromJSON(jsonData, "", parser, &tokens)

	return tokens, nil
}

// extractTokensFromJSON recursively extracts tokens from JSON data
func (d *TokenDetector) extractTokensFromJSON(data interface{}, path string, parser *JSONPathParser, tokens *[]Token) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair
		for key, value := range v {
			newPath := path
			if newPath == "" {
				newPath = "$." + key
			} else {
				newPath = path + "." + key
			}

			// Check if key suggests this might be a token
			if d.isLikelyTokenKey(key) {
				if strValue, ok := value.(string); ok {
					tokenType, confidence := d.detectTokenType(strValue)
					if tokenType != TokenTypeUnknown && confidence > 50 {
						*tokens = append(*tokens, Token{
							Value:      strValue,
							Type:       tokenType,
							Location:   LocationBody,
							Path:       newPath,
							Name:       key,
							Confidence: confidence,
						})
						// Don't recursively process this value since we already handled it
						continue
					}
				}
			}

			// Recursively process the value only if it wasn't already processed as a token
			d.extractTokensFromJSON(value, newPath, parser, tokens)
		}
	case []interface{}:
		// Process each element in the array
		for i, elem := range v {
			newPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			d.extractTokensFromJSON(elem, newPath, parser, tokens)
		}
	case string:
		// Only check strings that are NOT already handled by the map processing above
		// This prevents duplicate detection of tokens that were already found as key-value pairs
		if path != "" && !strings.Contains(path, ".") {
			tokenType, confidence := d.detectTokenType(v)
			if tokenType != TokenTypeUnknown && confidence > 70 {
				*tokens = append(*tokens, Token{
					Value:      v,
					Type:       tokenType,
					Location:   LocationBody,
					Path:       path,
					Confidence: confidence,
				})
			}
		}
	}
}

// isLikelyTokenKey checks if a key name suggests it might contain a token
func (d *TokenDetector) isLikelyTokenKey(key string) bool {
	key = strings.ToLower(key)
	tokenKeywords := []string{
		"token", "key", "api", "auth", "secret", "access", "jwt", "bearer", "oauth",
		"password", "credential", "apikey", "api_key", "auth_token", "access_token",
	}

	for _, keyword := range tokenKeywords {
		if strings.Contains(key, keyword) {
			return true
		}
	}

	return false
}

// detectTokenType attempts to determine the type of a token
func (d *TokenDetector) detectTokenType(value string) (TokenType, int) {
	// Check for JWT tokens first (most specific pattern)
	if d.patterns[TokenTypeJWT].MatchString(value) {
		return TokenTypeJWT, 90
	}

	// Check for Bearer tokens (includes Bearer prefix)
	if d.patterns[TokenTypeBearer].MatchString(value) {
		return TokenTypeBearer, 90
	}

	// Check for Basic auth (includes Basic prefix)
	if d.patterns[TokenTypeBasic].MatchString(value) {
		return TokenTypeBasic, 90
	}

	// Check for API keys (hexadecimal or alphanumeric, usually longer)
	if len(value) >= 20 && len(value) <= 128 && regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(value) {
		// Additional validation to ensure it looks like a real token
		hasNumbers := regexp.MustCompile(`[0-9]`).MatchString(value)
		hasLetters := regexp.MustCompile(`[a-zA-Z]`).MatchString(value)

		// Must have both letters and numbers to be considered a token
		if hasNumbers && hasLetters {
			// Additional heuristics for API keys vs OAuth tokens
			if len(value) >= 32 || strings.Contains(strings.ToLower(value), "api") {
				return TokenTypeAPIKey, 70
			}
			// OAuth tokens are typically shorter and more random
			return TokenTypeOAuth, 65
		}
	}

	// If no specific pattern matches but it looks like it could be a token
	if len(value) >= 16 && regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(value) {
		return TokenTypeUnknown, 40
	}

	return TokenTypeUnknown, 0
}

// DetectTokens is a convenience method on ResponseParser to detect tokens
func (p *ResponseParser) DetectTokens(data []byte, headers map[string][]string) ([]Token, error) {
	detector := NewTokenDetector()

	tokens := make([]Token, 0)

	// Check headers
	headerTokens, err := detector.detectTokensInHeaders(headers)
	if err != nil {
		return nil, err
	}
	tokens = append(tokens, headerTokens...)

	// Check body if it's JSON
	if p.format == FormatJSON {
		bodyTokens, err := detector.detectTokensInJSON(data)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, bodyTokens...)
	}

	return tokens, nil
}
