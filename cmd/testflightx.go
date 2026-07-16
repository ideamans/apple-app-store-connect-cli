package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// Extra TestFlight commands: build beta details, recruitment criteria,
// App Clip invocations, build bundles, and tester re-invitations.
// All subcommands attach to the existing `beta` (tfCmd) tree.

var (
	tfxApp           string
	tfxBuild         string
	tfxGroupID       string
	tfxAutoNotify    string
	tfxFilters       []string
	tfxBuildBundleID string
	tfxURL           string
	tfxLocale        string
	tfxTitle         string
	tfxID            string
	tfxEmail         string
)

// tfxRelIDs returns the ids of a to-many relationship on a fetched resource
// (present when the relationship was requested via include).
func tfxRelIDs(r *api.Resource, name string) []string {
	raw, ok := r.Relationships[name]
	if !ok {
		return nil
	}
	var rel struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &rel); err != nil {
		return nil
	}
	ids := make([]string, 0, len(rel.Data))
	for _, d := range rel.Data {
		ids = append(ids, d.ID)
	}
	return ids
}

// tfxParseFilter parses a device-family OS filter given as
// FAMILY[:MIN_OS[:MAX_OS]] into a deviceFamilyOsVersionFilters element.
func tfxParseFilter(s string) (map[string]any, error) {
	parts := strings.SplitN(s, ":", 3)
	family := strings.ToUpper(strings.TrimSpace(parts[0]))
	if family == "" {
		return nil, fmt.Errorf("invalid --filter %q: expected FAMILY[:MIN_OS[:MAX_OS]] (e.g. IPHONE:17.0:18.0)", s)
	}
	f := map[string]any{"deviceFamily": family}
	if len(parts) >= 2 && parts[1] != "" {
		f["minimumOsInclusive"] = parts[1]
	}
	if len(parts) == 3 && parts[2] != "" {
		f["maximumOsInclusive"] = parts[2]
	}
	return f, nil
}

// --- beta build-detail -----------------------------------------------------------

var tfxBuildDetailCmd = &cobra.Command{
	Use:   "build-detail",
	Short: "Show or set a build's TestFlight beta details (auto-notify, beta states)",
}

var tfxBuildDetailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a build's beta detail (autoNotifyEnabled, internal/external beta state)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfxApp, tfxBuild)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/builds/"+buildID+"/buildBetaDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Printf("Build %s has no beta detail.\n", buildID)
			return nil
		}
		b, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var tfxBuildDetailSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set whether testers are automatically notified about the build",
	Example: `  asc beta build-detail set --app 6790641087 --build 42 --auto-notify true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		val, err := strconv.ParseBool(tfxAutoNotify)
		if err != nil {
			return fmt.Errorf("--auto-notify must be true or false")
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfxApp, tfxBuild)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/builds/"+buildID+"/buildBetaDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			return fmt.Errorf("build %s has no beta detail", buildID)
		}
		_, err = c.Patch(ctx, "/v1/buildBetaDetails/"+detail.ID, api.Body{
			Data: api.Resource{Type: "buildBetaDetails", ID: detail.ID, Attributes: map[string]any{"autoNotifyEnabled": val}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Build %s autoNotifyEnabled set to %v.\n", buildID, val)
		return nil
	},
}

// --- beta recruitment --------------------------------------------------------------

var tfxRecruitCmd = &cobra.Command{
	Use:   "recruitment",
	Short: "Manage a beta group's tester recruitment criteria (device/OS filters)",
	Long: `Recruitment criteria limit which devices and OS versions can join a beta
group through its public link. They belong to a beta group, so commands take
--group-id (find it with: asc beta groups list --app <app>).`,
}

var tfxRecruitShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the beta group's recruitment criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		crit, err := c.GetOptional(cmd.Context(), "/v1/betaGroups/"+tfxGroupID+"/betaRecruitmentCriteria")
		if err != nil {
			return err
		}
		if crit == nil || crit.ID == "" {
			fmt.Printf("Beta group %s has no recruitment criteria.\n", tfxGroupID)
			return nil
		}
		b, err := json.MarshalIndent(crit, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var tfxRecruitSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or replace the beta group's recruitment criteria",
	Long: `Each --filter is FAMILY[:MIN_OS[:MAX_OS]] where FAMILY is one of
IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, VISION. Min/max OS versions are
inclusive and optional. Use "asc beta recruitment options" to list the OS
versions Apple accepts.`,
	Example: `  asc beta recruitment set --group-id <group-id> --filter IPHONE:17.0 --filter IPAD:17.0:18.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		filters := make([]map[string]any, 0, len(tfxFilters))
		for _, f := range tfxFilters {
			parsed, err := tfxParseFilter(f)
			if err != nil {
				return err
			}
			filters = append(filters, parsed)
		}
		ctx := cmd.Context()
		existing, err := c.GetOptional(ctx, "/v1/betaGroups/"+tfxGroupID+"/betaRecruitmentCriteria")
		if err != nil {
			return err
		}
		attrs := map[string]any{"deviceFamilyOsVersionFilters": filters}
		if existing != nil && existing.ID != "" {
			_, err = c.Patch(ctx, "/v1/betaRecruitmentCriteria/"+existing.ID, api.Body{
				Data: api.Resource{Type: "betaRecruitmentCriteria", ID: existing.ID, Attributes: attrs},
			})
		} else {
			_, err = c.Post(ctx, "/v1/betaRecruitmentCriteria", api.Body{
				Data: api.Resource{
					Type:          "betaRecruitmentCriteria",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"betaGroup": api.Rel("betaGroups", tfxGroupID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Recruitment criteria for group %s set (%d filter(s)).\n", tfxGroupID, len(filters))
		return nil
	},
}

var tfxRecruitDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the beta group's recruitment criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		crit, err := c.GetOptional(ctx, "/v1/betaGroups/"+tfxGroupID+"/betaRecruitmentCriteria")
		if err != nil {
			return err
		}
		if crit == nil || crit.ID == "" {
			fmt.Printf("Beta group %s has no recruitment criteria.\n", tfxGroupID)
			return nil
		}
		if err := c.Delete(ctx, "/v1/betaRecruitmentCriteria/"+crit.ID); err != nil {
			return err
		}
		fmt.Printf("Deleted recruitment criteria of group %s.\n", tfxGroupID)
		return nil
	},
}

var tfxRecruitOptionsCmd = &cobra.Command{
	Use:   "options",
	Short: "List the device families and OS versions usable in recruitment criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		options, err := c.List(cmd.Context(), "/v1/betaRecruitmentCriterionOptions?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "DEVICE FAMILY\tOS VERSIONS")
		for i := range options {
			var groups []struct {
				DeviceFamily string   `json:"deviceFamily"`
				OsVersions   []string `json:"osVersions"`
			}
			if err := options[i].DecodeAttr("deviceFamilyOsVersions", &groups); err != nil {
				continue
			}
			for _, g := range groups {
				fmt.Fprintf(w, "%s\t%s\n", g.DeviceFamily, strings.Join(g.OsVersions, ", "))
			}
		}
		return w.Flush()
	},
}

// --- beta build-bundles ---------------------------------------------------------------

var tfxBundlesCmd = &cobra.Command{
	Use:   "build-bundles",
	Short: "List a build's bundles (bundle ids, SDK, symbols/dSYM availability)",
	Long: `Build bundles are only reachable through the build resource
(GET /v1/builds/{id}?include=buildBundles); their ids are needed for App Clip
invocation commands and expose dSYM/symbol availability.`,
	Example: `  asc beta build-bundles --app 6790641087 --build 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfxApp, tfxBuild)
		if err != nil {
			return err
		}
		_, included, err := c.Get(ctx, "/v1/builds/"+buildID+"?include=buildBundles")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "BUNDLE ID\tTYPE\tSDK BUILD\tSYMBOLS\tID")
		var dsyms []string
		count := 0
		for i := range included {
			b := &included[i]
			if b.Type != "buildBundles" {
				continue
			}
			count++
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				b.Str("bundleId"), b.Str("bundleType"), b.Str("sdkBuild"), tfBoolAttr(b, "includesSymbols"), b.ID)
			if u := b.Str("dSYMUrl"); u != "" {
				dsyms = append(dsyms, fmt.Sprintf("%s: %s", b.Str("bundleId"), u))
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}
		if count == 0 {
			fmt.Printf("Build %s has no build bundles.\n", buildID)
		}
		for _, d := range dsyms {
			fmt.Println("dSYM " + d)
		}
		return nil
	},
}

// --- beta invocations (App Clip) ---------------------------------------------------------

var tfxInvocationsCmd = &cobra.Command{
	Use:   "invocations",
	Short: "Manage TestFlight App Clip invocations for a build bundle",
	Long: `Beta App Clip invocations let testers launch an App Clip experience from
TestFlight. They belong to a build bundle; find bundle ids with
"asc beta build-bundles --build <id|version> --app <app>".`,
}

var tfxInvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a build bundle's App Clip invocations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		// Decode manually to keep the included localizations that c.List drops.
		var invocations []api.Resource
		locs := map[string]*api.Resource{}
		next := "/v1/buildBundles/" + tfxBuildBundleID + "/betaAppClipInvocations?include=betaAppClipInvocationLocalizations&limit=200"
		for next != "" {
			raw, err := c.Do(ctx, http.MethodGet, next, nil)
			if err != nil {
				return err
			}
			var doc struct {
				Data     []api.Resource `json:"data"`
				Included []api.Resource `json:"included"`
				Links    struct {
					Next string `json:"next"`
				} `json:"links"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				return err
			}
			for i := range doc.Included {
				if doc.Included[i].Type == "betaAppClipInvocationLocalizations" {
					locs[doc.Included[i].ID] = &doc.Included[i]
				}
			}
			invocations = append(invocations, doc.Data...)
			next = doc.Links.Next
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "URL\tLOCALIZATIONS\tID")
		for i := range invocations {
			inv := &invocations[i]
			var titles []string
			for _, id := range tfxRelIDs(inv, "betaAppClipInvocationLocalizations") {
				if l, ok := locs[id]; ok {
					titles = append(titles, fmt.Sprintf("%s=%q (%s)", l.Str("locale"), l.Str("title"), l.ID))
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", inv.Str("url"), strings.Join(titles, ", "), inv.ID)
		}
		return w.Flush()
	},
}

var tfxInvCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an App Clip invocation with one localization",
	Long: `Create a beta App Clip invocation. The API requires at least one
localization (locale + title), created inline with the invocation. Add more
locales afterwards with "asc beta invocations localize".`,
	Example: `  asc beta invocations create --build-bundle-id <bundle-id> \
    --url https://example.com/clip --locale ja --title "クリップを試す"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		const localID = "${loc1}"
		locLinkage, err := json.Marshal(map[string]any{
			"data": []map[string]string{{"type": "betaAppClipInvocationLocalizations", "id": localID}},
		})
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/betaAppClipInvocations", api.Body{
			Data: api.Resource{
				Type:       "betaAppClipInvocations",
				Attributes: map[string]any{"url": tfxURL},
				Relationships: map[string]json.RawMessage{
					"buildBundle":                        api.Rel("buildBundles", tfxBuildBundleID),
					"betaAppClipInvocationLocalizations": locLinkage,
				},
			},
			Included: []api.Resource{{
				Type:       "betaAppClipInvocationLocalizations",
				ID:         localID,
				Attributes: map[string]any{"locale": tfxLocale, "title": tfxTitle},
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created App Clip invocation %s (%s, %s).\n", created.ID, tfxURL, tfxLocale)
		return nil
	},
}

var tfxInvDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an App Clip invocation",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/betaAppClipInvocations/"+tfxID); err != nil {
			return err
		}
		fmt.Printf("Deleted App Clip invocation %s.\n", tfxID)
		return nil
	},
}

var tfxInvLocalizeCmd = &cobra.Command{
	Use:     "localize",
	Short:   "Add a localized title to an App Clip invocation",
	Example: `  asc beta invocations localize --id <invocation-id> --locale en-US --title "Try the clip"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/betaAppClipInvocationLocalizations", api.Body{
			Data: api.Resource{
				Type:       "betaAppClipInvocationLocalizations",
				Attributes: map[string]any{"locale": tfxLocale, "title": tfxTitle},
				Relationships: map[string]json.RawMessage{
					"betaAppClipInvocation": api.Rel("betaAppClipInvocations", tfxID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Added localization %s (%s) to invocation %s.\n", created.ID, tfxLocale, tfxID)
		return nil
	},
}

var tfxInvDeleteLocCmd = &cobra.Command{
	Use:   "delete-localization",
	Short: "Delete an App Clip invocation localization",
	Long: `Delete a betaAppClipInvocationLocalization by its id (shown in the
LOCALIZATIONS column of "asc beta invocations list").`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/betaAppClipInvocationLocalizations/"+tfxID); err != nil {
			return err
		}
		fmt.Printf("Deleted App Clip invocation localization %s.\n", tfxID)
		return nil
	},
}

// --- beta testers reinvite -----------------------------------------------------------------

