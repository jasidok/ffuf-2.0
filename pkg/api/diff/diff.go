package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// ResponseDiff represents the differences between two API responses
type ResponseDiff struct {
	StatusCodeDiff bool
	HeadersDiff    map[string]HeaderDiff
	BodyDiff       BodyDiff
	TimingDiff     int64 // Difference in milliseconds
}

// HeaderDiff represents the difference in a specific header
type HeaderDiff struct {
	InFirst  []string
	InSecond []string
}

// BodyDiff represents the differences in response bodies
type BodyDiff struct {
	ContentTypeDiff bool
	JSONDiff        map[string]JSONFieldDiff
	RawDiff         string // For non-JSON responses
}

// JSONFieldDiff represents the difference in a specific JSON field
type JSONFieldDiff struct {
	InFirst      interface{}
	InSecond     interface{}
	OnlyInFirst  bool
	OnlyInSecond bool
}

// CompareResponses compares two API responses and returns the differences
func CompareResponses(resp1, resp2 *ffuf.Response) *ResponseDiff {
	diff := &ResponseDiff{
		StatusCodeDiff: resp1.StatusCode != resp2.StatusCode,
		HeadersDiff:    make(map[string]HeaderDiff),
		BodyDiff: BodyDiff{
			ContentTypeDiff: resp1.ContentType != resp2.ContentType,
			JSONDiff:        make(map[string]JSONFieldDiff),
		},
		TimingDiff: resp2.Duration.Milliseconds() - resp1.Duration.Milliseconds(),
	}

	// Compare headers
	allHeaders := make(map[string]bool)
	for k := range resp1.Headers {
		allHeaders[k] = true
	}
	for k := range resp2.Headers {
		allHeaders[k] = true
	}

	for header := range allHeaders {
		val1, ok1 := resp1.Headers[header]
		val2, ok2 := resp2.Headers[header]

		if !ok1 || !ok2 || !reflect.DeepEqual(val1, val2) {
			diff.HeadersDiff[header] = HeaderDiff{
				InFirst:  val1,
				InSecond: val2,
			}
		}
	}

	// Compare bodies
	if isJSONResponse(resp1) && isJSONResponse(resp2) {
		diff.BodyDiff.JSONDiff = compareJSONBodies(resp1.Data, resp2.Data)
	} else {
		// For non-JSON responses, just do a simple diff
		diff.BodyDiff.RawDiff = simpleDiff(string(resp1.Data), string(resp2.Data))
	}

	return diff
}

// isJSONResponse checks if the response is JSON
func isJSONResponse(resp *ffuf.Response) bool {
	return strings.Contains(resp.ContentType, "application/json") && json.Valid(resp.Data)
}

// compareJSONBodies compares two JSON bodies and returns the differences
func compareJSONBodies(data1, data2 []byte) map[string]JSONFieldDiff {
	var json1, json2 map[string]interface{}

	// Try to unmarshal as objects first
	err1 := json.Unmarshal(data1, &json1)
	err2 := json.Unmarshal(data2, &json2)

	// If either fails, try as arrays
	if err1 != nil || err2 != nil {
		var arr1, arr2 []interface{}
		err1Arr := json.Unmarshal(data1, &arr1)
		err2Arr := json.Unmarshal(data2, &arr2)

		if err1Arr == nil && err2Arr == nil {
			// Both are arrays, convert to map for comparison
			json1 = make(map[string]interface{})
			json2 = make(map[string]interface{})

			for i, v := range arr1 {
				json1[fmt.Sprintf("[%d]", i)] = v
			}

			for i, v := range arr2 {
				json2[fmt.Sprintf("[%d]", i)] = v
			}
		} else {
			// Can't compare as JSON, return empty diff
			return make(map[string]JSONFieldDiff)
		}
	}

	return compareJSONMaps(json1, json2, "")
}

