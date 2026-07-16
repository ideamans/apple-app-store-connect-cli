package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var nominationsCmd = &cobra.Command{
	Use:   "nominations",
	Short: "Work with featuring nominations (suggest content to the App Store editorial team)",
	Long: `Manage featuring nominations: pitches to the App Store editorial team for
featuring an app launch, app enhancements, or new content.

A nomination is created as a DRAFT and submitted by setting submitted=true
(pass --submit on create, or run: nominations update --id <id> --submit).
Dates are ISO 8601 date-times, e.g. 2026-08-01T00:00:00Z.`,
}

var eulaCmd = &cobra.Command{
	Use:   "eula",
	Short: "Work with the app's custom end user license agreement",
	Long: `Manage the app's custom EULA (endUserLicenseAgreement). Apps without a custom
EULA use Apple's standard EULA. A custom EULA applies to an explicit list of
territories.`,
}

var (
	nomApp          string
	nomID           string
	nomState        string
	nomName         string
	nomType         string
	nomDescription  string
	nomPublishStart string
	nomPublishEnd   string
	nomNotes        string
	nomSupplemental []string
	nomLocales      []string
	nomDeviceFams   []string
	nomRelatedApps  []string
	nomLaunchSelect bool
	nomSubmit       bool
	nomArchive      bool
	nomEulaText     string
	nomEulaTerrs    []string
)

// nomRelMany builds a to-many relationship value: {"data":[{"type":..,"id":..},...]}.
func nomRelMany(typ string, ids []string) json.RawMessage {
	items := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]string{"type": typ, "id": id})
	}
	b, _ := json.Marshal(map[string]any{"data": items})
	return b
}

// nomResolveApps resolves a list of app ids or bundle ids to app ids.
func nomResolveApps(ctx context.Context, c *api.Client, refs []string) ([]string, error) {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		id, err := resolveAppID(ctx, c, strings.TrimSpace(ref))
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

var nomListCmd = &cobra.Command{
	Use:   "list",
	Short: "List nominations, optionally filtered to one app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appFilter := ""
		if nomApp != "" {
			appID, err := resolveAppID(ctx, c, nomApp)
			if err != nil {
				return err
			}
			appFilter = "&filter[relatedApps]=" + appID
		}
		// filter[state] is required by the API, and despite the spec declaring it
		// an array the live API accepts only a single value — query per state.
		var noms []api.Resource
		for _, state := range strings.Split(nomState, ",") {
			state = strings.TrimSpace(state)
			if state == "" {
				continue
			}
			page, err := c.List(ctx, "/v1/nominations?filter[state]="+state+appFilter+"&limit=50")
			if err != nil {
				return err
			}
			noms = append(noms, page...)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATE\tPUBLISH_START\tID")
		for _, n := range noms {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				n.Str("name"), n.Str("type"), n.Str("state"), n.Str("publishStartDate"), n.ID)
		}
		return w.Flush()
	},
}

var nomShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a nomination",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		nom, _, err := c.Get(cmd.Context(), "/v1/nominations/"+nomID)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(nom, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var nomCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a nomination (a draft unless --submit is passed)",
	Example: `  asc nominations create --app 6790641087 --type NEW_CONTENT \
    --name "Summer update featuring" --description @nomination.txt \
    --publish-start 2026-08-01T00:00:00Z --submit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, nomApp)
		if err != nil {
			return err
		}
		desc, err := valueOrFile(nomDescription)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"name":             nomName,
			"type":             nomType,
			"description":      desc,
			"submitted":        nomSubmit,
			"publishStartDate": nomPublishStart,
		}
		if nomPublishEnd != "" {
			attrs["publishEndDate"] = nomPublishEnd
		}
		if len(nomSupplemental) > 0 {
			attrs["supplementalMaterialsUris"] = nomSupplemental
		}
		if cmd.Flags().Changed("launch-in-select-markets-first") {
			attrs["launchInSelectMarketsFirst"] = nomLaunchSelect
		}
		if len(nomLocales) > 0 {
			attrs["locales"] = nomLocales
		}
		if len(nomDeviceFams) > 0 {
			attrs["deviceFamilies"] = nomDeviceFams
		}
		if nomNotes != "" {
			attrs["notes"] = nomNotes
		}
		related, err := nomResolveApps(ctx, c, nomRelatedApps)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/nominations", api.Body{
			Data: api.Resource{
				Type:       "nominations",
				Attributes: attrs,
				Relationships: map[string]json.RawMessage{
					"relatedApps": nomRelMany("apps", append([]string{appID}, related...)),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created nomination %q (%s, state=%s).\n", nomName, created.ID, created.Str("state"))
		return nil
	},
}

var nomUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a nomination; --submit submits it, --archive archives it",
	Example: `  asc nominations update --id <id> --submit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = nomName
		}
		if cmd.Flags().Changed("type") {
			attrs["type"] = nomType
		}
		if cmd.Flags().Changed("description") {
			desc, err := valueOrFile(nomDescription)
			if err != nil {
				return err
			}
			attrs["description"] = desc
		}
		if cmd.Flags().Changed("publish-start") {
			attrs["publishStartDate"] = nomPublishStart
		}
		if cmd.Flags().Changed("publish-end") {
			attrs["publishEndDate"] = nomPublishEnd
		}
		if cmd.Flags().Changed("notes") {
			attrs["notes"] = nomNotes
		}
		if cmd.Flags().Changed("supplemental-materials-uris") {
			attrs["supplementalMaterialsUris"] = nomSupplemental
		}
		if cmd.Flags().Changed("launch-in-select-markets-first") {
			attrs["launchInSelectMarketsFirst"] = nomLaunchSelect
		}
		if cmd.Flags().Changed("locales") {
			attrs["locales"] = nomLocales
		}
		if cmd.Flags().Changed("device-families") {
			attrs["deviceFamilies"] = nomDeviceFams
		}
		if nomSubmit {
			attrs["submitted"] = true
		}
		if nomArchive {
			attrs["archived"] = true
		}
		var rels map[string]json.RawMessage
		if cmd.Flags().Changed("related-apps") {
			related, err := nomResolveApps(cmd.Context(), c, nomRelatedApps)
			if err != nil {
				return err
			}
			// Note: this replaces the nomination's whole related-apps list.
			rels = map[string]json.RawMessage{"relatedApps": nomRelMany("apps", related)}
		}
		if len(attrs) == 0 && rels == nil {
			return fmt.Errorf("nothing to set: pass attribute flags, --submit or --archive")
		}
		updated, err := c.Patch(cmd.Context(), "/v1/nominations/"+nomID, api.Body{
			Data: api.Resource{Type: "nominations", ID: nomID, Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Nomination updated (state=%s).\n", updated.Str("state"))
		return nil
	},
}

var nomDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a nomination",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/nominations/"+nomID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var nomEulaShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the app's custom EULA, if any",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, nomApp)
		if err != nil {
			return err
		}
		eula, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/endUserLicenseAgreement")
		if err != nil {
			return err
		}
		// The endpoint answers 200 with "data": null when there is no custom EULA.
		if eula == nil || eula.ID == "" {
			fmt.Println("No custom EULA; the App Store standard EULA applies.")
			return nil
		}
		terrs, err := c.List(ctx, "/v1/endUserLicenseAgreements/"+eula.ID+"/territories?limit=200")
		if err != nil {
			return err
		}
		ids := make([]string, 0, len(terrs))
		for _, t := range terrs {
			ids = append(ids, t.ID)
		}
		fmt.Printf("endUserLicenseAgreement %s\n", eula.ID)
		fmt.Printf("territories: %s\n", strings.Join(ids, ","))
		fmt.Println("---")
		fmt.Println(eula.Str("agreementText"))
		return nil
	},
}

var nomEulaSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Create or replace the app's custom EULA",
	Example: `  asc eula set --app 6790641087 --text @eula.txt --territories JPN,USA`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, nomApp)
		if err != nil {
			return err
		}
		text, err := valueOrFile(nomEulaText)
		if err != nil {
			return err
		}
		existing, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/endUserLicenseAgreement")
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != "" {
			rels := map[string]json.RawMessage{}
			if len(nomEulaTerrs) > 0 {
				rels["territories"] = nomRelMany("territories", nomEulaTerrs)
			}
			_, err = c.Patch(ctx, "/v1/endUserLicenseAgreements/"+existing.ID, api.Body{
				Data: api.Resource{
					Type:          "endUserLicenseAgreements",
					ID:            existing.ID,
					Attributes:    map[string]any{"agreementText": text},
					Relationships: rels,
				},
			})
			if err != nil {
				return err
			}
			fmt.Println("EULA updated.")
			return nil
		}
		if len(nomEulaTerrs) == 0 {
			return fmt.Errorf("--territories is required when creating a EULA (e.g. --territories JPN,USA)")
		}
		created, err := c.Post(ctx, "/v1/endUserLicenseAgreements", api.Body{
			Data: api.Resource{
				Type:       "endUserLicenseAgreements",
				Attributes: map[string]any{"agreementText": text},
				Relationships: map[string]json.RawMessage{
					"app":         api.Rel("apps", appID),
					"territories": nomRelMany("territories", nomEulaTerrs),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("EULA created (%s).\n", created.ID)
		return nil
	},
}

var nomEulaDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the app's custom EULA (reverting to Apple's standard EULA)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, nomApp)
		if err != nil {
			return err
		}
		existing, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/endUserLicenseAgreement")
		if err != nil {
			return err
		}
		if existing == nil || existing.ID == "" {
			fmt.Println("No custom EULA to delete.")
			return nil
		}
		if err := c.Delete(ctx, "/v1/endUserLicenseAgreements/"+existing.ID); err != nil {
			return err
		}
		fmt.Println("Deleted; the App Store standard EULA now applies.")
		return nil
	},
}

