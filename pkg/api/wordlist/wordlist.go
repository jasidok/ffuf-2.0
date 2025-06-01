// Package wordlist provides specialized functionality for handling API endpoint wordlists.
//
// This package extends ffuf's wordlist capabilities with API-specific features
// such as endpoint categorization, pattern matching, and integration with
// the api_wordlist repository format.
package wordlist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// Common API path patterns
var (
	// Common API prefixes
	CommonAPIPrefixes = []string{
		"/api/", "/v1/", "/v2/", "/v3/", "/rest/", "/graphql", "/gql",
		"/service/", "/services/", "/app/", "/apps/", "/integration/",
	}

	// Common API endpoint patterns
	CommonAPIPatterns = []string{
		"users", "auth", "login", "logout", "register", "profile",
		"admin", "config", "settings", "status", "health", "metrics",
		"data", "query", "search", "upload", "download", "file",
		"payment", "subscription", "webhook", "callback", "notification",
	}

	// API version pattern
	APIVersionPattern = regexp.MustCompile(`/v[0-9]+/`)
)

// APIEndpoint represents a single API endpoint with metadata
type APIEndpoint struct {
	Path       string            // The endpoint path
	Method     string            // HTTP method (GET, POST, etc.)
	Categories []string          // Categories this endpoint belongs to
	Parameters map[string]string // Parameters for this endpoint
}

// APIWordlist represents a wordlist specifically designed for API endpoint discovery
type APIWordlist struct {
	entries         []APIEndpoint
	categories      map[string][]int
	prefixIndex     map[string][]int
	patternIndex    map[string][]int
	config          *ffuf.Config
	position        int
	active          bool

	// api_wordlist repository specific fields
	actions         []string
	objects         []string
	wildcards       []string
	seenInWild      []string
}

// NewAPIWordlist creates a new APIWordlist from the given file path
func NewAPIWordlist(path string, conf *ffuf.Config) (*APIWordlist, error) {
	wl := &APIWordlist{
		entries:      make([]APIEndpoint, 0),
		categories:   make(map[string][]int),
		prefixIndex:  make(map[string][]int),
		patternIndex: make(map[string][]int),
		config:       conf,
		position:     0,
		active:       true,
		actions:      make([]string, 0),
		objects:      make([]string, 0),
		wildcards:    make([]string, 0),
		seenInWild:   make([]string, 0),
	}

	// Check if path is a directory (api_wordlist repository format)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, api.NewAPIError("Failed to access wordlist path: "+err.Error(), 0)
	}

	if fileInfo.IsDir() {
		// Process as api_wordlist repository
		err = wl.loadAPIWordlistRepo(path)
	} else {
		// Process as single wordlist file
		err = wl.loadWordlistFile(path)
	}

	if err != nil {
		return nil, err
	}

	return wl, nil
}

// loadAPIWordlistRepo loads API endpoints from an api_wordlist repository directory
func (w *APIWordlist) loadAPIWordlistRepo(dirPath string) error {
	// Check if this is an api_wordlist repository by looking for specific files
	isAPIWordlistRepo := false

	// List of key files to check for
	keyFiles := []string{
		"actions.txt",
		"objects.txt",
		"api_seen_in_wild.txt",
	}

	for _, file := range keyFiles {
		if _, err := os.Stat(filepath.Join(dirPath, file)); err == nil {
			isAPIWordlistRepo = true
			break
		}
	}

	if isAPIWordlistRepo {
		// This is an api_wordlist repository, handle it specifically
		return w.loadAPIWordlistSpecificRepo(dirPath)
	}

	// Otherwise, treat as a generic directory structure
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .txt files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".txt") {
			return nil
		}

		// Extract category from relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Use directory name as category
		category := filepath.Dir(relPath)
		if category == "." {
			category = "general"
		}

		// Load the file
		return w.loadWordlistFileWithCategory(path, category)
	})
}

