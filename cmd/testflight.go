package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var tfCmd = &cobra.Command{
	Use:   "beta",
	Short: "Manage TestFlight beta testing (groups, testers, localizations, review)",
}

var (
	tfApp             string
	tfGroupID         string
	tfName            string
	tfPublicLink      bool
	tfPublicLinkLimit int
	tfFeedback        bool
	tfBuild           string
	tfEmails          []string
	tfEmail           string
	tfFirstName       string
	tfLastName        string
	tfLocale          string
	tfDesc            string
	tfFeedbackEmail   string
	tfMarketingURL    string
	tfPrivacyURL      string
	tfTVOSPrivacy     string
	tfWhatsNew        string
	tfContactFirst    string
	tfContactLast     string
	tfContactPhone    string
	tfContactEmail    string
	tfDemoName        string
	tfDemoPassword    string
	tfDemoRequired    bool
	tfNotes           string
	tfLicenseText     string
	tfSandboxID       string
	tfSandboxTerr     string
	tfSandboxInter    bool
	tfSandboxRenewal  string
)

// tfLinkage sends a to-many relationship linkage request (POST to add, DELETE
// to remove) with a {"data":[{type,id},...]} body, honoring --dry-run.
func tfLinkage(ctx context.Context, c *api.Client, method, path, typ string, ids []string) error {
	items := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]string{"type": typ, "id": id})
	}
	payload, err := json.Marshal(map[string]any{"data": items})
	if err != nil {
		return err
	}
	if c.DryRun {
		fmt.Fprintf(os.Stderr, "DRY-RUN %s %s\n%s\n", method, path, payload)
		return nil
	}
	_, err = c.Do(ctx, method, path, bytes.NewReader(payload))
	return err
}

// tfResolveBuild accepts a build id (UUID form, returned as-is) or a build
// version string (the CFBundleVersion), which is looked up for the app.
func tfResolveBuild(ctx context.Context, c *api.Client, appRef, buildRef string) (string, error) {
	if buildRef == "" {
		return "", fmt.Errorf("--build is required (build id or build version string)")
	}
	if len(buildRef) == 36 && strings.Count(buildRef, "-") == 4 {
		return buildRef, nil // already a build id
	}
	if appRef == "" {
		return "", fmt.Errorf("--app is required when --build is a version string")
	}
	appID, err := resolveAppID(ctx, c, appRef)
	if err != nil {
		return "", err
	}
	builds, err := c.List(ctx, "/v1/builds?filter[app]="+appID+"&filter[version]="+url.QueryEscape(buildRef)+"&limit=1")
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", fmt.Errorf("no build with version %q for app %s (is it uploaded and processed?)", buildRef, appID)
	}
	return builds[0].ID, nil
}

// tfCreateTester creates (or re-associates) a beta tester by email and adds it
// to a beta group. This is the API's way to both invite new testers and add
// existing ones to a group.
func tfCreateTester(ctx context.Context, c *api.Client, groupID, email, first, last string) (*api.Resource, error) {
	attrs := map[string]any{"email": email}
	if first != "" {
		attrs["firstName"] = first
	}
	if last != "" {
		attrs["lastName"] = last
	}
	groups, err := json.Marshal(map[string]any{"data": []map[string]string{{"type": "betaGroups", "id": groupID}}})
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/v1/betaTesters", api.Body{
		Data: api.Resource{
			Type:          "betaTesters",
			Attributes:    attrs,
			Relationships: map[string]json.RawMessage{"betaGroups": groups},
		},
	})
}

// tfFindTester looks a beta tester up by email; extraFilter narrows the search
// (e.g. "&filter[apps]=<id>" or "&filter[betaGroups]=<id>").
func tfFindTester(ctx context.Context, c *api.Client, email, extraFilter string) (*api.Resource, error) {
	testers, err := c.List(ctx, "/v1/betaTesters?filter[email]="+url.QueryEscape(email)+extraFilter+"&limit=2")
	if err != nil {
		return nil, err
	}
	// The filter can over-match; confirm the attribute before trusting the result.
	if t := findByAttr(testers, "email", email); t != nil {
		return t, nil
	}
	return nil, fmt.Errorf("no beta tester with email %q", email)
}

