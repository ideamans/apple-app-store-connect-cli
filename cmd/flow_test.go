package cmd

import (
	"bytes"
	"encoding/json"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These tests execute real CLI commands in-process against a mock API server
// and assert the exact requests sent — regression coverage for the pitfalls
// found during real App Store submissions (feedback.md).

// TestPricingSetFree: free apps need a price schedule built from the
// customerPrice "0" price point, with ${...}-format local ids for the
// included inline-create resources (plain ids get 409 INVALID_ID).
func TestPricingSetFree(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/apps/123/appPricePoints",
		`{"data":[
		   {"type":"appPricePoints","id":"PP_099","attributes":{"customerPrice":"0.99","proceeds":"0.7"}},
		   {"type":"appPricePoints","id":"PP_FREE","attributes":{"customerPrice":"0","proceeds":"0"}}
		 ],"links":{}}`)
	ts.respond("POST /v1/appPriceSchedules", `{"data":{"type":"appPriceSchedules","id":"123"}}`)

	if err := runCommand(t, "pricing", "set", "--app", "123", "--free", "--base-territory", "USA"); err != nil {
		t.Fatal(err)
	}

	body := ts.lastBody("POST /v1/appPriceSchedules")
	if got := dig(t, body, "data", "relationships", "manualPrices", "data", 0, "id"); got != "${price1}" {
		t.Fatalf(`manualPrices local id = %v, want "${price1}"`, got)
	}
	if got := dig(t, body, "included", 0, "id"); got != "${price1}" {
		t.Fatalf(`included local id = %v, want "${price1}"`, got)
	}
	if got := dig(t, body, "included", 0, "relationships", "appPricePoint", "data", "id"); got != "PP_FREE" {
		t.Fatalf("price point id = %v, want PP_FREE (customerPrice 0)", got)
	}
	if got := dig(t, body, "data", "relationships", "baseTerritory", "data", "id"); got != "USA" {
		t.Fatalf("base territory = %v", got)
	}
}

// TestAvailabilitySetExpandsAllTerritories: the API requires a
// territoryAvailability for EVERY territory, so "JPN only" must send the full
// list with the others marked unavailable.
func TestAvailabilitySetExpandsAllTerritories(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/territories",
		`{"data":[
		   {"type":"territories","id":"FRA","attributes":{"currency":"EUR"}},
		   {"type":"territories","id":"JPN","attributes":{"currency":"JPY"}},
		   {"type":"territories","id":"USA","attributes":{"currency":"USD"}}
		 ],"links":{}}`)
	ts.respond("POST /v2/appAvailabilities", `{"data":{"type":"appAvailabilities","id":"AV1"}}`)

	if err := runCommand(t, "availability", "set", "--app", "123", "--territories", "jpn"); err != nil {
		t.Fatal(err)
	}

	body := ts.lastBody("POST /v2/appAvailabilities")
	included, ok := body["included"].([]any)
	if !ok || len(included) != 3 {
		t.Fatalf("included has %d entries, want one per territory (3)", len(included))
	}
	availByTerr := map[string]bool{}
	for i := range included {
		terr, _ := dig(t, body, "included", i, "relationships", "territory", "data", "id").(string)
		avail, _ := dig(t, body, "included", i, "attributes", "available").(bool)
		lid, _ := dig(t, body, "included", i, "id").(string)
		if !strings.HasPrefix(lid, "${ta") {
			t.Fatalf("local id %q is not ${...}-format", lid)
		}
		availByTerr[terr] = avail
	}
	if !availByTerr["JPN"] || availByTerr["USA"] || availByTerr["FRA"] {
		t.Fatalf("availability map = %v, want only JPN true", availByTerr)
	}
}

// TestAvailabilitySetRejectsUnknownTerritory guards the typo case before any
// mutation is sent.
func TestAvailabilitySetRejectsUnknownTerritory(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/territories",
		`{"data":[{"type":"territories","id":"JPN","attributes":{"currency":"JPY"}}],"links":{}}`)

	err := runCommand(t, "availability", "set", "--app", "123", "--territories", "JPX")
	if err == nil || !strings.Contains(err.Error(), "JPX") {
		t.Fatalf("want unknown-territory error, got %v", err)
	}
	if calls := ts.calls("POST /v2/appAvailabilities"); len(calls) != 0 {
		t.Fatal("mutation was sent despite the unknown territory")
	}
}