func init() {
	// nominations
	nomListCmd.Flags().StringVar(&nomApp, "app", "", "limit to nominations related to this app id or bundle id")
	nomListCmd.Flags().StringVar(&nomState, "state", "DRAFT,SUBMITTED,ARCHIVED", "comma-separated states: DRAFT, SUBMITTED, ARCHIVED")

	nomCreateCmd.Flags().StringVar(&nomApp, "app", "", "app id or bundle id (required)")
	_ = nomCreateCmd.MarkFlagRequired("app")

	for _, sub := range []*cobra.Command{nomShowCmd, nomUpdateCmd, nomDeleteCmd} {
		sub.Flags().StringVar(&nomID, "id", "", "nomination id (required)")
		_ = sub.MarkFlagRequired("id")
	}

	for _, sub := range []*cobra.Command{nomCreateCmd, nomUpdateCmd} {
		sub.Flags().StringVar(&nomName, "name", "", "nomination name")
		sub.Flags().StringVar(&nomType, "type", "", "type: NEW_CONTENT, APP_LAUNCH, APP_ENHANCEMENTS")
		sub.Flags().StringVar(&nomDescription, "description", "", "pitch to the editorial team (@file allowed)")
		sub.Flags().StringVar(&nomPublishStart, "publish-start", "", "publish start date (ISO 8601)")
		sub.Flags().StringVar(&nomPublishEnd, "publish-end", "", "publish end date (ISO 8601)")
		sub.Flags().StringVar(&nomNotes, "notes", "", "additional notes")
		sub.Flags().StringSliceVar(&nomSupplemental, "supplemental-materials-uris", nil, "URLs of supplemental materials (comma-separated or repeatable)")
		sub.Flags().StringSliceVar(&nomLocales, "locales", nil, "locales the content is available in")
		sub.Flags().StringSliceVar(&nomDeviceFams, "device-families", nil, "device families: IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, VISION")
		sub.Flags().StringSliceVar(&nomRelatedApps, "related-apps", nil, "additional related app ids")
		sub.Flags().BoolVar(&nomLaunchSelect, "launch-in-select-markets-first", false, "the app launches in select markets first")
		sub.Flags().BoolVar(&nomSubmit, "submit", false, "submit the nomination (submitted=true)")
	}
	_ = nomCreateCmd.MarkFlagRequired("name")
	_ = nomCreateCmd.MarkFlagRequired("type")
	_ = nomCreateCmd.MarkFlagRequired("description")
	_ = nomCreateCmd.MarkFlagRequired("publish-start")
	nomUpdateCmd.Flags().BoolVar(&nomArchive, "archive", false, "archive the nomination (archived=true)")

	nominationsCmd.AddCommand(nomListCmd, nomShowCmd, nomCreateCmd, nomUpdateCmd, nomDeleteCmd)

	// eula
	for _, sub := range []*cobra.Command{nomEulaShowCmd, nomEulaSetCmd, nomEulaDeleteCmd} {
		sub.Flags().StringVar(&nomApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	nomEulaSetCmd.Flags().StringVar(&nomEulaText, "text", "", "agreement text, or @file (required)")
	nomEulaSetCmd.Flags().StringSliceVar(&nomEulaTerrs, "territories", nil, "territory ids the EULA applies to, e.g. JPN,USA (required on create)")
	_ = nomEulaSetCmd.MarkFlagRequired("text")

	eulaCmd.AddCommand(nomEulaShowCmd, nomEulaSetCmd, nomEulaDeleteCmd)

	rootCmd.AddCommand(nominationsCmd, eulaCmd)
}
