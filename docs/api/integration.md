# Integrating ffuf with Other API Testing Tools

This document provides guidance on how to integrate ffuf with other popular API testing and security tools to create comprehensive API testing workflows.

## Table of Contents

1. [Postman Integration](#postman-integration)
2. [Swagger/OpenAPI Integration](#swaggeropenapi-integration)
3. [Burp Suite Integration](#burp-suite-integration)
4. [OWASP ZAP Integration](#owasp-zap-integration)
5. [CI/CD Integration](#cicd-integration)
6. [Custom Tool Integration](#custom-tool-integration)

## Postman Integration

[Postman](https://www.postman.com/) is a popular API client that allows you to design, test, and document APIs. ffuf can be integrated with Postman in several ways:

### Importing Postman Collections

ffuf can import Postman collections to use as a source of API endpoints and test cases:

```bash
# Export your collection from Postman as a JSON file
# Then use ffuf to import and test the endpoints
ffuf -u https://api.example.com/FUZZ -w postman_collection_endpoints.txt
```

To extract endpoints from a Postman collection:

```bash
# Using jq to extract endpoints from a Postman collection
jq -r '.item[].request.url.path | join("/")' postman_collection.json > postman_endpoints.txt

# Then use with ffuf
ffuf -u https://api.example.com/FUZZ -w postman_endpoints.txt
```

### Exporting ffuf Results to Postman

You can export ffuf results in a format that can be imported into Postman:

```bash
# Run ffuf with JSON output
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -of json -o results.json

# Convert the results to Postman collection format using a script
python convert_ffuf_to_postman.py results.json > postman_collection.json
```

Example conversion script (`convert_ffuf_to_postman.py`):

```python
import json
import sys
import uuid

def convert_ffuf_to_postman(ffuf_file):
    with open(ffuf_file, 'r') as f:
        ffuf_data = json.load(f)
    
    postman_collection = {
        "info": {
            "name": "ffuf API Results",
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
        },
        "item": []
    }
    
    for result in ffuf_data["results"]:
        item = {
            "name": result["url"],
            "request": {
                "method": "GET",
                "header": [],
                "url": {
                    "raw": result["url"],
                    "protocol": result["url"].split("://")[0],
                    "host": result["url"].split("://")[1].split("/")[0].split("."),
                    "path": result["url"].split("://")[1].split("/")[1:]
                }
            },
            "response": []
        }
        postman_collection["item"].append(item)
    
    return postman_collection

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python convert_ffuf_to_postman.py <ffuf_results.json>")
        sys.exit(1)
    
    postman_collection = convert_ffuf_to_postman(sys.argv[1])
    print(json.dumps(postman_collection, indent=2))
```

## Swagger/OpenAPI Integration

[Swagger/OpenAPI](https://swagger.io/) is a specification for describing RESTful APIs. ffuf can be integrated with Swagger/OpenAPI to test APIs based on their specifications.

### Testing APIs Based on OpenAPI Specifications

```bash
# First, download the OpenAPI specification
curl -o openapi.json https://api.example.com/openapi.json

# Extract endpoints from the OpenAPI specification
jq -r '.paths | keys[]' openapi.json > openapi_endpoints.txt

# Use ffuf to test the endpoints
ffuf -u https://api.example.com/FUZZ -w openapi_endpoints.txt
```

### Validating API Responses Against OpenAPI Specifications

You can use ffuf to test if API responses conform to their OpenAPI specifications:

```bash
# Run ffuf with output to a file
ffuf -u https://api.example.com/FUZZ -w openapi_endpoints.txt -of json -o results.json

# Use a script to validate the responses against the OpenAPI specification
python validate_responses.py results.json openapi.json
```

## Burp Suite Integration

[Burp Suite](https://portswigger.net/burp) is a popular web application security testing tool. ffuf can be integrated with Burp Suite to enhance API security testing.

### Using Burp Suite as a Proxy for ffuf

```bash
# Run ffuf through Burp Suite proxy
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -x http://127.0.0.1:8080
```

### Sending ffuf Results to Burp Suite for Further Analysis

```bash
# Run ffuf with replay proxy to send results to Burp Suite
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -replay-proxy http://127.0.0.1:8080
```

### Using Burp Suite Extensions with ffuf

The [Burp to ffuf](https://github.com/d3k4z/burp2ffuf) extension allows you to convert Burp Suite requests to ffuf commands:

1. Install the Burp to ffuf extension in Burp Suite
2. Right-click on a request in Burp Suite and select "Copy as ffuf command"
3. Paste and run the command in your terminal

## OWASP ZAP Integration

[OWASP ZAP](https://www.zaproxy.org/) (Zed Attack Proxy) is an open-source web application security scanner. ffuf can be integrated with ZAP for comprehensive API security testing.

### Using ZAP as a Proxy for ffuf

```bash
# Run ffuf through ZAP proxy
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -x http://127.0.0.1:8080
```

### Automating ZAP Scans with ffuf Results

```bash
# Run ffuf and output results to a file
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -of json -o results.json

# Use ZAP's API to scan the discovered endpoints
python zap_scan.py results.json
```

Example ZAP integration script (`zap_scan.py`):

```python
import json
import sys
import time
from zapv2 import ZAPv2

def scan_with_zap(ffuf_results_file):
    # Load ffuf results
    with open(ffuf_results_file, 'r') as f:
        ffuf_data = json.load(f)
    
    # Connect to ZAP
    zap = ZAPv2(apikey='your-api-key', proxies={'http': 'http://127.0.0.1:8080', 'https': 'http://127.0.0.1:8080'})
    
    # Scan each discovered endpoint
    for result in ffuf_data["results"]:
        url = result["url"]
        print(f"Scanning {url} with ZAP...")
        
        # Access the URL through ZAP
        zap.urlopen(url)
        
        # Spider the URL
        scan_id = zap.spider.scan(url)
        while int(zap.spider.status(scan_id)) < 100:
            print(f"Spider progress: {zap.spider.status(scan_id)}%")
            time.sleep(1)
        
        # Active scan
        scan_id = zap.ascan.scan(url)
        while int(zap.ascan.status(scan_id)) < 100:
            print(f"Active scan progress: {zap.ascan.status(scan_id)}%")
            time.sleep(5)
        
        # Get alerts
        alerts = zap.core.alerts(url)
        print(f"Found {len(alerts)} alerts for {url}")
        for alert in alerts:
            print(f"- {alert['alert']} ({alert['risk']}): {alert['url']}")
    
    print("ZAP scanning complete!")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python zap_scan.py <ffuf_results.json>")
        sys.exit(1)
    
    scan_with_zap(sys.argv[1])
```

## CI/CD Integration

ffuf can be integrated into CI/CD pipelines for automated API testing.

### GitHub Actions Integration

Example GitHub Actions workflow file (`.github/workflows/api-testing.yml`):

```yaml
name: API Testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * *'  # Run daily at midnight

jobs:
  ffuf-api-testing:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.17
      
      - name: Install ffuf
        run: go install github.com/ffuf/ffuf/v2@latest
      
      - name: Run API endpoint discovery
        run: |
          ffuf -u https://api.example.com/FUZZ -w ./wordlists/api_endpoints.txt -o results.json -of json
      
      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: ffuf-results
          path: results.json
```

### Jenkins Integration

Example Jenkins pipeline script:

```groovy
pipeline {
    agent any
    
    stages {
        stage('Install ffuf') {
            steps {
                sh 'go install github.com/ffuf/ffuf/v2@latest'
            }
        }
        
        stage('API Testing') {
            steps {
                sh 'ffuf -u https://api.example.com/FUZZ -w ./wordlists/api_endpoints.txt -o results.json -of json'
            }
        }
        
        stage('Process Results') {
            steps {
                sh 'python process_results.py results.json'
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'results.json', fingerprint: true
        }
    }
}
```

## Custom Tool Integration

ffuf can be integrated with custom tools and scripts to create tailored API testing workflows.

### Using ffuf with Custom Scripts

```bash
# Run ffuf and pipe results to a custom script
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -json | python process_results.py

# Or save results to a file and process them
ffuf -u https://api.example.com/FUZZ -w /path/to/wordlist.txt -of json -o results.json
python process_results.py results.json
```

Example processing script (`process_results.py`):

```python
import json
import sys

def process_results(results_file):
    with open(results_file, 'r') as f:
        data = json.load(f)
    
    # Process the results
    endpoints = []
    for result in data["results"]:
        if result["status"] == 200:
            endpoints.append(result["url"])
    
    # Output processed results
    print(f"Found {len(endpoints)} valid endpoints:")
    for endpoint in endpoints:
        print(f"- {endpoint}")
    
    # Save processed results to a new file
    with open('processed_results.json', 'w') as f:
        json.dump({"valid_endpoints": endpoints}, f, indent=2)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python process_results.py <ffuf_results.json>")
        sys.exit(1)
    
    process_results(sys.argv[1])
```

### Creating a Custom API Testing Framework

You can create a custom API testing framework that uses ffuf as its core engine:

```python
import subprocess
import json
import os

class APITestingFramework:
    def __init__(self, base_url, wordlist_dir):
        self.base_url = base_url
        self.wordlist_dir = wordlist_dir
        self.results = []
    
    def discover_endpoints(self):
        cmd = [
            "ffuf",
            "-u", f"{self.base_url}/FUZZ",
            "-w", f"{self.wordlist_dir}/api_endpoints.txt",
            "-of", "json",
            "-o", "endpoints.json"
        ]
        subprocess.run(cmd)
        
        with open("endpoints.json", "r") as f:
            data = json.load(f)
        
        endpoints = [result["url"] for result in data["results"]]
        return endpoints
    
    def test_endpoint(self, endpoint, method="GET", data=None, headers=None):
        cmd = ["ffuf", "-u", endpoint, "-X", method]
        
        if headers:
            for key, value in headers.items():
                cmd.extend(["-H", f"{key}: {value}"])
        
        if data:
            cmd.extend(["-d", json.dumps(data)])
        
        cmd.extend(["-of", "json", "-o", "endpoint_test.json"])
        
        subprocess.run(cmd)
        
        with open("endpoint_test.json", "r") as f:
            result = json.load(f)
        
        return result
    
    def run_security_tests(self, endpoint):
        # Run various security tests on the endpoint
        tests = [
            {
                "name": "SQL Injection",
                "wordlist": f"{self.wordlist_dir}/sql_injections.txt",
                "param": "query"
            },
            {
                "name": "XSS",
                "wordlist": f"{self.wordlist_dir}/xss_payloads.txt",
                "param": "input"
            }
        ]
        
        results = {}
        
        for test in tests:
            cmd = [
                "ffuf",
                "-u", f"{endpoint}?{test['param']}=FUZZ",
                "-w", test["wordlist"],
                "-of", "json",
                "-o", f"{test['name']}_test.json"
            ]
            subprocess.run(cmd)
            
            with open(f"{test['name']}_test.json", "r") as f:
                results[test["name"]] = json.load(f)
        
        return results
    
    def generate_report(self, results):
        # Generate a comprehensive report from all test results
        report = {
            "base_url": self.base_url,
            "endpoints_tested": len(results),
            "vulnerabilities_found": sum(1 for r in results.values() if r.get("vulnerable", False)),
            "details": results
        }
        
        with open("api_test_report.json", "w") as f:
            json.dump(report, f, indent=2)
        
        print(f"API Testing Report generated: api_test_report.json")
        print(f"Endpoints tested: {report['endpoints_tested']}")
        print(f"Vulnerabilities found: {report['vulnerabilities_found']}")

# Example usage
if __name__ == "__main__":
    framework = APITestingFramework("https://api.example.com", "./wordlists")
    
    # Discover endpoints
    endpoints = framework.discover_endpoints()
    print(f"Discovered {len(endpoints)} endpoints")
    
    # Test each endpoint
    results = {}
    for endpoint in endpoints:
        print(f"Testing endpoint: {endpoint}")
        results[endpoint] = {
            "basic_test": framework.test_endpoint(endpoint),
            "security_tests": framework.run_security_tests(endpoint)
        }
    
    # Generate report
    framework.generate_report(results)
```

## Conclusion

ffuf's flexibility and command-line interface make it an excellent tool for integration with other API testing and security tools. By combining ffuf with other specialized tools, you can create comprehensive API testing workflows that cover discovery, functional testing, security testing, and performance testing.

For more information on using ffuf for API testing, refer to the [API Testing Guide](guide.md) and [API Testing Workflows](workflows.md).