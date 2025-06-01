package parser

import (
	"testing"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestDetectTokenType(t *testing.T) {
	detector := NewTokenDetector()

	tests := []struct {
		name      string
		value     string
		wantType  TokenType
		wantConf  int
	}{
		{
			name:     "JWT Token",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantType: TokenTypeJWT,
			wantConf: 90,
		},
		{
			name:     "Bearer Token",
			value:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantType: TokenTypeBearer,
			wantConf: 90,
		},
		{
			name:     "Basic Auth",
			value:    "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			wantType: TokenTypeBasic,
			wantConf: 90,
		},
		{
			name:     "API Key",
			value:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
			wantType: TokenTypeAPIKey,
			wantConf: 60,
		},
		{
			name:     "Not a Token",
			value:    "hello world",
			wantType: TokenTypeUnknown,
			wantConf: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotConf := detector.detectTokenType(tt.value)
			if gotType != tt.wantType {
				t.Errorf("detectTokenType() type = %v, want %v", gotType, tt.wantType)
			}
			if gotConf != tt.wantConf {
				t.Errorf("detectTokenType() confidence = %v, want %v", gotConf, tt.wantConf)
			}
		})
	}
}

func TestDetectTokensInHeaders(t *testing.T) {
	detector := NewTokenDetector()

	headers := map[string][]string{
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		"X-API-Key":     {"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"},
		"Content-Type":  {"application/json"},
	}

	tokens, err := detector.detectTokensInHeaders(headers)
	if err != nil {
		t.Fatalf("detectTokensInHeaders() error = %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("detectTokensInHeaders() got %d tokens, want 2", len(tokens))
	}

	// Check first token (Bearer)
	if tokens[0].Type != TokenTypeBearer {
		t.Errorf("First token type = %v, want %v", tokens[0].Type, TokenTypeBearer)
	}
	if tokens[0].Location != LocationHeader {
		t.Errorf("First token location = %v, want %v", tokens[0].Location, LocationHeader)
	}
	if tokens[0].Name != "Authorization" {
		t.Errorf("First token name = %v, want %v", tokens[0].Name, "Authorization")
	}

	// Check second token (API Key)
	if tokens[1].Type != TokenTypeAPIKey {
		t.Errorf("Second token type = %v, want %v", tokens[1].Type, TokenTypeAPIKey)
	}
	if tokens[1].Location != LocationHeader {
		t.Errorf("Second token location = %v, want %v", tokens[1].Location, LocationHeader)
	}
	if tokens[1].Name != "X-API-Key" {
		t.Errorf("Second token name = %v, want %v", tokens[1].Name, "X-API-Key")
	}
}

func TestDetectTokensInJSON(t *testing.T) {
	detector := NewTokenDetector()

	jsonData := []byte(`{
		"auth": {
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			"api_key": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
		},
		"user": {
			"id": 123,
			"name": "John Doe"
		}
	}`)

	tokens, err := detector.detectTokensInJSON(jsonData)
	if err != nil {
		t.Fatalf("detectTokensInJSON() error = %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("detectTokensInJSON() got %d tokens, want 2", len(tokens))
	}

	// Check for JWT token
	foundJWT := false
	foundAPIKey := false

	for _, token := range tokens {
		if token.Type == TokenTypeJWT && token.Path == "$.auth.token" {
			foundJWT = true
		}
		if token.Type == TokenTypeAPIKey && token.Path == "$.auth.api_key" {
			foundAPIKey = true
		}
	}

	if !foundJWT {
		t.Errorf("JWT token not found in results")
	}
	if !foundAPIKey {
		t.Errorf("API key not found in results")
	}
}

func TestDetectTokensInResponse(t *testing.T) {
	detector := NewTokenDetector()

	// Create a mock response
	resp := &ffuf.Response{
		Headers: map[string][]string{
			"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		},
		ContentType: "application/json",
		Data: []byte(`{
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		}`),
	}

	tokens, err := detector.DetectTokensInResponse(resp)
	if err != nil {
		t.Fatalf("DetectTokensInResponse() error = %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("DetectTokensInResponse() got %d tokens, want 2", len(tokens))
	}

	// Check for header token
	foundHeader := false
	foundBody := false

	for _, token := range tokens {
		if token.Type == TokenTypeBearer && token.Location == LocationHeader {
			foundHeader = true
		}
		if token.Type == TokenTypeJWT && token.Location == LocationBody {
			foundBody = true
		}
	}

	if !foundHeader {
		t.Errorf("Header token not found in results")
	}
	if !foundBody {
		t.Errorf("Body token not found in results")
	}
}

func TestResponseParserDetectTokens(t *testing.T) {
	parser := NewResponseParser("application/json")

	headers := map[string][]string{
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	}

	data := []byte(`{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	}`)

	tokens, err := parser.DetectTokens(data, headers)
	if err != nil {
		t.Fatalf("DetectTokens() error = %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("DetectTokens() got %d tokens, want 2", len(tokens))
	}

	// Check for header token
	foundHeader := false
	foundBody := false

	for _, token := range tokens {
		if token.Type == TokenTypeBearer && token.Location == LocationHeader {
			foundHeader = true
		}
		if token.Type == TokenTypeJWT && token.Location == LocationBody {
			foundBody = true
		}
	}

	if !foundHeader {
		t.Errorf("Header token not found in results")
	}
	if !foundBody {
		t.Errorf("Body token not found in results")
	}
}
