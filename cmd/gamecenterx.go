package cmd

// Additional Game Center commands: Game Center app versions, activity and
// challenge CRUD, leaderboard set images (v2) and member localizations,
// matchmaking configuration, and server-to-server score / achievement
// submissions. Everything attaches under the existing `gamecenter` command
// defined in gamecenter.go.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	gczApp                 string
	gczVersionID           string
	gczAppVersionID        string
	gczCompatVersionID     string
	gczActivityID          string
	gczChallengeID         string
	gczSetID               string
	gczLeaderboardID       string
	gczLocalizationID      string
	gczImageID             string
	gczFile                string
	gczLocale              string
	gczName                string
	gczDesc                string
	gczRefName             string
	gczVendorID            string
	gczPlayStyle           string
	gczMinPlayers          int
	gczMaxPlayers          int
	gczSupportsPartyCode   bool
	gczArchived            bool
	gczChallengeType       string
	gczRepeatable          bool
	gczID                  string
	gczRuleSetID           string
	gczExperimentRuleSetID string
	gczClassicBundleIDs    []string
	gczRuleLanguageVersion int
	gczRuleType            string
	gczExpression          string
	gczWeight              float64
	gczBody                string
	gczBundleID            string
	gczScopedPlayerID      string
	gczScore               string
	gczContext             string
	gczSubmittedDate       string
	gczChallengeIDs        []string
	gczPreReleased         bool
	gczPercentage          int
)

// gczFloat formats a numeric attribute without truncating decimals.
func gczFloat(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(float64); ok {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// gczGameCenterAppVersion resolves an appStoreVersion id to its
// gameCenterAppVersion resource (nil when none exists yet).
func gczGameCenterAppVersion(ctx context.Context, c *api.Client, appStoreVersionID string) (*api.Resource, error) {
	return c.GetOptional(ctx, "/v1/appStoreVersions/"+appStoreVersionID+"/gameCenterAppVersion")
}

// gczLinkageMutation POSTs or DELETEs a to-many relationship linkage, honoring
// --dry-run (raw c.Do calls bypass the client's built-in guard).
func gczLinkageMutation(ctx context.Context, c *api.Client, method, path, typ string, ids []string) error {
	items := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]string{"type": typ, "id": id})
	}
	payload, err := json.Marshal(map[string]any{"data": items})
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(os.Stderr, "DRY-RUN %s %s\n%s\n", method, path, payload)
		return nil
	}
	_, err = c.Do(ctx, method, path, bytes.NewReader(payload))
	return err
}

// --- gamecenter app-versions ----------------------------------------------------

var gczAppVersionsCmd = &cobra.Command{
	Use:   "app-versions",
	Short: "Manage Game Center app versions (per-App-Store-version enablement)",
	Long: `Manage gameCenterAppVersions: the resources that enable Game Center for a
specific App Store version and declare multiplayer compatibility between
versions (via the compatibilityVersions relationship).`,
}

var gczAppVersionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's Game Center app versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gczApp)
		if err != nil {
			return err
		}
		vers, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterAppVersions?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ENABLED\tAPP_STORE_VERSION\tPLATFORM\tID")
		for i := range vers {
			v := &vers[i]
			asv, err := c.GetOptional(ctx, "/v1/gameCenterAppVersions/"+v.ID+"/appStoreVersion")
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", gcxBool(v, "enabled"), asv.Str("versionString"), asv.Str("platform"), v.ID)
		}
		return w.Flush()
	},
}

// gczSetAppVersionEnabled creates the gameCenterAppVersion for an
// appStoreVersion when needed, then patches its enabled attribute.
func gczSetAppVersionEnabled(cmd *cobra.Command, enabled bool) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	existing, err := gczGameCenterAppVersion(ctx, c, gczVersionID)
	if err != nil {
		return err
	}
	id := ""
	if existing != nil && existing.ID != "" {
		id = existing.ID
	} else {
		if !enabled {
			return fmt.Errorf("appStoreVersion %s has no gameCenterAppVersion; nothing to disable", gczVersionID)
		}
		created, err := c.Post(ctx, "/v1/gameCenterAppVersions", api.Body{
			Data: api.Resource{
				Type:          "gameCenterAppVersions",
				Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", gczVersionID)},
			},
		})
		if err != nil {
			return err
		}
		id = created.ID
	}
	_, err = c.Patch(ctx, "/v1/gameCenterAppVersions/"+id, api.Body{
		Data: api.Resource{Type: "gameCenterAppVersions", ID: id, Attributes: map[string]any{"enabled": enabled}},
	})
	if err != nil {
		return err
	}
	fmt.Printf("Game Center app version %s enabled=%t (appStoreVersion %s).\n", id, enabled, gczVersionID)
	return nil
}

var gczAppVersionsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Game Center for an App Store version",
	Long:  "Create the gameCenterAppVersion for the given appStoreVersion if needed and set enabled=true.",
	RunE:  func(cmd *cobra.Command, args []string) error { return gczSetAppVersionEnabled(cmd, true) },
}

var gczAppVersionsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Game Center for an App Store version",
	RunE:  func(cmd *cobra.Command, args []string) error { return gczSetAppVersionEnabled(cmd, false) },
}

var gczAppVersionsLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Mark two Game Center app versions as multiplayer-compatible",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/gameCenterAppVersions/" + gczAppVersionID + "/relationships/compatibilityVersions"
		if err := gczLinkageMutation(cmd.Context(), c, http.MethodPost, path, "gameCenterAppVersions", []string{gczCompatVersionID}); err != nil {
			return err
		}
		fmt.Printf("Linked compatibility version %s to %s.\n", gczCompatVersionID, gczAppVersionID)
		return nil
	},
}

var gczAppVersionsUnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Remove a multiplayer compatibility link between two app versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/gameCenterAppVersions/" + gczAppVersionID + "/relationships/compatibilityVersions"
		if err := gczLinkageMutation(cmd.Context(), c, http.MethodDelete, path, "gameCenterAppVersions", []string{gczCompatVersionID}); err != nil {
			return err
		}
		fmt.Printf("Unlinked compatibility version %s from %s.\n", gczCompatVersionID, gczAppVersionID)
		return nil
	},
}

