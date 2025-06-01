// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// VisFormat represents the format of the visualization
type VisFormat string

const (
	// VisFormatJSON represents JSON format
	VisFormatJSON VisFormat = "json"
	// VisFormatHTML represents HTML format
	VisFormatHTML VisFormat = "html"
	// VisFormatDOT represents DOT format (for Graphviz)
	VisFormatDOT VisFormat = "dot"
	// VisFormatMermaid represents Mermaid format
	VisFormatMermaid VisFormat = "mermaid"
)

// VisType represents the type of visualization
type VisType string

const (
	// VisTypeTree represents a tree visualization
	VisTypeTree VisType = "tree"
	// VisTypeGraph represents a graph visualization
	VisTypeGraph VisType = "graph"
	// VisTypeSchema represents a schema visualization
	VisTypeSchema VisType = "schema"
	// VisTypeSequence represents a sequence diagram
	VisTypeSequence VisType = "sequence"
)

// VisOptions contains options for generating visualizations
type VisOptions struct {
	// Format is the output format
	Format VisFormat
	// Type is the visualization type
	Type VisType
	// MaxDepth is the maximum depth to visualize
	MaxDepth int
	// IncludeValues indicates whether to include values in the visualization
	IncludeValues bool
	// ColorScheme is the color scheme to use
	ColorScheme string
	// Title is the title of the visualization
	Title string
}

// DefaultVisOptions returns default visualization options
func DefaultVisOptions() *VisOptions {
	return &VisOptions{
		Format:        VisFormatHTML,
		Type:          VisTypeTree,
		MaxDepth:      10,
		IncludeValues: true,
		ColorScheme:   "default",
		Title:         "API Response Visualization",
	}
}

// Visualizer provides methods for visualizing API responses
type Visualizer struct {
	options *VisOptions
}

// NewVisualizer creates a new Visualizer with the given options
func NewVisualizer(options *VisOptions) *Visualizer {
	if options == nil {
		options = DefaultVisOptions()
	}
	return &Visualizer{
		options: options,
	}
}

// VisualizeResponse generates a visualization of an API response
func (v *Visualizer) VisualizeResponse(resp *ffuf.Response) (string, error) {
	// Check if the response is JSON
	if !strings.Contains(resp.ContentType, "application/json") {
		return "", api.NewAPIError("Response is not in JSON format", 0)
	}

	// Parse the response body as JSON
	var jsonData interface{}
	if err := json.Unmarshal(resp.Data, &jsonData); err != nil {
		return "", api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
	}

	// Generate the visualization based on the format
	switch v.options.Format {
	case VisFormatJSON:
		return v.generateJSONVisualization(jsonData)
	case VisFormatHTML:
		return v.generateHTMLVisualization(jsonData, resp.Request.Url)
	case VisFormatDOT:
		return v.generateDOTVisualization(jsonData, resp.Request.Url)
	case VisFormatMermaid:
		return v.generateMermaidVisualization(jsonData, resp.Request.Url)
	default:
		return "", api.NewAPIError("Unsupported visualization format", 0)
	}
}

// VisualizeCorrelations generates a visualization of correlations between API responses
func (v *Visualizer) VisualizeCorrelations(correlations []Correlation) (string, error) {
	// Generate the visualization based on the format
	switch v.options.Format {
	case VisFormatJSON:
		return v.generateJSONCorrelationVisualization(correlations)
	case VisFormatHTML:
		return v.generateHTMLCorrelationVisualization(correlations)
	case VisFormatDOT:
		return v.generateDOTCorrelationVisualization(correlations)
	case VisFormatMermaid:
		return v.generateMermaidCorrelationVisualization(correlations)
	default:
		return "", api.NewAPIError("Unsupported visualization format", 0)
	}
}

// VisualizeSchema generates a visualization of an API schema
func (v *Visualizer) VisualizeSchema(schema *Schema) (string, error) {
	// Generate the visualization based on the format
	switch v.options.Format {
	case VisFormatJSON:
		return v.generateJSONSchemaVisualization(schema)
	case VisFormatHTML:
		return v.generateHTMLSchemaVisualization(schema)
	case VisFormatDOT:
		return v.generateDOTSchemaVisualization(schema)
	case VisFormatMermaid:
		return v.generateMermaidSchemaVisualization(schema)
	default:
		return "", api.NewAPIError("Unsupported visualization format", 0)
	}
}

