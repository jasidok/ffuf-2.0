// Package wordlist provides specialized functionality for handling API endpoint wordlists.
//
// This file implements intelligent pattern matching using api_wordlist signature detection
// to identify and categorize API endpoints based on common patterns.
package wordlist

import (
	"regexp"
	"strings"
)

// PatternSignature represents a signature pattern for API endpoint detection
type PatternSignature struct {
	// Name is the signature name
	Name string
	// Pattern is the regex pattern for matching
	Pattern *regexp.Regexp
	// Description provides information about the signature
	Description string
	// Category is the category this signature belongs to
	Category string
	// Confidence is the confidence level (0-100) for this signature
	Confidence int
}

// Common API signature patterns
var (
	// REST API patterns
	RestAPISignatures = []PatternSignature{
		{
			Name:        "standard_rest",
			Pattern:     regexp.MustCompile(`(?i)/api/v[0-9]+/[a-z0-9_-]+/?$`),
			Description: "Standard REST API endpoint with version",
			Category:    "api_type:rest",
			Confidence:  90,
		},
		{
			Name:        "resource_rest",
			Pattern:     regexp.MustCompile(`(?i)/api/[a-z0-9_-]+/[0-9]+/?$`),
			Description: "REST API resource with ID",
			Category:    "api_type:rest",
			Confidence:  85,
		},
		{
			Name:        "nested_rest",
			Pattern:     regexp.MustCompile(`(?i)/api/[a-z0-9_-]+/[0-9]+/[a-z0-9_-]+/?$`),
			Description: "Nested REST API resource",
			Category:    "api_type:rest",
			Confidence:  80,
		},
	}

	// GraphQL API patterns
	GraphQLSignatures = []PatternSignature{
		{
			Name:        "graphql_endpoint",
			Pattern:     regexp.MustCompile(`(?i)/graphql/?$`),
			Description: "Standard GraphQL endpoint",
			Category:    "api_type:graphql",
			Confidence:  95,
		},
		{
			Name:        "gql_endpoint",
			Pattern:     regexp.MustCompile(`(?i)/gql/?$`),
			Description: "Shortened GraphQL endpoint",
			Category:    "api_type:graphql",
			Confidence:  90,
		},
		{
			Name:        "graphql_api",
			Pattern:     regexp.MustCompile(`(?i)/api/graphql/?$`),
			Description: "GraphQL endpoint under API path",
			Category:    "api_type:graphql",
			Confidence:  85,
		},
	}

	// Mobile API patterns
	MobileAPISignatures = []PatternSignature{
		{
			Name:        "mobile_api",
			Pattern:     regexp.MustCompile(`(?i)/mobile/api/`),
			Description: "Mobile-specific API endpoint",
			Category:    "api_type:mobile",
			Confidence:  90,
		},
		{
			Name:        "app_api",
			Pattern:     regexp.MustCompile(`(?i)/app/api/`),
			Description: "App-specific API endpoint",
			Category:    "api_type:mobile",
			Confidence:  85,
		},
		{
			Name:        "android_api",
			Pattern:     regexp.MustCompile(`(?i)/android/`),
			Description: "Android-specific API endpoint",
			Category:    "api_type:mobile",
			Confidence:  90,
		},
		{
			Name:        "ios_api",
			Pattern:     regexp.MustCompile(`(?i)/ios/`),
			Description: "iOS-specific API endpoint",
			Category:    "api_type:mobile",
			Confidence:  90,
		},
	}

	// Authentication patterns
	AuthSignatures = []PatternSignature{
		{
			Name:        "oauth_token",
			Pattern:     regexp.MustCompile(`(?i)/oauth/token/?$`),
			Description: "OAuth token endpoint",
			Category:    "functional:auth",
			Confidence:  95,
		},
		{
			Name:        "login_endpoint",
			Pattern:     regexp.MustCompile(`(?i)/login/?$`),
			Description: "Login endpoint",
			Category:    "functional:auth",
			Confidence:  90,
		},
		{
			Name:        "auth_endpoint",
			Pattern:     regexp.MustCompile(`(?i)/auth/?$`),
			Description: "Authentication endpoint",
			Category:    "functional:auth",
			Confidence:  90,
		},
		{
			Name:        "jwt_endpoint",
			Pattern:     regexp.MustCompile(`(?i)/jwt/`),
			Description: "JWT-related endpoint",
			Category:    "functional:auth",
			Confidence:  85,
		},
	}

	// CRUD operation patterns
	CRUDSignatures = []PatternSignature{
		{
			Name:        "create_operation",
			Pattern:     regexp.MustCompile(`(?i)/(create|add|new)/`),
			Description: "Create operation endpoint",
			Category:    "functional:crud",
			Confidence:  85,
		},
		{
			Name:        "read_operation",
			Pattern:     regexp.MustCompile(`(?i)/(read|get|view|show)/`),
			Description: "Read operation endpoint",
			Category:    "functional:crud",
			Confidence:  85,
		},
		{
			Name:        "update_operation",
			Pattern:     regexp.MustCompile(`(?i)/(update|edit|modify)/`),
			Description: "Update operation endpoint",
			Category:    "functional:crud",
			Confidence:  85,
		},
		{
			Name:        "delete_operation",
			Pattern:     regexp.MustCompile(`(?i)/(delete|remove|destroy)/`),
			Description: "Delete operation endpoint",
			Category:    "functional:crud",
			Confidence:  85,
		},
	}

	// All signatures combined
	AllSignatures = []PatternSignature{}
)

