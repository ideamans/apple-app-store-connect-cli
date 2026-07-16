package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	vxApp          string
	vxVersion      string
	vxPlatform     string
	vxCopyright    string
	vxReleaseType  string
	vxEarliestDate string
	vxNewVersion   string
	vxPhasedState  string
)

// vxFindVersion looks up a version in any state, restricted to --platform when
// set (multi-platform apps list every platform's versions together). When
// versionString is given it must match. Otherwise the first version whose
// appStoreState is one of states (in priority order) is returned, or with no
// states, the newest version.
func vxFindVersion(ctx context.Context, c *api.Client, appID, versionString string, states ...string) (*api.Resource, error) {
	path := "/v1/apps/" + appID + "/appStoreVersions?limit=50"
	if vxPlatform != "" {
		path += "&filter[platform]=" + strings.ToUpper(vxPlatform)
	}
	versions, err := c.List(ctx, path)
	if err != nil {
		return nil, err
	}
	if versionString != "" {
		if v := findByAttr(versions, "versionString", versionString); v != nil {
			return v, nil
		}
		return nil, fmt.Errorf("app %s has no version %q", appID, versionString)
	}
	for _, state := range states {
		for i := range versions {
			if versions[i].Str("appStoreState") == state {
				return &versions[i], nil
			}
		}
	}
	if len(states) > 0 {
		return nil, fmt.Errorf("app %s has no version in state %s (pass --version to target one explicitly)", appID, strings.Join(states, "/"))
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("app %s has no App Store versions", appID)
	}
	return &versions[0], nil
}

var vxShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a version's state, release settings and attached build",
	Example: `  asc version show --app 6790641087
  asc version show --app 6790641087 --version 1.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, err := vxFindVersion(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		full, included, err := c.Get(ctx, "/v1/appStoreVersions/"+ver.ID+"?include=build")
		if err != nil {
			return err
		}
		out := map[string]any{"id": full.ID}
		for _, k := range []string{"versionString", "platform", "appStoreState", "appVersionState", "releaseType", "earliestReleaseDate", "copyright", "createdDate", "downloadable"} {
			if v, ok := full.Attributes[k]; ok {
				out[k] = v
			}
		}
		for i := range included {
			if included[i].Type == "builds" {
				out["build"] = map[string]any{
					"id":              included[i].ID,
					"version":         included[i].Str("version"),
					"processingState": included[i].Str("processingState"),
				}
			}
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var vxUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the editable version's copyright, release type or version string",
	Example: `  asc version update --app 6790641087 --copyright "2026 ideaman's Inc."
  asc version update --app 6790641087 --release-type SCHEDULED --earliest-release-date 2026-08-01T00:00:00Z
  asc version update --app 6790641087 --version 1.0 --version-string 1.0.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		fields := []struct {
			flag, attr string
			val        *string
		}{
			{"copyright", "copyright", &vxCopyright},
			{"release-type", "releaseType", &vxReleaseType},
			{"earliest-release-date", "earliestReleaseDate", &vxEarliestDate},
			{"version-string", "versionString", &vxNewVersion},
		}
		for _, f := range fields {
			if cmd.Flags().Changed(f.flag) {
				v, err := valueOrFile(*f.val)
				if err != nil {
					return err
				}
				attrs[f.attr] = v
			}
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --copyright/--release-type/--earliest-release-date/--version-string")
		}
		_, err = c.Patch(ctx, "/v1/appStoreVersions/"+ver.ID, api.Body{
			Data: api.Resource{Type: "appStoreVersions", ID: ver.ID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Version %s updated.\n", ver.Str("versionString"))
		return nil
	},
}

var vxDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an App Store version (requires an explicit --version)",
	Example: `  asc version delete --app 6790641087 --version 1.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if vxVersion == "" {
			return fmt.Errorf("--version is required for delete (refusing to guess which version to remove)")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, err := vxFindVersion(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		if err := c.Delete(ctx, "/v1/appStoreVersions/"+ver.ID); err != nil {
			return err
		}
		fmt.Printf("Deleted version %s (%s).\n", ver.Str("versionString"), ver.ID)
		return nil
	},
}

var vxReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release an approved version that is pending developer release",
	Long: `Request the release of a version in PENDING_DEVELOPER_RELEASE state (i.e. the
version was approved with releaseType MANUAL). Without --version, the version
currently pending developer release is used.`,
	Example: `  asc version release --app 6790641087
  asc version release --app 6790641087 --version 1.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, err := vxFindVersion(ctx, c, appID, vxVersion, "PENDING_DEVELOPER_RELEASE")
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/appStoreVersionReleaseRequests", api.Body{
			Data: api.Resource{
				Type:          "appStoreVersionReleaseRequests",
				Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", ver.ID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Release requested for version %s.\n", ver.Str("versionString"))
		return nil
	},
}

var vxPhasedCmd = &cobra.Command{
	Use:   "phased-release",
	Short: "Manage the version's phased release (gradual rollout over 7 days)",
}

// vxPhasedFind resolves the target version and its phased release (nil if none).
func vxPhasedFind(ctx context.Context, c *api.Client, appID, versionString string) (*api.Resource, *api.Resource, error) {
	ver, err := vxFindVersion(ctx, c, appID, versionString)
	if err != nil {
		return nil, nil, err
	}
	pr, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreVersionPhasedRelease")
	if err != nil {
		return nil, nil, err
	}
	return ver, pr, nil
}

