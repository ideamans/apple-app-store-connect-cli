package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	apxApp           string
	apxPrimaryLocale string
	apxContentRights string
)

var apxShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an app's attributes (name, bundle id, SKU, locale, content rights)",
	Example: `  asc apps show --app 6790641087
  asc apps show --app com.example.myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, apxApp)
		if err != nil {
			return err
		}
		app, _, err := c.Get(ctx, "/v1/apps/"+appID)
		if err != nil {
			return err
		}
		out := map[string]any{"id": app.ID}
		for k, v := range app.Attributes {
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

var apxUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update app-level settings (primary locale, content rights declaration)",
	Example: `  asc apps update --app 6790641087 --primary-locale ja
  asc apps update --app 6790641087 --content-rights DOES_NOT_USE_THIRD_PARTY_CONTENT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, apxApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("primary-locale") {
			attrs["primaryLocale"] = apxPrimaryLocale
		}
		if cmd.Flags().Changed("content-rights") {
			attrs["contentRightsDeclaration"] = apxContentRights
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --primary-locale and/or --content-rights")
		}
		_, err = c.Patch(ctx, "/v1/apps/"+appID, api.Body{
			Data: api.Resource{Type: "apps", ID: appID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("App %s updated.\n", appID)
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{apxShowCmd, apxUpdateCmd} {
		sub.Flags().StringVar(&apxApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	apxUpdateCmd.Flags().StringVar(&apxPrimaryLocale, "primary-locale", "", "primary locale, e.g. ja / en-US")
	apxUpdateCmd.Flags().StringVar(&apxContentRights, "content-rights", "", "DOES_NOT_USE_THIRD_PARTY_CONTENT or USES_THIRD_PARTY_CONTENT")

	appsCmd.AddCommand(apxShowCmd, apxUpdateCmd)
}
