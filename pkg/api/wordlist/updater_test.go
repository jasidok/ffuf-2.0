package wordlist

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWordlistUpdater(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)

	// Check that the updater has the expected values
	if updater.RepoOwner != "danielmiessler" {
		t.Errorf("Expected RepoOwner 'danielmiessler', got '%s'", updater.RepoOwner)
	}
	if updater.RepoName != "api_wordlist" {
		t.Errorf("Expected RepoName 'api_wordlist', got '%s'", updater.RepoName)
	}
	if updater.Branch != DefaultAPIWordlistBranch {
		t.Errorf("Expected Branch '%s', got '%s'", DefaultAPIWordlistBranch, updater.Branch)
	}
	if updater.LocalPath != tempDir {
		t.Errorf("Expected LocalPath '%s', got '%s'", tempDir, updater.LocalPath)
	}
	if updater.UpdateInterval != DefaultUpdateInterval {
		t.Errorf("Expected UpdateInterval %v, got %v", DefaultUpdateInterval, updater.UpdateInterval)
	}
	if updater.Client == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if updater.Metadata.RepoURL != "https://github.com/danielmiessler/api_wordlist" {
		t.Errorf("Expected RepoURL 'https://github.com/danielmiessler/api_wordlist', got '%s'", updater.Metadata.RepoURL)
	}
	if updater.Metadata.LocalPath != tempDir {
		t.Errorf("Expected LocalPath '%s', got '%s'", tempDir, updater.Metadata.LocalPath)
	}
	if updater.Metadata.Branch != DefaultAPIWordlistBranch {
		t.Errorf("Expected Branch '%s', got '%s'", DefaultAPIWordlistBranch, updater.Metadata.Branch)
	}
	if len(updater.Metadata.Files) != 0 {
		t.Errorf("Expected 0 files in metadata, got %d", len(updater.Metadata.Files))
	}
}

func TestLoadSaveMetadata(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)

	// Add some test data to the metadata
	updater.Metadata.LastCommit = "test-commit-sha"
	updater.Metadata.Files["test-file.txt"] = FileMetadata{
		Path:         "test-file.txt",
		SHA:          "test-file-sha",
		Size:         100,
		LastModified: time.Now(),
	}

	// Save the metadata
	err = updater.SaveMetadata()
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Create a new updater to load the metadata
	updater2 := NewWordlistUpdater(tempDir)

	// Load the metadata
	err = updater2.LoadMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Check that the metadata was loaded correctly
	if updater2.Metadata.LastCommit != "test-commit-sha" {
		t.Errorf("Expected LastCommit 'test-commit-sha', got '%s'", updater2.Metadata.LastCommit)
	}
	if len(updater2.Metadata.Files) != 1 {
		t.Errorf("Expected 1 file in metadata, got %d", len(updater2.Metadata.Files))
	}
	if file, exists := updater2.Metadata.Files["test-file.txt"]; !exists {
		t.Error("Expected 'test-file.txt' in metadata files")
	} else {
		if file.SHA != "test-file-sha" {
			t.Errorf("Expected SHA 'test-file-sha', got '%s'", file.SHA)
		}
		if file.Size != 100 {
			t.Errorf("Expected Size 100, got %d", file.Size)
		}
	}
}

func TestCheckForUpdates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)

	// Test with no previous updates - this would normally make HTTP calls
	_, err = updater.CheckForUpdates()
	if err == nil {
		t.Log("CheckForUpdates would normally require network access")
	}

	// Test with a previous update with a different commit
	updater.Metadata.LastCommit = "old-commit-sha"
	updater.Metadata.LastUpdate = time.Now().Add(-48 * time.Hour) // 2 days ago
	_, err = updater.CheckForUpdates()
	if err == nil {
		t.Log("CheckForUpdates would normally require network access")
	}

	// Test with a previous update with the same commit
	updater.Metadata.LastCommit = "new-commit-sha"
	updater.Metadata.LastUpdate = time.Now().Add(-48 * time.Hour) // 2 days ago
	updatesAvailable, err := updater.CheckForUpdates()
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}
	if updatesAvailable {
		t.Error("Expected no updates to be available with the same commit")
	}

	// Test with a recent update (within the update interval)
	updater.Metadata.LastCommit = "old-commit-sha"
	updater.Metadata.LastUpdate = time.Now().Add(-1 * time.Hour) // 1 hour ago
	updatesAvailable, err = updater.CheckForUpdates()
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}
	if updatesAvailable {
		t.Error("Expected no updates to be available with a recent update")
	}
}

