package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var gamecenterCmd = &cobra.Command{
	Use:     "gamecenter",
	Aliases: []string{"gc"},
	Short:   "Manage Game Center (achievements, leaderboards, leaderboard sets, groups)",
	Long: `Manage Game Center configuration. Uses the current (v2, versioned) Game
Center resources: achievements and leaderboards carry versions, localizations
hang off a version, and images hang off a localization.

Publishing note: the v1 "release" resources (gameCenterAchievementReleases etc.)
are deprecated in API 4.x. Versioned Game Center resources are published by
attaching their version to a review submission (reviewSubmissionItems with a
gameCenterAchievementVersion / gameCenterLeaderboardVersion relationship), so
there is no standalone release command here.`,
}

var (
	gcxApp               string
	gcxRefName           string
	gcxVendorID          string
	gcxPoints            int
	gcxRepeatable        bool
	gcxShowBeforeEarned  bool
	gcxArchived          bool
	gcxAchievementID     string
	gcxLeaderboardID     string
	gcxSetID             string
	gcxLocalizationID    string
	gcxFile              string
	gcxLocale            string
	gcxName              string
	gcxBeforeDesc        string
	gcxAfterDesc         string
	gcxDesc              string
	gcxSortType          string
	gcxFormatter         string
	gcxSubmissionType    string
	gcxRangeStart        string
	gcxRangeEnd          string
	gcxSuffix            string
	gcxSuffixSingular    string
	gcxFormatterOverride string
)

// gcxDetailID resolves the app reference to its gameCenterDetail id.
func gcxDetailID(ctx context.Context, c *api.Client, appRef string) (string, error) {
	appID, err := resolveAppID(ctx, c, appRef)
	if err != nil {
		return "", err
	}
	detail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/gameCenterDetail")
	if err != nil {
		return "", err
	}
	if detail == nil || detail.ID == "" {
		return "", fmt.Errorf("Game Center is not enabled for app %s (run: asc gamecenter enable --app %s)", appID, appRef)
	}
	return detail.ID, nil
}

// gcxLatestVersion returns the version resource with the highest version number.
func gcxLatestVersion(ctx context.Context, c *api.Client, path string) (*api.Resource, error) {
	vers, err := c.List(ctx, path)
	if err != nil {
		return nil, err
	}
	var best *api.Resource
	bestN := -1.0
	for i := range vers {
		n, _ := vers[i].Attributes["version"].(float64)
		if best == nil || n > bestN {
			best = &vers[i]
			bestN = n
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no versions found (GET %s)", path)
	}
	return best, nil
}

// gcxUploadImage runs reserve → upload → commit for the v2 Game Center image
// resources (they live under /v2 and commit with uploaded only, so the generic
// v1 uploadAsset helper does not apply).
func gcxUploadImage(ctx context.Context, c *api.Client, createPath, resType, relType, locID, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	reserved, err := c.Post(ctx, createPath, api.Body{
		Data: api.Resource{
			Type:          resType,
			Attributes:    map[string]any{"fileName": filepath.Base(filePath), "fileSize": len(data)},
			Relationships: map[string]json.RawMessage{"localization": api.Rel(relType, locID)},
		},
	})
	if err != nil {
		return "", err
	}
	if c.DryRun {
		return "dry-run", nil
	}
	var ops []api.UploadOperation
	if err := reserved.DecodeAttr("uploadOperations", &ops); err != nil {
		return "", fmt.Errorf("reservation for %s returned no upload operations: %w", filepath.Base(filePath), err)
	}
	if err := c.Upload(ctx, ops, data); err != nil {
		return "", err
	}
	_, err = c.Patch(ctx, createPath+"/"+reserved.ID, api.Body{
		Data: api.Resource{
			Type:       resType,
			ID:         reserved.ID,
			Attributes: map[string]any{"uploaded": true},
		},
	})
	if err != nil {
		return "", err
	}
	return reserved.ID, nil
}

// gcxToManyRel builds a to-many relationship value: {"data":[{"type":..,"id":..}]}.
func gcxToManyRel(typ, id string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"data": []map[string]string{{"type": typ, "id": id}}})
	return b
}

func gcxNum(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(float64); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	return ""
}

func gcxBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// --- show / enable -------------------------------------------------------------

var gcxShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the app's Game Center detail",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/gameCenterDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Printf("Game Center is not enabled for app %s. Enable it with: asc gamecenter enable --app %s\n", appID, gcxApp)
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

var gcxEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Game Center for an app (create its gameCenterDetail)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		detail, err := c.Post(ctx, "/v1/gameCenterDetails", api.Body{
			Data: api.Resource{
				Type:          "gameCenterDetails",
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Game Center enabled for app %s (detail %s).\n", appID, detail.ID)
		return nil
	},
}

// --- achievements ---------------------------------------------------------------

var gcxAchievementsCmd = &cobra.Command{
	Use:     "achievements",
	Aliases: []string{"achievement"},
	Short:   "Manage Game Center achievements (v2, versioned)",
}

var gcxAchievementsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's achievements",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		achievements, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterAchievementsV2?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tVENDOR_ID\tPOINTS\tARCHIVED\tID")
		for _, a := range achievements {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				a.Str("referenceName"), a.Str("vendorIdentifier"), gcxNum(&a, "points"), gcxBool(&a, "archived"), a.ID)
		}
		return w.Flush()
	},
}

var gcxAchievementsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an achievement (an initial version is created inline)",
	Example: `  asc gamecenter achievements create --app 6790641087 \
    --reference-name "First Win" --vendor-identifier com.example.first_win --points 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		const verLID = "ach-ver-1"
		created, err := c.Post(ctx, "/v2/gameCenterAchievements", api.Body{
			Data: api.Resource{
				Type: "gameCenterAchievements",
				Attributes: map[string]any{
					"referenceName":    gcxRefName,
					"vendorIdentifier": gcxVendorID,
					"points":           gcxPoints,
					"repeatable":       gcxRepeatable,
					"showBeforeEarned": gcxShowBeforeEarned,
				},
				Relationships: map[string]json.RawMessage{
					"gameCenterDetail": api.Rel("gameCenterDetails", detailID),
					"versions":         gcxToManyRel("gameCenterAchievementVersions", verLID),
				},
			},
			Included: []api.Resource{{Type: "gameCenterAchievementVersions", ID: verLID}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created achievement %s (%s).\n", gcxRefName, created.ID)
		return nil
	},
}

var gcxAchievementsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an achievement's attributes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("reference-name") {
			attrs["referenceName"] = gcxRefName
		}
		if cmd.Flags().Changed("points") {
			attrs["points"] = gcxPoints
		}
		if cmd.Flags().Changed("repeatable") {
			attrs["repeatable"] = gcxRepeatable
		}
		if cmd.Flags().Changed("show-before-earned") {
			attrs["showBeforeEarned"] = gcxShowBeforeEarned
		}
		if cmd.Flags().Changed("archived") {
			attrs["archived"] = gcxArchived
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --reference-name, --points, --repeatable, --show-before-earned and/or --archived")
		}
		_, err = c.Patch(cmd.Context(), "/v2/gameCenterAchievements/"+gcxAchievementID, api.Body{
			Data: api.Resource{Type: "gameCenterAchievements", ID: gcxAchievementID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Achievement %s updated.\n", gcxAchievementID)
		return nil
	},
}

var gcxAchievementsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an achievement",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v2/gameCenterAchievements/"+gcxAchievementID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gcxAchievementsLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set an achievement's localized name and descriptions",
	Long: `Set an achievement's localized name and earned/unearned descriptions.
