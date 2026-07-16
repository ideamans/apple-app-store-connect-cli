package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var llmFlag bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&llmFlag, "llm", false, "print detailed help for LLM agents")
	rootCmd.Args = cobra.NoArgs
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if llmFlag {
			fmt.Print(llmHelp())
			return nil
		}
		return cmd.Help()
	}
	// Makes --llm work on any subcommand as well (asc apps list --llm).
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if llmFlag {
			fmt.Print(llmHelp())
			os.Exit(0)
		}
	}
}

func llmHelp() string {
	var b strings.Builder
	b.WriteString(`# asc — Apple App Store Connect API CLI (reference for LLM agents)

asc is a non-interactive CLI for the App Store Connect API
(https://api.appstoreconnect.apple.com). Every command reads flags and
environment variables only; nothing prompts for input, so it is safe to run
from scripts and agents. Human-readable summaries go to stdout, errors go to
stderr prefixed with "Error:", and the exit code is 0 on success / 1 on any
failure.

## Credential model

The API authenticates with a JWT (ES256) built from three values:
- issuer_id: UUID identifying the team (from App Store Connect > Users and Access > Integrations)
- key_id: 10-character ID of the API key
- private key: the downloaded .p8 file (PKCS#8 ECDSA). It cannot be re-downloaded from Apple.

A "profile" is a named triple of these values. Tokens are short-lived (max 20
minutes) and are generated fresh for every invocation; there is no token cache.

## Credential resolution order (first match wins)

1. Environment variables, when ASC_ISSUER_ID and a key source are both set:
   - ASC_ISSUER_ID + ASC_PRIVATE_KEY_PATH (key ID derived from the AuthKey_XXXXXXXXXX.p8 filename, or ASC_KEY_ID)
   - ASC_ISSUER_ID + ASC_PRIVATE_KEY_BASE64 + ASC_KEY_ID (for CI; base64 of the .p8 file contents)
2. Profile lookup: --profile flag, else ASC_PROFILE, else default_profile in config.toml.

.env files are never read implicitly.

## Configuration files

~/.config/apple-app-store-connect/   (or $XDG_CONFIG_HOME/apple-app-store-connect, mode 0700)
├── config.toml                      (mode 0600)
└── keys/AuthKey_XXXXXXXXXX.p8       (mode 0600)

config.toml format:

    default_profile = "default"

    [profiles.default]
    issuer_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    key_id = "XXXXXXXXXX"            # optional; derived from private_key filename if omitted
    private_key = "keys/AuthKey_XXXXXXXXXX.p8"  # absolute, or relative to the config dir

Profiles are created with "asc configure" (preferred; it copies the key and
sets permissions) but the file may also be edited directly.

## Global flags

`)
	b.WriteString(codeBlock(flagUsages(rootCmd.PersistentFlags())))
	b.WriteString("\n## Commands\n\n")
	writeCommandReference(&b, rootCmd)
	b.WriteString(`## Using "asc api" effectively

"asc api" is the escape hatch for the full API surface
(https://developer.apple.com/documentation/appstoreconnectapi). Notes:

- The API follows the JSON:API convention: responses have "data" (with "id",
  "type", "attributes", "relationships"), plus "included" and "links".
- Pagination: pass a "limit" query parameter (max 200 for most collections) and
  follow the absolute URL in links.next until it is absent.
- Common query parameters: filter[attr]=value, sort=name / sort=-name,
  fields[type]=a,b, include=relationship, limit=N. Quote the path in shells
  because of the brackets: asc api "/v1/apps?filter[bundleId]=com.example.app"
- Write requests take JSON:API bodies:
  asc api -X POST /v1/betaGroups -d '{"data":{"type":"betaGroups","attributes":{...},"relationships":{...}}}'
- Errors return {"errors":[{"status","code","title","detail"}]}; asc formats
  them into the stderr message and exits 1.
- HTTP 401 usually means wrong issuer_id / key_id / key file or a revoked key;
  403 means the key's role lacks permission for the endpoint.

## Known pitfalls (learned from real App Store submissions)

- **Asset uploads validate asynchronously.** A 2xx on reserve/PUT/commit does
  NOT mean acceptance; failures (e.g. IMAGE_INCORRECT_DIMENSIONS) appear later
  in assetDeliveryState. asc's image/video upload commands poll for
  COMPLETE/FAILED automatically and exit non-zero on FAILED with the error
  codes. Build and Background Asset uploads use different state models — watch
  them with "asc builds list" / "asc background-assets versions".
- **IAP/subscription review screenshots only accept legacy sizes** (1242x2208,
  2208x1242, 2048x2732, 2732x2048). Current App Store screenshot sizes
  (1290x2796 etc.) are rejected — the appScreenshots validator and the
  review-screenshot validator are different. "asc iap screenshot" /
  "asc subscriptions screenshot" validate dimensions before uploading and offer
  --auto-fit (aspect-preserving resize + white padding).
- **whatsNew is not editable on an app's first version** (409). "asc version
  localize" skips it with a warning and still applies other attributes.
- **IAPs often already exist** (created for sandbox testing): POST 409s on a
  duplicate productId. "asc iap create" is find-or-create; localize/price/
  screenshot address products by productId and work on existing ones.
- **Inline-created ("included") resources need local ids in the form
  "${name}"** — a plain string id gets 409 ENTITY_ERROR.INCLUDED.INVALID_ID
  (appPriceSchedules, appAvailabilities, offer codes, ...). asc does this
  internally; remember it when composing raw "asc api" POSTs.
- **App availability requires ALL territories explicitly** (~175
  territoryAvailabilities in one POST, available true/false each). "Japan only"
  still needs every other territory marked false. "asc availability set"
  expands the full list automatically.
- **Free apps still need a price schedule** (customerPrice "0" price point):
  "asc pricing set --free". Price and availability are required for submission
  yet remain unset even when the version page looks complete; "asc submit"
  warns when either is missing.
- **The management API needs an ASC API key with a role.** App Store Server
  API (In-App Purchase) keys from the same issuer return 401 here.
- **The newer age-rating flow may not be readable via the API** (GET 404 even
  when set); writes work — verify the rating in the UI.
- **GUI-only (no public API):** the App Privacy questionnaire (data collection
  labels; /v1/apps/{id}/dataUsages 404s), App Store Server Notifications URL
  configuration, and Japan's 特定商取引法 disclosure (no dedicated field; the
  App Info "trader status" field is the EU DSA one).

## Typical flows

Register credentials, then verify:
    asc configure --issuer-id <UUID> --key ~/Downloads/AuthKey_XXXXXXXXXX.p8
    asc apps list

Multiple teams:
    asc configure --profile client-a --issuer-id <UUID> --key <p8>
    asc --profile client-a apps list
    asc profiles use client-a

Ad-hoc curl with a signed token:
    curl -H "Authorization: Bearer $(asc token)" "https://api.appstoreconnect.apple.com/v1/apps"
`)
	return b.String()
}

func writeCommandReference(b *strings.Builder, parent *cobra.Command) {
	for _, c := range parent.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		if c.Runnable() {
			fmt.Fprintf(b, "### `%s`\n\n", c.UseLine())
			desc := c.Long
			if desc == "" {
				desc = c.Short
			}
			b.WriteString(strings.TrimSpace(desc) + "\n\n")
			if usages := flagUsages(c.NonInheritedFlags()); usages != "" {
				b.WriteString(codeBlock(usages))
			}
			if c.Example != "" {
				b.WriteString("Examples:\n")
				b.WriteString(codeBlock(strings.TrimRight(c.Example, "\n") + "\n"))
			}
		}
		writeCommandReference(b, c)
	}
}

func flagUsages(flags *pflag.FlagSet) string {
	filtered := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			filtered.AddFlag(f)
		}
	})
	if !filtered.HasAvailableFlags() {
		return ""
	}
	return filtered.FlagUsages()
}

func codeBlock(content string) string {
	return "```\n" + content + "```\n\n"
}
