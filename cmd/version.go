package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Work with App Store versions (create, localize, select build)",
}

var (
	verApp       string
	verString    string
	verPlatform  string
	verLocale    string
	verDesc      string
	verKeywords  string
	verPromo     string
	verWhatsNew  string
	verSupport   string
	verMarketing string
	verBuild     string
)

var versionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's App Store versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, verApp)
		if err != nil {
			return err
		}
		versions, err := c.List(ctx, "/v1/apps/"+appID+"/appStoreVersions?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tPLATFORM\tSTATE\tID")
		for _, v := range versions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Str("versionString"), v.Str("platform"), v.Str("appStoreState"), v.ID)
		}
		return w.Flush()
	},
}

var versionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new App Store version",
	Example: `  asc version create --app 6790641087 --version 1.0
  asc version create --app 6790641087 --version 1.1 --platform IOS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, verApp)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/appStoreVersions", api.Body{
			Data: api.Resource{
				Type: "appStoreVersions",
				Attributes: map[string]any{
					"platform":      verPlatform,
					"versionString": verString,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created version %s (%s).\n", verString, created.ID)
		return nil
	},
}

var versionLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set localized version metadata (description, keywords, what's new, URLs)",
	Long: `Create or update the appStoreVersionLocalization for a locale on the app's
editable version. Any flag may take @file to read the value from a file.

On an app's FIRST version, whatsNew (release notes) is not editable — the API
409s. This command skips whatsNew with a warning in that case and still applies
the other attributes.`,
	Example: `  asc version localize --app 6790641087 --locale ja \
    --description @app-store/description.txt \
    --keywords "領収書,レシート,経費,Excel,スキャン" \
    --promo "撮るだけでExcelに。" \
    --support-url https://japan-receipt-scan.web.app/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, verApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, verString)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		fields := []struct {
			flag, attr string
			val        *string
		}{
			{"description", "description", &verDesc},
			{"keywords", "keywords", &verKeywords},
			{"promo", "promotionalText", &verPromo},
			{"whats-new", "whatsNew", &verWhatsNew},
			{"support-url", "supportUrl", &verSupport},
			{"marketing-url", "marketingUrl", &verMarketing},
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
			return fmt.Errorf("nothing to set: pass --description/--keywords/--promo/--whats-new/--support-url/--marketing-url")
		}
		locs, err := c.List(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreVersionLocalizations?limit=50")
		if err != nil {
			return err
		}
		apply := func(attrs map[string]any) error {
			if existing := findByAttr(locs, "locale", verLocale); existing != nil {
				_, err := c.Patch(ctx, "/v1/appStoreVersionLocalizations/"+existing.ID, api.Body{
					Data: api.Resource{Type: "appStoreVersionLocalizations", ID: existing.ID, Attributes: attrs},
				})
				return err
			}
			attrs["locale"] = verLocale
			_, err := c.Post(ctx, "/v1/appStoreVersionLocalizations", api.Body{
				Data: api.Resource{
					Type:          "appStoreVersionLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", ver.ID)},
				},
			})
			return err
		}
		err = apply(attrs)
		// An app's first version has no previous release, so whatsNew is not
		// editable and the API 409s. Retry without it rather than failing the
		// other attributes.
		if _, hadWhatsNew := attrs["whatsNew"]; err != nil && hadWhatsNew && strings.Contains(err.Error(), "whatsNew") {
			fmt.Fprintln(os.Stderr, "warning: whatsNew is not editable on this version (first versions have no release notes); skipped it")
			delete(attrs, "whatsNew")
			if len(attrs) == 0 {
				return nil
			}
			err = apply(attrs)
		}
		if err != nil {
			return err
		}
		fmt.Printf("Version %s localization (%s) updated.\n", ver.Str("versionString"), verLocale)
		return nil
	},
}

var versionSetBuildCmd = &cobra.Command{
	Use:   "set-build",
	Short: "Attach an uploaded build to the editable version",
	Long: `Select which uploaded build the version ships. The build must already be
uploaded (via Xcode or Transporter) and finished processing. --build accepts a
build id or a build version string (the CFBundleVersion, e.g. "42").`,
	Example: `  asc version set-build --app 6790641087 --build 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, verApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, verString)
		if err != nil {
			return err
		}
		buildID := verBuild
		if !isDigits(verBuild) || len(verBuild) < 6 {
			// Treat as a build version string; look it up for this app.
			builds, err := c.List(ctx, "/v1/builds?filter[app]="+appID+"&filter[version]="+verBuild+"&limit=1")
			if err != nil {
				return err
			}
			if len(builds) == 0 {
				return fmt.Errorf("no build with version %q for app %s (is it uploaded and processed?)", verBuild, appID)
			}
			buildID = builds[0].ID
		}
		_, err = c.Patch(ctx, "/v1/appStoreVersions/"+ver.ID, api.Body{
			Data: api.Resource{
				Type:          "appStoreVersions",
				ID:            ver.ID,
				Relationships: map[string]json.RawMessage{"build": api.Rel("builds", buildID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Build %s attached to version %s.\n", buildID, ver.Str("versionString"))
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{versionListCmd, versionCreateCmd, versionLocalizeCmd, versionSetBuildCmd} {
		sub.Flags().StringVar(&verApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{versionLocalizeCmd, versionSetBuildCmd} {
		sub.Flags().StringVar(&verString, "version", "", "version string (default: the editable version)")
	}
	versionCreateCmd.Flags().StringVar(&verString, "version", "", "version string, e.g. 1.0 (required)")
	versionCreateCmd.Flags().StringVar(&verPlatform, "platform", "IOS", "platform: IOS, MAC_OS, TV_OS, VISION_OS")
	_ = versionCreateCmd.MarkFlagRequired("version")

	versionLocalizeCmd.Flags().StringVar(&verLocale, "locale", "ja", "locale, e.g. ja / en-US")
	versionLocalizeCmd.Flags().StringVar(&verDesc, "description", "", "description (@file allowed)")
	versionLocalizeCmd.Flags().StringVar(&verKeywords, "keywords", "", "comma-separated keywords, max 100 chars (@file allowed)")
	versionLocalizeCmd.Flags().StringVar(&verPromo, "promo", "", "promotional text (@file allowed)")
	versionLocalizeCmd.Flags().StringVar(&verWhatsNew, "whats-new", "", "what's new in this version (@file allowed)")
	versionLocalizeCmd.Flags().StringVar(&verSupport, "support-url", "", "support URL")
	versionLocalizeCmd.Flags().StringVar(&verMarketing, "marketing-url", "", "marketing URL")

	versionSetBuildCmd.Flags().StringVar(&verBuild, "build", "", "build id or build version string (required)")
	_ = versionSetBuildCmd.MarkFlagRequired("build")

	versionCmd.AddCommand(versionListCmd, versionCreateCmd, versionLocalizeCmd, versionSetBuildCmd)
	rootCmd.AddCommand(versionCmd)
}
