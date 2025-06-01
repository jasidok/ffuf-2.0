# ffuf API Testing Cheat Sheet

This cheat sheet provides a quick reference for API testing commands and techniques using ffuf.

## Command Line Options

### Basic Options

| Option | Description | Example |
|--------|-------------|---------|
| `-u` | Target URL | `-u https://api.example.com/FUZZ` |
| `-w` | Wordlist file path | `-w /path/to/wordlist.txt` |
| `-X` | HTTP method | `-X POST` |
| `-d` | POST data | `-d '{"key":"value"}'` |
| `-H` | HTTP header | `-H "Content-Type: application/json"` |
| `-c` | Concurrency level | `-c 50` |
| `-rate` | Rate of requests per second | `-rate 10` |
| `-timeout` | HTTP request timeout in seconds | `-timeout 10` |
| `-api_mode` | Enable API mode | `-api_mode` |

### Matcher Options

| Option | Description | Example |
|--------|-------------|---------|
| `-mc` | Match HTTP status codes | `-mc 200,201,204` |
| `-ml` | Match response line count | `-ml 5` |
| `-mr` | Match response using regex | `-mr "success"` |
| `-ms` | Match response size | `-ms 1024` |
| `-mw` | Match response word count | `-mw 42` |

### Filter Options

| Option | Description | Example |
|--------|-------------|---------|
| `-fc` | Filter HTTP status codes | `-fc 404,500` |
| `-fl` | Filter by response line count | `-fl 5` |
| `-fr` | Filter response using regex | `-fr "error"` |
| `-fs` | Filter response size | `-fs 1024` |
| `-fw` | Filter response word count | `-fw 42` |
| `-ft` | Filter by response time | `-ft 100` |

### Output Options

| Option | Description | Example |
|--------|-------------|---------|
| `-o` | Output file path | `-o results.txt` |
| `-of` | Output format (json, ejson, html, md, csv, ecsv) | `-of json` |
| `-v` | Verbose output | `-v` |
| `-json` | JSON output | `-json` |
| `-silent` | Silent mode | `-silent` |
| `-debug-log` | Write debug logs to file | `-debug-log debug.txt` |

## API Testing Commands

### REST API Testing

```bash
# Basic GET request
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer TOKEN"

# POST request with JSON payload
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"John","email":"john@example.com"}'

# PUT request to update a resource
ffuf -u https://api.example.com/v1/users/1 -X PUT -H "Content-Type: application/json" -d '{"name":"John Updated"}'

# DELETE request
ffuf -u https://api.example.com/v1/users/1 -X DELETE -H "Authorization: Bearer TOKEN"
```

### API Endpoint Discovery

```bash
# Discover API endpoints
ffuf -u https://api.example.com/FUZZ -w /path/to/api_wordlist.txt -H "Content-Type: application/json"

# Discover API endpoints with version prefix
ffuf -u https://api.example.com/vFUZZ/users -w /path/to/versions.txt

# Recursive endpoint discovery
ffuf -u https://api.example.com/FUZZ -w /path/to/api_wordlist.txt -recursion -recursion-depth 2
```

### API Parameter Fuzzing

```bash
# Fuzz query parameters
ffuf -u https://api.example.com/v1/search?FUZZ=test -w /path/to/params.txt

# Fuzz query parameter values
ffuf -u https://api.example.com/v1/search?query=FUZZ -w /path/to/values.txt

# Fuzz JSON body parameters
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"FUZZ":"test"}' -w /path/to/params.txt

# Fuzz JSON body values
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"FUZZ"}' -w /path/to/names.txt

# Fuzz multiple parameters
ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZW","filter":"FUZZX"}' -w /path/to/queries.txt:FUZZW -w /path/to/filters.txt:FUZZX
```

### Authentication Testing

```bash
# Test API keys
ffuf -u https://api.example.com/v1/users -H "X-API-Key: FUZZ" -w /path/to/api_keys.txt -fc 401

# Test OAuth tokens
ffuf -u https://api.example.com/v1/users -H "Authorization: Bearer FUZZ" -w /path/to/tokens.txt -fc 401

# Test basic authentication
ffuf -u https://api.example.com/v1/users -X GET -u FUZZ:password -w /path/to/usernames.txt
```

### GraphQL API Testing

```bash
# GraphQL introspection
ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"{ __schema { types { name fields { name } } } }"}'

# GraphQL query fuzzing
ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"{ user(id: \"FUZZ\") { name email } }"}' -w /path/to/ids.txt

# GraphQL mutation testing
ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"mutation { updateUser(id: \"1\", input: { name: \"FUZZ\" }) { id name } }"}' -w /path/to/names.txt
```