// generateJSONVisualization generates a JSON visualization of the data
func (v *Visualizer) generateJSONVisualization(data interface{}) (string, error) {
	// Pretty-print the JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", api.NewAPIError("Failed to marshal JSON: "+err.Error(), 0)
	}
	return string(jsonBytes), nil
}

// generateHTMLVisualization generates an HTML visualization of the data
func (v *Visualizer) generateHTMLVisualization(data interface{}, url string) (string, error) {
	// Convert the data to a tree structure
	tree := v.convertToTree(data, v.options.MaxDepth)

	// Create an HTML template
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        .tree {
            margin-left: 20px;
        }
        .node {
            margin: 5px 0;
        }
        .key {
            font-weight: bold;
            color: #333;
        }
        .value {
            color: #666;
        }
        .object {
            color: #0066cc;
        }
        .array {
            color: #009900;
        }
        .string {
            color: #cc6600;
        }
        .number {
            color: #cc0000;
        }
        .boolean {
            color: #9900cc;
        }
        .null {
            color: #999;
            font-style: italic;
        }
        .collapsible {
            cursor: pointer;
        }
        .content {
            display: block;
            padding-left: 20px;
        }
        .hidden {
            display: none;
        }
        .endpoint {
            font-size: 1.2em;
            margin-bottom: 10px;
            padding: 5px;
            background-color: #f0f0f0;
            border-radius: 5px;
        }
    </style>
    <script>
        function toggleCollapse(id) {
            var content = document.getElementById(id);
            if (content.style.display === "none") {
                content.style.display = "block";
            } else {
                content.style.display = "none";
            }
        }
    </script>
</head>
<body>
    <h1>{{.Title}}</h1>
    <div class="endpoint">Endpoint: {{.URL}}</div>
    <div class="tree">
        {{.TreeHTML}}
    </div>
</body>
</html>`

	// Create a template data structure
	type TemplateData struct {
		Title    string
		URL      string
		TreeHTML template.HTML
	}

	// Generate the tree HTML
	treeHTML := v.generateTreeHTML(tree, 0)

	// Create the template data
	templateData := TemplateData{
		Title:    v.options.Title,
		URL:      url,
		TreeHTML: template.HTML(treeHTML),
	}

	// Execute the template
	tmplObj, err := template.New("visualization").Parse(tmpl)
	if err != nil {
		return "", api.NewAPIError("Failed to parse template: "+err.Error(), 0)
	}

	var buf bytes.Buffer
	if err := tmplObj.Execute(&buf, templateData); err != nil {
		return "", api.NewAPIError("Failed to execute template: "+err.Error(), 0)
	}

	return buf.String(), nil
}

// generateDOTVisualization generates a DOT visualization of the data
func (v *Visualizer) generateDOTVisualization(data interface{}, url string) (string, error) {
	// Start the DOT graph
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("digraph %s {\n", sanitizeID(v.options.Title)))
	buf.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	buf.WriteString(fmt.Sprintf("  root [label=\"%s\", shape=ellipse, fillcolor=lightgreen];\n", sanitizeLabel(url)))

	// Generate the DOT nodes and edges
	v.generateDOTNodes(&buf, data, "root", 0)

	// End the DOT graph
	buf.WriteString("}\n")

	return buf.String(), nil
}

// generateMermaidVisualization generates a Mermaid visualization of the data
func (v *Visualizer) generateMermaidVisualization(data interface{}, url string) (string, error) {
	// Start the Mermaid diagram
	var buf bytes.Buffer
	buf.WriteString("graph TD\n")
	buf.WriteString(fmt.Sprintf("  root[\"%s\"]\n", sanitizeLabel(url)))

	// Generate the Mermaid nodes and edges
	v.generateMermaidNodes(&buf, data, "root", 0)

	return buf.String(), nil
}

// generateJSONCorrelationVisualization generates a JSON visualization of correlations
func (v *Visualizer) generateJSONCorrelationVisualization(correlations []Correlation) (string, error) {
	// Pretty-print the JSON
	jsonBytes, err := json.MarshalIndent(correlations, "", "  ")
	if err != nil {
		return "", api.NewAPIError("Failed to marshal JSON: "+err.Error(), 0)
	}
	return string(jsonBytes), nil
}

// generateHTMLCorrelationVisualization generates an HTML visualization of correlations
func (v *Visualizer) generateHTMLCorrelationVisualization(correlations []Correlation) (string, error) {
	// Create an HTML template
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        table {
            border-collapse: collapse;
            width: 100%;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        .high-confidence {
            background-color: #d4edda;
        }
        .medium-confidence {
            background-color: #fff3cd;
        }
        .low-confidence {
            background-color: #f8d7da;
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <table>
        <tr>
            <th>Type</th>
            <th>Source Path</th>
            <th>Target Path</th>
            <th>Value</th>
            <th>Confidence</th>
            <th>Description</th>
        </tr>
        {{range .Correlations}}
        <tr class="{{.ConfidenceClass}}">
            <td>{{.Type}}</td>
            <td>{{.SourcePath}}</td>
            <td>{{.TargetPath}}</td>
            <td>{{.SourceValue}}</td>
            <td>{{.Confidence}}%</td>
            <td>{{.Description}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>`

	// Create a template data structure
	type CorrelationData struct {
		Type            string
		SourcePath      string
		TargetPath      string
		SourceValue     string
		Confidence      int
		Description     string
		ConfidenceClass string
	}

	type TemplateData struct {
		Title        string
		Correlations []CorrelationData
	}

	// Create the correlation data
	correlationData := make([]CorrelationData, 0, len(correlations))
	for _, correlation := range correlations {
		confidenceClass := "low-confidence"
		if correlation.Confidence >= 80 {
			confidenceClass = "high-confidence"
		} else if correlation.Confidence >= 50 {
			confidenceClass = "medium-confidence"
		}

		correlationData = append(correlationData, CorrelationData{
			Type:            string(correlation.Type),
			SourcePath:      correlation.SourcePath,
			TargetPath:      correlation.TargetPath,
			SourceValue:     correlation.SourceValue,
			Confidence:      correlation.Confidence,
			Description:     correlation.Description,
			ConfidenceClass: confidenceClass,
		})
	}

	// Create the template data
	templateData := TemplateData{
		Title:        v.options.Title,
		Correlations: correlationData,
	}

	// Execute the template
	tmplObj, err := template.New("visualization").Parse(tmpl)
	if err != nil {
		return "", api.NewAPIError("Failed to parse template: "+err.Error(), 0)
	}

	var buf bytes.Buffer
	if err := tmplObj.Execute(&buf, templateData); err != nil {
		return "", api.NewAPIError("Failed to execute template: "+err.Error(), 0)
	}

	return buf.String(), nil
}