func tfBoolAttr(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		if v {
			return "true"
		}
		return "false"
	}
	return ""
}

// --- beta groups ---------------------------------------------------------------

var tfGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage TestFlight beta groups",
}

var tfGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's beta groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		groups, err := c.List(ctx, "/v1/betaGroups?filter[app]="+appID+"&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tINTERNAL\tPUBLIC LINK\tID")
		for _, g := range groups {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", g.Str("name"), tfBoolAttr(&g, "isInternalGroup"), g.Str("publicLink"), g.ID)
		}
		return w.Flush()
	},
}

var tfGroupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a beta group",
	Example: `  asc beta groups create --app 6790641087 --name "External Testers" \
    --public-link --public-link-limit 100 --feedback-enabled`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{"name": tfName}
		if cmd.Flags().Changed("public-link") {
			attrs["publicLinkEnabled"] = tfPublicLink
		}
		if cmd.Flags().Changed("public-link-limit") {
			attrs["publicLinkLimitEnabled"] = true
			attrs["publicLinkLimit"] = tfPublicLinkLimit
		}
		if cmd.Flags().Changed("feedback-enabled") {
			attrs["feedbackEnabled"] = tfFeedback
		}
		created, err := c.Post(ctx, "/v1/betaGroups", api.Body{
			Data: api.Resource{
				Type:          "betaGroups",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created beta group %q (%s).\n", tfName, created.ID)
		if link := created.Str("publicLink"); link != "" {
			fmt.Printf("Public link: %s\n", link)
		}
		return nil
	},
}

var tfGroupsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a beta group's name, public link, or feedback settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = tfName
		}
		if cmd.Flags().Changed("public-link") {
			attrs["publicLinkEnabled"] = tfPublicLink
		}
		if cmd.Flags().Changed("public-link-limit") {
			attrs["publicLinkLimitEnabled"] = true
			attrs["publicLinkLimit"] = tfPublicLinkLimit
		}
		if cmd.Flags().Changed("feedback-enabled") {
			attrs["feedbackEnabled"] = tfFeedback
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass --name/--public-link/--public-link-limit/--feedback-enabled")
		}
		updated, err := c.Patch(cmd.Context(), "/v1/betaGroups/"+tfGroupID, api.Body{
			Data: api.Resource{Type: "betaGroups", ID: tfGroupID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Updated beta group %s.\n", tfGroupID)
		if link := updated.Str("publicLink"); link != "" {
			fmt.Printf("Public link: %s\n", link)
		}
		return nil
	},
}

var tfGroupsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a beta group",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/betaGroups/"+tfGroupID); err != nil {
			return err
		}
		fmt.Printf("Deleted beta group %s.\n", tfGroupID)
		return nil
	},
}

var tfGroupsAddBuildCmd = &cobra.Command{
	Use:     "add-build",
	Short:   "Give a beta group access to a build",
	Example: `  asc beta groups add-build --group-id <group-id> --app 6790641087 --build 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfApp, tfBuild)
		if err != nil {
			return err
		}
		if err := tfLinkage(ctx, c, http.MethodPost, "/v1/betaGroups/"+tfGroupID+"/relationships/builds", "builds", []string{buildID}); err != nil {
			return err
		}
		fmt.Printf("Build %s added to group %s.\n", buildID, tfGroupID)
		return nil
	},
}

var tfGroupsRemoveBuildCmd = &cobra.Command{
	Use:   "remove-build",
	Short: "Remove a build from a beta group",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfApp, tfBuild)
		if err != nil {
			return err
		}
		if err := tfLinkage(ctx, c, http.MethodDelete, "/v1/betaGroups/"+tfGroupID+"/relationships/builds", "builds", []string{buildID}); err != nil {
			return err
		}
		fmt.Printf("Build %s removed from group %s.\n", buildID, tfGroupID)
		return nil
	},
}

var tfGroupsAddTestersCmd = &cobra.Command{
	Use:     "add-testers",
	Short:   "Add testers to a beta group by email (invites new testers)",
	Example: `  asc beta groups add-testers --group-id <group-id> --email a@example.com --email b@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		for _, email := range tfEmails {
			tester, err := tfCreateTester(ctx, c, tfGroupID, email, "", "")
			if err != nil {
				return fmt.Errorf("%s: %w", email, err)
			}
			fmt.Printf("Added %s to group %s (tester %s).\n", email, tfGroupID, tester.ID)
		}
		return nil
	},
}

