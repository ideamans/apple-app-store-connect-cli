package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var sxCmd = &cobra.Command{
	Use:   "review-submissions",
	Short: "Inspect and manage App Store review submissions",
}

var (
	sxApp          string
	sxPlatform     string
	sxID           string
	sxVersionID    string
	sxEventID      string
	sxExperimentID string
	sxCPPVersionID string
)

var sxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's review submissions",
	Example: `  asc review-submissions list --app 6790641087
  asc review-submissions list --app 6790641087 --platform IOS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sxApp)
		if err != nil {
			return err
		}
		path := "/v1/apps/" + appID + "/reviewSubmissions?limit=50"
		if sxPlatform != "" {
			path += "&filter[platform]=" + sxPlatform
		}
		subs, err := c.List(ctx, path)
		if err != nil {
			return err
		}
		if len(subs) == 0 {
			fmt.Println("No review submissions found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STATE\tPLATFORM\tSUBMITTED\tID")
		for _, s := range subs {
			submitted := s.Str("submittedDate")
			if submitted == "" {
				submitted = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Str("state"), s.Str("platform"), submitted, s.ID)
		}
		return w.Flush()
	},
}

var sxShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show a review submission and its items",
	Example: `  asc review-submissions show --id 8b7e01b6-...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		sub, _, err := c.Get(ctx, "/v1/reviewSubmissions/"+sxID)
		if err != nil {
			return err
		}
		out := map[string]any{"id": sub.ID}
		for k, v := range sub.Attributes {
			out[k] = v
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

		data, err := c.Do(ctx, http.MethodGet,
			"/v1/reviewSubmissions/"+sxID+"/items?include=appStoreVersion,appEvent,appStoreVersionExperiment,appStoreVersionExperimentV2,appCustomProductPageVersion,inAppPurchaseVersion,subscriptionVersion,subscriptionGroupVersion,backgroundAssetVersion,gameCenterAchievementVersion,gameCenterActivityVersion,gameCenterChallengeVersion,gameCenterLeaderboardVersion,gameCenterLeaderboardSetVersion&limit=200", nil)
		if err != nil {
			return err
		}
		var doc struct {
			Data     []api.Resource `json:"data"`
			Included []api.Resource `json:"included"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			return err
		}
		if len(doc.Data) == 0 {
			fmt.Println("\nNo items.")
			return nil
		}
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STATE\tCONTENT\tID")
		for i := range doc.Data {
			fmt.Fprintf(w, "%s\t%s\t%s\n", doc.Data[i].Str("state"), sxItemContent(&doc.Data[i], doc.Included), doc.Data[i].ID)
		}
		return w.Flush()
	},
}

// sxItemContent describes what a review submission item points at, using the
// included resources for a human-readable label when available.
func sxItemContent(item *api.Resource, included []api.Resource) string {
	names := []string{
		"appStoreVersion", "appEvent", "appStoreVersionExperiment", "appStoreVersionExperimentV2",
		"appCustomProductPageVersion", "inAppPurchaseVersion", "subscriptionVersion", "subscriptionGroupVersion",
		"backgroundAssetVersion", "gameCenterAchievementVersion", "gameCenterActivityVersion",
		"gameCenterChallengeVersion", "gameCenterLeaderboardVersion", "gameCenterLeaderboardSetVersion",
	}
	for _, name := range names {
		raw, ok := item.Relationships[name]
		if !ok {
			continue
		}
		var rel struct {
			Data *struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"data"`
		}
		if json.Unmarshal(raw, &rel) != nil || rel.Data == nil {
			continue
		}
		label := name + " " + rel.Data.ID
		for j := range included {
			if included[j].ID != rel.Data.ID || included[j].Type != rel.Data.Type {
				continue
			}
			for _, attr := range []string{"versionString", "referenceName", "name"} {
				if v := included[j].Str(attr); v != "" {
					label = name + " " + v + " (" + rel.Data.ID + ")"
					break
				}
			}
		}
		return label
	}
	return "-"
}