// --- gamecenter activities (create/update/delete/localize/images) ----------------

var gczActivitiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an activity (an initial version is created inline)",
	Example: `  asc gamecenter activities create --app 6790641087 \
    --reference-name "Daily Race" --vendor-identifier com.example.daily_race \
    --play-style SYNCHRONOUS --min-players 2 --max-players 8`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gczApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"referenceName":    gczRefName,
			"vendorIdentifier": gczVendorID,
		}
		if cmd.Flags().Changed("play-style") {
			attrs["playStyle"] = gczPlayStyle
		}
		if cmd.Flags().Changed("min-players") {
			attrs["minimumPlayersCount"] = gczMinPlayers
		}
		if cmd.Flags().Changed("max-players") {
			attrs["maximumPlayersCount"] = gczMaxPlayers
		}
		if cmd.Flags().Changed("supports-party-code") {
			attrs["supportsPartyCode"] = gczSupportsPartyCode
		}
		const verLID = "${actVer1}"
		created, err := c.Post(ctx, "/v1/gameCenterActivities", api.Body{
			Data: api.Resource{
				Type:       "gameCenterActivities",
				Attributes: attrs,
				Relationships: map[string]json.RawMessage{
					"gameCenterDetail": api.Rel("gameCenterDetails", detailID),
					"versions":         gcxToManyRel("gameCenterActivityVersions", verLID),
				},
			},
			Included: []api.Resource{{Type: "gameCenterActivityVersions", ID: verLID}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created activity %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczActivitiesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an activity's attributes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("reference-name") {
			attrs["referenceName"] = gczRefName
		}
		if cmd.Flags().Changed("play-style") {
			attrs["playStyle"] = gczPlayStyle
		}
		if cmd.Flags().Changed("min-players") {
			attrs["minimumPlayersCount"] = gczMinPlayers
		}
		if cmd.Flags().Changed("max-players") {
			attrs["maximumPlayersCount"] = gczMaxPlayers
		}
		if cmd.Flags().Changed("supports-party-code") {
			attrs["supportsPartyCode"] = gczSupportsPartyCode
		}
		if cmd.Flags().Changed("archived") {
			attrs["archived"] = gczArchived
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --reference-name, --play-style, --min-players, --max-players, --supports-party-code and/or --archived")
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterActivities/"+gczActivityID, api.Body{
			Data: api.Resource{Type: "gameCenterActivities", ID: gczActivityID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Activity %s updated.\n", gczActivityID)
		return nil
	},
}

var gczActivitiesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an activity",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterActivities/"+gczActivityID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gczActivitiesLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set an activity's localized name and description",
	Long: `Set an activity's localized name and description. Localizations belong to
the activity's latest version. Creating a new localization requires --name;
updates accept any subset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, err := gcxLatestVersion(ctx, c, "/v1/gameCenterActivities/"+gczActivityID+"/versions?limit=50")
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = gczName
		}
		if cmd.Flags().Changed("description") {
			v, err := valueOrFile(gczDesc)
			if err != nil {
				return err
			}
			attrs["description"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name and/or --description")
		}
		locs, err := c.List(ctx, "/v1/gameCenterActivityVersions/"+ver.ID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gczLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/gameCenterActivityLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "gameCenterActivityLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			if _, ok := attrs["name"]; !ok {
				return fmt.Errorf("creating a new %s localization requires --name", gczLocale)
			}
			attrs["locale"] = gczLocale
			_, err = c.Post(ctx, "/v1/gameCenterActivityLocalizations", api.Body{
				Data: api.Resource{
					Type:          "gameCenterActivityLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"version": api.Rel("gameCenterActivityVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Activity %s localization (%s) updated on version %s.\n", gczActivityID, gczLocale, ver.ID)
		return nil
	},
}

var gczActivitiesUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload the image for an activity localization",
	Long: `Upload an activity image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the activity's latest version:
GET /v1/gameCenterActivities/{id}/versions then
GET /v1/gameCenterActivityVersions/{id}/localizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := gcxUploadImage(cmd.Context(), c,
			"/v1/gameCenterActivityImages", "gameCenterActivityImages",
			"gameCenterActivityLocalizations", gczLocalizationID, gczFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded activity image -> %s\n", id)
		return nil
	},
}

var gczActivitiesDeleteImageCmd = &cobra.Command{
	Use:   "delete-image",
	Short: "Delete an activity image",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterActivityImages/"+gczImageID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// --- gamecenter challenges (create/update/delete/localize/images) -----------------

var gczChallengesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a challenge (an initial version is created inline)",
	Example: `  asc gamecenter challenges create --app 6790641087 \
    --reference-name "Weekly Sprint" --vendor-identifier com.example.weekly_sprint \
    --leaderboard-id <gameCenterLeaderboard id> --repeatable`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gczApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"referenceName":    gczRefName,
			"vendorIdentifier": gczVendorID,
			"challengeType":    gczChallengeType,
		}
		if cmd.Flags().Changed("repeatable") {
			attrs["repeatable"] = gczRepeatable
		}
		const verLID = "${chalVer1}"
		rels := map[string]json.RawMessage{
			"gameCenterDetail": api.Rel("gameCenterDetails", detailID),
			"versions":         gcxToManyRel("gameCenterChallengeVersions", verLID),
		}
		if gczLeaderboardID != "" {
			rels["leaderboardV2"] = api.Rel("gameCenterLeaderboards", gczLeaderboardID)
		}
		created, err := c.Post(ctx, "/v1/gameCenterChallenges", api.Body{
			Data: api.Resource{
				Type:          "gameCenterChallenges",
				Attributes:    attrs,
				Relationships: rels,
			},
			Included: []api.Resource{{Type: "gameCenterChallengeVersions", ID: verLID}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created challenge %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczChallengesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a challenge's attributes or its leaderboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("reference-name") {
			attrs["referenceName"] = gczRefName
		}
		if cmd.Flags().Changed("repeatable") {
			attrs["repeatable"] = gczRepeatable
		}
		if cmd.Flags().Changed("archived") {
			attrs["archived"] = gczArchived
		}
		var rels map[string]json.RawMessage
		if cmd.Flags().Changed("leaderboard-id") {
			rels = map[string]json.RawMessage{"leaderboardV2": api.Rel("gameCenterLeaderboards", gczLeaderboardID)}
		}
		if len(attrs) == 0 && rels == nil {
			return fmt.Errorf("nothing to set: pass --reference-name, --repeatable, --archived and/or --leaderboard-id")
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterChallenges/"+gczChallengeID, api.Body{
			Data: api.Resource{Type: "gameCenterChallenges", ID: gczChallengeID, Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Challenge %s updated.\n", gczChallengeID)
		return nil
	},
}

var gczChallengesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a challenge",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterChallenges/"+gczChallengeID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gczChallengesLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set a challenge's localized name and description",
	Long: `Set a challenge's localized name and description. Localizations belong to
the challenge's latest version. Creating a new localization requires --name;
updates accept any subset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, err := gcxLatestVersion(ctx, c, "/v1/gameCenterChallenges/"+gczChallengeID+"/versions?limit=50")
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = gczName
		}
		if cmd.Flags().Changed("description") {
			v, err := valueOrFile(gczDesc)
			if err != nil {
				return err
			}
			attrs["description"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name and/or --description")
		}
		locs, err := c.List(ctx, "/v1/gameCenterChallengeVersions/"+ver.ID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gczLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/gameCenterChallengeLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "gameCenterChallengeLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			if _, ok := attrs["name"]; !ok {
				return fmt.Errorf("creating a new %s localization requires --name", gczLocale)
			}
			attrs["locale"] = gczLocale
			_, err = c.Post(ctx, "/v1/gameCenterChallengeLocalizations", api.Body{
				Data: api.Resource{
					Type:          "gameCenterChallengeLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"version": api.Rel("gameCenterChallengeVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Challenge %s localization (%s) updated on version %s.\n", gczChallengeID, gczLocale, ver.ID)
		return nil
	},
}

var gczChallengesUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload the image for a challenge localization",
	Long: `Upload a challenge image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the challenge's latest version:
GET /v1/gameCenterChallenges/{id}/versions then
GET /v1/gameCenterChallengeVersions/{id}/localizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := gcxUploadImage(cmd.Context(), c,
			"/v1/gameCenterChallengeImages", "gameCenterChallengeImages",
			"gameCenterChallengeLocalizations", gczLocalizationID, gczFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded challenge image -> %s\n", id)
		return nil
	},
}

var gczChallengesDeleteImageCmd = &cobra.Command{
	Use:   "delete-image",
	Short: "Delete a challenge image",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterChallengeImages/"+gczImageID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// --- gamecenter leaderboard-sets: images (v2) and member localizations ------------

var gczSetUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload the image for a leaderboard set localization",
	Long: `Upload a leaderboard set image (reserve → upload → commit, v2) and attach
it to a set localization. Find the localization id via the set's latest
version: GET /v2/gameCenterLeaderboardSets/{id}/versions then
GET /v2/gameCenterLeaderboardSetVersions/{id}/localizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := gcxUploadImage(cmd.Context(), c,
			"/v2/gameCenterLeaderboardSetImages", "gameCenterLeaderboardSetImages",
			"gameCenterLeaderboardSetLocalizations", gczLocalizationID, gczFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded leaderboard set image -> %s\n", id)
		return nil
	},
}

var gczSetDeleteImageCmd = &cobra.Command{
	Use:   "delete-image",
	Short: "Delete a leaderboard set image",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v2/gameCenterLeaderboardSetImages/"+gczImageID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gczMemberLocsCmd = &cobra.Command{
	Use:   "member-localizations",
	Short: "Manage per-set localized names of member leaderboards",
	Long: `A leaderboard set member localization overrides a leaderboard's display
name within a specific leaderboard set. The API addresses them by the
(leaderboard set, leaderboard) pair, so both ids are always required.`,
}

var gczMemberLocsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List member localizations for a (set, leaderboard) pair",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		locs, err := c.List(cmd.Context(),
			"/v1/gameCenterLeaderboardSetMemberLocalizations?filter[gameCenterLeaderboardSet]="+gczSetID+
				"&filter[gameCenterLeaderboard]="+gczLeaderboardID+"&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "LOCALE\tNAME\tID")
		for _, l := range locs {
			fmt.Fprintf(w, "%s\t%s\t%s\n", l.Str("locale"), l.Str("name"), l.ID)
		}
		return w.Flush()
	},
}

var gczMemberLocsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update a member localization for a locale",
	Example: `  asc gamecenter leaderboard-sets member-localizations set \
    --set-id <set id> --leaderboard-id <leaderboard id> --locale ja --name "週間スコア"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		locs, err := c.List(ctx,
			"/v1/gameCenterLeaderboardSetMemberLocalizations?filter[gameCenterLeaderboardSet]="+gczSetID+
				"&filter[gameCenterLeaderboard]="+gczLeaderboardID+"&limit=200")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gczLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/gameCenterLeaderboardSetMemberLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{
					Type:       "gameCenterLeaderboardSetMemberLocalizations",
					ID:         existing.ID,
					Attributes: map[string]any{"name": gczName},
				},
			})
		} else {
			_, err = c.Post(ctx, "/v1/gameCenterLeaderboardSetMemberLocalizations", api.Body{
				Data: api.Resource{
					Type:       "gameCenterLeaderboardSetMemberLocalizations",
					Attributes: map[string]any{"locale": gczLocale, "name": gczName},
					Relationships: map[string]json.RawMessage{
						"gameCenterLeaderboardSet": api.Rel("gameCenterLeaderboardSets", gczSetID),
						"gameCenterLeaderboard":    api.Rel("gameCenterLeaderboards", gczLeaderboardID),
					},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Member localization (%s) set for leaderboard %s in set %s.\n", gczLocale, gczLeaderboardID, gczSetID)
		return nil
	},
}

// --- gamecenter matchmaking --------------------------------------------------------

var gczMatchmakingCmd = &cobra.Command{
	Use:   "matchmaking",
	Short: "Manage Game Center matchmaking (queues, rule sets, rules, teams)",
}

// queues

var gczQueuesCmd = &cobra.Command{
	Use:     "queues",
	Aliases: []string{"queue"},
	Short:   "Manage matchmaking queues",
}

var gczQueuesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List matchmaking queues",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		queues, err := c.List(cmd.Context(), "/v1/gameCenterMatchmakingQueues?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tID")
		for _, q := range queues {
			fmt.Fprintf(w, "%s\t%s\n", q.Str("referenceName"), q.ID)
		}
		return w.Flush()
	},
}

var gczQueuesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a matchmaking queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{"referenceName": gczRefName}
		if cmd.Flags().Changed("classic-bundle-ids") {
			attrs["classicMatchmakingBundleIds"] = gczClassicBundleIDs
		}
		rels := map[string]json.RawMessage{"ruleSet": api.Rel("gameCenterMatchmakingRuleSets", gczRuleSetID)}
		if gczExperimentRuleSetID != "" {
			rels["experimentRuleSet"] = api.Rel("gameCenterMatchmakingRuleSets", gczExperimentRuleSetID)
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterMatchmakingQueues", api.Body{
			Data: api.Resource{Type: "gameCenterMatchmakingQueues", Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created queue %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczQueuesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a matchmaking queue (rule sets, classic bundle ids)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("classic-bundle-ids") {
			attrs["classicMatchmakingBundleIds"] = gczClassicBundleIDs
		}
		rels := map[string]json.RawMessage{}
		if cmd.Flags().Changed("rule-set-id") {
			rels["ruleSet"] = api.Rel("gameCenterMatchmakingRuleSets", gczRuleSetID)
		}
		if cmd.Flags().Changed("experiment-rule-set-id") {
			rels["experimentRuleSet"] = api.Rel("gameCenterMatchmakingRuleSets", gczExperimentRuleSetID)
		}
		if len(attrs) == 0 && len(rels) == 0 {
			return fmt.Errorf("nothing to set: pass --rule-set-id, --experiment-rule-set-id and/or --classic-bundle-ids")
		}
		if len(rels) == 0 {
			rels = nil
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterMatchmakingQueues/"+gczID, api.Body{
			Data: api.Resource{Type: "gameCenterMatchmakingQueues", ID: gczID, Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Queue %s updated.\n", gczID)
		return nil
	},
}

var gczQueuesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a matchmaking queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterMatchmakingQueues/"+gczID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// rule sets

var gczRuleSetsCmd = &cobra.Command{
	Use:     "rule-sets",
	Aliases: []string{"rule-set"},
	Short:   "Manage matchmaking rule sets",
}

var gczRuleSetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List matchmaking rule sets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		sets, err := c.List(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tLANG_VER\tMIN\tMAX\tID")
		for i := range sets {
			s := &sets[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.Str("referenceName"), gcxNum(s, "ruleLanguageVersion"), gcxNum(s, "minPlayers"), gcxNum(s, "maxPlayers"), s.ID)
		}
		return w.Flush()
	},
}

var gczRuleSetsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a matchmaking rule set",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets", api.Body{
			Data: api.Resource{
				Type: "gameCenterMatchmakingRuleSets",
				Attributes: map[string]any{
					"referenceName":       gczRefName,
					"ruleLanguageVersion": gczRuleLanguageVersion,
					"minPlayers":          gczMinPlayers,
					"maxPlayers":          gczMaxPlayers,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created rule set %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczRuleSetsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a matchmaking rule set's player counts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("min-players") {
			attrs["minPlayers"] = gczMinPlayers
		}
		if cmd.Flags().Changed("max-players") {
			attrs["maxPlayers"] = gczMaxPlayers
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --min-players and/or --max-players")
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets/"+gczID, api.Body{
			Data: api.Resource{Type: "gameCenterMatchmakingRuleSets", ID: gczID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Rule set %s updated.\n", gczID)
		return nil
	},
}

var gczRuleSetsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a matchmaking rule set",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets/"+gczID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gczRuleSetsTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test a rule set against sample matchmaking requests",
	Long: `POST /v1/gameCenterMatchmakingRuleSetTests. The request body is deeply
nested (a matchmakingRuleSet relationship, a matchmakingRequests to-many
relationship, and included gameCenterMatchmakingTestRequests /
gameCenterMatchmakingTestPlayerProperties inline resources), so pass the
complete JSON:API request document via --body (inline JSON or @file):

  {
    "data": {
      "type": "gameCenterMatchmakingRuleSetTests",
      "relationships": {
        "matchmakingRuleSet": {"data": {"type": "gameCenterMatchmakingRuleSets", "id": "..."}},
        "matchmakingRequests": {"data": [{"type": "gameCenterMatchmakingTestRequests", "id": "${req1}"}]}
      }
    },
    "included": [
      {"type": "gameCenterMatchmakingTestRequests", "id": "${req1}",
       "attributes": {"requestName": "r1", "secondsInQueue": 0, "bundleId": "com.example.app",
                      "platform": "IOS", "appVersion": "1.0", "playerCount": 2}}
    ]
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		body, err := valueOrFile(gczBody)
		if err != nil {
			return err
		}
		if !json.Valid([]byte(body)) {
			return fmt.Errorf("--body is not valid JSON")
		}
		const path = "/v1/gameCenterMatchmakingRuleSetTests"
		if dryRun {
			fmt.Fprintf(os.Stderr, "DRY-RUN POST %s\n%s\n", path, body)
			return nil
		}
		data, err := c.Do(cmd.Context(), http.MethodPost, path, bytes.NewReader([]byte(body)))
		if err != nil {
			return err
		}
		var pretty bytes.Buffer
		if json.Indent(&pretty, data, "", "  ") == nil {
			fmt.Println(pretty.String())
		} else {
			fmt.Println(string(data))
		}
		return nil
	},
}

// rules

var gczRulesCmd = &cobra.Command{
	Use:     "rules",
	Aliases: []string{"rule"},
	Short:   "Manage matchmaking rules within a rule set",
}

var gczRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the rules of a rule set",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		rules, err := c.List(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets/"+gczRuleSetID+"/rules?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tTYPE\tWEIGHT\tID")
		for i := range rules {
			r := &rules[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Str("referenceName"), r.Str("type"), gczFloat(r, "weight"), r.ID)
		}
		return w.Flush()
	},
}

var gczRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a matchmaking rule",
	Example: `  asc gamecenter matchmaking rules create --rule-set-id <id> \
    --reference-name skill --type MATCH --description "Match by skill" \
    --expression 'abs(requests[0].properties.skill - requests[1].properties.skill) < 10' --weight 0.5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		expr, err := valueOrFile(gczExpression)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"referenceName": gczRefName,
			"description":   gczDesc,
			"type":          gczRuleType,
			"expression":    expr,
		}
		if cmd.Flags().Changed("weight") {
			attrs["weight"] = gczWeight
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterMatchmakingRules", api.Body{
			Data: api.Resource{
				Type:          "gameCenterMatchmakingRules",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"ruleSet": api.Rel("gameCenterMatchmakingRuleSets", gczRuleSetID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created rule %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczRulesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a matchmaking rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("description") {
			attrs["description"] = gczDesc
		}
		if cmd.Flags().Changed("expression") {
			expr, err := valueOrFile(gczExpression)
			if err != nil {
				return err
			}
			attrs["expression"] = expr
		}
		if cmd.Flags().Changed("weight") {
			attrs["weight"] = gczWeight
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --description, --expression and/or --weight")
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterMatchmakingRules/"+gczID, api.Body{
			Data: api.Resource{Type: "gameCenterMatchmakingRules", ID: gczID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Rule %s updated.\n", gczID)
		return nil
	},
}

var gczRulesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a matchmaking rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterMatchmakingRules/"+gczID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// teams

var gczTeamsCmd = &cobra.Command{
	Use:     "teams",
	Aliases: []string{"team"},
	Short:   "Manage matchmaking teams within a rule set",
}

var gczTeamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the teams of a rule set",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		teams, err := c.List(cmd.Context(), "/v1/gameCenterMatchmakingRuleSets/"+gczRuleSetID+"/teams?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tMIN\tMAX\tID")
		for i := range teams {
			t := &teams[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Str("referenceName"), gcxNum(t, "minPlayers"), gcxNum(t, "maxPlayers"), t.ID)
		}
		return w.Flush()
	},
}

var gczTeamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a matchmaking team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterMatchmakingTeams", api.Body{
			Data: api.Resource{
				Type: "gameCenterMatchmakingTeams",
				Attributes: map[string]any{
					"referenceName": gczRefName,
					"minPlayers":    gczMinPlayers,
					"maxPlayers":    gczMaxPlayers,
				},
				Relationships: map[string]json.RawMessage{"ruleSet": api.Rel("gameCenterMatchmakingRuleSets", gczRuleSetID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created team %s (%s).\n", gczRefName, created.ID)
		return nil
	},
}

var gczTeamsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a matchmaking team's player counts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("min-players") {
			attrs["minPlayers"] = gczMinPlayers
		}
		if cmd.Flags().Changed("max-players") {
			attrs["maxPlayers"] = gczMaxPlayers
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --min-players and/or --max-players")
		}
		_, err = c.Patch(cmd.Context(), "/v1/gameCenterMatchmakingTeams/"+gczID, api.Body{
			Data: api.Resource{Type: "gameCenterMatchmakingTeams", ID: gczID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Team %s updated.\n", gczID)
		return nil
	},
}

var gczTeamsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a matchmaking team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/gameCenterMatchmakingTeams/"+gczID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// --- gamecenter submit-score / submit-achievement -----------------------------------

var gczSubmitScoreCmd = &cobra.Command{
	Use:   "submit-score",
	Short: "Submit a leaderboard score on behalf of a player (server-to-server)",
	Long: `POST /v1/gameCenterLeaderboardEntrySubmissions. Submits a score to a
leaderboard for a scoped player id. The score is passed as a string (64-bit
integer). Requires an API key authorized for server-to-server Game Center
submissions.`,
	Example: `  asc gamecenter submit-score --bundle-id com.example.app \
    --leaderboard-vendor-id com.example.high_scores --scoped-player-id <id> --score 12345`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"bundleId":         gczBundleID,
			"vendorIdentifier": gczVendorID,
			"scopedPlayerId":   gczScopedPlayerID,
			"score":            gczScore,
		}
		if cmd.Flags().Changed("context") {
			attrs["context"] = gczContext
		}
		if cmd.Flags().Changed("submitted-date") {
			attrs["submittedDate"] = gczSubmittedDate
		}
		if cmd.Flags().Changed("challenge-ids") {
			attrs["challengeIds"] = gczChallengeIDs
		}
		if cmd.Flags().Changed("pre-released") {
			attrs["preReleased"] = gczPreReleased
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterLeaderboardEntrySubmissions", api.Body{
			Data: api.Resource{Type: "gameCenterLeaderboardEntrySubmissions", Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted score %s to %s (submission %s).\n", gczScore, gczVendorID, created.ID)
		return nil
	},
}

var gczSubmitAchievementCmd = &cobra.Command{
	Use:   "submit-achievement",
	Short: "Submit achievement progress on behalf of a player (server-to-server)",
	Long: `POST /v1/gameCenterPlayerAchievementSubmissions. Reports a player's
achievement progress (0-100 percent) for a scoped player id. Requires an API
key authorized for server-to-server Game Center submissions.`,
	Example: `  asc gamecenter submit-achievement --bundle-id com.example.app \
    --achievement-vendor-id com.example.first_win --scoped-player-id <id> --percentage-achieved 100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"bundleId":           gczBundleID,
			"vendorIdentifier":   gczVendorID,
			"scopedPlayerId":     gczScopedPlayerID,
			"percentageAchieved": gczPercentage,
		}
		if cmd.Flags().Changed("submitted-date") {
			attrs["submittedDate"] = gczSubmittedDate
		}
		if cmd.Flags().Changed("challenge-ids") {
			attrs["challengeIds"] = gczChallengeIDs
		}
		if cmd.Flags().Changed("pre-released") {
			attrs["preReleased"] = gczPreReleased
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterPlayerAchievementSubmissions", api.Body{
			Data: api.Resource{Type: "gameCenterPlayerAchievementSubmissions", Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted %d%% for %s (submission %s).\n", gczPercentage, gczVendorID, created.ID)
		return nil
	},
}

func init() {
	// app-versions
	gczAppVersionsListCmd.Flags().StringVar(&gczApp, "app", "", "app id or bundle id (required)")
	_ = gczAppVersionsListCmd.MarkFlagRequired("app")
	for _, sub := range []*cobra.Command{gczAppVersionsEnableCmd, gczAppVersionsDisableCmd} {
		sub.Flags().StringVar(&gczVersionID, "version-id", "", "appStoreVersion id (required)")
		_ = sub.MarkFlagRequired("version-id")
	}
	for _, sub := range []*cobra.Command{gczAppVersionsLinkCmd, gczAppVersionsUnlinkCmd} {
		sub.Flags().StringVar(&gczAppVersionID, "id", "", "gameCenterAppVersion id (required)")
		sub.Flags().StringVar(&gczCompatVersionID, "compatibility-version-id", "", "compatible gameCenterAppVersion id (required)")
		_ = sub.MarkFlagRequired("id")
		_ = sub.MarkFlagRequired("compatibility-version-id")
	}
	gczAppVersionsCmd.AddCommand(gczAppVersionsListCmd, gczAppVersionsEnableCmd, gczAppVersionsDisableCmd,
		gczAppVersionsLinkCmd, gczAppVersionsUnlinkCmd)

	// activities
	gczActivitiesCreateCmd.Flags().StringVar(&gczApp, "app", "", "app id or bundle id (required)")
	_ = gczActivitiesCreateCmd.MarkFlagRequired("app")
	gczActivitiesCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "internal reference name (required)")
	gczActivitiesCreateCmd.Flags().StringVar(&gczVendorID, "vendor-identifier", "", "vendor identifier, e.g. com.example.daily_race (required)")
	gczActivitiesCreateCmd.Flags().StringVar(&gczPlayStyle, "play-style", "", "ASYNCHRONOUS or SYNCHRONOUS")
	gczActivitiesCreateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players count")
	gczActivitiesCreateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players count")
	gczActivitiesCreateCmd.Flags().BoolVar(&gczSupportsPartyCode, "supports-party-code", false, "activity supports party codes")
	_ = gczActivitiesCreateCmd.MarkFlagRequired("reference-name")
	_ = gczActivitiesCreateCmd.MarkFlagRequired("vendor-identifier")

	gczActivitiesUpdateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "internal reference name")
	gczActivitiesUpdateCmd.Flags().StringVar(&gczPlayStyle, "play-style", "", "ASYNCHRONOUS or SYNCHRONOUS")
	gczActivitiesUpdateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players count")
	gczActivitiesUpdateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players count")
	gczActivitiesUpdateCmd.Flags().BoolVar(&gczSupportsPartyCode, "supports-party-code", false, "activity supports party codes")
	gczActivitiesUpdateCmd.Flags().BoolVar(&gczArchived, "archived", false, "archive the activity")

	for _, sub := range []*cobra.Command{gczActivitiesUpdateCmd, gczActivitiesDeleteCmd, gczActivitiesLocalizeCmd} {
		sub.Flags().StringVar(&gczActivityID, "activity-id", "", "gameCenterActivity id (required)")
		_ = sub.MarkFlagRequired("activity-id")
	}
	gczActivitiesLocalizeCmd.Flags().StringVar(&gczLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gczActivitiesLocalizeCmd.Flags().StringVar(&gczName, "name", "", "localized display name")
	gczActivitiesLocalizeCmd.Flags().StringVar(&gczDesc, "description", "", "localized description (@file allowed)")

	// challenges
	gczChallengesCreateCmd.Flags().StringVar(&gczApp, "app", "", "app id or bundle id (required)")
	_ = gczChallengesCreateCmd.MarkFlagRequired("app")
	gczChallengesCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "internal reference name (required)")
	gczChallengesCreateCmd.Flags().StringVar(&gczVendorID, "vendor-identifier", "", "vendor identifier, e.g. com.example.weekly_sprint (required)")
	gczChallengesCreateCmd.Flags().StringVar(&gczChallengeType, "challenge-type", "LEADERBOARD", "challenge type (LEADERBOARD)")
	gczChallengesCreateCmd.Flags().BoolVar(&gczRepeatable, "repeatable", false, "challenge can be repeated")
	gczChallengesCreateCmd.Flags().StringVar(&gczLeaderboardID, "leaderboard-id", "", "gameCenterLeaderboard id backing the challenge")
	_ = gczChallengesCreateCmd.MarkFlagRequired("reference-name")
	_ = gczChallengesCreateCmd.MarkFlagRequired("vendor-identifier")

	gczChallengesUpdateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "internal reference name")
	gczChallengesUpdateCmd.Flags().BoolVar(&gczRepeatable, "repeatable", false, "challenge can be repeated")
	gczChallengesUpdateCmd.Flags().BoolVar(&gczArchived, "archived", false, "archive the challenge")
	gczChallengesUpdateCmd.Flags().StringVar(&gczLeaderboardID, "leaderboard-id", "", "gameCenterLeaderboard id backing the challenge")

	for _, sub := range []*cobra.Command{gczChallengesUpdateCmd, gczChallengesDeleteCmd, gczChallengesLocalizeCmd} {
		sub.Flags().StringVar(&gczChallengeID, "challenge-id", "", "gameCenterChallenge id (required)")
		_ = sub.MarkFlagRequired("challenge-id")
	}
	gczChallengesLocalizeCmd.Flags().StringVar(&gczLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gczChallengesLocalizeCmd.Flags().StringVar(&gczName, "name", "", "localized display name")
	gczChallengesLocalizeCmd.Flags().StringVar(&gczDesc, "description", "", "localized description (@file allowed)")

	// image uploads / deletions
	for _, sub := range []*cobra.Command{gczActivitiesUploadImageCmd, gczChallengesUploadImageCmd, gczSetUploadImageCmd} {
		sub.Flags().StringVar(&gczLocalizationID, "localization-id", "", "localization id (required)")
		sub.Flags().StringVar(&gczFile, "file", "", "image file (required)")
		_ = sub.MarkFlagRequired("localization-id")
		_ = sub.MarkFlagRequired("file")
	}
	for _, sub := range []*cobra.Command{gczActivitiesDeleteImageCmd, gczChallengesDeleteImageCmd, gczSetDeleteImageCmd} {
		sub.Flags().StringVar(&gczImageID, "id", "", "image id (required)")
		_ = sub.MarkFlagRequired("id")
	}

	// leaderboard set member localizations
	for _, sub := range []*cobra.Command{gczMemberLocsListCmd, gczMemberLocsSetCmd} {
		sub.Flags().StringVar(&gczSetID, "set-id", "", "gameCenterLeaderboardSet id (required)")
		sub.Flags().StringVar(&gczLeaderboardID, "leaderboard-id", "", "gameCenterLeaderboard id (required)")
		_ = sub.MarkFlagRequired("set-id")
		_ = sub.MarkFlagRequired("leaderboard-id")
	}
	gczMemberLocsSetCmd.Flags().StringVar(&gczLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gczMemberLocsSetCmd.Flags().StringVar(&gczName, "name", "", "localized name of the leaderboard within the set (required)")
	_ = gczMemberLocsSetCmd.MarkFlagRequired("name")
	gczMemberLocsCmd.AddCommand(gczMemberLocsListCmd, gczMemberLocsSetCmd)

	// matchmaking: queues
	gczQueuesCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "queue reference name (required)")
	gczQueuesCreateCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id (required)")
	gczQueuesCreateCmd.Flags().StringVar(&gczExperimentRuleSetID, "experiment-rule-set-id", "", "experimental rule set id")
	gczQueuesCreateCmd.Flags().StringSliceVar(&gczClassicBundleIDs, "classic-bundle-ids", nil, "bundle ids using classic matchmaking (comma-separated)")
	_ = gczQueuesCreateCmd.MarkFlagRequired("reference-name")
	_ = gczQueuesCreateCmd.MarkFlagRequired("rule-set-id")

	gczQueuesUpdateCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id")
	gczQueuesUpdateCmd.Flags().StringVar(&gczExperimentRuleSetID, "experiment-rule-set-id", "", "experimental rule set id")
	gczQueuesUpdateCmd.Flags().StringSliceVar(&gczClassicBundleIDs, "classic-bundle-ids", nil, "bundle ids using classic matchmaking (comma-separated)")

	for _, sub := range []*cobra.Command{gczQueuesUpdateCmd, gczQueuesDeleteCmd} {
		sub.Flags().StringVar(&gczID, "id", "", "gameCenterMatchmakingQueue id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	gczQueuesCmd.AddCommand(gczQueuesListCmd, gczQueuesCreateCmd, gczQueuesUpdateCmd, gczQueuesDeleteCmd)

	// matchmaking: rule sets
	gczRuleSetsCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "rule set reference name (required)")
	gczRuleSetsCreateCmd.Flags().IntVar(&gczRuleLanguageVersion, "rule-language-version", 1, "rule expression language version")
	gczRuleSetsCreateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players per match (required)")
	gczRuleSetsCreateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players per match (required)")
	_ = gczRuleSetsCreateCmd.MarkFlagRequired("reference-name")
	_ = gczRuleSetsCreateCmd.MarkFlagRequired("min-players")
	_ = gczRuleSetsCreateCmd.MarkFlagRequired("max-players")

	gczRuleSetsUpdateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players per match")
	gczRuleSetsUpdateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players per match")

	for _, sub := range []*cobra.Command{gczRuleSetsUpdateCmd, gczRuleSetsDeleteCmd} {
		sub.Flags().StringVar(&gczID, "id", "", "gameCenterMatchmakingRuleSet id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	gczRuleSetsTestCmd.Flags().StringVar(&gczBody, "body", "", "full JSON:API request document, inline or @file (required)")
	_ = gczRuleSetsTestCmd.MarkFlagRequired("body")
	gczRuleSetsCmd.AddCommand(gczRuleSetsListCmd, gczRuleSetsCreateCmd, gczRuleSetsUpdateCmd,
		gczRuleSetsDeleteCmd, gczRuleSetsTestCmd)

	// matchmaking: rules
	gczRulesListCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id (required)")
	_ = gczRulesListCmd.MarkFlagRequired("rule-set-id")

	gczRulesCreateCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id (required)")
	gczRulesCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "rule reference name (required)")
	gczRulesCreateCmd.Flags().StringVar(&gczDesc, "description", "", "rule description (required)")
	gczRulesCreateCmd.Flags().StringVar(&gczRuleType, "type", "", "COMPATIBLE, DISTANCE, MATCH or TEAM (required)")
	gczRulesCreateCmd.Flags().StringVar(&gczExpression, "expression", "", "rule expression (@file allowed) (required)")
	gczRulesCreateCmd.Flags().Float64Var(&gczWeight, "weight", 0, "rule weight")
	_ = gczRulesCreateCmd.MarkFlagRequired("rule-set-id")
	_ = gczRulesCreateCmd.MarkFlagRequired("reference-name")
	_ = gczRulesCreateCmd.MarkFlagRequired("description")
	_ = gczRulesCreateCmd.MarkFlagRequired("type")
	_ = gczRulesCreateCmd.MarkFlagRequired("expression")

	gczRulesUpdateCmd.Flags().StringVar(&gczDesc, "description", "", "rule description")
	gczRulesUpdateCmd.Flags().StringVar(&gczExpression, "expression", "", "rule expression (@file allowed)")
	gczRulesUpdateCmd.Flags().Float64Var(&gczWeight, "weight", 0, "rule weight")

	for _, sub := range []*cobra.Command{gczRulesUpdateCmd, gczRulesDeleteCmd} {
		sub.Flags().StringVar(&gczID, "id", "", "gameCenterMatchmakingRule id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	gczRulesCmd.AddCommand(gczRulesListCmd, gczRulesCreateCmd, gczRulesUpdateCmd, gczRulesDeleteCmd)

	// matchmaking: teams
	gczTeamsListCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id (required)")
	_ = gczTeamsListCmd.MarkFlagRequired("rule-set-id")

	gczTeamsCreateCmd.Flags().StringVar(&gczRuleSetID, "rule-set-id", "", "gameCenterMatchmakingRuleSet id (required)")
	gczTeamsCreateCmd.Flags().StringVar(&gczRefName, "reference-name", "", "team reference name (required)")
	gczTeamsCreateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players per team (required)")
	gczTeamsCreateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players per team (required)")
	_ = gczTeamsCreateCmd.MarkFlagRequired("rule-set-id")
	_ = gczTeamsCreateCmd.MarkFlagRequired("reference-name")
	_ = gczTeamsCreateCmd.MarkFlagRequired("min-players")
	_ = gczTeamsCreateCmd.MarkFlagRequired("max-players")

	gczTeamsUpdateCmd.Flags().IntVar(&gczMinPlayers, "min-players", 0, "minimum players per team")
	gczTeamsUpdateCmd.Flags().IntVar(&gczMaxPlayers, "max-players", 0, "maximum players per team")

	for _, sub := range []*cobra.Command{gczTeamsUpdateCmd, gczTeamsDeleteCmd} {
		sub.Flags().StringVar(&gczID, "id", "", "gameCenterMatchmakingTeam id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	gczTeamsCmd.AddCommand(gczTeamsListCmd, gczTeamsCreateCmd, gczTeamsUpdateCmd, gczTeamsDeleteCmd)

	gczMatchmakingCmd.AddCommand(gczQueuesCmd, gczRuleSetsCmd, gczRulesCmd, gczTeamsCmd)

	// submissions
	gczSubmitScoreCmd.Flags().StringVar(&gczBundleID, "bundle-id", "", "app bundle id (required)")
	gczSubmitScoreCmd.Flags().StringVar(&gczVendorID, "leaderboard-vendor-id", "", "leaderboard vendor identifier (required)")
	gczSubmitScoreCmd.Flags().StringVar(&gczScopedPlayerID, "scoped-player-id", "", "scoped player id (required)")
	gczSubmitScoreCmd.Flags().StringVar(&gczScore, "score", "", "score as a 64-bit integer string (required)")
	gczSubmitScoreCmd.Flags().StringVar(&gczContext, "context", "", "leaderboard context (64-bit integer string)")
	gczSubmitScoreCmd.Flags().StringVar(&gczSubmittedDate, "submitted-date", "", "submission time, RFC 3339 (default: now, server-side)")
	gczSubmitScoreCmd.Flags().StringSliceVar(&gczChallengeIDs, "challenge-ids", nil, "challenge ids the score counts toward (comma-separated)")
	gczSubmitScoreCmd.Flags().BoolVar(&gczPreReleased, "pre-released", false, "submit against the pre-release (unpublished) leaderboard")
	_ = gczSubmitScoreCmd.MarkFlagRequired("bundle-id")
	_ = gczSubmitScoreCmd.MarkFlagRequired("leaderboard-vendor-id")
	_ = gczSubmitScoreCmd.MarkFlagRequired("scoped-player-id")
	_ = gczSubmitScoreCmd.MarkFlagRequired("score")

	gczSubmitAchievementCmd.Flags().StringVar(&gczBundleID, "bundle-id", "", "app bundle id (required)")
	gczSubmitAchievementCmd.Flags().StringVar(&gczVendorID, "achievement-vendor-id", "", "achievement vendor identifier (required)")
	gczSubmitAchievementCmd.Flags().StringVar(&gczScopedPlayerID, "scoped-player-id", "", "scoped player id (required)")
	gczSubmitAchievementCmd.Flags().IntVar(&gczPercentage, "percentage-achieved", 0, "progress percentage, 0-100 (required)")
	gczSubmitAchievementCmd.Flags().StringVar(&gczSubmittedDate, "submitted-date", "", "submission time, RFC 3339 (default: now, server-side)")
	gczSubmitAchievementCmd.Flags().StringSliceVar(&gczChallengeIDs, "challenge-ids", nil, "challenge ids the progress counts toward (comma-separated)")
	gczSubmitAchievementCmd.Flags().BoolVar(&gczPreReleased, "pre-released", false, "submit against the pre-release (unpublished) achievement")
	_ = gczSubmitAchievementCmd.MarkFlagRequired("bundle-id")
	_ = gczSubmitAchievementCmd.MarkFlagRequired("achievement-vendor-id")
	_ = gczSubmitAchievementCmd.MarkFlagRequired("scoped-player-id")
	_ = gczSubmitAchievementCmd.MarkFlagRequired("percentage-achieved")

	// Attach to the existing gamecenter command tree (gamecenter.go). The
	// activities / challenges groups already exist with a list subcommand;
	// they gain full CRUD here, so refresh their summaries.
	gcxActivitiesCmd.Short = "Manage Game Center activities (versioned)"
	gcxChallengesCmd.Short = "Manage Game Center challenges (versioned)"
	gcxActivitiesCmd.AddCommand(gczActivitiesCreateCmd, gczActivitiesUpdateCmd, gczActivitiesDeleteCmd,
		gczActivitiesLocalizeCmd, gczActivitiesUploadImageCmd, gczActivitiesDeleteImageCmd)
	gcxChallengesCmd.AddCommand(gczChallengesCreateCmd, gczChallengesUpdateCmd, gczChallengesDeleteCmd,
		gczChallengesLocalizeCmd, gczChallengesUploadImageCmd, gczChallengesDeleteImageCmd)
	gcxLeaderboardSetsCmd.AddCommand(gczSetUploadImageCmd, gczSetDeleteImageCmd, gczMemberLocsCmd)
	gamecenterCmd.AddCommand(gczAppVersionsCmd, gczMatchmakingCmd, gczSubmitScoreCmd, gczSubmitAchievementCmd)
}