var tfGroupsRemoveTestersCmd = &cobra.Command{
	Use:   "remove-testers",
	Short: "Remove testers from a beta group by email",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		var ids []string
		for _, email := range tfEmails {
			tester, err := tfFindTester(ctx, c, email, "&filter[betaGroups]="+tfGroupID)
			if err != nil {
				return err
			}
			ids = append(ids, tester.ID)
		}
		if err := tfLinkage(ctx, c, http.MethodDelete, "/v1/betaGroups/"+tfGroupID+"/relationships/betaTesters", "betaTesters", ids); err != nil {
			return err
		}
		fmt.Printf("Removed %d tester(s) from group %s.\n", len(ids), tfGroupID)
		return nil
	},
}

// --- beta testers ----------------------------------------------------------------

var tfTestersCmd = &cobra.Command{
	Use:   "testers",
	Short: "Manage TestFlight beta testers",
}

var tfTestersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's beta testers (optionally within one group)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		path := "/v1/betaTesters?filter[apps]=" + appID + "&limit=200"
		if tfGroupID != "" {
			path += "&filter[betaGroups]=" + tfGroupID
		}
		testers, err := c.List(ctx, path)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "EMAIL\tNAME\tSTATE\tINVITE\tID")
		for _, t := range testers {
			name := strings.TrimSpace(t.Str("firstName") + " " + t.Str("lastName"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Str("email"), name, t.Str("state"), t.Str("inviteType"), t.ID)
		}
		return w.Flush()
	},
}

var tfTestersInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a tester by email to a beta group",
	Long: `Create a beta tester and add it to a group; Apple emails the TestFlight
invitation. The group determines the app, so no --app flag is needed.`,
	Example: `  asc beta testers invite --group-id <group-id> --email tester@example.com --first-name Hanako`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		tester, err := tfCreateTester(cmd.Context(), c, tfGroupID, tfEmail, tfFirstName, tfLastName)
		if err != nil {
			return err
		}
		fmt.Printf("Invited %s to group %s (tester %s).\n", tfEmail, tfGroupID, tester.ID)
		return nil
	},
}

var tfTestersRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a tester's access to the app (all groups and builds)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		tester, err := tfFindTester(ctx, c, tfEmail, "&filter[apps]="+appID)
		if err != nil {
			return err
		}
		if err := tfLinkage(ctx, c, http.MethodDelete, "/v1/betaTesters/"+tester.ID+"/relationships/apps", "apps", []string{appID}); err != nil {
			return err
		}
		fmt.Printf("Removed %s from app %s.\n", tfEmail, appID)
		return nil
	},
}

// --- beta app / build localizations ------------------------------------------------

