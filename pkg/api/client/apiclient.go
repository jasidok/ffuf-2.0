// Package client provides an optimized HTTP client for API testing.
package client

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"

	"github.com/andybalholm/brotli"
)

// APIClient is an optimized HTTP client for API testing
type APIClient struct {
	config        *ffuf.Config
	client        *http.Client
	baseRunner    *runner.SimpleRunner
	commonHeaders map[string]string
	authTokens    map[string]string
}

// NewAPIClient creates a new API client with optimized settings for API testing
func NewAPIClient(conf *ffuf.Config) *APIClient {
	baseRunner := runner.NewSimpleRunner(conf, false).(*runner.SimpleRunner)

	// Create a transport with optimized settings for API testing
	proxyURL := http.ProxyFromEnvironment
	if len(conf.ProxyURL) > 0 {
		pu, err := url.Parse(conf.ProxyURL)
		if err == nil {
			proxyURL = http.ProxyURL(pu)
		}
	}

	cert := []tls.Certificate{}
	if conf.ClientCert != "" && conf.ClientKey != "" {
		tmp, _ := tls.LoadX509KeyPair(conf.ClientCert, conf.ClientKey)
		cert = []tls.Certificate{tmp}
	}

	transport := &http.Transport{
		ForceAttemptHTTP2:   true, // Always try HTTP/2 for APIs
		Proxy:               proxyURL,
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 500,
		MaxConnsPerHost:     500,
		IdleConnTimeout:     90 * time.Second, // Increased for API testing
		TLSHandshakeTimeout: time.Duration(time.Duration(conf.Timeout) * time.Second),
		DisableCompression:  false, // Enable compression for APIs
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(time.Duration(conf.Timeout) * time.Second),
			KeepAlive: 30 * time.Second, // Increased for API testing
			DualStack: true,             // Support IPv4 and IPv6
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12, // Minimum TLS 1.2 for security
			Renegotiation:      tls.RenegotiateOnceAsClient,
			ServerName:         conf.SNI,
			Certificates:       cert,
		},
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:       time.Duration(time.Duration(conf.Timeout) * time.Second),
		Transport:     transport,
	}

	if conf.FollowRedirects {
		client.CheckRedirect = nil
	}

	// Common headers for API requests
	commonHeaders := map[string]string{
		"Accept":       "application/json, application/xml, */*",
		"Content-Type": "application/json",
		"User-Agent":   fmt.Sprintf("%s v%s (API Client)", "Fuzz Faster U Fool", ffuf.Version()),
	}

	return &APIClient{
		config:        conf,
		client:        client,
		baseRunner:    baseRunner,
		commonHeaders: commonHeaders,
		authTokens:    make(map[string]string),
	}
}

// SetAuthToken sets an authentication token for API requests
func (c *APIClient) SetAuthToken(tokenType, token string) {
	c.authTokens[tokenType] = token
}

// GetAuthToken gets an authentication token
func (c *APIClient) GetAuthToken(tokenType string) string {
	return c.authTokens[tokenType]
}

// SetCommonHeader sets a common header for all API requests
func (c *APIClient) SetCommonHeader(name, value string) {
	c.commonHeaders[name] = value
}

// Prepare prepares a request with API-specific optimizations
func (c *APIClient) Prepare(input map[string][]byte, basereq *ffuf.Request) (ffuf.Request, error) {
	req := ffuf.CopyRequest(basereq)

	// Apply common headers for API requests if not already set
	for k, v := range c.commonHeaders {
		if _, exists := req.Headers[k]; !exists {
			req.Headers[k] = v
		}
	}

	// Apply auth tokens if available
	for tokenType, token := range c.authTokens {
		switch strings.ToLower(tokenType) {
		case "bearer":
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
		case "basic":
			req.Headers["Authorization"] = fmt.Sprintf("Basic %s", token)
		case "apikey":
			req.Headers["X-API-Key"] = token
		}
	}

	// Process input replacements
	for keyword, inputitem := range input {
		req.Method = strings.ReplaceAll(req.Method, keyword, string(inputitem))
		headers := make(map[string]string, len(req.Headers))
		for h, v := range req.Headers {
			replacedHeader := strings.ReplaceAll(h, keyword, string(inputitem))
			// Validate that the replaced header name is valid
			if replacedHeader != "" && !strings.ContainsAny(replacedHeader, " \t\r\n") {
				var CanonicalHeader string = textproto.CanonicalMIMEHeaderKey(replacedHeader)
				headers[CanonicalHeader] = strings.ReplaceAll(v, keyword, string(inputitem))
			}
		}
		req.Headers = headers
		req.Url = strings.ReplaceAll(req.Url, keyword, string(inputitem))
		req.Data = []byte(strings.ReplaceAll(string(req.Data), keyword, string(inputitem)))
	}

	req.Input = input
	return req, nil
}

