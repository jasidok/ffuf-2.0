// Package integration provides functionality for integrating ffuf with external API tools and formats.
package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// PostmanCollection represents the structure of a Postman collection
type PostmanCollection struct {
	Info  PostmanInfo       `json:"info"`
	Item  []PostmanItem     `json:"item"`
	Auth  *PostmanAuth      `json:"auth,omitempty"`
	Event []PostmanEvent    `json:"event,omitempty"`
	Variable []PostmanVariable `json:"variable,omitempty"`
}

// PostmanInfo contains metadata about the collection
type PostmanInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
	Version     string `json:"version,omitempty"`
}

// PostmanItem represents a request or a folder in a Postman collection
type PostmanItem struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Item        []PostmanItem `json:"item,omitempty"`
	Request     *PostmanRequest `json:"request,omitempty"`
	Response    []PostmanResponse `json:"response,omitempty"`
}

// PostmanRequest represents a request in a Postman collection
type PostmanRequest struct {
	Method      string           `json:"method"`
	URL         PostmanURL       `json:"url"`
	Description string           `json:"description,omitempty"`
	Header      []PostmanHeader  `json:"header,omitempty"`
	Body        *PostmanBody     `json:"body,omitempty"`
	Auth        *PostmanAuth     `json:"auth,omitempty"`
}

// PostmanURL represents a URL in a Postman request
type PostmanURL struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol,omitempty"`
	Host     []string `json:"host,omitempty"`
	Path     []string `json:"path,omitempty"`
	Query    []PostmanQuery `json:"query,omitempty"`
}

// PostmanQuery represents a query parameter in a Postman URL
type PostmanQuery struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanHeader represents a header in a Postman request
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody represents the body of a Postman request
type PostmanBody struct {
	Mode    string `json:"mode"`
	Raw     string `json:"raw,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// PostmanResponse represents a response in a Postman collection
type PostmanResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Code   int    `json:"code"`
	Header []PostmanHeader `json:"header,omitempty"`
	Body   string `json:"body,omitempty"`
}

// PostmanAuth represents authentication details in a Postman collection
type PostmanAuth struct {
	Type   string `json:"type"`
	Bearer []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"bearer,omitempty"`
	Basic []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"basic,omitempty"`
	// Other auth types can be added as needed
}

// PostmanVariable represents a variable in a Postman collection
type PostmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanEvent represents an event in a Postman collection
type PostmanEvent struct {
	Listen string `json:"listen"`
	Script struct {
		Type string   `json:"type"`
		Exec []string `json:"exec"`
	} `json:"script"`
}

// ImportPostmanCollection imports a Postman collection from a file and converts it to ffuf requests
func ImportPostmanCollection(filePath string) ([]ffuf.Request, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Postman collection file: %v", err)
	}

	// Parse the JSON
	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse Postman collection: %v", err)
	}

	// Convert to ffuf requests
	var requests []ffuf.Request
	err = processItems(collection.Item, &requests, collection.Auth)
	if err != nil {
		return nil, err
	}

	return requests, nil
}

// processItems recursively processes Postman items and converts them to ffuf requests
func processItems(items []PostmanItem, requests *[]ffuf.Request, collectionAuth *PostmanAuth) error {
	for _, item := range items {
		// If this is a folder, process its items recursively
		if len(item.Item) > 0 {
			if err := processItems(item.Item, requests, collectionAuth); err != nil {
				return err
			}
			continue
		}

		// Skip items without requests
		if item.Request == nil {
			continue
		}

		// Convert Postman request to ffuf request
		req, err := convertPostmanRequestToFfuf(item.Request, collectionAuth)
		if err != nil {
			return err
		}

		*requests = append(*requests, req)
	}

	return nil
}

// convertPostmanRequestToFfuf converts a Postman request to an ffuf request
func convertPostmanRequestToFfuf(postmanReq *PostmanRequest, collectionAuth *PostmanAuth) (ffuf.Request, error) {
	req := ffuf.Request{
		Method:  postmanReq.Method,
		Url:     postmanReq.URL.Raw,
		Headers: make(map[string]string),
	}

	// Add headers
	for _, header := range postmanReq.Header {
		req.Headers[header.Key] = header.Value
	}

	// Add authentication if present
	auth := postmanReq.Auth
	if auth == nil {
		auth = collectionAuth
	}
	if auth != nil {
		if err := addAuthToRequest(auth, &req); err != nil {
			return req, err
		}
	}

	// Add body if present
	if postmanReq.Body != nil && postmanReq.Body.Mode == "raw" {
		req.Data = []byte(postmanReq.Body.Raw)
	}

	return req, nil
}

// addAuthToRequest adds authentication details to an ffuf request
func addAuthToRequest(auth *PostmanAuth, req *ffuf.Request) error {
	switch auth.Type {
	case "bearer":
		for _, bearer := range auth.Bearer {
			if bearer.Key == "token" {
				req.Headers["Authorization"] = "Bearer " + bearer.Value
				break
			}
		}
	case "basic":
		// Basic auth would typically be handled differently in ffuf
		// This is a simplified implementation
		var username, password string
		for _, basic := range auth.Basic {
			if basic.Key == "username" {
				username = basic.Value
			} else if basic.Key == "password" {
				password = basic.Value
			}
		}
		if username != "" && password != "" {
			// In a real implementation, you would encode these properly
			req.Headers["Authorization"] = "Basic " + username + ":" + password
		}
	}
	return nil
}

// ExportToPostmanCollection exports ffuf requests to a Postman collection file
func ExportToPostmanCollection(requests []ffuf.Request, name, outputPath string) error {
	// Create a new Postman collection
	collection := PostmanCollection{
		Info: PostmanInfo{
			Name:   name,
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: make([]PostmanItem, 0, len(requests)),
	}

	// Convert ffuf requests to Postman items
	for i, req := range requests {
		item := PostmanItem{
			Name: fmt.Sprintf("Request %d", i+1),
			Request: &PostmanRequest{
				Method: req.Method,
				URL: PostmanURL{
					Raw: req.Url,
				},
				Header: make([]PostmanHeader, 0, len(req.Headers)),
			},
		}

		// Add headers
		for key, value := range req.Headers {
			item.Request.Header = append(item.Request.Header, PostmanHeader{
				Key:   key,
				Value: value,
			})
		}

		// Add body if present
		if len(req.Data) > 0 {
			item.Request.Body = &PostmanBody{
				Mode: "raw",
				Raw:  string(req.Data),
			}
		}

		collection.Item = append(collection.Item, item)
	}

	// Convert to JSON
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Postman collection: %v", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write to file
	if err := ioutil.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Postman collection file: %v", err)
	}

	return nil
}
