# Command catalog

Generated from the cobra command tree by `go generate ./...`.
Do not edit by hand — edit the command definitions instead.

## Global flags

| flag | type | default | description |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | print mutating requests instead of sending them |
| `--profile` | string | — | profile name (env: ASC_PROFILE) |

## `asc accessibility`

Manage accessibility declarations (Accessibility Nutrition Labels)

### `asc accessibility create`

Create an accessibility declaration for an app and device family

Create an accessibility declaration. --device-family is one of IPHONE, IPAD,
APPLE_TV, APPLE_WATCH, MAC, VISION. Pass any of the supports flags (e.g.
--voiceover=true) to declare supported accessibility features; omitted flags
are left unset. Publish it afterwards with "asc accessibility update --publish".

Example:

```
asc accessibility create --app com.example.app --device-family IPHONE \
    --voiceover=true --larger-text=true --dark-interface=true
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--audio-descriptions` | bool | `false` | declare support: supportsAudioDescriptions |
| `--captions` | bool | `false` | declare support: supportsCaptions |
| `--dark-interface` | bool | `false` | declare support: supportsDarkInterface |
| `--device-family` | string | — | IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, or VISION (required) |
| `--differentiate-without-color` | bool | `false` | declare support: supportsDifferentiateWithoutColorAlone |
| `--larger-text` | bool | `false` | declare support: supportsLargerText |
| `--reduced-motion` | bool | `false` | declare support: supportsReducedMotion |
| `--sufficient-contrast` | bool | `false` | declare support: supportsSufficientContrast |
| `--voice-control` | bool | `false` | declare support: supportsVoiceControl |
| `--voiceover` | bool | `false` | declare support: supportsVoiceover |

### `asc accessibility delete`

Delete an accessibility declaration

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | accessibilityDeclaration id (required) |

### `asc accessibility list`

List accessibility declarations of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc accessibility show`

Show an accessibility declaration

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | accessibilityDeclaration id (required) |

### `asc accessibility update`

Update (and optionally publish) an accessibility declaration

Update the supports flags of a DRAFT accessibility declaration. Only the flags
you pass are changed. Pass --publish to publish the declaration to the App Store.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--audio-descriptions` | bool | `false` | declare support: supportsAudioDescriptions |
| `--captions` | bool | `false` | declare support: supportsCaptions |
| `--dark-interface` | bool | `false` | declare support: supportsDarkInterface |
| `--differentiate-without-color` | bool | `false` | declare support: supportsDifferentiateWithoutColorAlone |
| `--id` | string | — | accessibilityDeclaration id (required) |
| `--larger-text` | bool | `false` | declare support: supportsLargerText |
| `--publish` | bool | `false` | publish the declaration to the App Store |
| `--reduced-motion` | bool | `false` | declare support: supportsReducedMotion |
| `--sufficient-contrast` | bool | `false` | declare support: supportsSufficientContrast |
| `--voice-control` | bool | `false` | declare support: supportsVoiceControl |
| `--voiceover` | bool | `false` | declare support: supportsVoiceover |

## `asc actors`

Look up actors (users, API keys, Xcode Cloud, Apple) referenced by other resources

Actors identify who performed an action (e.g. in app store version release
history). The list endpoint requires explicit ids, so pass the actor ids you
found on other resources.

### `asc actors list`

List actors by id (the API requires an id filter)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--ids` | stringSlice | `[]` | comma-separated actor ids (required) |

### `asc actors show`

Show an actor

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | actor id (required) |

## `asc alt-distribution`

Manage EU alternative distribution keys, domains, and packages

### `asc alt-distribution deltas`

List deltas of an alternative distribution package version

List the deltas of a package version with their download URL, expiration
date, and file checksum. Use --json to include the full resources (e.g. the
alternativeDistributionKeyBlob, which is too large for the table).

| flag | type | default | description |
| --- | --- | --- | --- |
| `--json` | bool | `false` | print the full resources as JSON instead of a table |
| `--version-id` | string | — | alternativeDistributionPackageVersion id (required) |

### `asc alt-distribution domains`

Alternative distribution domains (web distribution)

#### `asc alt-distribution domains create`

Register an alternative distribution domain

Example:

```
asc alt-distribution domains create --domain example.com --reference-name production
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--domain` | string | — | domain name, e.g. example.com (required) |
| `--reference-name` | string | — | reference name for the domain (required) |

#### `asc alt-distribution domains delete`

Delete an alternative distribution domain

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | alternativeDistributionDomain id (required) |

#### `asc alt-distribution domains list`

List alternative distribution domains

### `asc alt-distribution keys`

Alternative distribution keys (marketplace/web distribution public keys)

#### `asc alt-distribution keys create`

Register an alternative distribution public key

Example:

```
asc alt-distribution keys create --public-key @public_key.pem --app com.example.app
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id to associate the key with (optional) |
| `--public-key` | string | — | PEM public key, or @file to read it from a file (required) |

#### `asc alt-distribution keys delete`

Delete an alternative distribution key

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | alternativeDistributionKey id (required) |

#### `asc alt-distribution keys list`

List alternative distribution keys

### `asc alt-distribution packages`

Alternative distribution packages of App Store versions

#### `asc alt-distribution packages show`

Show the alternative distribution package of an App Store version (and its versions)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | appStoreVersion id (required) |

### `asc alt-distribution variants`

List variants of an alternative distribution package version

List the variants of a package version with their download URL, expiration
date, and file checksum. Use --json to include the full resources (e.g. the
alternativeDistributionKeyBlob, which is too large for the table).

| flag | type | default | description |
| --- | --- | --- | --- |
| `--json` | bool | `false` | print the full resources as JSON instead of a table |
| `--version-id` | string | — | alternativeDistributionPackageVersion id (required) |

### `asc alt-distribution versions`

List versions of an alternative distribution package (newest first)

List the versions of an alternative distribution package. Find package ids
with "asc alt-distribution packages show --version-id <appStoreVersion-id>".

| flag | type | default | description |
| --- | --- | --- | --- |
| `--limit` | int | `50` | maximum number of versions to list (max 200) |
| `--package-id` | string | — | alternativeDistributionPackage id (required) |

## `asc analytics`

Request and download App Store analytics reports

Work with the Analytics Reports API. The flow is:

  1. analytics request    -- create a report request for an app
  2. analytics requests   -- find the request id
  3. analytics reports    -- list reports available for a request
  4. analytics instances  -- list a report's instances (per processing date)
  5. analytics segments   -- list or download an instance's data segments (.csv.gz)

### `asc analytics instances`

List instances of an analytics report

| flag | type | default | description |
| --- | --- | --- | --- |
| `--granularity` | string | — | filter by granularity: DAILY, WEEKLY, MONTHLY |
| `--report-id` | string | — | analytics report id (required) |

### `asc analytics reports`

List reports available for an analytics report request

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | string | — | filter by category: APP_USAGE, APP_STORE_ENGAGEMENT, COMMERCE, FRAMEWORK_USAGE, PERFORMANCE |
| `--request-id` | string | — | analytics report request id (required) |

### `asc analytics request`

Create an analytics report request for an app

Create an analytics report request (POST /v1/analyticsReportRequests).

--access-type ONGOING generates reports continuously; ONE_TIME_SNAPSHOT
generates historical data once. Reports become available asynchronously.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--access-type` | string | `ONGOING` | ONGOING or ONE_TIME_SNAPSHOT |
| `--app` | string | — | app id or bundle id (required) |

### `asc analytics requests`

List an app's analytics report requests

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc analytics segments`

List or download an analytics report instance's data segments

List the data segments of a report instance. Without --output-dir the
segment URLs and sizes are printed; with --output-dir each segment is
downloaded to <dir>/<checksum>.csv.gz (the files are gzip-compressed CSV).

| flag | type | default | description |
| --- | --- | --- | --- |
| `--instance-id` | string | — | analytics report instance id (required) |
| `--output-dir` | string | — | download segments into this directory (list only when omitted) |

## `asc android-mapping`

Map Android apps to an iOS app (androidToIosAppMappingDetails)

Manage the Android-to-iOS app mapping details of an app. Each mapping names
an Android package and the SHA-256 fingerprints of its app signing key's
public certificates. An app can have multiple mappings (one per Android package).

### `asc android-mapping delete`

Delete an Android-to-iOS app mapping

Delete a mapping by --id, or by --app (plus --package-name when the app has
more than one mapping).

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id |
| `--id` | string | — | androidToIosAppMappingDetail id |
| `--package-name` | string | — | Android package name to delete (when the app has several mappings) |

### `asc android-mapping set`

Create or update the mapping for an Android package name

Set the Android-to-iOS mapping of an app for one Android package. If the app
already has a mapping with the same package name its fingerprints are updated;
otherwise a new mapping is created.

Example:

```
asc android-mapping set --app com.example.app \
    --package-name com.example.android \
    --fingerprints AA:BB:...,CC:DD:...   # or --fingerprints @fingerprints.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--fingerprints` | string | — | SHA-256 fingerprints of the Android app signing key's public certificates, comma-separated or @file (required) |
| `--package-name` | string | — | Android package name (required) |

### `asc android-mapping show`

Show the Android-to-iOS app mappings of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

## `asc api`

Send a raw request to the App Store Connect API

Sends an authenticated request to https://api.appstoreconnect.apple.com
and prints the JSON response. Use this for endpoints not yet covered by a
dedicated subcommand.

```
asc api <path>
```

Example:

```
asc api /v1/apps
  asc api "/v1/apps?filter[bundleId]=com.example.app"
  asc api -X POST /v1/betaGroups -d '{"data": {...}}'
  asc api -X POST /v1/betaGroups -d @payload.json
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `-d`, `--data` | string | — | JSON request body, or @file to read from a file |
| `-X`, `--method` | string | `GET` | HTTP method |

## `asc app-tags`

App Store tags assigned to an app

### `asc app-tags list`

List App Store tags of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

## `asc appclips`

Manage App Clips (default experiences, localizations, header images, review details)

### `asc appclips advanced`

Advanced App Clip experiences (read-only)

Advanced App Clip experiences. Creating one requires header image and
localization relationships in a single request (plus optional place data), which
is beyond this command's scope — use "asc api" with a hand-written body for
creation. Listing is supported here.

#### `asc appclips advanced list`

List an App Clip's advanced experiences

| flag | type | default | description |
| --- | --- | --- | --- |
| `--clip-id` | string | — | appClip id (see: asc appclips list) (required) |

### `asc appclips create-experience`

Create a default experience for an App Clip

Example:

```
asc appclips create-experience --clip-id <appClip id> --action OPEN \
    --release-version-id <appStoreVersion id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action` | string | — | invocation action: OPEN, VIEW or PLAY |
| `--clip-id` | string | — | appClip id (see: asc appclips list) (required) |
| `--release-version-id` | string | — | appStoreVersion id to release the experience with |

### `asc appclips experiences`

List an App Clip's default experiences

| flag | type | default | description |
| --- | --- | --- | --- |
| `--clip-id` | string | — | appClip id (see: asc appclips list) (required) |

### `asc appclips list`

List the app's App Clips

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc appclips localize`

Set a default experience's localized subtitle

Example:

```
asc appclips localize --experience-id <id> --locale ja --subtitle "その場ですぐ注文"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--experience-id` | string | — | appClipDefaultExperience id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--subtitle` | string | — | App Clip card subtitle (@file allowed) (required) |

### `asc appclips review-detail`

Show or set a default experience's App Store review detail (invocation URLs)

#### `asc appclips review-detail set`

Set the invocation URLs on a default experience's review detail

Example:

```
asc appclips review-detail set --experience-id <id> \
    --invocation-urls "https://example.com/clip/a,https://example.com/clip/b"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--experience-id` | string | — | appClipDefaultExperience id (required) |
| `--invocation-urls` | stringSlice | `[]` | comma-separated invocation URLs for review (required) |

#### `asc appclips review-detail show`

Show the review detail for a default experience

| flag | type | default | description |
| --- | --- | --- | --- |
| `--experience-id` | string | — | appClipDefaultExperience id (required) |

### `asc appclips upload-header`

Upload the header image for a default experience localization

Upload an App Clip card header image (reserve → upload → commit) and attach
it to a default experience localization. Find the localization id via:
GET /v1/appClipDefaultExperiences/{id}/appClipDefaultExperienceLocalizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | header image file (required) |
| `--localization-id` | string | — | appClipDefaultExperienceLocalization id (required) |

## `asc appinfo`

Work with app-level information (name, subtitle, category, age rating)

Work with the app's editable appInfo: localized name/subtitle/privacy URL,
primary/secondary category, and the age-rating declaration.

These fields live on the appInfo resource, not on a specific version.

### `asc appinfo age-rating`

Set the age-rating declaration from a JSON attributes file

Set the age-rating declaration. Because Apple periodically changes the
questionnaire fields, this command takes the raw attributes object as JSON
rather than hardcoding a preset.

Note: with the newer age-rating flow the declaration may not be readable via
the API (GETs can 404 even when the UI shows a rating). Writing still works;
verify the resulting rating in the App Store Connect UI.

Dump the current attributes, edit them (e.g. set every content field to "NONE"
for a 4+ rating), and apply:

    asc appinfo age-rating --app 6790641087 --attrs @agerating.json

Example:

```
asc appinfo age-rating --app 6790641087 --attrs @agerating.json
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--attrs` | string | — | JSON attributes object, or @file (required) |

### `asc appinfo category`

Set the primary and/or secondary category

Set the app's primary and/or secondary category. Category ids are the App
Store Connect enum values, e.g. PRODUCTIVITY, BUSINESS, FINANCE, UTILITIES.
List available ids with: asc api /v1/appCategories

Example:

```
asc appinfo category --app 6790641087 --primary PRODUCTIVITY --secondary BUSINESS
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--primary` | string | — | primary category id (e.g. PRODUCTIVITY) |
| `--secondary` | string | — | secondary category id (e.g. BUSINESS) |

### `asc appinfo localize`

Set the localized name, subtitle and privacy policy URL

Example:

```
asc appinfo localize --app 6790641087 --locale ja \
    --name "日本領収書スキャン" --subtitle "データ化するだけのAIレシートスキャナ" \
    --privacy-url https://japan-receipt-scan.web.app/privacy
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | app name (max 30 chars) |
| `--privacy-url` | string | — | privacy policy URL |
| `--subtitle` | string | — | subtitle (max 30 chars) |

### `asc appinfo show`

Show the app's appInfo (state, categories, age rating, localizations)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

## `asc apps`

Work with apps

### `asc apps list`

List all apps in the team

### `asc apps show`

Show an app's attributes (name, bundle id, SKU, locale, content rights)

Example:

```
asc apps show --app 6790641087
  asc apps show --app com.example.myapp
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc apps update`

Update app-level settings (primary locale, content rights declaration)

Example:

```
asc apps update --app 6790641087 --primary-locale ja
  asc apps update --app 6790641087 --content-rights DOES_NOT_USE_THIRD_PARTY_CONTENT
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--content-rights` | string | — | DOES_NOT_USE_THIRD_PARTY_CONTENT or USES_THIRD_PARTY_CONTENT |
| `--primary-locale` | string | — | primary locale, e.g. ja / en-US |

## `asc assets`

Upload and manage media assets (screenshots) via the reserve→upload→commit flow

Upload media assets to App Store Connect. Uploading is the one part of the
submission flow that "asc api" cannot do on its own: each asset is reserved
(POST), its bytes are PUT to a per-asset pre-signed URL, then the upload is
committed (PATCH uploaded=true with an MD5 checksum). These commands do all
three steps.

A 2xx commit does NOT mean the asset was accepted: Apple validates
asynchronously and rejections (e.g. IMAGE_INCORRECT_DIMENSIONS) only appear in
assetDeliveryState afterwards. These commands therefore poll until validation
reaches COMPLETE and exit non-zero on FAILED with the error codes.

### `asc assets delete-preview`

Delete an app preview by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appPreview id (required) |

### `asc assets delete-preview-set`

Delete a preview set (and its previews) by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appPreviewSet id (required) |

### `asc assets delete-screenshot`

Delete a screenshot by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appScreenshot id (required) |

### `asc assets delete-screenshot-set`

Delete a screenshot set (and its screenshots) by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appScreenshotSet id (required) |

### `asc assets list`

List screenshot sets and screenshots for a locale

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--version` | string | — | version string (default: the editable version) |

### `asc assets list-previews`

List app preview sets and previews for a locale

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--version` | string | — | version string (default: the editable version) |

### `asc assets upload-preview`

Upload one or more app preview videos for a locale and preview type

