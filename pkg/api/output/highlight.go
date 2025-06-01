// Package output provides API-specific output formatting with syntax highlighting.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// Highlighter provides syntax highlighting for API responses.
type Highlighter struct {
	// Style is the syntax highlighting style to use
	Style string
	// MaxResponseSize is the maximum size of response to highlight
	MaxResponseSize int
}

// NewHighlighter creates a new Highlighter instance.
func NewHighlighter() *Highlighter {
	return &Highlighter{
		Style:           "monokai",
		MaxResponseSize: 10000, // 10KB max to avoid performance issues
	}
}

// Highlight highlights a response based on its content type.
func (h *Highlighter) Highlight(res ffuf.Result, responseDataCache map[string][]byte) string {
	// Check if we have data to highlight
	if res.ContentLength == 0 {
		return ""
	}

	// Get the content type
	contentType := res.ContentType

	// Determine the lexer to use based on content type
	var lexer chroma.Lexer
	if strings.Contains(contentType, "application/json") {
		lexer = lexers.Get("json")
	} else if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		lexer = lexers.Get("xml")
	} else if strings.Contains(contentType, "application/graphql") {
		lexer = lexers.Get("graphql")
	} else {
		// Default to plain text for unknown content types
		lexer = lexers.Get("text")
	}

	// If we couldn't determine the lexer, use fallback
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get the formatter and style
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	style := styles.Get(h.Style)
	if style == nil {
		style = styles.Fallback
	}

	// Get the response data from the cache
	var data []byte

	// Generate the cache key
	cacheKey := res.ResultFile
	if cacheKey == "" {
		cacheKey = fmt.Sprintf("%s-%d-%d", res.Url, res.StatusCode, res.ContentLength)
	}

	// Try to get the data from the cache
	if responseDataCache != nil {
		if cachedData, ok := responseDataCache[cacheKey]; ok {
			data = cachedData
		}
	}

	// If we couldn't get the data from the cache, generate placeholder data
	if data == nil || len(data) == 0 {
		// Create a placeholder response based on the content type
		switch {
		case strings.Contains(contentType, "application/json"):
			// Create a simple JSON object with the result information
			jsonObj := map[string]interface{}{
				"status":       res.StatusCode,
				"content_type": res.ContentType,
				"url":          res.Url,
				"duration_ms":  res.Duration.Milliseconds(),
				"note":         "This is a placeholder. Enable response caching for actual data.",
			}
			data, _ = json.MarshalIndent(jsonObj, "", "  ")
		case strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml"):
			// Create a simple XML document with the result information
			data = []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<response>
  <status>%d</status>
  <content_type>%s</content_type>
  <url>%s</url>
  <duration_ms>%d</duration_ms>
  <note>This is a placeholder. Enable response caching for actual data.</note>
</response>`, res.StatusCode, res.ContentType, res.Url, res.Duration.Milliseconds()))
		default:
			// For other content types, just use a simple text representation
			data = []byte(fmt.Sprintf("Status: %d\nContent-Type: %s\nURL: %s\nDuration: %dms\nNote: This is a placeholder. Enable response caching for actual data.",
				res.StatusCode, res.ContentType, res.Url, res.Duration.Milliseconds()))
		}
	}

	// Limit the size of the data to highlight
	if len(data) > h.MaxResponseSize {
		data = append(data[:h.MaxResponseSize], []byte("\n... (truncated)")...)
	}

	// Highlight the data
	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, string(data))
	if err != nil {
		return string(data) // Return the original data if tokenization fails
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return string(data) // Return the original data if formatting fails
	}

	return buf.String()
}

// HighlightJSON highlights a JSON result.
func (h *Highlighter) HighlightJSON(res ffuf.Result) string {
	// Convert the result to JSON
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		// If marshaling fails, return a simple string representation
		return fmt.Sprintf("Error marshaling result to JSON: %v", err)
	}

	// Highlight the JSON
	lexer := lexers.Get("json")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	style := styles.Get(h.Style)
	if style == nil {
		style = styles.Fallback
	}

	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, string(jsonBytes))
	if err != nil {
		return string(jsonBytes) // Return the original JSON if tokenization fails
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return string(jsonBytes) // Return the original JSON if formatting fails
	}

	return buf.String()
}
