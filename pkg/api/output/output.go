// Package output provides API-specific output formatting with syntax highlighting.
package output

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// APIOutput is an output provider that adds syntax highlighting for API responses.
type APIOutput struct {
	config            *ffuf.Config
	fuzzkeywords      []string
	Results           []ffuf.Result
	CurrentResults    []ffuf.Result
	highlighter       *Highlighter
	responseDataCache map[string][]byte // Cache of response data for highlighting
}

// NewAPIOutput creates a new APIOutput instance.
func NewAPIOutput(conf *ffuf.Config) *APIOutput {
	var outp APIOutput
	outp.config = conf
	outp.Results = make([]ffuf.Result, 0)
	outp.CurrentResults = make([]ffuf.Result, 0)
	outp.fuzzkeywords = make([]string, 0)
	for _, ip := range conf.InputProviders {
		outp.fuzzkeywords = append(outp.fuzzkeywords, ip.Keyword)
	}
	sort.Strings(outp.fuzzkeywords)
	outp.highlighter = NewHighlighter()
	outp.responseDataCache = make(map[string][]byte)
	return &outp
}

// Banner prints the banner.
func (a *APIOutput) Banner() {
	version := strings.ReplaceAll(ffuf.Version(), "<3", fmt.Sprintf("%s<3%s", ANSI_RED, ANSI_CLEAR))
	fmt.Fprintf(os.Stderr, "%s\n       v%s\n%s\n\n", BANNER_HEADER, version, BANNER_SEP)
	fmt.Fprintf(os.Stderr, "%sAPI Output Mode Enabled - Syntax highlighting for API responses%s\n\n", ANSI_GREEN, ANSI_CLEAR)
	printOption([]byte("Method"), []byte(a.config.Method))
	printOption([]byte("URL"), []byte(a.config.Url))

	// Print wordlists
	for _, provider := range a.config.InputProviders {
		if provider.Name == "wordlist" {
			printOption([]byte("Wordlist"), []byte(provider.Keyword+": "+provider.Value))
		}
	}

	// Print headers
	if len(a.config.Headers) > 0 {
		for k, v := range a.config.Headers {
			printOption([]byte("Header"), []byte(fmt.Sprintf("%s: %s", k, v)))
		}
	}
	// Print POST data
	if len(a.config.Data) > 0 {
		printOption([]byte("Data"), []byte(a.config.Data))
	}

	// Output file info
	if len(a.config.OutputFile) > 0 {
		// Use filename as specified by user
		OutputFile := a.config.OutputFile

		if a.config.OutputFormat == "all" {
			// Actually... append all extensions
			OutputFile += ".{json,ejson,html,md,csv,ecsv}"
		}

		printOption([]byte("Output file"), []byte(OutputFile))
		printOption([]byte("File format"), []byte(a.config.OutputFormat))
	}

	// Follow redirects?
	follow := fmt.Sprintf("%t", a.config.FollowRedirects)
	printOption([]byte("Follow redirects"), []byte(follow))

	// Autocalibration
	autocalib := fmt.Sprintf("%t", a.config.AutoCalibration)
	printOption([]byte("Calibration"), []byte(autocalib))

	// Proxies
	if len(a.config.ProxyURL) > 0 {
		printOption([]byte("Proxy"), []byte(a.config.ProxyURL))
	}
	if len(a.config.ReplayProxyURL) > 0 {
		printOption([]byte("ReplayProxy"), []byte(a.config.ReplayProxyURL))
	}
}

// Finalize prints the final results.
func (a *APIOutput) Finalize() error {
	fmt.Fprintf(os.Stderr, "\n\nResults Summary:\n")
	fmt.Fprintf(os.Stderr, "\n%d requests made\n", len(a.Results))

	if len(a.config.OutputFile) > 0 {
		// No need to create file handle for stdout
		var err error
		var filename string
		var format string

		if a.config.OutputFormat == "all" {
			formatsToUse := []string{"json", "ejson", "html", "md", "csv", "ecsv"}
			for _, format = range formatsToUse {
				filename = fmt.Sprintf("%s.%s", a.config.OutputFile, format)
				err = a.writeResultsToFile(filename, format)
				if err != nil {
					a.Error(err.Error())
				}
			}
		} else {
			filename = a.config.OutputFile
			format = a.config.OutputFormat
			err = a.writeResultsToFile(filename, format)
			if err != nil {
				a.Error(err.Error())
			}
		}
	}
	if !a.config.Quiet {
		fmt.Fprintf(os.Stderr, "\n")
	}
	return nil
}

