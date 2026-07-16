package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var appinfoCmd = &cobra.Command{
	Use:   "appinfo",
	Short: "Work with app-level information (name, subtitle, category, age rating)",
	Long: `Work with the app's editable appInfo: localized name/subtitle/privacy URL,
primary/secondary category, and the age-rating declaration.

These fields live on the appInfo resource, not on a specific version.`,
}

var (
	appinfoApp      string
	appinfoLocale   string
	appinfoName     string
	appinfoSubtitle string
	appinfoPrivacy  string
	appinfoPrimary  string
	appinfoSecond   string
	appinfoAttrs    string
)

var appinfoShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the app's appInfo (state, categories, age rating, localizations)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, appinfoApp)
		if err != nil {
			return err
		}
		info, err := editableAppInfo(ctx, c, appID)
		if err != nil {
			return err
		}
		full, included, err := c.Get(ctx, "/v1/appInfos/"+info.ID+"?include=primaryCategory,secondaryCategory,ageRatingDeclaration")
		if err != nil {
			return err
		}
		fmt.Printf("appInfo %s  state=%s\n", full.ID, full.Str("state"))
		for _, inc := range included {
			switch inc.Type {
			case "appCategories":
				fmt.Printf("  category: %s\n", inc.ID)
			case "ageRatingDeclarations":
				fmt.Printf("  ageRatingDeclaration: %s\n", inc.ID)
			}
		}
		locs, err := c.List(ctx, "/v1/appInfos/"+info.ID+"/appInfoLocalizations?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "LOCALE\tNAME\tSUBTITLE")
		for _, l := range locs {
			fmt.Fprintf(w, "%s\t%s\t%s\n", l.Str("locale"), l.Str("name"), l.Str("subtitle"))
		}
		return w.Flush()
	},
}

var appinfoLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set the localized name, subtitle and privacy policy URL",
	Example: `  asc appinfo localize --app 6790641087 --locale ja \
    --name "日本領収書スキャン" --subtitle "データ化するだけのAIレシートスキャナ" \
    --privacy-url https://japan-receipt-scan.web.app/privacy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, appinfoApp)
		if err != nil {
			return err
		}
		info, err := editableAppInfo(ctx, c, appID)
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v1/appInfos/"+info.ID+"/appInfoLocalizations?limit=50")
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = appinfoName
		}
		if cmd.Flags().Changed("subtitle") {
			attrs["subtitle"] = appinfoSubtitle
		}
		if cmd.Flags().Changed("privacy-url") {
			attrs["privacyPolicyUrl"] = appinfoPrivacy
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name, --subtitle and/or --privacy-url")
		}
		if existing := findByAttr(locs, "locale", appinfoLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/appInfoLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "appInfoLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			attrs["locale"] = appinfoLocale
			_, err = c.Post(ctx, "/v1/appInfoLocalizations", api.Body{
				Data: api.Resource{
					Type:          "appInfoLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appInfo": api.Rel("appInfos", info.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("appInfo localization (%s) updated.\n", appinfoLocale)
		return nil
	},
}

var appinfoCategoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Set the primary and/or secondary category",
	Long: `Set the app's primary and/or secondary category. Category ids are the App
Store Connect enum values, e.g. PRODUCTIVITY, BUSINESS, FINANCE, UTILITIES.
List available ids with: asc api /v1/appCategories`,
	Example: `  asc appinfo category --app 6790641087 --primary PRODUCTIVITY --secondary BUSINESS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, appinfoApp)
		if err != nil {
			return err
		}
		info, err := editableAppInfo(ctx, c, appID)
		if err != nil {
			return err
		}
		rels := map[string]json.RawMessage{}
		if cmd.Flags().Changed("primary") {
			rels["primaryCategory"] = api.Rel("appCategories", appinfoPrimary)
		}
		if cmd.Flags().Changed("secondary") {
			rels["secondaryCategory"] = api.Rel("appCategories", appinfoSecond)
		}
		if len(rels) == 0 {
			return fmt.Errorf("nothing to set: pass --primary and/or --secondary")
		}
		_, err = c.Patch(ctx, "/v1/appInfos/"+info.ID, api.Body{
			Data: api.Resource{Type: "appInfos", ID: info.ID, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Println("Categories updated.")
		return nil
	},
}

var appinfoAgeRatingCmd = &cobra.Command{
	Use:   "age-rating",
	Short: "Set the age-rating declaration from a JSON attributes file",
	Long: `Set the age-rating declaration. Because Apple periodically changes the
questionnaire fields, this command takes the raw attributes object as JSON
rather than hardcoding a preset.

Note: with the newer age-rating flow the declaration may not be readable via
the API (GETs can 404 even when the UI shows a rating). Writing still works;
verify the resulting rating in the App Store Connect UI.

Dump the current attributes, edit them (e.g. set every content field to "NONE"
for a 4+ rating), and apply:

    asc appinfo age-rating --app 6790641087 --attrs @agerating.json`,
	Example: `  asc appinfo age-rating --app 6790641087 --attrs @agerating.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, appinfoApp)
		if err != nil {
			return err
		}
		info, err := editableAppInfo(ctx, c, appID)
		if err != nil {
			return err
		}
		_, included, err := c.Get(ctx, "/v1/appInfos/"+info.ID+"?include=ageRatingDeclaration")
		if err != nil {
			return err
		}
		var declID string
		for _, inc := range included {
			if inc.Type == "ageRatingDeclarations" {
				declID = inc.ID
			}
		}
		if declID == "" {
			return fmt.Errorf("no ageRatingDeclaration found for appInfo %s", info.ID)
		}
		raw, err := valueOrFile(appinfoAttrs)
		if err != nil {
			return err
		}
		var attrs map[string]any
		if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
			return fmt.Errorf("parse --attrs JSON: %w", err)
		}
		_, err = c.Patch(ctx, "/v1/ageRatingDeclarations/"+declID, api.Body{
			Data: api.Resource{Type: "ageRatingDeclarations", ID: declID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Println("Age-rating declaration updated.")
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{appinfoShowCmd, appinfoLocalizeCmd, appinfoCategoryCmd, appinfoAgeRatingCmd} {
		sub.Flags().StringVar(&appinfoApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	appinfoLocalizeCmd.Flags().StringVar(&appinfoLocale, "locale", "ja", "locale, e.g. ja / en-US")
	appinfoLocalizeCmd.Flags().StringVar(&appinfoName, "name", "", "app name (max 30 chars)")
	appinfoLocalizeCmd.Flags().StringVar(&appinfoSubtitle, "subtitle", "", "subtitle (max 30 chars)")
	appinfoLocalizeCmd.Flags().StringVar(&appinfoPrivacy, "privacy-url", "", "privacy policy URL")

	appinfoCategoryCmd.Flags().StringVar(&appinfoPrimary, "primary", "", "primary category id (e.g. PRODUCTIVITY)")
	appinfoCategoryCmd.Flags().StringVar(&appinfoSecond, "secondary", "", "secondary category id (e.g. BUSINESS)")

	appinfoAgeRatingCmd.Flags().StringVar(&appinfoAttrs, "attrs", "", "JSON attributes object, or @file (required)")
	_ = appinfoAgeRatingCmd.MarkFlagRequired("attrs")

	appinfoCmd.AddCommand(appinfoShowCmd, appinfoLocalizeCmd, appinfoCategoryCmd, appinfoAgeRatingCmd)
	rootCmd.AddCommand(appinfoCmd)
}
