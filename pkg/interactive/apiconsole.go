package interactive

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api/diff"
	"github.com/ffuf/ffuf/v2/pkg/api/parser"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// APIConsole provides an interactive console for API testing
type APIConsole struct {
	Job           *ffuf.Job
	CurrentMethod string
	CurrentURL    string
	CurrentPath   string
	Headers       map[string]string
	QueryParams   map[string]string
	BodyParams    map[string]interface{}
	BodyFormat    string // json, xml, form, etc.
	LastResponse  *ffuf.Response
	PrevResponse  *ffuf.Response // For comparison
}

// NewAPIConsole creates a new APIConsole instance
func NewAPIConsole(job *ffuf.Job) *APIConsole {
	// Initialize with values from the job config
	headers := make(map[string]string)
	for k, v := range job.Config.Headers {
		headers[k] = v
	}

	return &APIConsole{
		Job:           job,
		CurrentMethod: job.Config.Method,
		CurrentURL:    job.Config.Url,
		CurrentPath:   "",
		Headers:       headers,
		QueryParams:   make(map[string]string),
		BodyParams:    make(map[string]interface{}),
		BodyFormat:    "json", // Default to JSON
	}
}

// HandleCommand processes API console commands
func (a *APIConsole) HandleCommand(args []string) {
	if len(args) == 0 {
		a.printHelp()
		return
	}

	switch args[0] {
	case "help":
		a.printHelp()
	case "method":
		a.handleMethod(args)
	case "url":
		a.handleURL(args)
	case "path":
		a.handlePath(args)
	case "header":
		a.handleHeader(args)
	case "query":
		a.handleQuery(args)
	case "body":
		a.handleBody(args)
	case "format":
		a.handleFormat(args)
	case "show":
		a.showRequest()
	case "send":
		a.sendRequest()
	case "diff":
		a.compareResponses()
	case "map":
		a.visualizeAPIMap(args)
	default:
		a.Job.Output.Warning(fmt.Sprintf("Unknown API console command: %s", args[0]))
	}
}

// handleMethod sets the HTTP method for the request
func (a *APIConsole) handleMethod(args []string) {
	if len(args) < 2 {
		a.Job.Output.Error("Please specify a method (GET, POST, PUT, DELETE, etc.)")
		return
	}
	method := strings.ToUpper(args[1])
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	valid := false
	for _, m := range validMethods {
		if m == method {
			valid = true
			break
		}
	}
	if !valid {
		a.Job.Output.Warning(fmt.Sprintf("Unusual HTTP method: %s", method))
	}
	a.CurrentMethod = method
	a.Job.Output.Info(fmt.Sprintf("HTTP method set to: %s", method))
}

// handleURL sets the base URL for the request
func (a *APIConsole) handleURL(args []string) {
	if len(args) < 2 {
		a.Job.Output.Error("Please specify a URL")
		return
	}
	a.CurrentURL = args[1]
	a.Job.Output.Info(fmt.Sprintf("Base URL set to: %s", a.CurrentURL))
}

// handlePath sets the path for the request
func (a *APIConsole) handlePath(args []string) {
	if len(args) < 2 {
		a.Job.Output.Error("Please specify a path")
		return
	}
	a.CurrentPath = args[1]
	a.Job.Output.Info(fmt.Sprintf("Path set to: %s", a.CurrentPath))
}

// handleHeader adds or removes a header
func (a *APIConsole) handleHeader(args []string) {
	if len(args) < 2 {
		// List all headers
		a.Job.Output.Raw("Current headers:\n")
		for k, v := range a.Headers {
			a.Job.Output.Raw(fmt.Sprintf("  %s: %s\n", k, v))
		}
		return
	}

	if args[1] == "clear" {
		a.Headers = make(map[string]string)
		a.Job.Output.Info("All headers cleared")
		return
	}

	if args[1] == "remove" && len(args) > 2 {
		delete(a.Headers, args[2])
		a.Job.Output.Info(fmt.Sprintf("Removed header: %s", args[2]))
		return
	}

	if len(args) < 3 {
		a.Job.Output.Error("Please specify header name and value, or 'clear' to remove all headers")
		return
	}

	a.Headers[args[1]] = args[2]
	a.Job.Output.Info(fmt.Sprintf("Added header: %s: %s", args[1], args[2]))
}