Upload app preview videos to the editable version. Files are appended to the
preview set for the given preview type in the order passed.

Common --display values: IPHONE_67, IPHONE_61, IPHONE_65, IPHONE_58, IPHONE_55,
IPAD_PRO_3GEN_129, IPAD_PRO_3GEN_11, DESKTOP, APPLE_TV, APPLE_VISION_PRO.
Note preview types have no APP_ prefix, unlike screenshot display types.

Example:

```
asc assets upload-preview --app 6790641087 --locale ja --display IPHONE_67 \
    --file app-store/previews/demo.mp4 --frame-time-code 00:00:05:00
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--display` | string | — | preview type, e.g. IPHONE_67 (required) |
| `--file` | stringArray | `[]` | preview video file (repeatable, applied in order) |
| `--frame-time-code` | string | — | poster frame time code, e.g. 00:00:05:00 |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--version` | string | — | version string (default: the editable version) |

### `asc assets upload-routing-coverage`

Upload a GeoJSON routing app coverage file for the editable version

Upload the .geojson file that defines where a routing app offers coverage.
A version has at most one routing app coverage; any existing one is deleted
and replaced.

Example:

```
asc assets upload-routing-coverage --app 6790641087 --file coverage.geojson
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--file` | string | — | GeoJSON coverage file (required) |
| `--version` | string | — | version string (default: the editable version) |

### `asc assets upload-screenshot`

Upload one or more screenshots for a locale and display type

Upload screenshots to the editable version. Files are appended to the
screenshot set for the given display type in filename order, so pass them in the
order you want them shown.

Common --display values: APP_IPHONE_67 (6.5"/6.7"/6.9"), APP_IPHONE_61,
APP_IPHONE_55, APP_IPAD_PRO_129, APP_IPAD_PRO_3GEN_129. Full list:
asc api "/v1/appScreenshotSets?..." or Apple's ScreenshotDisplayType docs.

Example:

```
asc assets upload-screenshot --app 6790641087 --locale ja --display APP_IPHONE_67 \
    --file app-store/screenshots/01-hero.png --file app-store/screenshots/02.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--display` | string | — | screenshot display type, e.g. APP_IPHONE_67 (required) |
| `--file` | stringArray | `[]` | screenshot file (repeatable, applied in order) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--version` | string | — | version string (default: the editable version) |

## `asc availability`

Manage the app's territory availability

### `asc availability end-preorder`

End the app's pre-order in every territory where it is enabled

Example:

```
asc availability end-preorder --app 6790641087
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc availability set`

Set the territories where the app is available

Replace the app's availability with the given territory list. Territories
not listed become unavailable. The API requires a territoryAvailability entry
for EVERY App Store territory (~175), so the command fetches the full territory
list and marks the ones you pass available and all others unavailable. Pass
--available-in-new-territories to opt in to territories Apple adds later.

Example:

```
asc availability set --app 6790641087 --territories JPN,USA,GBR
  asc availability set --app 6790641087 --territories JPN --available-in-new-territories
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--available-in-new-territories` | bool | `false` | automatically become available in future new territories |
| `--territories` | string | — | comma-separated territory codes, e.g. JPN,USA,GBR (required) |

### `asc availability show`

Show the app's territory availability and pre-order status

Example:

```
asc availability show --app 6790641087
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

## `asc background-assets`

Manage Background Assets asset packs (list, create, archive, versions, upload, release status)

Manage Background Assets asset packs through the App Store Connect API.

An asset pack (backgroundAsset) belongs to an app and is identified by its
assetPackIdentifier. Each upload creates a new backgroundAssetVersion whose
version number Apple assigns automatically. Releases to internal beta,
external beta, and the App Store are separate read-only resources; App Store
and external beta releases are initiated by adding the backgroundAssetVersion
to a review submission (reviewSubmissionItems), while internal beta releases
happen automatically when a version finishes processing.

### `asc background-assets archive`

Archive a background asset pack

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | backgroundAsset id (required) |

### `asc background-assets create`

Create a background asset pack for an app

Example:

```
asc background-assets create --app 6790641087 --pack-id com.example.app.levels
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--pack-id` | string | — | assetPackIdentifier, e.g. com.example.app.levels (required) |

### `asc background-assets list`

List the app's background asset packs

Example:

```
asc background-assets list --app 6790641087
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc background-assets releases`

Show a version's release state on each track (internal-beta, external-beta, app-store)

Show the release resources attached to a backgroundAssetVersion. The App
Store Connect API exposes releases as read-only state: internal beta releases
are created automatically when a version finishes processing, and external
beta / App Store releases are created by submitting the backgroundAssetVersion
for review (reviewSubmissionItems with a backgroundAssetVersion relationship).

Example:

```
asc background-assets releases --version-id 5678efgh-...
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | backgroundAssetVersion id (required) |

### `asc background-assets unarchive`

Unarchive a background asset pack

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | backgroundAsset id (required) |

### `asc background-assets upload`

Upload a new version of an asset pack (create version, then reserve→PUT→commit)

Upload a new background asset version: a backgroundAssetVersion is created
for the asset pack (Apple assigns the version number), the asset pack file is
reserved as a backgroundAssetUploadFile, its bytes are PUT to Apple's
pre-signed URLs, and the upload is committed with an MD5 checksum. Pass
--manifest to also upload a MANIFEST file for the same version. Apple then
processes the version (see "background-assets versions").

Example:

```
asc background-assets upload --asset-id 1234abcd-... --file pack.aar
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--asset-id` | string | — | backgroundAsset id (required) |
| `--file` | string | — | asset pack file to upload (required) |
| `--manifest` | string | — | optional MANIFEST file to upload for the same version |

### `asc background-assets versions`

List a background asset pack's versions, newest first

Example:

```
asc background-assets versions --asset-id 1234abcd-...
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--asset-id` | string | — | backgroundAsset id (required) |

## `asc beta`

Manage TestFlight beta testing (groups, testers, localizations, review)

### `asc beta app-localize`

Set the TestFlight beta app information for a locale

Example:

```
asc beta app-localize --app 6790641087 --locale ja \
    --description @testflight/description.txt --feedback-email support@example.com
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--description` | string | — | beta app description (@file allowed) |
| `--feedback-email` | string | — | email testers send feedback to |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--marketing-url` | string | — | marketing URL |
| `--privacy-policy-url` | string | — | privacy policy URL |
| `--tv-os-privacy-policy` | string | — | tvOS privacy policy text (@file allowed) |

### `asc beta build-bundles`

List a build's bundles (bundle ids, SDK, symbols/dSYM availability)

Build bundles are only reachable through the build resource
(GET /v1/builds/{id}?include=buildBundles); their ids are needed for App Clip
invocation commands and expose dSYM/symbol availability.

Example:

```
asc beta build-bundles --app 6790641087 --build 42
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc beta build-detail`

Show or set a build's TestFlight beta details (auto-notify, beta states)

#### `asc beta build-detail set`

Set whether testers are automatically notified about the build

Example:

```
asc beta build-detail set --app 6790641087 --build 42 --auto-notify true
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--auto-notify` | string | — | true or false (required) |
| `--build` | string | — | build id or build version string (required) |

#### `asc beta build-detail show`

Show a build's beta detail (autoNotifyEnabled, internal/external beta state)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc beta build-localize`

Set a build's TestFlight "what to test" notes for a locale

Example:

```
asc beta build-localize --app 6790641087 --build 42 --locale ja --whats-new @testflight/whats-new.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--whats-new` | string | — | what to test in this build (@file allowed, required) |

### `asc beta groups`

Manage TestFlight beta groups

#### `asc beta groups add-build`

Give a beta group access to a build

Example:

```
asc beta groups add-build --group-id <group-id> --app 6790641087 --build 42
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta groups add-testers`

Add testers to a beta group by email (invites new testers)

Example:

```
asc beta groups add-testers --group-id <group-id> --email a@example.com --email b@example.com
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | stringArray | `[]` | tester email (repeatable, required) |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta groups create`

Create a beta group

Example:

```
asc beta groups create --app 6790641087 --name "External Testers" \
    --public-link --public-link-limit 100 --feedback-enabled
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--feedback-enabled` | bool | `false` | enable tester feedback |
| `--name` | string | — | group name (required) |
| `--public-link` | bool | `false` | enable the public TestFlight invitation link |
| `--public-link-limit` | int | `0` | maximum testers via the public link (enables the limit) |

#### `asc beta groups delete`

Delete a beta group

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta groups list`

List the app's beta groups

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc beta groups remove-build`

Remove a build from a beta group

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta groups remove-testers`

Remove testers from a beta group by email

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | stringArray | `[]` | tester email (repeatable, required) |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta groups update`

Update a beta group's name, public link, or feedback settings

| flag | type | default | description |
| --- | --- | --- | --- |
| `--feedback-enabled` | bool | `false` | enable tester feedback |
| `--group-id` | string | — | beta group id (required) |
| `--name` | string | — | group name |
| `--public-link` | bool | `false` | enable the public TestFlight invitation link |
| `--public-link-limit` | int | `0` | maximum testers via the public link (enables the limit) |

### `asc beta invocations`

Manage TestFlight App Clip invocations for a build bundle

Beta App Clip invocations let testers launch an App Clip experience from
TestFlight. They belong to a build bundle; find bundle ids with
"asc beta build-bundles --build <id|version> --app <app>".

#### `asc beta invocations create`

Create an App Clip invocation with one localization

Create a beta App Clip invocation. The API requires at least one
localization (locale + title), created inline with the invocation. Add more
locales afterwards with "asc beta invocations localize".

Example:

```
asc beta invocations create --build-bundle-id <bundle-id> \
    --url https://example.com/clip --locale ja --title "クリップを試す"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--build-bundle-id` | string | — | build bundle id (required; see: asc beta build-bundles) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--title` | string | — | localized title shown in TestFlight (required) |
| `--url` | string | — | App Clip invocation URL (required) |

#### `asc beta invocations delete`

Delete an App Clip invocation

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | App Clip invocation id (required) |

#### `asc beta invocations delete-localization`

Delete an App Clip invocation localization

Delete a betaAppClipInvocationLocalization by its id (shown in the
LOCALIZATIONS column of "asc beta invocations list").

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | App Clip invocation localization id (required) |

#### `asc beta invocations list`

List a build bundle's App Clip invocations

| flag | type | default | description |
| --- | --- | --- | --- |
| `--build-bundle-id` | string | — | build bundle id (required; see: asc beta build-bundles) |

#### `asc beta invocations localize`

Add a localized title to an App Clip invocation

Example:

```
asc beta invocations localize --id <invocation-id> --locale en-US --title "Try the clip"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | App Clip invocation id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--title` | string | — | localized title shown in TestFlight (required) |

### `asc beta license`

Show or set the app's TestFlight beta license agreement

#### `asc beta license set`

Set the beta license agreement text

Example:

```
asc beta license set --app 6790641087 --text @testflight/license.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--text` | string | — | agreement text (@file allowed, required) |

#### `asc beta license show`

Show the beta license agreement text

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc beta recruitment`

Manage a beta group's tester recruitment criteria (device/OS filters)

Recruitment criteria limit which devices and OS versions can join a beta
group through its public link. They belong to a beta group, so commands take
--group-id (find it with: asc beta groups list --app <app>).

#### `asc beta recruitment delete`

Delete the beta group's recruitment criteria

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta recruitment options`

List the device families and OS versions usable in recruitment criteria

#### `asc beta recruitment set`

Create or replace the beta group's recruitment criteria

Each --filter is FAMILY[:MIN_OS[:MAX_OS]] where FAMILY is one of
IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, VISION. Min/max OS versions are
inclusive and optional. Use "asc beta recruitment options" to list the OS
versions Apple accepts.

Example:

```
asc beta recruitment set --group-id <group-id> --filter IPHONE:17.0 --filter IPAD:17.0:18.1
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--filter` | stringArray | `[]` | device/OS filter FAMILY[:MIN_OS[:MAX_OS]], e.g. IPHONE:17.0:18.0 (repeatable, required) |
| `--group-id` | string | — | beta group id (required) |

#### `asc beta recruitment show`

Show the beta group's recruitment criteria

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | beta group id (required) |

### `asc beta review-detail`

Show or set the app's TestFlight beta review contact information

#### `asc beta review-detail set`

Set beta review contact and demo account fields

Example:

```
asc beta review-detail set --app 6790641087 \
    --contact-first-name Taro --contact-last-name Yamada \
    --contact-email review@example.com --contact-phone +81-3-0000-0000
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--contact-email` | string | — | review contact email |
| `--contact-first-name` | string | — | review contact first name |
| `--contact-last-name` | string | — | review contact last name |
| `--contact-phone` | string | — | review contact phone |
| `--demo-account-name` | string | — | demo account user name |
| `--demo-account-password` | string | — | demo account password |
| `--demo-account-required` | bool | `false` | a demo account is required to review |
| `--notes` | string | — | notes for the reviewer (@file allowed) |

#### `asc beta review-detail show`

Show the beta app review detail

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc beta submit`

Submit a build for beta review (required before external testing)

Example:

```
asc beta submit --app 6790641087 --build 42
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc beta testers`

Manage TestFlight beta testers

#### `asc beta testers invite`

Invite a tester by email to a beta group

Create a beta tester and add it to a group; Apple emails the TestFlight
invitation. The group determines the app, so no --app flag is needed.

Example:

```
asc beta testers invite --group-id <group-id> --email tester@example.com --first-name Hanako
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | string | — | tester email (required) |
| `--first-name` | string | — | tester first name |
| `--group-id` | string | — | beta group id (required) |
| `--last-name` | string | — | tester last name |

#### `asc beta testers list`

List the app's beta testers (optionally within one group)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--group-id` | string | — | limit to one beta group |

#### `asc beta testers reinvite`

Resend the TestFlight invitation email to an existing tester

Example:

```
asc beta testers reinvite --app 6790641087 --email tester@example.com
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--email` | string | — | tester email (required) |

#### `asc beta testers remove`

Remove a tester's access to the app (all groups and builds)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--email` | string | — | tester email (required) |

## `asc builds`

Manage builds (list, expire, encryption, notify, upload binaries)

### `asc builds encryption`

Set a build's export compliance (usesNonExemptEncryption)

Example:

```
asc builds encryption --app 6790641087 --build 42 --uses-non-exempt-encryption false
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |
| `--uses-non-exempt-encryption` | string | — | true or false (required) |

### `asc builds expire`

Expire a build (removes it from TestFlight)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc builds list`

List the app's builds, newest first

Example:

```
asc builds list --app 6790641087
  asc builds list --app 6790641087 --version 1.2.0 --limit 10
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--limit` | int | `50` | maximum number of builds to list |
| `--version` | string | — | filter by marketing (prerelease) version, e.g. 1.2.0 |

### `asc builds notify`

Notify TestFlight testers that a build is available

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc builds prerelease-versions`

List the app's prerelease (TestFlight) versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc builds show`

Show a build's full details

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required when --build is a version string) |
| `--build` | string | — | build id or build version string (required) |

### `asc builds upload`

Upload a build binary via the buildUploads API (no Transporter needed)

Upload an .ipa/.pkg/.zip build binary directly through the App Store
Connect API: a buildUpload is created for the app/version, a buildUploadFile is
reserved, the bytes are PUT to Apple's pre-signed URLs, and the upload is
committed with an MD5 checksum. Apple then processes the binary into a build.

Example:

```
asc builds upload --app 6790641087 --file build/App.ipa \
    --platform IOS --version 1.2.0 --build-number 42
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--build-number` | string | — | CFBundleVersion, e.g. 42 (required) |
| `--file` | string | — | build binary: .ipa, .pkg, or .zip (required) |
| `--platform` | string | `IOS` | platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | CFBundleShortVersionString, e.g. 1.2.0 (required) |

## `asc bundle-ids`

Manage bundle IDs and their capabilities

### `asc bundle-ids capabilities`

List the capabilities enabled on a bundle ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | bundle ID resource id (required) |

### `asc bundle-ids create`

