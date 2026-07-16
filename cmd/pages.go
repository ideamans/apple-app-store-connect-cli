package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var cppCmd = &cobra.Command{
	Use:   "product-pages",
	Short: "Work with custom product pages (create, localize, upload screenshots)",
	Long: `Manage custom product pages (appCustomProductPages): alternate versions of the
App Store product page reachable via their own URL, e.g. for ad campaigns.

A page owns versions (appCustomProductPageVersions); each version owns per-locale
localizations, which own screenshot sets. Page versions are submitted for review
via: asc review-submissions add-item ...`,
}

var cppExpCmd = &cobra.Command{
	Use:   "experiments",
	Short: "Work with product page optimization experiments (A/B tests)",
	Long: `Manage product page optimization experiments (appStoreVersionExperiments v2):
A/B tests of alternate icons, screenshots and previews against the default page.

An experiment owns treatments; each treatment owns per-locale localizations,
which own screenshot sets. Experiments are submitted for review via:
asc review-submissions add-item ...`,
}

var (
	cppApp           string
	cppID            string
	cppName          string
	cppFromVersion   string
	cppPageID        string
	cppVersionID     string
	cppDeepLink      string
	cppLocale        string
	cppPromoText     string
	cppLocID         string
	cppDisplay       string
	cppFiles         []string
	cppVersion       string
	cppExpID         string
	cppExpName       string
	cppExpTraffic    int
	cppExpPlatform   string
	cppTreatmentID   string
	cppTreatmentName string
	cppIconName      string
	cppTreatLocID    string
)

// cppBool formats a boolean attribute, or "" when absent.
func cppBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return fmt.Sprintf("%t", v)
	}
	return ""
}

var cppListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's custom product pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, cppApp)
		if err != nil {
			return err
		}
		pages, err := c.List(ctx, "/v1/apps/"+appID+"/appCustomProductPages?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVISIBLE\tID")
		for i := range pages {
			p := &pages[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Str("name"), cppBool(p, "visible"), p.ID)
		}
		return w.Flush()
	},
}

var cppCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom product page",
	Long: `Create a custom product page. With --from-version-id, the page's first version
is copied from an existing appStoreVersion (the appStoreVersionTemplate
relationship); otherwise create a version afterwards with create-version.`,
	Example: `  asc product-pages create --app 6790641087 --name "Campaign A"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, cppApp)
		if err != nil {
			return err
		}
		rels := map[string]json.RawMessage{"app": api.Rel("apps", appID)}
		if cppFromVersion != "" {
			rels["appStoreVersionTemplate"] = api.Rel("appStoreVersions", cppFromVersion)
		}
		created, err := c.Post(ctx, "/v1/appCustomProductPages", api.Body{
			Data: api.Resource{
				Type:          "appCustomProductPages",
				Attributes:    map[string]any{"name": cppName},
				Relationships: rels,
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created custom product page %q (%s).\n", cppName, created.ID)
		return nil
	},
}

var cppShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a custom product page and its versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		page, included, err := c.Get(cmd.Context(), "/v1/appCustomProductPages/"+cppID+"?include=appCustomProductPageVersions")
		if err != nil {
			return err
		}
		out := struct {
			Page     *api.Resource  `json:"page"`
			Versions []api.Resource `json:"versions"`
		}{page, included}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var cppDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a custom product page",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appCustomProductPages/"+cppID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var cppVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List a custom product page's versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		versions, err := c.List(cmd.Context(), "/v1/appCustomProductPages/"+cppPageID+"/appCustomProductPageVersions?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tSTATE\tDEEP_LINK\tID")
		for _, v := range versions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Str("version"), v.Str("state"), v.Str("deepLink"), v.ID)
		}
		return w.Flush()
	},
}

var cppCreateVersionCmd = &cobra.Command{
	Use:   "create-version",
	Short: "Create a new version of a custom product page",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cppDeepLink != "" {
			attrs["deepLink"] = cppDeepLink
		}
		created, err := c.Post(cmd.Context(), "/v1/appCustomProductPageVersions", api.Body{
			Data: api.Resource{
				Type:          "appCustomProductPageVersions",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"appCustomProductPage": api.Rel("appCustomProductPages", cppPageID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created page version %s.\n", created.ID)
		return nil
	},
}

var cppLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Create or update a page version's localization",
	Long: `Create the appCustomProductPageLocalization for a locale on a page version (or
update its promotional text). The localization id printed on success is what
upload-screenshot takes as --localization-id.`,
	Example: `  asc product-pages localize --version-id <id> --locale ja --promotional-text "夏のキャンペーン実施中"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		attrs := map[string]any{}
		if cmd.Flags().Changed("promotional-text") {
			v, err := valueOrFile(cppPromoText)
			if err != nil {
				return err
			}
			attrs["promotionalText"] = v
		}
		locs, err := c.List(ctx, "/v1/appCustomProductPageVersions/"+cppVersionID+"/appCustomProductPageLocalizations?limit=50")
		if err != nil {
			return err
		}
		locID := ""
		if existing := findByAttr(locs, "locale", cppLocale); existing != nil {
			locID = existing.ID
			if len(attrs) > 0 {
				_, err = c.Patch(ctx, "/v1/appCustomProductPageLocalizations/"+existing.ID, api.Body{
					Data: api.Resource{Type: "appCustomProductPageLocalizations", ID: existing.ID, Attributes: attrs},
				})
			}
		} else {
			attrs["locale"] = cppLocale
			var created *api.Resource
			created, err = c.Post(ctx, "/v1/appCustomProductPageLocalizations", api.Body{
				Data: api.Resource{
					Type:          "appCustomProductPageLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appCustomProductPageVersion": api.Rel("appCustomProductPageVersions", cppVersionID)},
				},
			})
			if created != nil {
				locID = created.ID
			}
		}
		if err != nil {
			return err
		}
		fmt.Printf("Page localization (%s): %s\n", cppLocale, locID)
		return nil
	},
}