// loadAPIWordlistSpecificRepo loads API endpoints from an api_wordlist repository
func (w *APIWordlist) loadAPIWordlistSpecificRepo(dirPath string) error {
	// Load actions files
	actionFiles := []string{
		"actions.txt",
		"actions-lowercase.txt",
		"actions-uppercase.txt",
	}

	for _, file := range actionFiles {
		filePath := filepath.Join(dirPath, file)
		if _, err := os.Stat(filePath); err == nil {
			if err := w.loadActionsFile(filePath); err != nil {
				return err
			}
		}
	}

	// Load objects files
	objectFiles := []string{
		"objects.txt",
		"objects-lowercase.txt",
		"objects-uppercase.txt",
	}

	for _, file := range objectFiles {
		filePath := filepath.Join(dirPath, file)
		if _, err := os.Stat(filePath); err == nil {
			if err := w.loadObjectsFile(filePath); err != nil {
				return err
			}
		}
	}

	// Load seen in wild file
	wildFilePath := filepath.Join(dirPath, "api_seen_in_wild.txt")
	if _, err := os.Stat(wildFilePath); err == nil {
		if err := w.loadSeenInWildFile(wildFilePath); err != nil {
			return err
		}
	}

	// Load common paths file
	commonPathsFile := filepath.Join(dirPath, "common_paths.txt")
	if _, err := os.Stat(commonPathsFile); err == nil {
		if err := w.loadWordlistFileWithCategory(commonPathsFile, "common"); err != nil {
			return err
		}
	}

	// Generate API endpoints by combining actions and objects
	w.generateAPIEndpoints()

	return nil
}

// loadActionsFile loads API function name verbs from an actions file
func (w *APIWordlist) loadActionsFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return api.NewAPIError("Failed to open actions file: "+err.Error(), 0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		w.actions = append(w.actions, line)
	}

	if err := scanner.Err(); err != nil {
		return api.NewAPIError("Error reading actions file: "+err.Error(), 0)
	}

	return nil
}

// loadObjectsFile loads API function name nouns from an objects file
func (w *APIWordlist) loadObjectsFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return api.NewAPIError("Failed to open objects file: "+err.Error(), 0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		w.objects = append(w.objects, line)
	}

	if err := scanner.Err(); err != nil {
		return api.NewAPIError("Error reading objects file: "+err.Error(), 0)
	}

	return nil
}

// loadSeenInWildFile loads API function names seen in the wild
func (w *APIWordlist) loadSeenInWildFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return api.NewAPIError("Failed to open seen-in-wild file: "+err.Error(), 0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Add to seen in wild list
		w.seenInWild = append(w.seenInWild, line)

		// Also process as a regular endpoint with "seen_in_wild" category
		w.processLine(line, "seen_in_wild")
	}

	if err := scanner.Err(); err != nil {
		return api.NewAPIError("Error reading seen-in-wild file: "+err.Error(), 0)
	}

	return nil
}

// generateAPIEndpoints generates API endpoints by combining actions and objects
func (w *APIWordlist) generateAPIEndpoints() {
	// Common API path patterns
	patterns := []string{
		"/api/%s/%s",
		"/api/v1/%s/%s",
		"/api/v2/%s/%s",
		"/v1/%s/%s",
		"/v2/%s/%s",
		"/%s/%s",
	}

	// Generate endpoints by combining actions and objects
	for _, action := range w.actions {
		for _, object := range w.objects {
			for _, pattern := range patterns {
				endpoint := fmt.Sprintf(pattern, action, object)

				// Add as a generated endpoint with appropriate categories
				w.processLine(endpoint, "generated,action_object")
			}
		}
	}

	// Note: Seen in wild entries are already added in loadSeenInWildFile
}

// loadWordlistFile loads API endpoints from a single wordlist file
func (w *APIWordlist) loadWordlistFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return api.NewAPIError("Failed to open wordlist file: "+err.Error(), 0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		w.processLine(line, "")
	}

	if err := scanner.Err(); err != nil {
		return api.NewAPIError("Error reading wordlist file: "+err.Error(), 0)
	}

	return nil
}