var tfxReinviteCmd = &cobra.Command{
	Use:     "reinvite",
	Short:   "Resend the TestFlight invitation email to an existing tester",
	Example: `  asc beta testers reinvite --app 6790641087 --email tester@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfxApp)
		if err != nil {
			return err
		}
		tester, err := tfFindTester(ctx, c, tfxEmail, "&filter[apps]="+appID)
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/betaTesterInvitations", api.Body{
			Data: api.Resource{
				Type: "betaTesterInvitations",
				Relationships: map[string]json.RawMessage{
					"betaTester": api.Rel("betaTesters", tester.ID),
					"app":        api.Rel("apps", appID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Re-sent TestFlight invitation to %s (tester %s) for app %s.\n", tfxEmail, tester.ID, appID)
		return nil
	},
}

func init() {
	// --app / --build (build resolution)
	for _, sub := range []*cobra.Command{tfxBuildDetailShowCmd, tfxBuildDetailSetCmd, tfxBundlesCmd} {
		sub.Flags().StringVar(&tfxApp, "app", "", "app id or bundle id (required when --build is a version string)")
		sub.Flags().StringVar(&tfxBuild, "build", "", "build id or build version string (required)")
		_ = sub.MarkFlagRequired("build")
	}
	tfxBuildDetailSetCmd.Flags().StringVar(&tfxAutoNotify, "auto-notify", "", "true or false (required)")
	_ = tfxBuildDetailSetCmd.MarkFlagRequired("auto-notify")

	// recruitment
	for _, sub := range []*cobra.Command{tfxRecruitShowCmd, tfxRecruitSetCmd, tfxRecruitDeleteCmd} {
		sub.Flags().StringVar(&tfxGroupID, "group-id", "", "beta group id (required)")
		_ = sub.MarkFlagRequired("group-id")
	}
	tfxRecruitSetCmd.Flags().StringArrayVar(&tfxFilters, "filter", nil,
		"device/OS filter FAMILY[:MIN_OS[:MAX_OS]], e.g. IPHONE:17.0:18.0 (repeatable, required)")
	_ = tfxRecruitSetCmd.MarkFlagRequired("filter")

	// invocations
	for _, sub := range []*cobra.Command{tfxInvListCmd, tfxInvCreateCmd} {
		sub.Flags().StringVar(&tfxBuildBundleID, "build-bundle-id", "", "build bundle id (required; see: asc beta build-bundles)")
		_ = sub.MarkFlagRequired("build-bundle-id")
	}
	tfxInvCreateCmd.Flags().StringVar(&tfxURL, "url", "", "App Clip invocation URL (required)")
	_ = tfxInvCreateCmd.MarkFlagRequired("url")
	for _, sub := range []*cobra.Command{tfxInvCreateCmd, tfxInvLocalizeCmd} {
		sub.Flags().StringVar(&tfxLocale, "locale", "ja", "locale, e.g. ja / en-US")
		sub.Flags().StringVar(&tfxTitle, "title", "", "localized title shown in TestFlight (required)")
		_ = sub.MarkFlagRequired("title")
	}
	tfxInvDeleteCmd.Flags().StringVar(&tfxID, "id", "", "App Clip invocation id (required)")
	_ = tfxInvDeleteCmd.MarkFlagRequired("id")
	tfxInvLocalizeCmd.Flags().StringVar(&tfxID, "id", "", "App Clip invocation id (required)")
	_ = tfxInvLocalizeCmd.MarkFlagRequired("id")
	tfxInvDeleteLocCmd.Flags().StringVar(&tfxID, "id", "", "App Clip invocation localization id (required)")
	_ = tfxInvDeleteLocCmd.MarkFlagRequired("id")

	// reinvite
	tfxReinviteCmd.Flags().StringVar(&tfxApp, "app", "", "app id or bundle id (required)")
	_ = tfxReinviteCmd.MarkFlagRequired("app")
	tfxReinviteCmd.Flags().StringVar(&tfxEmail, "email", "", "tester email (required)")
	_ = tfxReinviteCmd.MarkFlagRequired("email")

	tfxBuildDetailCmd.AddCommand(tfxBuildDetailShowCmd, tfxBuildDetailSetCmd)
	tfxRecruitCmd.AddCommand(tfxRecruitShowCmd, tfxRecruitSetCmd, tfxRecruitDeleteCmd, tfxRecruitOptionsCmd)
	tfxInvocationsCmd.AddCommand(tfxInvListCmd, tfxInvCreateCmd, tfxInvDeleteCmd, tfxInvLocalizeCmd, tfxInvDeleteLocCmd)
	tfTestersCmd.AddCommand(tfxReinviteCmd)
	tfCmd.AddCommand(tfxBuildDetailCmd, tfxRecruitCmd, tfxBundlesCmd, tfxInvocationsCmd)
}