// cppScreenshotSetID finds or creates the screenshot set for a custom product
// page localization + display type. Distinct from screenshotSetID in assets.go,
// which targets appStoreVersionLocalizations.
func cppScreenshotSetID(ctx context.Context, c *api.Client, locID, display string) (string, error) {
	sets, err := c.List(ctx, "/v1/appCustomProductPageLocalizations/"+locID+"/appScreenshotSets?limit=50")
	if err != nil {
		return "", err
	}
	if set := findByAttr(sets, "screenshotDisplayType", display); set != nil {
		return set.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appScreenshotSets", api.Body{
		Data: api.Resource{
			Type:          "appScreenshotSets",
			Attributes:    map[string]any{"screenshotDisplayType": display},
			Relationships: map[string]json.RawMessage{"appCustomProductPageLocalization": api.Rel("appCustomProductPageLocalizations", locID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

var cppUploadScreenshotCmd = &cobra.Command{
	Use:   "upload-screenshot",
	Short: "Upload screenshots to a page localization",
	Example: `  asc product-pages upload-screenshot --localization-id <id> --display APP_IPHONE_67 \
    --file app-store/cpp/01.png --file app-store/cpp/02.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		setID, err := cppScreenshotSetID(ctx, c, cppLocID, cppDisplay)
		if err != nil {
			return err
		}
		for _, f := range cppFiles {
			id, err := uploadAsset(ctx, c, assetSpec{
				reserveType: "appScreenshots",
				relName:     "appScreenshotSet",
				relType:     "appScreenshotSets",
				relID:       setID,
				filePath:    f,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			fmt.Printf("uploaded %s -> %s\n", filepath.Base(f), id)
		}
		return nil
	},
}

var cppExpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List product page experiments (v2)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, cppApp)
		if err != nil {
			return err
		}
		path := "/v1/apps/" + appID + "/appStoreVersionExperimentsV2?limit=50"
		if cppVersion != "" {
			ver, err := editableVersion(ctx, c, appID, cppVersion)
			if err != nil {
				return err
			}
			path = "/v1/appStoreVersions/" + ver.ID + "/appStoreVersionExperimentsV2?limit=50"
		}
		exps, err := c.List(ctx, path)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPLATFORM\tSTATE\tTRAFFIC\tSTART\tID")
		for i := range exps {
			e := &exps[i]
			traffic := ""
			if v, ok := e.Attributes["trafficProportion"].(float64); ok {
				traffic = fmt.Sprintf("%.0f%%", v)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				e.Str("name"), e.Str("platform"), e.Str("state"), traffic, e.Str("startDate"), e.ID)
		}
		return w.Flush()
	},
}

var cppExpCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a product page experiment (v2)",
	Example: `  asc experiments create --app 6790641087 --name "Icon test" --traffic-proportion 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, cppApp)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v2/appStoreVersionExperiments", api.Body{
			Data: api.Resource{
				Type: "appStoreVersionExperiments",
				Attributes: map[string]any{
					"name":              cppExpName,
					"platform":          cppExpPlatform,
					"trafficProportion": cppExpTraffic,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created experiment %q (%s).\n", cppExpName, created.ID)
		return nil
	},
}

var cppExpStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an approved experiment (started=true)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v2/appStoreVersionExperiments/"+cppExpID, api.Body{
			Data: api.Resource{Type: "appStoreVersionExperiments", ID: cppExpID, Attributes: map[string]any{"started": true}},
		})
		if err != nil {
			return err
		}
		fmt.Println("Experiment started.")
		return nil
	},
}

var cppExpStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running experiment (started=false)",
	Long: `Stop a running experiment. The update request exposes a single boolean
attribute "started"; stopping sets it back to false.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v2/appStoreVersionExperiments/"+cppExpID, api.Body{
			Data: api.Resource{Type: "appStoreVersionExperiments", ID: cppExpID, Attributes: map[string]any{"started": false}},
		})
		if err != nil {
			return err
		}
		fmt.Println("Experiment stopped.")
		return nil
	},
}

var cppExpDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v2/appStoreVersionExperiments/"+cppExpID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var cppExpTreatmentsCmd = &cobra.Command{
	Use:   "treatments",
	Short: "List an experiment's treatments",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		treatments, err := c.List(cmd.Context(), "/v2/appStoreVersionExperiments/"+cppExpID+"/appStoreVersionExperimentTreatments?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tAPP_ICON_NAME\tPROMOTED\tID")
		for _, t := range treatments {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Str("name"), t.Str("appIconName"), t.Str("promotedDate"), t.ID)
		}
		return w.Flush()
	},
}

var cppExpCreateTreatmentCmd = &cobra.Command{
	Use:   "create-treatment",
	Short: "Add a treatment (variant) to an experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{"name": cppTreatmentName}
		if cppIconName != "" {
			attrs["appIconName"] = cppIconName
		}
		created, err := c.Post(cmd.Context(), "/v1/appStoreVersionExperimentTreatments", api.Body{
			Data: api.Resource{
				Type:          "appStoreVersionExperimentTreatments",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"appStoreVersionExperimentV2": api.Rel("appStoreVersionExperiments", cppExpID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created treatment %q (%s).\n", cppTreatmentName, created.ID)
		return nil
	},
}

var cppExpLocalizeTreatmentCmd = &cobra.Command{
	Use:   "localize-treatment",
	Short: "Create a treatment localization for a locale",
	Long: `Create the appStoreVersionExperimentTreatmentLocalization for a locale (or
print the existing one). The printed id is what upload-screenshot takes as
--treatment-localization-id.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		locs, err := c.List(ctx, "/v1/appStoreVersionExperimentTreatments/"+cppTreatmentID+"/appStoreVersionExperimentTreatmentLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", cppLocale); existing != nil {
			fmt.Printf("Treatment localization (%s): %s\n", cppLocale, existing.ID)
			return nil
		}
		created, err := c.Post(ctx, "/v1/appStoreVersionExperimentTreatmentLocalizations", api.Body{
			Data: api.Resource{
				Type:          "appStoreVersionExperimentTreatmentLocalizations",
				Attributes:    map[string]any{"locale": cppLocale},
				Relationships: map[string]json.RawMessage{"appStoreVersionExperimentTreatment": api.Rel("appStoreVersionExperimentTreatments", cppTreatmentID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Treatment localization (%s): %s\n", cppLocale, created.ID)
		return nil
	},
}

// cppTreatmentScreenshotSetID finds or creates the screenshot set for a
// treatment localization + display type.
func cppTreatmentScreenshotSetID(ctx context.Context, c *api.Client, locID, display string) (string, error) {
	sets, err := c.List(ctx, "/v1/appStoreVersionExperimentTreatmentLocalizations/"+locID+"/appScreenshotSets?limit=50")
	if err != nil {
		return "", err
	}
	if set := findByAttr(sets, "screenshotDisplayType", display); set != nil {
		return set.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appScreenshotSets", api.Body{
		Data: api.Resource{
			Type:          "appScreenshotSets",
			Attributes:    map[string]any{"screenshotDisplayType": display},
			Relationships: map[string]json.RawMessage{"appStoreVersionExperimentTreatmentLocalization": api.Rel("appStoreVersionExperimentTreatmentLocalizations", locID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

var cppExpUploadScreenshotCmd = &cobra.Command{
	Use:   "upload-screenshot",
	Short: "Upload screenshots to a treatment localization",
	Example: `  asc experiments upload-screenshot --treatment-localization-id <id> \
    --display APP_IPHONE_67 --file variants/01.png --file variants/02.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		setID, err := cppTreatmentScreenshotSetID(ctx, c, cppTreatLocID, cppDisplay)
		if err != nil {
			return err
		}
		for _, f := range cppFiles {
			id, err := uploadAsset(ctx, c, assetSpec{
				reserveType: "appScreenshots",
				relName:     "appScreenshotSet",
				relType:     "appScreenshotSets",
				relID:       setID,
				filePath:    f,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			fmt.Printf("uploaded %s -> %s\n", filepath.Base(f), id)
		}
		return nil
	},
}

var cppExpPromoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote a winning treatment to the app's product page",
	Long: `Apply a winning experiment treatment to an App Store version by creating an
appStoreVersionPromotion. The API relates a promotion to an appStoreVersion and
an appStoreVersionExperimentTreatment (custom product page versions cannot be
promoted through this endpoint).`,
	Example: `  asc experiments promote --app 6790641087 --treatment-id <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, cppApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, cppVersion)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/appStoreVersionPromotions", api.Body{
			Data: api.Resource{
				Type: "appStoreVersionPromotions",
				Relationships: map[string]json.RawMessage{
					"appStoreVersion":                    api.Rel("appStoreVersions", ver.ID),
					"appStoreVersionExperimentTreatment": api.Rel("appStoreVersionExperimentTreatments", cppTreatmentID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Promoted treatment %s to version %s (%s).\n", cppTreatmentID, ver.Str("versionString"), created.ID)
		return nil
	},
}

func init() {
	// product-pages
	for _, sub := range []*cobra.Command{cppListCmd, cppCreateCmd} {
		sub.Flags().StringVar(&cppApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	cppCreateCmd.Flags().StringVar(&cppName, "name", "", "page name (required)")
	cppCreateCmd.Flags().StringVar(&cppFromVersion, "from-version-id", "", "appStoreVersion id to copy metadata from (appStoreVersionTemplate)")
	_ = cppCreateCmd.MarkFlagRequired("name")

	for _, sub := range []*cobra.Command{cppShowCmd, cppDeleteCmd} {
		sub.Flags().StringVar(&cppID, "id", "", "appCustomProductPage id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	for _, sub := range []*cobra.Command{cppVersionsCmd, cppCreateVersionCmd} {
		sub.Flags().StringVar(&cppPageID, "page-id", "", "appCustomProductPage id (required)")
		_ = sub.MarkFlagRequired("page-id")
	}
	cppCreateVersionCmd.Flags().StringVar(&cppDeepLink, "deep-link", "", "deep link URL for this page version")

	cppLocalizeCmd.Flags().StringVar(&cppVersionID, "version-id", "", "appCustomProductPageVersion id (required)")
	cppLocalizeCmd.Flags().StringVar(&cppLocale, "locale", "ja", "locale, e.g. ja / en-US")
	cppLocalizeCmd.Flags().StringVar(&cppPromoText, "promotional-text", "", "promotional text (@file allowed)")
	_ = cppLocalizeCmd.MarkFlagRequired("version-id")

	cppUploadScreenshotCmd.Flags().StringVar(&cppLocID, "localization-id", "", "appCustomProductPageLocalization id (required)")
	cppUploadScreenshotCmd.Flags().StringVar(&cppDisplay, "display", "", "screenshot display type, e.g. APP_IPHONE_67 (required)")
	cppUploadScreenshotCmd.Flags().StringArrayVar(&cppFiles, "file", nil, "screenshot file (repeatable, applied in order)")
	_ = cppUploadScreenshotCmd.MarkFlagRequired("localization-id")
	_ = cppUploadScreenshotCmd.MarkFlagRequired("display")
	_ = cppUploadScreenshotCmd.MarkFlagRequired("file")

	cppCmd.AddCommand(cppListCmd, cppCreateCmd, cppShowCmd, cppDeleteCmd,
		cppVersionsCmd, cppCreateVersionCmd, cppLocalizeCmd, cppUploadScreenshotCmd)

	// experiments
	for _, sub := range []*cobra.Command{cppExpListCmd, cppExpCreateCmd, cppExpPromoteCmd} {
		sub.Flags().StringVar(&cppApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{cppExpListCmd, cppExpPromoteCmd} {
		sub.Flags().StringVar(&cppVersion, "version", "", "version string (default: list at app level / the editable version)")
	}
	cppExpCreateCmd.Flags().StringVar(&cppExpName, "name", "", "experiment name (required)")
	cppExpCreateCmd.Flags().IntVar(&cppExpTraffic, "traffic-proportion", 0, "percentage of traffic in the experiment, 1-100 (required)")
	cppExpCreateCmd.Flags().StringVar(&cppExpPlatform, "platform", "IOS", "platform: IOS, MAC_OS, TV_OS, VISION_OS")
	_ = cppExpCreateCmd.MarkFlagRequired("name")
	_ = cppExpCreateCmd.MarkFlagRequired("traffic-proportion")

	for _, sub := range []*cobra.Command{cppExpStartCmd, cppExpStopCmd, cppExpDeleteCmd} {
		sub.Flags().StringVar(&cppExpID, "id", "", "appStoreVersionExperiment (v2) id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	for _, sub := range []*cobra.Command{cppExpTreatmentsCmd, cppExpCreateTreatmentCmd} {
		sub.Flags().StringVar(&cppExpID, "experiment-id", "", "appStoreVersionExperiment (v2) id (required)")
		_ = sub.MarkFlagRequired("experiment-id")
	}
	cppExpCreateTreatmentCmd.Flags().StringVar(&cppTreatmentName, "name", "", "treatment name (required)")
	cppExpCreateTreatmentCmd.Flags().StringVar(&cppIconName, "app-icon-name", "", "alternate app icon asset name shipped in the build")
	_ = cppExpCreateTreatmentCmd.MarkFlagRequired("name")

	cppExpLocalizeTreatmentCmd.Flags().StringVar(&cppTreatmentID, "treatment-id", "", "appStoreVersionExperimentTreatment id (required)")
	cppExpLocalizeTreatmentCmd.Flags().StringVar(&cppLocale, "locale", "ja", "locale, e.g. ja / en-US")
	_ = cppExpLocalizeTreatmentCmd.MarkFlagRequired("treatment-id")

	cppExpUploadScreenshotCmd.Flags().StringVar(&cppTreatLocID, "treatment-localization-id", "", "appStoreVersionExperimentTreatmentLocalization id (required)")
	cppExpUploadScreenshotCmd.Flags().StringVar(&cppDisplay, "display", "", "screenshot display type, e.g. APP_IPHONE_67 (required)")
	cppExpUploadScreenshotCmd.Flags().StringArrayVar(&cppFiles, "file", nil, "screenshot file (repeatable, applied in order)")
	_ = cppExpUploadScreenshotCmd.MarkFlagRequired("treatment-localization-id")
	_ = cppExpUploadScreenshotCmd.MarkFlagRequired("display")
	_ = cppExpUploadScreenshotCmd.MarkFlagRequired("file")

	cppExpPromoteCmd.Flags().StringVar(&cppTreatmentID, "treatment-id", "", "appStoreVersionExperimentTreatment id to promote (required)")
	_ = cppExpPromoteCmd.MarkFlagRequired("treatment-id")

	cppExpCmd.AddCommand(cppExpListCmd, cppExpCreateCmd, cppExpStartCmd, cppExpStopCmd, cppExpDeleteCmd,
		cppExpTreatmentsCmd, cppExpCreateTreatmentCmd, cppExpLocalizeTreatmentCmd, cppExpUploadScreenshotCmd, cppExpPromoteCmd)

	rootCmd.AddCommand(cppCmd, cppExpCmd)
}