// handleQuery adds or removes a query parameter
func (a *APIConsole) handleQuery(args []string) {
	if len(args) < 2 {
		// List all query parameters
		a.Job.Output.Raw("Current query parameters:\n")
		for k, v := range a.QueryParams {
			a.Job.Output.Raw(fmt.Sprintf("  %s=%s\n", k, v))
		}
		return
	}

	if args[1] == "clear" {
		a.QueryParams = make(map[string]string)
		a.Job.Output.Info("All query parameters cleared")
		return
	}

	if args[1] == "remove" && len(args) > 2 {
		delete(a.QueryParams, args[2])
		a.Job.Output.Info(fmt.Sprintf("Removed query parameter: %s", args[2]))
		return
	}

	if len(args) < 3 {
		a.Job.Output.Error("Please specify parameter name and value, or 'clear' to remove all parameters")
		return
	}

	a.QueryParams[args[1]] = args[2]
	a.Job.Output.Info(fmt.Sprintf("Added query parameter: %s=%s", args[1], args[2]))
}

// handleBody adds or removes a body parameter
func (a *APIConsole) handleBody(args []string) {
	if len(args) < 2 {
		// List all body parameters
		a.Job.Output.Raw("Current body parameters:\n")
		jsonData, _ := json.MarshalIndent(a.BodyParams, "", "  ")
		a.Job.Output.Raw(fmt.Sprintf("%s\n", string(jsonData)))
		return
	}

	if args[1] == "clear" {
		a.BodyParams = make(map[string]interface{})
		a.Job.Output.Info("All body parameters cleared")
		return
	}

	if args[1] == "remove" && len(args) > 2 {
		delete(a.BodyParams, args[2])
		a.Job.Output.Info(fmt.Sprintf("Removed body parameter: %s", args[2]))
		return
	}

	if args[1] == "json" && len(args) > 2 {
		// Parse JSON string
		var jsonData map[string]interface{}
		jsonStr := strings.Join(args[2:], " ")
		err := json.Unmarshal([]byte(jsonStr), &jsonData)
		if err != nil {
			a.Job.Output.Error(fmt.Sprintf("Invalid JSON: %s", err))
			return
		}
		a.BodyParams = jsonData
		a.Job.Output.Info("Body set from JSON")
		return
	}

	if len(args) < 3 {
		a.Job.Output.Error("Please specify parameter name and value, or 'clear' to remove all parameters")
		return
	}

	a.BodyParams[args[1]] = args[2]
	a.Job.Output.Info(fmt.Sprintf("Added body parameter: %s=%s", args[1], args[2]))
}

// handleFormat sets the body format
func (a *APIConsole) handleFormat(args []string) {
	if len(args) < 2 {
		a.Job.Output.Info(fmt.Sprintf("Current body format: %s", a.BodyFormat))
		return
	}

	format := strings.ToLower(args[1])
	validFormats := []string{"json", "xml", "form"}
	valid := false
	for _, f := range validFormats {
		if f == format {
			valid = true
			break
		}
	}
	if !valid {
		a.Job.Output.Error(fmt.Sprintf("Invalid format: %s. Valid formats are: json, xml, form", format))
		return
	}

	a.BodyFormat = format
	a.Job.Output.Info(fmt.Sprintf("Body format set to: %s", format))
}

