// Package reporting provides functionality for generating reports on API testing activities.
//
// This package includes modules for analyzing API coverage, generating reports in various
// formats, and visualizing testing progress and results.
package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/api/parser"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// CoverageFormat represents the format of a coverage report
type CoverageFormat string

const (
	// FormatJSON represents a JSON report format
	FormatJSON CoverageFormat = "json"
	// FormatHTML represents an HTML report format
	FormatHTML CoverageFormat = "html"
	// FormatMarkdown represents a Markdown report format
	FormatMarkdown CoverageFormat = "md"
	// FormatText represents a plain text report format
	FormatText CoverageFormat = "txt"
)

// EndpointStatus represents the testing status of an API endpoint
type EndpointStatus string

const (
	// StatusUntested represents an endpoint that has not been tested
	StatusUntested EndpointStatus = "untested"
	// StatusTested represents an endpoint that has been tested
	StatusTested EndpointStatus = "tested"
	// StatusPartial represents an endpoint that has been partially tested (e.g., not all parameters)
	StatusPartial EndpointStatus = "partial"
	// StatusError represents an endpoint that encountered errors during testing
	StatusError EndpointStatus = "error"
)

// CoverageOptions contains configuration options for the coverage analyzer
type CoverageOptions struct {
	// IncludeUntested determines whether untested endpoints are included in the report
	IncludeUntested bool
	// DetailLevel controls the amount of detail in the report (1-3)
	DetailLevel int
	// GroupByTags determines whether endpoints are grouped by tags in the report
	GroupByTags bool
	// Format specifies the report format
	Format CoverageFormat
	// OutputFile specifies the file to write the report to (empty for stdout)
	OutputFile string
}

// DefaultCoverageOptions returns the default coverage options
func DefaultCoverageOptions() *CoverageOptions {
	return &CoverageOptions{
		IncludeUntested: true,
		DetailLevel:     2,
		GroupByTags:     true,
		Format:          FormatHTML,
		OutputFile:      "",
	}
}

// EndpointCoverage represents coverage information for a single API endpoint
type EndpointCoverage struct {
	// Path is the endpoint path
	Path string `json:"path"`
	// Method is the HTTP method
	Method string `json:"method"`
	// Status is the testing status
	Status EndpointStatus `json:"status"`
	// Tags are the endpoint tags
	Tags []string `json:"tags,omitempty"`
	// Parameters are the endpoint parameters
	Parameters []ParameterCoverage `json:"parameters,omitempty"`
	// ResponseStatus is the HTTP status code of the last response
	ResponseStatus int `json:"response_status,omitempty"`
	// LastTested is the timestamp of the last test
	LastTested time.Time `json:"last_tested,omitempty"`
	// TestCount is the number of times the endpoint has been tested
	TestCount int `json:"test_count"`
	// ErrorCount is the number of errors encountered during testing
	ErrorCount int `json:"error_count"`
}

// ParameterCoverage represents coverage information for a single API parameter
type ParameterCoverage struct {
	// Name is the parameter name
	Name string `json:"name"`
	// Type is the parameter type
	Type string `json:"type,omitempty"`
	// Required indicates whether the parameter is required
	Required bool `json:"required,omitempty"`
	// Tested indicates whether the parameter has been tested
	Tested bool `json:"tested"`
	// TestCount is the number of times the parameter has been tested
	TestCount int `json:"test_count"`
}

// CoverageAnalyzer tracks and analyzes API testing coverage
type CoverageAnalyzer struct {
	// options contains the configuration options
	options *CoverageOptions
	// endpoints maps endpoint paths to coverage information
	endpoints map[string]*EndpointCoverage
	// discovery is the API endpoint discovery instance
	discovery *parser.APIEndpointDiscovery
	// visualizer is used for generating visualizations
	visualizer *parser.Visualizer
	// startTime is the time when the analyzer was created
	startTime time.Time
}

// NewCoverageAnalyzer creates a new coverage analyzer with the given options
func NewCoverageAnalyzer(options *CoverageOptions) *CoverageAnalyzer {
	if options == nil {
		options = DefaultCoverageOptions()
	}

	return &CoverageAnalyzer{
		options:    options,
		endpoints:  make(map[string]*EndpointCoverage),
		visualizer: parser.NewVisualizer(nil),
		startTime:  time.Now(),
	}
}

