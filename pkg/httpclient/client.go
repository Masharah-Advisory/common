package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceClient is a smart HTTP client for service-to-service communication
type ServiceClient struct {
	client        *http.Client
	serviceID     string
	serviceSecret string
	serviceHosts  map[string]string
}

// ServiceConfig holds service host mappings (only configure what you need)
type ServiceConfig map[string]string

// NewServiceClient creates a new service client
func NewServiceClient(serviceID, serviceSecret string, config ServiceConfig) *ServiceClient {
	return &ServiceClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		serviceID:     serviceID,
		serviceSecret: serviceSecret,
		serviceHosts:  config,
	}
}

// Get performs a smart GET request with auto context extraction
func (c *ServiceClient) Get(ctx context.Context, route string) (*http.Response, error) {
	return c.smartRequest(ctx, "GET", route, nil)
}

// Post performs a smart POST request with auto context extraction
func (c *ServiceClient) Post(ctx context.Context, route string, payload interface{}) (*http.Response, error) {
	return c.smartRequest(ctx, "POST", route, payload)
}

// Put performs a smart PUT request with auto context extraction
func (c *ServiceClient) Put(ctx context.Context, route string, payload interface{}) (*http.Response, error) {
	return c.smartRequest(ctx, "PUT", route, payload)
}

// Delete performs a smart DELETE request with auto context extraction
func (c *ServiceClient) Delete(ctx context.Context, route string) (*http.Response, error) {
	return c.smartRequest(ctx, "DELETE", route, nil)
}

// smartRequest auto-detects service and extracts headers from context
func (c *ServiceClient) smartRequest(ctx context.Context, method, route string, payload interface{}) (*http.Response, error) {
	// Build full URL by detecting service
	fullURL, err := c.buildURL(route)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Extract headers from context
	headers := c.extractHeaders(ctx)

	return c.doRequest(method, fullURL, payload, headers)
}

// buildURL detects service from route and builds full URL
func (c *ServiceClient) buildURL(route string) (string, error) {
	// Clean route
	route = strings.TrimPrefix(route, "/")
	// Route has api/vX/service format - extract service name
	parts := strings.Split(route, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid API route format: %s", route)
	}

	// parts[0] = "api", parts[1] = "v1", parts[2] = service name
	serviceName := parts[2]
	host, exists := c.serviceHosts[serviceName]
	if !exists {
		return "", fmt.Errorf("no host configured for service: %s", serviceName)
	}

	// Build full URL preserving the API version
	fullURL := strings.TrimSuffix(host, "/") + "/" + route
	return fullURL, nil
}

// extractHeaders gets headers from Gin context or standard context
func (c *ServiceClient) extractHeaders(ctx context.Context) map[string]string {
	headers := make(map[string]string)

	// Try Gin context first
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userID := ginCtx.GetHeader("X-User-ID"); userID != "" {
			headers["X-User-ID"] = userID
		}
		if userID, exists := ginCtx.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok {
				headers["X-User-ID"] = strconv.FormatUint(uint64(uid), 10)
			}
		}
		if requestID := ginCtx.GetHeader("X-Request-ID"); requestID != "" {
			headers["X-Request-ID"] = requestID
		}
		if acceptLang := ginCtx.GetHeader("Accept-Language"); acceptLang != "" {
			headers["Accept-Language"] = acceptLang
		}
		return headers
	}

	// Try standard context values
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(uint); ok {
			headers["X-User-ID"] = strconv.FormatUint(uint64(uid), 10)
		}
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			headers["X-Request-ID"] = rid
		}
	}

	return headers
}

// doRequest is the core method that handles all requests
func (c *ServiceClient) doRequest(method, url string, payload interface{}, contextHeaders map[string]string) (*http.Response, error) {
	var body []byte
	var err error

	// Marshal payload if provided
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	// Create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-ID", c.serviceID)
	req.Header.Set("X-Service-Secret", c.serviceSecret)

	// Set extracted context headers
	for key, value := range contextHeaders {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("service returned error [%d]: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// DecodeJSON is a helper to decode JSON response
func DecodeJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// DecodeStandardResponse decodes standard API response format
func DecodeStandardResponse(resp *http.Response, dataStruct interface{}) error {
	defer resp.Body.Close()

	var standardResp struct {
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
		Success bool            `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&standardResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !standardResp.Success {
		return fmt.Errorf("service error: %s", standardResp.Message)
	}

	if dataStruct != nil {
		return json.Unmarshal(standardResp.Data, dataStruct)
	}

	return nil
}
