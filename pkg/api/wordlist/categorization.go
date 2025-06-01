// Package wordlist provides specialized functionality for handling API endpoint wordlists.
//
// This file implements a comprehensive categorization system for API endpoints
// based on patterns from the api_wordlist repository.
package wordlist

import (
	"regexp"
)

// CategoryType represents the type of category
type CategoryType string

const (
	// APITypeCategory represents API architecture type categories (REST, GraphQL, etc.)
	APITypeCategory CategoryType = "api_type"
	// FunctionalCategory represents functional categories (auth, data, admin, etc.)
	FunctionalCategory CategoryType = "functional"
	// SecurityCategory represents security-related categories (sensitive, pii, etc.)
	SecurityCategory CategoryType = "security"
	// ArchitecturalCategory represents architectural categories (microservice, monolith, etc.)
	ArchitecturalCategory CategoryType = "architectural"
	// TechnologyCategory represents technology-specific categories (aws, azure, etc.)
	TechnologyCategory CategoryType = "technology"
)

// CategoryDefinition defines a category with its type, patterns, and description
type CategoryDefinition struct {
	// Name is the category name
	Name string
	// Type is the category type
	Type CategoryType
	// Patterns are regex patterns that match this category
	Patterns []*regexp.Regexp
	// Description provides information about the category
	Description string
	// Parent is the parent category (if any)
	Parent string
}

// Standard API type categories
var apiTypeCategories = []CategoryDefinition{
	{
		Name:        "rest",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/api/", "/rest/", "/v[0-9]+/"}),
		Description: "RESTful API endpoints",
	},
	{
		Name:        "graphql",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/graphql", "/gql"}),
		Description: "GraphQL API endpoints",
	},
	{
		Name:        "soap",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/soap", "/ws/", "/services/"}),
		Description: "SOAP or XML-based web services",
	},
	{
		Name:        "rpc",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/rpc", "/jsonrpc", "/api-rpc"}),
		Description: "RPC-style API endpoints",
	},
	{
		Name:        "mobile",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/mobile/", "/app/", "/device/", "/android/", "/ios/"}),
		Description: "Mobile-specific API endpoints",
	},
	{
		Name:        "internal",
		Type:        APITypeCategory,
		Patterns:    compilePatterns([]string{"/internal/", "/private/", "/_internal/", "/_private/"}),
		Description: "Internal API endpoints not intended for public use",
	},
}

// Standard functional categories
var functionalCategories = []CategoryDefinition{
	{
		Name:        "auth",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/auth", "/login", "/logout", "/register", "/signup", "/signin", "/token", "/oauth", "/password"}),
		Description: "Authentication and authorization endpoints",
	},
	{
		Name:        "user",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/user", "/users", "/profile", "/account", "/me"}),
		Description: "User management and profile endpoints",
	},
	{
		Name:        "admin",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/admin", "/manage", "/management", "/dashboard", "/console"}),
		Description: "Administrative endpoints",
	},
	{
		Name:        "data",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/data", "/query", "/search", "/filter", "/list", "/get"}),
		Description: "Data retrieval and querying endpoints",
	},
	{
		Name:        "crud",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/create", "/read", "/update", "/delete", "/add", "/edit", "/remove"}),
		Description: "CRUD operation endpoints",
	},
	{
		Name:        "file",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/file", "/files", "/upload", "/download", "/document", "/attachment", "/image", "/media"}),
		Description: "File handling endpoints",
	},
	{
		Name:        "config",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/config", "/settings", "/preferences", "/options"}),
		Description: "Configuration and settings endpoints",
	},
	{
		Name:        "payment",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/payment", "/pay", "/checkout", "/billing", "/invoice", "/subscription", "/order"}),
		Description: "Payment and billing endpoints",
	},
	{
		Name:        "notification",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/notification", "/notify", "/alert", "/message", "/email", "/sms", "/push"}),
		Description: "Notification and messaging endpoints",
	},
	{
		Name:        "analytics",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/analytics", "/stats", "/statistics", "/metrics", "/report", "/dashboard"}),
		Description: "Analytics and reporting endpoints",
	},
	{
		Name:        "health",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/health", "/status", "/ping", "/alive", "/ready", "/heartbeat"}),
		Description: "Health check and status endpoints",
	},
	{
		Name:        "webhook",
		Type:        FunctionalCategory,
		Patterns:    compilePatterns([]string{"/webhook", "/callback", "/hook", "/event"}),
		Description: "Webhook and callback endpoints",
	},
}