// showRequest displays the current request configuration
func (a *APIConsole) showRequest() {
	a.Job.Output.Raw("Current API Request:\n")
	a.Job.Output.Raw(fmt.Sprintf("  Method: %s\n", a.CurrentMethod))

	// Construct full URL with path and query parameters
	url := a.CurrentURL
	if a.CurrentPath != "" {
		if !strings.HasPrefix(a.CurrentPath, "/") && !strings.HasSuffix(url, "/") {
			url += "/"
		}
		url += a.CurrentPath
	}

	if len(a.QueryParams) > 0 {
		url += "?"
		params := make([]string, 0)
		for k, v := range a.QueryParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		url += strings.Join(params, "&")
	}

	a.Job.Output.Raw(fmt.Sprintf("  URL: %s\n", url))

	a.Job.Output.Raw("  Headers:\n")
	for k, v := range a.Headers {
		a.Job.Output.Raw(fmt.Sprintf("    %s: %s\n", k, v))
	}

	if len(a.BodyParams) > 0 {
		a.Job.Output.Raw(fmt.Sprintf("  Body Format: %s\n", a.BodyFormat))
		a.Job.Output.Raw("  Body:\n")

		switch a.BodyFormat {
		case "json":
			jsonData, _ := json.MarshalIndent(a.BodyParams, "    ", "  ")
			a.Job.Output.Raw(fmt.Sprintf("    %s\n", string(jsonData)))
		case "form":
			params := make([]string, 0)
			for k, v := range a.BodyParams {
				params = append(params, fmt.Sprintf("%s=%v", k, v))
			}
			a.Job.Output.Raw(fmt.Sprintf("    %s\n", strings.Join(params, "&")))
		case "xml":
			// Simple XML formatting for display purposes
			a.Job.Output.Raw("    <request>\n")
			for k, v := range a.BodyParams {
				a.Job.Output.Raw(fmt.Sprintf("      <%s>%v</%s>\n", k, v, k))
			}
			a.Job.Output.Raw("    </request>\n")
		}
	}
}

// sendRequest sends the current request
func (a *APIConsole) sendRequest() {
	// Construct the request
	url := a.CurrentURL
	if a.CurrentPath != "" {
		if !strings.HasPrefix(a.CurrentPath, "/") && !strings.HasSuffix(url, "/") {
			url += "/"
		}
		url += a.CurrentPath
	}

	if len(a.QueryParams) > 0 {
		url += "?"
		params := make([]string, 0)
		for k, v := range a.QueryParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		url += strings.Join(params, "&")
	}

	// Prepare body data
	var bodyData []byte
	if len(a.BodyParams) > 0 {
		switch a.BodyFormat {
		case "json":
			bodyData, _ = json.Marshal(a.BodyParams)
		case "form":
			params := make([]string, 0)
			for k, v := range a.BodyParams {
				params = append(params, fmt.Sprintf("%s=%v", k, v))
			}
			bodyData = []byte(strings.Join(params, "&"))
		case "xml":
			// Simple XML formatting
			xml := "<request>"
			for k, v := range a.BodyParams {
				xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
			}
			xml += "</request>"
			bodyData = []byte(xml)
		}
	}

	// Create a request
	req := &ffuf.Request{
		Method:  a.CurrentMethod,
		Url:     url,
		Headers: a.Headers,
	}

	if len(bodyData) > 0 {
		req.Data = bodyData
	}

	a.Job.Output.Info(fmt.Sprintf("Sending %s request to %s", a.CurrentMethod, url))

	// Execute the request
	resp, err := a.Job.Runner.Execute(req)
	if err != nil {
		a.Job.Output.Error(fmt.Sprintf("Request failed: %s", err))
		return
	}

	// Store the response for later comparison
	a.PrevResponse = a.LastResponse
	respCopy := resp // Create a copy to take its address
	a.LastResponse = &respCopy

	// Display the response
	a.Job.Output.Raw(fmt.Sprintf("\nResponse:\n"))
	a.Job.Output.Raw(fmt.Sprintf("  Status: %d\n", resp.StatusCode))
	a.Job.Output.Raw(fmt.Sprintf("  Content-Type: %s\n", resp.ContentType))
	a.Job.Output.Raw(fmt.Sprintf("  Content-Length: %d\n", resp.ContentLength))
	a.Job.Output.Raw(fmt.Sprintf("  Response Time: %dms\n\n", resp.Duration.Milliseconds()))

	// If we have API output formatting enabled, use it to display the response body
	if a.Job.Config.APIOutputFormat {
		// Create a result from the response for highlighting
		result := ffuf.Result{
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			ContentWords:  resp.ContentWords,
			ContentLines:  resp.ContentLines,
			ContentType:   resp.ContentType,
			Url:           url,
			Duration:      resp.Duration,
		}

		// Use the API output provider to highlight the response
		a.Job.Output.PrintResult(result)
	} else {
		// Simple output of response body
		a.Job.Output.Raw(fmt.Sprintf("  Body:\n%s\n", string(resp.Data)))
	}
}

