# Known pitfalls

Learned from real App Store submissions. None of these are inferable from the
API documentation or the command list — read this chapter before driving a
submission end to end.

- **Asset uploads validate asynchronously.** A 2xx on reserve/PUT/commit does
  **not** mean acceptance; failures (e.g. `IMAGE_INCORRECT_DIMENSIONS`) appear
  later in `assetDeliveryState`. asc's image/video upload commands poll for
  COMPLETE/FAILED automatically and exit non-zero on FAILED with the error
  codes. Build and Background Asset uploads use different state models — watch
  them with `asc builds list` / `asc background-assets versions`.

- **IAP/subscription review screenshots only accept legacy sizes**
  (1242x2208, 2208x1242, 2048x2732, 2732x2048). Current App Store screenshot
  sizes (1290x2796 etc.) are rejected — the appScreenshots validator and the
  review-screenshot validator are different. `asc iap screenshot` /
  `asc subscriptions screenshot` validate dimensions before uploading and offer
  `--auto-fit` (aspect-preserving resize + white padding).

- **`whatsNew` is not editable on an app's first version** (409).
  `asc version localize` skips it with a warning and still applies other
  attributes.

- **IAPs often already exist** (created for sandbox testing): POST 409s on a
  duplicate productId. `asc iap create` is find-or-create; localize / price /
  screenshot address products by productId and work on existing ones.

- **Inline-created ("included") resources need local ids in the form
  `${name}`** — a plain string id gets `409 ENTITY_ERROR.INCLUDED.INVALID_ID`
  (appPriceSchedules, appAvailabilities, offer codes, …). asc does this
  internally; remember it when composing raw `asc api` POSTs.

- **App availability requires ALL territories explicitly** (~175
  territoryAvailabilities in one POST, `available` true/false each). "Japan
  only" still needs every other territory marked false. `asc availability set`
  expands the full list automatically.

- **Free apps still need a price schedule** (customerPrice "0" price point):
  `asc pricing set --free`. Price and availability are required for submission
  yet remain unset even when the version page looks complete; `asc submit` warns
  when either is missing.

- **The management API needs an ASC API key with a role.** App Store Server API
  (In-App Purchase) keys from the same issuer return 401 here.

- **The newer age-rating flow may not be readable via the API** (GET 404 even
  when set); writes work — verify the rating in the UI.

- **GUI-only (no public API):** the App Privacy questionnaire (data collection
  labels; `/v1/apps/{id}/dataUsages` 404s), App Store Server Notifications URL
  configuration, and Japan's 特定商取引法 disclosure (no dedicated field; the App
  Info "trader status" field is the EU DSA one).