// writeResultsToFile writes the results to a file.
func (a *APIOutput) writeResultsToFile(filename, format string) error {
	// Simplified implementation that only supports JSON format
	// For other formats, return an error
	switch format {
	case "json":
		// Convert results to JSON
		jsonData, err := json.Marshal(a.Results)
		if err != nil {
			return err
		}
		// Write JSON to file
		return os.WriteFile(filename, jsonData, 0644)
	default:
		return fmt.Errorf("Format %s not supported in API output mode", format)
	}
}

// Progress prints the progress.
func (a *APIOutput) Progress(status ffuf.Progress) {
	if a.config.Quiet {
		return
	}
	dur := time.Since(status.StartedAt)
	runningSecs := int(dur.Seconds())
	var reqRate int64
	if runningSecs > 0 {
		reqRate = status.ReqSec
	} else {
		reqRate = 0
	}
	hours := dur / time.Hour
	dur -= hours * time.Hour
	mins := dur / time.Minute
	dur -= mins * time.Minute
	secs := dur / time.Second
	fmt.Fprintf(os.Stderr, "%s :: Progress: [%d/%d] :: Job: [%d/%d] :: %d req/sec :: Duration: [%d:%02d:%02d] :: Errors: %d %s", TERMINAL_CLEAR_LINE, status.ReqCount, status.ReqTotal, status.QueuePos, status.QueueTotal, reqRate, hours, mins, secs, status.ErrorCount, CURSOR_CLEAR_LINE)
}

// Info prints an info message.
func (a *APIOutput) Info(infostring string) {
	if a.config.Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "%s[INFO] %s%s\n", ANSI_BLUE, infostring, ANSI_CLEAR)
}

// Error prints an error message.
func (a *APIOutput) Error(errstring string) {
	if a.config.Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "%s[ERROR] %s%s\n", ANSI_RED, errstring, ANSI_CLEAR)
}

// Warning prints a warning message.
func (a *APIOutput) Warning(warnstring string) {
	if a.config.Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "%s[WARNING] %s%s\n", ANSI_YELLOW, warnstring, ANSI_CLEAR)
}

// Raw prints a raw message.
func (a *APIOutput) Raw(output string) {
	fmt.Println(output)
}

// Result processes a result.
func (a *APIOutput) Result(resp ffuf.Response) {
	// Do we want to write request and response to a file
	if len(a.config.OutputDirectory) > 0 {
		resp.ResultFile = writeResponseToFile(resp, a.config.OutputDirectory)
	}

	// Generate a unique key for the response data cache
	// If we have a ResultFile, use that as the key
	// Otherwise, generate a key based on the URL, status code, and content length
	cacheKey := resp.ResultFile
	if cacheKey == "" {
		cacheKey = fmt.Sprintf("%s-%d-%d", resp.Request.Url, resp.StatusCode, resp.ContentLength)
	}

	// Store the response data in the cache
	if a.responseDataCache == nil {
		a.responseDataCache = make(map[string][]byte)
	}
	a.responseDataCache[cacheKey] = resp.Data

	inputs := make(map[string][]byte, len(resp.Request.Input))
	for k, v := range resp.Request.Input {
		inputs[k] = v
	}
	sResult := ffuf.Result{
		Input:            inputs,
		Position:         resp.Request.Position,
		StatusCode:       resp.StatusCode,
		ContentLength:    resp.ContentLength,
		ContentWords:     resp.ContentWords,
		ContentLines:     resp.ContentLines,
		ContentType:      resp.ContentType,
		RedirectLocation: resp.GetRedirectLocation(false),
		ScraperData:      resp.ScraperData,
		Url:              resp.Request.Url,
		Duration:         resp.Duration,
		ResultFile:       resp.ResultFile,
		Host:             resp.Request.Host,
	}
	a.CurrentResults = append(a.CurrentResults, sResult)
	// Output the result
	a.PrintResult(sResult)
}

