---
name: asc-usage
description: Drive App Store Connect from the command line with the asc CLI — list apps, manage versions and builds, TestFlight groups, in-app purchases and subscriptions, pricing, territory availability, screenshots, age ratings and App Store submission. Use when the user asks about their iOS/macOS app on App Store Connect, an App Store release or review, TestFlight distribution, or App Store Connect API calls.
license: MIT
compatibility: Requires the `asc` binary on PATH — run the asc-install skill if it is missing. Needs an App Store Connect API key (issuer id, key id, .p8 file) registered with `asc configure`, or ASC_ISSUER_ID plus a key source in the environment. The API key must have a management role; App Store Server API keys return 401.
allowed-tools: Bash(asc:*) Bash(jq:*) Bash(command:*) Read Write
---

# asc-usage

Operate App Store Connect through the `asc` CLI. The command surface is large
(384 commands), and the API has traps that no amount of reasoning will reveal —
so the workflow is *read the reference, check the pitfalls, then act*.

## 1. Confirm the tool and the credentials

```bash
command -v asc && asc --version
```

Missing? Run the `asc-install` skill.

```bash
asc apps list
```

This is the cheapest call that proves the credentials work. A 401 usually means
the wrong issuer id / key id / key file, a revoked key, **or an App Store Server
API key being used for the management API** — those are different keys from the
same issuer. A 403 means the key's role lacks permission for that endpoint.

If nothing is configured, walk the user through:

```bash
asc configure --issuer-id <UUID> --key ~/Downloads/AuthKey_XXXXXXXXXX.p8
```

**Never print the contents of the `.p8` file and never pass a key inline** — it
would land in the transcript and in shell history. The `.p8` cannot be
re-downloaded from Apple, so also never move or delete the user's copy.

## 2. Read the reference before composing a command

```bash
asc llm | head -170            # conventions, credential model, api escape hatch
asc llm | sed -n '/# Known pitfalls/,/# Command catalog/p'
asc llm | grep -i 'asc iap'    # find the commands for a topic
```

The reference is embedded in the binary, so it matches the installed version
exactly. It is ~5,700 lines — grep it, do not paste it wholesale.

**The pitfalls chapter is the valuable part.** It records things learned from
real submissions that the API documentation does not state: asynchronous asset
validation, legacy-only screenshot sizes for IAP review, `whatsNew` being
unavailable on a first version, availability needing all ~175 territories, free
apps still needing a price schedule. Read it before any submission work.

## 3. Prefer a real command over `asc api`

Most tasks have a dedicated command that already handles the traps —
find-or-create for IAPs, dimension validation for screenshots, full-territory
expansion for availability. Reach for `asc api` only when nothing covers the
endpoint, and remember it takes JSON:API bodies and needs the path quoted
because of the brackets:

```bash
asc api "/v1/apps?filter[bundleId]=com.example.app"
```

## 4. Dry-run anything that mutates

```bash
asc <command> --dry-run
```

Submitting a version, changing price, changing availability and editing store
metadata are **customer-visible and hard to walk back**. Before running one:
state which app, which version and what will change, and get the user's
agreement. Then run it for real.

## 5. Report

Give back what changed and what state it is now in — the app id, the version
string, the review state. Asset uploads are the exception that needs saying out
loud: a 2xx does not mean acceptance, so quote the final `assetDeliveryState`
rather than declaring success on the HTTP status.

## Failure modes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `command not found: asc` | not installed | run the `asc-install` skill |
| 401 on every command | wrong key, revoked key, or a Server API key | verify with `asc apps list`; the management API needs an ASC key with a role |
| 403 on one endpoint | key role lacks permission | ask the user to raise the key's role in App Store Connect |
| upload returns 2xx then nothing works | asynchronous validation failed | check `assetDeliveryState`; for images use `--auto-fit` if dimensions were rejected |
| 409 on IAP create | product already exists from sandbox testing | `asc iap create` is find-or-create — proceed with localize/price/screenshot |
| 409 `INCLUDED.INVALID_ID` | inline-created resource used a plain id | use the `${name}` local-id form |
| `whatsNew` rejected with 409 | first version of the app | expected; the field is not editable yet |
| age rating reads back as 404 | newer flow is not readable via the API | writes still work; verify in the UI |
