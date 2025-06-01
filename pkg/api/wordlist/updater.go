// Package wordlist provides specialized functionality for handling API endpoint wordlists.
//
// This file implements an automated update mechanism to keep api_wordlist data current.
package wordlist

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// Repository information
const (
	// DefaultAPIWordlistRepo is the default GitHub repository for api_wordlist
	DefaultAPIWordlistRepo = "danielmiessler/api_wordlist"
	// DefaultAPIWordlistBranch is the default branch to use
	DefaultAPIWordlistBranch = "main"
	// DefaultUpdateInterval is the default interval between update checks (24 hours)
	DefaultUpdateInterval = 24 * time.Hour
	// UpdateMetadataFile is the file where update metadata is stored
	UpdateMetadataFile = ".api_wordlist_metadata.json"
)

// UpdateMetadata contains metadata about the api_wordlist repository
type UpdateMetadata struct {
	// LastUpdate is the timestamp of the last update
	LastUpdate time.Time `json:"last_update"`
	// LastCommit is the SHA of the last commit that was pulled
	LastCommit string `json:"last_commit"`
	// RepoURL is the URL of the repository
	RepoURL string `json:"repo_url"`
	// LocalPath is the local path where the repository is stored
	LocalPath string `json:"local_path"`
	// Branch is the branch that was pulled
	Branch string `json:"branch"`
	// Files is a map of filenames to file metadata
	Files map[string]FileMetadata `json:"files"`
}

// FileMetadata contains metadata about a file in the api_wordlist repository
type FileMetadata struct {
	// Path is the path of the file in the repository
	Path string `json:"path"`
	// SHA is the SHA of the file
	SHA string `json:"sha"`
	// Size is the size of the file in bytes
	Size int `json:"size"`
	// LastModified is the timestamp when the file was last modified
	LastModified time.Time `json:"last_modified"`
}

// GitHubCommit represents a commit in a GitHub repository
type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

// GitHubContent represents a file in a GitHub repository
type GitHubContent struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	SHA          string    `json:"sha"`
	Size         int       `json:"size"`
	Type         string    `json:"type"`
	DownloadURL  string    `json:"download_url"`
	LastModified time.Time `json:"-"`
}

// WordlistUpdater manages updates for the api_wordlist repository
type WordlistUpdater struct {
	// RepoOwner is the owner of the GitHub repository
	RepoOwner string
	// RepoName is the name of the GitHub repository
	RepoName string
	// Branch is the branch to use
	Branch string
	// LocalPath is the local path where the repository is stored
	LocalPath string
	// UpdateInterval is the interval between update checks
	UpdateInterval time.Duration
	// Metadata is the update metadata
	Metadata UpdateMetadata
	// Client is the HTTP client to use for API requests
	Client *http.Client

	// Background update fields
	backgroundRunning bool
	stopBackground    chan struct{}
	backgroundWg      sync.WaitGroup
	mutex             sync.Mutex
	logger            *log.Logger

	// Function fields for testing
	checkForUpdatesFn func() (bool, error)
	updateFn          func() error
	autoUpdateFn      func() (bool, error)
}

// NewWordlistUpdater creates a new WordlistUpdater with the given parameters
func NewWordlistUpdater(localPath string) *WordlistUpdater {
	// Parse the default repository
	parts := strings.Split(DefaultAPIWordlistRepo, "/")
	owner := parts[0]
	name := parts[1]

	updater := &WordlistUpdater{
		RepoOwner:      owner,
		RepoName:       name,
		Branch:         DefaultAPIWordlistBranch,
		LocalPath:      localPath,
		UpdateInterval: DefaultUpdateInterval,
		Client:         &http.Client{Timeout: 30 * time.Second},
		Metadata: UpdateMetadata{
			RepoURL:   fmt.Sprintf("https://github.com/%s/%s", owner, name),
			LocalPath: localPath,
			Branch:    DefaultAPIWordlistBranch,
			Files:     make(map[string]FileMetadata),
		},
		backgroundRunning: false,
		stopBackground:    make(chan struct{}),
		logger:            log.New(os.Stderr, "[api_wordlist] ", log.LstdFlags),
	}

	// Initialize function fields with default implementations
	updater.checkForUpdatesFn = updater.CheckForUpdates
	updater.updateFn = updater.Update
	updater.autoUpdateFn = updater.AutoUpdate

	return updater
}

// LoadMetadata loads update metadata from the metadata file
func (u *WordlistUpdater) LoadMetadata() error {
	metadataPath := filepath.Join(u.LocalPath, UpdateMetadataFile)

	// Check if the metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Metadata file doesn't exist, use default metadata
		return nil
	}

	// Read the metadata file
	data, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return api.NewAPIError("Failed to read metadata file: "+err.Error(), 0)
	}

	// Parse the metadata
	err = json.Unmarshal(data, &u.Metadata)
	if err != nil {
		return api.NewAPIError("Failed to parse metadata: "+err.Error(), 0)
	}

	return nil
}

