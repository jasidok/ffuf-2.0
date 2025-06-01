# Comprehensive API Testing Guide for ffuf

## Introduction

ffuf (Fuzz Faster U Fool) is a powerful web fuzzing tool that can be used for API testing. This guide provides comprehensive information on how to use ffuf for API testing, from basic concepts to advanced techniques.

## Table of Contents

1. [Getting Started with API Testing](#getting-started-with-api-testing)
2. [Basic API Testing](#basic-api-testing)
3. [Advanced API Testing](#advanced-api-testing)
4. [Interactive API Console](#interactive-api-console)
5. [API Response Analysis](#api-response-analysis)
6. [API Security Testing](#api-security-testing)
7. [Troubleshooting](#troubleshooting)

## Getting Started with API Testing

### What is API Testing?

API testing involves testing application programming interfaces directly and as part of integration testing to determine if they meet expectations for functionality, reliability, performance, and security.

### Why Use ffuf for API Testing?

ffuf offers several advantages for API testing:

- Fast and efficient fuzzing capabilities
- Support for various API formats (REST, GraphQL, etc.)
- Flexible configuration options
- Interactive API testing console
- Advanced response analysis features
- Integration with other API testing tools

### Installation

To use ffuf for API testing, you need to install it first:

```bash
# Using go install
go install github.com/ffuf/ffuf/v2@latest

# Or build from source
git clone https://github.com/ffuf/ffuf.git
cd ffuf
go build
```

### Configuration

ffuf uses a configuration file for default settings. You can create a configuration file based on the example provided in the repository:

```bash
cp ffufrc.example ~/.config/ffuf/ffufrc
```

To enable API mode, add the following to your configuration file:

```toml
[general]
api_mode = true
```

## Basic API Testing

### Testing REST APIs

To test a REST API endpoint:

```bash
ffuf -u https://api.example.com/v1/FUZZ -w /path/to/endpoints.txt -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_TOKEN"
```

### Testing with Different HTTP Methods

```bash
# GET request
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN"

# POST request with JSON payload
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"FUZZ","email":"test@example.com"}' -w /path/to/names.txt
```

### Handling Authentication

```bash
# Basic authentication
ffuf -u https://api.example.com/v1/users -X GET -u username:password

# API key in header
ffuf -u https://api.example.com/v1/users -X GET -H "X-API-Key: YOUR_API_KEY"

# OAuth token
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN"
```

## Advanced API Testing

### JSON Payload Fuzzing

To fuzz specific fields in a JSON payload:

```bash
ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"FUZZ","email":"test@example.com"}' -w /path/to/names.txt
```

### GraphQL Query Fuzzing

To fuzz GraphQL queries:

```bash
ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"{ user(id: \"FUZZ\") { name email } }"}' -w /path/to/ids.txt
```

### API Parameter Discovery

To discover API parameters:

```bash
ffuf -u https://api.example.com/v1/users?FUZZ=test -w /path/to/params.txt
```

### Testing API Endpoints with Multiple Parameters

```bash
ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZW","filter":"FUZZX"}' -w /path/to/queries.txt:FUZZW -w /path/to/filters.txt:FUZZX
```

## Interactive API Console

ffuf provides an interactive API console for testing APIs. To use it:

1. Start ffuf with API mode enabled:
   ```bash
   ffuf -u https://api.example.com -api_mode
   ```

2. Press ENTER during execution to enter interactive mode.

3. Type `api start` to enter the API console.

### API Console Commands

- `method [METHOD]` - Set HTTP method (GET, POST, PUT, DELETE, etc.)
- `url [URL]` - Set base URL
- `path [PATH]` - Set request path
- `header [NAME] [VALUE]` - Add a header
- `header remove [NAME]` - Remove a header
- `header clear` - Clear all headers
- `query [NAME] [VALUE]` - Add a query parameter
- `query remove [NAME]` - Remove a query parameter
- `query clear` - Clear all query parameters
- `body [NAME] [VALUE]` - Add a body parameter
- `body remove [NAME]` - Remove a body parameter
- `body clear` - Clear all body parameters
- `body json [JSON_STRING]` - Set body from JSON string
- `format [FORMAT]` - Set body format (json, xml, form)
- `show` - Show current request
- `send` - Send the request
- `diff` - Compare the last two API responses
- `map [FORMAT]` - Generate API structure visualization (formats: json, html, dot, mermaid)
- `help` - Show help

### Example API Console Session

```
> api start
Entering API console mode

> method GET
HTTP method set to: GET

> url https://api.example.com
Base URL set to: https://api.example.com

> path /v1/users
Path set to: /v1/users

> header Authorization Bearer YOUR_TOKEN
Added header: Authorization: Bearer YOUR_TOKEN

> send
Sending GET request to https://api.example.com/v1/users

Response:
  Status: 200
  Content-Type: application/json
  Content-Length: 1234
  Response Time: 42ms

  Body:
  {
    "users": [
      {"id": 1, "name": "John Doe"},
      {"id": 2, "name": "Jane Smith"}
    ]
  }

> map html
API Map Visualization (html format):

<!DOCTYPE html>
<html>
...
</html>

To view the HTML visualization, save it to a file and open in a browser.
```

## API Response Analysis

### JSON Response Parsing

ffuf can parse JSON responses and extract specific data using JSONPath:

```bash
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -jr "$.users[*].name"
```

### Response Diffing

To compare API responses, use the `diff` command in the interactive API console:

```
> send
Sending GET request to https://api.example.com/v1/users/1

> method POST
HTTP method set to: POST

> body name Updated User
Added body parameter: name=Updated User

> send
Sending POST request to https://api.example.com/v1/users/1

> diff
Comparing previous two responses:

Status Code: DIFFERENT

Headers Differences:
  * Content-Length: changed from '[123]' to '[145]'

Content-Type: SAME

JSON Body Differences:
  * name: changed from 'John Doe' to 'Updated User'
  * updated_at: changed from '2023-01-01T00:00:00Z' to '2023-06-15T12:34:56Z'

Timing Difference: 5ms
```

### API Mapping Visualization

To visualize the structure of an API, use the `map` command in the interactive API console:

```
> map html
API Map Visualization (html format):

<!DOCTYPE html>
<html>
...
</html>

To view the HTML visualization, save it to a file and open in a browser.
```

## API Security Testing

ffuf can be used for API security testing, including:

- Authentication and authorization testing
- Input validation testing
- Rate limiting testing
- Injection testing
- Sensitive data exposure testing

### Testing for BOLA/IDOR Vulnerabilities

```bash
ffuf -u https://api.example.com/v1/users/FUZZ/profile -w /path/to/ids.txt -H "Authorization: Bearer YOUR_TOKEN"
```

### Testing for Injection Vulnerabilities

```bash
ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZ"}' -w /path/to/injections.txt
```

### Testing for Rate Limiting

```bash
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -rate 100
```

## Troubleshooting

### Common Issues

- **Authentication Failures**: Ensure you're using the correct authentication method and credentials.
- **Rate Limiting**: If you're being rate limited, reduce the request rate using the `-rate` flag.
- **JSON Parsing Errors**: Verify that your JSON payloads are correctly formatted.
- **Connection Issues**: Check network connectivity and proxy settings.

### Debugging

Use the `-v` flag for verbose output:

```bash
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -v
```

For more detailed debugging, use the `-debug-log` flag:

```bash
ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -debug-log debug.log
```

## Conclusion

ffuf is a powerful tool for API testing, offering a wide range of features from basic fuzzing to advanced API analysis. By following this guide, you should be able to effectively use ffuf for your API testing needs.

For more specific examples and workflows, refer to the [API Testing Workflows](workflows.md) document.