// Initialize all signatures
func init() {
	// Combine all signature types
	AllSignatures = append(AllSignatures, RestAPISignatures...)
	AllSignatures = append(AllSignatures, GraphQLSignatures...)
	AllSignatures = append(AllSignatures, MobileAPISignatures...)
	AllSignatures = append(AllSignatures, AuthSignatures...)
	AllSignatures = append(AllSignatures, CRUDSignatures...)
}

// SignatureMatch represents a match between an endpoint and a signature
type SignatureMatch struct {
	// Path is the endpoint path
	Path string
	// Signature is the matched signature
	Signature PatternSignature
	// Confidence is the confidence level for this match
	Confidence int
}

// MatchSignatures matches an endpoint path against all signatures
func MatchSignatures(path string) []SignatureMatch {
	matches := make([]SignatureMatch, 0)
	
	for _, sig := range AllSignatures {
		if sig.Pattern.MatchString(path) {
			matches = append(matches, SignatureMatch{
				Path:       path,
				Signature:  sig,
				Confidence: sig.Confidence,
			})
		}
	}
	
	return matches
}

// MatchSignaturesWithThreshold matches an endpoint path against all signatures with a minimum confidence threshold
func MatchSignaturesWithThreshold(path string, minConfidence int) []SignatureMatch {
	matches := make([]SignatureMatch, 0)
	
	for _, sig := range AllSignatures {
		if sig.Pattern.MatchString(path) && sig.Confidence >= minConfidence {
			matches = append(matches, SignatureMatch{
				Path:       path,
				Signature:  sig,
				Confidence: sig.Confidence,
			})
		}
	}
	
	return matches
}

// GetSignaturesByCategory returns all signatures for a specific category
func GetSignaturesByCategory(category string) []PatternSignature {
	signatures := make([]PatternSignature, 0)
	
	for _, sig := range AllSignatures {
		if sig.Category == category {
			signatures = append(signatures, sig)
		}
	}
	
	return signatures
}

// EnhanceAPIWordlistWithSignatures adds signature-based categorization to an existing APIWordlist
func EnhanceAPIWordlistWithSignatures(wl *APIWordlist, minConfidence int) {
	// Process each endpoint
	for i, endpoint := range wl.entries {
		// Match signatures
		matches := MatchSignaturesWithThreshold(endpoint.Path, minConfidence)
		
		// Add categories from matched signatures
		for _, match := range matches {
			category := match.Signature.Category
			
			// Add category if it doesn't already exist
			if !contains(endpoint.Categories, category) {
				wl.entries[i].Categories = append(wl.entries[i].Categories, category)
				
				// Update the category index
				wl.categories[category] = append(wl.categories[category], i)
			}
		}
	}
}

// DetectAPIType detects the API type of an endpoint path
func DetectAPIType(path string) string {
	// Check for GraphQL first (most specific)
	for _, sig := range GraphQLSignatures {
		if sig.Pattern.MatchString(path) {
			return "graphql"
		}
	}
	
	// Check for Mobile API
	for _, sig := range MobileAPISignatures {
		if sig.Pattern.MatchString(path) {
			return "mobile"
		}
	}
	
	// Check for REST API (most common)
	for _, sig := range RestAPISignatures {
		if sig.Pattern.MatchString(path) {
			return "rest"
		}
	}
	
	// Default to REST if it contains common REST patterns
	if strings.Contains(path, "/api/") || strings.Contains(path, "/v1/") || strings.Contains(path, "/v2/") {
		return "rest"
	}
	
	// Unknown API type
	return "unknown"
}

// DetectEndpointFunction detects the function of an endpoint path
func DetectEndpointFunction(path string) string {
	pathLower := strings.ToLower(path)
	
	// Check for authentication endpoints
	for _, sig := range AuthSignatures {
		if sig.Pattern.MatchString(path) {
			return "auth"
		}
	}
	
	// Check for CRUD operations
	for _, sig := range CRUDSignatures {
		if sig.Pattern.MatchString(path) {
			return "crud"
		}
	}
	
	// Check for common patterns
	if strings.Contains(pathLower, "user") || strings.Contains(pathLower, "profile") {
		return "user"
	}
	
	if strings.Contains(pathLower, "admin") || strings.Contains(pathLower, "manage") {
		return "admin"
	}
	
	if strings.Contains(pathLower, "file") || strings.Contains(pathLower, "upload") || strings.Contains(pathLower, "download") {
		return "file"
	}
	
	if strings.Contains(pathLower, "payment") || strings.Contains(pathLower, "billing") {
		return "payment"
	}
	
	if strings.Contains(pathLower, "config") || strings.Contains(pathLower, "setting") {
		return "config"
	}
	
	// Unknown function
	return "unknown"
}

// GenerateSignatureReport generates a report of signature matches for a list of paths
func GenerateSignatureReport(paths []string, minConfidence int) map[string][]SignatureMatch {
	report := make(map[string][]SignatureMatch)
	
	for _, path := range paths {
		matches := MatchSignaturesWithThreshold(path, minConfidence)
		if len(matches) > 0 {
			report[path] = matches
		}
	}
	
	return report
}

// AddCustomSignature adds a custom signature to the list of signatures
func AddCustomSignature(signature PatternSignature) {
	AllSignatures = append(AllSignatures, signature)
}

// CreateCustomSignature creates a new signature with the given parameters
func CreateCustomSignature(name, pattern, description, category string, confidence int) (PatternSignature, error) {
	// Compile the pattern
	compiledPattern, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return PatternSignature{}, err
	}
	
	// Create the signature
	signature := PatternSignature{
		Name:        name,
		Pattern:     compiledPattern,
		Description: description,
		Category:    category,
		Confidence:  confidence,
	}
	
	return signature, nil
}