// loadWordlistFileWithCategory loads API endpoints from a file with a specific category
func (w *APIWordlist) loadWordlistFileWithCategory(path, category string) error {
	file, err := os.Open(path)
	if err != nil {
		return api.NewAPIError("Failed to open wordlist file: "+err.Error(), 0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		w.processLine(line, category)
	}

	if err := scanner.Err(); err != nil {
		return api.NewAPIError("Error reading wordlist file: "+err.Error(), 0)
	}

	return nil
}

// processLine processes a single line from a wordlist file
func (w *APIWordlist) processLine(line, defaultCategory string) {
	var endpoint APIEndpoint
	var categories []string

	// Check if line contains category information (format: endpoint:category)
	parts := strings.SplitN(line, ":", 2)
	endpoint.Path = parts[0]

	// If category is specified in the line, use it
	if len(parts) > 1 {
		categories = append(categories, strings.Split(parts[1], ",")...)
	}

	// If a default category was provided, add it
	if defaultCategory != "" {
		categories = append(categories, defaultCategory)
	}

	// If no categories are specified, try to infer from the path
	if len(categories) == 0 {
		categories = w.inferCategories(endpoint.Path)
	}

	endpoint.Categories = categories
	endpoint.Parameters = make(map[string]string)

	// Try to infer HTTP method from the path or endpoint name
	endpoint.Method = w.inferMethod(endpoint.Path)

	// Add the endpoint to the wordlist
	w.entries = append(w.entries, endpoint)
	index := len(w.entries) - 1

	// Index by categories
	for _, category := range categories {
		w.categories[category] = append(w.categories[category], index)
	}

	// Index by common API prefixes
	for _, prefix := range CommonAPIPrefixes {
		if strings.Contains(endpoint.Path, prefix) {
			w.prefixIndex[prefix] = append(w.prefixIndex[prefix], index)
		}
	}

	// Index by path segments for more comprehensive prefix matching
	segments := strings.Split(endpoint.Path, "/")
	currentPath := ""
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		currentPath += "/" + segment
		w.prefixIndex[currentPath] = append(w.prefixIndex[currentPath], index)
	}

	// Index by common API patterns
	for _, pattern := range CommonAPIPatterns {
		if strings.Contains(strings.ToLower(endpoint.Path), pattern) {
			w.patternIndex[pattern] = append(w.patternIndex[pattern], index)
		}
	}

	// Index by action-object combinations if this is from an api_wordlist repository
	pathLower := strings.ToLower(endpoint.Path)

	// Check for actions
	for _, action := range w.actions {
		actionLower := strings.ToLower(action)
		if strings.Contains(pathLower, actionLower) {
			// Add to pattern index for this action
			actionPattern := "action:" + actionLower
			w.patternIndex[actionPattern] = append(w.patternIndex[actionPattern], index)

			// Add action category
			if !contains(categories, "action:"+actionLower) {
				categories = append(categories, "action:"+actionLower)
				w.categories["action:"+actionLower] = append(w.categories["action:"+actionLower], index)
			}
		}
	}

	// Check for objects
	for _, object := range w.objects {
		objectLower := strings.ToLower(object)
		if strings.Contains(pathLower, objectLower) {
			// Add to pattern index for this object
			objectPattern := "object:" + objectLower
			w.patternIndex[objectPattern] = append(w.patternIndex[objectPattern], index)

			// Add object category
			if !contains(categories, "object:"+objectLower) {
				categories = append(categories, "object:"+objectLower)
				w.categories["object:"+objectLower] = append(w.categories["object:"+objectLower], index)
			}
		}
	}

	// Update the endpoint categories with any new inferred categories
	w.entries[index].Categories = categories
}

// inferCategories attempts to determine categories for an API endpoint based on its path
func (w *APIWordlist) inferCategories(path string) []string {
	categories := make([]string, 0)

	// Check for API version
	if APIVersionPattern.MatchString(path) {
		categories = append(categories, "versioned")
	}

	// Check for common patterns
	pathLower := strings.ToLower(path)
	for _, pattern := range CommonAPIPatterns {
		if strings.Contains(pathLower, pattern) {
			categories = append(categories, pattern)
		}
	}

	// If no categories were inferred, use "general"
	if len(categories) == 0 {
		categories = append(categories, "general")
	}

	return categories
}

// inferMethod attempts to determine the HTTP method for an API endpoint based on its path
func (w *APIWordlist) inferMethod(path string) string {
	pathLower := strings.ToLower(path)

	// Check for common patterns that suggest specific HTTP methods
	if strings.Contains(pathLower, "get") || 
	   strings.Contains(pathLower, "list") || 
	   strings.Contains(pathLower, "search") || 
	   strings.Contains(pathLower, "find") {
		return "GET"
	}

	if strings.Contains(pathLower, "create") || 
	   strings.Contains(pathLower, "add") || 
	   strings.Contains(pathLower, "register") || 
	   strings.Contains(pathLower, "upload") {
		return "POST"
	}

	if strings.Contains(pathLower, "update") || 
	   strings.Contains(pathLower, "edit") || 
	   strings.Contains(pathLower, "modify") {
		return "PUT"
	}

	if strings.Contains(pathLower, "delete") || 
	   strings.Contains(pathLower, "remove") {
		return "DELETE"
	}

	// Default to GET if no method could be inferred
	return "GET"
}