// ImportFromDiscovery imports endpoints from an API endpoint discovery instance
func (c *CoverageAnalyzer) ImportFromDiscovery(discovery *parser.APIEndpointDiscovery) {
	c.discovery = discovery
	
	// Import all discovered endpoints
	for _, endpoint := range discovery.GetEndpoints() {
		key := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		
		// Create coverage entry if it doesn't exist
		if _, exists := c.endpoints[key]; !exists {
			params := make([]ParameterCoverage, 0, len(endpoint.Parameters))
			for _, param := range endpoint.Parameters {
				params = append(params, ParameterCoverage{
					Name:      param.Name,
					Type:      param.Type,
					Required:  param.Required,
					Tested:    false,
					TestCount: 0,
				})
			}
			
			c.endpoints[key] = &EndpointCoverage{
				Path:       endpoint.Path,
				Method:     endpoint.Method,
				Status:     StatusUntested,
				Tags:       endpoint.Tags,
				Parameters: params,
				TestCount:  0,
				ErrorCount: 0,
			}
		}
	}
}

// RecordTest records a test of an API endpoint
func (c *CoverageAnalyzer) RecordTest(method, path string, resp *ffuf.Response, testedParams []string) {
	key := fmt.Sprintf("%s %s", method, path)
	
	// Create the endpoint if it doesn't exist
	if _, exists := c.endpoints[key]; !exists {
		c.endpoints[key] = &EndpointCoverage{
			Path:       path,
			Method:     method,
			Status:     StatusUntested,
			Parameters: []ParameterCoverage{},
			TestCount:  0,
			ErrorCount: 0,
		}
	}
	
	endpoint := c.endpoints[key]
	endpoint.TestCount++
	endpoint.LastTested = time.Now()
	
	// Record response status if available
	if resp != nil {
		endpoint.ResponseStatus = int(resp.StatusCode)
		
		// Record error if status code indicates an error
		if resp.StatusCode >= 400 {
			endpoint.ErrorCount++
		}
	}
	
	// Update parameter testing status
	paramsTested := 0
	for i, param := range endpoint.Parameters {
		for _, testedParam := range testedParams {
			if param.Name == testedParam {
				endpoint.Parameters[i].Tested = true
				endpoint.Parameters[i].TestCount++
				paramsTested++
				break
			}
		}
	}
	
	// Update endpoint status
	if paramsTested == 0 && len(endpoint.Parameters) > 0 {
		endpoint.Status = StatusPartial
	} else if paramsTested == len(endpoint.Parameters) || len(endpoint.Parameters) == 0 {
		endpoint.Status = StatusTested
	} else {
		endpoint.Status = StatusPartial
	}
	
	// If there were errors, mark as error status
	if endpoint.ErrorCount > 0 {
		endpoint.Status = StatusError
	}
}

// GetCoverageStats returns overall coverage statistics
func (c *CoverageAnalyzer) GetCoverageStats() map[string]interface{} {
	totalEndpoints := len(c.endpoints)
	testedEndpoints := 0
	partialEndpoints := 0
	errorEndpoints := 0
	totalParams := 0
	testedParams := 0
	
	for _, endpoint := range c.endpoints {
		if endpoint.Status == StatusTested {
			testedEndpoints++
		} else if endpoint.Status == StatusPartial {
			partialEndpoints++
		} else if endpoint.Status == StatusError {
			errorEndpoints++
		}
		
		for _, param := range endpoint.Parameters {
			totalParams++
			if param.Tested {
				testedParams++
			}
		}
	}
	
	// Calculate percentages
	endpointCoverage := 0.0
	if totalEndpoints > 0 {
		endpointCoverage = float64(testedEndpoints+partialEndpoints) / float64(totalEndpoints) * 100
	}
	
	paramCoverage := 0.0
	if totalParams > 0 {
		paramCoverage = float64(testedParams) / float64(totalParams) * 100
	}
	
	return map[string]interface{}{
		"total_endpoints":     totalEndpoints,
		"tested_endpoints":    testedEndpoints,
		"partial_endpoints":   partialEndpoints,
		"error_endpoints":     errorEndpoints,
		"untested_endpoints":  totalEndpoints - testedEndpoints - partialEndpoints - errorEndpoints,
		"endpoint_coverage":   endpointCoverage,
		"total_parameters":    totalParams,
		"tested_parameters":   testedParams,
		"parameter_coverage":  paramCoverage,
		"duration":            time.Since(c.startTime).String(),
		"timestamp":           time.Now().Format(time.RFC3339),
	}
}

