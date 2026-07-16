package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ideamans/apple-app-store-connect-cli/internal/auth"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

// BaseURL is the API origin. It is a variable so tests can point the client at
// a local mock server.
var BaseURL = "https://api.appstoreconnect.apple.com"

type Client struct {
	creds *config.Credentials
	http  *http.Client
	// DryRun, when true, makes mutating helpers (Post/Patch/Delete/Upload) print
	// the intended request to stderr and skip it. Reads still execute.
	DryRun bool
}

func New(creds *config.Credentials) *Client {
	return &Client{
		creds: creds,
		http:  &http.Client{Timeout: 120 * time.Second},
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

// Download performs a GET and returns the raw response body, for endpoints
// that do not speak JSON:API (gzipped sales/finance reports, CSV code files,
// certificate content, analytics segment URLs). accept, when non-empty, is
// sent as the Accept header. pathOrURL may be an API path or an absolute URL;
// the bearer token is only attached for the API host.
func (c *Client) Download(ctx context.Context, pathOrURL, accept string) ([]byte, error) {
	url := pathOrURL
	if strings.HasPrefix(pathOrURL, "/") {
		url = BaseURL + pathOrURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, BaseURL) {
		token, err := auth.Token(c.creds, auth.DefaultTTL)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
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

// Error is an API error that carries the HTTP status code.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

// IsNotFound reports whether err is an API error with HTTP 404.
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Status == http.StatusNotFound
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
		msg := fmt.Sprintf("HTTP %d: %s", status, strings.Join(parts, "; "))
		if status == http.StatusUnauthorized {
			msg += "\nHint: the management API needs an App Store Connect API key with a role (Admin/App Manager). In-App Purchase / App Store Server API keys from the same issuer get 401 here."
		}
		return &Error{Status: status, Message: msg}
	}
	snippet := string(body)
	if len(snippet) > 500 {
		snippet = snippet[:500] + "..."
	}
	return &Error{Status: status, Message: fmt.Sprintf("HTTP %d: %s", status, snippet)}
}