// PrintResult prints a result.
func (a *APIOutput) PrintResult(res ffuf.Result) {
	switch {
	case a.config.Json:
		a.resultJson(res)
	case a.config.Quiet:
		a.resultQuiet(res)
	case len(a.fuzzkeywords) > 1 || a.config.Verbose || len(a.config.OutputDirectory) > 0 || len(res.ScraperData) > 0:
		// Print a multi-line result (when using multiple input keywords and wordlists)
		a.resultMultiline(res)
	default:
		a.resultNormal(res)
	}
}

// prepareInputsOneLine prepares the inputs for one-line output.
func (a *APIOutput) prepareInputsOneLine(res ffuf.Result) string {
	inputs := ""
	if len(a.fuzzkeywords) > 1 {
		for _, k := range a.fuzzkeywords {
			if ffuf.StrInSlice(k, a.config.CommandKeywords) {
				// If we're using external command for input, display the position instead of input
				inputs = fmt.Sprintf("%s%s : %d ", inputs, k, res.Position)
			} else {
				inputs = fmt.Sprintf("%s%s : %s ", inputs, k, res.Input[k])
			}
		}
	} else {
		for _, k := range a.fuzzkeywords {
			if ffuf.StrInSlice(k, a.config.CommandKeywords) {
				// If we're using external command for input, display the position instead of input
				inputs = fmt.Sprintf("%d", res.Position)
			} else {
				inputs = string(res.Input[k])
			}
		}
	}
	return inputs
}

// resultQuiet prints a result in quiet mode.
func (a *APIOutput) resultQuiet(res ffuf.Result) {
	fmt.Println(a.prepareInputsOneLine(res))
}

// resultMultiline prints a result in multi-line mode.
func (a *APIOutput) resultMultiline(res ffuf.Result) {
	var res_hdr, res_str string
	res_str = "%s%s    * %s: %s\n"
	res_hdr = fmt.Sprintf("%s%s[Status: %d, Size: %d, Words: %d, Lines: %d, Duration: %dms]%s", TERMINAL_CLEAR_LINE, colorize(res.StatusCode), res.StatusCode, res.ContentLength, res.ContentWords, res.ContentLines, res.Duration.Milliseconds(), ANSI_CLEAR)
	reslines := ""
	if a.config.Verbose {
		reslines = fmt.Sprintf("%s%s| URL | %s\n", reslines, TERMINAL_CLEAR_LINE, res.Url)
		redirectLocation := res.RedirectLocation
		if redirectLocation != "" {
			reslines = fmt.Sprintf("%s%s| --> | %s\n", reslines, TERMINAL_CLEAR_LINE, redirectLocation)
		}
	}
	if res.ResultFile != "" {
		reslines = fmt.Sprintf("%s%s| RES | %s\n", reslines, TERMINAL_CLEAR_LINE, res.ResultFile)
	}
	for _, k := range a.fuzzkeywords {
		if ffuf.StrInSlice(k, a.config.CommandKeywords) {
			// If we're using external command for input, display the position instead of input
			reslines = fmt.Sprintf(res_str, reslines, TERMINAL_CLEAR_LINE, k, fmt.Sprintf("%d", res.Position))
		} else {
			// Wordlist input
			reslines = fmt.Sprintf(res_str, reslines, TERMINAL_CLEAR_LINE, k, res.Input[k])
		}
	}
	if len(res.ScraperData) > 0 {
		reslines = fmt.Sprintf("%s%s| SCR |\n", reslines, TERMINAL_CLEAR_LINE)
		for k, vslice := range res.ScraperData {
			for _, v := range vslice {
				reslines = fmt.Sprintf(res_str, reslines, TERMINAL_CLEAR_LINE, k, v)
			}
		}
	}
	fmt.Printf("%s\n%s\n", res_hdr, reslines)

	// Add API-specific output with syntax highlighting
	if isAPIResponse(res) {
		fmt.Printf("%s%s| API Response | %s\n", TERMINAL_CLEAR_LINE, ANSI_GREEN, ANSI_CLEAR)
		highlightedResponse := a.highlighter.Highlight(res, a.responseDataCache)
		fmt.Println(highlightedResponse)
	}
}

// resultNormal prints a result in normal mode.
func (a *APIOutput) resultNormal(res ffuf.Result) {
	resnormal := fmt.Sprintf("%s%s%-23s [Status: %d, Size: %d, Words: %d, Lines: %d, Duration: %dms]%s", TERMINAL_CLEAR_LINE, colorize(res.StatusCode), a.prepareInputsOneLine(res), res.StatusCode, res.ContentLength, res.ContentWords, res.ContentLines, res.Duration.Milliseconds(), ANSI_CLEAR)
	fmt.Println(resnormal)

	// Add API-specific output with syntax highlighting
	if isAPIResponse(res) {
		highlightedResponse := a.highlighter.Highlight(res, a.responseDataCache)
		fmt.Println(highlightedResponse)
	}
}

