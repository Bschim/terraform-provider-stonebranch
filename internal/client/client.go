package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the StoneBranch API client.
type Client struct {
	BaseURL    string
	APIToken   string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// NewClient creates a new StoneBranch API client authenticating with a
// Bearer token.
func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewBasicAuthClient creates a new StoneBranch API client authenticating
// with HTTP Basic Auth instead of a Bearer token, for UAC instances fronted
// by a reverse proxy that only accepts Basic credentials.
func NewBasicAuthClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIError represents an error returned by the StoneBranch API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, query url.Values, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	reqURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	if query != nil && len(query) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, query.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, endpoint, query, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, endpoint string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, endpoint, nil, body)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, endpoint string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, endpoint, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, endpoint, query, nil)
}
