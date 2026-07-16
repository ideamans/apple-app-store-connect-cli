package cmd

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// makePNG encodes a solid-colored PNG of the given size.
func makePNG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeTempPNG(t *testing.T, name string, w, h int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, makePNG(t, w, h, color.RGBA{R: 255, A: 255}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImageDims(t *testing.T) {
	w, h, err := imageDims(makePNG(t, 12, 34, color.White))
	if err != nil {
		t.Fatal(err)
	}
	if w != 12 || h != 34 {
		t.Fatalf("imageDims = %dx%d, want 12x34", w, h)
	}
	if _, _, err := imageDims([]byte("not an image")); err == nil {
		t.Fatal("imageDims accepted garbage")
	}
}

func TestFitImageScalesAndPads(t *testing.T) {
	// A 100x100 red square fitted into 1242x2208 must be centered with white
	// padding above and below.
	fitted, err := fitImage(makePNG(t, 100, 100, color.RGBA{R: 255, A: 255}), 1242, 2208)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(fitted))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 1242 || b.Dy() != 2208 {
		t.Fatalf("fitted size = %dx%d, want 1242x2208", b.Dx(), b.Dy())
	}
	// Top edge should be pure white padding; the exact center should be red.
	if r, g, b, _ := img.At(621, 5).RGBA(); r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Fatalf("top padding = %d,%d,%d; want white", r>>8, g>>8, b>>8)
	}
	if r, g, b, _ := img.At(621, 1104).RGBA(); r>>8 < 200 || g>>8 > 60 || b>>8 > 60 {
		t.Fatalf("center = %d,%d,%d; want red", r>>8, g>>8, b>>8)
	}
}