// TestVersionLocalizeSkipsWhatsNewOnFirstVersion: the first version 409s on
// whatsNew; the command must retry without it so the other attributes land.
func TestVersionLocalizeSkipsWhatsNewOnFirstVersion(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/apps/123/appStoreVersions",
		`{"data":[{"type":"appStoreVersions","id":"V1","attributes":{"versionString":"1.0","appStoreState":"PREPARE_FOR_SUBMISSION"}}],"links":{}}`)
	ts.respond("GET /v1/appStoreVersions/V1/appStoreVersionLocalizations",
		`{"data":[{"type":"appStoreVersionLocalizations","id":"L1","attributes":{"locale":"ja"}}],"links":{}}`)
	patches := 0
	ts.handle("PATCH /v1/appStoreVersionLocalizations/L1", func(w http.ResponseWriter, r *http.Request) {
		patches++
		if patches == 1 {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"errors":[{"status":"409","code":"STATE_ERROR","title":"Attribute 'whatsNew' cannot be edited at this time"}]}`))
			return
		}
		w.Write([]byte(`{"data":{"type":"appStoreVersionLocalizations","id":"L1"}}`))
	})

	err := runCommand(t, "version", "localize", "--app", "123", "--locale", "ja",
		"--description", "great app", "--whats-new", "first release")
	if err != nil {
		t.Fatalf("command should succeed by skipping whatsNew: %v", err)
	}
	if patches != 2 {
		t.Fatalf("PATCH count = %d, want 2 (409 then retry)", patches)
	}
	retry := ts.lastBody("PATCH /v1/appStoreVersionLocalizations/L1")
	attrs, _ := dig(t, retry, "data", "attributes").(map[string]any)
	if _, has := attrs["whatsNew"]; has {
		t.Fatal("retry still contains whatsNew")
	}
	if attrs["description"] != "great app" {
		t.Fatalf("retry lost description: %v", attrs)
	}
}

// TestIAPCreateFindOrCreate: product ids often already exist (sandbox
// testing); create must return the existing IAP instead of 409ing.
func TestIAPCreateFindOrCreate(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/apps/123/inAppPurchasesV2",
		`{"data":[{"type":"inAppPurchases","id":"IAP1","attributes":{"productId":"com.x.credits","state":"APPROVED"}}],"links":{}}`)

	if err := runCommand(t, "iap", "create", "--app", "123",
		"--product-id", "com.x.credits", "--name", "Credits", "--type", "CONSUMABLE"); err != nil {
		t.Fatal(err)
	}
	if calls := ts.calls("POST /v2/inAppPurchases"); len(calls) != 0 {
		t.Fatal("POST was sent even though the product id already exists")
	}
}

// TestIAPScreenshotAutoFitUploadsLegacySize: end-to-end — a modern-size image
// is auto-fitted to 1242x2208 before reserve/PUT/commit, and the delivery
// state is polled to completion.
func TestIAPScreenshotAutoFitUploadsLegacySize(t *testing.T) {
	prevInterval := assetPollInterval
	assetPollInterval = time.Millisecond
	defer func() { assetPollInterval = prevInterval }()

	ts := newTestServer(t)
	iapList := `{"data":[{"type":"inAppPurchases","id":"IAP1","attributes":{"productId":"com.x.credits","state":"MISSING_METADATA"}}],"links":{}}`
	ts.respond("GET /v1/apps/123/inAppPurchasesV2", iapList)

	var uploaded []byte
	ts.handle("PUT /upload/shot", func(w http.ResponseWriter, r *http.Request) {
		uploaded, _ = readAll(r)
		w.WriteHeader(http.StatusOK)
	})
	ts.handle("POST /v1/inAppPurchaseAppStoreReviewScreenshots", func(w http.ResponseWriter, r *http.Request) {
		body, _ := readAll(r)
		var doc map[string]any
		_ = json.Unmarshal(body, &doc)
		size := int(dig(t, doc, "data", "attributes", "fileSize").(float64))
		w.Write([]byte(`{"data":{"type":"inAppPurchaseAppStoreReviewScreenshots","id":"SS1","attributes":{"uploadOperations":[{"method":"PUT","url":"` + ts.url() + `/upload/shot","length":` + strconv.Itoa(size) + `,"offset":0,"requestHeaders":[]}]}}}`))
	})
	ts.respond("PATCH /v1/inAppPurchaseAppStoreReviewScreenshots/SS1",
		`{"data":{"type":"inAppPurchaseAppStoreReviewScreenshots","id":"SS1"}}`)
	ts.respond("GET /v1/inAppPurchaseAppStoreReviewScreenshots/SS1",
		`{"data":{"type":"inAppPurchaseAppStoreReviewScreenshots","id":"SS1","attributes":{"assetDeliveryState":{"state":"COMPLETE"}}}}`)

	shot := writeTempPNG(t, "modern.png", 1290, 2796)
	if err := runCommand(t, "iap", "screenshot", "--app", "123",
		"--product", "com.x.credits", "--file", shot, "--auto-fit"); err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(uploaded))
	if err != nil {
		t.Fatalf("uploaded bytes are not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 1242 || b.Dy() != 2208 {
		t.Fatalf("uploaded image is %dx%d, want 1242x2208 (auto-fit)", b.Dx(), b.Dy())
	}
}

// TestResolveAppIDEscapesBundleID: query metacharacters in a bundle id must
// not restructure the request.
func TestResolveAppIDEscapesBundleID(t *testing.T) {
	ts := newTestServer(t)
	ts.respond("GET /v1/apps", `{"data":[{"type":"apps","id":"999"}],"links":{}}`)

	c, err := newClient()
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveAppID(t.Context(), c, "com.example&limit=99")
	if err != nil {
		t.Fatal(err)
	}
	if id != "999" {
		t.Fatalf("id = %q", id)
	}
	reqs := ts.calls("GET /v1/apps")
	if len(reqs) != 1 {
		t.Fatalf("calls = %d", len(reqs))
	}
	if !strings.Contains(reqs[0].Query, "com.example%26limit%3D99") {
		t.Fatalf("bundle id was not escaped: %s", reqs[0].Query)
	}
}

// TestSubscriptionAvailabilityUpsert: the deprecated subscriptionAvailabilities
// family was replaced by subscriptionPlanAvailabilities — one per plan type,
// PATCHed when it already exists, POSTed otherwise.
func TestSubscriptionAvailabilityUpsert(t *testing.T) {
	ts := newTestServer(t)
	groups := `{"data":[{"type":"subscriptionGroups","id":"G1","attributes":{"referenceName":"Premium"}}],"links":{}}`
	subs := `{"data":[{"type":"subscriptions","id":"SUB1","attributes":{"productId":"com.x.pro.monthly","name":"Pro"}}],"links":{}}`
	ts.respond("GET /v1/apps/123/subscriptionGroups", groups)
	ts.respond("GET /v1/subscriptionGroups/G1/subscriptions", subs)

	// Case 1: no existing plan availability -> POST with planType + subscription rel.
	ts.respond("GET /v1/subscriptions/SUB1/planAvailabilities", `{"data":[],"links":{}}`)
	ts.respond("POST /v1/subscriptionPlanAvailabilities", `{"data":{"type":"subscriptionPlanAvailabilities","id":"PA1"}}`)
	if err := runCommand(t, "subscriptions", "availability", "set",
		"--app", "123", "--sub", "com.x.pro.monthly", "--territories", "JPN"); err != nil {
		t.Fatal(err)
	}
	created := ts.lastBody("POST /v1/subscriptionPlanAvailabilities")
	if got := dig(t, created, "data", "attributes", "planType"); got != "MONTHLY" {
		t.Fatalf("planType = %v, want MONTHLY (default)", got)
	}
	if got := dig(t, created, "data", "relationships", "subscription", "data", "id"); got != "SUB1" {
		t.Fatalf("subscription rel = %v", got)
	}

	// Case 2: existing plan availability for the plan type -> PATCH it, no new POST.
	ts.respond("GET /v1/subscriptions/SUB1/planAvailabilities",
		`{"data":[{"type":"subscriptionPlanAvailabilities","id":"PA1","attributes":{"planType":"MONTHLY"}}],"links":{}}`)
	ts.respond("PATCH /v1/subscriptionPlanAvailabilities/PA1", `{"data":{"type":"subscriptionPlanAvailabilities","id":"PA1"}}`)
	if err := runCommand(t, "subscriptions", "availability", "set",
		"--app", "123", "--sub", "com.x.pro.monthly", "--territories", "JPN,USA"); err != nil {
		t.Fatal(err)
	}
	if got := len(ts.calls("POST /v1/subscriptionPlanAvailabilities")); got != 1 {
		t.Fatalf("POST count = %d, want still 1 (second run must PATCH)", got)
	}
	patched := ts.lastBody("PATCH /v1/subscriptionPlanAvailabilities/PA1")
	terrs, _ := dig(t, patched, "data", "relationships", "availableTerritories", "data").([]any)
	if len(terrs) != 2 {
		t.Fatalf("patched territories = %d, want 2", len(terrs))
	}
}
