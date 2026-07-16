package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var appclipsCmd = &cobra.Command{
	Use:   "appclips",
	Short: "Manage App Clips (default experiences, localizations, header images, review details)",
}

var (
	aclApp              string
	aclClipID           string
	aclAction           string
	aclReleaseVersionID string
	aclExperienceID     string
	aclLocale           string
	aclSubtitle         string
	aclLocalizationID   string
	aclFile             string
	aclInvocationURLs   []string
)

var appclipsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's App Clips",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, aclApp)
		if err != nil {
			return err
		}
		clips, err := c.List(ctx, "/v1/apps/"+appID+"/appClips?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "BUNDLE_ID\tID")
		for _, clip := range clips {
			fmt.Fprintf(w, "%s\t%s\n", clip.Str("bundleId"), clip.ID)
		}
		return w.Flush()
	},
}

var appclipsExperiencesCmd = &cobra.Command{
	Use:   "experiences",
	Short: "List an App Clip's default experiences",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		experiences, err := c.List(cmd.Context(), "/v1/appClips/"+aclClipID+"/appClipDefaultExperiences?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ACTION\tID")
		for _, e := range experiences {
			fmt.Fprintf(w, "%s\t%s\n", e.Str("action"), e.ID)
		}
		return w.Flush()
	},
}

