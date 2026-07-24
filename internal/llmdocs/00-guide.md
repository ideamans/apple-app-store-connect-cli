# asc — Apple App Store Connect API CLI (reference for AI agents)

`asc` is a non-interactive CLI for the App Store Connect API
(<https://api.appstoreconnect.apple.com>). Every command reads flags and
environment variables only; nothing prompts for input, so it is safe to run from
scripts and agents. Human-readable summaries go to stdout, errors go to stderr
prefixed with `Error:`, and the exit code is 0 on success / 1 on any failure.

This reference is embedded in the binary — `asc llm` always describes the exact
version you are running.

## Credential model

The API authenticates with a JWT (ES256) built from three values:

- **issuer_id** — UUID identifying the team (App Store Connect → Users and
  Access → Integrations)
- **key_id** — 10-character ID of the API key
- **private key** — the downloaded `.p8` file (PKCS#8 ECDSA). **It cannot be
  re-downloaded from Apple.**

A *profile* is a named triple of these values. Tokens are short-lived (max 20
minutes) and are generated fresh for every invocation; there is no token cache.

### Resolution order (first match wins)

1. Environment variables, when `ASC_ISSUER_ID` and a key source are both set:
   - `ASC_ISSUER_ID` + `ASC_PRIVATE_KEY_PATH` (key ID derived from the
     `AuthKey_XXXXXXXXXX.p8` filename, or `ASC_KEY_ID`)
   - `ASC_ISSUER_ID` + `ASC_PRIVATE_KEY_BASE64` + `ASC_KEY_ID` (for CI; base64
     of the `.p8` file contents)
2. Profile lookup: `--profile` flag, else `ASC_PROFILE`, else `default_profile`
   in `config.toml`.

`.env` files are never read implicitly.

### Configuration files

```
~/.config/apple-app-store-connect/   (or $XDG_CONFIG_HOME/apple-app-store-connect, mode 0700)
├── config.toml                      (mode 0600)
└── keys/AuthKey_XXXXXXXXXX.p8       (mode 0600)
```

```toml
default_profile = "default"

[profiles.default]
issuer_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
key_id = "XXXXXXXXXX"                       # optional; derived from private_key filename if omitted
private_key = "keys/AuthKey_XXXXXXXXXX.p8"  # absolute, or relative to the config dir
```

Profiles are created with `asc configure` (preferred; it copies the key and sets
permissions) but the file may also be edited directly.

**Never print the contents of a `.p8` file, and never pass a private key on a
command line** — it lands in shell history and process listings. Point `asc
configure` at the file instead.

## Using `asc api` effectively

`asc api` is the escape hatch for the full API surface
(<https://developer.apple.com/documentation/appstoreconnectapi>).

- The API follows the JSON:API convention: responses have `data` (with `id`,
  `type`, `attributes`, `relationships`), plus `included` and `links`.
- **Pagination**: pass a `limit` query parameter (max 200 for most collections)
  and follow the absolute URL in `links.next` until it is absent.
- Common query parameters: `filter[attr]=value`, `sort=name` / `sort=-name`,
  `fields[type]=a,b`, `include=relationship`, `limit=N`. **Quote the path in
  shells** because of the brackets:
  `asc api "/v1/apps?filter[bundleId]=com.example.app"`
- Write requests take JSON:API bodies:
  `asc api -X POST /v1/betaGroups -d '{"data":{"type":"betaGroups","attributes":{...},"relationships":{...}}}'`
- Errors return `{"errors":[{"status","code","title","detail"}]}`; asc formats
  them into the stderr message and exits 1.
- HTTP 401 usually means wrong issuer_id / key_id / key file or a revoked key;
  403 means the key's role lacks permission for the endpoint.

## Typical flows

Register credentials, then verify:

```bash
asc configure --issuer-id <UUID> --key ~/Downloads/AuthKey_XXXXXXXXXX.p8
asc apps list
```

Multiple teams:

```bash
asc configure --profile client-a --issuer-id <UUID> --key <p8>
asc --profile client-a apps list
asc profiles use client-a
```

Ad-hoc curl with a signed token:

```bash
curl -H "Authorization: Bearer $(asc token)" "https://api.appstoreconnect.apple.com/v1/apps"
```

## Before doing anything destructive or public-facing

`--dry-run` prints mutating requests instead of sending them. Use it first for
anything that changes app metadata, pricing, availability or submission state.

Submitting a version, changing price or changing availability is visible to
customers and hard to walk back. Confirm with the user in prose before running
those, and say exactly which app and version you are about to touch.
