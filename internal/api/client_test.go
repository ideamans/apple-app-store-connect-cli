package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

// newTestClient points BaseURL at a mock server and returns a client with a
// freshly generated throwaway key.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	prev := BaseURL
	BaseURL = srv.URL
	t.Cleanup(func() { BaseURL = prev })

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return New(&config.Credentials{
		IssuerID:      "00000000-0000-0000-0000-000000000000",
		KeyID:         "TESTKEY123",
		PrivateKeyPEM: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}),
	})
}

func TestListFollowsPagination(t *testing.T) {
	var c *Client
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/things", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "2" {
			w.Write([]byte(`{"data":[{"type":"things","id":"B"}],"links":{}}`))
			return
		}
		w.Write([]byte(`{"data":[{"type":"things","id":"A"}],"links":{"next":"` + BaseURL + `/v1/things?cursor=2"}}`))
	})
	c = newTestClient(t, mux)

	items, err := c.List(context.Background(), "/v1/things")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != "A" || items[1].ID != "B" {
		t.Fatalf("items = %+v, want A then B across two pages", items)
	}
}

// TestGetOptionalNullData documents the live-API quirk: to-one relationship
// endpoints answer 200 with "data": null, which yields a NON-nil resource with
// an empty ID — callers must guard res.ID == "".
func TestGetOptionalNullData(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps/1/thing", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":null,"links":{}}`))
	})
	mux.HandleFunc("/v1/apps/1/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"status":"404","code":"NOT_FOUND","title":"nope"}]}`))
	})
	c := newTestClient(t, mux)

	res, err := c.GetOptional(context.Background(), "/v1/apps/1/thing")
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Fatal("200-with-null yields a non-nil zero resource by contract; if this changed, revisit every res.ID == \"\" guard")
	}
	if res.ID != "" {
		t.Fatalf("res.ID = %q, want empty", res.ID)
	}

	res, err = c.GetOptional(context.Background(), "/v1/apps/1/missing")
	if err != nil || res != nil {
		t.Fatalf("404 should give (nil, nil), got (%v, %v)", res, err)
	}
}

func TestUnauthorizedErrorCarriesRoleHint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"status":"401","code":"NOT_AUTHORIZED","title":"Authentication credentials are missing or invalid."}]}`))
	})
	c := newTestClient(t, mux)

	_, err := c.List(context.Background(), "/v1/apps")
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "App Store Connect API key with a role") {
		t.Fatalf("401 error should carry the key-role hint, got: %v", err)
	}
}

// TestDownloadBearerOnlyOnAPIHost: pre-signed segment/artifact URLs live on
// other hosts and must not receive the API bearer token.
func TestDownloadBearerOnlyOnAPIHost(t *testing.T) {
	var apiAuth, extAuth string
	ext := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extAuth = r.Header.Get("Authorization")
		w.Write([]byte("external-bytes"))
	}))
	defer ext.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/report", func(w http.ResponseWriter, r *http.Request) {
		apiAuth = r.Header.Get("Authorization")
		if r.Header.Get("Accept") != "application/a-gzip" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Write([]byte("api-bytes"))
	})
	c := newTestClient(t, mux)

	if data, err := c.Download(context.Background(), "/v1/report", "application/a-gzip"); err != nil || string(data) != "api-bytes" {
		t.Fatalf("api download = %q, %v", data, err)
	}
	if !strings.HasPrefix(apiAuth, "Bearer ") {
		t.Fatalf("API host request lacked bearer token: %q", apiAuth)
	}

	if data, err := c.Download(context.Background(), ext.URL+"/blob", ""); err != nil || string(data) != "external-bytes" {
		t.Fatalf("external download = %q, %v", data, err)
	}
	if extAuth != "" {
		t.Fatalf("external host received Authorization %q; token must not leak", extAuth)
	}
}

func TestDryRunSkipsMutations(t *testing.T) {
	hit := false
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { hit = true })
	c := newTestClient(t, mux)
	c.DryRun = true

	if _, err := c.Post(context.Background(), "/v1/things", Body{Data: Resource{Type: "things"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Patch(context.Background(), "/v1/things/1", Body{Data: Resource{Type: "things", ID: "1"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(context.Background(), "/v1/things/1"); err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("dry-run sent a request to the server")
	}
}
