package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var reviewCmd = &cobra.Command{
	Use:   "review-detail",
	Short: "Manage the App Review contact info and notes for the editable version",
}

var (
	revApp      string
	revVersion  string
	revFirst    string
	revLast     string
	revPhone    string
	revEmail    string
	revNotes    string
	revDemoUser string
	revDemoPass string
	revDemoReq  bool
)

var reviewShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current App Review detail",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, revApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, revVersion)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreReviewDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Println("No App Review detail set yet.")
			return nil
		}
		for _, k := range []string{"contactFirstName", "contactLastName", "contactPhone", "contactEmail", "demoAccountName", "demoAccountRequired", "notes"} {
			if v, ok := detail.Attributes[k]; ok {
				fmt.Printf("%-20s %v\n", k, v)
			}
		}
		return nil
	},
}

var reviewSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update the App Review contact info, demo account and notes",
	Example: `  asc review-detail set --app 6790641087 \
    --first 邦彦 --last 宮永 --phone +81-50-3159-6143 --email contact@ideamans.com \
    --notes @app-store/review-notes.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, revApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, revVersion)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		strFields := []struct {
			flag, attr string
			val        *string
		}{
			{"first", "contactFirstName", &revFirst},
			{"last", "contactLastName", &revLast},
			{"phone", "contactPhone", &revPhone},
			{"email", "contactEmail", &revEmail},
			{"demo-user", "demoAccountName", &revDemoUser},
			{"demo-pass", "demoAccountPassword", &revDemoPass},
			{"notes", "notes", &revNotes},
		}
		for _, f := range strFields {
			if cmd.Flags().Changed(f.flag) {
				v, err := valueOrFile(*f.val)
				if err != nil {
					return err
				}
				attrs[f.attr] = v
			}
		}
		if cmd.Flags().Changed("demo-required") {
			attrs["demoAccountRequired"] = revDemoReq
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set")
		}

		detail, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreReviewDetail")
		if err != nil {
			return err
		}
		if detail != nil && detail.ID != "" {
			_, err = c.Patch(ctx, "/v1/appStoreReviewDetails/"+detail.ID, api.Body{
				Data: api.Resource{Type: "appStoreReviewDetails", ID: detail.ID, Attributes: attrs},
			})
		} else {
			_, err = c.Post(ctx, "/v1/appStoreReviewDetails", api.Body{
				Data: api.Resource{
					Type:          "appStoreReviewDetails",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Println("App Review detail updated.")
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{reviewShowCmd, reviewSetCmd} {
		sub.Flags().StringVar(&revApp, "app", "", "app id or bundle id (required)")
		sub.Flags().StringVar(&revVersion, "version", "", "version string (default: the editable version)")
		_ = sub.MarkFlagRequired("app")
	}
	reviewSetCmd.Flags().StringVar(&revFirst, "first", "", "contact first name")
	reviewSetCmd.Flags().StringVar(&revLast, "last", "", "contact last name")
	reviewSetCmd.Flags().StringVar(&revPhone, "phone", "", "contact phone")
	reviewSetCmd.Flags().StringVar(&revEmail, "email", "", "contact email")
	reviewSetCmd.Flags().StringVar(&revNotes, "notes", "", "review notes (@file allowed)")
	reviewSetCmd.Flags().StringVar(&revDemoUser, "demo-user", "", "demo account username")
	reviewSetCmd.Flags().StringVar(&revDemoPass, "demo-pass", "", "demo account password")
	reviewSetCmd.Flags().BoolVar(&revDemoReq, "demo-required", false, "whether a demo account is required")

	reviewCmd.AddCommand(reviewShowCmd, reviewSetCmd)
	rootCmd.AddCommand(reviewCmd)
}
