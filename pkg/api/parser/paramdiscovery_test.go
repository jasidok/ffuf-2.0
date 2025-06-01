package parser

import (
	"testing"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestDiscoverURLParameters(t *testing.T) {
	discovery := NewParameterDiscovery()
	
	tests := []struct {
		name     string
		url      string
		wantPath []string
		wantQuery []string
	}{
		{
			name:      "Simple URL with query parameters",
			url:       "https://api.example.com/users?page=1&limit=10",
			wantPath:  []string{},
			wantQuery: []string{"page", "limit"},
		},
		{
			name:      "URL with path parameters",
			url:       "https://api.example.com/users/{id}/posts/{postId}",
			wantPath:  []string{"id", "postId"},
			wantQuery: []string{},
		},
		{
			name:      "URL with both path and query parameters",
			url:       "https://api.example.com/users/{userId}?include=posts&sort=desc",
			wantPath:  []string{"userId"},
			wantQuery: []string{"include", "sort"},
		},
		{
			name:      "URL with no parameters",
			url:       "https://api.example.com/status",
			wantPath:  []string{},
			wantQuery: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := discovery.discoverURLParameters(tt.url)
			if err != nil {
				t.Fatalf("discoverURLParameters() error = %v", err)
			}
			
			// Count path and query parameters
			pathParams := make([]string, 0)
			queryParams := make([]string, 0)
			
			for _, param := range params {
				if param.Type == ParamTypePath {
					pathParams = append(pathParams, param.Name)
				} else if param.Type == ParamTypeQuery {
					queryParams = append(queryParams, param.Name)
				}
			}
			
			// Check path parameters
			if len(pathParams) != len(tt.wantPath) {
				t.Errorf("Got %d path parameters, want %d", len(pathParams), len(tt.wantPath))
			}
			
			for _, wantParam := range tt.wantPath {
				found := false
				for _, gotParam := range pathParams {
					if gotParam == wantParam {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Path parameter %s not found", wantParam)
				}
			}
			
			// Check query parameters
			if len(queryParams) != len(tt.wantQuery) {
				t.Errorf("Got %d query parameters, want %d", len(queryParams), len(tt.wantQuery))
			}
			
			for _, wantParam := range tt.wantQuery {
				found := false
				for _, gotParam := range queryParams {
					if gotParam == wantParam {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Query parameter %s not found", wantParam)
				}
			}
		})
	}
}

func TestDiscoverJSONParameters(t *testing.T) {
	discovery := NewParameterDiscovery()
	
	jsonData := []byte(`{
		"user": {
			"id": 123,
			"name": "John Doe",
			"email": "john@example.com"
		},
		"posts": [
			{
				"id": 1,
				"title": "First Post",
				"content": "This is the first post"
			},
			{
				"id": 2,
				"title": "Second Post",
				"content": "This is the second post"
			}
		],
		"pagination": {
			"page": 1,
			"limit": 10,
			"total": 2
		}
	}`)
	
	params, err := discovery.discoverJSONParameters(jsonData)
	if err != nil {
		t.Fatalf("discoverJSONParameters() error = %v", err)
	}
	
	// Check that we found the expected parameters
	expectedParams := map[string]string{
		"user":       "object",
		"id":         "number",
		"name":       "string",
		"email":      "string",
		"posts":      "array",
		"title":      "string",
		"content":    "string",
		"pagination": "object",
		"page":       "number",
		"limit":      "number",
		"total":      "number",
	}
	
	for paramName, dataType := range expectedParams {
		found := false
		for _, param := range params {
			if param.Name == paramName && param.DataType == dataType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Parameter %s with data type %s not found", paramName, dataType)
		}
	}
	
	// Check that known parameters have the correct type
	for _, param := range params {
		if param.Name == "page" || param.Name == "limit" {
			if param.Type != ParamTypeQuery {
				t.Errorf("Parameter %s has type %s, want %s", param.Name, param.Type, ParamTypeQuery)
			}
		}
		if param.Name == "id" {
			if param.Type != ParamTypePath {
				t.Errorf("Parameter %s has type %s, want %s", param.Name, param.Type, ParamTypePath)
			}
		}
	}
}

func TestExtractURLsFromJSON(t *testing.T) {
	discovery := NewParameterDiscovery()
	
	jsonData := map[string]interface{}{
		"links": map[string]interface{}{
			"self": "https://api.example.com/users/123",
			"next": "https://api.example.com/users?page=2",
		},
		"user": map[string]interface{}{
			"id": 123,
			"profile_url": "https://example.com/profiles/123",
		},
		"related": []interface{}{
			map[string]interface{}{
				"href": "https://api.example.com/posts/456",
			},
			map[string]interface{}{
				"href": "https://api.example.com/comments/789",
			},
		},
	}
	
	urls := discovery.extractURLsFromJSON(jsonData)
	
	expectedURLs := []string{
		"https://api.example.com/users/123",
		"https://api.example.com/users?page=2",
		"https://example.com/profiles/123",
		"https://api.example.com/posts/456",
		"https://api.example.com/comments/789",
	}
	
	if len(urls) != len(expectedURLs) {
		t.Errorf("extractURLsFromJSON() got %d URLs, want %d", len(urls), len(expectedURLs))
	}
	
	for _, expectedURL := range expectedURLs {
		found := false
		for _, url := range urls {
			if url == expectedURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("URL %s not found", expectedURL)
		}
	}
}

func TestDiscoverParameters(t *testing.T) {
	discovery := NewParameterDiscovery()
	
	// Create a mock response
	resp := &ffuf.Response{
		Request: &ffuf.Request{
			Url: "https://api.example.com/users/123?include=posts",
		},
		Headers: map[string][]string{
			"Link": {"<https://api.example.com/users?page=2>; rel=\"next\""},
		},
		ContentType: "application/json",
		Data: []byte(`{
			"user": {
				"id": 123,
				"name": "John Doe"
			},
			"links": {
				"self": "https://api.example.com/users/123",
				"posts": "https://api.example.com/users/123/posts"
			}
		}`),
	}
	
	params, err := discovery.DiscoverParameters(resp)
	if err != nil {
		t.Fatalf("DiscoverParameters() error = %v", err)
	}
	
	// Check that we found parameters from different sources
	foundURLParam := false
	foundLinkParam := false
	foundJSONParam := false
	
	for _, param := range params {
		if param.Name == "include" && param.Type == ParamTypeQuery {
			foundURLParam = true
		}
		if param.Name == "page" && param.Type == ParamTypeQuery {
			foundLinkParam = true
		}
		if param.Name == "id" && param.Type == ParamTypePath {
			foundJSONParam = true
		}
	}
	
	if !foundURLParam {
		t.Errorf("URL parameter 'include' not found")
	}
	if !foundLinkParam {
		t.Errorf("Link parameter 'page' not found")
	}
	if !foundJSONParam {
		t.Errorf("JSON parameter 'id' not found")
	}
}

func TestResponseParserDiscoverParameters(t *testing.T) {
	parser := NewResponseParser("application/json")
	
	// Create a mock response
	resp := &ffuf.Response{
		Request: &ffuf.Request{
			Url: "https://api.example.com/users?page=1",
		},
		Headers: map[string][]string{},
		ContentType: "application/json",
		Data: []byte(`{"id": 123, "name": "John"}`),
	}
	
	params, err := parser.DiscoverParameters(resp)
	if err != nil {
		t.Fatalf("DiscoverParameters() error = %v", err)
	}
	
	if len(params) < 3 {
		t.Errorf("DiscoverParameters() got %d parameters, want at least 3", len(params))
	}
	
	// Check for specific parameters
	foundPage := false
	foundId := false
	foundName := false
	
	for _, param := range params {
		if param.Name == "page" && param.Type == ParamTypeQuery {
			foundPage = true
		}
		if param.Name == "id" && param.Type == ParamTypePath {
			foundId = true
		}
		if param.Name == "name" && param.Type == ParamTypeBody {
			foundName = true
		}
	}
	
	if !foundPage {
		t.Errorf("Parameter 'page' not found")
	}
	if !foundId {
		t.Errorf("Parameter 'id' not found")
	}
	if !foundName {
		t.Errorf("Parameter 'name' not found")
	}
}