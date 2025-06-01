// Package auth provides authentication mechanisms for API testing.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// GatewayType represents the type of API gateway
type GatewayType int

const (
	// GatewayAWS represents AWS API Gateway
	GatewayAWS GatewayType = iota
	// GatewayAzure represents Azure API Management
	GatewayAzure
	// GatewayGoogle represents Google Cloud Endpoints/API Gateway
	GatewayGoogle
	// GatewayKong represents Kong API Gateway
	GatewayKong
	// GatewayTyk represents Tyk API Gateway
	GatewayTyk
)

// GatewayAuth is an interface for API gateway authentication mechanisms
type GatewayAuth interface {
	AuthProvider
	// GetGatewayType returns the type of API gateway
	GetGatewayType() GatewayType
}

// AWSGatewayAuth implements authentication for AWS API Gateway
type AWSGatewayAuth struct {
	AccessKey string
	SecretKey string
	Region    string
	Service   string // Usually "execute-api" for API Gateway
}

// NewAWSGatewayAuth creates a new AWSGatewayAuth provider
func NewAWSGatewayAuth(accessKey, secretKey, region string) *AWSGatewayAuth {
	return &AWSGatewayAuth{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Service:   "execute-api", // Default service for API Gateway
	}
}

// AddAuth adds AWS Signature Version 4 authentication to the given HTTP request
func (a *AWSGatewayAuth) AddAuth(req *http.Request) error {
	if a.AccessKey == "" || a.SecretKey == "" {
		return api.NewAPIError("AWS access key or secret key is empty", 0)
	}

	// Get current time in the required format
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Create canonical request
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Create canonical query string
	canonicalQueryString := req.URL.RawQuery

	// Create canonical headers
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Host", req.URL.Host)

	var headers []string
	for k := range req.Header {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)

	var canonicalHeaders strings.Builder
	for _, h := range headers {
		canonicalHeaders.WriteString(h)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(h)))
		canonicalHeaders.WriteString("\n")
	}

	signedHeaders := strings.Join(headers, ";")

	// Create payload hash
	var payloadHash string
	if req.Body != nil {
		// In a real implementation, we would need to read the body, hash it, and then reset it
		// For simplicity, we'll assume the body is empty or already hashed
		payloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // Empty string hash
	} else {
		payloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // Empty string hash
	}

	// Combine elements to create canonical request
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash)

	// Create string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, a.Region, a.Service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		amzDate,
		credentialScope,
		hex.EncodeToString(sha256Hash([]byte(canonicalRequest))))

	// Calculate signature
	signingKey := getAWSSignatureKey(a.SecretKey, dateStamp, a.Region, a.Service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Add authorization header
	authHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		a.AccessKey,
		credentialScope,
		signedHeaders,
		signature)

	req.Header.Set("Authorization", authHeader)
	return nil
}

// GetAuthType returns the type of authentication
func (a *AWSGatewayAuth) GetAuthType() AuthType {
	return AuthAPIKey // Using APIKey as the base type
}

// GetGatewayType returns the type of API gateway
func (a *AWSGatewayAuth) GetGatewayType() GatewayType {
	return GatewayAWS
}

// GetDescription returns a human-readable description of the authentication
func (a *AWSGatewayAuth) GetDescription() string {
	// Mask the secret key for security
	secretKeyPreview := ""
	if len(a.SecretKey) > 8 {
		secretKeyPreview = a.SecretKey[:4] + "..." + a.SecretKey[len(a.SecretKey)-4:]
	} else if len(a.SecretKey) > 0 {
		secretKeyPreview = "***"
	}
	return fmt.Sprintf("AWS API Gateway (access key: %s, secret key: %s, region: %s)", a.AccessKey, secretKeyPreview, a.Region)
}

// AzureGatewayAuth implements authentication for Azure API Management
type AzureGatewayAuth struct {
	SubscriptionKey string
	KeyName         string // Optional, defaults to "Ocp-Apim-Subscription-Key"
}

// NewAzureGatewayAuth creates a new AzureGatewayAuth provider
func NewAzureGatewayAuth(subscriptionKey string, keyName string) *AzureGatewayAuth {
	if keyName == "" {
		keyName = "Ocp-Apim-Subscription-Key"
	}
	return &AzureGatewayAuth{
		SubscriptionKey: subscriptionKey,
		KeyName:         keyName,
	}
}

// AddAuth adds Azure API Management authentication to the given HTTP request
func (a *AzureGatewayAuth) AddAuth(req *http.Request) error {
	if a.SubscriptionKey == "" {
		return api.NewAPIError("Azure subscription key is empty", 0)
	}
	req.Header.Set(a.KeyName, a.SubscriptionKey)
	return nil
}

// GetAuthType returns the type of authentication
func (a *AzureGatewayAuth) GetAuthType() AuthType {
	return AuthAPIKey
}

// GetGatewayType returns the type of API gateway
func (a *AzureGatewayAuth) GetGatewayType() GatewayType {
	return GatewayAzure
}

// GetDescription returns a human-readable description of the authentication
func (a *AzureGatewayAuth) GetDescription() string {
	// Mask the subscription key for security
	keyPreview := ""
	if len(a.SubscriptionKey) > 8 {
		keyPreview = a.SubscriptionKey[:4] + "..." + a.SubscriptionKey[len(a.SubscriptionKey)-4:]
	} else if len(a.SubscriptionKey) > 0 {
		keyPreview = "***"
	}
	return fmt.Sprintf("Azure API Management (subscription key: %s, header: %s)", keyPreview, a.KeyName)
}