// compareResponses compares the last two API responses and displays the differences
func (a *APIConsole) compareResponses() {
	if a.LastResponse == nil {
		a.Job.Output.Error("No responses to compare. Send at least one request first.")
		return
	}

	if a.PrevResponse == nil {
		a.Job.Output.Error("Only one response available. Send at least two requests to compare.")
		return
	}

	a.Job.Output.Info("Comparing previous two responses:")

	// Use the diff package to compare responses
	responseDiff := diff.CompareResponses(a.PrevResponse, a.LastResponse)

	// Display the formatted diff
	a.Job.Output.Raw(responseDiff.FormatDiff())
}

// visualizeAPIMap generates and displays a visualization of the API structure
func (a *APIConsole) visualizeAPIMap(args []string) {
	// Check if we have a response to visualize
	if a.LastResponse == nil {
		a.Job.Output.Error("No API response to visualize. Send a request first.")
		return
	}

	// Default format is HTML
	format := parser.VisFormatHTML

	// Check if a format was specified
	if len(args) > 1 {
		switch strings.ToLower(args[1]) {
		case "json":
			format = parser.VisFormatJSON
		case "html":
			format = parser.VisFormatHTML
		case "dot":
			format = parser.VisFormatDOT
		case "mermaid":
			format = parser.VisFormatMermaid
		default:
			a.Job.Output.Warning(fmt.Sprintf("Unknown visualization format: %s. Using HTML.", args[1]))
		}
	}

	// Create visualization options
	options := parser.DefaultVisOptions()
	options.Format = format
	options.Title = "API Map for " + a.CurrentURL

	// Create a visualizer
	visualizer := parser.NewVisualizer(options)

	// Generate the visualization
	visualization, err := visualizer.VisualizeResponse(a.LastResponse)
	if err != nil {
		a.Job.Output.Error(fmt.Sprintf("Failed to generate visualization: %s", err))
		return
	}

	// Display the visualization
	a.Job.Output.Raw(fmt.Sprintf("\nAPI Map Visualization (%s format):\n\n", format))
	a.Job.Output.Raw(visualization)

	// If it's HTML, suggest saving to a file
	if format == parser.VisFormatHTML {
		a.Job.Output.Info("To view the HTML visualization, save it to a file and open in a browser.")
	}
}

// printHelp displays the help for the API console
func (a *APIConsole) printHelp() {
	help := `
API Console Commands:
  method [METHOD]           - Set HTTP method (GET, POST, PUT, DELETE, etc.)
  url [URL]                 - Set base URL
  path [PATH]               - Set request path
  header [NAME] [VALUE]     - Add a header
  header remove [NAME]      - Remove a header
  header clear              - Clear all headers
  query [NAME] [VALUE]      - Add a query parameter
  query remove [NAME]       - Remove a query parameter
  query clear               - Clear all query parameters
  body [NAME] [VALUE]       - Add a body parameter
  body remove [NAME]        - Remove a body parameter
  body clear                - Clear all body parameters
  body json [JSON_STRING]   - Set body from JSON string
  format [FORMAT]           - Set body format (json, xml, form)
  show                      - Show current request
  send                      - Send the request
  diff                      - Compare the last two API responses
  map [FORMAT]              - Generate API structure visualization (formats: json, html, dot, mermaid)
  help                      - Show this help
`
	a.Job.Output.Raw(help)
}