var sxCancelCmd = &cobra.Command{
	Use:     "cancel",
	Short:   "Cancel a review submission that has not completed review",
	Example: `  asc review-submissions cancel --id 8b7e01b6-...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/reviewSubmissions/"+sxID, api.Body{
			Data: api.Resource{Type: "reviewSubmissions", ID: sxID, Attributes: map[string]any{"canceled": true}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Review submission %s canceled.\n", sxID)
		return nil
	},
}

var sxRemoveItemCmd = &cobra.Command{
	Use:     "remove-item",
	Short:   "Remove an item from its review submission",
	Example: `  asc review-submissions remove-item --id <reviewSubmissionItem id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/reviewSubmissionItems/"+sxID); err != nil {
			return err
		}
		fmt.Printf("Review submission item %s removed.\n", sxID)
		return nil
	},
}

var sxAddItemCmd = &cobra.Command{
	Use:   "add-item",
	Short: "Add content to an open review submission",
	Long: `Add one piece of reviewable content to a review submission. Pass exactly one
of --version-id (appStoreVersions id), --event-id (appEvents id),
--experiment-id (appStoreVersionExperiments id, submitted as the V2
relationship) or --cpp-version-id (appCustomProductPageVersions id).`,
	Example: `  asc review-submissions add-item --id 8b7e01b6-... --version-id d3e7a1c2-...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rels := map[string]json.RawMessage{"reviewSubmission": api.Rel("reviewSubmissions", sxID)}
		count := 0
		if sxVersionID != "" {
			rels["appStoreVersion"] = api.Rel("appStoreVersions", sxVersionID)
			count++
		}
		if sxEventID != "" {
			rels["appEvent"] = api.Rel("appEvents", sxEventID)
			count++
		}
		if sxExperimentID != "" {
			rels["appStoreVersionExperimentV2"] = api.Rel("appStoreVersionExperiments", sxExperimentID)
			count++
		}
		if sxCPPVersionID != "" {
			rels["appCustomProductPageVersion"] = api.Rel("appCustomProductPageVersions", sxCPPVersionID)
			count++
		}
		if count != 1 {
			return fmt.Errorf("pass exactly one of --version-id, --event-id, --experiment-id, --cpp-version-id")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/reviewSubmissionItems", api.Body{
			Data: api.Resource{Type: "reviewSubmissionItems", Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Added item %s to review submission %s.\n", created.ID, sxID)
		return nil
	},
}

func init() {
	sxListCmd.Flags().StringVar(&sxApp, "app", "", "app id or bundle id (required)")
	sxListCmd.Flags().StringVar(&sxPlatform, "platform", "", "filter by platform: IOS, MAC_OS, TV_OS, VISION_OS")
	_ = sxListCmd.MarkFlagRequired("app")

	sxShowCmd.Flags().StringVar(&sxID, "id", "", "review submission id (required)")
	sxCancelCmd.Flags().StringVar(&sxID, "id", "", "review submission id (required)")
	sxRemoveItemCmd.Flags().StringVar(&sxID, "id", "", "review submission item id (required)")
	sxAddItemCmd.Flags().StringVar(&sxID, "id", "", "review submission id (required)")
	for _, sub := range []*cobra.Command{sxShowCmd, sxCancelCmd, sxRemoveItemCmd, sxAddItemCmd} {
		_ = sub.MarkFlagRequired("id")
	}

	sxAddItemCmd.Flags().StringVar(&sxVersionID, "version-id", "", "appStoreVersions id to submit")
	sxAddItemCmd.Flags().StringVar(&sxEventID, "event-id", "", "appEvents id to submit")
	sxAddItemCmd.Flags().StringVar(&sxExperimentID, "experiment-id", "", "appStoreVersionExperiments id to submit")
	sxAddItemCmd.Flags().StringVar(&sxCPPVersionID, "cpp-version-id", "", "appCustomProductPageVersions id to submit")

	sxCmd.AddCommand(sxListCmd, sxShowCmd, sxCancelCmd, sxRemoveItemCmd, sxAddItemCmd)
	rootCmd.AddCommand(sxCmd)
}