### API Security Testing

```bash
# BOLA/IDOR testing
ffuf -u https://api.example.com/v1/users/FUZZ/profile -w /path/to/ids.txt -H "Authorization: Bearer TOKEN" -mc all

# SQL injection testing
ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZ"}' -w /path/to/sql_injections.txt -fr "error"

# NoSQL injection testing
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"username":{"$ne":"FUZZ"}}' -w /path/to/nosql_payloads.txt

# Mass assignment testing
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"John","email":"john@example.com","FUZZ":"value"}' -w /path/to/privileged_params.txt

# Rate limiting testing
ffuf -u https://api.example.com/v1/login -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}' -rate 50
```

## Interactive API Console Commands

| Command | Description | Example |
|---------|-------------|---------|
| `method [METHOD]` | Set HTTP method | `method POST` |
| `url [URL]` | Set base URL | `url https://api.example.com` |
| `path [PATH]` | Set request path | `path /v1/users` |
| `header [NAME] [VALUE]` | Add a header | `header Authorization Bearer TOKEN` |
| `header remove [NAME]` | Remove a header | `header remove Authorization` |
| `header clear` | Clear all headers | `header clear` |
| `query [NAME] [VALUE]` | Add a query parameter | `query sort desc` |
| `query remove [NAME]` | Remove a query parameter | `query remove sort` |
| `query clear` | Clear all query parameters | `query clear` |
| `body [NAME] [VALUE]` | Add a body parameter | `body name John` |
| `body remove [NAME]` | Remove a body parameter | `body remove name` |
| `body clear` | Clear all body parameters | `body clear` |
| `body json [JSON_STRING]` | Set body from JSON string | `body json {"name":"John"}` |
| `format [FORMAT]` | Set body format | `format json` |
| `show` | Show current request | `show` |
| `send` | Send the request | `send` |
| `diff` | Compare the last two API responses | `diff` |
| `map [FORMAT]` | Generate API structure visualization | `map html` |
| `help` | Show help | `help` |

## Common API Wordlists

| Wordlist | Description | Path |
|----------|-------------|------|
| API Endpoints | Common API endpoint names | `/path/to/api_endpoints.txt` |
| API Versions | Common API version numbers | `/path/to/api_versions.txt` |
| API Parameters | Common API parameter names | `/path/to/api_params.txt` |
| API Methods | Common API method names | `/path/to/api_methods.txt` |
| API Headers | Common API header names | `/path/to/api_headers.txt` |
| API Status Codes | Common API status codes | `/path/to/api_status_codes.txt` |

## Response Filtering Examples

```bash
# Filter out 404 responses
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -fc 404

# Match only 2xx responses
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -mc 200,201,202,203,204

# Filter by response size
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -fs 42

# Filter by response time (ms)
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -ft 500

# Match responses containing specific text
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -mr "success"

# Filter responses containing error messages
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -fr "error|exception|invalid"
```

## Advanced Techniques

```bash
# Use custom HTTP method
ffuf -u https://api.example.com/v1/users -X PATCH -H "Content-Type: application/json" -d '{"status":"active"}'

# Test with client certificates
ffuf -u https://api.example.com/v1/secure -X GET -cert /path/to/cert.pem -key /path/to/key.pem

# Follow redirects
ffuf -u https://api.example.com/v1/redirect -X GET -r

# Set custom cookies
ffuf -u https://api.example.com/v1/users -X GET -b "session=FUZZ" -w /path/to/session_tokens.txt

# Use proxy
ffuf -u https://api.example.com/v1/users -x http://127.0.0.1:8080

# Save matched responses to files
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -od /path/to/output/dir

# Replay a previous request with modifications
ffuf -replay-proxy http://127.0.0.1:8080 -u https://api.example.com/v1/users
```

## Tips & Tricks

- Use `-v` for verbose output to see full request and response details
- Use `-json` for structured output that can be parsed by other tools
- Combine multiple wordlists with different FUZZ keywords (FUZZX, FUZZY, etc.)
- Use `-rate` to control request rate and avoid triggering rate limiting
- Use `-timeout` to handle slow API responses
- Use `-recursion` to discover nested API endpoints
- Use the interactive mode for exploratory API testing
- Use the `diff` command to compare API responses
- Use the `map` command to visualize API structures