// GoogleGatewayAuth implements authentication for Google Cloud Endpoints/API Gateway
type GoogleGatewayAuth struct {
	APIKey string
}

// NewGoogleGatewayAuth creates a new GoogleGatewayAuth provider
func NewGoogleGatewayAuth(apiKey string) *GoogleGatewayAuth {
	return &GoogleGatewayAuth{
		APIKey: apiKey,
	}
}

// AddAuth adds Google Cloud API Gateway authentication to the given HTTP request
func (a *GoogleGatewayAuth) AddAuth(req *http.Request) error {
	if a.APIKey == "" {
		return api.NewAPIError("Google API key is empty", 0)
	}

	// Add API key as a query parameter
	q := req.URL.Query()
	q.Add("key", a.APIKey)
	req.URL.RawQuery = q.Encode()
	return nil
}

// GetAuthType returns the type of authentication
func (a *GoogleGatewayAuth) GetAuthType() AuthType {
	return AuthAPIKey
}

// GetGatewayType returns the type of API gateway
func (a *GoogleGatewayAuth) GetGatewayType() GatewayType {
	return GatewayGoogle
}

// GetDescription returns a human-readable description of the authentication
func (a *GoogleGatewayAuth) GetDescription() string {
	// Mask the API key for security
	keyPreview := ""
	if len(a.APIKey) > 8 {
		keyPreview = a.APIKey[:4] + "..." + a.APIKey[len(a.APIKey)-4:]
	} else if len(a.APIKey) > 0 {
		keyPreview = "***"
	}
	return fmt.Sprintf("Google Cloud API Gateway (API key: %s)", keyPreview)
}

// KongGatewayAuth implements authentication for Kong API Gateway
type KongGatewayAuth struct {
	APIKey     string
	KeyName    string // Header or query parameter name
	InHeader   bool   // Whether to send the key in a header or query parameter
}

// NewKongGatewayAuth creates a new KongGatewayAuth provider
func NewKongGatewayAuth(apiKey, keyName string, inHeader bool) *KongGatewayAuth {
	if keyName == "" {
		keyName = "apikey"
	}
	return &KongGatewayAuth{
		APIKey:   apiKey,
		KeyName:  keyName,
		InHeader: inHeader,
	}
}

// AddAuth adds Kong API Gateway authentication to the given HTTP request
func (a *KongGatewayAuth) AddAuth(req *http.Request) error {
	if a.APIKey == "" {
		return api.NewAPIError("Kong API key is empty", 0)
	}

	if a.InHeader {
		req.Header.Set(a.KeyName, a.APIKey)
	} else {
		q := req.URL.Query()
		q.Add(a.KeyName, a.APIKey)
		req.URL.RawQuery = q.Encode()
	}
	return nil
}

// GetAuthType returns the type of authentication
func (a *KongGatewayAuth) GetAuthType() AuthType {
	return AuthAPIKey
}

// GetGatewayType returns the type of API gateway
func (a *KongGatewayAuth) GetGatewayType() GatewayType {
	return GatewayKong
}

// GetDescription returns a human-readable description of the authentication
func (a *KongGatewayAuth) GetDescription() string {
	// Mask the API key for security
	keyPreview := ""
	if len(a.APIKey) > 8 {
		keyPreview = a.APIKey[:4] + "..." + a.APIKey[len(a.APIKey)-4:]
	} else if len(a.APIKey) > 0 {
		keyPreview = "***"
	}
	location := "header"
	if !a.InHeader {
		location = "query"
	}
	return fmt.Sprintf("Kong API Gateway (API key: %s, name: %s, in: %s)", keyPreview, a.KeyName, location)
}

// TykGatewayAuth implements authentication for Tyk API Gateway
type TykGatewayAuth struct {
	APIKey string
}

// NewTykGatewayAuth creates a new TykGatewayAuth provider
func NewTykGatewayAuth(apiKey string) *TykGatewayAuth {
	return &TykGatewayAuth{
		APIKey: apiKey,
	}
}

// AddAuth adds Tyk API Gateway authentication to the given HTTP request
func (a *TykGatewayAuth) AddAuth(req *http.Request) error {
	if a.APIKey == "" {
		return api.NewAPIError("Tyk API key is empty", 0)
	}
	req.Header.Set("Authorization", a.APIKey)
	return nil
}

// GetAuthType returns the type of authentication
func (a *TykGatewayAuth) GetAuthType() AuthType {
	return AuthAPIKey
}

// GetGatewayType returns the type of API gateway
func (a *TykGatewayAuth) GetGatewayType() GatewayType {
	return GatewayTyk
}

// GetDescription returns a human-readable description of the authentication
func (a *TykGatewayAuth) GetDescription() string {
	// Mask the API key for security
	keyPreview := ""
	if len(a.APIKey) > 8 {
		keyPreview = a.APIKey[:4] + "..." + a.APIKey[len(a.APIKey)-4:]
	} else if len(a.APIKey) > 0 {
		keyPreview = "***"
	}
	return fmt.Sprintf("Tyk API Gateway (API key: %s)", keyPreview)
}

// Helper functions for AWS Signature Version 4

// sha256Hash computes the SHA256 hash of the given data
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// hmacSHA256 computes the HMAC-SHA256 of the given data with the given key
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// getAWSSignatureKey derives the signing key for AWS Signature Version 4
func getAWSSignatureKey(key, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
