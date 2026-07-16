package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

// dryRun is wired to the persistent --dry-run flag; when set, mutating requests
// are printed instead of sent.
var dryRun bool

// newClient resolves credentials and returns a client honoring --dry-run.
func newClient() (*api.Client, error) {
	creds, err := config.Resolve(profileFlag)
	if err != nil {
		return nil, err
	}
	c := api.New(creds)
	c.DryRun = dryRun
	return c, nil
}

// editableStates are the appStoreVersion states whose metadata can still be edited.
var editableStates = map[string]bool{
	"PREPARE_FOR_SUBMISSION": true,
	"DEVELOPER_REJECTED":     true,
	"REJECTED":               true,
	"METADATA_REJECTED":      true,
	"INVALID_BINARY":         true,
}

// resolveAppID accepts a numeric app id (returned as-is) or a bundle id (looked up).
func resolveAppID(ctx context.Context, c *api.Client, appRef string) (string, error) {
	if appRef == "" {
		return "", fmt.Errorf("--app is required (numeric app id or bundle id)")
	}
	if isDigits(appRef) {
		return appRef, nil
	}
	apps, err := c.List(ctx, "/v1/apps?filter[bundleId]="+appRef+"&limit=1")
	if err != nil {
		return "", err
	}
	if len(apps) == 0 {
		return "", fmt.Errorf("no app found with bundle id %q", appRef)
	}
	return apps[0].ID, nil
}

// editableVersion returns the app's editable version. If versionString is given
// it must match; otherwise the first editable version is returned.
func editableVersion(ctx context.Context, c *api.Client, appID, versionString string) (*api.Resource, error) {
	versions, err := c.List(ctx, "/v1/apps/"+appID+"/appStoreVersions?limit=50")
	if err != nil {
		return nil, err
	}
	for i := range versions {
		v := &versions[i]
		if versionString != "" && v.Str("versionString") != versionString {
			continue
		}
		if versionString != "" || editableStates[v.Str("appStoreState")] {
			return v, nil
		}
	}
	if versionString != "" {
		return nil, fmt.Errorf("app %s has no version %q", appID, versionString)
	}
	return nil, fmt.Errorf("app %s has no editable version (states: %s)", appID, strings.Join(editableStateNames(), ", "))
}

// editableAppInfo returns the app's editable appInfo (the one being prepared).
func editableAppInfo(ctx context.Context, c *api.Client, appID string) (*api.Resource, error) {
	infos, err := c.List(ctx, "/v1/apps/"+appID+"/appInfos?limit=10")
	if err != nil {
		return nil, err
	}
	for i := range infos {
		if editableStates[infos[i].Str("appStoreState")] || editableStates[infos[i].Str("state")] {
			return &infos[i], nil
		}
	}
	if len(infos) > 0 {
		return &infos[0], nil
	}
	return nil, fmt.Errorf("app %s has no appInfo", appID)
}

// findByAttr returns the first resource whose string attribute equals value.
func findByAttr(list []api.Resource, key, value string) *api.Resource {
	for i := range list {
		if list[i].Str(key) == value {
			return &list[i]
		}
	}
	return nil
}

// valueOrFile returns s, or the file contents when s starts with "@".
func valueOrFile(s string) (string, error) {
	if strings.HasPrefix(s, "@") {
		b, err := os.ReadFile(s[1:])
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(b), "\n"), nil
	}
	return s, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func editableStateNames() []string {
	out := make([]string, 0, len(editableStates))
	for s := range editableStates {
		out = append(out, s)
	}
	return out
}