// generateDOTCorrelationVisualization generates a DOT visualization of correlations
func (v *Visualizer) generateDOTCorrelationVisualization(correlations []Correlation) (string, error) {
	// Start the DOT graph
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("digraph %s {\n", sanitizeID(v.options.Title)))
	buf.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	buf.WriteString("  edge [fontsize=10];\n")

	// Create a map to track nodes
	nodes := make(map[string]bool)

	// Generate the DOT nodes and edges
	for i, correlation := range correlations {
		sourceID := fmt.Sprintf("source_%d", i)
		targetID := fmt.Sprintf("target_%d", i)

		// Add source node if not already added
		if !nodes[correlation.SourcePath] {
			buf.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", sourceID, sanitizeLabel(correlation.SourcePath)))
			nodes[correlation.SourcePath] = true
		}

		// Add target node if not already added
		if !nodes[correlation.TargetPath] {
			buf.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", targetID, sanitizeLabel(correlation.TargetPath)))
			nodes[correlation.TargetPath] = true
		}

		// Add edge
		edgeColor := "black"
		if correlation.Confidence >= 80 {
			edgeColor = "green"
		} else if correlation.Confidence >= 50 {
			edgeColor = "orange"
		} else {
			edgeColor = "red"
		}

		buf.WriteString(fmt.Sprintf("  %s -> %s [label=\"%s\\n%d%%\", color=%s];\n",
			sourceID, targetID, correlation.Type, correlation.Confidence, edgeColor))
	}

	// End the DOT graph
	buf.WriteString("}\n")

	return buf.String(), nil
}

