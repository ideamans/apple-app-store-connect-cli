package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	subApp         string
	subVersion     string
	subPlatform    string
	subPrepareOnly bool
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit the app version for App Store review",
	Long: `Submit the editable version for review using the reviewSubmissions flow:
find or create an open review submission for the platform, add the version as a
submission item, then mark the submission submitted.

Preconditions (not checked here; the API will reject if unmet): metadata and
localizations complete, screenshots uploaded, a build attached, pricing set,
age rating set, and — done in the web UI — the App Privacy questionnaire and,
for paid in-app purchases, the Paid Apps agreement. Use --prepare-only to stage
the submission without finalizing.`,
	Example: `  asc submit --app 6790641087
  asc submit --app 6790641087 --version 1.0 --prepare-only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, subVersion)
		if err != nil {
			return err
		}
		subID, err := openReviewSubmission(ctx, c, appID, subPlatform)
		if err != nil {
			return err
		}
		if _, err := c.Post(ctx, "/v1/reviewSubmissionItems", api.Body{
			Data: api.Resource{
				Type: "reviewSubmissionItems",
				Relationships: map[string]json.RawMessage{
					"reviewSubmission": api.Rel("reviewSubmissions", subID),
					"appStoreVersion":  api.Rel("appStoreVersions", ver.ID),
				},
			},
		}); err != nil {
			return err
		}
		fmt.Printf("Added version %s to review submission %s.\n", ver.Str("versionString"), subID)

		if subPrepareOnly {
			fmt.Println("--prepare-only: not finalizing. Submit later with: asc submit --app", subApp)
			return nil
		}
		if _, err := c.Patch(ctx, "/v1/reviewSubmissions/"+subID, api.Body{
			Data: api.Resource{Type: "reviewSubmissions", ID: subID, Attributes: map[string]any{"submitted": true}},
		}); err != nil {
			return err
		}
		fmt.Println("Submitted for review.")
		return nil
	},
}

// openReviewSubmission returns an existing not-yet-submitted review submission
// for the platform, or creates one.
func openReviewSubmission(ctx context.Context, c *api.Client, appID, platform string) (string, error) {
	subs, err := c.List(ctx, "/v1/apps/"+appID+"/reviewSubmissions?filter[platform]="+platform+"&limit=50")
	if err != nil {
		return "", err
	}
	for i := range subs {
		if subs[i].Str("state") == "READY_FOR_REVIEW" {
			return subs[i].ID, nil
		}
	}
	created, err := c.Post(ctx, "/v1/reviewSubmissions", api.Body{
		Data: api.Resource{
			Type:          "reviewSubmissions",
			Attributes:    map[string]any{"platform": platform},
			Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

func init() {
	submitCmd.Flags().StringVar(&subApp, "app", "", "app id or bundle id (required)")
	submitCmd.Flags().StringVar(&subVersion, "version", "", "version string (default: the editable version)")
	submitCmd.Flags().StringVar(&subPlatform, "platform", "IOS", "platform: IOS, MAC_OS, TV_OS, VISION_OS")
	submitCmd.Flags().BoolVar(&subPrepareOnly, "prepare-only", false, "stage the submission item without finalizing")
	_ = submitCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(submitCmd)
}