var appclipsCreateExperienceCmd = &cobra.Command{
	Use:   "create-experience",
	Short: "Create a default experience for an App Clip",
	Example: `  asc appclips create-experience --clip-id <appClip id> --action OPEN \
    --release-version-id <appStoreVersion id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		res := api.Resource{
			Type:          "appClipDefaultExperiences",
			Relationships: map[string]json.RawMessage{"appClip": api.Rel("appClips", aclClipID)},
		}
		if cmd.Flags().Changed("action") {
			res.Attributes = map[string]any{"action": aclAction}
		}
		if aclReleaseVersionID != "" {
			res.Relationships["releaseWithAppStoreVersion"] = api.Rel("appStoreVersions", aclReleaseVersionID)
		}
		created, err := c.Post(cmd.Context(), "/v1/appClipDefaultExperiences", api.Body{Data: res})
		if err != nil {
			return err
		}
		fmt.Printf("Created default experience %s.\n", created.ID)
		return nil
	},
}

var appclipsLocalizeCmd = &cobra.Command{
	Use:     "localize",
	Short:   "Set a default experience's localized subtitle",
	Example: `  asc appclips localize --experience-id <id> --locale ja --subtitle "その場ですぐ注文"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		subtitle, err := valueOrFile(aclSubtitle)
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v1/appClipDefaultExperiences/"+aclExperienceID+"/appClipDefaultExperienceLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", aclLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/appClipDefaultExperienceLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{
					Type:       "appClipDefaultExperienceLocalizations",
					ID:         existing.ID,
					Attributes: map[string]any{"subtitle": subtitle},
				},
			})
		} else {
			_, err = c.Post(ctx, "/v1/appClipDefaultExperienceLocalizations", api.Body{
				Data: api.Resource{
					Type:          "appClipDefaultExperienceLocalizations",
					Attributes:    map[string]any{"locale": aclLocale, "subtitle": subtitle},
					Relationships: map[string]json.RawMessage{"appClipDefaultExperience": api.Rel("appClipDefaultExperiences", aclExperienceID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Default experience %s localization (%s) updated.\n", aclExperienceID, aclLocale)
		return nil
	},
}

var appclipsUploadHeaderCmd = &cobra.Command{
	Use:   "upload-header",
	Short: "Upload the header image for a default experience localization",
	Long: `Upload an App Clip card header image (reserve → upload → commit) and attach
it to a default experience localization. Find the localization id via:
GET /v1/appClipDefaultExperiences/{id}/appClipDefaultExperienceLocalizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := uploadAsset(cmd.Context(), c, assetSpec{
			reserveType: "appClipHeaderImages",
			relName:     "appClipDefaultExperienceLocalization",
			relType:     "appClipDefaultExperienceLocalizations",
			relID:       aclLocalizationID,
			filePath:    aclFile,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded App Clip header image -> %s\n", id)
		return nil
	},
}

var appclipsReviewDetailCmd = &cobra.Command{
	Use:   "review-detail",
	Short: "Show or set a default experience's App Store review detail (invocation URLs)",
}

var appclipsReviewDetailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the review detail for a default experience",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(cmd.Context(), "/v1/appClipDefaultExperiences/"+aclExperienceID+"/appClipAppStoreReviewDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Println("No review detail set for this experience.")
			return nil
		}
		out, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	},
}

var appclipsReviewDetailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the invocation URLs on a default experience's review detail",
	Example: `  asc appclips review-detail set --experience-id <id> \
    --invocation-urls "https://example.com/clip/a,https://example.com/clip/b"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		attrs := map[string]any{"invocationUrls": aclInvocationURLs}
		existing, err := c.GetOptional(ctx, "/v1/appClipDefaultExperiences/"+aclExperienceID+"/appClipAppStoreReviewDetail")
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != "" {
			_, err = c.Patch(ctx, "/v1/appClipAppStoreReviewDetails/"+existing.ID, api.Body{
				Data: api.Resource{Type: "appClipAppStoreReviewDetails", ID: existing.ID, Attributes: attrs},
			})
		} else {
			_, err = c.Post(ctx, "/v1/appClipAppStoreReviewDetails", api.Body{
				Data: api.Resource{
					Type:          "appClipAppStoreReviewDetails",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appClipDefaultExperience": api.Rel("appClipDefaultExperiences", aclExperienceID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Review detail for experience %s updated (%d invocation URL(s)).\n", aclExperienceID, len(aclInvocationURLs))
		return nil
	},
}

var appclipsAdvancedCmd = &cobra.Command{
	Use:   "advanced",
	Short: "Advanced App Clip experiences (read-only)",
	Long: `Advanced App Clip experiences. Creating one requires header image and
localization relationships in a single request (plus optional place data), which
is beyond this command's scope — use "asc api" with a hand-written body for
creation. Listing is supported here.`,
}

var appclipsAdvancedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List an App Clip's advanced experiences",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		experiences, err := c.List(cmd.Context(), "/v1/appClips/"+aclClipID+"/appClipAdvancedExperiences?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "LINK\tACTION\tSTATUS\tID")
		for _, e := range experiences {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Str("link"), e.Str("action"), e.Str("status"), e.ID)
		}
		return w.Flush()
	},
}

func init() {
	appclipsListCmd.Flags().StringVar(&aclApp, "app", "", "app id or bundle id (required)")
	_ = appclipsListCmd.MarkFlagRequired("app")

	for _, sub := range []*cobra.Command{appclipsExperiencesCmd, appclipsCreateExperienceCmd, appclipsAdvancedListCmd} {
		sub.Flags().StringVar(&aclClipID, "clip-id", "", "appClip id (see: asc appclips list) (required)")
		_ = sub.MarkFlagRequired("clip-id")
	}

	appclipsCreateExperienceCmd.Flags().StringVar(&aclAction, "action", "", "invocation action: OPEN, VIEW or PLAY")
	appclipsCreateExperienceCmd.Flags().StringVar(&aclReleaseVersionID, "release-version-id", "", "appStoreVersion id to release the experience with")

	for _, sub := range []*cobra.Command{appclipsLocalizeCmd, appclipsReviewDetailShowCmd, appclipsReviewDetailSetCmd} {
		sub.Flags().StringVar(&aclExperienceID, "experience-id", "", "appClipDefaultExperience id (required)")
		_ = sub.MarkFlagRequired("experience-id")
	}
	appclipsLocalizeCmd.Flags().StringVar(&aclLocale, "locale", "ja", "locale, e.g. ja / en-US")
	appclipsLocalizeCmd.Flags().StringVar(&aclSubtitle, "subtitle", "", "App Clip card subtitle (@file allowed) (required)")
	_ = appclipsLocalizeCmd.MarkFlagRequired("subtitle")

	appclipsUploadHeaderCmd.Flags().StringVar(&aclLocalizationID, "localization-id", "", "appClipDefaultExperienceLocalization id (required)")
	appclipsUploadHeaderCmd.Flags().StringVar(&aclFile, "file", "", "header image file (required)")
	_ = appclipsUploadHeaderCmd.MarkFlagRequired("localization-id")
	_ = appclipsUploadHeaderCmd.MarkFlagRequired("file")

	appclipsReviewDetailSetCmd.Flags().StringSliceVar(&aclInvocationURLs, "invocation-urls", nil, "comma-separated invocation URLs for review (required)")
	_ = appclipsReviewDetailSetCmd.MarkFlagRequired("invocation-urls")

	appclipsReviewDetailCmd.AddCommand(appclipsReviewDetailShowCmd, appclipsReviewDetailSetCmd)
	appclipsAdvancedCmd.AddCommand(appclipsAdvancedListCmd)
	appclipsCmd.AddCommand(appclipsListCmd, appclipsExperiencesCmd, appclipsCreateExperienceCmd,
		appclipsLocalizeCmd, appclipsUploadHeaderCmd, appclipsReviewDetailCmd, appclipsAdvancedCmd)
	rootCmd.AddCommand(appclipsCmd)
}