var tfAppLocalizeCmd = &cobra.Command{
	Use:   "app-localize",
	Short: "Set the TestFlight beta app information for a locale",
	Example: `  asc beta app-localize --app 6790641087 --locale ja \
    --description @testflight/description.txt --feedback-email support@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		fields := []struct {
			flag, attr string
			val        *string
		}{
			{"description", "description", &tfDesc},
			{"feedback-email", "feedbackEmail", &tfFeedbackEmail},
			{"marketing-url", "marketingUrl", &tfMarketingURL},
			{"privacy-policy-url", "privacyPolicyUrl", &tfPrivacyURL},
			{"tv-os-privacy-policy", "tvOsPrivacyPolicy", &tfTVOSPrivacy},
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
			return fmt.Errorf("nothing to set: pass --description/--feedback-email/--marketing-url/--privacy-policy-url/--tv-os-privacy-policy")
		}
		locs, err := c.List(ctx, "/v1/betaAppLocalizations?filter[app]="+appID+"&limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", tfLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/betaAppLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "betaAppLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			attrs["locale"] = tfLocale
			_, err = c.Post(ctx, "/v1/betaAppLocalizations", api.Body{
				Data: api.Resource{
					Type:          "betaAppLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Beta app localization (%s) updated.\n", tfLocale)
		return nil
	},
}

var tfBuildLocalizeCmd = &cobra.Command{
	Use:     "build-localize",
	Short:   "Set a build's TestFlight \"what to test\" notes for a locale",
	Example: `  asc beta build-localize --app 6790641087 --build 42 --locale ja --whats-new @testflight/whats-new.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfApp, tfBuild)
		if err != nil {
			return err
		}
		whatsNew, err := valueOrFile(tfWhatsNew)
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v1/builds/"+buildID+"/betaBuildLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", tfLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/betaBuildLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "betaBuildLocalizations", ID: existing.ID, Attributes: map[string]any{"whatsNew": whatsNew}},
			})
		} else {
			_, err = c.Post(ctx, "/v1/betaBuildLocalizations", api.Body{
				Data: api.Resource{
					Type:          "betaBuildLocalizations",
					Attributes:    map[string]any{"locale": tfLocale, "whatsNew": whatsNew},
					Relationships: map[string]json.RawMessage{"build": api.Rel("builds", buildID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Build %s localization (%s) updated.\n", buildID, tfLocale)
		return nil
	},
}

// --- beta review detail --------------------------------------------------------------

var tfReviewDetailCmd = &cobra.Command{
	Use:   "review-detail",
	Short: "Show or set the app's TestFlight beta review contact information",
}

var tfReviewDetailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the beta app review detail",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/betaAppReviewDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Println("No beta app review detail.")
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

var tfReviewDetailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set beta review contact and demo account fields",
	Example: `  asc beta review-detail set --app 6790641087 \
    --contact-first-name Taro --contact-last-name Yamada \
    --contact-email review@example.com --contact-phone +81-3-0000-0000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/betaAppReviewDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			return fmt.Errorf("app %s has no beta app review detail", appID)
		}
		attrs := map[string]any{}
		fields := []struct {
			flag, attr string
			val        *string
		}{
			{"contact-first-name", "contactFirstName", &tfContactFirst},
			{"contact-last-name", "contactLastName", &tfContactLast},
			{"contact-phone", "contactPhone", &tfContactPhone},
			{"contact-email", "contactEmail", &tfContactEmail},
			{"demo-account-name", "demoAccountName", &tfDemoName},
			{"demo-account-password", "demoAccountPassword", &tfDemoPassword},
		}
		for _, f := range fields {
			if cmd.Flags().Changed(f.flag) {
				attrs[f.attr] = *f.val
			}
		}
		if cmd.Flags().Changed("demo-account-required") {
			attrs["demoAccountRequired"] = tfDemoRequired
		}
		if cmd.Flags().Changed("notes") {
			v, err := valueOrFile(tfNotes)
			if err != nil {
				return err
			}
			attrs["notes"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass one or more --contact-*/--demo-account-*/--notes flags")
		}
		_, err = c.Patch(ctx, "/v1/betaAppReviewDetails/"+detail.ID, api.Body{
			Data: api.Resource{Type: "betaAppReviewDetails", ID: detail.ID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Println("Beta app review detail updated.")
		return nil
	},
}

// --- beta license agreement ------------------------------------------------------------

var tfLicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Show or set the app's TestFlight beta license agreement",
}

var tfLicenseShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the beta license agreement text",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		lic, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/betaLicenseAgreement")
		if err != nil {
			return err
		}
		if lic == nil || lic.Str("agreementText") == "" {
			fmt.Println("No beta license agreement text set.")
			return nil
		}
		fmt.Println(lic.Str("agreementText"))
		return nil
	},
}

var tfLicenseSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set the beta license agreement text",
	Example: `  asc beta license set --app 6790641087 --text @testflight/license.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, tfApp)
		if err != nil {
			return err
		}
		lic, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/betaLicenseAgreement")
		if err != nil {
			return err
		}
		if lic == nil || lic.ID == "" {
			return fmt.Errorf("app %s has no beta license agreement resource", appID)
		}
		text, err := valueOrFile(tfLicenseText)
		if err != nil {
			return err
		}
		_, err = c.Patch(ctx, "/v1/betaLicenseAgreements/"+lic.ID, api.Body{
			Data: api.Resource{Type: "betaLicenseAgreements", ID: lic.ID, Attributes: map[string]any{"agreementText": text}},
		})
		if err != nil {
			return err
		}
		fmt.Println("Beta license agreement updated.")
		return nil
	},
}

// --- beta submit ----------------------------------------------------------------------

var tfSubmitCmd = &cobra.Command{
	Use:     "submit",
	Short:   "Submit a build for beta review (required before external testing)",
	Example: `  asc beta submit --app 6790641087 --build 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := tfResolveBuild(ctx, c, tfApp, tfBuild)
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/betaAppReviewSubmissions", api.Body{
			Data: api.Resource{
				Type:          "betaAppReviewSubmissions",
				Relationships: map[string]json.RawMessage{"build": api.Rel("builds", buildID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted build %s for beta review.\n", buildID)
		return nil
	},
}

// --- sandbox testers (v2) ---------------------------------------------------------------

var tfSandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Manage sandbox testers (list, update, clear purchase history)",
	Long: `Manage App Store sandbox test accounts via the v2 sandboxTesters API.
The API only supports listing, updating, and clearing purchase history;
creating or deleting sandbox testers must be done in App Store Connect.`,
}

var tfSandboxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sandbox testers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		testers, err := c.List(cmd.Context(), "/v2/sandboxTesters?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ACCOUNT\tNAME\tTERRITORY\tINTERRUPT\tID")
		for _, t := range testers {
			name := strings.TrimSpace(t.Str("firstName") + " " + t.Str("lastName"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Str("acAccountName"), name, t.Str("territory"), tfBoolAttr(&t, "interruptPurchases"), t.ID)
		}
		return w.Flush()
	},
}

var tfSandboxUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a sandbox tester (territory, interrupted purchases, renewal rate)",
	Example: `  asc sandbox update --id <tester-id> --territory USA --interrupt-purchases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("territory") {
			attrs["territory"] = tfSandboxTerr
		}
		if cmd.Flags().Changed("interrupt-purchases") {
			attrs["interruptPurchases"] = tfSandboxInter
		}
		if cmd.Flags().Changed("renewal-rate") {
			attrs["subscriptionRenewalRate"] = tfSandboxRenewal
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass --territory/--interrupt-purchases/--renewal-rate")
		}
		_, err = c.Patch(cmd.Context(), "/v2/sandboxTesters/"+tfSandboxID, api.Body{
			Data: api.Resource{Type: "sandboxTesters", ID: tfSandboxID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Sandbox tester %s updated.\n", tfSandboxID)
		return nil
	},
}

var tfSandboxClearHistoryCmd = &cobra.Command{
	Use:   "clear-history",
	Short: "Clear a sandbox tester's purchase history",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		testers, err := json.Marshal(map[string]any{"data": []map[string]string{{"type": "sandboxTesters", "id": tfSandboxID}}})
		if err != nil {
			return err
		}
		_, err = c.Post(cmd.Context(), "/v2/sandboxTestersClearPurchaseHistoryRequest", api.Body{
			Data: api.Resource{
				Type:          "sandboxTestersClearPurchaseHistoryRequest",
				Relationships: map[string]json.RawMessage{"sandboxTesters": testers},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Purchase history cleared for sandbox tester %s.\n", tfSandboxID)
		return nil
	},
}

func init() {
	// --app
	for _, sub := range []*cobra.Command{
		tfGroupsListCmd, tfTestersListCmd, tfTestersRemoveCmd, tfAppLocalizeCmd,
		tfReviewDetailShowCmd, tfReviewDetailSetCmd, tfLicenseShowCmd, tfLicenseSetCmd, tfGroupsCreateCmd,
	} {
		sub.Flags().StringVar(&tfApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{tfGroupsAddBuildCmd, tfGroupsRemoveBuildCmd, tfBuildLocalizeCmd, tfSubmitCmd} {
		sub.Flags().StringVar(&tfApp, "app", "", "app id or bundle id (required when --build is a version string)")
		sub.Flags().StringVar(&tfBuild, "build", "", "build id or build version string (required)")
		_ = sub.MarkFlagRequired("build")
	}

	// --group-id
	for _, sub := range []*cobra.Command{
		tfGroupsUpdateCmd, tfGroupsDeleteCmd, tfGroupsAddBuildCmd, tfGroupsRemoveBuildCmd,
		tfGroupsAddTestersCmd, tfGroupsRemoveTestersCmd, tfTestersInviteCmd,
	} {
		sub.Flags().StringVar(&tfGroupID, "group-id", "", "beta group id (required)")
		_ = sub.MarkFlagRequired("group-id")
	}
	tfTestersListCmd.Flags().StringVar(&tfGroupID, "group-id", "", "limit to one beta group")

	// group attributes
	tfGroupsCreateCmd.Flags().StringVar(&tfName, "name", "", "group name (required)")
	_ = tfGroupsCreateCmd.MarkFlagRequired("name")
	tfGroupsUpdateCmd.Flags().StringVar(&tfName, "name", "", "group name")
	for _, sub := range []*cobra.Command{tfGroupsCreateCmd, tfGroupsUpdateCmd} {
		sub.Flags().BoolVar(&tfPublicLink, "public-link", false, "enable the public TestFlight invitation link")
		sub.Flags().IntVar(&tfPublicLinkLimit, "public-link-limit", 0, "maximum testers via the public link (enables the limit)")
		sub.Flags().BoolVar(&tfFeedback, "feedback-enabled", false, "enable tester feedback")
	}

	// tester emails
	for _, sub := range []*cobra.Command{tfGroupsAddTestersCmd, tfGroupsRemoveTestersCmd} {
		sub.Flags().StringArrayVar(&tfEmails, "email", nil, "tester email (repeatable, required)")
		_ = sub.MarkFlagRequired("email")
	}
	for _, sub := range []*cobra.Command{tfTestersInviteCmd, tfTestersRemoveCmd} {
		sub.Flags().StringVar(&tfEmail, "email", "", "tester email (required)")
		_ = sub.MarkFlagRequired("email")
	}
	tfTestersInviteCmd.Flags().StringVar(&tfFirstName, "first-name", "", "tester first name")
	tfTestersInviteCmd.Flags().StringVar(&tfLastName, "last-name", "", "tester last name")

	// localizations
	tfAppLocalizeCmd.Flags().StringVar(&tfLocale, "locale", "ja", "locale, e.g. ja / en-US")
	tfAppLocalizeCmd.Flags().StringVar(&tfDesc, "description", "", "beta app description (@file allowed)")
	tfAppLocalizeCmd.Flags().StringVar(&tfFeedbackEmail, "feedback-email", "", "email testers send feedback to")
	tfAppLocalizeCmd.Flags().StringVar(&tfMarketingURL, "marketing-url", "", "marketing URL")
	tfAppLocalizeCmd.Flags().StringVar(&tfPrivacyURL, "privacy-policy-url", "", "privacy policy URL")
	tfAppLocalizeCmd.Flags().StringVar(&tfTVOSPrivacy, "tv-os-privacy-policy", "", "tvOS privacy policy text (@file allowed)")

	tfBuildLocalizeCmd.Flags().StringVar(&tfLocale, "locale", "ja", "locale, e.g. ja / en-US")
	tfBuildLocalizeCmd.Flags().StringVar(&tfWhatsNew, "whats-new", "", "what to test in this build (@file allowed, required)")
	_ = tfBuildLocalizeCmd.MarkFlagRequired("whats-new")

	// review detail
	tfReviewDetailSetCmd.Flags().StringVar(&tfContactFirst, "contact-first-name", "", "review contact first name")
	tfReviewDetailSetCmd.Flags().StringVar(&tfContactLast, "contact-last-name", "", "review contact last name")
	tfReviewDetailSetCmd.Flags().StringVar(&tfContactPhone, "contact-phone", "", "review contact phone")
	tfReviewDetailSetCmd.Flags().StringVar(&tfContactEmail, "contact-email", "", "review contact email")
	tfReviewDetailSetCmd.Flags().StringVar(&tfDemoName, "demo-account-name", "", "demo account user name")
	tfReviewDetailSetCmd.Flags().StringVar(&tfDemoPassword, "demo-account-password", "", "demo account password")
	tfReviewDetailSetCmd.Flags().BoolVar(&tfDemoRequired, "demo-account-required", false, "a demo account is required to review")
	tfReviewDetailSetCmd.Flags().StringVar(&tfNotes, "notes", "", "notes for the reviewer (@file allowed)")

	// license
	tfLicenseSetCmd.Flags().StringVar(&tfLicenseText, "text", "", "agreement text (@file allowed, required)")
	_ = tfLicenseSetCmd.MarkFlagRequired("text")

	// sandbox
	for _, sub := range []*cobra.Command{tfSandboxUpdateCmd, tfSandboxClearHistoryCmd} {
		sub.Flags().StringVar(&tfSandboxID, "id", "", "sandbox tester id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	tfSandboxUpdateCmd.Flags().StringVar(&tfSandboxTerr, "territory", "", "territory code, e.g. JPN, USA")
	tfSandboxUpdateCmd.Flags().BoolVar(&tfSandboxInter, "interrupt-purchases", false, "simulate interrupted purchases")
	tfSandboxUpdateCmd.Flags().StringVar(&tfSandboxRenewal, "renewal-rate", "", "subscription renewal rate, e.g. MONTHLY_RENEWAL_EVERY_ONE_HOUR")

	tfGroupsCmd.AddCommand(tfGroupsListCmd, tfGroupsCreateCmd, tfGroupsUpdateCmd, tfGroupsDeleteCmd,
		tfGroupsAddBuildCmd, tfGroupsRemoveBuildCmd, tfGroupsAddTestersCmd, tfGroupsRemoveTestersCmd)
	tfTestersCmd.AddCommand(tfTestersListCmd, tfTestersInviteCmd, tfTestersRemoveCmd)
	tfReviewDetailCmd.AddCommand(tfReviewDetailShowCmd, tfReviewDetailSetCmd)
	tfLicenseCmd.AddCommand(tfLicenseShowCmd, tfLicenseSetCmd)
	tfCmd.AddCommand(tfGroupsCmd, tfTestersCmd, tfAppLocalizeCmd, tfBuildLocalizeCmd, tfReviewDetailCmd, tfLicenseCmd, tfSubmitCmd)
	tfSandboxCmd.AddCommand(tfSandboxListCmd, tfSandboxUpdateCmd, tfSandboxClearHistoryCmd)
	rootCmd.AddCommand(tfCmd, tfSandboxCmd)
}