// GenerateReport generates a coverage report in the specified format
func (c *CoverageAnalyzer) GenerateReport() (string, error) {
	switch c.options.Format {
	case FormatJSON:
		return c.generateJSONReport()
	case FormatHTML:
		return c.generateHTMLReport()
	case FormatMarkdown:
		return c.generateMarkdownReport()
	case FormatText:
		return c.generateTextReport()
	default:
		return "", api.NewAPIError("Unsupported report format", 0)
	}
}

// generateJSONReport generates a JSON coverage report
func (c *CoverageAnalyzer) generateJSONReport() (string, error) {
	// Prepare the report data
	report := map[string]interface{}{
		"stats":     c.GetCoverageStats(),
		"endpoints": c.getEndpointsForReport(),
	}
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", api.NewAPIError("Failed to generate JSON report: "+err.Error(), 0)
	}
	
	return string(jsonData), nil
}

// generateHTMLReport generates an HTML coverage report
func (c *CoverageAnalyzer) generateHTMLReport() (string, error) {
	// HTML template for the report
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>API Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2, h3 { color: #333; }
        .stats { display: flex; flex-wrap: wrap; margin-bottom: 20px; }
        .stat-box { 
            background-color: #f5f5f5; 
            border-radius: 5px; 
            padding: 15px; 
            margin: 10px; 
            min-width: 200px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-value { font-size: 24px; font-weight: bold; margin-bottom: 5px; }
        .stat-label { color: #666; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        tr:hover { background-color: #f5f5f5; }
        .status-tested { color: green; }
        .status-partial { color: orange; }
        .status-error { color: red; }
        .status-untested { color: gray; }
        .progress-bar {
            height: 20px;
            background-color: #e0e0e0;
            border-radius: 10px;
            margin-top: 5px;
        }
        .progress-value {
            height: 100%;
            background-color: #4CAF50;
            border-radius: 10px;
        }
        .tag { 
            display: inline-block;
            background-color: #e0e0e0;
            padding: 3px 8px;
            border-radius: 10px;
            margin-right: 5px;
            font-size: 12px;
        }
        .parameters {
            margin-left: 20px;
            font-size: 14px;
        }
        .parameter-name {
            font-weight: bold;
        }
        .parameter-tested {
            color: green;
        }
        .parameter-untested {
            color: gray;
        }
    </style>
</head>
<body>
    <h1>API Coverage Report</h1>
    <p>Generated on {{.stats.timestamp}}</p>
    
    <div class="stats">
        <div class="stat-box">
            <div class="stat-value">{{.stats.endpoint_coverage | printf "%.1f"}}%</div>
            <div class="stat-label">Endpoint Coverage</div>
            <div class="progress-bar">
                <div class="progress-value" style="width: {{.stats.endpoint_coverage}}%;"></div>
            </div>
        </div>
        <div class="stat-box">
            <div class="stat-value">{{.stats.parameter_coverage | printf "%.1f"}}%</div>
            <div class="stat-label">Parameter Coverage</div>
            <div class="progress-bar">
                <div class="progress-value" style="width: {{.stats.parameter_coverage}}%;"></div>
            </div>
        </div>
        <div class="stat-box">
            <div class="stat-value">{{.stats.total_endpoints}}</div>
            <div class="stat-label">Total Endpoints</div>
        </div>
        <div class="stat-box">
            <div class="stat-value">{{.stats.tested_endpoints}}</div>
            <div class="stat-label">Fully Tested</div>
        </div>
        <div class="stat-box">
            <div class="stat-value">{{.stats.partial_endpoints}}</div>
            <div class="stat-label">Partially Tested</div>
        </div>
        <div class="stat-box">
            <div class="stat-value">{{.stats.error_endpoints}}</div>
            <div class="stat-label">Errors</div>
        </div>
    </div>
    
    <h2>Endpoint Details</h2>
    <table>
        <tr>
            <th>Method</th>
            <th>Path</th>
            <th>Status</th>
            <th>Tags</th>
            <th>Tests</th>
            <th>Last Tested</th>
        </tr>
        {{range .endpoints}}
        <tr>
            <td>{{.Method}}</td>
            <td>{{.Path}}</td>
            <td class="status-{{.Status}}">{{.Status}}</td>
            <td>
                {{range .Tags}}
                <span class="tag">{{.}}</span>
                {{end}}
            </td>
            <td>{{.TestCount}}</td>
            <td>{{if .LastTested}}{{.LastTested | formatTime}}{{else}}Never{{end}}</td>
        </tr>
        {{if gt (len .Parameters) 0}}
        <tr>
            <td colspan="6" class="parameters">
                <strong>Parameters:</strong>
                {{range .Parameters}}
                <div class="{{if .Tested}}parameter-tested{{else}}parameter-untested{{end}}">
                    <span class="parameter-name">{{.Name}}</span> 
                    ({{.Type}}{{if .Required}}, required{{end}}) - 
                    {{if .Tested}}Tested {{.TestCount}} times{{else}}Not tested{{end}}
                </div>
                {{end}}
            </td>
        </tr>
        {{end}}
        {{end}}
    </table>
    
    <p><small>Report generated by ffuf API Coverage Analyzer. Duration: {{.stats.duration}}</small></p>
</body>
</html>`

	// Create template with custom function
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
	}
	
	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", api.NewAPIError("Failed to parse HTML template: "+err.Error(), 0)
	}
	
	// Prepare the report data
	report := map[string]interface{}{
		"stats":     c.GetCoverageStats(),
		"endpoints": c.getEndpointsForReport(),
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		return "", api.NewAPIError("Failed to generate HTML report: "+err.Error(), 0)
	}
	
	return buf.String(), nil
}

// generateMarkdownReport generates a Markdown coverage report
func (c *CoverageAnalyzer) generateMarkdownReport() (string, error) {
	stats := c.GetCoverageStats()
	endpoints := c.getEndpointsForReport()
	
	var buf bytes.Buffer
	
	// Write header
	buf.WriteString("# API Coverage Report\n\n")
	buf.WriteString(fmt.Sprintf("Generated on %s\n\n", stats["timestamp"]))
	
	// Write summary
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- **Endpoint Coverage**: %.1f%%\n", stats["endpoint_coverage"]))
	buf.WriteString(fmt.Sprintf("- **Parameter Coverage**: %.1f%%\n", stats["parameter_coverage"]))
	buf.WriteString(fmt.Sprintf("- **Total Endpoints**: %d\n", stats["total_endpoints"]))
	buf.WriteString(fmt.Sprintf("- **Fully Tested**: %d\n", stats["tested_endpoints"]))
	buf.WriteString(fmt.Sprintf("- **Partially Tested**: %d\n", stats["partial_endpoints"]))
	buf.WriteString(fmt.Sprintf("- **Errors**: %d\n", stats["error_endpoints"]))
	buf.WriteString(fmt.Sprintf("- **Untested**: %d\n\n", stats["untested_endpoints"]))
	
	// Write endpoint details
	buf.WriteString("## Endpoint Details\n\n")
	buf.WriteString("| Method | Path | Status | Tags | Tests | Last Tested |\n")
	buf.WriteString("|--------|------|--------|------|-------|-------------|\n")
	
	for _, endpoint := range endpoints {
		// Format tags
		tags := strings.Join(endpoint.Tags, ", ")
		
		// Format last tested time
		lastTested := "Never"
		if !endpoint.LastTested.IsZero() {
			lastTested = endpoint.LastTested.Format("2006-01-02 15:04:05")
		}
		
		// Write endpoint row
		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %d | %s |\n", 
			endpoint.Method, 
			endpoint.Path, 
			endpoint.Status, 
			tags, 
			endpoint.TestCount, 
			lastTested))
		
		// Write parameter details if detail level > 1
		if c.options.DetailLevel > 1 && len(endpoint.Parameters) > 0 {
			buf.WriteString("\n**Parameters:**\n\n")
			for _, param := range endpoint.Parameters {
				required := ""
				if param.Required {
					required = ", required"
				}
				
				tested := "Not tested"
				if param.Tested {
					tested = fmt.Sprintf("Tested %d times", param.TestCount)
				}
				
				buf.WriteString(fmt.Sprintf("- **%s** (%s%s) - %s\n", 
					param.Name, 
					param.Type, 
					required, 
					tested))
			}
			buf.WriteString("\n")
		}
	}
	
	// Write footer
	buf.WriteString(fmt.Sprintf("\n*Report generated by ffuf API Coverage Analyzer. Duration: %s*\n", stats["duration"]))
	
	return buf.String(), nil
}

// generateTextReport generates a plain text coverage report
func (c *CoverageAnalyzer) generateTextReport() (string, error) {
	stats := c.GetCoverageStats()
	endpoints := c.getEndpointsForReport()
	
	var buf bytes.Buffer
	
	// Write header
	buf.WriteString("API COVERAGE REPORT\n")
	buf.WriteString("===================\n\n")
	buf.WriteString(fmt.Sprintf("Generated on %s\n\n", stats["timestamp"]))
	
	// Write summary
	buf.WriteString("SUMMARY\n-------\n\n")
	buf.WriteString(fmt.Sprintf("Endpoint Coverage:  %.1f%%\n", stats["endpoint_coverage"]))
	buf.WriteString(fmt.Sprintf("Parameter Coverage: %.1f%%\n", stats["parameter_coverage"]))
	buf.WriteString(fmt.Sprintf("Total Endpoints:    %d\n", stats["total_endpoints"]))
	buf.WriteString(fmt.Sprintf("Fully Tested:       %d\n", stats["tested_endpoints"]))
	buf.WriteString(fmt.Sprintf("Partially Tested:   %d\n", stats["partial_endpoints"]))
	buf.WriteString(fmt.Sprintf("Errors:             %d\n", stats["error_endpoints"]))
	buf.WriteString(fmt.Sprintf("Untested:           %d\n\n", stats["untested_endpoints"]))
	
	// Write endpoint details
	buf.WriteString("ENDPOINT DETAILS\n----------------\n\n")
	
	for _, endpoint := range endpoints {
		// Format last tested time
		lastTested := "Never"
		if !endpoint.LastTested.IsZero() {
			lastTested = endpoint.LastTested.Format("2006-01-02 15:04:05")
		}
		
		// Write endpoint details
		buf.WriteString(fmt.Sprintf("%s %s\n", endpoint.Method, endpoint.Path))
		buf.WriteString(fmt.Sprintf("  Status:     %s\n", endpoint.Status))
		if len(endpoint.Tags) > 0 {
			buf.WriteString(fmt.Sprintf("  Tags:       %s\n", strings.Join(endpoint.Tags, ", ")))
		}
		buf.WriteString(fmt.Sprintf("  Tests:      %d\n", endpoint.TestCount))
		buf.WriteString(fmt.Sprintf("  Last Tested: %s\n", lastTested))
		
		// Write parameter details if detail level > 1
		if c.options.DetailLevel > 1 && len(endpoint.Parameters) > 0 {
			buf.WriteString("  Parameters:\n")
			for _, param := range endpoint.Parameters {
				required := ""
				if param.Required {
					required = ", required"
				}
				
				tested := "Not tested"
				if param.Tested {
					tested = fmt.Sprintf("Tested %d times", param.TestCount)
				}
				
				buf.WriteString(fmt.Sprintf("    - %s (%s%s) - %s\n", 
					param.Name, 
					param.Type, 
					required, 
					tested))
			}
		}
		
		buf.WriteString("\n")
	}
	
	// Write footer
	buf.WriteString(fmt.Sprintf("Report generated by ffuf API Coverage Analyzer. Duration: %s\n", stats["duration"]))
	
	return buf.String(), nil
}

// getEndpointsForReport returns a slice of endpoints for the report
func (c *CoverageAnalyzer) getEndpointsForReport() []*EndpointCoverage {
	endpoints := make([]*EndpointCoverage, 0, len(c.endpoints))
	
	for _, endpoint := range c.endpoints {
		// Skip untested endpoints if not including them
		if !c.options.IncludeUntested && endpoint.Status == StatusUntested {
			continue
		}
		
		// Add a copy of the endpoint to the slice
		endpointCopy := *endpoint
		endpoints = append(endpoints, &endpointCopy)
	}
	
	// Sort endpoints by path and method
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path == endpoints[j].Path {
			return endpoints[i].Method < endpoints[j].Method
		}
		return endpoints[i].Path < endpoints[j].Path
	})
	
	return endpoints
}