// vxPhasedSetState patches the phased release of the target version to state.
func vxPhasedSetState(cmd *cobra.Command, state string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	appID, err := resolveAppID(ctx, c, vxApp)
	if err != nil {
		return err
	}
	ver, pr, err := vxPhasedFind(ctx, c, appID, vxVersion)
	if err != nil {
		return err
	}
	if pr == nil || pr.ID == "" {
		return fmt.Errorf("version %s has no phased release", ver.Str("versionString"))
	}
	_, err = c.Patch(ctx, "/v1/appStoreVersionPhasedReleases/"+pr.ID, api.Body{
		Data: api.Resource{
			Type:       "appStoreVersionPhasedReleases",
			ID:         pr.ID,
			Attributes: map[string]any{"phasedReleaseState": state},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("Phased release for version %s set to %s.\n", ver.Str("versionString"), state)
	return nil
}

var vxPhasedShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show the version's phased release status",
	Example: `  asc version phased-release show --app 6790641087`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, pr, err := vxPhasedFind(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		if pr == nil || pr.ID == "" {
			fmt.Printf("Version %s has no phased release.\n", ver.Str("versionString"))
			return nil
		}
		out := map[string]any{"id": pr.ID}
		for k, v := range pr.Attributes {
			out[k] = v
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var vxPhasedCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Enable phased release for the version",
	Example: `  asc version phased-release create --app 6790641087
  asc version phased-release create --app 6790641087 --state ACTIVE`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, err := vxFindVersion(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		data := api.Resource{
			Type:          "appStoreVersionPhasedReleases",
			Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", ver.ID)},
		}
		if cmd.Flags().Changed("state") {
			data.Attributes = map[string]any{"phasedReleaseState": vxPhasedState}
		}
		created, err := c.Post(ctx, "/v1/appStoreVersionPhasedReleases", api.Body{Data: data})
		if err != nil {
			return err
		}
		fmt.Printf("Phased release created for version %s (%s).\n", ver.Str("versionString"), created.ID)
		return nil
	},
}

var vxPhasedPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the phased release",
	RunE: func(cmd *cobra.Command, args []string) error {
		return vxPhasedSetState(cmd, "PAUSED")
	},
}

var vxPhasedResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused phased release",
	RunE: func(cmd *cobra.Command, args []string) error {
		return vxPhasedSetState(cmd, "ACTIVE")
	},
}

var vxPhasedCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Finish the rollout and release to all users immediately",
	RunE: func(cmd *cobra.Command, args []string) error {
		return vxPhasedSetState(cmd, "COMPLETE")
	},
}

var vxPhasedDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove the phased release (version releases to everyone at once)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, vxApp)
		if err != nil {
			return err
		}
		ver, pr, err := vxPhasedFind(ctx, c, appID, vxVersion)
		if err != nil {
			return err
		}
		if pr == nil || pr.ID == "" {
			return fmt.Errorf("version %s has no phased release", ver.Str("versionString"))
		}
		if err := c.Delete(ctx, "/v1/appStoreVersionPhasedReleases/"+pr.ID); err != nil {
			return err
		}
		fmt.Printf("Phased release deleted for version %s.\n", ver.Str("versionString"))
		return nil
	},
}

func init() {
	subs := []*cobra.Command{
		vxShowCmd, vxUpdateCmd, vxDeleteCmd, vxReleaseCmd,
		vxPhasedShowCmd, vxPhasedCreateCmd, vxPhasedPauseCmd, vxPhasedResumeCmd, vxPhasedCompleteCmd, vxPhasedDeleteCmd,
	}
	for _, sub := range subs {
		sub.Flags().StringVar(&vxApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
		sub.Flags().StringVar(&vxVersion, "version", "", "version string (default: picked by command)")
		sub.Flags().StringVar(&vxPlatform, "platform", "", "restrict to a platform: IOS, MAC_OS, TV_OS, VISION_OS")
	}
	_ = vxDeleteCmd.MarkFlagRequired("version")

	vxUpdateCmd.Flags().StringVar(&vxCopyright, "copyright", "", "copyright line, e.g. \"2026 ideaman's Inc.\" (@file allowed)")
	vxUpdateCmd.Flags().StringVar(&vxReleaseType, "release-type", "", "release type: MANUAL, AFTER_APPROVAL, SCHEDULED")
	vxUpdateCmd.Flags().StringVar(&vxEarliestDate, "earliest-release-date", "", "earliest release date (ISO8601, requires --release-type SCHEDULED)")
	vxUpdateCmd.Flags().StringVar(&vxNewVersion, "version-string", "", "rename the version string")

	vxPhasedCreateCmd.Flags().StringVar(&vxPhasedState, "state", "", "initial state: INACTIVE or ACTIVE (default: server default)")

	vxPhasedCmd.AddCommand(vxPhasedShowCmd, vxPhasedCreateCmd, vxPhasedPauseCmd, vxPhasedResumeCmd, vxPhasedCompleteCmd, vxPhasedDeleteCmd)
	versionCmd.AddCommand(vxShowCmd, vxUpdateCmd, vxDeleteCmd, vxReleaseCmd, vxPhasedCmd)
}