// generateMermaidCorrelationVisualization generates a Mermaid visualization of correlations
func (v *Visualizer) generateMermaidCorrelationVisualization(correlations []Correlation) (string, error) {
	// Start the Mermaid diagram
	var buf bytes.Buffer
	buf.WriteString("graph TD\n")

	// Create a map to track nodes
	nodes := make(map[string]bool)

	// Generate the Mermaid nodes and edges
	for i, correlation := range correlations {
		sourceID := fmt.Sprintf("source_%d", i)
		targetID := fmt.Sprintf("target_%d", i)

		// Add source node if not already added
		if !nodes[correlation.SourcePath] {
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", sourceID, sanitizeLabel(correlation.SourcePath)))
			nodes[correlation.SourcePath] = true
		}

		// Add target node if not already added
		if !nodes[correlation.TargetPath] {
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", targetID, sanitizeLabel(correlation.TargetPath)))
			nodes[correlation.TargetPath] = true
		}

		// Add edge
		linkStyle := ""
		if correlation.Confidence >= 80 {
			linkStyle = "stroke:green,stroke-width:2px"
		} else if correlation.Confidence >= 50 {
			linkStyle = "stroke:orange,stroke-width:1.5px"
		} else {
			linkStyle = "stroke:red,stroke-width:1px"
		}

		buf.WriteString(fmt.Sprintf("  %s -->|%s %d%%| %s\n",
			sourceID, correlation.Type, correlation.Confidence, targetID))
		buf.WriteString(fmt.Sprintf("  linkStyle %d %s\n", i, linkStyle))
	}

	return buf.String(), nil
}

// generateJSONSchemaVisualization generates a JSON visualization of a schema
func (v *Visualizer) generateJSONSchemaVisualization(schema *Schema) (string, error) {
	// Pretty-print the JSON
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", api.NewAPIError("Failed to marshal JSON: "+err.Error(), 0)
	}
	return string(jsonBytes), nil
}

