package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Resource is a JSON:API primary resource object (the "data" element).
type Resource struct {
	Type          string                     `json:"type"`
	ID            string                     `json:"id,omitempty"`
	Attributes    map[string]any             `json:"attributes,omitempty"`
	Relationships map[string]json.RawMessage `json:"relationships,omitempty"`
}

// Str returns a string attribute, or "" when absent or not a string.
func (r *Resource) Str(key string) string {
	if r == nil {
		return ""
	}
	if v, ok := r.Attributes[key].(string); ok {
		return v
	}
	return ""
}

// DecodeAttr re-decodes a single attribute into v (for nested structures like
// uploadOperations that arrive inside the generic attributes map).
func (r *Resource) DecodeAttr(key string, v any) error {
	raw, ok := r.Attributes[key]
	if !ok {
		return fmt.Errorf("no attribute %q on %s", key, r.Type)
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

type singleDoc struct {
	Data     Resource   `json:"data"`
	Included []Resource `json:"included"`
}

type listDoc struct {
	Data     []Resource `json:"data"`
	Included []Resource `json:"included"`
	Links    struct {
		Next string `json:"next"`
	} `json:"links"`
}

// Get fetches a single resource (and any included resources).
func (c *Client) Get(ctx context.Context, path string) (*Resource, []Resource, error) {
	data, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	var doc singleDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, err
	}
	return &doc.Data, doc.Included, nil
}

// GetOptional is like Get but returns (nil, nil) on a 404 instead of an error,
// which the API uses for absent to-one relationships (e.g. an unset review detail).
func (c *Client) GetOptional(ctx context.Context, path string) (*Resource, error) {
	data, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	var doc singleDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc.Data, nil
}

// List fetches a collection, following pagination to the end.
func (c *Client) List(ctx context.Context, path string) ([]Resource, error) {
	var out []Resource
	next := path
	for next != "" {
		data, err := c.Do(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, err
		}
		var doc listDoc
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		out = append(out, doc.Data...)
		next = doc.Links.Next
	}
	return out, nil
}

// Body is a JSON:API request document with a single primary resource.
type Body struct {
	Data     Resource   `json:"data"`
	Included []Resource `json:"included,omitempty"`
}

// Post creates a resource. In dry-run mode it prints the request and returns a
// stub resource without calling the API.
func (c *Client) Post(ctx context.Context, path string, body Body) (*Resource, error) {
	return c.mutate(ctx, http.MethodPost, path, body)
}

// Patch updates a resource.
func (c *Client) Patch(ctx context.Context, path string, body Body) (*Resource, error) {
	return c.mutate(ctx, http.MethodPatch, path, body)
}

// Delete removes a resource.
func (c *Client) Delete(ctx context.Context, path string) error {
	if c.DryRun {
		fmt.Fprintf(os.Stderr, "DRY-RUN DELETE %s\n", path)
		return nil
	}
	_, err := c.Do(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *Client) mutate(ctx context.Context, method, path string, body Body) (*Resource, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	if c.DryRun {
		pretty, _ := json.MarshalIndent(body, "", "  ")
		fmt.Fprintf(os.Stderr, "DRY-RUN %s %s\n%s\n", method, path, pretty)
		return &Resource{Type: body.Data.Type, ID: "dry-run"}, nil
	}
	data, err := c.Do(ctx, method, path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return &Resource{Type: body.Data.Type}, nil
	}
	var doc singleDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc.Data, nil
}

// Rel builds a to-one relationship value: {"data":{"type":..,"id":..}}.
func Rel(typ, id string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"data": map[string]string{"type": typ, "id": id}})
	return b
}

// --- Asset upload transport ---------------------------------------------------

// UploadOperation is one chunk instruction returned in an asset reservation's
// uploadOperations attribute.
type UploadOperation struct {
	Method         string         `json:"method"`
	URL            string         `json:"url"`
	Length         int            `json:"length"`
	Offset         int            `json:"offset"`
	RequestHeaders []UploadHeader `json:"requestHeaders"`
}

type UploadHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MD5Hex returns the hex-encoded MD5 checksum the API expects in
// sourceFileChecksum when committing an uploaded asset.
func MD5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// Upload sends the file bytes to the reserved upload operations. Each operation
// targets a pre-signed URL on a different host and carries its own headers, so
// it must NOT include the API bearer token or the JSON content type.
func (c *Client) Upload(ctx context.Context, ops []UploadOperation, data []byte) error {
	for i, op := range ops {
		if err := c.uploadChunk(ctx, op, data); err != nil {
			return fmt.Errorf("upload operation %d/%d: %w", i+1, len(ops), err)
		}
	}
	return nil
}

func (c *Client) uploadChunk(ctx context.Context, op UploadOperation, data []byte) error {
	if op.Offset+op.Length > len(data) {
		return fmt.Errorf("operation range [%d:%d] exceeds file size %d", op.Offset, op.Offset+op.Length, len(data))
	}
	if c.DryRun {
		fmt.Fprintf(os.Stderr, "DRY-RUN %s %s (%d bytes @ offset %d)\n", op.Method, op.URL, op.Length, op.Offset)
		return nil
	}
	chunk := data[op.Offset : op.Offset+op.Length]
	req, err := http.NewRequestWithContext(ctx, op.Method, op.URL, bytes.NewReader(chunk))
	if err != nil {
		return err
	}
	for _, h := range op.RequestHeaders {
		if h.Value != "" {
			req.Header.Set(h.Name, h.Value)
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		snippet := string(body)
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
	}
	return nil
}