// SaveMetadata saves update metadata to the metadata file
func (u *WordlistUpdater) SaveMetadata() error {
	metadataPath := filepath.Join(u.LocalPath, UpdateMetadataFile)

	// Update the last update timestamp
	u.Metadata.LastUpdate = time.Now()

	// Marshal the metadata to JSON
	data, err := json.MarshalIndent(u.Metadata, "", "  ")
	if err != nil {
		return api.NewAPIError("Failed to marshal metadata: "+err.Error(), 0)
	}

	// Write the metadata file
	err = ioutil.WriteFile(metadataPath, data, 0644)
	if err != nil {
		return api.NewAPIError("Failed to write metadata file: "+err.Error(), 0)
	}

	return nil
}

// CheckForUpdates checks if there are updates available for the api_wordlist repository
func (u *WordlistUpdater) CheckForUpdates() (bool, error) {
	// Load metadata
	err := u.LoadMetadata()
	if err != nil {
		return false, err
	}

	// Check if the update interval has passed
	if time.Since(u.Metadata.LastUpdate) < u.UpdateInterval {
		// Not time to check for updates yet
		return false, nil
	}

	// Get the latest commit
	latestCommit, err := u.getLatestCommit()
	if err != nil {
		return false, err
	}

	// Check if the latest commit is different from the last commit
	if latestCommit.SHA == u.Metadata.LastCommit {
		// No updates available
		return false, nil
	}

	// Updates are available
	return true, nil
}

// Update updates the api_wordlist repository
func (u *WordlistUpdater) Update() error {
	// Load metadata
	err := u.LoadMetadata()
	if err != nil {
		return err
	}

	// Get the latest commit
	latestCommit, err := u.getLatestCommit()
	if err != nil {
		return err
	}

	// Get the repository contents
	contents, err := u.getRepositoryContents("")
	if err != nil {
		return err
	}

	// Create the local directory if it doesn't exist
	err = os.MkdirAll(u.LocalPath, 0755)
	if err != nil {
		return api.NewAPIError("Failed to create local directory: "+err.Error(), 0)
	}

	// Process each item in the repository
	for _, item := range contents {
		if item.Type == "dir" {
			// Process directory
			err = u.processDirectory(item.Path)
			if err != nil {
				return err
			}
		} else {
			// Process file
			err = u.processFile(item)
			if err != nil {
				return err
			}
		}
	}

	// Update metadata
	u.Metadata.LastCommit = latestCommit.SHA
	u.Metadata.LastUpdate = time.Now()

	// Save metadata
	err = u.SaveMetadata()
	if err != nil {
		return err
	}

	return nil
}