func TestPrepareReviewScreenshot(t *testing.T) {
	// Accepted size passes through untouched.
	acceptedPath := writeTempPNG(t, "ok.png", 1242, 2208)
	orig, _ := os.ReadFile(acceptedPath)
	data, name, err := prepareReviewScreenshot(acceptedPath, false)
	if err != nil {
		t.Fatal(err)
	}
	if name != "ok.png" || !bytes.Equal(data, orig) {
		t.Fatal("accepted-size image should pass through unchanged")
	}

	// Unsupported size without --auto-fit errors before any upload, listing sizes.
	badPath := writeTempPNG(t, "modern.png", 1290, 2796)
	if _, _, err := prepareReviewScreenshot(badPath, false); err == nil {
		t.Fatal("expected rejection for 1290x2796")
	} else if !strings.Contains(err.Error(), "1242×2208") || !strings.Contains(err.Error(), "--auto-fit") {
		t.Fatalf("rejection should list accepted sizes and suggest --auto-fit: %v", err)
	}

	// Auto-fit portrait converts to exactly 1242x2208 PNG.
	data, name, err = prepareReviewScreenshot(badPath, true)
	if err != nil {
		t.Fatal(err)
	}
	if name != "modern.png" {
		t.Fatalf("fileName = %q, want modern.png", name)
	}
	if w, h, _ := imageDims(data); w != 1242 || h != 2208 {
		t.Fatalf("auto-fit portrait = %dx%d, want 1242x2208", w, h)
	}

	// Auto-fit landscape targets 2208x1242.
	widePath := writeTempPNG(t, "wide.png", 2796, 1290)
	data, _, err = prepareReviewScreenshot(widePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if w, h, _ := imageDims(data); w != 2208 || h != 1242 {
		t.Fatalf("auto-fit landscape = %dx%d, want 2208x1242", w, h)
	}
}

func TestAssetDeliveryInfo(t *testing.T) {
	res := &api.Resource{Attributes: map[string]any{
		"assetDeliveryState": map[string]any{
			"state": "FAILED",
			"errors": []any{
				map[string]any{"code": "IMAGE_INCORRECT_DIMENSIONS", "description": "bad size"},
				map[string]any{"code": "OTHER"},
			},
		},
	}}
	state, detail := assetDeliveryInfo(res)
	if state != "FAILED" {
		t.Fatalf("state = %q", state)
	}
	if !strings.Contains(detail, "IMAGE_INCORRECT_DIMENSIONS (bad size)") || !strings.Contains(detail, "OTHER") {
		t.Fatalf("detail = %q", detail)
	}

	// videoDeliveryState is consulted when assetDeliveryState is absent (app previews).
	res = &api.Resource{Attributes: map[string]any{
		"videoDeliveryState": map[string]any{"state": "PROCESSING"},
	}}
	if state, _ := assetDeliveryInfo(res); state != "PROCESSING" {
		t.Fatalf("video state = %q, want PROCESSING", state)
	}

	// No delivery state at all -> empty (treated as done).
	if state, _ := assetDeliveryInfo(&api.Resource{Attributes: map[string]any{}}); state != "" {
		t.Fatalf("state = %q, want empty", state)
	}
}

func TestWaitAssetDelivery(t *testing.T) {
	prevInterval, prevTimeout := assetPollInterval, assetPollTimeout
	assetPollInterval, assetPollTimeout = time.Millisecond, time.Second
	defer func() { assetPollInterval, assetPollTimeout = prevInterval, prevTimeout }()

	ts := newTestServer(t)
	c, err := newClient()
	if err != nil {
		t.Fatal(err)
	}

	// PROCESSING twice, then COMPLETE.
	polls := 0
	ts.handle("GET /v1/appScreenshots/S1", func(w http.ResponseWriter, r *http.Request) {
		polls++
		state := "PROCESSING"
		if polls >= 3 {
			state = "COMPLETE"
		}
		w.Write([]byte(`{"data":{"type":"appScreenshots","id":"S1","attributes":{"assetDeliveryState":{"state":"` + state + `"}}}}`))
	})
	if err := waitAssetDelivery(context.Background(), c, "/v1/appScreenshots/S1"); err != nil {
		t.Fatalf("COMPLETE path errored: %v", err)
	}
	if polls < 3 {
		t.Fatalf("polled %d times, want >= 3", polls)
	}

	// FAILED surfaces the error codes and fails the command.
	ts.respond("GET /v1/appScreenshots/S2",
		`{"data":{"type":"appScreenshots","id":"S2","attributes":{"assetDeliveryState":{"state":"FAILED","errors":[{"code":"IMAGE_INCORRECT_DIMENSIONS","description":"1290x2796 not allowed"}]}}}}`)
	err = waitAssetDelivery(context.Background(), c, "/v1/appScreenshots/S2")
	if err == nil || !strings.Contains(err.Error(), "IMAGE_INCORRECT_DIMENSIONS") {
		t.Fatalf("FAILED path should error with the code, got: %v", err)
	}
}

func TestUploadAssetBytesFullFlow(t *testing.T) {
	prevInterval := assetPollInterval
	assetPollInterval = time.Millisecond
	defer func() { assetPollInterval = prevInterval }()

	ts := newTestServer(t)
	c, err := newClient()
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("fake image bytes")
	var uploaded []byte
	ts.handle("PUT /upload/1", func(w http.ResponseWriter, r *http.Request) {
		uploaded, _ = readAll(r)
		w.WriteHeader(http.StatusOK)
	})
	ts.respond("POST /v1/appScreenshots",
		`{"data":{"type":"appScreenshots","id":"NEW1","attributes":{"uploadOperations":[{"method":"PUT","url":"`+ts.url()+`/upload/1","length":16,"offset":0,"requestHeaders":[]}]}}}`)
	ts.respond("PATCH /v1/appScreenshots/NEW1", `{"data":{"type":"appScreenshots","id":"NEW1"}}`)
	ts.respond("GET /v1/appScreenshots/NEW1",
		`{"data":{"type":"appScreenshots","id":"NEW1","attributes":{"assetDeliveryState":{"state":"COMPLETE"}}}}`)

	id, err := uploadAssetBytes(context.Background(), c, assetSpec{
		reserveType: "appScreenshots",
		relName:     "appScreenshotSet",
		relType:     "appScreenshotSets",
		relID:       "SET1",
	}, payload, "shot.png")
	if err != nil {
		t.Fatal(err)
	}
	if id != "NEW1" {
		t.Fatalf("id = %q", id)
	}
	if !bytes.Equal(uploaded, payload) {
		t.Fatalf("uploaded bytes = %q, want %q", uploaded, payload)
	}

	reserve := ts.lastBody("POST /v1/appScreenshots")
	if got := dig(t, reserve, "data", "attributes", "fileName"); got != "shot.png" {
		t.Fatalf("reserve fileName = %v", got)
	}
	if got := dig(t, reserve, "data", "relationships", "appScreenshotSet", "data", "id"); got != "SET1" {
		t.Fatalf("reserve relationship id = %v", got)
	}
	commit := ts.lastBody("PATCH /v1/appScreenshots/NEW1")
	if got := dig(t, commit, "data", "attributes", "uploaded"); got != true {
		t.Fatalf("commit uploaded = %v", got)
	}
	if got := dig(t, commit, "data", "attributes", "sourceFileChecksum"); got != api.MD5Hex(payload) {
		t.Fatalf("commit checksum = %v", got)
	}
}