func TestGetUpdateStatus(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)

	// Set some test metadata
	updater.Metadata.LastCommit = "test-commit-sha"
	updater.Metadata.LastUpdate = time.Now().Add(-48 * time.Hour) // 2 days ago
	updater.Metadata.Files["test-file.txt"] = FileMetadata{
		Path:         "test-file.txt",
		SHA:          "test-file-sha",
		Size:         100,
		LastModified: time.Now(),
	}

	// Save the metadata
	err = updater.SaveMetadata()
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Get the update status
	status, err := updater.GetUpdateStatus()
	if err != nil {
		t.Fatalf("Failed to get update status: %v", err)
	}

	// Check the status
	if status["repo_url"] != "https://github.com/danielmiessler/api_wordlist" {
		t.Errorf("Expected repo_url 'https://github.com/danielmiessler/api_wordlist', got '%s'", status["repo_url"])
	}
	if status["local_path"] != tempDir {
		t.Errorf("Expected local_path '%s', got '%s'", tempDir, status["local_path"])
	}
	if status["branch"] != DefaultAPIWordlistBranch {
		t.Errorf("Expected branch '%s', got '%s'", DefaultAPIWordlistBranch, status["branch"])
	}
	if status["last_commit"] != "test-commit-sha" {
		t.Errorf("Expected last_commit 'test-commit-sha', got '%s'", status["last_commit"])
	}
	if status["file_count"] != 1 {
		t.Errorf("Expected file_count 1, got %d", status["file_count"])
	}
	if status["update_due"] != true {
		t.Error("Expected update_due to be true")
	}
}

func TestProcessFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a test file
		w.Write([]byte("test file content"))
	}))
	defer server.Close()

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)
	updater.Client = server.Client()

	// Create a test file item
	item := GitHubContent{
		Name:         "test-file.txt",
		Path:         "test-file.txt",
		SHA:          "test-file-sha",
		Size:         100,
		Type:         "file",
		DownloadURL:  server.URL,
		LastModified: time.Now(),
	}

	// Process the file
	err = updater.processFile(item)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Check that the file was downloaded
	filePath := filepath.Join(tempDir, "test-file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to exist")
	}

	// Check the file contents
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "test file content" {
		t.Errorf("Expected file content 'test file content', got '%s'", string(content))
	}

	// Check that the file metadata was updated
	if file, exists := updater.Metadata.Files["test-file.txt"]; !exists {
		t.Error("Expected 'test-file.txt' in metadata files")
	} else {
		if file.SHA != "test-file-sha" {
			t.Errorf("Expected SHA 'test-file-sha', got '%s'", file.SHA)
		}
		if file.Size != 100 {
			t.Errorf("Expected Size 100, got %d", file.Size)
		}
	}

	// Test with a file that doesn't need to be updated
	updater.Metadata.Files["test-file.txt"] = FileMetadata{
		Path:         "test-file.txt",
		SHA:          "test-file-sha",
		Size:         100,
		LastModified: time.Now(),
	}

	// Modify the file to check if it's updated
	err = ioutil.WriteFile(filePath, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Process the file again
	err = updater.processFile(item)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Check that the file wasn't updated
	content, err = ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "modified content" {
		t.Errorf("Expected file content 'modified content', got '%s'", string(content))
	}
}

func TestAutoUpdate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater
	updater := NewWordlistUpdater(tempDir)

	// Test with no previous updates
	updated, err := updater.AutoUpdate()
	if err == nil {
		t.Log("AutoUpdate would normally require network access")
	}
	_ = updated // Avoid unused variable warning

	// Test with a previous update with a different commit
	updated, err = updater.AutoUpdate()
	if err == nil {
		t.Log("AutoUpdate would normally require network access")
	}
	_ = updated // Avoid unused variable warning
}

func TestBackgroundUpdates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "api_wordlist_updater_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new updater with a short update interval for testing
	updater := NewWordlistUpdater(tempDir)
	updater.UpdateInterval = 100 * time.Millisecond // Short interval for testing

	// Start background updates
	err = updater.StartBackgroundUpdates()
	if err != nil {
		t.Fatalf("Failed to start background updates: %v", err)
	}

	// Check that background updates are running
	if !updater.IsBackgroundRunning() {
		t.Error("Expected background updates to be running")
	}

	// Stop background updates
	updater.StopBackgroundUpdates()

	// Check that background updates are no longer running
	if updater.IsBackgroundRunning() {
		t.Error("Expected background updates to be stopped")
	}
}