Register a new bundle ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--identifier` | string | — | bundle identifier, e.g. com.example.app (required) |
| `--name` | string | — | bundle ID name (required) |
| `--platform` | string | — | platform: IOS, MAC_OS or UNIVERSAL (required) |

### `asc bundle-ids delete`

Delete a bundle ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | bundle ID resource id (required) |

### `asc bundle-ids disable`

Disable a capability on a bundle ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--capability` | string | — | capability: ICLOUD, IN_APP_PURCHASE, GAME_CENTER, PUSH_NOTIFICATIONS, WALLET, INTER_APP_AUDIO, MAPS, ASSOCIATED_DOMAINS, PERSONAL_VPN, APP_GROUPS, HEALTHKIT, HOMEKIT, WIRELESS_ACCESSORY_CONFIGURATION, APPLE_PAY, DATA_PROTECTION, SIRIKIT, NETWORK_EXTENSIONS, MULTIPATH, HOT_SPOT, NFC_TAG_READING, CLASSKIT, AUTOFILL_CREDENTIAL_PROVIDER, ACCESS_WIFI_INFORMATION, NETWORK_CUSTOM_PROTOCOL, COREMEDIA_HLS_LOW_LATENCY, SYSTEM_EXTENSION_INSTALL, USER_MANAGEMENT, APPLE_ID_AUTH (required) |
| `--id` | string | — | bundle ID resource id (required) |

### `asc bundle-ids enable`

Enable a capability on a bundle ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--capability` | string | — | capability: ICLOUD, IN_APP_PURCHASE, GAME_CENTER, PUSH_NOTIFICATIONS, WALLET, INTER_APP_AUDIO, MAPS, ASSOCIATED_DOMAINS, PERSONAL_VPN, APP_GROUPS, HEALTHKIT, HOMEKIT, WIRELESS_ACCESSORY_CONFIGURATION, APPLE_PAY, DATA_PROTECTION, SIRIKIT, NETWORK_EXTENSIONS, MULTIPATH, HOT_SPOT, NFC_TAG_READING, CLASSKIT, AUTOFILL_CREDENTIAL_PROVIDER, ACCESS_WIFI_INFORMATION, NETWORK_CUSTOM_PROTOCOL, COREMEDIA_HLS_LOW_LATENCY, SYSTEM_EXTENSION_INSTALL, USER_MANAGEMENT, APPLE_ID_AUTH (required) |
| `--id` | string | — | bundle ID resource id (required) |

### `asc bundle-ids list`

List bundle IDs

## `asc certificates`

Manage signing certificates

### `asc certificates create`

Create a signing certificate from a CSR file

Example:

```
asc certificates create --type IOS_DISTRIBUTION --csr CertificateSigningRequest.certSigningRequest
  asc certificates create --type DEVELOPMENT --csr request.csr --output development.cer
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--csr` | string | — | path to a .certSigningRequest / PEM CSR file (required) |
| `--merchant-id-id` | string | — | merchantIds resource id (required for APPLE_PAY* types) |
| `--output` | string | — | output .cer path (default: <serial>.cer) |
| `--pass-type-id-id` | string | — | passTypeIds resource id (required for PASS_TYPE_ID* types) |
| `--type` | string | — | certificate type: IOS_DEVELOPMENT, IOS_DISTRIBUTION, MAC_APP_DEVELOPMENT, MAC_APP_DISTRIBUTION, MAC_INSTALLER_DISTRIBUTION, DEVELOPMENT, DISTRIBUTION, DEVELOPER_ID_APPLICATION, DEVELOPER_ID_APPLICATION_G2, DEVELOPER_ID_KEXT, DEVELOPER_ID_KEXT_G2, APPLE_PAY, APPLE_PAY_MERCHANT_IDENTITY, APPLE_PAY_PSP_IDENTITY, APPLE_PAY_RSA, IDENTITY_ACCESS, PASS_TYPE_ID, PASS_TYPE_ID_WITH_NFC (required) |

### `asc certificates download`

Download a certificate as a .cer file

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | certificate id (required) |
| `--output` | string | — | output .cer path (default: <serial>.cer) |

### `asc certificates list`

List signing certificates

| flag | type | default | description |
| --- | --- | --- | --- |
| `--type` | string | — | filter by certificate type: IOS_DEVELOPMENT, IOS_DISTRIBUTION, MAC_APP_DEVELOPMENT, MAC_APP_DISTRIBUTION, MAC_INSTALLER_DISTRIBUTION, DEVELOPMENT, DISTRIBUTION, DEVELOPER_ID_APPLICATION, DEVELOPER_ID_APPLICATION_G2, DEVELOPER_ID_KEXT, DEVELOPER_ID_KEXT_G2, APPLE_PAY, APPLE_PAY_MERCHANT_IDENTITY, APPLE_PAY_PSP_IDENTITY, APPLE_PAY_RSA, IDENTITY_ACCESS, PASS_TYPE_ID, PASS_TYPE_ID_WITH_NFC |

### `asc certificates revoke`

Revoke a certificate

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | certificate id (required) |

## `asc configure`

Add a profile from an issuer ID and a downloaded .p8 key

Copies the .p8 private key into ~/.config/apple-app-store-connect/keys/
(with 0600 permissions) and registers a profile in config.toml.

The key ID is derived from the AuthKey_XXXXXXXXXX.p8 filename unless --key-id
is given. The first registered profile becomes the default profile.

Use an App Store Connect API key WITH A ROLE (Users and Access > Integrations >
App Store Connect API; Admin or App Manager for write access). In-App Purchase
keys for the App Store Server API look similar but get 401 on this API.

Example:

```
asc configure --issuer-id 12345678-aaaa-bbbb-cccc-1234567890ab --key ~/Downloads/AuthKey_ABC123DEF4.p8
  asc configure --profile client-a --issuer-id ... --key ... --key-id ABC123DEF4
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--force` | bool | `false` | overwrite an existing profile |
| `--issuer-id` | string | — | App Store Connect API issuer ID (required) |
| `--key` | string | — | path to the downloaded AuthKey_XXXXXXXXXX.p8 file (required) |
| `--key-id` | string | — | key ID (default: derived from the .p8 filename) |

## `asc devices`

Manage registered devices

### `asc devices list`

List registered devices

| flag | type | default | description |
| --- | --- | --- | --- |
| `--platform` | string | — | filter by platform: IOS, MAC_OS or UNIVERSAL |

### `asc devices register`

Register a new device

| flag | type | default | description |
| --- | --- | --- | --- |
| `--name` | string | — | device name (required) |
| `--platform` | string | — | platform: IOS, MAC_OS or UNIVERSAL (required) |
| `--udid` | string | — | device UDID (required) |

### `asc devices update`

Rename a device or change its status

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | device id (required) |
| `--name` | string | — | new device name |
| `--status` | string | — | ENABLED or DISABLED |

## `asc diagnostics`

Inspect build diagnostic signatures and logs

### `asc diagnostics logs`

Download the diagnostic logs for a diagnostic signature

| flag | type | default | description |
| --- | --- | --- | --- |
| `--output` | string | — | write to file instead of stdout |
| `--signature-id` | string | — | diagnostic signature id (required) |

### `asc diagnostics signatures`

List a build's diagnostic signatures

Example:

```
asc diagnostics signatures --build <build id>
  asc diagnostics signatures --build <build id> --type HANGS
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--build` | string | — | build id (required) |
| `--type` | string | — | filter by diagnostic type: DISK_WRITES, HANGS, LAUNCHES |

## `asc encryption`

Manage app encryption declarations (export compliance)

### `asc encryption assign-build`

Assign a build to an encryption declaration

| flag | type | default | description |
| --- | --- | --- | --- |
| `--build-id` | string | — | build id (required) |
| `--id` | string | — | appEncryptionDeclaration id (required) |

### `asc encryption create`

Create an encryption declaration for an app

Create an app encryption declaration. --description explains how the app uses
encryption. Pass --proprietary and/or --third-party when the app contains
proprietary or third-party cryptography (beyond Apple's OS crypto), and
--available-on-french-store when the app is distributed in France.

Example:

```
asc encryption create --app com.example.app \
    --description "Uses HTTPS only" --available-on-french-store
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--available-on-french-store` | bool | `false` | the app is available on the French store |
| `--description` | string | — | description of how the app uses encryption (required) |
| `--proprietary` | bool | `false` | contains proprietary cryptography |
| `--third-party` | bool | `false` | contains third-party cryptography |

### `asc encryption list`

List encryption declarations of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc encryption show`

Show an encryption declaration

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appEncryptionDeclaration id (required) |

### `asc encryption upload-document`

Upload a compliance document (e.g. PDF) for an encryption declaration

Example:

```
asc encryption upload-document --id DECLARATION_ID --file compliance.pdf
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | document file, e.g. a PDF (required) |
| `--id` | string | — | appEncryptionDeclaration id (required) |

## `asc eula`

Work with the app's custom end user license agreement

Manage the app's custom EULA (endUserLicenseAgreement). Apps without a custom
EULA use Apple's standard EULA. A custom EULA applies to an explicit list of
territories.

### `asc eula delete`

Delete the app's custom EULA (reverting to Apple's standard EULA)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc eula set`

Create or replace the app's custom EULA

Example:

```
asc eula set --app 6790641087 --text @eula.txt --territories JPN,USA
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--territories` | stringSlice | `[]` | territory ids the EULA applies to, e.g. JPN,USA (required on create) |
| `--text` | string | — | agreement text, or @file (required) |

### `asc eula show`

Show the app's custom EULA, if any

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

## `asc events`

Work with in-app events (create, schedule, localize, upload media)

Manage in-app events (appEvents): timed events like challenges, competitions
and premieres that appear on the App Store product page.

A typical flow: create the event, localize it (name, descriptions), upload the
event card / details page media, then submit it for review with:

    asc review-submissions add-item --event-id <event-id> ...

Dates are ISO 8601 date-times, e.g. 2026-08-01T10:00:00Z.

### `asc events create`

Create an in-app event

Example:

```
asc events create --app 6790641087 --reference-name "Summer Challenge" \
    --badge CHALLENGE --purpose ATTRACT_NEW_USERS --priority HIGH \
    --publish-start 2026-08-01T00:00:00Z --start 2026-08-08T00:00:00Z --end 2026-08-15T00:00:00Z
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--badge` | string | — | badge: LIVE_EVENT, PREMIERE, CHALLENGE, COMPETITION, NEW_SEASON, MAJOR_UPDATE, SPECIAL_EVENT |
| `--deep-link` | string | — | deep link URL opened from the event |
| `--end` | string | — | event end (ISO 8601) |
| `--primary-locale` | string | — | primary locale, e.g. ja / en-US |
| `--priority` | string | — | priority: HIGH or NORMAL |
| `--publish-start` | string | — | when the event page becomes visible (ISO 8601) |
| `--purchase-requirement` | string | — | purchase requirement, e.g. NO_COST_ASSOCIATED, IN_APP_PURCHASE, SUBSCRIPTION |
| `--purpose` | string | — | purpose: APPROPRIATE_FOR_ALL_USERS, ATTRACT_NEW_USERS, KEEP_ACTIVE_USERS_INFORMED, BRING_BACK_LAPSED_USERS |
| `--reference-name` | string | — | internal reference name |
| `--start` | string | — | event start (ISO 8601) |
| `--territories` | string | — | comma-separated territory ids for the schedule, e.g. JPN,USA (default: all territories) |

### `asc events delete`

Delete an in-app event

| flag | type | default | description |
| --- | --- | --- | --- |
| `--event-id` | string | — | appEvent id (required) |

### `asc events delete-screenshot`

Delete an event screenshot by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | asset id (required) |

### `asc events delete-video`

Delete an event video clip by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | asset id (required) |

### `asc events list`

List the app's in-app events

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--state` | string | — | filter by eventState, e.g. DRAFT, READY_FOR_REVIEW, PUBLISHED |

### `asc events list-assets`

List uploaded screenshots and video clips for a locale

| flag | type | default | description |
| --- | --- | --- | --- |
| `--event-id` | string | — | appEvent id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |

### `asc events localize`

Set the localized event name and descriptions

Example:

```
asc events localize --event-id <id> --locale ja \
    --name "夏のチャレンジ" --short-description "1週間の限定イベント" \
    --long-description @app-store/event-long.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--event-id` | string | — | appEvent id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--long-description` | string | — | long description, max 120 chars (@file allowed) |
| `--name` | string | — | event name (max 30 chars) |
| `--short-description` | string | — | short description, max 50 chars (@file allowed) |

### `asc events update`

Update an in-app event's attributes or schedule

| flag | type | default | description |
| --- | --- | --- | --- |
| `--badge` | string | — | badge: LIVE_EVENT, PREMIERE, CHALLENGE, COMPETITION, NEW_SEASON, MAJOR_UPDATE, SPECIAL_EVENT |
| `--deep-link` | string | — | deep link URL opened from the event |
| `--end` | string | — | event end (ISO 8601) |
| `--event-id` | string | — | appEvent id (required) |
| `--primary-locale` | string | — | primary locale, e.g. ja / en-US |
| `--priority` | string | — | priority: HIGH or NORMAL |
| `--publish-start` | string | — | when the event page becomes visible (ISO 8601) |
| `--purchase-requirement` | string | — | purchase requirement, e.g. NO_COST_ASSOCIATED, IN_APP_PURCHASE, SUBSCRIPTION |
| `--purpose` | string | — | purpose: APPROPRIATE_FOR_ALL_USERS, ATTRACT_NEW_USERS, KEEP_ACTIVE_USERS_INFORMED, BRING_BACK_LAPSED_USERS |
| `--reference-name` | string | — | internal reference name |
| `--start` | string | — | event start (ISO 8601) |
| `--territories` | string | — | comma-separated territory ids for the schedule, e.g. JPN,USA (default: all territories) |

### `asc events upload-screenshot`

Upload an event screenshot for a locale and display target

Upload an image to the event localization. --display is the appEventAssetType:
EVENT_CARD (the card shown on the App Store) or EVENT_DETAILS_PAGE.

Example:

```
asc events upload-screenshot --event-id <id> --locale ja \
    --display EVENT_CARD --file app-store/event-card.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--display` | string | — | asset target: EVENT_CARD or EVENT_DETAILS_PAGE (required) |
