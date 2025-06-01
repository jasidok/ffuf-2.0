# ffuf Development Guidelines - API Focus

This document provides guidelines for developing and working with ffuf, with a specific focus on API interactions. ffuf is a web fuzzer that sends HTTP requests to target websites, making it an excellent tool for testing and fuzzing web APIs.

## Build/Configuration Instructions

### Building ffuf

ffuf is written in Go and requires Go 1.16 or greater. Here are the ways to build ffuf:

1. **Using go install**:
   ```
   go install github.com/ffuf/ffuf/v2@latest
   ```

2. **Building from source**:
   ```
   git clone https://github.com/ffuf/ffuf
   cd ffuf
   go build
   ```

### Configuration

ffuf uses configuration files to store default settings. The default path for the configuration file is `$XDG_CONFIG_HOME/ffuf/ffufrc`. You can also specify a custom configuration file using the `-config` flag.

Example configuration file (ffufrc.example):
```toml
[http]
# Timeout in seconds
timeout = 10
# Follow redirects
followredirects = false
# HTTP proxy URL
proxyurl = "http://127.0.0.1:8080"
# Target URL
url = "https://example.org/FUZZ"
```

## Testing Information

### Running Tests

ffuf uses Go's standard testing framework. To run all tests:

```
go test ./...
```

To run tests in a specific package:

```
go test ./pkg/runner
```

To run a specific test with verbose output:

```
go test -v ./pkg/runner -run TestSimpleRunnerExecute
```

### Writing Tests for API Interactions

When testing API interactions, use Go's `httptest` package to create mock HTTP servers. This allows you to test your code without making actual HTTP requests to external servers.

Example test for API interaction:

```
func TestAPIInteraction(t *testing.T) {
    // Create a mock HTTP server
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer ts.Close()

    // Create a config for the runner
    config := &ffuf.Config{
        Context:         context.Background(),
        Timeout:         10,
        FollowRedirects: false,
    }

    // Create a runner
    runner := NewSimpleRunner(config, false)

    // Create a request
    req := &ffuf.Request{
        Method:  "GET",
        Url:     ts.URL,
        Headers: make(map[string]string),
    }

    // Execute the request
    resp, err := runner.Execute(req)
    if err != nil {
        t.Errorf("Error executing request: %v", err)
    }

    // Check the response
    if resp.StatusCode != 200 {
        t.Errorf("Expected status code 200, got %d", resp.StatusCode)
    }

    // Check the content type
    if resp.ContentType != "application/json" {
        t.Errorf("Expected content type application/json, got %s", resp.ContentType)
    }

    // Check the response body
    if string(resp.Data) != `{"status":"ok"}` {
        t.Errorf("Expected response body {\"status\":\"ok\"}, got %s", string(resp.Data))
    }
}
```

## API Development Information

### Key Components for API Interaction

1. **Request Structure**: The `Request` struct in `pkg/ffuf/request.go` represents an HTTP request. It contains fields for the HTTP method, URL, headers, and data (body).

2. **Response Structure**: The `Response` struct in `pkg/ffuf/response.go` represents an HTTP response. It contains fields for the status code, headers, body, and timing information.

3. **Runner**: The `SimpleRunner` in `pkg/runner/simple.go` is responsible for executing HTTP requests. It handles the actual HTTP request/response cycle, including timing, handling different content encodings, and processing the response.

### Making API Requests

To make an API request:

1. Create a `Request` struct with the appropriate method, URL, headers, and data.
2. Create a `SimpleRunner` with the desired configuration.
3. Call the `Execute` method on the runner with the request.
4. Process the response.

Example:

```
// Create a request
req := &ffuf.Request{
    Method:  "POST",
    Url:     "https://api.example.com/endpoint",
    Headers: map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer token",
    },
    Data: []byte(`{"key":"value"}`),
}

// Create a runner
config := &ffuf.Config{
    Context:         context.Background(),
    Timeout:         10,
    FollowRedirects: true,
}
runner := NewSimpleRunner(config, false)

// Execute the request
resp, err := runner.Execute(req)
if err != nil {
    // Handle error
}

// Process the response
fmt.Printf("Status: %d\n", resp.StatusCode)
fmt.Printf("Body: %s\n", string(resp.Data))
```

### Handling Different API Response Types

ffuf can handle various response types, including JSON, XML, and HTML. The `Response` struct contains the raw response data, which you can parse according to the content type.

For JSON responses, you can use Go's `encoding/json` package:

```
var data map[string]interface{}
err := json.Unmarshal(resp.Data, &data)
if err != nil {
    // Handle error
}
```

### API Fuzzing

ffuf is designed for fuzzing web applications, including APIs. You can use it to fuzz API endpoints, parameters, headers, and request bodies.

Example of fuzzing an API endpoint:

```
ffuf -w /path/to/wordlist -u https://api.example.com/FUZZ -H "Content-Type: application/json" -H "Authorization: Bearer token"
```

Example of fuzzing a JSON parameter:

```
ffuf -w /path/to/wordlist -u https://api.example.com/endpoint -X POST -H "Content-Type: application/json" -d '{"key":"FUZZ"}' -fr "error"
```

## Conclusion

ffuf is a powerful tool for testing and fuzzing web APIs. By understanding its API interaction components and how to test them, you can effectively develop and extend ffuf for your API testing needs.