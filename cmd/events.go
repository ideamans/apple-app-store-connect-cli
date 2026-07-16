package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Work with in-app events (create, schedule, localize, upload media)",
	Long: `Manage in-app events (appEvents): timed events like challenges, competitions
and premieres that appear on the App Store product page.

A typical flow: create the event, localize it (name, descriptions), upload the
event card / details page media, then submit it for review with:

    asc review-submissions add-item --event-id <event-id> ...

Dates are ISO 8601 date-times, e.g. 2026-08-01T10:00:00Z.`,
}

var (
	evtApp          string
	evtEventID      string
	evtState        string
	evtRefName      string
	evtBadge        string
	evtDeepLink     string
	evtPurchaseReq  string
	evtPrimaryLoc   string
	evtPriority     string
	evtPurpose      string
	evtTerritories  string
	evtPublishStart string
	evtStart        string
	evtEnd          string
	evtLocale       string
	evtName         string
	evtShortDesc    string
	evtLongDesc     string
	evtFile         string
	evtDisplay      string
	evtAssetID      string
)

// evtSchedule mirrors one entry of the territorySchedules array attribute.
type evtSchedule struct {
	Territories  []string `json:"territories"`
	PublishStart string   `json:"publishStart"`
	EventStart   string   `json:"eventStart"`
	EventEnd     string   `json:"eventEnd"`
}

// evtScheduleFromFlags builds one territorySchedules entry from the schedule
// flags, or nil when none were passed.
func evtScheduleFromFlags(cmd *cobra.Command) map[string]any {
	sched := map[string]any{}
	if cmd.Flags().Changed("publish-start") {
		sched["publishStart"] = evtPublishStart
	}
	if cmd.Flags().Changed("start") {
		sched["eventStart"] = evtStart
	}
	if cmd.Flags().Changed("end") {
		sched["eventEnd"] = evtEnd
	}
	if cmd.Flags().Changed("territories") {
		sched["territories"] = strings.Split(evtTerritories, ",")
	}
	if len(sched) == 0 {
		return nil
	}
	return sched
}

// evtAttrsFromFlags collects the appEvent attributes set via flags.
func evtAttrsFromFlags(cmd *cobra.Command) map[string]any {
	attrs := map[string]any{}
	set := func(flag, attr string, val *string) {
		if cmd.Flags().Changed(flag) {
			attrs[attr] = *val
		}
	}
	set("reference-name", "referenceName", &evtRefName)
	set("badge", "badge", &evtBadge)
	set("deep-link", "deepLink", &evtDeepLink)
	set("purchase-requirement", "purchaseRequirement", &evtPurchaseReq)
	set("primary-locale", "primaryLocale", &evtPrimaryLoc)
	set("priority", "priority", &evtPriority)
	set("purpose", "purpose", &evtPurpose)
	if sched := evtScheduleFromFlags(cmd); sched != nil {
		attrs["territorySchedules"] = []any{sched}
	}
	return attrs
}

var evtListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's in-app events",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, evtApp)
		if err != nil {
			return err
		}
		path := "/v1/apps/" + appID + "/appEvents?limit=50"
		if evtState != "" {
			path += "&filter[eventState]=" + evtState
		}
		events, err := c.List(ctx, path)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REFERENCE_NAME\tBADGE\tSTATE\tSTART\tEND\tID")
		for _, e := range events {
			start, end := "", ""
			var scheds []evtSchedule
			if err := e.DecodeAttr("territorySchedules", &scheds); err == nil && len(scheds) > 0 {
				start, end = scheds[0].EventStart, scheds[0].EventEnd
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				e.Str("referenceName"), e.Str("badge"), e.Str("eventState"), start, end, e.ID)
		}
		return w.Flush()
	},
}

var evtCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an in-app event",
	Example: `  asc events create --app 6790641087 --reference-name "Summer Challenge" \
    --badge CHALLENGE --purpose ATTRACT_NEW_USERS --priority HIGH \
    --publish-start 2026-08-01T00:00:00Z --start 2026-08-08T00:00:00Z --end 2026-08-15T00:00:00Z`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, evtApp)
		if err != nil {
			return err
		}
		attrs := evtAttrsFromFlags(cmd)
		created, err := c.Post(ctx, "/v1/appEvents", api.Body{
			Data: api.Resource{
				Type:          "appEvents",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created event %q (%s).\n", evtRefName, created.ID)
		return nil
	},
}

var evtUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an in-app event's attributes or schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := evtAttrsFromFlags(cmd)
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass at least one attribute flag")
		}
		// territorySchedules is a whole-array attribute: PATCHing a partial entry
		// would wipe the fields not passed as flags. Merge into the current value.
		if sched, ok := attrs["territorySchedules"].([]any); ok && len(sched) == 1 {
			current, _, err := c.Get(cmd.Context(), "/v1/appEvents/"+evtEventID)
			if err != nil {
				return err
			}
			var existing []map[string]any
			if err := current.DecodeAttr("territorySchedules", &existing); err == nil && len(existing) > 0 {
				merged := existing[0]
				for k, v := range sched[0].(map[string]any) {
					merged[k] = v
				}
				out := make([]any, 0, len(existing))
				out = append(out, merged)
				for _, rest := range existing[1:] {
					out = append(out, rest)
				}
				attrs["territorySchedules"] = out
			}
		}
		_, err = c.Patch(cmd.Context(), "/v1/appEvents/"+evtEventID, api.Body{
			Data: api.Resource{Type: "appEvents", ID: evtEventID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Println("Event updated.")
		return nil
	},
}

var evtDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an in-app event",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appEvents/"+evtEventID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// evtLocalizationID returns the event localization id for a locale, creating an
// empty one when it does not exist.
func evtLocalizationID(ctx context.Context, c *api.Client, eventID, locale string) (string, error) {
	locs, err := c.List(ctx, "/v1/appEvents/"+eventID+"/localizations?limit=50")
	if err != nil {
		return "", err
	}
	if loc := findByAttr(locs, "locale", locale); loc != nil {
		return loc.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appEventLocalizations", api.Body{
		Data: api.Resource{
			Type:          "appEventLocalizations",
			Attributes:    map[string]any{"locale": locale},
			Relationships: map[string]json.RawMessage{"appEvent": api.Rel("appEvents", eventID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

var evtLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set the localized event name and descriptions",
	Example: `  asc events localize --event-id <id> --locale ja \
    --name "夏のチャレンジ" --short-description "1週間の限定イベント" \
    --long-description @app-store/event-long.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		attrs := map[string]any{}
		fields := []struct {
			flag, attr string
			val        *string
		}{
			{"name", "name", &evtName},
			{"short-description", "shortDescription", &evtShortDesc},
			{"long-description", "longDescription", &evtLongDesc},
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
			return fmt.Errorf("nothing to set: pass --name, --short-description and/or --long-description")
		}
		locs, err := c.List(ctx, "/v1/appEvents/"+evtEventID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", evtLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/appEventLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "appEventLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			attrs["locale"] = evtLocale
			_, err = c.Post(ctx, "/v1/appEventLocalizations", api.Body{
				Data: api.Resource{
					Type:          "appEventLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"appEvent": api.Rel("appEvents", evtEventID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Event localization (%s) updated.\n", evtLocale)
		return nil
	},
}

// evtUploadAsset uploads an event screenshot or video clip via reserve → PUT →
// commit. It differs from the generic uploadAsset in two spec-mandated ways:
// the reservation carries the appEventAssetType attribute, and the commit sends
// only uploaded=true (AppEventScreenshot/VideoClipUpdateRequest have no
// sourceFileChecksum attribute).
func evtUploadAsset(ctx context.Context, c *api.Client, resType, locID, assetType, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	fileName := filepath.Base(filePath)
	reserved, err := c.Post(ctx, "/v1/"+resType, api.Body{
		Data: api.Resource{
			Type: resType,
			Attributes: map[string]any{
				"fileName":          fileName,
				"fileSize":          len(data),
				"appEventAssetType": assetType,
			},
			Relationships: map[string]json.RawMessage{"appEventLocalization": api.Rel("appEventLocalizations", locID)},
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
		return "", fmt.Errorf("reservation for %s returned no upload operations: %w", fileName, err)
	}
	if err := c.Upload(ctx, ops, data); err != nil {
		return "", err
	}
	_, err = c.Patch(ctx, "/v1/"+resType+"/"+reserved.ID, api.Body{
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

var evtUploadScreenshotCmd = &cobra.Command{
	Use:   "upload-screenshot",
	Short: "Upload an event screenshot for a locale and display target",
	Long: `Upload an image to the event localization. --display is the appEventAssetType:
EVENT_CARD (the card shown on the App Store) or EVENT_DETAILS_PAGE.`,
	Example: `  asc events upload-screenshot --event-id <id> --locale ja \
    --display EVENT_CARD --file app-store/event-card.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		locID, err := evtLocalizationID(ctx, c, evtEventID, evtLocale)
		if err != nil {
			return err
		}
		id, err := evtUploadAsset(ctx, c, "appEventScreenshots", locID, evtDisplay, evtFile)
		if err != nil {
			return fmt.Errorf("%s: %w", evtFile, err)
		}
		fmt.Printf("uploaded %s -> %s\n", filepath.Base(evtFile), id)
		return nil
	},
}

var evtUploadVideoCmd = &cobra.Command{
	Use:   "upload-video",
	Short: "Upload an event video clip for a locale and display target",
	Long: `Upload a video clip to the event localization. --display is the
appEventAssetType: EVENT_CARD or EVENT_DETAILS_PAGE.`,
	Example: `  asc events upload-video --event-id <id> --locale ja \
    --display EVENT_DETAILS_PAGE --file app-store/event-clip.mp4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		locID, err := evtLocalizationID(ctx, c, evtEventID, evtLocale)
		if err != nil {
			return err
		}
		id, err := evtUploadAsset(ctx, c, "appEventVideoClips", locID, evtDisplay, evtFile)
		if err != nil {
			return fmt.Errorf("%s: %w", evtFile, err)
		}
		fmt.Printf("uploaded %s -> %s\n", filepath.Base(evtFile), id)
		return nil
	},
}

var evtListAssetsCmd = &cobra.Command{
	Use:   "list-assets",
	Short: "List uploaded screenshots and video clips for a locale",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		locs, err := c.List(ctx, "/v1/appEvents/"+evtEventID+"/localizations?limit=50")
		if err != nil {
			return err
		}
		loc := findByAttr(locs, "locale", evtLocale)
		if loc == nil {
			fmt.Printf("No localization for %s.\n", evtLocale)
			return nil
		}
		deliveryState := func(r *api.Resource) string {
			if ds, ok := r.Attributes["assetDeliveryState"].(map[string]any); ok {
				if st, ok := ds["state"].(string); ok {
					return st
				}
			}
			return ""
		}
		shots, err := c.List(ctx, "/v1/appEventLocalizations/"+loc.ID+"/appEventScreenshots?limit=50")
		if err != nil {
			return err
		}
		clips, err := c.List(ctx, "/v1/appEventLocalizations/"+loc.ID+"/appEventVideoClips?limit=50")
		if err != nil {
			return err
		}
		fmt.Printf("%d screenshot(s)\n", len(shots))
		for i := range shots {
			s := &shots[i]
			fmt.Printf("  %s  %s  %s  %s\n", s.ID, s.Str("appEventAssetType"), s.Str("fileName"), deliveryState(s))
		}
		fmt.Printf("%d video clip(s)\n", len(clips))
		for i := range clips {
			v := &clips[i]
			fmt.Printf("  %s  %s  %s  %s\n", v.ID, v.Str("appEventAssetType"), v.Str("fileName"), deliveryState(v))
		}
		return nil
	},
}

var evtDeleteScreenshotCmd = &cobra.Command{
	Use:   "delete-screenshot",
	Short: "Delete an event screenshot by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appEventScreenshots/"+evtAssetID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var evtDeleteVideoCmd = &cobra.Command{
	Use:   "delete-video",
	Short: "Delete an event video clip by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appEventVideoClips/"+evtAssetID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{evtListCmd, evtCreateCmd} {
		sub.Flags().StringVar(&evtApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	evtListCmd.Flags().StringVar(&evtState, "state", "", "filter by eventState, e.g. DRAFT, READY_FOR_REVIEW, PUBLISHED")

	for _, sub := range []*cobra.Command{evtCreateCmd, evtUpdateCmd} {
		sub.Flags().StringVar(&evtRefName, "reference-name", "", "internal reference name")
		sub.Flags().StringVar(&evtBadge, "badge", "", "badge: LIVE_EVENT, PREMIERE, CHALLENGE, COMPETITION, NEW_SEASON, MAJOR_UPDATE, SPECIAL_EVENT")
		sub.Flags().StringVar(&evtDeepLink, "deep-link", "", "deep link URL opened from the event")
		sub.Flags().StringVar(&evtPurchaseReq, "purchase-requirement", "", "purchase requirement, e.g. NO_COST_ASSOCIATED, IN_APP_PURCHASE, SUBSCRIPTION")
		sub.Flags().StringVar(&evtPrimaryLoc, "primary-locale", "", "primary locale, e.g. ja / en-US")
		sub.Flags().StringVar(&evtPriority, "priority", "", "priority: HIGH or NORMAL")
		sub.Flags().StringVar(&evtPurpose, "purpose", "", "purpose: APPROPRIATE_FOR_ALL_USERS, ATTRACT_NEW_USERS, KEEP_ACTIVE_USERS_INFORMED, BRING_BACK_LAPSED_USERS")
		sub.Flags().StringVar(&evtTerritories, "territories", "", "comma-separated territory ids for the schedule, e.g. JPN,USA (default: all territories)")
		sub.Flags().StringVar(&evtPublishStart, "publish-start", "", "when the event page becomes visible (ISO 8601)")
		sub.Flags().StringVar(&evtStart, "start", "", "event start (ISO 8601)")
		sub.Flags().StringVar(&evtEnd, "end", "", "event end (ISO 8601)")
	}
	_ = evtCreateCmd.MarkFlagRequired("reference-name")

	for _, sub := range []*cobra.Command{evtUpdateCmd, evtDeleteCmd, evtLocalizeCmd, evtUploadScreenshotCmd, evtUploadVideoCmd, evtListAssetsCmd} {
		sub.Flags().StringVar(&evtEventID, "event-id", "", "appEvent id (required)")
		_ = sub.MarkFlagRequired("event-id")
	}
	for _, sub := range []*cobra.Command{evtLocalizeCmd, evtUploadScreenshotCmd, evtUploadVideoCmd, evtListAssetsCmd} {
		sub.Flags().StringVar(&evtLocale, "locale", "ja", "locale, e.g. ja / en-US")
	}

	evtLocalizeCmd.Flags().StringVar(&evtName, "name", "", "event name (max 30 chars)")
	evtLocalizeCmd.Flags().StringVar(&evtShortDesc, "short-description", "", "short description, max 50 chars (@file allowed)")
	evtLocalizeCmd.Flags().StringVar(&evtLongDesc, "long-description", "", "long description, max 120 chars (@file allowed)")

	for _, sub := range []*cobra.Command{evtUploadScreenshotCmd, evtUploadVideoCmd} {
		sub.Flags().StringVar(&evtFile, "file", "", "media file to upload (required)")
		sub.Flags().StringVar(&evtDisplay, "display", "", "asset target: EVENT_CARD or EVENT_DETAILS_PAGE (required)")
		_ = sub.MarkFlagRequired("file")
		_ = sub.MarkFlagRequired("display")
	}

	for _, sub := range []*cobra.Command{evtDeleteScreenshotCmd, evtDeleteVideoCmd} {
		sub.Flags().StringVar(&evtAssetID, "id", "", "asset id (required)")
		_ = sub.MarkFlagRequired("id")
	}

	eventsCmd.AddCommand(evtListCmd, evtCreateCmd, evtUpdateCmd, evtDeleteCmd, evtLocalizeCmd,
		evtUploadScreenshotCmd, evtUploadVideoCmd, evtListAssetsCmd, evtDeleteScreenshotCmd, evtDeleteVideoCmd)
	rootCmd.AddCommand(eventsCmd)
}