| `--event-id` | string | — | appEvent id (required) |
| `--file` | string | — | media file to upload (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |

### `asc events upload-video`

Upload an event video clip for a locale and display target

Upload a video clip to the event localization. --display is the
appEventAssetType: EVENT_CARD or EVENT_DETAILS_PAGE.

Example:

```
asc events upload-video --event-id <id> --locale ja \
    --display EVENT_DETAILS_PAGE --file app-store/event-clip.mp4
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--display` | string | — | asset target: EVENT_CARD or EVENT_DETAILS_PAGE (required) |
| `--event-id` | string | — | appEvent id (required) |
| `--file` | string | — | media file to upload (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |

## `asc experiments`

Work with product page optimization experiments (A/B tests)

Manage product page optimization experiments (appStoreVersionExperiments v2):
A/B tests of alternate icons, screenshots and previews against the default page.

An experiment owns treatments; each treatment owns per-locale localizations,
which own screenshot sets. Experiments are submitted for review via:
asc review-submissions add-item ...

### `asc experiments create`

Create a product page experiment (v2)

Example:

```
asc experiments create --app 6790641087 --name "Icon test" --traffic-proportion 50
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--name` | string | — | experiment name (required) |
| `--platform` | string | `IOS` | platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--traffic-proportion` | int | `0` | percentage of traffic in the experiment, 1-100 (required) |

### `asc experiments create-treatment`

Add a treatment (variant) to an experiment

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app-icon-name` | string | — | alternate app icon asset name shipped in the build |
| `--experiment-id` | string | — | appStoreVersionExperiment (v2) id (required) |
| `--name` | string | — | treatment name (required) |

### `asc experiments delete`

Delete an experiment

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appStoreVersionExperiment (v2) id (required) |

### `asc experiments list`

List product page experiments (v2)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--version` | string | — | version string (default: list at app level / the editable version) |

### `asc experiments localize-treatment`

Create a treatment localization for a locale

Create the appStoreVersionExperimentTreatmentLocalization for a locale (or
print the existing one). The printed id is what upload-screenshot takes as
--treatment-localization-id.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--treatment-id` | string | — | appStoreVersionExperimentTreatment id (required) |

### `asc experiments promote`

Promote a winning treatment to the app's product page

Apply a winning experiment treatment to an App Store version by creating an
appStoreVersionPromotion. The API relates a promotion to an appStoreVersion and
an appStoreVersionExperimentTreatment (custom product page versions cannot be
promoted through this endpoint).

Example:

```
asc experiments promote --app 6790641087 --treatment-id <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--treatment-id` | string | — | appStoreVersionExperimentTreatment id to promote (required) |
| `--version` | string | — | version string (default: list at app level / the editable version) |

### `asc experiments start`

Start an approved experiment (started=true)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appStoreVersionExperiment (v2) id (required) |

### `asc experiments stop`

Stop a running experiment (started=false)

Stop a running experiment. The update request exposes a single boolean
attribute "started"; stopping sets it back to false.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appStoreVersionExperiment (v2) id (required) |

### `asc experiments treatments`

List an experiment's treatments

| flag | type | default | description |
| --- | --- | --- | --- |
| `--experiment-id` | string | — | appStoreVersionExperiment (v2) id (required) |

### `asc experiments upload-screenshot`

Upload screenshots to a treatment localization

Example:

```
asc experiments upload-screenshot --treatment-localization-id <id> \
    --display APP_IPHONE_67 --file variants/01.png --file variants/02.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--display` | string | — | screenshot display type, e.g. APP_IPHONE_67 (required) |
| `--file` | stringArray | `[]` | screenshot file (repeatable, applied in order) |
| `--treatment-localization-id` | string | — | appStoreVersionExperimentTreatmentLocalization id (required) |

## `asc feedback`

Inspect TestFlight beta feedback (crashes and screenshots)

### `asc feedback crash-log`

Download the crash log of a beta feedback crash submission

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | crash submission id (required) |
| `--output` | string | — | write to file instead of stdout |

### `asc feedback crashes`

List an app's beta feedback crash submissions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--build` | string | — | filter by build id |

### `asc feedback delete-crash`

Delete a beta feedback crash submission

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | crash submission id (required) |

### `asc feedback delete-screenshot`

Delete a beta feedback screenshot submission

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | screenshot submission id (required) |

### `asc feedback screenshots`

List an app's beta feedback screenshot submissions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--build` | string | — | filter by build id |

## `asc gamecenter`

Manage Game Center (achievements, leaderboards, leaderboard sets, groups)

Manage Game Center configuration. Uses the current (v2, versioned) Game
Center resources: achievements and leaderboards carry versions, localizations
hang off a version, and images hang off a localization.

Publishing note: the v1 "release" resources (gameCenterAchievementReleases etc.)
are deprecated in API 4.x. Versioned Game Center resources are published by
attaching their version to a review submission (reviewSubmissionItems with a
gameCenterAchievementVersion / gameCenterLeaderboardVersion relationship), so
there is no standalone release command here.

Aliases: gc

### `asc gamecenter achievements`

Manage Game Center achievements (v2, versioned)

Aliases: achievement

#### `asc gamecenter achievements create`

Create an achievement (an initial version is created inline)

Example:

```
asc gamecenter achievements create --app 6790641087 \
    --reference-name "First Win" --vendor-identifier com.example.first_win --points 10
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--points` | int | `0` | point value, 1-100 (required) |
| `--reference-name` | string | — | internal reference name (required) |
| `--repeatable` | bool | `false` | achievement can be earned more than once |
| `--show-before-earned` | bool | `false` | show the achievement before it is earned |
| `--vendor-identifier` | string | — | vendor identifier, e.g. com.example.first_win (required) |

#### `asc gamecenter achievements delete`

Delete an achievement

| flag | type | default | description |
| --- | --- | --- | --- |
| `--achievement-id` | string | — | gameCenterAchievement id (required) |

#### `asc gamecenter achievements list`

List the app's achievements

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter achievements localize`

Set an achievement's localized name and descriptions

Set an achievement's localized name and earned/unearned descriptions.
Localizations belong to the achievement's latest version (versions are created
automatically when the achievement is created). Creating a new localization
requires --name, --before-earned-desc and --after-earned-desc; updating an
existing one accepts any subset.

Example:

```
asc gamecenter achievements localize --achievement-id <id> --locale ja \
    --name "初勝利" --before-earned-desc "最初の対戦に勝つ" --after-earned-desc "最初の対戦に勝った"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--achievement-id` | string | — | gameCenterAchievement id (required) |
| `--after-earned-desc` | string | — | description after earned (@file allowed) |
| `--before-earned-desc` | string | — | description before earned (@file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name |

#### `asc gamecenter achievements update`

Update an achievement's attributes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--achievement-id` | string | — | gameCenterAchievement id (required) |
| `--archived` | bool | `false` | archive the achievement |
| `--points` | int | `0` | point value, 1-100 |
| `--reference-name` | string | — | internal reference name |
| `--repeatable` | bool | `false` | achievement can be earned more than once |
| `--show-before-earned` | bool | `false` | show the achievement before it is earned |

#### `asc gamecenter achievements upload-image`

Upload the image for an achievement localization

Upload an achievement image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the achievement's latest version:
GET /v2/gameCenterAchievements/{id}/versions then
GET /v2/gameCenterAchievementVersions/{id}/localizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | image file (required) |
| `--localization-id` | string | — | localization id (required) |

### `asc gamecenter activities`

Manage Game Center activities (versioned)

#### `asc gamecenter activities create`

Create an activity (an initial version is created inline)

Example:

```
asc gamecenter activities create --app 6790641087 \
    --reference-name "Daily Race" --vendor-identifier com.example.daily_race \
    --play-style SYNCHRONOUS --min-players 2 --max-players 8
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--max-players` | int | `0` | maximum players count |
| `--min-players` | int | `0` | minimum players count |
| `--play-style` | string | — | ASYNCHRONOUS or SYNCHRONOUS |
| `--reference-name` | string | — | internal reference name (required) |
| `--supports-party-code` | bool | `false` | activity supports party codes |
| `--vendor-identifier` | string | — | vendor identifier, e.g. com.example.daily_race (required) |

#### `asc gamecenter activities delete`

Delete an activity

| flag | type | default | description |
| --- | --- | --- | --- |
| `--activity-id` | string | — | gameCenterActivity id (required) |

#### `asc gamecenter activities delete-image`

Delete an activity image

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | image id (required) |

#### `asc gamecenter activities list`

List the app's Game Center activities

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter activities localize`

Set an activity's localized name and description

Set an activity's localized name and description. Localizations belong to
the activity's latest version. Creating a new localization requires --name;
updates accept any subset.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--activity-id` | string | — | gameCenterActivity id (required) |
| `--description` | string | — | localized description (@file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name |

#### `asc gamecenter activities update`

Update an activity's attributes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--activity-id` | string | — | gameCenterActivity id (required) |
| `--archived` | bool | `false` | archive the activity |
| `--max-players` | int | `0` | maximum players count |
| `--min-players` | int | `0` | minimum players count |
| `--play-style` | string | — | ASYNCHRONOUS or SYNCHRONOUS |
| `--reference-name` | string | — | internal reference name |
| `--supports-party-code` | bool | `false` | activity supports party codes |

#### `asc gamecenter activities upload-image`

Upload the image for an activity localization

Upload an activity image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the activity's latest version:
GET /v1/gameCenterActivities/{id}/versions then
GET /v1/gameCenterActivityVersions/{id}/localizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | image file (required) |
| `--localization-id` | string | — | localization id (required) |

### `asc gamecenter app-versions`

Manage Game Center app versions (per-App-Store-version enablement)

Manage gameCenterAppVersions: the resources that enable Game Center for a
specific App Store version and declare multiplayer compatibility between
versions (via the compatibilityVersions relationship).

#### `asc gamecenter app-versions disable`

Disable Game Center for an App Store version

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | appStoreVersion id (required) |

#### `asc gamecenter app-versions enable`

Enable Game Center for an App Store version

Create the gameCenterAppVersion for the given appStoreVersion if needed and set enabled=true.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | appStoreVersion id (required) |

#### `asc gamecenter app-versions link`

Mark two Game Center app versions as multiplayer-compatible

| flag | type | default | description |
| --- | --- | --- | --- |
| `--compatibility-version-id` | string | — | compatible gameCenterAppVersion id (required) |
| `--id` | string | — | gameCenterAppVersion id (required) |

#### `asc gamecenter app-versions list`

List the app's Game Center app versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter app-versions unlink`

Remove a multiplayer compatibility link between two app versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--compatibility-version-id` | string | — | compatible gameCenterAppVersion id (required) |
| `--id` | string | — | gameCenterAppVersion id (required) |

### `asc gamecenter challenges`

Manage Game Center challenges (versioned)

#### `asc gamecenter challenges create`

Create a challenge (an initial version is created inline)

Example:

```
asc gamecenter challenges create --app 6790641087 \
    --reference-name "Weekly Sprint" --vendor-identifier com.example.weekly_sprint \
    --leaderboard-id <gameCenterLeaderboard id> --repeatable
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--challenge-type` | string | `LEADERBOARD` | challenge type (LEADERBOARD) |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id backing the challenge |
| `--reference-name` | string | — | internal reference name (required) |
| `--repeatable` | bool | `false` | challenge can be repeated |
| `--vendor-identifier` | string | — | vendor identifier, e.g. com.example.weekly_sprint (required) |

#### `asc gamecenter challenges delete`

Delete a challenge

| flag | type | default | description |
| --- | --- | --- | --- |
| `--challenge-id` | string | — | gameCenterChallenge id (required) |

#### `asc gamecenter challenges delete-image`

Delete a challenge image

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | image id (required) |

#### `asc gamecenter challenges list`

List the app's Game Center challenges

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter challenges localize`

Set a challenge's localized name and description

Set a challenge's localized name and description. Localizations belong to
the challenge's latest version. Creating a new localization requires --name;
updates accept any subset.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--challenge-id` | string | — | gameCenterChallenge id (required) |
| `--description` | string | — | localized description (@file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name |

#### `asc gamecenter challenges update`

Update a challenge's attributes or its leaderboard

| flag | type | default | description |
| --- | --- | --- | --- |
| `--archived` | bool | `false` | archive the challenge |
| `--challenge-id` | string | — | gameCenterChallenge id (required) |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id backing the challenge |
| `--reference-name` | string | — | internal reference name |
| `--repeatable` | bool | `false` | challenge can be repeated |

#### `asc gamecenter challenges upload-image`

Upload the image for a challenge localization

Upload a challenge image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the challenge's latest version:
GET /v1/gameCenterChallenges/{id}/versions then
GET /v1/gameCenterChallengeVersions/{id}/localizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | image file (required) |
| `--localization-id` | string | — | localization id (required) |

### `asc gamecenter enable`

Enable Game Center for an app (create its gameCenterDetail)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc gamecenter groups`

Manage Game Center groups

Aliases: group

#### `asc gamecenter groups create`

Create a Game Center group

| flag | type | default | description |
| --- | --- | --- | --- |
| `--reference-name` | string | — | internal reference name |

#### `asc gamecenter groups list`

List Game Center groups visible to this team

### `asc gamecenter leaderboard-sets`

Manage Game Center leaderboard sets (v2, versioned)

Aliases: leaderboard-set

#### `asc gamecenter leaderboard-sets add-leaderboard`

Add a leaderboard to a leaderboard set

| flag | type | default | description |
| --- | --- | --- | --- |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id to add (required) |
| `--set-id` | string | — | gameCenterLeaderboardSet id (required) |

#### `asc gamecenter leaderboard-sets create`

Create a leaderboard set (an initial version is created inline)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--reference-name` | string | — | internal reference name (required) |
| `--vendor-identifier` | string | — | vendor identifier (required) |

#### `asc gamecenter leaderboard-sets delete-image`

Delete a leaderboard set image

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | image id (required) |

#### `asc gamecenter leaderboard-sets list`

List the app's leaderboard sets

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter leaderboard-sets localize`

Set a leaderboard set's localized name

| flag | type | default | description |
| --- | --- | --- | --- |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name (required) |
| `--set-id` | string | — | gameCenterLeaderboardSet id (required) |

#### `asc gamecenter leaderboard-sets member-localizations`

Manage per-set localized names of member leaderboards

A leaderboard set member localization overrides a leaderboard's display
name within a specific leaderboard set. The API addresses them by the
(leaderboard set, leaderboard) pair, so both ids are always required.

##### `asc gamecenter leaderboard-sets member-localizations list`

List member localizations for a (set, leaderboard) pair

| flag | type | default | description |
| --- | --- | --- | --- |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id (required) |
| `--set-id` | string | — | gameCenterLeaderboardSet id (required) |

##### `asc gamecenter leaderboard-sets member-localizations set`

Create or update a member localization for a locale

Example:

```
asc gamecenter leaderboard-sets member-localizations set \
    --set-id <set id> --leaderboard-id <leaderboard id> --locale ja --name "週間スコア"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized name of the leaderboard within the set (required) |
| `--set-id` | string | — | gameCenterLeaderboardSet id (required) |

#### `asc gamecenter leaderboard-sets upload-image`

Upload the image for a leaderboard set localization

Upload a leaderboard set image (reserve → upload → commit, v2) and attach
it to a set localization. Find the localization id via the set's latest
version: GET /v2/gameCenterLeaderboardSets/{id}/versions then
GET /v2/gameCenterLeaderboardSetVersions/{id}/localizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | image file (required) |
| `--localization-id` | string | — | localization id (required) |

### `asc gamecenter leaderboards`

Manage Game Center leaderboards (v2, versioned)

Aliases: leaderboard

#### `asc gamecenter leaderboards create`

Create a leaderboard (an initial version is created inline)

Example:

```
asc gamecenter leaderboards create --app 6790641087 \
    --reference-name "High Scores" --vendor-identifier com.example.high_scores \
    --score-sort-type DESC --default-formatter INTEGER --submission-type BEST_SCORE
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--default-formatter` | string | `INTEGER` | score formatter, e.g. INTEGER, DECIMAL_POINT_1_PLACE, ELAPSED_TIME_SECOND, MONEY_YEN |
| `--reference-name` | string | — | internal reference name (required) |
| `--score-range-end` | string | — | maximum accepted score |
| `--score-range-start` | string | — | minimum accepted score |
| `--score-sort-type` | string | — | ASC or DESC (required) |
| `--submission-type` | string | `BEST_SCORE` | BEST_SCORE or MOST_RECENT_SCORE |
| `--vendor-identifier` | string | — | vendor identifier, e.g. com.example.high_scores (required) |

#### `asc gamecenter leaderboards delete`

Delete a leaderboard

| flag | type | default | description |
| --- | --- | --- | --- |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id (required) |

#### `asc gamecenter leaderboards list`

List the app's leaderboards

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc gamecenter leaderboards localize`

Set a leaderboard's localized name and formatting

Set a leaderboard's localized name, description and score formatting.
Localizations belong to the leaderboard's latest version. Creating a new
localization requires --name; updates accept any subset.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--description` | string | — | localized description (@file allowed) |
| `--formatter-override` | string | — | per-locale formatter override |
| `--formatter-suffix` | string | — | score suffix, e.g. " points" |
| `--formatter-suffix-singular` | string | — | singular score suffix, e.g. " point" |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name |

#### `asc gamecenter leaderboards update`

Update a leaderboard's attributes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--archived` | bool | `false` | archive the leaderboard |
| `--default-formatter` | string | — | score formatter |
| `--leaderboard-id` | string | — | gameCenterLeaderboard id (required) |
| `--reference-name` | string | — | internal reference name |
| `--score-range-end` | string | — | maximum accepted score |
| `--score-range-start` | string | — | minimum accepted score |
| `--score-sort-type` | string | — | ASC or DESC |
| `--submission-type` | string | — | BEST_SCORE or MOST_RECENT_SCORE |

#### `asc gamecenter leaderboards upload-image`

Upload the image for a leaderboard localization

Upload a leaderboard image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the leaderboard's latest version:
GET /v2/gameCenterLeaderboards/{id}/versions then
GET /v2/gameCenterLeaderboardVersions/{id}/localizations.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--file` | string | — | image file (required) |
| `--localization-id` | string | — | localization id (required) |

### `asc gamecenter matchmaking`

Manage Game Center matchmaking (queues, rule sets, rules, teams)

#### `asc gamecenter matchmaking queues`

Manage matchmaking queues

Aliases: queue

##### `asc gamecenter matchmaking queues create`

Create a matchmaking queue

| flag | type | default | description |
| --- | --- | --- | --- |
| `--classic-bundle-ids` | stringSlice | `[]` | bundle ids using classic matchmaking (comma-separated) |
| `--experiment-rule-set-id` | string | — | experimental rule set id |
| `--reference-name` | string | — | queue reference name (required) |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id (required) |

##### `asc gamecenter matchmaking queues delete`

Delete a matchmaking queue

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingQueue id (required) |

##### `asc gamecenter matchmaking queues list`

List matchmaking queues

##### `asc gamecenter matchmaking queues update`

Update a matchmaking queue (rule sets, classic bundle ids)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--classic-bundle-ids` | stringSlice | `[]` | bundle ids using classic matchmaking (comma-separated) |
| `--experiment-rule-set-id` | string | — | experimental rule set id |
| `--id` | string | — | gameCenterMatchmakingQueue id (required) |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id |

#### `asc gamecenter matchmaking rule-sets`

Manage matchmaking rule sets

Aliases: rule-set

##### `asc gamecenter matchmaking rule-sets create`

Create a matchmaking rule set

| flag | type | default | description |
| --- | --- | --- | --- |
| `--max-players` | int | `0` | maximum players per match (required) |
| `--min-players` | int | `0` | minimum players per match (required) |
| `--reference-name` | string | — | rule set reference name (required) |
| `--rule-language-version` | int | `1` | rule expression language version |

##### `asc gamecenter matchmaking rule-sets delete`

Delete a matchmaking rule set

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingRuleSet id (required) |

##### `asc gamecenter matchmaking rule-sets list`

List matchmaking rule sets

##### `asc gamecenter matchmaking rule-sets test`

Test a rule set against sample matchmaking requests

POST /v1/gameCenterMatchmakingRuleSetTests. The request body is deeply
nested (a matchmakingRuleSet relationship, a matchmakingRequests to-many
relationship, and included gameCenterMatchmakingTestRequests /
gameCenterMatchmakingTestPlayerProperties inline resources), so pass the
complete JSON:API request document via --body (inline JSON or @file):

  {
    "data": {
      "type": "gameCenterMatchmakingRuleSetTests",
      "relationships": {
        "matchmakingRuleSet": {"data": {"type": "gameCenterMatchmakingRuleSets", "id": "..."}},
        "matchmakingRequests": {"data": [{"type": "gameCenterMatchmakingTestRequests", "id": "${req1}"}]}
      }
    },
    "included": [
      {"type": "gameCenterMatchmakingTestRequests", "id": "${req1}",
       "attributes": {"requestName": "r1", "secondsInQueue": 0, "bundleId": "com.example.app",
                      "platform": "IOS", "appVersion": "1.0", "playerCount": 2}}
    ]
  }

| flag | type | default | description |
| --- | --- | --- | --- |
| `--body` | string | — | full JSON:API request document, inline or @file (required) |

##### `asc gamecenter matchmaking rule-sets update`

Update a matchmaking rule set's player counts

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingRuleSet id (required) |
| `--max-players` | int | `0` | maximum players per match |
| `--min-players` | int | `0` | minimum players per match |

#### `asc gamecenter matchmaking rules`

Manage matchmaking rules within a rule set

Aliases: rule

##### `asc gamecenter matchmaking rules create`

Create a matchmaking rule

Example:

```
asc gamecenter matchmaking rules create --rule-set-id <id> \
    --reference-name skill --type MATCH --description "Match by skill" \
    --expression 'abs(requests[0].properties.skill - requests[1].properties.skill) < 10' --weight 0.5
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--description` | string | — | rule description (required) |
| `--expression` | string | — | rule expression (@file allowed) (required) |
| `--reference-name` | string | — | rule reference name (required) |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id (required) |
| `--type` | string | — | COMPATIBLE, DISTANCE, MATCH or TEAM (required) |
| `--weight` | float64 | `0` | rule weight |

##### `asc gamecenter matchmaking rules delete`

Delete a matchmaking rule

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingRule id (required) |

##### `asc gamecenter matchmaking rules list`

List the rules of a rule set

| flag | type | default | description |
| --- | --- | --- | --- |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id (required) |

##### `asc gamecenter matchmaking rules update`

Update a matchmaking rule

| flag | type | default | description |
| --- | --- | --- | --- |
| `--description` | string | — | rule description |
| `--expression` | string | — | rule expression (@file allowed) |
| `--id` | string | — | gameCenterMatchmakingRule id (required) |
| `--weight` | float64 | `0` | rule weight |

#### `asc gamecenter matchmaking teams`

Manage matchmaking teams within a rule set

Aliases: team

##### `asc gamecenter matchmaking teams create`

Create a matchmaking team

| flag | type | default | description |
| --- | --- | --- | --- |
| `--max-players` | int | `0` | maximum players per team (required) |
| `--min-players` | int | `0` | minimum players per team (required) |
| `--reference-name` | string | — | team reference name (required) |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id (required) |

##### `asc gamecenter matchmaking teams delete`

Delete a matchmaking team

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingTeam id (required) |

##### `asc gamecenter matchmaking teams list`

List the teams of a rule set

| flag | type | default | description |
| --- | --- | --- | --- |
| `--rule-set-id` | string | — | gameCenterMatchmakingRuleSet id (required) |

##### `asc gamecenter matchmaking teams update`

Update a matchmaking team's player counts

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | gameCenterMatchmakingTeam id (required) |
| `--max-players` | int | `0` | maximum players per team |
| `--min-players` | int | `0` | minimum players per team |

### `asc gamecenter show`

Show the app's Game Center detail

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc gamecenter submit-achievement`

Submit achievement progress on behalf of a player (server-to-server)

POST /v1/gameCenterPlayerAchievementSubmissions. Reports a player's
achievement progress (0-100 percent) for a scoped player id. Requires an API
key authorized for server-to-server Game Center submissions.

Example:

```
asc gamecenter submit-achievement --bundle-id com.example.app \
    --achievement-vendor-id com.example.first_win --scoped-player-id <id> --percentage-achieved 100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--achievement-vendor-id` | string | — | achievement vendor identifier (required) |
| `--bundle-id` | string | — | app bundle id (required) |
| `--challenge-ids` | stringSlice | `[]` | challenge ids the progress counts toward (comma-separated) |
| `--percentage-achieved` | int | `0` | progress percentage, 0-100 (required) |
| `--pre-released` | bool | `false` | submit against the pre-release (unpublished) achievement |
| `--scoped-player-id` | string | — | scoped player id (required) |
| `--submitted-date` | string | — | submission time, RFC 3339 (default: now, server-side) |

### `asc gamecenter submit-score`

Submit a leaderboard score on behalf of a player (server-to-server)

POST /v1/gameCenterLeaderboardEntrySubmissions. Submits a score to a
leaderboard for a scoped player id. The score is passed as a string (64-bit
integer). Requires an API key authorized for server-to-server Game Center
submissions.

Example:

```
asc gamecenter submit-score --bundle-id com.example.app \
    --leaderboard-vendor-id com.example.high_scores --scoped-player-id <id> --score 12345
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--bundle-id` | string | — | app bundle id (required) |
| `--challenge-ids` | stringSlice | `[]` | challenge ids the score counts toward (comma-separated) |
| `--context` | string | — | leaderboard context (64-bit integer string) |
| `--leaderboard-vendor-id` | string | — | leaderboard vendor identifier (required) |
| `--pre-released` | bool | `false` | submit against the pre-release (unpublished) leaderboard |
| `--scoped-player-id` | string | — | scoped player id (required) |
| `--score` | string | — | score as a 64-bit integer string (required) |
| `--submitted-date` | string | — | submission time, RFC 3339 (default: now, server-side) |

## `asc iap`

Manage in-app purchases (localizations, price, review screenshot, submit)

### `asc iap availability`

Show or set the territories where an in-app purchase is available

#### `asc iap availability set`

Set the territories where an in-app purchase is available

Example:

```
asc iap availability set --app 6790641087 --product com.example.app.credits100 \
    --territories JPN,USA --available-in-new-territories
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--available-in-new-territories` | bool | `false` | automatically become available in future new territories |
| `--product` | string | — | product id or ASC in-app purchase id (required) |
| `--territories` | string | — | comma-separated territory codes, e.g. JPN,USA (required) |

#### `asc iap availability show`

Show an in-app purchase's territory availability

Example:

```
asc iap availability show --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap content`

Show an in-app purchase's hosted content info (file name, size, download URL)

Show the Apple-hosted content of an in-app purchase
(GET /v2/inAppPurchases/{id}/content). Read-only: the spec exposes no create or
update operations for inAppPurchaseContents; hosted content is uploaded with
Xcode or Transporter.

Example:

```
asc iap content --app com.example.app --product com.example.app.premiumpack
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap create`

Create an in-app purchase

Example:

```
asc iap create --app 6790641087 --product-id com.example.app.credits100 \
    --name "100 Credits" --type CONSUMABLE --review-note "Unlocks 100 analysis credits."
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--family-sharable` | bool | `false` | allow Family Sharing |
| `--name` | string | — | reference name shown in App Store Connect (required) |
| `--product-id` | string | — | new product id, e.g. com.example.app.credits100 (required) |
| `--review-note` | string | — | note for App Review (@file allowed) |
| `--type` | string | — | CONSUMABLE, NON_CONSUMABLE or NON_RENEWING_SUBSCRIPTION (required) |

### `asc iap delete`

Delete an in-app purchase

Example:

```
asc iap delete --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap delete-image`

Delete an in-app purchase promotional image by id

Example:

```
asc iap delete-image --id 12345678-1234-1234-1234-123456789012
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | inAppPurchaseImage id (required) |

### `asc iap list`

List the app's in-app purchases

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc iap list-images`

List an in-app purchase's promotional images

Example:

```
asc iap list-images --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap localize`

Set an in-app purchase's localized display name and description

Example:

```
asc iap localize --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --locale ja --name "AI解析チケット（100枚）" \
    --description "領収書のAI解析に使えるチケット。1枚の解析につき1枚消費します。"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--description` | string | — | description (max 45 chars, @file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | display name (max 30 chars) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap offer-codes`

Manage in-app purchase offer codes

#### `asc iap offer-codes create`

Create an offer code for an in-app purchase

Create an offer code. The offer price is resolved from --price (customer
price) and --territory to an in-app purchase price point; Apple derives the
other territories' offer prices from it. After creating, generate redeemable
codes with "asc iap offer-codes create-codes".

Example:

```
asc iap offer-codes create --app 6790641087 --product com.example.app.credits100 \
    --name "Summer Promo" --territory JPN --price 100 --eligibility NON_SPENDER,CHURNED_SPENDER
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--eligibility` | stringSlice | `[NON_SPENDER,ACTIVE_SPENDER,CHURNED_SPENDER]` | customer eligibilities: NON_SPENDER, ACTIVE_SPENDER, CHURNED_SPENDER |
| `--name` | string | — | offer code reference name (required) |
| `--price` | string | — | offer customer price in the territory currency (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |
| `--territory` | string | `JPN` | territory code the offer price is defined in |

#### `asc iap offer-codes create-codes`

Generate one-time-use codes for an offer code

Example:

```
asc iap offer-codes create-codes --id 12345678-1234-1234-1234-123456789012 \
    --number 100 --expiration-date 2026-12-31
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--expiration-date` | string | — | expiration date YYYY-MM-DD (required) |
| `--id` | string | — | inAppPurchaseOfferCode id (required) |
| `--number` | int | `0` | number of codes to generate (required) |

#### `asc iap offer-codes custom-codes`

Manage custom (multi-redemption) codes for an in-app purchase offer code

##### `asc iap offer-codes custom-codes create`

Create a custom code for an in-app purchase offer code

Example:

```
asc iap offer-codes custom-codes create --id 12345678-1234-1234-1234-123456789012 --code SPRING24 --count 500
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--code` | string | — | custom code value, e.g. SPRING24 (required) |
| `--count` | int | `0` | number of redemptions the code allows (required) |
| `--expiration` | string | — | expiration date YYYY-MM-DD (optional) |
| `--id` | string | — | inAppPurchaseOfferCode id (required) |

##### `asc iap offer-codes custom-codes list`

List an in-app purchase offer code's custom codes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | inAppPurchaseOfferCode id (required) |

#### `asc iap offer-codes deactivate`

Deactivate an offer code

Example:

```
asc iap offer-codes deactivate --id 12345678-1234-1234-1234-123456789012
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | inAppPurchaseOfferCode id (required) |

#### `asc iap offer-codes download-codes`

Download a one-time-use code batch as CSV

Download the code values of a one-time-use code batch (created with
"create-codes") as CSV. Code generation is asynchronous; retry shortly after
creation if the values are not ready yet.

Example:

```
asc iap offer-codes download-codes --id 12345678-1234-1234-1234-123456789012 --output codes.csv
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | one-time-use code batch id from create-codes (required) |
| `--output` | string | — | write CSV to this file (default: stdout) |

#### `asc iap offer-codes list`

List an in-app purchase's offer codes

Example:

```
asc iap offer-codes list --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap price`

Set an in-app purchase's price by customer price in a base territory

Set the price by matching a customer price in a base territory to an App
Store price point, then creating a price schedule effective immediately.

Apple derives every other territory's price from the base territory's price
point. --price is the customer-facing price in the territory's currency
(e.g. 150 for ¥150 with --territory JPN).

Example:

```
asc iap price --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --territory JPN --price 150
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--price` | string | — | customer price in the territory currency, e.g. 150 (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |
| `--territory` | string | `JPN` | base territory code, e.g. JPN, USA |

### `asc iap schedule`

Show an in-app purchase's price schedule (base territory and manual prices)

Example:

```
asc iap schedule --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap screenshot`

Upload the review screenshot for an in-app purchase

Upload the App Review screenshot for an in-app purchase.

IMPORTANT: the IAP review-screenshot validator only accepts LEGACY device
sizes — current App Store screenshot sizes (1290×2796, 1320×2868, 1284×2778,
1170×2532, ...) are all rejected asynchronously with IMAGE_INCORRECT_DIMENSIONS.
Known-accepted sizes: 1242×2208, 2208×1242, 2048×2732, 2732×2048. The command
validates dimensions before uploading; pass --auto-fit to scale and pad the
image to an accepted size automatically. After upload it waits for Apple's
validation and reports the IAP state (READY_TO_SUBMIT when metadata is done).

Example:

```
asc iap screenshot --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --file app-store/iap-review.png --auto-fit
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--auto-fit` | bool | `false` | scale and pad the image to an accepted review-screenshot size (keeps aspect ratio, white padding) |
| `--file` | string | — | review screenshot file (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap show`

Show an in-app purchase's attributes and state

Example:

```
asc iap show --app 6790641087 --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap submit`

Submit an in-app purchase for review

Submit an in-app purchase for review on its own (create an
inAppPurchaseSubmission). For a first app submission, IAPs are usually reviewed
together with the app version; use this for IAPs added or changed later.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap upload-image`

Upload a promotional image for an in-app purchase

Example:

```
asc iap upload-image --app 6790641087 --product com.example.app.credits100 \
    --file app-store/iap-promo.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--file` | string | — | image file (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

### `asc iap version`

Manage in-app purchase draft versions (to edit a live in-app purchase)

#### `asc iap version create`

Create a draft version of an in-app purchase

Create an inAppPurchaseVersion, a draft for editing an in-app purchase that
is already live. The create request carries no attributes, only the
inAppPurchase relationship.

Example:

```
asc iap version create --app com.example.app --product com.example.app.credits100
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--product` | string | — | product id or ASC in-app purchase id (required) |

#### `asc iap version show`

Show an in-app purchase version with its localizations and images

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | inAppPurchaseVersion id (required) |

## `asc invitations`

Manage user invitations

### `asc invitations cancel`

Cancel a pending user invitation

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | string | — | invitee email |
| `--id` | string | — | invitation id |

### `asc invitations create`

Invite a new user to the team

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all-apps-visible` | bool | `false` | grant access to all apps |
| `--app-ids` | stringSlice | `[]` | apps to make visible (app ids or bundle ids); required for app-scoped roles without --all-apps-visible |
| `--email` | string | — | invitee email (required) |
| `--first-name` | string | — | invitee first name (required) |
| `--last-name` | string | — | invitee last name (required) |
| `--roles` | string | — | comma-separated roles: ADMIN, FINANCE, ACCOUNT_HOLDER, SALES, MARKETING, APP_MANAGER, DEVELOPER, ACCESS_TO_REPORTS, CUSTOMER_SUPPORT, CREATE_APPS, CLOUD_MANAGED_DEVELOPER_ID, CLOUD_MANAGED_APP_DISTRIBUTION, GENERATE_INDIVIDUAL_KEYS (required) |

### `asc invitations list`

List pending user invitations

## `asc marketplace`

Manage alternative app marketplace search details and webhooks

### `asc marketplace search-detail`

Marketplace search detail (catalog URL) of a marketplace app

#### `asc marketplace search-detail set`

Set the marketplace catalog URL of an app (creates or updates the search detail)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--url` | string | — | catalog URL (required) |

#### `asc marketplace search-detail show`

Show the marketplace search detail of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc marketplace webhook`

Marketplace webhooks (deprecated by Apple in favor of the webhooks API)

#### `asc marketplace webhook create`

Create a marketplace webhook

| flag | type | default | description |
| --- | --- | --- | --- |
| `--secret` | string | — | signing secret, or @file (required) |
| `--url` | string | — | endpoint URL (required) |

#### `asc marketplace webhook delete`

Delete a marketplace webhook

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | marketplaceWebhook id (required) |

#### `asc marketplace webhook list`

List marketplace webhooks

#### `asc marketplace webhook update`

Update a marketplace webhook (only the flags you pass are changed)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | marketplaceWebhook id (required) |
| `--secret` | string | — | new signing secret, or @file |
| `--url` | string | — | new endpoint URL |

## `asc merchant-ids`

Manage Apple Pay merchant IDs

### `asc merchant-ids create`

Register a new Apple Pay merchant ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--identifier` | string | — | merchant identifier, e.g. merchant.com.example (required) |
| `--name` | string | — | merchant ID name (required) |

### `asc merchant-ids delete`

Delete an Apple Pay merchant ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | merchant ID resource id (required) |

### `asc merchant-ids list`

List Apple Pay merchant IDs

## `asc metrics`

Download Xcode performance and power metrics

### `asc metrics power`

Download perfPowerMetrics for an app or a build

Download performance and power metrics (perfPowerMetrics) for an app or a
specific build. The response is Apple's xcode-metrics JSON (not JSON:API) and
is pretty-printed to stdout.

Example:

```
asc metrics power --app com.example.app
  asc metrics power --build <build id> --metric-type LAUNCH --device-type iPhone15,2
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id |
| `--build` | string | — | build id (instead of --app) |
| `--device-type` | string | — | filter by device type, e.g. iPhone15,2 |
| `--metric-type` | string | — | filter: DISK, HANG, BATTERY, LAUNCH, MEMORY, ANIMATION, TERMINATION, STORAGE |
| `--platform` | string | — | filter by platform: IOS |

## `asc nominations`

Work with featuring nominations (suggest content to the App Store editorial team)

Manage featuring nominations: pitches to the App Store editorial team for
featuring an app launch, app enhancements, or new content.

A nomination is created as a DRAFT and submitted by setting submitted=true
(pass --submit on create, or run: nominations update --id <id> --submit).
Dates are ISO 8601 date-times, e.g. 2026-08-01T00:00:00Z.

### `asc nominations create`

Create a nomination (a draft unless --submit is passed)

Example:

```
asc nominations create --app 6790641087 --type NEW_CONTENT \
    --name "Summer update featuring" --description @nomination.txt \
    --publish-start 2026-08-01T00:00:00Z --submit
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--description` | string | — | pitch to the editorial team (@file allowed) |
| `--device-families` | stringSlice | `[]` | device families: IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, VISION |
| `--launch-in-select-markets-first` | bool | `false` | the app launches in select markets first |
| `--locales` | stringSlice | `[]` | locales the content is available in |
| `--name` | string | — | nomination name |
| `--notes` | string | — | additional notes |
| `--publish-end` | string | — | publish end date (ISO 8601) |
| `--publish-start` | string | — | publish start date (ISO 8601) |
| `--related-apps` | stringSlice | `[]` | additional related app ids |
| `--submit` | bool | `false` | submit the nomination (submitted=true) |
| `--supplemental-materials-uris` | stringSlice | `[]` | URLs of supplemental materials (comma-separated or repeatable) |
| `--type` | string | — | type: NEW_CONTENT, APP_LAUNCH, APP_ENHANCEMENTS |

### `asc nominations delete`

Delete a nomination

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | nomination id (required) |

### `asc nominations list`

List nominations, optionally filtered to one app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | limit to nominations related to this app id or bundle id |
| `--state` | string | `DRAFT,SUBMITTED,ARCHIVED` | comma-separated states: DRAFT, SUBMITTED, ARCHIVED |

### `asc nominations show`

Show a nomination

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | nomination id (required) |

### `asc nominations update`

Update a nomination; --submit submits it, --archive archives it

Example:

```
asc nominations update --id <id> --submit
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--archive` | bool | `false` | archive the nomination (archived=true) |
| `--description` | string | — | pitch to the editorial team (@file allowed) |
| `--device-families` | stringSlice | `[]` | device families: IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, VISION |
| `--id` | string | — | nomination id (required) |
| `--launch-in-select-markets-first` | bool | `false` | the app launches in select markets first |
| `--locales` | stringSlice | `[]` | locales the content is available in |
| `--name` | string | — | nomination name |
| `--notes` | string | — | additional notes |
| `--publish-end` | string | — | publish end date (ISO 8601) |
| `--publish-start` | string | — | publish start date (ISO 8601) |
| `--related-apps` | stringSlice | `[]` | additional related app ids |
| `--submit` | bool | `false` | submit the nomination (submitted=true) |
| `--supplemental-materials-uris` | stringSlice | `[]` | URLs of supplemental materials (comma-separated or repeatable) |
| `--type` | string | — | type: NEW_CONTENT, APP_LAUNCH, APP_ENHANCEMENTS |

## `asc pass-type-ids`

Manage Wallet pass type IDs

### `asc pass-type-ids create`

Register a new Wallet pass type ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--identifier` | string | — | pass type identifier, e.g. pass.com.example.coupon (required) |
| `--name` | string | — | pass type name (required) |

### `asc pass-type-ids delete`

Delete a Wallet pass type ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | pass type ID resource id (required) |

### `asc pass-type-ids list`

List Wallet pass type IDs

## `asc pricing`

Manage the app's price schedule and price points

### `asc pricing points`

List app price points for a territory

Example:

```
asc pricing points --app 6790641087 --territory JPN --limit 30
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--limit` | int | `0` | maximum number of price points to print (default: all) |
| `--territory` | string | `JPN` | territory code, e.g. JPN, USA |

### `asc pricing set`

Set the app's price by customer price in a base territory

Set the price by matching a customer price in a base territory to an App
Store price point, then creating a price schedule. Apple derives every other
territory's price from the base territory's price point. --price is the
customer-facing price in the territory's currency (e.g. 300 for ¥300 with
--base-territory JPN). Without --start-date the price takes effect immediately.

Free apps also need a price schedule (a required submission item that is easy
to miss): pass --free to resolve the customerPrice "0" price point.

Example:

```
asc pricing set --app 6790641087 --free
  asc pricing set --app 6790641087 --price 300 --base-territory JPN --start-date 2026-08-01
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--base-territory` | string | `USA` | base territory code, e.g. USA, JPN |
| `--free` | bool | `false` | make the app free (shorthand for --price 0) |
| `--price` | string | — | customer price in the base territory currency, e.g. 0.99 or 300 |
| `--start-date` | string | — | start date YYYY-MM-DD (default: effective immediately) |

### `asc pricing show`

Show the app's price schedule (base territory and manual prices)

Example:

```
asc pricing show --app 6790641087 --automatic
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--automatic` | bool | `false` | also list automatic (derived) prices for every territory |

## `asc product-pages`

Work with custom product pages (create, localize, upload screenshots)

Manage custom product pages (appCustomProductPages): alternate versions of the
App Store product page reachable via their own URL, e.g. for ad campaigns.

A page owns versions (appCustomProductPageVersions); each version owns per-locale
localizations, which own screenshot sets. Page versions are submitted for review
via: asc review-submissions add-item ...

### `asc product-pages create`

Create a custom product page

Create a custom product page. With --from-version-id, the page's first version
is copied from an existing appStoreVersion (the appStoreVersionTemplate
relationship); otherwise create a version afterwards with create-version.

Example:

```
asc product-pages create --app 6790641087 --name "Campaign A"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--from-version-id` | string | — | appStoreVersion id to copy metadata from (appStoreVersionTemplate) |
| `--name` | string | — | page name (required) |

### `asc product-pages create-version`

Create a new version of a custom product page

| flag | type | default | description |
| --- | --- | --- | --- |
| `--deep-link` | string | — | deep link URL for this page version |
| `--page-id` | string | — | appCustomProductPage id (required) |

### `asc product-pages delete`

Delete a custom product page

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appCustomProductPage id (required) |

### `asc product-pages list`

List the app's custom product pages

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc product-pages localize`

Create or update a page version's localization

Create the appCustomProductPageLocalization for a locale on a page version (or
update its promotional text). The localization id printed on success is what
upload-screenshot takes as --localization-id.

Example:

```
asc product-pages localize --version-id <id> --locale ja --promotional-text "夏のキャンペーン実施中"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--promotional-text` | string | — | promotional text (@file allowed) |
| `--version-id` | string | — | appCustomProductPageVersion id (required) |

### `asc product-pages show`

Show a custom product page and its versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appCustomProductPage id (required) |

### `asc product-pages upload-screenshot`

Upload screenshots to a page localization

Example:

```
asc product-pages upload-screenshot --localization-id <id> --display APP_IPHONE_67 \
    --file app-store/cpp/01.png --file app-store/cpp/02.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--display` | string | — | screenshot display type, e.g. APP_IPHONE_67 (required) |
| `--file` | stringArray | `[]` | screenshot file (repeatable, applied in order) |
| `--localization-id` | string | — | appCustomProductPageLocalization id (required) |

### `asc product-pages versions`

List a custom product page's versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--page-id` | string | — | appCustomProductPage id (required) |

## `asc profiles`

Manage credential profiles

### `asc profiles list`

List configured profiles

### `asc profiles remove`

Remove a profile (the key file is kept)

```
asc profiles remove <name>
```

### `asc profiles use`

Set the default profile

```
asc profiles use <name>
```

## `asc provisioning-profiles`

Manage provisioning profiles

### `asc provisioning-profiles create`

Create a provisioning profile

Example:

```
asc provisioning-profiles create --name "My App Store" --type IOS_APP_STORE \
    --bundle-id-id ABC123 --certificate-ids CERT1
  asc provisioning-profiles create --name "My Dev" --type IOS_APP_DEVELOPMENT \
    --bundle-id-id ABC123 --certificate-ids CERT1,CERT2 --device-ids DEV1,DEV2
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--bundle-id-id` | string | — | bundle ID resource id (required) |
| `--certificate-ids` | string | — | comma-separated certificate ids (required) |
| `--device-ids` | string | — | comma-separated device ids (for development/ad hoc profiles) |
| `--name` | string | — | profile name (required) |
| `--type` | string | — | profile type: IOS_APP_DEVELOPMENT, IOS_APP_STORE, IOS_APP_ADHOC, IOS_APP_INHOUSE, MAC_APP_DEVELOPMENT, MAC_APP_STORE, MAC_APP_DIRECT, TVOS_APP_DEVELOPMENT, TVOS_APP_STORE, TVOS_APP_ADHOC, TVOS_APP_INHOUSE, MAC_CATALYST_APP_DEVELOPMENT, MAC_CATALYST_APP_STORE, MAC_CATALYST_APP_DIRECT (required) |

### `asc provisioning-profiles delete`

Delete a provisioning profile

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | profile id (required) |

### `asc provisioning-profiles download`

Download a provisioning profile as a .mobileprovision file

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | profile id (required) |
| `--output` | string | — | output path (default: <name>.mobileprovision) |

### `asc provisioning-profiles list`

List provisioning profiles

| flag | type | default | description |
| --- | --- | --- | --- |
| `--type` | string | — | filter by profile type: IOS_APP_DEVELOPMENT, IOS_APP_STORE, IOS_APP_ADHOC, IOS_APP_INHOUSE, MAC_APP_DEVELOPMENT, MAC_APP_STORE, MAC_APP_DIRECT, TVOS_APP_DEVELOPMENT, TVOS_APP_STORE, TVOS_APP_ADHOC, TVOS_APP_INHOUSE, MAC_CATALYST_APP_DEVELOPMENT, MAC_CATALYST_APP_STORE, MAC_CATALYST_APP_DIRECT |

## `asc reports`

Download sales and finance reports

### `asc reports finance`

Download a finance report (TSV)

Download a finance report from GET /v1/financeReports.

--region is a two-character financial region code (e.g. JP, US, or ZZ for the
consolidated report; Z1 with --type FINANCE_DETAIL). --date is the fiscal
period, e.g. 2026-06. The report arrives gzip-compressed and is decompressed
by default; pass --gzip to keep the raw .gz bytes.

The vendor number is shown in App Store Connect under Payments and Financial
Reports.

Example:

```
asc reports finance --vendor 8XXXXXXX --region ZZ --date 2026-06
  asc reports finance --vendor 8XXXXXXX --region Z1 --type FINANCE_DETAIL --date 2026-06 --output detail.tsv
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date` | string | — | fiscal period, e.g. 2026-06 (required) |
| `--gzip` | bool | `false` | keep the raw gzip bytes instead of decompressing |
| `--output` | string | — | write to file instead of stdout |
| `--region` | string | — | financial region code, e.g. JP, US, ZZ (required) |
| `--type` | string | `FINANCIAL` | report type: FINANCIAL or FINANCE_DETAIL |
| `--vendor` | string | — | vendor number from Payments and Financial Reports (required) |

### `asc reports sales`

Download a sales and trends report (TSV)

Download a sales report from GET /v1/salesReports.

The report arrives gzip-compressed. By default it is decompressed and the
tab-separated content is written to stdout (or --output); pass --gzip to keep
the raw .gz bytes instead.

The vendor number is shown in App Store Connect under Payments and Financial
Reports (an 8-digit number starting with 8).

Example:

```
asc reports sales --vendor 8XXXXXXX --date 2026-07-15
  asc reports sales --vendor 8XXXXXXX --type SUBSCRIPTION --subtype SUMMARY --frequency DAILY --version 1_3 --output subs.tsv
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date` | string | — | report date, e.g. 2026-07-15 (daily), 2026-07 (monthly); latest when omitted |
| `--frequency` | string | `DAILY` | frequency: DAILY, WEEKLY, MONTHLY, YEARLY |
| `--gzip` | bool | `false` | keep the raw gzip bytes instead of decompressing |
| `--output` | string | — | write to file instead of stdout |
| `--subtype` | string | `SUMMARY` | report subtype: SUMMARY, DETAILED, SUMMARY_INSTALL_TYPE, SUMMARY_TERRITORY, SUMMARY_CHANNEL |
| `--type` | string | `SALES` | report type: SALES, PRE_ORDER, NEWSSTAND, SUBSCRIPTION, SUBSCRIPTION_EVENT, SUBSCRIBER, SUBSCRIPTION_OFFER_CODE_REDEMPTION, INSTALLS, FIRST_ANNUAL, WIN_BACK_ELIGIBILITY |
| `--vendor` | string | — | vendor number from Payments and Financial Reports (required) |
| `--version` | string | — | report version, e.g. 1_0, 1_1 |

## `asc review-detail`

Manage the App Review contact info and notes for the editable version

### `asc review-detail attach`

Upload a file attachment for App Review

Upload an attachment (screenshot, video, document) to the version's App
Review detail. The review detail must exist; create it with
"asc review-detail set" first.

Example:

```
asc review-detail attach --app 6790641087 --file app-store/demo-walkthrough.mp4
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--file` | string | — | attachment file (required) |
| `--version` | string | — | version string (default: the editable version) |

### `asc review-detail attachments`

List App Review attachments for the editable version

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--version` | string | — | version string (default: the editable version) |

### `asc review-detail delete-attachment`

Delete an App Review attachment by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | appStoreReviewAttachment id (required) |

### `asc review-detail set`

Create or update the App Review contact info, demo account and notes

Example:

```
asc review-detail set --app 6790641087 \
    --first 邦彦 --last 宮永 --phone +81-50-3159-6143 --email contact@ideamans.com \
    --notes @app-store/review-notes.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--demo-pass` | string | — | demo account password |
| `--demo-required` | bool | `false` | whether a demo account is required |
| `--demo-user` | string | — | demo account username |
| `--email` | string | — | contact email |
| `--first` | string | — | contact first name |
| `--last` | string | — | contact last name |
| `--notes` | string | — | review notes (@file allowed) |
| `--phone` | string | — | contact phone |
| `--version` | string | — | version string (default: the editable version) |

### `asc review-detail show`

Show the current App Review detail

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--version` | string | — | version string (default: the editable version) |

## `asc review-submissions`

Inspect and manage App Store review submissions

### `asc review-submissions add-item`

Add content to an open review submission

Add one piece of reviewable content to a review submission. Pass exactly one
of --version-id (appStoreVersions id), --event-id (appEvents id),
--experiment-id (appStoreVersionExperiments id, submitted as the V2
relationship) or --cpp-version-id (appCustomProductPageVersions id).

Example:

```
asc review-submissions add-item --id 8b7e01b6-... --version-id d3e7a1c2-...
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--cpp-version-id` | string | — | appCustomProductPageVersions id to submit |
| `--event-id` | string | — | appEvents id to submit |
| `--experiment-id` | string | — | appStoreVersionExperiments id to submit |
| `--id` | string | — | review submission id (required) |
| `--version-id` | string | — | appStoreVersions id to submit |

### `asc review-submissions cancel`

Cancel a review submission that has not completed review

Example:

```
asc review-submissions cancel --id 8b7e01b6-...
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | review submission id (required) |

### `asc review-submissions list`

List the app's review submissions

Example:

```
asc review-submissions list --app 6790641087
  asc review-submissions list --app 6790641087 --platform IOS
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | filter by platform: IOS, MAC_OS, TV_OS, VISION_OS |

### `asc review-submissions remove-item`

Remove an item from its review submission

Example:

```
asc review-submissions remove-item --id <reviewSubmissionItem id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | review submission item id (required) |

### `asc review-submissions show`

Show a review submission and its items

Example:

```
asc review-submissions show --id 8b7e01b6-...
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | review submission id (required) |

## `asc reviews`

Work with App Store customer reviews and responses

### `asc reviews delete-response`

Delete the developer response to a customer review

Delete a customer review response. --id accepts either the review id (its
response is looked up) or the response id itself.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | review id or response id (required) |

### `asc reviews list`

List an app's customer reviews

Example:

```
asc reviews list --app com.example.app
  asc reviews list --app 6790641087 --territory JPN --rating 1 --sort createdDate
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--rating` | int | `0` | filter by star rating (1-5) |
| `--sort` | string | `-createdDate` | sort: createdDate, -createdDate, rating, -rating |
| `--territory` | string | — | filter by territory code, e.g. JPN, USA |

### `asc reviews respond`

Create or replace the developer response to a customer review

Example:

```
asc reviews respond --id <review id> --body "Thank you for the feedback!"
  asc reviews respond --id <review id> --body @response.txt
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--body` | string | — | response body (@file allowed) (required) |
| `--id` | string | — | customer review id (required) |

### `asc reviews show`

Show a customer review's full body and any developer response

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | customer review id (required) |

## `asc sandbox`

Manage sandbox testers (list, update, clear purchase history)

Manage App Store sandbox test accounts via the v2 sandboxTesters API.
The API only supports listing, updating, and clearing purchase history;
creating or deleting sandbox testers must be done in App Store Connect.

### `asc sandbox clear-history`

Clear a sandbox tester's purchase history

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | sandbox tester id (required) |

### `asc sandbox list`

List sandbox testers

### `asc sandbox update`

Update a sandbox tester (territory, interrupted purchases, renewal rate)

Example:

```
asc sandbox update --id <tester-id> --territory USA --interrupt-purchases
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | sandbox tester id (required) |
| `--interrupt-purchases` | bool | `false` | simulate interrupted purchases |
| `--renewal-rate` | string | — | subscription renewal rate, e.g. MONTHLY_RENEWAL_EVERY_ONE_HOUR |
| `--territory` | string | — | territory code, e.g. JPN, USA |

## `asc submit`

Submit the app version for App Store review

Submit the editable version for review using the reviewSubmissions flow:
find or create an open review submission for the platform, add the version as a
submission item, then mark the submission submitted.

Preconditions (not checked here; the API will reject if unmet): metadata and
localizations complete, screenshots uploaded, a build attached, pricing set,
age rating set, and — done in the web UI — the App Privacy questionnaire and,
for paid in-app purchases, the Paid Apps agreement. Use --prepare-only to stage
the submission without finalizing.

Example:

```
asc submit --app 6790641087
  asc submit --app 6790641087 --version 1.0 --prepare-only
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | `IOS` | platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--prepare-only` | bool | `false` | stage the submission item without finalizing |
| `--version` | string | — | version string (default: the editable version) |

## `asc subscriptions`

Manage auto-renewable subscriptions (groups, prices, offers, availability, review)

Aliases: subs

### `asc subscriptions availability`

Show or set the territories where a subscription is available

#### `asc subscriptions availability set`

Set the territories where a subscription is available

Example:

```
asc subscriptions availability set --app com.example.app --sub com.example.app.pro.monthly \
    --territories JPN,USA --available-in-new-territories
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--available-in-new-territories` | bool | `true` | automatically include future App Store territories |
| `--plan-type` | string | `MONTHLY` | subscription plan type: MONTHLY or UPFRONT |
| `--sub` | string | — | subscription id or productId (required) |
| `--territories` | stringSlice | `[]` | territory codes, e.g. JPN,USA (required) |

#### `asc subscriptions availability show`

Show a subscription's territory availability

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions create`

Create an auto-renewable subscription in a group

Example:

```
asc subscriptions create --app com.example.app --group-id 21489000 \
    --product-id com.example.app.pro.monthly --name "Pro Monthly" --period P1M
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--family-sharable` | bool | `false` | enable Family Sharing |
| `--group-id` | string | — | subscription group id (required) |
| `--group-level` | int | `0` | group level (1 = highest tier) |
| `--name` | string | — | reference name (required) |
| `--period` | string | — | subscription period: P1W, P1M, P2M, P3M, P6M or P1Y (required) |
| `--product-id` | string | — | product id, e.g. com.example.app.pro.monthly (required) |
| `--review-note` | string | — | review note (@file allowed) |

### `asc subscriptions delete`

Delete a subscription (only possible before it is approved)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions grace-period`

Show or configure the app's billing grace period

#### `asc subscriptions grace-period set`

Update the app's billing grace period settings

Example:

```
asc subscriptions grace-period set --app com.example.app \
    --opt-in --duration P16D --renewal-type ALL_RENEWALS
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--duration` | string | — | grace period duration: P3D, P16D or P28D |
| `--opt-in` | bool | `false` | enable the grace period in production |
| `--renewal-type` | string | — | ALL_RENEWALS or PAID_TO_PAID_ONLY |
| `--sandbox-opt-in` | bool | `false` | enable the grace period in sandbox |

#### `asc subscriptions grace-period show`

Show the app's billing grace period settings

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc subscriptions groups`

Manage subscription groups and their localizations

#### `asc subscriptions groups create`

Create a subscription group

Example:

```
asc subscriptions groups create --app com.example.app --name "Premium Plans"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--name` | string | — | group reference name (required) |

#### `asc subscriptions groups delete`

Delete a subscription group

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | subscription group id (required) |

#### `asc subscriptions groups list`

List the app's subscription groups

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc subscriptions groups localize`

Set a subscription group's localized display name

Example:

```
asc subscriptions groups localize --group-id 21489000 --locale ja --name "プレミアムプラン"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--custom-app-name` | string | — | custom app name shown with the group |
| `--group-id` | string | — | subscription group id (required) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized group display name |

#### `asc subscriptions groups submit`

Submit a subscription group (its display names) for review

Example:

```
asc subscriptions groups submit --group-id 21489000
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | subscription group id (required) |

#### `asc subscriptions groups version`

Manage subscription group draft versions

##### `asc subscriptions groups version create`

Create a draft version of a subscription group

Example:

```
asc subscriptions groups version create --group-id 21489000
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--group-id` | string | — | subscription group id (required) |

##### `asc subscriptions groups version show`

Show a subscription group version with its localizations

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | subscriptionGroupVersion id (required) |

### `asc subscriptions image`

Manage a subscription's promotional image

#### `asc subscriptions image delete`

Delete a subscription image by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionImage id (required) |

#### `asc subscriptions image list`

List a subscription's images

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

#### `asc subscriptions image upload`

Upload a promotional image for a subscription

Upload a promotional image via the reserve/upload/commit flow
(POST /v1/subscriptionImages relates directly to the subscription; the commit
sets uploaded plus sourceFileChecksum).

Example:

```
asc subscriptions image upload --app com.example.app --sub com.example.app.pro.monthly \
    --file app-store/sub-promo.png
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--file` | string | — | image file (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions intro-offer`

Manage introductory offers (free trial, pay as you go, pay up front)

#### `asc subscriptions intro-offer delete`

Delete an introductory offer by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionIntroductoryOffer id (required) |

#### `asc subscriptions intro-offer list`

List a subscription's introductory offers

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

#### `asc subscriptions intro-offer set`

Create an introductory offer

Create an introductory offer. FREE_TRIAL needs no price; PAY_AS_YOU_GO and
PAY_UP_FRONT require --territory and --price to resolve a subscription price
point. Without --territory a FREE_TRIAL applies to all territories.

Example:

```
asc subscriptions intro-offer set --app com.example.app --sub com.example.app.pro.monthly \
    --duration P1W --offer-mode FREE_TRIAL
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--duration` | string | — | offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required) |
| `--end-date` | string | — | end date YYYY-MM-DD |
| `--offer-mode` | string | — | FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required) |
| `--periods` | int | `1` | number of periods |
| `--price` | string | — | customer price in the territory currency (paid modes) |
| `--start-date` | string | — | start date YYYY-MM-DD |
| `--sub` | string | — | subscription id or productId (required) |
| `--territory` | string | `JPN` | territory code (paid modes; optional for FREE_TRIAL) |

### `asc subscriptions list`

List all subscriptions across the app's subscription groups

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc subscriptions localize`

Set a subscription's localized display name and description

Example:

```
asc subscriptions localize --app com.example.app --sub com.example.app.pro.monthly \
    --locale ja --name "プロ（月額）" --description "全機能が使える月額プラン"
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--description` | string | — | localized description (@file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--name` | string | — | localized display name |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions offer-codes`

Manage subscription offer codes and their one-time-use codes

#### `asc subscriptions offer-codes create`

Create an offer code campaign

Create an offer code with a price for one territory (Apple equalizes the
remaining territories from that price point). FREE_TRIAL needs no --price.

Example:

```
asc subscriptions offer-codes create --app com.example.app --sub com.example.app.pro.monthly \
    --name "Launch trial" --duration P1M --offer-mode FREE_TRIAL --periods 1 --territory JPN
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--customer-eligibility` | stringSlice | `[NEW,EXISTING,EXPIRED]` | customer eligibility: NEW, EXISTING, EXPIRED |
| `--duration` | string | — | offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required) |
| `--eligibility` | string | `STACK_WITH_INTRO_OFFERS` | STACK_WITH_INTRO_OFFERS or REPLACE_INTRO_OFFERS |
| `--name` | string | — | campaign name (required) |
| `--offer-mode` | string | — | FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required) |
| `--periods` | int | `1` | number of periods |
| `--price` | string | — | customer price in the territory currency (paid modes) |
| `--sub` | string | — | subscription id or productId (required) |
| `--territory` | string | `JPN` | territory code for the price |

#### `asc subscriptions offer-codes create-codes`

Generate a batch of one-time-use codes for an offer code campaign

Example:

```
asc subscriptions offer-codes create-codes --id 21500000 --count 100 --expiration 2026-12-31
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--count` | int | `0` | number of codes to generate (required) |
| `--expiration` | string | — | expiration date YYYY-MM-DD (required) |
| `--id` | string | — | subscriptionOfferCode id (required) |

#### `asc subscriptions offer-codes custom-codes`

Manage custom (multi-redemption) codes for a subscription offer code campaign

##### `asc subscriptions offer-codes custom-codes create`

Create a custom code for an offer code campaign

Example:

```
asc subscriptions offer-codes custom-codes create --id 21500000 --code SPRING24 --count 500 --expiration 2026-12-31
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--code` | string | — | custom code value, e.g. SPRING24 (required) |
| `--count` | int | `0` | number of redemptions the code allows (required) |
| `--expiration` | string | — | expiration date YYYY-MM-DD (optional) |
| `--id` | string | — | subscriptionOfferCode id (required) |

##### `asc subscriptions offer-codes custom-codes list`

List an offer code campaign's custom codes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionOfferCode id (required) |

#### `asc subscriptions offer-codes deactivate`

Deactivate an offer code campaign by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionOfferCode id (required) |

#### `asc subscriptions offer-codes download-codes`

Download a one-time-use code batch as CSV

Download the code values of a one-time-use code batch (the id printed by
create-codes, or listed via
"asc api /v1/subscriptionOfferCodes/{id}/oneTimeUseCodes"). Prints CSV to
stdout, or writes to --output.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionOfferCodeOneTimeUseCodes batch id (required) |
| `--output` | string | — | write CSV to this file instead of stdout |

#### `asc subscriptions offer-codes list`

List a subscription's offer code campaigns

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions price`

Manage subscription prices

#### `asc subscriptions price list`

List a subscription's scheduled and current prices

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

#### `asc subscriptions price set`

Set a subscription's price by customer price in a territory

Set the price by matching a customer price in a territory to a subscription
price point, then creating a subscriptionPrice. Apple derives the other
territories' prices from this price point. Without --start-date the change is
scheduled immediately; --preserve-current keeps the current price for existing
subscribers.

Example:

```
asc subscriptions price set --app com.example.app --sub com.example.app.pro.monthly \
    --territory JPN --price 600
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--preserve-current` | bool | `false` | keep the current price for existing subscribers |
| `--price` | string | — | customer price in the territory currency, e.g. 600 (required) |
| `--start-date` | string | — | start date YYYY-MM-DD (default: immediately) |
| `--sub` | string | — | subscription id or productId (required) |
| `--territory` | string | `JPN` | territory code, e.g. JPN, USA |

### `asc subscriptions promo-offer`

Manage promotional offers for existing subscribers

#### `asc subscriptions promo-offer create`

Create a promotional offer

Create a promotional offer with a price for one territory (Apple equalizes
the remaining territories from that price point). FREE_TRIAL needs no --price.

Example:

```
asc subscriptions promo-offer create --app com.example.app --sub com.example.app.pro.monthly \
    --name "Win back 50%" --offer-code WINBACK50 --duration P1M --offer-mode PAY_AS_YOU_GO \
    --periods 3 --territory JPN --price 300
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--duration` | string | — | offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required) |
| `--name` | string | — | offer reference name (required) |
| `--offer-code` | string | — | offer identifier used by your app, e.g. WINBACK50 (required) |
| `--offer-mode` | string | — | FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required) |
| `--periods` | int | `1` | number of periods |
| `--price` | string | — | customer price in the territory currency (paid modes) |
| `--sub` | string | — | subscription id or productId (required) |
| `--territory` | string | `JPN` | territory code for the price |

#### `asc subscriptions promo-offer delete`

Delete a promotional offer by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | subscriptionPromotionalOffer id (required) |

#### `asc subscriptions promo-offer list`

List a subscription's promotional offers

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions promoted`

Manage promoted in-app purchases and subscriptions on the App Store

#### `asc subscriptions promoted list`

List the app's promoted purchases

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

#### `asc subscriptions promoted set`

Promote a subscription or in-app purchase (creates or updates the promoted purchase)

Example:

```
asc subscriptions promoted set --app com.example.app --sub com.example.app.pro.monthly \
    --visible-for-all-users
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--enabled` | bool | `false` | enable or disable the promotion |
| `--iap` | string | — | in-app purchase product id or ASC id to promote |
| `--sub` | string | — | subscription id or productId to promote |
| `--visible-for-all-users` | bool | `true` | visible to all users (not only past purchasers) |

### `asc subscriptions screenshot`

Upload the review screenshot for a subscription

Upload the App Review screenshot for a subscription. Like IAP review
screenshots, the validator only accepts legacy sizes (1242×2208, 2208×1242,
2048×2732, 2732×2048) — current App Store screenshot sizes are rejected
asynchronously. Dimensions are validated before uploading; pass --auto-fit to
scale and pad automatically.

Example:

```
asc subscriptions screenshot --app com.example.app --sub com.example.app.pro.monthly \
    --file app-store/sub-review.png --auto-fit
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--auto-fit` | bool | `false` | scale and pad the image to an accepted review-screenshot size (keeps aspect ratio, white padding) |
| `--file` | string | — | review screenshot file (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions show`

Show a subscription as JSON

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions submit`

Submit a subscription for review

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions update`

Update a subscription's name, period, group level, family sharing or review note

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--family-sharable` | bool | `false` | enable Family Sharing |
| `--group-level` | int | `0` | group level (1 = highest tier) |
| `--name` | string | — | reference name |
| `--period` | string | — | subscription period: P1W, P1M, P2M, P3M, P6M or P1Y |
| `--review-note` | string | — | review note (@file allowed) |
| `--sub` | string | — | subscription id or productId (required) |

### `asc subscriptions version`

Manage subscription draft versions (to edit a live subscription)

#### `asc subscriptions version create`

Create a draft version of a subscription

Create a subscriptionVersion, a draft for editing a subscription that is
already live. The create request carries no attributes, only the subscription
relationship.

Example:

```
asc subscriptions version create --app com.example.app --sub com.example.app.pro.monthly
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

#### `asc subscriptions version show`

Show a subscription version with its localizations and images

| flag | type | default | description |
| --- | --- | --- | --- |
| `--version-id` | string | — | subscriptionVersion id (required) |

### `asc subscriptions win-back`

List or delete win-back offers (create is not supported; see help)

List or delete a subscription's win-back offers. Creating win-back offers is
not supported here: the API's inline price schema for winBackOffers carries no
territory or price-point relationships, so a reliable create cannot be built
from the published spec. Create them in App Store Connect, or use "asc api"
with a hand-crafted body.

#### `asc subscriptions win-back delete`

Delete a win-back offer by id

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | winBackOffer id (required) |

#### `asc subscriptions win-back list`

List a subscription's win-back offers

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--sub` | string | — | subscription id or productId (required) |

## `asc territories`

List App Store territories and their currencies

## `asc token`

Print a signed JWT for use with curl etc.

Example:

```
curl -H "Authorization: Bearer $(asc token)" https://api.appstoreconnect.apple.com/v1/apps
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--ttl` | duration | `15m0s` | token lifetime (max 20m) |

## `asc users`

Manage users on the team

### `asc users apps`

List the apps visible to a user

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | string | — | user email (Apple ID) |
| `--id` | string | — | user id |

### `asc users list`

List all users on the team

### `asc users remove`

Remove a user from the team

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | string | — | user email (Apple ID) |
| `--id` | string | — | user id |

### `asc users set-apps`

Replace the set of apps visible to a user

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app-ids` | string | — | comma-separated app ids to make visible (required) |
| `--email` | string | — | user email (Apple ID) |
| `--id` | string | — | user id |

### `asc users show`

Show a user by id or email

| flag | type | default | description |
| --- | --- | --- | --- |
| `--email` | string | — | user email (Apple ID) |
| `--id` | string | — | user id |

### `asc users update`

Update a user's roles and permissions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all-apps-visible` | bool | `false` | grant access to all apps |
| `--email` | string | — | user email (Apple ID) |
| `--id` | string | — | user id |
| `--provisioning-allowed` | bool | `false` | allow access to Certificates, Identifiers & Profiles |
| `--roles` | string | — | comma-separated roles: ADMIN, FINANCE, ACCOUNT_HOLDER, SALES, MARKETING, APP_MANAGER, DEVELOPER, ACCESS_TO_REPORTS, CUSTOMER_SUPPORT, CREATE_APPS, CLOUD_MANAGED_DEVELOPER_ID, CLOUD_MANAGED_APP_DISTRIBUTION, GENERATE_INDIVIDUAL_KEYS |

## `asc version`

Work with App Store versions (create, localize, select build)

### `asc version create`

Create a new App Store version

Example:

```
asc version create --app 6790641087 --version 1.0
  asc version create --app 6790641087 --version 1.1 --platform IOS
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | `IOS` | platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string, e.g. 1.0 (required) |

### `asc version delete`

Delete an App Store version (requires an explicit --version)

Example:

```
asc version delete --app 6790641087 --version 1.1
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

### `asc version list`

List the app's App Store versions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc version localize`

Set localized version metadata (description, keywords, what's new, URLs)

Create or update the appStoreVersionLocalization for a locale on the app's
editable version. Any flag may take @file to read the value from a file.

On an app's FIRST version, whatsNew (release notes) is not editable — the API
409s. This command skips whatsNew with a warning in that case and still applies
the other attributes.

Example:

```
asc version localize --app 6790641087 --locale ja \
    --description @app-store/description.txt \
    --keywords "領収書,レシート,経費,Excel,スキャン" \
    --promo "撮るだけでExcelに。" \
    --support-url https://japan-receipt-scan.web.app/
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--description` | string | — | description (@file allowed) |
| `--keywords` | string | — | comma-separated keywords, max 100 chars (@file allowed) |
| `--locale` | string | `ja` | locale, e.g. ja / en-US |
| `--marketing-url` | string | — | marketing URL |
| `--promo` | string | — | promotional text (@file allowed) |
| `--support-url` | string | — | support URL |
| `--version` | string | — | version string (default: the editable version) |
| `--whats-new` | string | — | what's new in this version (@file allowed) |

### `asc version phased-release`

Manage the version's phased release (gradual rollout over 7 days)

#### `asc version phased-release complete`

Finish the rollout and release to all users immediately

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

#### `asc version phased-release create`

Enable phased release for the version

Example:

```
asc version phased-release create --app 6790641087
  asc version phased-release create --app 6790641087 --state ACTIVE
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--state` | string | — | initial state: INACTIVE or ACTIVE (default: server default) |
| `--version` | string | — | version string (default: picked by command) |

#### `asc version phased-release delete`

Remove the phased release (version releases to everyone at once)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

#### `asc version phased-release pause`

Pause the phased release

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

#### `asc version phased-release resume`

Resume a paused phased release

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

#### `asc version phased-release show`

Show the version's phased release status

Example:

```
asc version phased-release show --app 6790641087
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

### `asc version release`

Release an approved version that is pending developer release

Request the release of a version in PENDING_DEVELOPER_RELEASE state (i.e. the
version was approved with releaseType MANUAL). Without --version, the version
currently pending developer release is used.

Example:

```
asc version release --app 6790641087
  asc version release --app 6790641087 --version 1.0
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

### `asc version set-build`

Attach an uploaded build to the editable version

Select which uploaded build the version ships. The build must already be
uploaded (via Xcode or Transporter) and finished processing. --build accepts a
build id or a build version string (the CFBundleVersion, e.g. "42").

Example:

```
asc version set-build --app 6790641087 --build 42
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--build` | string | — | build id or build version string (required) |
| `--version` | string | — | version string (default: the editable version) |

### `asc version show`

Show a version's state, release settings and attached build

Example:

```
asc version show --app 6790641087
  asc version show --app 6790641087 --version 1.0
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--version` | string | — | version string (default: picked by command) |

### `asc version update`

Update the editable version's copyright, release type or version string

Example:

```
asc version update --app 6790641087 --copyright "2026 ideaman's Inc."
  asc version update --app 6790641087 --release-type SCHEDULED --earliest-release-date 2026-08-01T00:00:00Z
  asc version update --app 6790641087 --version 1.0 --version-string 1.0.1
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--copyright` | string | — | copyright line, e.g. "2026 ideaman's Inc." (@file allowed) |
| `--earliest-release-date` | string | — | earliest release date (ISO8601, requires --release-type SCHEDULED) |
| `--platform` | string | — | restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS |
| `--release-type` | string | — | release type: MANUAL, AFTER_APPROVAL, SCHEDULED |
| `--version` | string | — | version string (default: picked by command) |
| `--version-string` | string | — | rename the version string |

## `asc webhooks`

Manage App Store Connect webhooks and inspect their deliveries

Create, update, ping, and delete webhooks for an app, and inspect delivery attempts.

Event types:
  ALTERNATIVE_DISTRIBUTION_PACKAGE_AVAILABLE_UPDATED
  ALTERNATIVE_DISTRIBUTION_PACKAGE_VERSION_CREATED
  ALTERNATIVE_DISTRIBUTION_TERRITORY_AVAILABILITY_UPDATED
  APP_STORE_VERSION_APP_VERSION_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_APP_STORE_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_EXTERNAL_BETA_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_INTERNAL_BETA_RELEASE_CREATED
  BACKGROUND_ASSET_VERSION_STATE_UPDATED
  BETA_FEEDBACK_CRASH_SUBMISSION_CREATED
  BETA_FEEDBACK_SCREENSHOT_SUBMISSION_CREATED
  BUILD_BETA_DETAIL_EXTERNAL_BUILD_STATE_UPDATED
  BUILD_UPLOAD_STATE_UPDATED

### `asc webhooks create`

Create a webhook for an app

Create a webhook. The secret signs each delivery payload (X-Apple-CK-Signature).

Event types:
  ALTERNATIVE_DISTRIBUTION_PACKAGE_AVAILABLE_UPDATED
  ALTERNATIVE_DISTRIBUTION_PACKAGE_VERSION_CREATED
  ALTERNATIVE_DISTRIBUTION_TERRITORY_AVAILABILITY_UPDATED
  APP_STORE_VERSION_APP_VERSION_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_APP_STORE_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_EXTERNAL_BETA_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_INTERNAL_BETA_RELEASE_CREATED
  BACKGROUND_ASSET_VERSION_STATE_UPDATED
  BETA_FEEDBACK_CRASH_SUBMISSION_CREATED
  BETA_FEEDBACK_SCREENSHOT_SUBMISSION_CREATED
  BUILD_BETA_DETAIL_EXTERNAL_BUILD_STATE_UPDATED
  BUILD_UPLOAD_STATE_UPDATED

Example:

```
asc webhooks create --app com.example.app --name ci-hook \
    --url https://example.com/hook --secret @secret.txt \
    --event-types APP_STORE_VERSION_APP_VERSION_STATE_UPDATED,BUILD_UPLOAD_STATE_UPDATED
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |
| `--enabled` | bool | `true` | create the webhook enabled |
| `--event-types` | stringSlice | `[]` | comma-separated event types (required); see command help for values |
| `--name` | string | — | webhook name (required) |
| `--secret` | string | — | signing secret, or @file to read it from a file (required) |
| `--url` | string | — | HTTPS endpoint URL (required) |

### `asc webhooks delete`

Delete a webhook

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | webhook id (required) |

### `asc webhooks deliveries`

List delivery attempts of a webhook

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | webhook id (required) |
| `--limit` | int | `50` | maximum number of deliveries to list (max 200) |
| `--state` | string | — | filter by delivery state: SUCCEEDED, FAILED, or PENDING |

### `asc webhooks delivery`

Show one delivery attempt of a webhook (full request/response details)

Show a single webhook delivery as JSON, including the request URL, the
response HTTP status code and body, the delivery state, and any error message.

The API has no per-delivery endpoint, so the delivery is located by scanning
the webhook's deliveries; both --webhook-id and --id are required.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | webhookDelivery id (required) |
| `--webhook-id` | string | — | webhook id the delivery belongs to (required) |

### `asc webhooks list`

List webhooks of an app

| flag | type | default | description |
| --- | --- | --- | --- |
| `--app` | string | — | app id or bundle id (required) |

### `asc webhooks ping`

Send a test ping event to a webhook

Send a ping event to the webhook's URL. The result of the attempt shows up as
a delivery; check it with "asc webhooks deliveries --id <webhook-id>".

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | webhook id (required) |

### `asc webhooks redeliver`

Redeliver a webhook event using a past delivery as the template

Ask App Store Connect to send the event of a previous delivery again. The
original delivery is referenced as the "template" of the new one (find
delivery ids with "asc webhooks deliveries --id <webhook-id>").

Example:

```
asc webhooks redeliver --delivery-id DELIVERY_ID
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--delivery-id` | string | — | webhookDelivery id to redeliver (required) |

### `asc webhooks update`

Update a webhook (only the flags you pass are changed)

Update a webhook's name, URL, secret, enabled state, or event types.

Event types:
  ALTERNATIVE_DISTRIBUTION_PACKAGE_AVAILABLE_UPDATED
  ALTERNATIVE_DISTRIBUTION_PACKAGE_VERSION_CREATED
  ALTERNATIVE_DISTRIBUTION_TERRITORY_AVAILABILITY_UPDATED
  APP_STORE_VERSION_APP_VERSION_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_APP_STORE_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_EXTERNAL_BETA_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_INTERNAL_BETA_RELEASE_CREATED
  BACKGROUND_ASSET_VERSION_STATE_UPDATED
  BETA_FEEDBACK_CRASH_SUBMISSION_CREATED
  BETA_FEEDBACK_SCREENSHOT_SUBMISSION_CREATED
  BUILD_BETA_DETAIL_EXTERNAL_BUILD_STATE_UPDATED
  BUILD_UPLOAD_STATE_UPDATED

| flag | type | default | description |
| --- | --- | --- | --- |
| `--enabled` | bool | `true` | enable (true) or disable (false) the webhook |
| `--event-types` | stringSlice | `[]` | new comma-separated event types |
| `--id` | string | — | webhook id (required) |
| `--name` | string | — | new name |
| `--secret` | string | — | new signing secret, or @file |
| `--url` | string | — | new endpoint URL |

## `asc xcode-cloud`

Work with Xcode Cloud products, workflows, and build runs

Inspect Xcode Cloud CI products and workflows, start build runs, and
retrieve run results (actions, artifacts, issues, test results).

Aliases: ci

### `asc xcode-cloud artifacts`

List (and optionally download) artifacts of a build action

List the artifacts produced by a build action (use "asc xcode-cloud runs show"
to find action ids). With --download-dir each artifact's download URL is
fetched and the file is saved into that directory.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action-id` | string | — | ciBuildAction id (required) |
| `--download-dir` | string | — | directory to download artifact files into |

### `asc xcode-cloud environment`

Browse the build environments Xcode Cloud offers (macOS and Xcode versions)

#### `asc xcode-cloud environment macos-versions`

List macOS versions available to Xcode Cloud builds

#### `asc xcode-cloud environment xcode-versions`

List Xcode versions available to Xcode Cloud builds

List the Xcode versions Xcode Cloud can build with. With --id a single
Xcode version is shown as JSON, including its testDestinations (simulator
devices and runtimes available for testing).

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciXcodeVersion id (show one version, incl. test destinations) |

### `asc xcode-cloud issues`

List issues (errors, warnings, test failures) of a build action

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action-id` | string | — | ciBuildAction id (required) |

### `asc xcode-cloud products`

List, show, and delete Xcode Cloud products

#### `asc xcode-cloud products delete`

Delete an Xcode Cloud product (removes all its workflows and build data)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciProduct id (required) |

#### `asc xcode-cloud products list`

List Xcode Cloud products

#### `asc xcode-cloud products show`

Show an Xcode Cloud product

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciProduct id (required) |

### `asc xcode-cloud run`

Start a new build run for a workflow

Start an Xcode Cloud build run. With --branch or --tag the name is resolved
to a git reference of the workflow's repository and used as the source;
otherwise Xcode Cloud builds the workflow's default reference.

Example:

```
asc xcode-cloud run --workflow-id WORKFLOW_ID --branch main
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--branch` | string | — | branch name to build (default: the workflow's default reference) |
| `--clean` | bool | `false` | perform a clean build (no derived data cache) |
| `--tag` | string | — | tag name to build |
| `--workflow-id` | string | — | ciWorkflow id (required) |

### `asc xcode-cloud runs`

List and inspect build runs

#### `asc xcode-cloud runs list`

List build runs of a workflow or product (newest first)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--limit` | int | `25` | maximum number of runs to list (max 200) |
| `--product-id` | string | — | ciProduct id |
| `--workflow-id` | string | — | ciWorkflow id |

#### `asc xcode-cloud runs show`

Show a build run and its actions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciBuildRun id (required) |

### `asc xcode-cloud scm`

Browse Git providers and repositories connected to Xcode Cloud

#### `asc xcode-cloud scm providers`

Source control providers

##### `asc xcode-cloud scm providers list`

List source control providers

#### `asc xcode-cloud scm pull-requests`

List pull requests of a repository

List the pull requests Xcode Cloud knows about for a repository (find
repository ids with "asc xcode-cloud scm repositories list").

| flag | type | default | description |
| --- | --- | --- | --- |
| `--limit` | int | `50` | maximum number of pull requests to list (max 200) |
| `--repository-id` | string | — | scmRepository id (required) |

#### `asc xcode-cloud scm repositories`

Source control repositories

##### `asc xcode-cloud scm repositories list`

List repositories (all, or those of one provider)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--provider-id` | string | — | scmProvider id (default: all repositories) |

### `asc xcode-cloud tests`

List test results of a build action

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action-id` | string | — | ciBuildAction id (required) |

### `asc xcode-cloud workflows`

List, show, enable, and disable Xcode Cloud workflows

#### `asc xcode-cloud workflows disable`

Disable a workflow

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciWorkflow id (required) |

#### `asc xcode-cloud workflows enable`

Enable a workflow

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciWorkflow id (required) |

#### `asc xcode-cloud workflows list`

List workflows of a product

| flag | type | default | description |
| --- | --- | --- | --- |
| `--product-id` | string | — | ciProduct id (required) |

#### `asc xcode-cloud workflows show`

Show a workflow (start conditions, actions, environment)

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | string | — | ciWorkflow id (required) |
