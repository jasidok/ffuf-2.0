# Getting Started with ffuf

This guide will help you get up and running with ffuf, from basic web fuzzing to advanced API testing.

## Installation

### Quick Install Options

```bash
# Download prebuilt binary (recommended)
wget https://github.com/ffuf/ffuf/releases/latest/download/ffuf_2.1.0_linux_amd64.tar.gz
tar -xzf ffuf_2.1.0_linux_amd64.tar.gz
sudo mv ffuf /usr/local/bin/

# Using Go (if you have Go installed)
go install github.com/ffuf/ffuf/v2@latest

# macOS with Homebrew
brew install ffuf

# Build from source
git clone https://github.com/ffuf/ffuf
cd ffuf
go build
```

### Verify Installation

```bash
ffuf -V
# Should output: ffuf version v2.1.0
```

## Basic Concepts

### The FUZZ Keyword

ffuf uses the `FUZZ` keyword as a placeholder that gets replaced with entries from your wordlist:

```bash
# FUZZ in URL path
ffuf -w wordlist.txt -u https://example.com/FUZZ

# FUZZ in headers
ffuf -w wordlist.txt -u https://example.com/ -H "Host: FUZZ.example.com"

# FUZZ in POST data
ffuf -w wordlist.txt -u https://example.com/login -d "username=FUZZ&password=test"
```

### Basic Filtering

Control what results are shown:

```bash
# Match specific status codes
ffuf -w wordlist.txt -u https://example.com/FUZZ -mc 200,301,302

# Filter out specific status codes
ffuf -w wordlist.txt -u https://example.com/FUZZ -fc 404

# Filter by response size
ffuf -w wordlist.txt -u https://example.com/FUZZ -fs 4242

# Match responses containing specific text
ffuf -w wordlist.txt -u https://example.com/FUZZ -mr "admin"
```

## Your First Scans

### 1. Directory Discovery

Find hidden directories and files:

```bash
# Basic directory discovery
ffuf -w /usr/share/wordlists/dirb/common.txt -u https://example.com/FUZZ

# With better output formatting
ffuf -w /usr/share/wordlists/dirb/common.txt -u https://example.com/FUZZ -c -v

# Filter out common 404 response size
ffuf -w /usr/share/wordlists/dirb/common.txt -u https://example.com/FUZZ -fs 4242 -c -v
```

### 2. Virtual Host Discovery

Find subdomains and virtual hosts:

```bash
# Subdomain enumeration
ffuf -w /usr/share/wordlists/seclists/Discovery/DNS/subdomains-top1million-5000.txt \
     -u https://example.com/ -H "Host: FUZZ.example.com" -fs 4242

# Virtual host discovery
ffuf -w vhosts.txt -u https://10.10.10.10/ -H "Host: FUZZ" -fs 4242
```

### 3. Parameter Discovery

Find hidden parameters:

```bash
# GET parameter names
ffuf -w params.txt -u https://example.com/search?FUZZ=test -fs 4242

# GET parameter values
ffuf -w values.txt -u https://example.com/search?id=FUZZ -fc 404

# POST parameter names
ffuf -w params.txt -u https://example.com/login -X POST -d "FUZZ=test" -fc 404
```

## API Testing Basics

ffuf excels at testing modern APIs. Here's how to get started:

### 1. Enable API Mode

```bash
# Basic API endpoint discovery
ffuf -api-mode -u https://api.example.com/FUZZ -H "Authorization: Bearer YOUR_TOKEN"
```

API mode automatically:

- Uses built-in API endpoint wordlists
- Applies appropriate filters for API responses
- Enables JSON response parsing

### 2. REST API Testing

```bash
# Test API endpoints
ffuf -w api_endpoints.txt -u https://api.example.com/FUZZ \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer TOKEN" \
     -mc all -fc 404

# Fuzz JSON parameters
ffuf -w params.txt -u https://api.example.com/users \
     -X POST -H "Content-Type: application/json" \
     -d '{"FUZZ":"value"}' -mc all -fr "error"

# Fuzz JSON values
ffuf -w payloads.txt -u https://api.example.com/users \
     -X POST -H "Content-Type: application/json" \
     -d '{"username":"FUZZ"}' -mc all -fr "error"
```

### 3. GraphQL API Testing