// resultJson prints a result in JSON mode.
func (a *APIOutput) resultJson(res ffuf.Result) {
	// Use the highlighter to format the JSON output
	jsonOutput := a.highlighter.HighlightJSON(res)
	fmt.Fprint(os.Stderr, TERMINAL_CLEAR_LINE)
	fmt.Println(jsonOutput)
}

// GetCurrentResults returns the current results.
func (a *APIOutput) GetCurrentResults() []ffuf.Result {
	return a.CurrentResults
}

// SetCurrentResults sets the current results.
func (a *APIOutput) SetCurrentResults(results []ffuf.Result) {
	a.CurrentResults = results
}

// Reset resets the output.
func (a *APIOutput) Reset() {
	a.CurrentResults = make([]ffuf.Result, 0)
}

// Cycle cycles the output.
func (a *APIOutput) Cycle() {
	a.Results = append(a.Results, a.CurrentResults...)
	a.Reset()
}

// SaveFile saves the results to a file.
func (a *APIOutput) SaveFile(filename, format string) error {
	return a.writeResultsToFile(filename, format)
}

// isAPIResponse checks if a response is an API response.
func isAPIResponse(res ffuf.Result) bool {
	contentType := res.ContentType
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "text/xml") ||
		strings.Contains(contentType, "application/graphql")
}

// writeResponseToFile writes a response to a file.
func writeResponseToFile(resp ffuf.Response, outputDir string) string {
	var fileContent, fileName, filePath string
	// Create directory if needed
	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0750)
		if err != nil {
			if !os.IsExist(err) {
				fmt.Fprintf(os.Stderr, "Failed to create output directory: %s\n", err)
				return ""
			}
		}
	}
	fileContent = fmt.Sprintf("%s\n---- ↑ Request ---- Response ↓ ----\n\n%s", resp.Request.Raw, resp.Raw)

	// Create file name
	fileName = fmt.Sprintf("%x", md5.Sum([]byte(fileContent)))

	filePath = fmt.Sprintf("%s/%s", outputDir, fileName)
	err := os.WriteFile(filePath, []byte(fileContent), 0640)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write response to file: %s\n", err)
		return ""
	}
	return fileName
}

// ANSI color codes
const (
	ANSI_CLEAR  = "\x1b[0m"
	ANSI_RED    = "\x1b[31m"
	ANSI_GREEN  = "\x1b[32m"
	ANSI_YELLOW = "\x1b[33m"
	ANSI_BLUE   = "\x1b[34m"
	ANSI_PURPLE = "\x1b[35m"
	ANSI_CYAN   = "\x1b[36m"
	ANSI_WHITE  = "\x1b[37m"
)

// Terminal control sequences
const (
	TERMINAL_CLEAR_LINE = "\r\x1b[2K"
	CURSOR_CLEAR_LINE   = "\x1b[0K"
)

// Banner constants
const (
	BANNER_HEADER = `
        /'___\  /'___\           /'___\       
       /\ \__/ /\ \__/  __  __  /\ \__/       
       \ \ ,__\\ \ ,__\/\ \/\ \ \ \ ,__\      
        \ \ \_/ \ \ \_/\ \ \_\ \ \ \ \_/      
         \ \_\   \ \_\  \ \____/  \ \_\       
          \/_/    \/_/   \/___/    \/_/       
`
	BANNER_SEP = "________________________________________________"
)

// colorize returns the ANSI color code for a status code.
func colorize(status int64) string {
	colorCode := ANSI_CLEAR
	if status >= 200 && status < 300 {
		colorCode = ANSI_GREEN
	}
	if status >= 300 && status < 400 {
		colorCode = ANSI_BLUE
	}
	if status >= 400 && status < 500 {
		colorCode = ANSI_YELLOW
	}
	if status >= 500 && status < 600 {
		colorCode = ANSI_RED
	}
	return colorCode
}

// printOption prints an option.
func printOption(name []byte, value []byte) {
	fmt.Fprintf(os.Stderr, " :: %-16s : %s\n", name, value)
}