// generateHTMLSchemaVisualization generates an HTML visualization of a schema
func (v *Visualizer) generateHTMLSchemaVisualization(schema *Schema) (string, error) {
	// Create an HTML template
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        .schema {
            margin-left: 20px;
        }
        .field {
            margin: 10px 0;
            padding: 5px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        .field-name {
            font-weight: bold;
            color: #333;
        }
        .field-type {
            color: #0066cc;
        }
        .field-format {
            color: #009900;
        }
        .field-pattern {
            color: #cc6600;
        }
        .field-required {
            color: #cc0000;
        }
        .object-fields {
            margin-left: 20px;
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <div class="schema">
        {{.SchemaHTML}}
    </div>
</body>
</html>`

	// Create a template data structure
	type TemplateData struct {
		Title      string
		SchemaHTML template.HTML
	}

	// Generate the schema HTML
	schemaHTML := v.generateSchemaHTML(schema)

	// Create the template data
	templateData := TemplateData{
		Title:      v.options.Title,
		SchemaHTML: template.HTML(schemaHTML),
	}

	// Execute the template
	tmplObj, err := template.New("visualization").Parse(tmpl)
	if err != nil {
		return "", api.NewAPIError("Failed to parse template: "+err.Error(), 0)
	}

	var buf bytes.Buffer
	if err := tmplObj.Execute(&buf, templateData); err != nil {
		return "", api.NewAPIError("Failed to execute template: "+err.Error(), 0)
	}

	return buf.String(), nil
}

// generateDOTSchemaVisualization generates a DOT visualization of a schema
func (v *Visualizer) generateDOTSchemaVisualization(schema *Schema) (string, error) {
	// Start the DOT graph
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("digraph %s {\n", sanitizeID(v.options.Title)))
	buf.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	buf.WriteString("  edge [fontsize=10];\n")

	// Add the root node
	buf.WriteString("  root [label=\"Schema\", shape=ellipse, fillcolor=lightgreen];\n")

	// Generate the DOT nodes and edges for the schema
	v.generateDOTSchemaNodes(&buf, schema, "root")

	// End the DOT graph
	buf.WriteString("}\n")

	return buf.String(), nil
}

// generateMermaidSchemaVisualization generates a Mermaid visualization of a schema
func (v *Visualizer) generateMermaidSchemaVisualization(schema *Schema) (string, error) {
	// Start the Mermaid diagram
	var buf bytes.Buffer
	buf.WriteString("classDiagram\n")

	// Generate the Mermaid class diagram for the schema
	v.generateMermaidSchemaClasses(&buf, schema)

	return buf.String(), nil
}

// Helper functions

// TreeNode represents a node in the visualization tree
type TreeNode struct {
	Key      string
	Type     string
	Value    interface{}
	Children []*TreeNode
}

// convertToTree converts data to a tree structure
func (v *Visualizer) convertToTree(data interface{}, maxDepth int) *TreeNode {
	// Ensure maxDepth is at least 1 to show the root node
	if maxDepth < 1 {
		maxDepth = 1
	}
	return v.convertToTreeRecursive(data, "", maxDepth, 0)
}

// convertToTreeRecursive recursively converts data to a tree structure
func (v *Visualizer) convertToTreeRecursive(data interface{}, key string, maxDepth, depth int) *TreeNode {
	if depth >= maxDepth {
		return &TreeNode{
			Key:   key,
			Type:  "truncated",
			Value: "...",
		}
	}

	node := &TreeNode{
		Key: key,
	}

	switch val := data.(type) {
	case map[string]interface{}:
		node.Type = "object"
		node.Children = make([]*TreeNode, 0, len(val))
		for k, value := range val {
			child := v.convertToTreeRecursive(value, k, maxDepth, depth+1)
			node.Children = append(node.Children, child)
		}
	case []interface{}:
		node.Type = "array"
		node.Children = make([]*TreeNode, 0, len(val))
		for i, value := range val {
			child := v.convertToTreeRecursive(value, fmt.Sprintf("[%d]", i), maxDepth, depth+1)
			node.Children = append(node.Children, child)
		}
	case string:
		node.Type = "string"
		node.Value = val
	case float64:
		node.Type = "number"
		node.Value = val
	case bool:
		node.Type = "boolean"
		node.Value = val
	case nil:
		node.Type = "null"
		node.Value = "null"
	default:
		node.Type = "unknown"
		node.Value = fmt.Sprintf("%v", val)
	}

	return node
}

// generateTreeHTML generates HTML for a tree node
func (v *Visualizer) generateTreeHTML(node *TreeNode, id int) string {
	var buf bytes.Buffer

	contentID := fmt.Sprintf("content_%d", id)
	nextID := id + 1

	// Generate the node HTML
	if node.Key != "" {
		buf.WriteString(fmt.Sprintf("<div class=\"node\"><span class=\"key\">%s: </span>", node.Key))
	} else {
		buf.WriteString("<div class=\"node\">")
	}

	// Add the value or children
	switch node.Type {
	case "object":
		buf.WriteString(fmt.Sprintf("<span class=\"object collapsible\" onclick=\"toggleCollapse('%s')\">Object {%d}</span>", contentID, len(node.Children)))
		buf.WriteString(fmt.Sprintf("<div id=\"%s\" class=\"content\">", contentID))
		for _, child := range node.Children {
			buf.WriteString(v.generateTreeHTML(child, nextID))
			nextID += countNodes(child)
		}
		buf.WriteString("</div>")
	case "array":
		buf.WriteString(fmt.Sprintf("<span class=\"array collapsible\" onclick=\"toggleCollapse('%s')\">Array [%d]</span>", contentID, len(node.Children)))
		buf.WriteString(fmt.Sprintf("<div id=\"%s\" class=\"content\">", contentID))
		for _, child := range node.Children {
			buf.WriteString(v.generateTreeHTML(child, nextID))
			nextID += countNodes(child)
		}
		buf.WriteString("</div>")
	case "string":
		buf.WriteString(fmt.Sprintf("<span class=\"string\">\"%s\"</span>", node.Value))
	case "number":
		buf.WriteString(fmt.Sprintf("<span class=\"number\">%v</span>", node.Value))
	case "boolean":
		buf.WriteString(fmt.Sprintf("<span class=\"boolean\">%v</span>", node.Value))
	case "null":
		buf.WriteString("<span class=\"null\">null</span>")
	case "truncated":
		buf.WriteString("<span class=\"truncated\">...</span>")
	default:
		buf.WriteString(fmt.Sprintf("<span class=\"value\">%v</span>", node.Value))
	}

	buf.WriteString("</div>")

	return buf.String()
}

// countNodes counts the number of nodes in a tree
func countNodes(node *TreeNode) int {
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}

// generateDOTNodes generates DOT nodes and edges for a data structure
func (v *Visualizer) generateDOTNodes(buf *bytes.Buffer, data interface{}, parentID string, depth int) int {
	if depth > v.options.MaxDepth {
		return 0
	}

	nodeCount := 0

	switch val := data.(type) {
	case map[string]interface{}:
		for key, value := range val {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, sanitizeLabel(key)))
			buf.WriteString(fmt.Sprintf("  %s -> %s;\n", parentID, nodeID))
			nodeCount += 1 + v.generateDOTNodes(buf, value, nodeID, depth+1)
		}
	case []interface{}:
		for i, value := range val {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s [label=\"[%d]\"];\n", nodeID, i))
			buf.WriteString(fmt.Sprintf("  %s -> %s;\n", parentID, nodeID))
			nodeCount += 1 + v.generateDOTNodes(buf, value, nodeID, depth+1)
		}
	default:
		if v.options.IncludeValues {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s [label=\"%s\", shape=ellipse, fillcolor=lightyellow];\n", nodeID, sanitizeLabel(fmt.Sprintf("%v", val))))
			buf.WriteString(fmt.Sprintf("  %s -> %s;\n", parentID, nodeID))
			nodeCount++
		}
	}

	return nodeCount
}

// generateMermaidNodes generates Mermaid nodes and edges for a data structure
func (v *Visualizer) generateMermaidNodes(buf *bytes.Buffer, data interface{}, parentID string, depth int) int {
	if depth > v.options.MaxDepth {
		return 0
	}

	nodeCount := 0

	switch val := data.(type) {
	case map[string]interface{}:
		for key, value := range val {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeID, sanitizeLabel(key)))
			buf.WriteString(fmt.Sprintf("  %s --> %s\n", parentID, nodeID))
			nodeCount += 1 + v.generateMermaidNodes(buf, value, nodeID, depth+1)
		}
	case []interface{}:
		for i, value := range val {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s[\"[%d]\"]\n", nodeID, i))
			buf.WriteString(fmt.Sprintf("  %s --> %s\n", parentID, nodeID))
			nodeCount += 1 + v.generateMermaidNodes(buf, value, nodeID, depth+1)
		}
	default:
		if v.options.IncludeValues {
			nodeID := fmt.Sprintf("%s_%d", parentID, nodeCount)
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeID, sanitizeLabel(fmt.Sprintf("%v", val))))
			buf.WriteString(fmt.Sprintf("  %s --> %s\n", parentID, nodeID))
			nodeCount++
		}
	}

	return nodeCount
}

// generateDOTSchemaNodes generates DOT nodes and edges for a schema
func (v *Visualizer) generateDOTSchemaNodes(buf *bytes.Buffer, schema *Schema, parentID string) {
	// Add nodes for each property in the schema
	for name, field := range schema.Properties {
		nodeID := fmt.Sprintf("%s_%s", parentID, sanitizeID(name))
		label := fmt.Sprintf("%s: %s", name, field.Type)
		if field.Format != "" {
			label += fmt.Sprintf(" (%s)", field.Format)
		}
		if field.Required {
			label += " *"
		}

		buf.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, sanitizeLabel(label)))
		buf.WriteString(fmt.Sprintf("  %s -> %s;\n", parentID, nodeID))

		// If the field has properties, add them
		if field.Properties != nil {
			for subName, subField := range field.Properties {
				subNodeID := fmt.Sprintf("%s_%s", nodeID, sanitizeID(subName))
				subLabel := fmt.Sprintf("%s: %s", subName, subField.Type)
				if subField.Format != "" {
					subLabel += fmt.Sprintf(" (%s)", subField.Format)
				}
				if subField.Required {
					subLabel += " *"
				}

				buf.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", subNodeID, sanitizeLabel(subLabel)))
				buf.WriteString(fmt.Sprintf("  %s -> %s;\n", nodeID, subNodeID))
			}
		}
	}
}

// generateMermaidSchemaClasses generates Mermaid class diagram for a schema
func (v *Visualizer) generateMermaidSchemaClasses(buf *bytes.Buffer, schema *Schema) {
	// Add the main class
	buf.WriteString("  class Schema {\n")
	for name, field := range schema.Properties {
		fieldStr := fmt.Sprintf("    +%s %s", field.Type, name)
		if field.Format != "" {
			fieldStr += fmt.Sprintf(" (%s)", field.Format)
		}
		if field.Required {
			fieldStr += " *"
		}
		buf.WriteString(fieldStr + "\n")
	}
	buf.WriteString("  }\n")

	// Add relationships for object fields
	for name, field := range schema.Properties {
		if field.Type == "object" && field.Properties != nil {
			className := sanitizeID(name)
			buf.WriteString(fmt.Sprintf("  class %s {\n", className))
			for subName, subField := range field.Properties {
				fieldStr := fmt.Sprintf("    +%s %s", subField.Type, subName)
				if subField.Format != "" {
					fieldStr += fmt.Sprintf(" (%s)", subField.Format)
				}
				if subField.Required {
					fieldStr += " *"
				}
				buf.WriteString(fieldStr + "\n")
			}
			buf.WriteString("  }\n")
			buf.WriteString(fmt.Sprintf("  Schema --> %s\n", className))
		}
	}
}

// generateSchemaHTML generates HTML for a schema
func (v *Visualizer) generateSchemaHTML(schema *Schema) string {
	var buf bytes.Buffer

	// Generate HTML for each property in the schema
	for name, field := range schema.Properties {
		buf.WriteString("<div class=\"field\">")
		buf.WriteString(fmt.Sprintf("<span class=\"field-name\">%s</span>: ", name))
		buf.WriteString(fmt.Sprintf("<span class=\"field-type\">%s</span>", field.Type))

		if field.Format != "" {
			buf.WriteString(fmt.Sprintf(" <span class=\"field-format\">(%s)</span>", field.Format))
		}

		if field.Pattern != "" {
			buf.WriteString(fmt.Sprintf(" <span class=\"field-pattern\">pattern: %s</span>", field.Pattern))
		}

		if field.Required {
			buf.WriteString(" <span class=\"field-required\">*required</span>")
		}

		// If the field has properties, add them
		if field.Properties != nil {
			buf.WriteString("<div class=\"object-fields\">")
			for subName, subField := range field.Properties {
				buf.WriteString("<div class=\"field\">")
				buf.WriteString(fmt.Sprintf("<span class=\"field-name\">%s</span>: ", subName))
				buf.WriteString(fmt.Sprintf("<span class=\"field-type\">%s</span>", subField.Type))

				if subField.Format != "" {
					buf.WriteString(fmt.Sprintf(" <span class=\"field-format\">(%s)</span>", subField.Format))
				}

				if subField.Pattern != "" {
					buf.WriteString(fmt.Sprintf(" <span class=\"field-pattern\">pattern: %s</span>", subField.Pattern))
				}

				if subField.Required {
					buf.WriteString(" <span class=\"field-required\">*required</span>")
				}

				buf.WriteString("</div>")
			}
			buf.WriteString("</div>")
		}

		buf.WriteString("</div>")
	}

	return buf.String()
}

// sanitizeID sanitizes a string for use as a DOT ID
func sanitizeID(s string) string {
	// Replace non-alphanumeric characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(s, "_")
}

// sanitizeLabel sanitizes a string for use as a DOT label
func sanitizeLabel(s string) string {
	// Escape quotes and backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// VisualizeResponse is a convenience method on ResponseParser to visualize a response
func (p *ResponseParser) VisualizeResponse(resp *ffuf.Response, format VisFormat) (string, error) {
	options := DefaultVisOptions()
	options.Format = format
	visualizer := NewVisualizer(options)
	return visualizer.VisualizeResponse(resp)
}
