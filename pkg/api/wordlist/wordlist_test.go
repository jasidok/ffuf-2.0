package wordlist

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestAPIWordlistFromFile(t *testing.T) {
	// Create a temporary wordlist file
	content := `/api/users
/api/users/:id
/api/auth/login
/api/products:products
/api/orders:orders
/v1/settings
/graphql
/rest/data
`
	tmpfile, err := ioutil.TempFile("", "api_wordlist_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Create a config for testing
	config := &ffuf.Config{}

	// Create a new APIWordlist
	wl, err := NewAPIWordlist(tmpfile.Name(), config)
	if err != nil {
		t.Fatalf("Failed to create APIWordlist: %v", err)
	}

	// Test total entries
	if wl.Total() != 8 {
		t.Errorf("Expected 8 entries, got %d", wl.Total())
	}

	// Test categories
	categories := wl.GetCategories()
	if len(categories) < 3 {
		t.Errorf("Expected at least 3 categories, got %d", len(categories))
	}

	// Test entries by category
	products := wl.GetEntriesByCategory("products")
	if len(products) != 1 || products[0] != "/api/products" {
		t.Errorf("Expected 1 product entry, got %d: %v", len(products), products)
	}

	// Test prefix lookup
	apiEntries := wl.GetEntriesByPrefix("/api/")
	if len(apiEntries) != 5 {
		t.Errorf("Expected 5 API entries, got %d: %v", len(apiEntries), apiEntries)
	}

	// Test pattern lookup
	userEntries := wl.GetEntriesByPattern("users")
	if len(userEntries) != 2 {
		t.Errorf("Expected 2 user entries, got %d: %v", len(userEntries), userEntries)
	}

	// Test iteration
	count := 0
	wl.ResetPosition()
	for wl.Next() {
		entry := wl.Value()
		if entry == "" {
			t.Errorf("Got empty entry at position %d", wl.Position())
		}
		count++
		wl.IncrementPosition()
	}
	if count != 8 {
		t.Errorf("Expected to iterate through 8 entries, got %d", count)
	}

	// Test endpoint metadata
	wl.ResetPosition()
	endpoint, err := wl.GetCurrentEndpoint()
	if err != nil {
		t.Fatalf("Failed to get current endpoint: %v", err)
	}
	if endpoint.Path == "" || endpoint.Method == "" {
		t.Errorf("Expected non-empty path and method, got path=%s, method=%s", endpoint.Path, endpoint.Method)
	}
}

func TestAPIWordlistFromRepo(t *testing.T) {
	// Create a temporary directory structure mimicking api_wordlist repository
	tmpdir, err := ioutil.TempDir("", "api_wordlist_repo")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Create subdirectories
	authDir := filepath.Join(tmpdir, "auth")
	dataDir := filepath.Join(tmpdir, "data")
	if err := os.Mkdir(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth directory: %v", err)
	}
	if err := os.Mkdir(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	// Create wordlist files
	authContent := `/api/auth/login
/api/auth/logout
/api/auth/register
/api/auth/reset-password
`
	dataContent := `/api/data/query
/api/data/export
/api/data/import
/v1/data/stream
`

	if err := ioutil.WriteFile(filepath.Join(authDir, "endpoints.txt"), []byte(authContent), 0644); err != nil {
		t.Fatalf("Failed to write auth wordlist: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(dataDir, "endpoints.txt"), []byte(dataContent), 0644); err != nil {
		t.Fatalf("Failed to write data wordlist: %v", err)
	}

	// Create a config for testing
	config := &ffuf.Config{}

	// Create a new APIWordlist
	wl, err := NewAPIWordlist(tmpdir, config)
	if err != nil {
		t.Fatalf("Failed to create APIWordlist: %v", err)
	}

	// Test total entries
	if wl.Total() != 8 {
		t.Errorf("Expected 8 entries, got %d", wl.Total())
	}

	// Test categories
	categories := wl.GetCategories()
	if len(categories) < 2 {
		t.Errorf("Expected at least 2 categories, got %d", len(categories))
	}
	hasAuth := false
	hasData := false
	for _, cat := range categories {
		if cat == "auth" {
			hasAuth = true
		}
		if cat == "data" {
			hasData = true
		}
	}
	if !hasAuth || !hasData {
		t.Errorf("Missing expected categories. Has auth: %v, Has data: %v", hasAuth, hasData)
	}

	// Test entries by category
	authEntries := wl.GetEntriesByCategory("auth")
	if len(authEntries) != 4 {
		t.Errorf("Expected 4 auth entries, got %d: %v", len(authEntries), authEntries)
	}

	dataEntries := wl.GetEntriesByCategory("data")
	if len(dataEntries) != 4 {
		t.Errorf("Expected 4 data entries, got %d: %v", len(dataEntries), dataEntries)
	}
}

func TestAPIWordlistFromAPIWordlistRepo(t *testing.T) {
	// Create a temporary directory mimicking the api_wordlist repository format
	tmpdir, err := ioutil.TempDir("", "api_wordlist_specific_repo")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Create actions.txt
	actionsContent := `get
list
create
update
delete
search
find
add
remove
`
	if err := ioutil.WriteFile(filepath.Join(tmpdir, "actions.txt"), []byte(actionsContent), 0644); err != nil {
		t.Fatalf("Failed to write actions.txt: %v", err)
	}

	// Create objects.txt
	objectsContent := `user
account
product
order
payment
setting
config
profile
document
`
	if err := ioutil.WriteFile(filepath.Join(tmpdir, "objects.txt"), []byte(objectsContent), 0644); err != nil {
		t.Fatalf("Failed to write objects.txt: %v", err)
	}

	// Create api_seen_in_wild.txt
	seenInWildContent := `api/v1/getUserProfile
api/v2/listOrders
api/auth/login
api/products/search
`
	if err := ioutil.WriteFile(filepath.Join(tmpdir, "api_seen_in_wild.txt"), []byte(seenInWildContent), 0644); err != nil {
		t.Fatalf("Failed to write api_seen_in_wild.txt: %v", err)
	}

	// Create a config for testing
	config := &ffuf.Config{}

	// Create a new APIWordlist
	wl, err := NewAPIWordlist(tmpdir, config)
	if err != nil {
		t.Fatalf("Failed to create APIWordlist: %v", err)
	}

	// Test that it was recognized as an api_wordlist repository
	if !wl.IsAPIWordlistRepo() {
		t.Errorf("Expected wordlist to be recognized as an api_wordlist repository")
	}

	// Test actions were loaded
	actions := wl.GetActions()
	if len(actions) != 9 {
		t.Errorf("Expected 9 actions, got %d: %v", len(actions), actions)
	}

	// Test objects were loaded
	objects := wl.GetObjects()
	if len(objects) != 9 {
		t.Errorf("Expected 9 objects, got %d: %v", len(objects), objects)
	}

	// Test seen in wild entries were loaded
	seenInWild := wl.GetSeenInWild()
	if len(seenInWild) != 4 {
		t.Errorf("Expected 4 seen in wild entries, got %d: %v", len(seenInWild), seenInWild)
	}

	// Test total entries (should include generated combinations)
	// 9 actions * 9 objects * 6 patterns + 4 seen in wild = 490
	expectedMinEntries := 400 // Use a lower bound to be safe
	if wl.Total() < expectedMinEntries {
		t.Errorf("Expected at least %d entries, got %d", expectedMinEntries, wl.Total())
	}

	// Test entries by action category
	getUserEntries := wl.GetEntriesByCategory("action:get")
	if len(getUserEntries) < 9 {
		t.Errorf("Expected at least 9 get entries, got %d: %v", len(getUserEntries), getUserEntries)
	}

	// Test entries by object category
	userEntries := wl.GetEntriesByCategory("object:user")
	if len(userEntries) < 9 {
		t.Errorf("Expected at least 9 user entries, got %d: %v", len(userEntries), userEntries)
	}

	// Test entries by prefix
	apiV1Entries := wl.GetEntriesByPrefix("/api/v1")
	if len(apiV1Entries) < 10 {
		t.Errorf("Expected at least 10 /api/v1 entries, got %d: %v", len(apiV1Entries), apiV1Entries)
	}

	// Test entries by pattern
	getUserEntries = wl.GetEntriesByPattern("action:get")
	if len(getUserEntries) < 9 {
		t.Errorf("Expected at least 9 get pattern entries, got %d: %v", len(getUserEntries), getUserEntries)
	}

	// Test seen in wild entries were added as regular endpoints
	seenInWildEntries := wl.GetEntriesByCategory("seen_in_wild")
	if len(seenInWildEntries) != 4 {
		t.Errorf("Expected 4 seen_in_wild category entries, got %d: %v", len(seenInWildEntries), seenInWildEntries)
	}
}

// These methods are now implemented in the main wordlist.go file

// Helper method to reset position (not in the original implementation)
func (w *APIWordlist) ResetPosition() {
	w.position = 0
}
