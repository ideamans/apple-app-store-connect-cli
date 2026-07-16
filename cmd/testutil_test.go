package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// testServer wires a mock App Store Connect API: it records every request and
// serves canned JSON:API responses per "METHOD path" key. Handlers registered
// with handle() win over responses.
type testServer struct {
	t         *testing.T
	mu        sync.Mutex
	requests  []recordedRequest
	responses map[string]string
	handlers  map[string]http.HandlerFunc
	srv       *httptest.Server
}

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   []byte
}

// newTestServer starts the mock API, points api.BaseURL at it, and installs
// throwaway env credentials so newClient() works without a config file.
func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{
		t:         t,
		responses: map[string]string{},
		handlers:  map[string]http.HandlerFunc{},
	}
	ts.srv = httptest.NewServer(http.HandlerFunc(ts.serve))
	t.Cleanup(ts.srv.Close)

	prevBase := api.BaseURL
	api.BaseURL = ts.srv.URL
	t.Cleanup(func() { api.BaseURL = prevBase })

	setTestCredentials(t)
	return ts
}

// url returns the mock server origin (for upload operations etc.).
func (ts *testServer) url() string { return ts.srv.URL }

// respond registers a canned response body for "METHOD /path" (query ignored).
func (ts *testServer) respond(methodAndPath, body string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.responses[methodAndPath] = body
}

// handle registers a custom handler for "METHOD /path".
func (ts *testServer) handle(methodAndPath string, h http.HandlerFunc) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.handlers[methodAndPath] = h
}

func (ts *testServer) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := readAll(r)
	// Recording consumed the body; restore it for custom handlers.
	r.Body = io.NopCloser(bytes.NewReader(body))
	key := r.Method + " " + r.URL.Path
	ts.mu.Lock()
	ts.requests = append(ts.requests, recordedRequest{
		Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Body: body,
	})
	h := ts.handlers[key]
	resp, ok := ts.responses[key]
	ts.mu.Unlock()
	if h != nil {
		h(w, r)
		return
	}
	if !ok {
		ts.t.Errorf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"status":"404","code":"NOT_FOUND","title":"not stubbed"}]}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(resp))
}

// calls returns the recorded requests matching "METHOD /path".
func (ts *testServer) calls(methodAndPath string) []recordedRequest {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	var out []recordedRequest
	for _, req := range ts.requests {
		if req.Method+" "+req.Path == methodAndPath {
			out = append(out, req)
		}
	}
	return out
}

// lastBody unmarshals the last request body for "METHOD /path" into a map.
func (ts *testServer) lastBody(methodAndPath string) map[string]any {
	ts.t.Helper()
	reqs := ts.calls(methodAndPath)
	if len(reqs) == 0 {
		ts.t.Fatalf("no request recorded for %s", methodAndPath)
	}
	var doc map[string]any
	if err := json.Unmarshal(reqs[len(reqs)-1].Body, &doc); err != nil {
		ts.t.Fatalf("unmarshal body of %s: %v", methodAndPath, err)
	}
	return doc
}

func readAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 4096)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			return buf, nil
		}
	}
}

// setTestCredentials generates a throwaway P-256 key and points the env-based
// credential resolution at it (issuer id + AuthKey_<KEYID>.p8 file).
func setTestCredentials(t *testing.T) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	path := filepath.Join(t.TempDir(), "AuthKey_TESTKEY123.p8")
	if err := os.WriteFile(path, pemData, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASC_ISSUER_ID", "00000000-0000-0000-0000-000000000000")
	t.Setenv("ASC_PRIVATE_KEY_PATH", path)
	t.Setenv("ASC_KEY_ID", "TESTKEY123")
}

// runCommand executes the CLI in-process (asc <args...>) and returns the error.
// Flag state persists on the package-level cobra commands between Execute
// calls (slice flags even append), so everything is reset to defaults first.
func runCommand(t *testing.T, args ...string) error {
	t.Helper()
	resetFlags(rootCmd)
	rootCmd.SetArgs(args)
	defer rootCmd.SetArgs(nil)
	return rootCmd.Execute()
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

// dig walks a decoded JSON document by keys/indexes.
func dig(t *testing.T, doc any, path ...any) any {
	t.Helper()
	cur := doc
	for _, step := range path {
		switch key := step.(type) {
		case string:
			m, ok := cur.(map[string]any)
			if !ok {
				t.Fatalf("dig %v: not an object at %v", path, step)
			}
			cur = m[key]
		case int:
			arr, ok := cur.([]any)
			if !ok || key >= len(arr) {
				t.Fatalf("dig %v: not an array (or too short) at %v", path, step)
			}
			cur = arr[key]
		}
	}
	return cur
}