// processDirectory processes a directory in the repository
func (u *WordlistUpdater) processDirectory(path string) error {
	// Get the directory contents
	contents, err := u.getRepositoryContents(path)
	if err != nil {
		return err
	}

	// Create the local directory if it doesn't exist
	localDir := filepath.Join(u.LocalPath, path)
	err = os.MkdirAll(localDir, 0755)
	if err != nil {
		return api.NewAPIError("Failed to create local directory: "+err.Error(), 0)
	}

	// Process each item in the directory
	for _, item := range contents {
		if item.Type == "dir" {
			// Process subdirectory
			err = u.processDirectory(item.Path)
			if err != nil {
				return err
			}
		} else {
			// Process file
			err = u.processFile(item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processFile processes a file in the repository
func (u *WordlistUpdater) processFile(item GitHubContent) error {
	// Check if the file needs to be updated
	needsUpdate := true
	if metadata, exists := u.Metadata.Files[item.Path]; exists {
		if metadata.SHA == item.SHA {
			// File hasn't changed
			needsUpdate = false
		}
	}

	if needsUpdate {
		// Download the file
		err := u.downloadFile(item)
		if err != nil {
			return err
		}

		// Update file metadata
		u.Metadata.Files[item.Path] = FileMetadata{
			Path:         item.Path,
			SHA:          item.SHA,
			Size:         item.Size,
			LastModified: item.LastModified,
		}
	}

	return nil
}

// downloadFile downloads a file from the repository
func (u *WordlistUpdater) downloadFile(item GitHubContent) error {
	// Create the local directory if it doesn't exist
	localDir := filepath.Dir(filepath.Join(u.LocalPath, item.Path))
	err := os.MkdirAll(localDir, 0755)
	if err != nil {
		return api.NewAPIError("Failed to create local directory: "+err.Error(), 0)
	}

	// Download the file
	resp, err := u.Client.Get(item.DownloadURL)
	if err != nil {
		return api.NewAPIError("Failed to download file: "+err.Error(), 0)
	}
	defer resp.Body.Close()

	// Create the local file
	localPath := filepath.Join(u.LocalPath, item.Path)
	file, err := os.Create(localPath)
	if err != nil {
		return api.NewAPIError("Failed to create local file: "+err.Error(), 0)
	}
	defer file.Close()

	// Copy the file contents
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return api.NewAPIError("Failed to write file contents: "+err.Error(), 0)
	}

	return nil
}

// getLatestCommit gets the latest commit in the repository
func (u *WordlistUpdater) getLatestCommit() (GitHubCommit, error) {
	// Build the API URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", u.RepoOwner, u.RepoName, u.Branch)

	// Make the API request
	resp, err := u.Client.Get(url)
	if err != nil {
		return GitHubCommit{}, api.NewAPIError("Failed to get latest commit: "+err.Error(), 0)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return GitHubCommit{}, api.NewAPIError(fmt.Sprintf("Failed to get latest commit: %s", resp.Status), 0)
	}

	// Parse the response
	var commit GitHubCommit
	err = json.NewDecoder(resp.Body).Decode(&commit)
	if err != nil {
		return GitHubCommit{}, api.NewAPIError("Failed to parse commit data: "+err.Error(), 0)
	}

	return commit, nil
}

// getRepositoryContents gets the contents of a directory in the repository
func (u *WordlistUpdater) getRepositoryContents(path string) ([]GitHubContent, error) {
	// Build the API URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", u.RepoOwner, u.RepoName, path)
	if path == "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", u.RepoOwner, u.RepoName)
	}
	if u.Branch != "" {
		url += fmt.Sprintf("?ref=%s", u.Branch)
	}

	// Make the API request
	resp, err := u.Client.Get(url)
	if err != nil {
		return nil, api.NewAPIError("Failed to get repository contents: "+err.Error(), 0)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, api.NewAPIError(fmt.Sprintf("Failed to get repository contents: %s", resp.Status), 0)
	}

	// Parse the response
	var contents []GitHubContent
	err = json.NewDecoder(resp.Body).Decode(&contents)
	if err != nil {
		return nil, api.NewAPIError("Failed to parse contents data: "+err.Error(), 0)
	}

	// Set the last modified time for each item
	for i := range contents {
		contents[i].LastModified = time.Now()
	}

	return contents, nil
}

// AutoUpdate checks for updates and updates the repository if needed
func (u *WordlistUpdater) AutoUpdate() (bool, error) {
	// Check for updates
	updatesAvailable, err := u.CheckForUpdates()
	if err != nil {
		return false, err
	}

	// If updates are available, update the repository
	if updatesAvailable {
		err = u.Update()
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// GetUpdateStatus returns the status of the api_wordlist repository
func (u *WordlistUpdater) GetUpdateStatus() (map[string]interface{}, error) {
	// Load metadata
	err := u.LoadMetadata()
	if err != nil {
		return nil, err
	}

	// Build the status
	status := map[string]interface{}{
		"repo_url":           u.Metadata.RepoURL,
		"local_path":         u.Metadata.LocalPath,
		"branch":             u.Metadata.Branch,
		"last_update":        u.Metadata.LastUpdate,
		"last_commit":        u.Metadata.LastCommit,
		"file_count":         len(u.Metadata.Files),
		"next_update":        u.Metadata.LastUpdate.Add(u.UpdateInterval),
		"update_due":         time.Since(u.Metadata.LastUpdate) >= u.UpdateInterval,
		"background_running": u.IsBackgroundRunning(),
	}

	return status, nil
}

// StartBackgroundUpdates starts a background goroutine that periodically checks for updates
// and applies them if available. The interval between checks is determined by UpdateInterval.
// If background updates are already running, this method does nothing.
func (u *WordlistUpdater) StartBackgroundUpdates() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Check if background updates are already running
	if u.backgroundRunning {
		return nil
	}

	// Create a new stop channel
	u.stopBackground = make(chan struct{})
	u.backgroundRunning = true

	// Start the background goroutine
	u.backgroundWg.Add(1)
	go func() {
		defer u.backgroundWg.Done()
		u.logger.Printf("Starting background updates for api_wordlist repository")

		ticker := time.NewTicker(u.UpdateInterval)
		defer ticker.Stop()

		// Perform an initial update
		updated, err := u.AutoUpdate()
		if err != nil {
			u.logger.Printf("Error during initial api_wordlist update: %v", err)
		} else if updated {
			u.logger.Printf("Initial api_wordlist update completed successfully")
		} else {
			u.logger.Printf("No updates available during initial check")
		}

		// Main update loop
		for {
			select {
			case <-ticker.C:
				// Time to check for updates
				updated, err := u.AutoUpdate()
				if err != nil {
					u.logger.Printf("Error during api_wordlist update: %v", err)
				} else if updated {
					u.logger.Printf("api_wordlist update completed successfully")
				}
			case <-u.stopBackground:
				// Stop signal received
				u.logger.Printf("Stopping background updates for api_wordlist repository")
				return
			}
		}
	}()

	return nil
}

// StopBackgroundUpdates stops the background update goroutine if it's running.
// This method waits for the goroutine to finish before returning.
func (u *WordlistUpdater) StopBackgroundUpdates() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Check if background updates are running
	if !u.backgroundRunning {
		return
	}

	// Signal the background goroutine to stop
	select {
	case <-u.stopBackground:
		// Channel already closed
	default:
		close(u.stopBackground)
	}
	u.backgroundWg.Wait()

	// Reset the state
	u.backgroundRunning = false
}

// IsBackgroundRunning returns true if background updates are currently running.
func (u *WordlistUpdater) IsBackgroundRunning() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	return u.backgroundRunning
}
