package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RoadieClient struct {
	BaseURL     string
	Token       string
	WorkspaceID string
	HTTPClient  *http.Client
	UserAgent   string
}

func New(baseURL, token, workspaceID, version string) *RoadieClient {
	return &RoadieClient{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Token:       token,
		WorkspaceID: workspaceID,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		UserAgent:   "terraform-provider-roadie/" + version,
	}
}

func (c *RoadieClient) doRequest(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if c.WorkspaceID != "" {
		req.Header.Set("x-openroadie-workspace-id", c.WorkspaceID)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, parseAPIError(resp.StatusCode, respBody)
	}

	return respBody, resp.StatusCode, nil
}

func (c *RoadieClient) Get(ctx context.Context, path string) ([]byte, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, path, nil)
	return body, err
}

func (c *RoadieClient) Post(ctx context.Context, path string, body any) ([]byte, error) {
	respBody, _, err := c.doRequest(ctx, http.MethodPost, path, body)
	return respBody, err
}

func (c *RoadieClient) Put(ctx context.Context, path string, body any) ([]byte, error) {
	respBody, _, err := c.doRequest(ctx, http.MethodPut, path, body)
	return respBody, err
}

func (c *RoadieClient) Patch(ctx context.Context, path string, body any) ([]byte, error) {
	respBody, _, err := c.doRequest(ctx, http.MethodPatch, path, body)
	return respBody, err
}

func (c *RoadieClient) Delete(ctx context.Context, path string) error {
	_, statusCode, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	_ = statusCode
	return nil
}
