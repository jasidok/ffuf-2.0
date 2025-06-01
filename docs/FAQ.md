# Frequently Asked Questions (FAQ)

## General Usage

### Q: What does "FUZZ" mean and how do I use it?

**A:** FUZZ is a keyword placeholder that gets replaced with entries from your wordlist. You can use it anywhere in your
request:

```bash
# In URL path
ffuf -w wordlist.txt -u https://example.com/FUZZ

# In headers
ffuf -w wordlist.txt -u https://example.com/ -H "Host: FUZZ.example.com"

# In POST data
ffuf -w wordlist.txt -u https://example.com/login -d "username=FUZZ&password=test"

# Multiple FUZZ keywords
ffuf -w users.txt:USER -w passwords.txt:PASS -u https://example.com/login -d "username=USER&password=PASS"
```

### Q: Why am I getting too many results?

**A:** You likely need to add filters. Common approaches:

```bash
# Filter out 404 responses
ffuf -w wordlist.txt -u https://example.com/FUZZ -fc 404

# Filter by response size (if all 404s are same size)
ffuf -w wordlist.txt -u https://example.com/FUZZ -fs 4242

# Filter by content using regex
ffuf -w wordlist.txt -u https://example.com/FUZZ -fr "not found|error"

# Start with no filters to see what you're getting
ffuf -w wordlist.txt -u https://example.com/FUZZ -mc all -v | head -20
```

### Q: How do I find the right filter values?

**A:** Run a few test requests first:

```bash
# See all responses with details
ffuf -w small_wordlist.txt -u https://example.com/FUZZ -mc all -v

# Use auto-calibration
ffuf -w wordlist.txt -u https://example.com/FUZZ -ac

# Check a few manual requests
curl -I https://example.com/nonexistent
curl -I https://example.com/admin
```

### Q: Why is ffuf so slow?

**A:** Several factors can affect speed:

```bash
# Reduce threads if target can't handle load
ffuf -w wordlist.txt -u https://example.com/FUZZ -t 10

# Add rate limiting
ffuf -w wordlist.txt -u https://example.com/FUZZ -rate 50

# Add delays between requests
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 0.1

# Increase timeout if responses are slow
ffuf -w wordlist.txt -u https://example.com/FUZZ -timeout 30
```

## API Testing

### Q: What's the difference between regular mode and API mode?

**A:** API mode (`-api-mode`) provides several enhancements for API testing:

- Uses built-in API endpoint wordlists
- Enables JSON response parsing
- Applies API-appropriate default filters
- Optimizes output for API responses

```bash
# Regular mode
ffuf -w api_endpoints.txt -u https://api.example.com/FUZZ

# API mode (recommended for APIs)
ffuf -api-mode -u https://api.example.com/FUZZ
```

### Q: How do I test APIs that require authentication?

**A:** ffuf supports various authentication methods:

```bash
# Bearer token
ffuf -w endpoints.txt -u https://api.example.com/FUZZ -H "Authorization: Bearer YOUR_TOKEN"

# API key in header
ffuf -w endpoints.txt -u https://api.example.com/FUZZ -H "X-API-Key: YOUR_KEY"

# API key in query parameter
ffuf -w endpoints.txt -u https://api.example.com/FUZZ?api_key=YOUR_KEY

# Basic authentication
ffuf -w endpoints.txt -u https://user:pass@api.example.com/FUZZ

# Custom authentication headers
ffuf -w endpoints.txt -u https://api.example.com/FUZZ -H "Custom-Auth: value"
```

### Q: How do I fuzz JSON data?

**A:** Use POST with JSON content type:

```bash
# Fuzz JSON parameter names
ffuf -w params.txt -u https://api.example.com/endpoint -X POST \
     -H "Content-Type: application/json" \
     -d '{"FUZZ":"value"}' -mc all -fr "error"

# Fuzz JSON values
ffuf -w payloads.txt -u https://api.example.com/endpoint -X POST \
     -H "Content-Type: application/json" \
     -d '{"username":"FUZZ"}' -mc all -fr "error"

# Complex JSON structures
ffuf -w payloads.txt -u https://api.example.com/endpoint -X POST \
     -H "Content-Type: application/json" \
     -d '{"user":{"name":"FUZZ","role":"user"}}' -mc all -fr "error"
```

### Q: How do I test GraphQL APIs?

