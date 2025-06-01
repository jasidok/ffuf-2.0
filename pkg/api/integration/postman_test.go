package integration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestImportPostmanCollection(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "postman_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample Postman collection file
	collectionPath := filepath.Join(tempDir, "test_collection.json")
	collection := createSampleCollection()
	collectionData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal sample collection: %v", err)
	}
	err = ioutil.WriteFile(collectionPath, collectionData, 0644)
	if err != nil {
		t.Fatalf("Failed to write sample collection file: %v", err)
	}

	// Test importing the collection
	requests, err := ImportPostmanCollection(collectionPath)
	if err != nil {
		t.Fatalf("ImportPostmanCollection failed: %v", err)
	}

	// Verify the imported requests
	if len(requests) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(requests))
	}

	// Check the first request
	if requests[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", requests[0].Method)
	}
	if requests[0].Url != "https://api.example.com/users" {
		t.Errorf("Expected URL https://api.example.com/users, got %s", requests[0].Url)
	}
	if val, ok := requests[0].Headers["Accept"]; !ok || val != "application/json" {
		t.Errorf("Expected Accept header application/json, got %s", val)
	}

	// Check the second request
	if requests[1].Method != "POST" {
		t.Errorf("Expected method POST, got %s", requests[1].Method)
	}
	if requests[1].Url != "https://api.example.com/users" {
		t.Errorf("Expected URL https://api.example.com/users, got %s", requests[1].Url)
	}
	if val, ok := requests[1].Headers["Content-Type"]; !ok || val != "application/json" {
		t.Errorf("Expected Content-Type header application/json, got %s", val)
	}
	if string(requests[1].Data) != `{"name":"Test User","email":"test@example.com"}` {
		t.Errorf("Expected request body {\"name\":\"Test User\",\"email\":\"test@example.com\"}, got %s", string(requests[1].Data))
	}
}

func TestExportToPostmanCollection(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "postman_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample ffuf requests
	requests := []ffuf.Request{
		{
			Method: "GET",
			Url:    "https://api.example.com/users",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		},
		{
			Method: "POST",
			Url:    "https://api.example.com/users",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Data: []byte(`{"name":"Test User","email":"test@example.com"}`),
		},
	}

	// Test exporting the requests
	outputPath := filepath.Join(tempDir, "exported_collection.json")
	err = ExportToPostmanCollection(requests, "Test Collection", outputPath)
	if err != nil {
		t.Fatalf("ExportToPostmanCollection failed: %v", err)
	}

	// Verify the exported collection
	data, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported collection: %v", err)
	}

	var collection PostmanCollection
	err = json.Unmarshal(data, &collection)
	if err != nil {
		t.Fatalf("Failed to parse exported collection: %v", err)
	}

	// Check collection metadata
	if collection.Info.Name != "Test Collection" {
		t.Errorf("Expected collection name Test Collection, got %s", collection.Info.Name)
	}
	if collection.Info.Schema != "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" {
		t.Errorf("Expected schema URL, got %s", collection.Info.Schema)
	}

	// Check items
	if len(collection.Item) != 2 {
		t.Errorf("Expected 2 items, got %d", len(collection.Item))
	}

	// Check the first item
	if collection.Item[0].Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", collection.Item[0].Request.Method)
	}
	if collection.Item[0].Request.URL.Raw != "https://api.example.com/users" {
		t.Errorf("Expected URL https://api.example.com/users, got %s", collection.Item[0].Request.URL.Raw)
	}
	if len(collection.Item[0].Request.Header) != 1 || collection.Item[0].Request.Header[0].Key != "Accept" || collection.Item[0].Request.Header[0].Value != "application/json" {
		t.Errorf("Expected Accept header application/json, got %v", collection.Item[0].Request.Header)
	}

	// Check the second item
	if collection.Item[1].Request.Method != "POST" {
		t.Errorf("Expected method POST, got %s", collection.Item[1].Request.Method)
	}
	if collection.Item[1].Request.URL.Raw != "https://api.example.com/users" {
		t.Errorf("Expected URL https://api.example.com/users, got %s", collection.Item[1].Request.URL.Raw)
	}
	if len(collection.Item[1].Request.Header) != 1 || collection.Item[1].Request.Header[0].Key != "Content-Type" || collection.Item[1].Request.Header[0].Value != "application/json" {
		t.Errorf("Expected Content-Type header application/json, got %v", collection.Item[1].Request.Header)
	}
	if collection.Item[1].Request.Body == nil || collection.Item[1].Request.Body.Mode != "raw" || collection.Item[1].Request.Body.Raw != `{"name":"Test User","email":"test@example.com"}` {
		t.Errorf("Expected request body, got %v", collection.Item[1].Request.Body)
	}
}

// Helper function to create a sample Postman collection for testing
func createSampleCollection() PostmanCollection {
	return PostmanCollection{
		Info: PostmanInfo{
			Name:   "Test Collection",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: []PostmanItem{
			{
				Name: "Get Users",
				Request: &PostmanRequest{
					Method: "GET",
					URL: PostmanURL{
						Raw: "https://api.example.com/users",
					},
					Header: []PostmanHeader{
						{
							Key:   "Accept",
							Value: "application/json",
						},
					},
				},
			},
			{
				Name: "Create User",
				Request: &PostmanRequest{
					Method: "POST",
					URL: PostmanURL{
						Raw: "https://api.example.com/users",
					},
					Header: []PostmanHeader{
						{
							Key:   "Content-Type",
							Value: "application/json",
						},
					},
					Body: &PostmanBody{
						Mode: "raw",
						Raw:  `{"name":"Test User","email":"test@example.com"}`,
					},
				},
			},
		},
	}
}