```bash
# GraphQL query fuzzing
ffuf -w graphql_queries.txt -u https://api.example.com/graphql \
     -X POST -H "Content-Type: application/json" \
     -d '{"query":"FUZZ"}' -mc all -fr "errors"

# GraphQL field fuzzing
ffuf -w fields.txt -u https://api.example.com/graphql \
     -X POST -H "Content-Type: application/json" \
     -d '{"query":"{ user { FUZZ } }"}' -mc all -fr "error"
```

## Advanced Features

### 1. Multiple Wordlists

Use multiple wordlists with different keywords:

```bash
# Use different keywords for different positions
ffuf -w users.txt:USER -w passwords.txt:PASS \
     -u https://example.com/login \
     -X POST -d "username=USER&password=PASS" \
     -fc 401
```

### 2. Recursive Scanning

Automatically scan discovered directories:

```bash
# Enable recursion
ffuf -w wordlist.txt -u https://example.com/FUZZ -recursion -recursion-depth 2

# Recursive with API mode
ffuf -api-mode -u https://api.example.com/FUZZ -recursion -recursion-depth 1
```

### 3. Rate Limiting

Control request rate to avoid overwhelming targets:

```bash
# 10 requests per second
ffuf -w wordlist.txt -u https://example.com/FUZZ -rate 10

# Add delay between requests
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 0.1

# Random delay between 0.1 and 2 seconds
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 0.1-2.0
```

## Configuration Files

Save common settings in a configuration file:

```bash
# Create config file
mkdir -p ~/.config/ffuf
cat > ~/.config/ffuf/ffufrc << EOF
# Default ffuf configuration
threads = 100
timeout = 10
rate = 50
colors = true
verbose = true

# Default filters
fs = 4242
fc = 404

# Default headers for API testing
header = Content-Type: application/json
header = User-Agent: ffuf/2.1.0
EOF

# Use config
ffuf -w wordlist.txt -u https://example.com/FUZZ
```

## Common Patterns

### Web Application Testing

```bash
# Complete web app discovery
ffuf -w common.txt -u https://example.com/FUZZ -e .php,.html,.js,.txt -c -v

# Login form testing
ffuf -w passwords.txt -u https://example.com/login \
     -X POST -d "username=admin&password=FUZZ" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -fc 401
```

### API Security Testing

```bash
# Test for admin endpoints
ffuf -w admin_endpoints.txt -u https://api.example.com/FUZZ \
     -H "Authorization: Bearer USER_TOKEN" \
     -mc all -fc 403

# Test for IDOR vulnerabilities
ffuf -w numbers.txt -u https://api.example.com/users/FUZZ \
     -H "Authorization: Bearer TOKEN" \
     -mc 200

# Test for injection vulnerabilities
ffuf -w sql_payloads.txt -u https://api.example.com/search \
     -X POST -H "Content-Type: application/json" \
     -d '{"query":"FUZZ"}' -mr "error|exception|syntax"
```

## Next Steps

Once you're comfortable with the basics:

1. **Read the [API Testing Guide](api/guide.md)** for comprehensive API testing techniques
2. **Explore [API Workflows](api/workflows.md)** for real-world scenarios
3. **Use the [API Cheat Sheet](api/cheatsheet.md)** for quick reference
4. **Learn about [Tool Integration](api/integration.md)** for advanced workflows

## Getting Help

- **Built-in help**: `ffuf -h`
- **Interactive mode**: Press ENTER during execution
- **Documentation**: https://github.com/ffuf/ffuf/tree/master/docs
- **Community**: https://github.com/ffuf/ffuf/discussions
- **Issues**: https://github.com/ffuf/ffuf/issues

## Common Troubleshooting

### No Results Found

```bash
# Check if target is responding
curl -I https://example.com/

# Try with no filters first
ffuf -w wordlist.txt -u https://example.com/FUZZ -mc all

# Check response sizes
ffuf -w wordlist.txt -u https://example.com/FUZZ -mc all -v | head -20
```

### Too Many Results

```bash
# Add filters based on common response size
ffuf -w wordlist.txt -u https://example.com/FUZZ -fs 4242

# Filter by status code
ffuf -w wordlist.txt -u https://example.com/FUZZ -fc 404

# Use regex to filter content
ffuf -w wordlist.txt -u https://example.com/FUZZ -fr "not found|error"
```

### Rate Limiting

```bash
# Reduce concurrent threads
ffuf -w wordlist.txt -u https://example.com/FUZZ -t 10

# Add delays
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 1

# Use rate limiting
ffuf -w wordlist.txt -u https://example.com/FUZZ -rate 5
```