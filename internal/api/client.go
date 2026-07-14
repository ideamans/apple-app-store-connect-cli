package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ideamans/apple-app-store-connect/internal/auth"
	"github.com/ideamans/apple-app-store-connect/internal/config"
)

const BaseURL = "https://api.appstoreconnect.apple.com"

type Client struct {
	creds *config.Credentials
	http  *http.Client
}

func New(creds *config.Credentials) *Client {
	return &Client{
		creds: creds,
		http:  &http.Client{Timeout: 60 * time.Second},
	}
}

// Do sends a request. pathOrURL may be an API path like "/v1/apps" (query
// string allowed) or an absolute URL such as a pagination "next" link.
func (c *Client) Do(ctx context.Context, method, pathOrURL string, body io.Reader) ([]byte, error) {
	url := pathOrURL
	if strings.HasPrefix(pathOrURL, "/") {
		url = BaseURL + pathOrURL
	}
	token, err := auth.Token(c.creds, auth.DefaultTTL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return data, apiError(resp.StatusCode, data)
	}
	return data, nil
}

func apiError(status int, body []byte) error {
	var payload struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && len(payload.Errors) > 0 {
		var parts []string
		for _, e := range payload.Errors {
			msg := e.Title
			if e.Detail != "" {
				msg += ": " + e.Detail
			}
			if e.Code != "" {
				msg += " (" + e.Code + ")"
			}
			parts = append(parts, msg)
		}
		return fmt.Errorf("HTTP %d: %s", status, strings.Join(parts, "; "))
	}
	snippet := string(body)
	if len(snippet) > 500 {
		snippet = snippet[:500] + "..."
	}
	return fmt.Errorf("HTTP %d: %s", status, snippet)
}
