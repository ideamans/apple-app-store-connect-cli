package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestMD5Hex(t *testing.T) {
	data := []byte("hello world")
	sum := md5.Sum(data)
	want := hex.EncodeToString(sum[:])
	if got := MD5Hex(data); got != want {
		t.Fatalf("MD5Hex = %s, want %s", got, want)
	}
}

func TestUploadReassemblesChunksWithoutAuth(t *testing.T) {
	data := []byte("0123456789ABCDEFGHIJ") // 20 bytes

	var mu sync.Mutex
	received := make([]byte, len(data))
	sawAuth := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.Header.Get("Authorization") != "" {
			sawAuth = true
		}
		body, _ := io.ReadAll(r.Body)
		off := r.URL.Query().Get("offset")
		mu.Lock()
		defer mu.Unlock()
		// offset is encoded in the query for the test
		var start int
		switch off {
		case "0":
			start = 0
		case "12":
			start = 12
		}
		copy(received[start:], body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ops := []UploadOperation{
		{Method: "PUT", URL: srv.URL + "?offset=0", Offset: 0, Length: 12,
			RequestHeaders: []UploadHeader{{Name: "Content-Type", Value: "image/png"}}},
		{Method: "PUT", URL: srv.URL + "?offset=12", Offset: 12, Length: 8},
	}

	c := New(nil)
	if err := c.Upload(context.Background(), ops, data); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if sawAuth {
		t.Error("upload requests must not carry the API Authorization header")
	}
	if string(received) != string(data) {
		t.Fatalf("reassembled = %q, want %q", received, data)
	}
}

func TestUploadRejectsOutOfRangeOperation(t *testing.T) {
	c := New(nil)
	ops := []UploadOperation{{Method: "PUT", URL: "http://example.invalid", Offset: 0, Length: 100}}
	if err := c.Upload(context.Background(), ops, []byte("short")); err == nil {
		t.Fatal("expected an out-of-range error, got nil")
	}
}

func TestDryRunSkipsUpload(t *testing.T) {
	c := New(nil)
	c.DryRun = true
	// URL is unreachable; dry-run must not attempt the request.
	ops := []UploadOperation{{Method: "PUT", URL: "http://127.0.0.1:0/nope", Offset: 0, Length: 3}}
	if err := c.Upload(context.Background(), ops, []byte("abc")); err != nil {
		t.Fatalf("dry-run Upload should be a no-op, got %v", err)
	}
}

func TestRel(t *testing.T) {
	got := string(Rel("apps", "123"))
	want := `{"data":{"id":"123","type":"apps"}}`
	if got != want {
		t.Fatalf("Rel = %s, want %s", got, want)
	}
}
