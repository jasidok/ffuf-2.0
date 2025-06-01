package payload

import (
	"encoding/json"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestGenerateQueryParams(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name      string
		baseURL   string
		paramName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Simple URL without query params",
			baseURL:   "https://api.example.com/users",
			paramName: "id",
			want:      "https://api.example.com/users?id=FUZZ",
			wantErr:   false,
		},
		{
			name:      "URL with existing query params",
			baseURL:   "https://api.example.com/users?page=1",
			paramName: "id",
			want:      "https://api.example.com/users?id=FUZZ&page=1",
			wantErr:   false,
		},
		{
			name:      "Replace existing param",
			baseURL:   "https://api.example.com/users?id=123",
			paramName: "id",
			want:      "https://api.example.com/users?id=FUZZ",
			wantErr:   false,
		},
		{
			name:      "Invalid URL",
			baseURL:   "://invalid-url",
			paramName: "id",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateQueryParams(tt.baseURL, tt.paramName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateQueryParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePathParam(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name        string
		urlTemplate string
		paramName   string
		want        string
		wantErr     bool
	}{
		{
			name:        "Simple path parameter",
			urlTemplate: "https://api.example.com/users/{id}",
			paramName:   "id",
			want:        "https://api.example.com/users/FUZZ",
			wantErr:     false,
		},
		{
			name:        "Multiple occurrences of the same parameter",
			urlTemplate: "https://api.example.com/users/{id}/posts/{id}",
			paramName:   "id",
			want:        "https://api.example.com/users/FUZZ/posts/FUZZ",
			wantErr:     false,
		},
		{
			name:        "Parameter not found",
			urlTemplate: "https://api.example.com/users/{userId}",
			paramName:   "id",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GeneratePathParam(tt.urlTemplate, tt.paramName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePathParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GeneratePathParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateJSONWithArrays(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name      string
		template  string
		path      string
		want      string
		wantErr   bool
	}{
		{
			name:     "Simple array element",
			template: "",
			path:     "users.0.name",
			want:     `{
  "users": [
    {
      "name": "FUZZ"
    }
  ]
}`,
			wantErr:  false,
		},
		{
			name:     "Nested array element",
			template: "",
			path:     "users.0.posts.1.title",
			want:     `{
  "users": [
    {
      "posts": [
        null,
        {
          "title": "FUZZ"
        }
      ]
    }
  ]
}`,
			wantErr:  false,
		},
		{
			name:     "Array with existing template",
			template: `{"users":[{"name":"John"},{"name":"Jane"}]}`,
			path:     "users.1.name",
			want:     `{
  "users": [
    {
      "name": "John"
    },
    {
      "name": "FUZZ"
    }
  ]
}`,
			wantErr:  false,
		},
		{
			name:     "Invalid array index",
			template: "",
			path:     "users.invalid.name",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateJSON(tt.template, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Normalize JSON for comparison
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
					t.Errorf("Failed to parse generated JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Errorf("Failed to parse expected JSON: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotObj)
				wantJSON, _ := json.Marshal(wantObj)

				if string(gotJSON) != string(wantJSON) {
					t.Errorf("GenerateJSON() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGenerateJSONWithMultipleFuzzPoints(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name     string
		template string
		paths    []string
		want     string
		wantErr  bool
	}{
		{
			name:     "Multiple simple paths",
			template: "",
			paths:    []string{"username", "email"},
			want:     `{
  "email": "FUZZ",
  "username": "FUZZ"
}`,
			wantErr:  false,
		},
		{
			name:     "Multiple nested paths",
			template: "",
			paths:    []string{"user.name", "user.email", "settings.theme"},
			want:     `{
  "settings": {
    "theme": "FUZZ"
  },
  "user": {
    "email": "FUZZ",
    "name": "FUZZ"
  }
}`,
			wantErr:  false,
		},
		{
			name:     "With existing template",
			template: `{"user":{"name":"John","email":"john@example.com"},"settings":{"theme":"dark"}}`,
			paths:    []string{"user.email", "settings.theme"},
			want:     `{
  "settings": {
    "theme": "FUZZ"
  },
  "user": {
    "email": "FUZZ",
    "name": "John"
  }
}`,
			wantErr:  false,
		},
		{
			name:     "With array paths",
			template: "",
			paths:    []string{"users.0.name", "users.1.name"},
			want:     `{
  "users": [
    {
      "name": "FUZZ"
    },
    {
      "name": "FUZZ"
    }
  ]
}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateJSONWithMultipleFuzzPoints(tt.template, tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJSONWithMultipleFuzzPoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Normalize JSON for comparison
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
					t.Errorf("Failed to parse generated JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Errorf("Failed to parse expected JSON: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotObj)
				wantJSON, _ := json.Marshal(wantObj)

				if string(gotJSON) != string(wantJSON) {
					t.Errorf("GenerateJSONWithMultipleFuzzPoints() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFuzzJSON(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name     string
		template string
		path     string
		values   []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "Simple JSON with single value",
			template: "",
			path:     "username",
			values:   []string{"admin"},
			want:     []string{`{
  "username": "admin"
}`},
			wantErr: false,
		},
		{
			name:     "Simple JSON with multiple values",
			template: "",
			path:     "username",
			values:   []string{"admin", "user", "guest"},
			want: []string{
				`{
  "username": "admin"
}`,
				`{
  "username": "user"
}`,
				`{
  "username": "guest"
}`,
			},
			wantErr: false,
		},
		{
			name:     "Nested JSON with single value",
			template: "",
			path:     "user.credentials.password",
			values:   []string{"password123"},
			want:     []string{`{
  "user": {
    "credentials": {
      "password": "password123"
    }
  }
}`},
			wantErr: false,
		},
		{
			name:     "Array JSON with single value",
			template: "",
			path:     "users.0.name",
			values:   []string{"John"},
			want:     []string{`{
  "users": [
    {
      "name": "John"
    }
  ]
}`},
			wantErr: false,
		},
		{
			name:     "With existing template",
			template: `{"user":{"name":"John","email":"john@example.com"}}`,
			path:     "user.email",
			values:   []string{"admin@example.com"},
			want:     []string{`{
  "user": {
    "email": "admin@example.com",
    "name": "John"
  }
}`},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.FuzzJSON(tt.template, tt.path, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("FuzzJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FuzzJSON() returned %d payloads, want %d", len(got), len(tt.want))
					return
				}

				for i := range got {
					// Normalize JSON for comparison
					var gotObj, wantObj interface{}
					if err := json.Unmarshal([]byte(got[i]), &gotObj); err != nil {
						t.Errorf("Failed to parse generated JSON: %v", err)
						return
					}
					if err := json.Unmarshal([]byte(tt.want[i]), &wantObj); err != nil {
						t.Errorf("Failed to parse expected JSON: %v", err)
						return
					}

					gotJSON, _ := json.Marshal(gotObj)
					wantJSON, _ := json.Marshal(wantObj)

					if string(gotJSON) != string(wantJSON) {
						t.Errorf("FuzzJSON() payload %d = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestFuzzJSONWithMultipleFuzzPoints(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name     string
		template string
		paths    []string
		values   []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "Multiple simple paths with single value",
			template: "",
			paths:    []string{"username", "email"},
			values:   []string{"admin"},
			want:     []string{`{
  "email": "admin",
  "username": "admin"
}`},
			wantErr: false,
		},
		{
			name:     "Multiple simple paths with multiple values",
			template: "",
			paths:    []string{"username", "email"},
			values:   []string{"admin", "user"},
			want: []string{
				`{
  "email": "admin",
  "username": "admin"
}`,
				`{
  "email": "user",
  "username": "user"
}`,
			},
			wantErr: false,
		},
		{
			name:     "Multiple nested paths with single value",
			template: "",
			paths:    []string{"user.name", "user.email", "settings.theme"},
			values:   []string{"admin"},
			want:     []string{`{
  "settings": {
    "theme": "admin"
  },
  "user": {
    "email": "admin",
    "name": "admin"
  }
}`},
			wantErr: false,
		},
		{
			name:     "With existing template",
			template: `{"user":{"name":"John","email":"john@example.com"},"settings":{"theme":"dark"}}`,
			paths:    []string{"user.email", "settings.theme"},
			values:   []string{"admin@example.com"},
			want:     []string{`{
  "settings": {
    "theme": "admin@example.com"
  },
  "user": {
    "email": "admin@example.com",
    "name": "John"
  }
}`},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.FuzzJSONWithMultipleFuzzPoints(tt.template, tt.paths, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("FuzzJSONWithMultipleFuzzPoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FuzzJSONWithMultipleFuzzPoints() returned %d payloads, want %d", len(got), len(tt.want))
					return
				}

				for i := range got {
					// Normalize JSON for comparison
					var gotObj, wantObj interface{}
					if err := json.Unmarshal([]byte(got[i]), &gotObj); err != nil {
						t.Errorf("Failed to parse generated JSON: %v", err)
						return
					}
					if err := json.Unmarshal([]byte(tt.want[i]), &wantObj); err != nil {
						t.Errorf("Failed to parse expected JSON: %v", err)
						return
					}

					gotJSON, _ := json.Marshal(gotObj)
					wantJSON, _ := json.Marshal(wantObj)

					if string(gotJSON) != string(wantJSON) {
						t.Errorf("FuzzJSONWithMultipleFuzzPoints() payload %d = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestGenerateGraphQLWithFuzzPoint(t *testing.T) {
	generator := NewPayloadGenerator(FormatGraphQL)

	tests := []struct {
		name      string
		query     string
		variables map[string]interface{}
		want      string
		wantErr   bool
	}{
		{
			name:      "Empty query",
			query:     "",
			variables: nil,
			want: `{
  "query": "query {\n  FUZZ\n}"
}`,
			wantErr: false,
		},
		{
			name:      "Query without fuzz marker",
			query:     "query { user { name email } }",
			variables: nil,
			want: `{
  "query": "query {\n  FUZZ\n}"
}`,
			wantErr: false,
		},
		{
			name:      "Query with fuzz marker",
			query:     "query { user { FUZZ } }",
			variables: nil,
			want: `{
  "query": "query { user { FUZZ } }"
}`,
			wantErr: false,
		},
		{
			name: "Query with variables",
			query: "query($id: ID!) { user(id: $id) { FUZZ } }",
			variables: map[string]interface{}{
				"id": "123",
			},
			want: `{
  "query": "query($id: ID!) { user(id: $id) { FUZZ } }",
  "variables": {
    "id": "123"
  }
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateGraphQLWithFuzzPoint(tt.query, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateGraphQLWithFuzzPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Normalize JSON for comparison
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
					t.Errorf("Failed to parse generated JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Errorf("Failed to parse expected JSON: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotObj)
				wantJSON, _ := json.Marshal(wantObj)

				if string(gotJSON) != string(wantJSON) {
					t.Errorf("GenerateGraphQLWithFuzzPoint() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGenerateGraphQLWithVariableFuzzPoint(t *testing.T) {
	generator := NewPayloadGenerator(FormatGraphQL)

	tests := []struct {
		name         string
		query        string
		variableName string
		want         string
		wantErr      bool
	}{
		{
			name:         "Empty query",
			query:        "",
			variableName: "id",
			want: `{
  "query": "query($id: String!) {\n  field(param: $id)\n}",
  "variables": {
    "id": "FUZZ"
  }
}`,
			wantErr: false,
		},
		{
			name:         "Custom query",
			query:        "query($id: ID!) { user(id: $id) { name email } }",
			variableName: "id",
			want: `{
  "query": "query($id: ID!) { user(id: $id) { name email } }",
  "variables": {
    "id": "FUZZ"
  }
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateGraphQLWithVariableFuzzPoint(tt.query, tt.variableName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateGraphQLWithVariableFuzzPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Normalize JSON for comparison
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
					t.Errorf("Failed to parse generated JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Errorf("Failed to parse expected JSON: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotObj)
				wantJSON, _ := json.Marshal(wantObj)

				if string(gotJSON) != string(wantJSON) {
					t.Errorf("GenerateGraphQLWithVariableFuzzPoint() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFuzzGraphQL(t *testing.T) {
	generator := NewPayloadGenerator(FormatGraphQL)

	tests := []struct {
		name      string
		query     string
		variables map[string]interface{}
		values    []string
		want      []string
		wantErr   bool
	}{
		{
			name:      "Simple query with single value",
			query:     "query { FUZZ }",
			variables: nil,
			values:    []string{"user { name email }"},
			want: []string{`{
  "query": "query { user { name email } }"
}`},
			wantErr: false,
		},
		{
			name:      "Simple query with multiple values",
			query:     "query { FUZZ }",
			variables: nil,
			values:    []string{"user { name }", "post { title }"},
			want: []string{
				`{
  "query": "query { user { name } }"
}`,
				`{
  "query": "query { post { title } }"
}`,
			},
			wantErr: false,
		},
		{
			name: "Query with variables",
			query: "query($id: ID!) { user(id: $id) { FUZZ } }",
			variables: map[string]interface{}{
				"id": "123",
			},
			values: []string{"name email", "posts { id title }"},
			want: []string{
				`{
  "query": "query($id: ID!) { user(id: $id) { name email } }",
  "variables": {
    "id": "123"
  }
}`,
				`{
  "query": "query($id: ID!) { user(id: $id) { posts { id title } } }",
  "variables": {
    "id": "123"
  }
}`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.FuzzGraphQL(tt.query, tt.variables, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("FuzzGraphQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FuzzGraphQL() returned %d payloads, want %d", len(got), len(tt.want))
					return
				}

				for i := range got {
					// Normalize JSON for comparison
					var gotObj, wantObj interface{}
					if err := json.Unmarshal([]byte(got[i]), &gotObj); err != nil {
						t.Errorf("Failed to parse generated JSON: %v", err)
						return
					}
					if err := json.Unmarshal([]byte(tt.want[i]), &wantObj); err != nil {
						t.Errorf("Failed to parse expected JSON: %v", err)
						return
					}

					gotJSON, _ := json.Marshal(gotObj)
					wantJSON, _ := json.Marshal(wantObj)

					if string(gotJSON) != string(wantJSON) {
						t.Errorf("FuzzGraphQL() payload %d = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestFuzzGraphQLVariable(t *testing.T) {
	generator := NewPayloadGenerator(FormatGraphQL)

	tests := []struct {
		name         string
		query        string
		variableName string
		values       []string
		want         []string
		wantErr      bool
	}{
		{
			name:         "Simple query with single value",
			query:        "query($id: ID!) { user(id: $id) { name email } }",
			variableName: "id",
			values:       []string{"123"},
			want: []string{`{
  "query": "query($id: ID!) { user(id: $id) { name email } }",
  "variables": {
    "id": "123"
  }
}`},
			wantErr: false,
		},
		{
			name:         "Simple query with multiple values",
			query:        "query($id: ID!) { user(id: $id) { name email } }",
			variableName: "id",
			values:       []string{"123", "456"},
			want: []string{
				`{
  "query": "query($id: ID!) { user(id: $id) { name email } }",
  "variables": {
    "id": "123"
  }
}`,
				`{
  "query": "query($id: ID!) { user(id: $id) { name email } }",
  "variables": {
    "id": "456"
  }
}`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.FuzzGraphQLVariable(tt.query, tt.variableName, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("FuzzGraphQLVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FuzzGraphQLVariable() returned %d payloads, want %d", len(got), len(tt.want))
					return
				}

				for i := range got {
					// Normalize JSON for comparison
					var gotObj, wantObj interface{}
					if err := json.Unmarshal([]byte(got[i]), &gotObj); err != nil {
						t.Errorf("Failed to parse generated JSON: %v", err)
						return
					}
					if err := json.Unmarshal([]byte(tt.want[i]), &wantObj); err != nil {
						t.Errorf("Failed to parse expected JSON: %v", err)
						return
					}

					gotJSON, _ := json.Marshal(gotObj)
					wantJSON, _ := json.Marshal(wantObj)

					if string(gotJSON) != string(wantJSON) {
						t.Errorf("FuzzGraphQLVariable() payload %d = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestGenerateRESTRequest(t *testing.T) {
	generator := NewPayloadGenerator(FormatJSON)

	tests := []struct {
		name      string
		baseReq   *ffuf.Request
		paramType string
		paramName string
		check     func(*testing.T, *ffuf.Request)
		wantErr   bool
	}{
		{
			name: "Query parameter",
			baseReq: &ffuf.Request{
				Method:  "GET",
				Url:     "https://api.example.com/users",
				Headers: make(map[string]string),
			},
			paramType: "query",
			paramName: "id",
			check: func(t *testing.T, req *ffuf.Request) {
				if req.Url != "https://api.example.com/users?id=FUZZ" {
					t.Errorf("Expected URL with query param, got %s", req.Url)
				}
			},
			wantErr: false,
		},
		{
			name: "Path parameter",
			baseReq: &ffuf.Request{
				Method:  "GET",
				Url:     "https://api.example.com/users/{id}",
				Headers: make(map[string]string),
			},
			paramType: "path",
			paramName: "id",
			check: func(t *testing.T, req *ffuf.Request) {
				if req.Url != "https://api.example.com/users/FUZZ" {
					t.Errorf("Expected URL with path param, got %s", req.Url)
				}
			},
			wantErr: false,
		},
		{
			name: "Header parameter",
			baseReq: &ffuf.Request{
				Method:  "GET",
				Url:     "https://api.example.com/users",
				Headers: make(map[string]string),
			},
			paramType: "header",
			paramName: "Authorization",
			check: func(t *testing.T, req *ffuf.Request) {
				if req.Headers["Authorization"] != "FUZZ" {
					t.Errorf("Expected header with FUZZ value, got %s", req.Headers["Authorization"])
				}
			},
			wantErr: false,
		},
		{
			name: "JSON body parameter",
			baseReq: &ffuf.Request{
				Method: "POST",
				Url:    "https://api.example.com/users",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Data: []byte(`{"user":{"name":"John","email":"john@example.com"}}`),
			},
			paramType: "body",
			paramName: "user.email",
			check: func(t *testing.T, req *ffuf.Request) {
				var data map[string]interface{}
				err := json.Unmarshal(req.Data, &data)
				if err != nil {
					t.Errorf("Failed to parse JSON body: %v", err)
					return
				}

				user, ok := data["user"].(map[string]interface{})
				if !ok {
					t.Errorf("Expected user object in body")
					return
				}

				if user["email"] != "FUZZ" {
					t.Errorf("Expected email to be FUZZ, got %v", user["email"])
				}
			},
			wantErr: false,
		},
		{
			name: "Unsupported parameter type",
			baseReq: &ffuf.Request{
				Method:  "GET",
				Url:     "https://api.example.com/users",
				Headers: make(map[string]string),
			},
			paramType: "invalid",
			paramName: "id",
			check:     func(t *testing.T, req *ffuf.Request) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.GenerateRESTRequest(tt.baseReq, tt.paramType, tt.paramName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRESTRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				tt.check(t, got)
			}
		})
	}
}