// compareJSONMaps compares two JSON objects recursively
func compareJSONMaps(json1, json2 map[string]interface{}, prefix string) map[string]JSONFieldDiff {
	result := make(map[string]JSONFieldDiff)

	// Get all keys from both maps
	allKeys := make(map[string]bool)
	for k := range json1 {
		allKeys[k] = true
	}
	for k := range json2 {
		allKeys[k] = true
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Compare each key
	for _, key := range keys {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		val1, ok1 := json1[key]
		val2, ok2 := json2[key]

		if !ok1 {
			// Only in second
			result[fullKey] = JSONFieldDiff{
				InSecond:     val2,
				OnlyInSecond: true,
			}
			continue
		}

		if !ok2 {
			// Only in first
			result[fullKey] = JSONFieldDiff{
				InFirst:     val1,
				OnlyInFirst: true,
			}
			continue
		}

		// Both exist, check if they're equal
		if !reflect.DeepEqual(val1, val2) {
			// If both are maps, recurse
			map1, isMap1 := val1.(map[string]interface{})
			map2, isMap2 := val2.(map[string]interface{})

			if isMap1 && isMap2 {
				// Recurse into nested objects
				nestedDiffs := compareJSONMaps(map1, map2, fullKey)
				for k, v := range nestedDiffs {
					result[k] = v
				}
			} else {
				// Simple value difference
				result[fullKey] = JSONFieldDiff{
					InFirst:  val1,
					InSecond: val2,
				}
			}
		}
	}

	return result
}

// simpleDiff creates a simple line-by-line diff of two strings
func simpleDiff(s1, s2 string) string {
	lines1 := strings.Split(s1, "\n")
	lines2 := strings.Split(s2, "\n")

	var diff bytes.Buffer

	// Find the maximum number of lines
	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		line1 := ""
		if i < len(lines1) {
			line1 = lines1[i]
		}

		line2 := ""
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			if line1 == "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", line2))
			} else if line2 == "" {
				diff.WriteString(fmt.Sprintf("- %s\n", line1))
			} else {
				diff.WriteString(fmt.Sprintf("- %s\n+ %s\n", line1, line2))
			}
		}
	}

	return diff.String()
}

// FormatDiff returns a formatted string representation of the diff
func (d *ResponseDiff) FormatDiff() string {
	var result bytes.Buffer

	// Status code diff
	if d.StatusCodeDiff {
		result.WriteString("Status Code: DIFFERENT\n")
	} else {
		result.WriteString("Status Code: SAME\n")
	}

	// Headers diff
	if len(d.HeadersDiff) > 0 {
		result.WriteString("\nHeaders Differences:\n")
		for header, diff := range d.HeadersDiff {
			if len(diff.InFirst) == 0 {
				result.WriteString(fmt.Sprintf("  + %s: %v (only in second response)\n", header, diff.InSecond))
			} else if len(diff.InSecond) == 0 {
				result.WriteString(fmt.Sprintf("  - %s: %v (only in first response)\n", header, diff.InFirst))
			} else {
				result.WriteString(fmt.Sprintf("  * %s: changed from '%v' to '%v'\n", header, diff.InFirst, diff.InSecond))
			}
		}
	} else {
		result.WriteString("\nHeaders: SAME\n")
	}

	// Content type diff
	if d.BodyDiff.ContentTypeDiff {
		result.WriteString("\nContent-Type: DIFFERENT\n")
	} else {
		result.WriteString("\nContent-Type: SAME\n")
	}

	// Body diff
	if len(d.BodyDiff.JSONDiff) > 0 {
		result.WriteString("\nJSON Body Differences:\n")
		for field, diff := range d.BodyDiff.JSONDiff {
			if diff.OnlyInFirst {
				result.WriteString(fmt.Sprintf("  - %s: %v (only in first response)\n", field, diff.InFirst))
			} else if diff.OnlyInSecond {
				result.WriteString(fmt.Sprintf("  + %s: %v (only in second response)\n", field, diff.InSecond))
			} else {
				result.WriteString(fmt.Sprintf("  * %s: changed from '%v' to '%v'\n", field, diff.InFirst, diff.InSecond))
			}
		}
	} else if d.BodyDiff.RawDiff != "" {
		result.WriteString("\nBody Differences:\n")
		result.WriteString(d.BodyDiff.RawDiff)
	} else {
		result.WriteString("\nBody: SAME\n")
	}

	// Timing diff
	result.WriteString(fmt.Sprintf("\nTiming Difference: %dms\n", d.TimingDiff))

	return result.String()
}