// Standard security categories
var securityCategories = []CategoryDefinition{
	{
		Name:        "sensitive",
		Type:        SecurityCategory,
		Patterns:    compilePatterns([]string{"/admin", "/internal", "/key", "/secret", "/token", "/password", "/auth", "/sudo"}),
		Description: "Endpoints that may contain sensitive functionality",
	},
	{
		Name:        "pii",
		Type:        SecurityCategory,
		Patterns:    compilePatterns([]string{"/user", "/profile", "/account", "/personal", "/address", "/contact", "/email", "/phone"}),
		Description: "Endpoints that may handle personally identifiable information",
	},
	{
		Name:        "payment_data",
		Type:        SecurityCategory,
		Patterns:    compilePatterns([]string{"/payment", "/card", "/credit", "/billing", "/bank", "/financial"}),
		Description: "Endpoints that may handle payment or financial data",
	},
	{
		Name:        "privileged",
		Type:        SecurityCategory,
		Patterns:    compilePatterns([]string{"/admin", "/superuser", "/root", "/sudo", "/manage", "/control"}),
		Description: "Endpoints that may require elevated privileges",
	},
}

// Standard architectural categories
var architecturalCategories = []CategoryDefinition{
	{
		Name:        "microservice",
		Type:        ArchitecturalCategory,
		Patterns:    compilePatterns([]string{"/service/", "/services/", "/ms/", "/api/[a-z-]+/v[0-9]+/"}),
		Description: "Endpoints that appear to be part of a microservice architecture",
	},
	{
		Name:        "gateway",
		Type:        ArchitecturalCategory,
		Patterns:    compilePatterns([]string{"/gateway", "/gw/", "/api-gw/", "/proxy/"}),
		Description: "API gateway or proxy endpoints",
	},
	{
		Name:        "legacy",
		Type:        ArchitecturalCategory,
		Patterns:    compilePatterns([]string{"/legacy", "/old", "/deprecated", "/v[0-9]+/"}),
		Description: "Legacy or deprecated API endpoints",
	},
}

// Standard technology categories
var technologyCategories = []CategoryDefinition{
	{
		Name:        "aws",
		Type:        TechnologyCategory,
		Patterns:    compilePatterns([]string{"/aws", "/amazon", "/s3", "/ec2", "/lambda", "/dynamodb"}),
		Description: "AWS-related API endpoints",
	},
	{
		Name:        "azure",
		Type:        TechnologyCategory,
		Patterns:    compilePatterns([]string{"/azure", "/microsoft", "/ms-", "/blob", "/functions"}),
		Description: "Azure-related API endpoints",
	},
	{
		Name:        "gcp",
		Type:        TechnologyCategory,
		Patterns:    compilePatterns([]string{"/gcp", "/google", "/firebase", "/cloud", "/storage", "/bigquery"}),
		Description: "Google Cloud Platform-related API endpoints",
	},
	{
		Name:        "salesforce",
		Type:        TechnologyCategory,
		Patterns:    compilePatterns([]string{"/salesforce", "/sfdc", "/sf-", "/force"}),
		Description: "Salesforce-related API endpoints",
	},
}

// All predefined categories
var allCategories = []CategoryDefinition{}

// Initialize all categories
func init() {
	// Combine all category types
	allCategories = append(allCategories, apiTypeCategories...)
	allCategories = append(allCategories, functionalCategories...)
	allCategories = append(allCategories, securityCategories...)
	allCategories = append(allCategories, architecturalCategories...)
	allCategories = append(allCategories, technologyCategories...)
}

// compilePatterns compiles a list of string patterns into regular expressions
func compilePatterns(patterns []string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		// Make the pattern case-insensitive
		result[i] = regexp.MustCompile("(?i)" + pattern)
	}
	return result
}