// Execute executes an API request with optimized handling
func (c *APIClient) Execute(req *ffuf.Request) (ffuf.Response, error) {
	var httpreq *http.Request
	var err error
	var rawreq []byte
	data := bytes.NewReader(req.Data)

	var start time.Time
	var firstByteTime time.Duration

	trace := &httptrace.ClientTrace{
		WroteRequest: func(wri httptrace.WroteRequestInfo) {
			start = time.Now() // begin the timer after the request is fully written
		},
		GotFirstResponseByte: func() {
			firstByteTime = time.Since(start) // record when the first byte of the response was received
		},
	}

	// Create request with context timeout to prevent hanging
	ctx := c.config.Context
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(c.config.Context, time.Duration(c.config.Timeout)*time.Second)
		defer cancel()
	}

	httpreq, err = http.NewRequestWithContext(ctx, req.Method, req.Url, data)

	if err != nil {
		return ffuf.Response{}, err
	}

	// Handle Go http.Request special cases
	if _, ok := req.Headers["Host"]; ok {
		httpreq.Host = req.Headers["Host"]
	}

	req.Host = httpreq.Host
	httpreq = httpreq.WithContext(httptrace.WithClientTrace(c.config.Context, trace))

	// Set all headers
	for k, v := range req.Headers {
		httpreq.Header.Set(k, v)
	}

	// Apply auth tokens if not already set (for cases where Prepare wasn't called)
	if httpreq.Header.Get("Authorization") == "" {
		for tokenType, token := range c.authTokens {
			switch strings.ToLower(tokenType) {
			case "bearer":
				httpreq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			case "basic":
				httpreq.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))
			}
		}
	}

	// Apply API key if not already set
	if httpreq.Header.Get("X-API-Key") == "" {
		if apiKey, exists := c.authTokens["apikey"]; exists {
			httpreq.Header.Set("X-API-Key", apiKey)
		}
	}

	// Apply common headers if not already set
	for k, v := range c.commonHeaders {
		if httpreq.Header.Get(k) == "" {
			httpreq.Header.Set(k, v)
		}
	}

	if len(c.config.OutputDirectory) > 0 || len(c.config.AuditLog) > 0 {
		rawreq, _ = httputil.DumpRequestOut(httpreq, true)
		req.Raw = string(rawreq)
	}

	// Execute the request with optimized client
	httpresp, err := c.client.Do(httpreq)
	if err != nil {
		return ffuf.Response{}, err
	}

	req.Timestamp = start

	resp := ffuf.NewResponse(httpresp, req)
	defer httpresp.Body.Close()

	// Check if we should download the resource or not
	size, err := strconv.Atoi(httpresp.Header.Get("Content-Length"))
	if err == nil {
		resp.ContentLength = int64(size)
		if (c.config.IgnoreBody) || (size > runner.MAX_DOWNLOAD_SIZE) {
			resp.Cancelled = true
			return resp, nil
		}
	}

	if len(c.config.OutputDirectory) > 0 || len(c.config.AuditLog) > 0 {
		rawresp, _ := httputil.DumpResponse(httpresp, true)
		resp.Request.Raw = string(rawreq)
		resp.Raw = string(rawresp)
	}

	// Handle different content encodings
	var bodyReader io.ReadCloser
	switch httpresp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(httpresp.Body)
		if err != nil {
			bodyReader = httpresp.Body
		}
	case "br":
		bodyReader = io.NopCloser(brotli.NewReader(httpresp.Body))
	case "deflate":
		bodyReader = flate.NewReader(httpresp.Body)
	default:
		bodyReader = httpresp.Body
	}

	// Read response body with streaming for large API responses
	if respbody, err := io.ReadAll(bodyReader); err == nil {
		resp.ContentLength = int64(len(string(respbody)))
		resp.Data = respbody
	}

	// Process response data
	wordsSize := len(strings.Split(string(resp.Data), " "))
	linesSize := len(strings.Split(string(resp.Data), "\n"))
	resp.ContentWords = int64(wordsSize)
	resp.ContentLines = int64(linesSize)
	resp.Duration = firstByteTime
	resp.Timestamp = start.Add(firstByteTime)

	// Set content type for easier API response handling
	resp.ContentType = httpresp.Header.Get("Content-Type")

	return resp, nil
}

// Dump dumps a request to bytes
func (c *APIClient) Dump(req *ffuf.Request) ([]byte, error) {
	var httpreq *http.Request
	var err error
	data := bytes.NewReader(req.Data)
	httpreq, err = http.NewRequestWithContext(c.config.Context, req.Method, req.Url, data)
	if err != nil {
		return []byte{}, err
	}

	// Handle Go http.Request special cases
	if _, ok := req.Headers["Host"]; ok {
		httpreq.Host = req.Headers["Host"]
	}

	req.Host = httpreq.Host
	for k, v := range req.Headers {
		httpreq.Header.Set(k, v)
	}
	return httputil.DumpRequestOut(httpreq, true)
}