// GetEntries returns all entries in the wordlist as strings
func (w *APIWordlist) GetEntries() []string {
	result := make([]string, len(w.entries))
	for i, endpoint := range w.entries {
		result[i] = endpoint.Path
	}
	return result
}

// GetEntriesByCategory returns all entries in the specified category as strings
func (w *APIWordlist) GetEntriesByCategory(category string) []string {
	indices, exists := w.categories[category]
	if !exists {
		return []string{}
	}

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = w.entries[idx].Path
	}
	return result
}

// GetNextWithPrefix returns the next entry that contains the specified prefix
func (w *APIWordlist) GetNextWithPrefix(prefix string) string {
	indices, exists := w.prefixIndex[prefix]
	if !exists || len(indices) == 0 {
		return ""
	}

	// Return the first entry and remove it from the index
	index := indices[0]
	w.prefixIndex[prefix] = indices[1:]

	return w.entries[index].Path
}

// GetCategories returns all available categories in the wordlist
func (w *APIWordlist) GetCategories() []string {
	categories := make([]string, 0, len(w.categories))
	for category := range w.categories {
		categories = append(categories, category)
	}
	return categories
}

// Position returns the current position in the wordlist
func (w *APIWordlist) Position() int {
	return w.position
}

// SetPosition sets the current position in the wordlist
func (w *APIWordlist) SetPosition(pos int) {
	w.position = pos
}

// Next returns true if there are more entries in the wordlist
func (w *APIWordlist) Next() bool {
	return w.position < len(w.entries)
}

// Value returns the current entry in the wordlist
func (w *APIWordlist) Value() string {
	if w.position >= len(w.entries) {
		return ""
	}
	return w.entries[w.position].Path
}

// IncrementPosition moves to the next entry in the wordlist
func (w *APIWordlist) IncrementPosition() {
	w.position++
}

// Total returns the total number of entries in the wordlist
func (w *APIWordlist) Total() int {
	return len(w.entries)
}

// Active returns whether the wordlist is active
func (w *APIWordlist) Active() bool {
	return w.active
}

// Enable activates the wordlist
func (w *APIWordlist) Enable() {
	w.active = true
}

// Disable deactivates the wordlist
func (w *APIWordlist) Disable() {
	w.active = false
}

// GetEndpoint returns the full APIEndpoint at the specified index
func (w *APIWordlist) GetEndpoint(index int) (APIEndpoint, error) {
	if index < 0 || index >= len(w.entries) {
		return APIEndpoint{}, api.NewAPIError("Index out of range", 0)
	}
	return w.entries[index], nil
}

// GetCurrentEndpoint returns the full APIEndpoint at the current position
func (w *APIWordlist) GetCurrentEndpoint() (APIEndpoint, error) {
	return w.GetEndpoint(w.position)
}

// GetActions returns all API function name verbs
func (w *APIWordlist) GetActions() []string {
	return w.actions
}

// GetObjects returns all API function name nouns
func (w *APIWordlist) GetObjects() []string {
	return w.objects
}

// GetSeenInWild returns all API function names seen in the wild
func (w *APIWordlist) GetSeenInWild() []string {
	return w.seenInWild
}

// IsAPIWordlistRepo returns true if this wordlist was loaded from an api_wordlist repository
func (w *APIWordlist) IsAPIWordlistRepo() bool {
	return len(w.actions) > 0 || len(w.objects) > 0 || len(w.seenInWild) > 0
}

// GetEntriesByPrefix returns all entries that start with the given prefix
func (w *APIWordlist) GetEntriesByPrefix(prefix string) []string {
	indices, exists := w.prefixIndex[prefix]
	if !exists {
		return []string{}
	}

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = w.entries[idx].Path
	}
	return result
}

// GetEntriesByPattern returns all entries that contain the given pattern
func (w *APIWordlist) GetEntriesByPattern(pattern string) []string {
	indices, exists := w.patternIndex[pattern]
	if !exists {
		return []string{}
	}

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = w.entries[idx].Path
	}
	return result
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
