// Package security provides testing modules for API security vulnerabilities.
//
// This package implements testing modules for the OWASP API Security Top 10,
// including Broken Object Level Authorization, Broken Authentication, Excessive Data Exposure,
// Lack of Resources & Rate Limiting, Broken Function Level Authorization, Mass Assignment,
// Security Misconfiguration, Injection, Improper Assets Management, and Insufficient Logging & Monitoring.
package security

import (
	"context"
	"net/http"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// VulnerabilityType represents the type of vulnerability
type VulnerabilityType int

const (
	// VulnBrokenObjectLevelAuth represents Broken Object Level Authorization (API1:2019)
	VulnBrokenObjectLevelAuth VulnerabilityType = iota + 1
	// VulnBrokenAuth represents Broken User Authentication (API2:2019)
	VulnBrokenAuth
	// VulnExcessiveDataExposure represents Excessive Data Exposure (API3:2019)
	VulnExcessiveDataExposure
	// VulnLackOfResources represents Lack of Resources & Rate Limiting (API4:2019)
	VulnLackOfResources
	// VulnBrokenFunctionLevelAuth represents Broken Function Level Authorization (API5:2019)
	VulnBrokenFunctionLevelAuth
	// VulnMassAssignment represents Mass Assignment (API6:2019)
	VulnMassAssignment
	// VulnSecurityMisconfig represents Security Misconfiguration (API7:2019)
	VulnSecurityMisconfig
	// VulnInjection represents Injection (API8:2019)
	VulnInjection
	// VulnImproperAssetsMgmt represents Improper Assets Management (API9:2019)
	VulnImproperAssetsMgmt
	// VulnInsufficientLogging represents Insufficient Logging & Monitoring (API10:2019)
	VulnInsufficientLogging
)

// VulnerabilityInfo contains information about a detected vulnerability
type VulnerabilityInfo struct {
	Type        VulnerabilityType
	Name        string
	Description string
	Severity    string // "Critical", "High", "Medium", "Low", "Info"
	Request     *http.Request
	Response    *http.Response
	Evidence    string
	Remediation string
	CVSS        float64 // Common Vulnerability Scoring System score
	CWE         string  // Common Weakness Enumeration ID
	References  []string
	DetectedAt  time.Time
}

// TestResult represents the result of a security test
type TestResult struct {
	Vulnerabilities []VulnerabilityInfo
	TestName        string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Error           error
}

// SecurityTester is an interface for security testing modules
type SecurityTester interface {
	// Test runs the security test against the target
	Test(ctx context.Context, config *ffuf.Config) (*TestResult, error)
	// GetType returns the type of vulnerability this tester checks for
	GetType() VulnerabilityType
	// GetName returns the name of the security test
	GetName() string
	// GetDescription returns a description of the security test
	GetDescription() string
}

// SecurityTestRegistry holds registered security testers
type SecurityTestRegistry struct {
	testers map[VulnerabilityType]SecurityTester
}

// NewSecurityTestRegistry creates a new security test registry
func NewSecurityTestRegistry() *SecurityTestRegistry {
	return &SecurityTestRegistry{
		testers: make(map[VulnerabilityType]SecurityTester),
	}
}

// Register adds a security tester to the registry
func (r *SecurityTestRegistry) Register(tester SecurityTester) {
	r.testers[tester.GetType()] = tester
}

// Get retrieves a security tester from the registry
func (r *SecurityTestRegistry) Get(vulnType VulnerabilityType) (SecurityTester, bool) {
	tester, exists := r.testers[vulnType]
	return tester, exists
}

// GetAll returns all registered security testers
func (r *SecurityTestRegistry) GetAll() []SecurityTester {
	var testers []SecurityTester
	for _, tester := range r.testers {
		testers = append(testers, tester)
	}
	return testers
}

// RunAll runs all registered security tests
func (r *SecurityTestRegistry) RunAll(ctx context.Context, config *ffuf.Config) ([]*TestResult, error) {
	var results []*TestResult
	for _, tester := range r.testers {
		result, err := tester.Test(ctx, config)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// DefaultRegistry is the global security test registry
var DefaultRegistry = NewSecurityTestRegistry()

// RegisterSecurityTester registers a security tester with the default registry
func RegisterSecurityTester(tester SecurityTester) {
	DefaultRegistry.Register(tester)
}

// GetSecurityTester retrieves a security tester from the default registry
func GetSecurityTester(vulnType VulnerabilityType) (SecurityTester, bool) {
	return DefaultRegistry.Get(vulnType)
}

// GetAllSecurityTesters returns all registered security testers from the default registry
func GetAllSecurityTesters() []SecurityTester {
	return DefaultRegistry.GetAll()
}

// RunAllSecurityTests runs all registered security tests using the default registry
func RunAllSecurityTests(ctx context.Context, config *ffuf.Config) ([]*TestResult, error) {
	return DefaultRegistry.RunAll(ctx, config)
}