// CategorizeEndpoint categorizes an API endpoint based on its path
func CategorizeEndpoint(path string) []string {
	categories := make([]string, 0)

	// Check each category definition
	for _, category := range allCategories {
		for _, pattern := range category.Patterns {
			if pattern.MatchString(path) {
				// Add the category if it's not already in the list
				categoryName := string(category.Type) + ":" + category.Name
				if !contains(categories, categoryName) {
					categories = append(categories, categoryName)
				}
				break // No need to check other patterns for this category
			}
		}
	}

	return categories
}

// CategorizeEndpoints categorizes multiple API endpoints
func CategorizeEndpoints(paths []string) map[string][]string {
	result := make(map[string][]string)

	for _, path := range paths {
		result[path] = CategorizeEndpoint(path)
	}

	return result
}

// GetCategoryDefinition returns the definition for a specific category
func GetCategoryDefinition(categoryType CategoryType, name string) (CategoryDefinition, bool) {
	for _, category := range allCategories {
		if category.Type == categoryType && category.Name == name {
			return category, true
		}
	}

	return CategoryDefinition{}, false
}

// GetAllCategoryDefinitions returns all category definitions
func GetAllCategoryDefinitions() []CategoryDefinition {
	return allCategories
}

// GetCategoryDefinitionsByType returns all category definitions of a specific type
func GetCategoryDefinitionsByType(categoryType CategoryType) []CategoryDefinition {
	result := make([]CategoryDefinition, 0)

	for _, category := range allCategories {
		if category.Type == categoryType {
			result = append(result, category)
		}
	}

	return result
}

// EnhanceAPIWordlist adds comprehensive categorization to an existing APIWordlist
func EnhanceAPIWordlist(wl *APIWordlist) {
	// Process each endpoint
	for i, endpoint := range wl.entries {
		// Get additional categories
		additionalCategories := CategorizeEndpoint(endpoint.Path)

		// Add new categories
		for _, category := range additionalCategories {
			if !contains(endpoint.Categories, category) {
				wl.entries[i].Categories = append(wl.entries[i].Categories, category)

				// Update the category index
				wl.categories[category] = append(wl.categories[category], i)
			}
		}
	}
}

// GetEndpointsByCategory returns all endpoints in a specific category
func (w *APIWordlist) GetEndpointsByCategory(categoryType CategoryType, name string) []APIEndpoint {
	categoryName := string(categoryType) + ":" + name
	indices, exists := w.categories[categoryName]
	if !exists {
		return []APIEndpoint{}
	}

	result := make([]APIEndpoint, len(indices))
	for i, idx := range indices {
		result[i] = w.entries[idx]
	}

	return result
}

// GetEndpointsByCategoryType returns all endpoints in categories of a specific type
func (w *APIWordlist) GetEndpointsByCategoryType(categoryType CategoryType) []APIEndpoint {
	// Get all categories of this type
	categories := GetCategoryDefinitionsByType(categoryType)

	// Create a map to avoid duplicates
	endpointMap := make(map[int]bool)

	// Collect all endpoints in these categories
	for _, category := range categories {
		categoryName := string(categoryType) + ":" + category.Name
		indices, exists := w.categories[categoryName]
		if exists {
			for _, idx := range indices {
				endpointMap[idx] = true
			}
		}
	}

	// Convert map keys to a slice of endpoints
	result := make([]APIEndpoint, 0, len(endpointMap))
	for idx := range endpointMap {
		result = append(result, w.entries[idx])
	}

	return result
}

// GetCategoryTypes returns all available category types
func GetCategoryTypes() []CategoryType {
	return []CategoryType{
		APITypeCategory,
		FunctionalCategory,
		SecurityCategory,
		ArchitecturalCategory,
		TechnologyCategory,
	}
}

// GetCategoriesByType returns all category names of a specific type
func GetCategoriesByType(categoryType CategoryType) []string {
	categories := GetCategoryDefinitionsByType(categoryType)
	result := make([]string, len(categories))

	for i, category := range categories {
		result[i] = category.Name
	}

	return result
}