Localizations belong to the achievement's latest version (versions are created
automatically when the achievement is created). Creating a new localization
requires --name, --before-earned-desc and --after-earned-desc; updating an
existing one accepts any subset.`,
	Example: `  asc gamecenter achievements localize --achievement-id <id> --locale ja \
    --name "初勝利" --before-earned-desc "最初の対戦に勝つ" --after-earned-desc "最初の対戦に勝った"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, err := gcxLatestVersion(ctx, c, "/v2/gameCenterAchievements/"+gcxAchievementID+"/versions?limit=50")
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = gcxName
		}
		if cmd.Flags().Changed("before-earned-desc") {
			v, err := valueOrFile(gcxBeforeDesc)
			if err != nil {
				return err
			}
			attrs["beforeEarnedDescription"] = v
		}
		if cmd.Flags().Changed("after-earned-desc") {
			v, err := valueOrFile(gcxAfterDesc)
			if err != nil {
				return err
			}
			attrs["afterEarnedDescription"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name, --before-earned-desc and/or --after-earned-desc")
		}
		locs, err := c.List(ctx, "/v2/gameCenterAchievementVersions/"+ver.ID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gcxLocale); existing != nil {
			_, err = c.Patch(ctx, "/v2/gameCenterAchievementLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "gameCenterAchievementLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			for _, k := range []string{"name", "beforeEarnedDescription", "afterEarnedDescription"} {
				if _, ok := attrs[k]; !ok {
					return fmt.Errorf("creating a new %s localization requires --name, --before-earned-desc and --after-earned-desc", gcxLocale)
				}
			}
			attrs["locale"] = gcxLocale
			_, err = c.Post(ctx, "/v2/gameCenterAchievementLocalizations", api.Body{
				Data: api.Resource{
					Type:          "gameCenterAchievementLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"version": api.Rel("gameCenterAchievementVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Achievement %s localization (%s) updated on version %s.\n", gcxAchievementID, gcxLocale, ver.ID)
		return nil
	},
}

var gcxAchievementsUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload the image for an achievement localization",
	Long: `Upload an achievement image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the achievement's latest version:
GET /v2/gameCenterAchievements/{id}/versions then
GET /v2/gameCenterAchievementVersions/{id}/localizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := gcxUploadImage(cmd.Context(), c,
			"/v2/gameCenterAchievementImages", "gameCenterAchievementImages",
			"gameCenterAchievementLocalizations", gcxLocalizationID, gcxFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded achievement image -> %s\n", id)
		return nil
	},
}

// --- leaderboards ----------------------------------------------------------------

var gcxLeaderboardsCmd = &cobra.Command{
	Use:     "leaderboards",
	Aliases: []string{"leaderboard"},
	Short:   "Manage Game Center leaderboards (v2, versioned)",
}

var gcxLeaderboardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's leaderboards",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		boards, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterLeaderboardsV2?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tVENDOR_ID\tSORT\tSUBMISSION\tID")
		for _, b := range boards {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				b.Str("referenceName"), b.Str("vendorIdentifier"), b.Str("scoreSortType"), b.Str("submissionType"), b.ID)
		}
		return w.Flush()
	},
}

var gcxLeaderboardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a leaderboard (an initial version is created inline)",
	Example: `  asc gamecenter leaderboards create --app 6790641087 \
    --reference-name "High Scores" --vendor-identifier com.example.high_scores \
    --score-sort-type DESC --default-formatter INTEGER --submission-type BEST_SCORE`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"referenceName":    gcxRefName,
			"vendorIdentifier": gcxVendorID,
			"scoreSortType":    gcxSortType,
			"defaultFormatter": gcxFormatter,
			"submissionType":   gcxSubmissionType,
		}
		if cmd.Flags().Changed("score-range-start") {
			attrs["scoreRangeStart"] = gcxRangeStart
		}
		if cmd.Flags().Changed("score-range-end") {
			attrs["scoreRangeEnd"] = gcxRangeEnd
		}
		const verLID = "lb-ver-1"
		created, err := c.Post(ctx, "/v2/gameCenterLeaderboards", api.Body{
			Data: api.Resource{
				Type:       "gameCenterLeaderboards",
				Attributes: attrs,
				Relationships: map[string]json.RawMessage{
					"gameCenterDetail": api.Rel("gameCenterDetails", detailID),
					"versions":         gcxToManyRel("gameCenterLeaderboardVersions", verLID),
				},
			},
			Included: []api.Resource{{Type: "gameCenterLeaderboardVersions", ID: verLID}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created leaderboard %s (%s).\n", gcxRefName, created.ID)
		return nil
	},
}

var gcxLeaderboardsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a leaderboard's attributes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("reference-name") {
			attrs["referenceName"] = gcxRefName
		}
		if cmd.Flags().Changed("score-sort-type") {
			attrs["scoreSortType"] = gcxSortType
		}
		if cmd.Flags().Changed("default-formatter") {
			attrs["defaultFormatter"] = gcxFormatter
		}
		if cmd.Flags().Changed("submission-type") {
			attrs["submissionType"] = gcxSubmissionType
		}
		if cmd.Flags().Changed("score-range-start") {
			attrs["scoreRangeStart"] = gcxRangeStart
		}
		if cmd.Flags().Changed("score-range-end") {
			attrs["scoreRangeEnd"] = gcxRangeEnd
		}
		if cmd.Flags().Changed("archived") {
			attrs["archived"] = gcxArchived
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set")
		}
		_, err = c.Patch(cmd.Context(), "/v2/gameCenterLeaderboards/"+gcxLeaderboardID, api.Body{
			Data: api.Resource{Type: "gameCenterLeaderboards", ID: gcxLeaderboardID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Leaderboard %s updated.\n", gcxLeaderboardID)
		return nil
	},
}

var gcxLeaderboardsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a leaderboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v2/gameCenterLeaderboards/"+gcxLeaderboardID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var gcxLeaderboardsLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set a leaderboard's localized name and formatting",
	Long: `Set a leaderboard's localized name, description and score formatting.
Localizations belong to the leaderboard's latest version. Creating a new
localization requires --name; updates accept any subset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, err := gcxLatestVersion(ctx, c, "/v2/gameCenterLeaderboards/"+gcxLeaderboardID+"/versions?limit=50")
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = gcxName
		}
		if cmd.Flags().Changed("description") {
			v, err := valueOrFile(gcxDesc)
			if err != nil {
				return err
			}
			attrs["description"] = v
		}
		if cmd.Flags().Changed("formatter-override") {
			attrs["formatterOverride"] = gcxFormatterOverride
		}
		if cmd.Flags().Changed("formatter-suffix") {
			attrs["formatterSuffix"] = gcxSuffix
		}
		if cmd.Flags().Changed("formatter-suffix-singular") {
			attrs["formatterSuffixSingular"] = gcxSuffixSingular
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name, --description and/or formatter flags")
		}
		locs, err := c.List(ctx, "/v2/gameCenterLeaderboardVersions/"+ver.ID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gcxLocale); existing != nil {
			_, err = c.Patch(ctx, "/v2/gameCenterLeaderboardLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "gameCenterLeaderboardLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			if _, ok := attrs["name"]; !ok {
				return fmt.Errorf("creating a new %s localization requires --name", gcxLocale)
			}
			attrs["locale"] = gcxLocale
			_, err = c.Post(ctx, "/v2/gameCenterLeaderboardLocalizations", api.Body{
				Data: api.Resource{
					Type:          "gameCenterLeaderboardLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"version": api.Rel("gameCenterLeaderboardVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Leaderboard %s localization (%s) updated on version %s.\n", gcxLeaderboardID, gcxLocale, ver.ID)
		return nil
	},
}

var gcxLeaderboardsUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload the image for a leaderboard localization",
	Long: `Upload a leaderboard image (reserve → upload → commit) and attach it to a
localization. Find the localization id via the leaderboard's latest version:
GET /v2/gameCenterLeaderboards/{id}/versions then
GET /v2/gameCenterLeaderboardVersions/{id}/localizations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := gcxUploadImage(cmd.Context(), c,
			"/v2/gameCenterLeaderboardImages", "gameCenterLeaderboardImages",
			"gameCenterLeaderboardLocalizations", gcxLocalizationID, gcxFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded leaderboard image -> %s\n", id)
		return nil
	},
}

// --- leaderboard sets --------------------------------------------------------------

var gcxLeaderboardSetsCmd = &cobra.Command{
	Use:     "leaderboard-sets",
	Aliases: []string{"leaderboard-set"},
	Short:   "Manage Game Center leaderboard sets (v2, versioned)",
}

var gcxLeaderboardSetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's leaderboard sets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		sets, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterLeaderboardSetsV2?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tVENDOR_ID\tID")
		for _, s := range sets {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.Str("referenceName"), s.Str("vendorIdentifier"), s.ID)
		}
		return w.Flush()
	},
}

var gcxLeaderboardSetsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a leaderboard set (an initial version is created inline)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		const verLID = "lbset-ver-1"
		created, err := c.Post(ctx, "/v2/gameCenterLeaderboardSets", api.Body{
			Data: api.Resource{
				Type: "gameCenterLeaderboardSets",
				Attributes: map[string]any{
					"referenceName":    gcxRefName,
					"vendorIdentifier": gcxVendorID,
				},
				Relationships: map[string]json.RawMessage{
					"gameCenterDetail": api.Rel("gameCenterDetails", detailID),
					"versions":         gcxToManyRel("gameCenterLeaderboardSetVersions", verLID),
				},
			},
			Included: []api.Resource{{Type: "gameCenterLeaderboardSetVersions", ID: verLID}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created leaderboard set %s (%s).\n", gcxRefName, created.ID)
		return nil
	},
}

var gcxLeaderboardSetsLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set a leaderboard set's localized name",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, err := gcxLatestVersion(ctx, c, "/v2/gameCenterLeaderboardSets/"+gcxSetID+"/versions?limit=50")
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v2/gameCenterLeaderboardSetVersions/"+ver.ID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", gcxLocale); existing != nil {
			_, err = c.Patch(ctx, "/v2/gameCenterLeaderboardSetLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "gameCenterLeaderboardSetLocalizations", ID: existing.ID, Attributes: map[string]any{"name": gcxName}},
			})
		} else {
			_, err = c.Post(ctx, "/v2/gameCenterLeaderboardSetLocalizations", api.Body{
				Data: api.Resource{
					Type:          "gameCenterLeaderboardSetLocalizations",
					Attributes:    map[string]any{"locale": gcxLocale, "name": gcxName},
					Relationships: map[string]json.RawMessage{"version": api.Rel("gameCenterLeaderboardSetVersions", ver.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Leaderboard set %s localization (%s) updated on version %s.\n", gcxSetID, gcxLocale, ver.ID)
		return nil
	},
}

var gcxLeaderboardSetsAddLeaderboardCmd = &cobra.Command{
	Use:   "add-leaderboard",
	Short: "Add a leaderboard to a leaderboard set",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v2/gameCenterLeaderboardSets/" + gcxSetID + "/relationships/gameCenterLeaderboards"
		payload, err := json.Marshal(map[string]any{
			"data": []map[string]string{{"type": "gameCenterLeaderboards", "id": gcxLeaderboardID}},
		})
		if err != nil {
			return err
		}
		if dryRun {
			fmt.Fprintf(os.Stderr, "DRY-RUN POST %s\n%s\n", path, payload)
			return nil
		}
		if _, err := c.Do(cmd.Context(), http.MethodPost, path, bytes.NewReader(payload)); err != nil {
			return err
		}
		fmt.Printf("Added leaderboard %s to set %s.\n", gcxLeaderboardID, gcxSetID)
		return nil
	},
}

// --- groups -------------------------------------------------------------------------

var gcxGroupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"group"},
	Short:   "Manage Game Center groups",
}

var gcxGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Game Center groups visible to this team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		groups, err := c.List(cmd.Context(), "/v1/gameCenterGroups?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tID")
		for _, g := range groups {
			fmt.Fprintf(w, "%s\t%s\n", g.Str("referenceName"), g.ID)
		}
		return w.Flush()
	},
}

var gcxGroupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Game Center group",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		res := api.Resource{Type: "gameCenterGroups"}
		if cmd.Flags().Changed("reference-name") {
			res.Attributes = map[string]any{"referenceName": gcxRefName}
		}
		created, err := c.Post(cmd.Context(), "/v1/gameCenterGroups", api.Body{Data: res})
		if err != nil {
			return err
		}
		fmt.Printf("Created group %s.\n", created.ID)
		return nil
	},
}

// --- activities / challenges (read-only) -----------------------------------------------

var gcxActivitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "Game Center activities (read-only)",
}

var gcxActivitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's Game Center activities",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		activities, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterActivities?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tVENDOR_ID\tPLAY_STYLE\tARCHIVED\tID")
		for _, a := range activities {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				a.Str("referenceName"), a.Str("vendorIdentifier"), a.Str("playStyle"), gcxBool(&a, "archived"), a.ID)
		}
		return w.Flush()
	},
}

var gcxChallengesCmd = &cobra.Command{
	Use:   "challenges",
	Short: "Game Center challenges (read-only)",
}

var gcxChallengesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's Game Center challenges",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detailID, err := gcxDetailID(ctx, c, gcxApp)
		if err != nil {
			return err
		}
		challenges, err := c.List(ctx, "/v1/gameCenterDetails/"+detailID+"/gameCenterChallenges?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REF_NAME\tVENDOR_ID\tTYPE\tREPEATABLE\tID")
		for _, ch := range challenges {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				ch.Str("referenceName"), ch.Str("vendorIdentifier"), ch.Str("challengeType"), gcxBool(&ch, "repeatable"), ch.ID)
		}
		return w.Flush()
	},
}

func init() {
	// --app flag on commands that resolve the app's gameCenterDetail.
	for _, sub := range []*cobra.Command{
		gcxShowCmd, gcxEnableCmd,
		gcxAchievementsListCmd, gcxAchievementsCreateCmd,
		gcxLeaderboardsListCmd, gcxLeaderboardsCreateCmd,
		gcxLeaderboardSetsListCmd, gcxLeaderboardSetsCreateCmd,
		gcxActivitiesListCmd, gcxChallengesListCmd,
	} {
		sub.Flags().StringVar(&gcxApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}

	// achievements
	gcxAchievementsCreateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name (required)")
	gcxAchievementsCreateCmd.Flags().StringVar(&gcxVendorID, "vendor-identifier", "", "vendor identifier, e.g. com.example.first_win (required)")
	gcxAchievementsCreateCmd.Flags().IntVar(&gcxPoints, "points", 0, "point value, 1-100 (required)")
	gcxAchievementsCreateCmd.Flags().BoolVar(&gcxRepeatable, "repeatable", false, "achievement can be earned more than once")
	gcxAchievementsCreateCmd.Flags().BoolVar(&gcxShowBeforeEarned, "show-before-earned", false, "show the achievement before it is earned")
	_ = gcxAchievementsCreateCmd.MarkFlagRequired("reference-name")
	_ = gcxAchievementsCreateCmd.MarkFlagRequired("vendor-identifier")
	_ = gcxAchievementsCreateCmd.MarkFlagRequired("points")

	gcxAchievementsUpdateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name")
	gcxAchievementsUpdateCmd.Flags().IntVar(&gcxPoints, "points", 0, "point value, 1-100")
	gcxAchievementsUpdateCmd.Flags().BoolVar(&gcxRepeatable, "repeatable", false, "achievement can be earned more than once")
	gcxAchievementsUpdateCmd.Flags().BoolVar(&gcxShowBeforeEarned, "show-before-earned", false, "show the achievement before it is earned")
	gcxAchievementsUpdateCmd.Flags().BoolVar(&gcxArchived, "archived", false, "archive the achievement")

	for _, sub := range []*cobra.Command{gcxAchievementsUpdateCmd, gcxAchievementsDeleteCmd, gcxAchievementsLocalizeCmd} {
		sub.Flags().StringVar(&gcxAchievementID, "achievement-id", "", "gameCenterAchievement id (required)")
		_ = sub.MarkFlagRequired("achievement-id")
	}
	gcxAchievementsLocalizeCmd.Flags().StringVar(&gcxLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gcxAchievementsLocalizeCmd.Flags().StringVar(&gcxName, "name", "", "localized display name")
	gcxAchievementsLocalizeCmd.Flags().StringVar(&gcxBeforeDesc, "before-earned-desc", "", "description before earned (@file allowed)")
	gcxAchievementsLocalizeCmd.Flags().StringVar(&gcxAfterDesc, "after-earned-desc", "", "description after earned (@file allowed)")

	// leaderboards
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name (required)")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxVendorID, "vendor-identifier", "", "vendor identifier, e.g. com.example.high_scores (required)")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxSortType, "score-sort-type", "", "ASC or DESC (required)")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxFormatter, "default-formatter", "INTEGER", "score formatter, e.g. INTEGER, DECIMAL_POINT_1_PLACE, ELAPSED_TIME_SECOND, MONEY_YEN")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxSubmissionType, "submission-type", "BEST_SCORE", "BEST_SCORE or MOST_RECENT_SCORE")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxRangeStart, "score-range-start", "", "minimum accepted score")
	gcxLeaderboardsCreateCmd.Flags().StringVar(&gcxRangeEnd, "score-range-end", "", "maximum accepted score")
	_ = gcxLeaderboardsCreateCmd.MarkFlagRequired("reference-name")
	_ = gcxLeaderboardsCreateCmd.MarkFlagRequired("vendor-identifier")
	_ = gcxLeaderboardsCreateCmd.MarkFlagRequired("score-sort-type")

	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name")
	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxSortType, "score-sort-type", "", "ASC or DESC")
	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxFormatter, "default-formatter", "", "score formatter")
	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxSubmissionType, "submission-type", "", "BEST_SCORE or MOST_RECENT_SCORE")
	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxRangeStart, "score-range-start", "", "minimum accepted score")
	gcxLeaderboardsUpdateCmd.Flags().StringVar(&gcxRangeEnd, "score-range-end", "", "maximum accepted score")
	gcxLeaderboardsUpdateCmd.Flags().BoolVar(&gcxArchived, "archived", false, "archive the leaderboard")

	for _, sub := range []*cobra.Command{gcxLeaderboardsUpdateCmd, gcxLeaderboardsDeleteCmd, gcxLeaderboardsLocalizeCmd} {
		sub.Flags().StringVar(&gcxLeaderboardID, "leaderboard-id", "", "gameCenterLeaderboard id (required)")
		_ = sub.MarkFlagRequired("leaderboard-id")
	}
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxName, "name", "", "localized display name")
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxDesc, "description", "", "localized description (@file allowed)")
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxFormatterOverride, "formatter-override", "", "per-locale formatter override")
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxSuffix, "formatter-suffix", "", "score suffix, e.g. \" points\"")
	gcxLeaderboardsLocalizeCmd.Flags().StringVar(&gcxSuffixSingular, "formatter-suffix-singular", "", "singular score suffix, e.g. \" point\"")

	// image uploads
	for _, sub := range []*cobra.Command{gcxAchievementsUploadImageCmd, gcxLeaderboardsUploadImageCmd} {
		sub.Flags().StringVar(&gcxLocalizationID, "localization-id", "", "localization id (required)")
		sub.Flags().StringVar(&gcxFile, "file", "", "image file (required)")
		_ = sub.MarkFlagRequired("localization-id")
		_ = sub.MarkFlagRequired("file")
	}

	// leaderboard sets
	gcxLeaderboardSetsCreateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name (required)")
	gcxLeaderboardSetsCreateCmd.Flags().StringVar(&gcxVendorID, "vendor-identifier", "", "vendor identifier (required)")
	_ = gcxLeaderboardSetsCreateCmd.MarkFlagRequired("reference-name")
	_ = gcxLeaderboardSetsCreateCmd.MarkFlagRequired("vendor-identifier")

	for _, sub := range []*cobra.Command{gcxLeaderboardSetsLocalizeCmd, gcxLeaderboardSetsAddLeaderboardCmd} {
		sub.Flags().StringVar(&gcxSetID, "set-id", "", "gameCenterLeaderboardSet id (required)")
		_ = sub.MarkFlagRequired("set-id")
	}
	gcxLeaderboardSetsLocalizeCmd.Flags().StringVar(&gcxLocale, "locale", "ja", "locale, e.g. ja / en-US")
	gcxLeaderboardSetsLocalizeCmd.Flags().StringVar(&gcxName, "name", "", "localized display name (required)")
	_ = gcxLeaderboardSetsLocalizeCmd.MarkFlagRequired("name")
	gcxLeaderboardSetsAddLeaderboardCmd.Flags().StringVar(&gcxLeaderboardID, "leaderboard-id", "", "gameCenterLeaderboard id to add (required)")
	_ = gcxLeaderboardSetsAddLeaderboardCmd.MarkFlagRequired("leaderboard-id")

	// groups
	gcxGroupsCreateCmd.Flags().StringVar(&gcxRefName, "reference-name", "", "internal reference name")

	gcxAchievementsCmd.AddCommand(gcxAchievementsListCmd, gcxAchievementsCreateCmd, gcxAchievementsUpdateCmd,
		gcxAchievementsDeleteCmd, gcxAchievementsLocalizeCmd, gcxAchievementsUploadImageCmd)
	gcxLeaderboardsCmd.AddCommand(gcxLeaderboardsListCmd, gcxLeaderboardsCreateCmd, gcxLeaderboardsUpdateCmd,
		gcxLeaderboardsDeleteCmd, gcxLeaderboardsLocalizeCmd, gcxLeaderboardsUploadImageCmd)
	gcxLeaderboardSetsCmd.AddCommand(gcxLeaderboardSetsListCmd, gcxLeaderboardSetsCreateCmd,
		gcxLeaderboardSetsLocalizeCmd, gcxLeaderboardSetsAddLeaderboardCmd)
	gcxGroupsCmd.AddCommand(gcxGroupsListCmd, gcxGroupsCreateCmd)
	gcxActivitiesCmd.AddCommand(gcxActivitiesListCmd)
	gcxChallengesCmd.AddCommand(gcxChallengesListCmd)

	gamecenterCmd.AddCommand(gcxShowCmd, gcxEnableCmd, gcxAchievementsCmd, gcxLeaderboardsCmd,
		gcxLeaderboardSetsCmd, gcxGroupsCmd, gcxActivitiesCmd, gcxChallengesCmd)
	rootCmd.AddCommand(gamecenterCmd)
}