**A:** GraphQL testing requires POST requests with JSON payloads:

```bash
# Fuzz GraphQL queries
ffuf -w graphql_queries.txt -u https://api.example.com/graphql -X POST \
     -H "Content-Type: application/json" \
     -d '{"query":"FUZZ"}' -mc all -fr "errors"

# Fuzz GraphQL fields
ffuf -w fields.txt -u https://api.example.com/graphql -X POST \
     -H "Content-Type: application/json" \
     -d '{"query":"{ user { FUZZ } }"}' -mc all -fr "error"

# Fuzz GraphQL variables
ffuf -w values.txt -u https://api.example.com/graphql -X POST \
     -H "Content-Type: application/json" \
     -d '{"query":"query($var: String) { user(name: $var) { id } }","variables":{"var":"FUZZ"}}' \
     -mc all -fr "error"
```

## Configuration and Setup

### Q: Where should I put my configuration file?

**A:** ffuf looks for configuration files in several locations:

```bash
# Primary location (recommended)
~/.config/ffuf/ffufrc

# Alternative locations
~/.ffufrc
./ffufrc

# Custom location
ffuf -config /path/to/custom/ffufrc -w wordlist.txt -u https://example.com/FUZZ
```

### Q: What should I put in my configuration file?

**A:** Common configuration options:

```bash
# ~/.config/ffuf/ffufrc
threads = 40
timeout = 10
colors = true
verbose = false

# Default filters
fc = 404
fs = 4242

# Default headers
header = User-Agent: ffuf/2.1.0

# API testing defaults
header = Content-Type: application/json
rate = 50
```

### Q: How do I get better wordlists?

**A:** Several good sources:

- **SecLists**: https://github.com/danielmiessler/SecLists
- **FuzzDB**: https://github.com/fuzzdb-project/fuzzdb
- **PayloadsAllTheThings**: https://github.com/swisskyrepo/PayloadsAllTheThings
- **API Wordlists**: https://github.com/hAPI-hacker/Hacking-APIs (for API testing)

```bash
# Install SecLists (Ubuntu/Debian)
sudo apt install seclists

# Common locations after installation
/usr/share/seclists/
/usr/share/wordlists/
```

## Troubleshooting

### Q: I'm getting "connection refused" errors

**A:** Check connectivity and target availability:

```bash
# Test basic connectivity
curl -I https://example.com/

# Check if target is blocking your IP
curl -I https://example.com/ -H "User-Agent: Mozilla/5.0..."

# Try with different user agent
ffuf -w wordlist.txt -u https://example.com/FUZZ -H "User-Agent: Mozilla/5.0 (compatible)"

# Reduce concurrent connections
ffuf -w wordlist.txt -u https://example.com/FUZZ -t 1 -p 1
```

### Q: The target seems to be rate limiting me

**A:** Implement proper rate limiting:

```bash
# Reduce request rate
ffuf -w wordlist.txt -u https://example.com/FUZZ -rate 5

# Add delays between requests
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 2

# Use fewer threads
ffuf -w wordlist.txt -u https://example.com/FUZZ -t 5

# Random delays to appear more human-like
ffuf -w wordlist.txt -u https://example.com/FUZZ -p 1-3
```

### Q: How do I handle HTTPS/TLS errors?

**A:** Several options for TLS issues:

```bash
# Skip certificate verification (not recommended for production)
ffuf -w wordlist.txt -u https://example.com/FUZZ -k

# Specify custom CA certificate
ffuf -w wordlist.txt -u https://example.com/FUZZ -cc /path/to/ca.crt

# Use HTTP/2
ffuf -w wordlist.txt -u https://example.com/FUZZ -http2

# Set custom SNI
ffuf -w wordlist.txt -u https://example.com/FUZZ -sni custom.example.com
```

### Q: My wordlist isn't working

**A:** Common wordlist issues:

```bash
# Check wordlist format
head -10 /path/to/wordlist.txt

# Ensure wordlist is readable
ls -la /path/to/wordlist.txt

# Test with a small wordlist first
echo -e "admin\ntest\napi" > test_wordlist.txt
ffuf -w test_wordlist.txt -u https://example.com/FUZZ -v

# Check for encoding issues
file /path/to/wordlist.txt
```

## Advanced Usage

### Q: How do I use multiple wordlists?

**A:** Use different keywords for each wordlist:

```bash
# Two wordlists with different keywords
ffuf -w users.txt:USER -w passwords.txt:PASS \
     -u https://example.com/login \
     -d "username=USER&password=PASS" \
     -X POST -fc 401

# Different fuzzing modes
ffuf -w dirs.txt:DIR -w files.txt:FILE \
     -u https://example.com/DIR/FILE \
     -mode clusterbomb  # or pitchfork, sniper
```

### Q: How do I save and resume scans?

**A:** Use output files and filtering:

```bash
# Save results to file
ffuf -w wordlist.txt -u https://example.com/FUZZ -o results.json

# Save only successful results
ffuf -w wordlist.txt -u https://example.com/FUZZ -o results.json -or

# Use different output formats
ffuf -w wordlist.txt -u https://example.com/FUZZ -o results.html -of html

# Resume using interactive mode
# Press ENTER during scan, use 'show' command to see current results
```

### Q: How do I recursively scan directories?

**A:** Enable recursion:

```bash
# Basic recursion
ffuf -w wordlist.txt -u https://example.com/FUZZ -recursion

# Limit recursion depth
ffuf -w wordlist.txt -u https://example.com/FUZZ -recursion -recursion-depth 3

# Recursive strategy
ffuf -w wordlist.txt -u https://example.com/FUZZ -recursion -recursion-strategy greedy
```

### Q: How do I integrate ffuf with other tools?

**A:** Several integration options:

```bash
# Pipe results to other tools
ffuf -w wordlist.txt -u https://example.com/FUZZ -s | grep "200" | cut -d' ' -f1

# Use with Burp Suite
ffuf -w wordlist.txt -u https://example.com/FUZZ -replay-proxy http://127.0.0.1:8080

# Output for parsing
ffuf -w wordlist.txt -u https://example.com/FUZZ -json | jq '.results[].url'

# Use results in bash scripts
for url in $(ffuf -w wordlist.txt -u https://example.com/FUZZ -s | awk '{print $1}'); do
    echo "Testing: $url"
done
```

## Performance and Optimization

### Q: How can I make ffuf faster?

**A:** Several optimization strategies:

```bash
# Increase threads (if target can handle it)
ffuf -w wordlist.txt -u https://example.com/FUZZ -t 100

# Disable response body fetching if not needed
ffuf -w wordlist.txt -u https://example.com/FUZZ -ignore-body

# Use appropriate timeout
ffuf -w wordlist.txt -u https://example.com/FUZZ -timeout 5

# Skip unnecessary extensions
ffuf -w wordlist.txt -u https://example.com/FUZZ
# Instead of: ffuf -w wordlist.txt -u https://example.com/FUZZ -e .php,.html,.js
```

### Q: How do I limit the scan duration?

**A:** Use time limits:

```bash
# Maximum total runtime (5 minutes)
ffuf -w wordlist.txt -u https://example.com/FUZZ -maxtime 300

# Maximum time per job (useful with recursion)
ffuf -w wordlist.txt -u https://example.com/FUZZ -maxtime-job 60 -recursion
```

## Getting Help

### Q: Where can I get more help?

**A:** Several resources available:

- **Built-in help**: `ffuf -h`
- **Interactive help**: Press ENTER during execution, then type `help`
- **Documentation**: https://github.com/ffuf/ffuf/tree/master/docs
- **Wiki**: https://github.com/ffuf/ffuf/wiki
- **Community discussions**: https://github.com/ffuf/ffuf/discussions
- **Bug reports**: https://github.com/ffuf/ffuf/issues

### Q: How do I report bugs or request features?

**A:** Use GitHub:

1. **Search existing issues** first: https://github.com/ffuf/ffuf/issues
2. **For bugs**: Provide ffuf version, command used, expected vs actual behavior
3. **For features**: Describe the use case and proposed solution
4. **For questions**: Use GitHub Discussions instead of issues

### Q: How can I contribute?

**A:** Several ways to contribute:

- **Code**: Submit pull requests for bug fixes or features
- **Documentation**: Improve guides, examples, or translations
- **Testing**: Test new features and report issues
- **Community**: Help answer questions in discussions
- **Wordlists**: Contribute specialized wordlists for different use cases

See [CONTRIBUTORS.md](../CONTRIBUTORS.md) for detailed guidelines.