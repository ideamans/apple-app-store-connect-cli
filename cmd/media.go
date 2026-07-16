package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	medApp       string
	medVersion   string
	medLocale    string
	medDisplay   string
	medFiles     []string
	medFile      string
	medFrameTime string
	medID        string
)

// medPreviewSetID finds or creates the preview set for a localization + preview type.
func medPreviewSetID(ctx context.Context, c *api.Client, locID, previewType string) (string, error) {
	sets, err := c.List(ctx, "/v1/appStoreVersionLocalizations/"+locID+"/appPreviewSets?limit=50")
	if err != nil {
		return "", err
	}
	if set := findByAttr(sets, "previewType", previewType); set != nil {
		return set.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appPreviewSets", api.Body{
		Data: api.Resource{
			Type:          "appPreviewSets",
			Attributes:    map[string]any{"previewType": previewType},
			Relationships: map[string]json.RawMessage{"appStoreVersionLocalization": api.Rel("appStoreVersionLocalizations", locID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// medDeliveryState extracts the delivery state from a media resource. On app
// previews assetDeliveryState is deprecated in favor of videoDeliveryState, so
// both are consulted.
func medDeliveryState(r *api.Resource) string {
	for _, key := range []string{"assetDeliveryState", "videoDeliveryState"} {
		if ds, ok := r.Attributes[key].(map[string]any); ok {
			if st, ok := ds["state"].(string); ok && st != "" {
				return st
			}
		}
	}
	return ""
}

// medReviewDetail returns the App Review detail for the editable version,
// erroring with a hint when none has been created yet.
func medReviewDetail(ctx context.Context, c *api.Client, appRef, versionString string) (*api.Resource, error) {
	appID, err := resolveAppID(ctx, c, appRef)
	if err != nil {
		return nil, err
	}
	ver, err := editableVersion(ctx, c, appID, versionString)
	if err != nil {
		return nil, err
	}
	detail, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreReviewDetail")
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.ID == "" {
		return nil, fmt.Errorf("version %s has no App Review detail yet; create one first with `asc review-detail set`", ver.Str("versionString"))
	}
	return detail, nil
}

var medUploadPreviewCmd = &cobra.Command{
	Use:   "upload-preview",
	Short: "Upload one or more app preview videos for a locale and preview type",
	Long: `Upload app preview videos to the editable version. Files are appended to the
preview set for the given preview type in the order passed.

Common --display values: IPHONE_67, IPHONE_61, IPHONE_65, IPHONE_58, IPHONE_55,
IPAD_PRO_3GEN_129, IPAD_PRO_3GEN_11, DESKTOP, APPLE_TV, APPLE_VISION_PRO.
Note preview types have no APP_ prefix, unlike screenshot display types.`,
	Example: `  asc assets upload-preview --app 6790641087 --locale ja --display IPHONE_67 \
    --file app-store/previews/demo.mp4 --frame-time-code 00:00:05:00`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, medApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, medVersion)
		if err != nil {
			return err
		}
		locID, err := versionLocalizationID(ctx, c, ver.ID, medLocale)
		if err != nil {
			return err
		}
		setID, err := medPreviewSetID(ctx, c, locID, medDisplay)
		if err != nil {
			return err
		}
		for _, f := range medFiles {
			id, err := uploadAsset(ctx, c, assetSpec{
				reserveType: "appPreviews",
				relName:     "appPreviewSet",
				relType:     "appPreviewSets",
				relID:       setID,
				filePath:    f,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			if cmd.Flags().Changed("frame-time-code") {
				_, err = c.Patch(ctx, "/v1/appPreviews/"+id, api.Body{
					Data: api.Resource{
						Type:       "appPreviews",
						ID:         id,
						Attributes: map[string]any{"previewFrameTimeCode": medFrameTime},
					},
				})
				if err != nil {
					// The API rejects previewFrameTimeCode while the video is still
					// processing; the upload itself succeeded, so warn instead of failing.
					fmt.Fprintf(os.Stderr, "warning: %s: setting frame time code failed (retry once the video is processed): %v\n", f, err)
				}
			}
			fmt.Printf("uploaded %s -> %s\n", filepath.Base(f), id)
		}
		return nil
	},
}

var medListPreviewsCmd = &cobra.Command{
	Use:   "list-previews",
	Short: "List app preview sets and previews for a locale",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, medApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, medVersion)
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreVersionLocalizations?limit=50")
		if err != nil {
			return err
		}
		loc := findByAttr(locs, "locale", medLocale)
		if loc == nil {
			fmt.Printf("No localization for %s.\n", medLocale)
			return nil
		}
		sets, err := c.List(ctx, "/v1/appStoreVersionLocalizations/"+loc.ID+"/appPreviewSets?limit=50")
		if err != nil {
			return err
		}
		if len(sets) == 0 {
			fmt.Printf("No preview sets for %s.\n", medLocale)
			return nil
		}
		for _, set := range sets {
			previews, err := c.List(ctx, "/v1/appPreviewSets/"+set.ID+"/appPreviews?limit=50")
			if err != nil {
				return err
			}
			fmt.Printf("%s (%s)  %d preview(s)\n", set.Str("previewType"), set.ID, len(previews))
			for i := range previews {
				p := &previews[i]
				fmt.Printf("  %s  %s  %s\n", p.ID, p.Str("fileName"), medDeliveryState(p))
			}
		}
		return nil
	},
}

var medDeletePreviewCmd = &cobra.Command{
	Use:   "delete-preview",
	Short: "Delete an app preview by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appPreviews/"+medID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var medDeleteScreenshotSetCmd = &cobra.Command{
	Use:   "delete-screenshot-set",
	Short: "Delete a screenshot set (and its screenshots) by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appScreenshotSets/"+medID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var medDeletePreviewSetCmd = &cobra.Command{
	Use:   "delete-preview-set",
	Short: "Delete a preview set (and its previews) by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appPreviewSets/"+medID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var medUploadRoutingCoverageCmd = &cobra.Command{
	Use:   "upload-routing-coverage",
	Short: "Upload a GeoJSON routing app coverage file for the editable version",
	Long: `Upload the .geojson file that defines where a routing app offers coverage.
A version has at most one routing app coverage; any existing one is deleted
and replaced.`,
	Example: `  asc assets upload-routing-coverage --app 6790641087 --file coverage.geojson`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, medApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, medVersion)
		if err != nil {
			return err
		}
		existing, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+ver.ID+"/routingAppCoverage")
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != "" {
			if err := c.Delete(ctx, "/v1/routingAppCoverages/"+existing.ID); err != nil {
				return err
			}
			if !c.DryRun {
				fmt.Printf("Deleted existing coverage %s.\n", existing.ID)
			}
		}
		id, err := uploadAsset(ctx, c, assetSpec{
			reserveType: "routingAppCoverages",
			relName:     "appStoreVersion",
			relType:     "appStoreVersions",
			relID:       ver.ID,
			filePath:    medFile,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", medFile, err)
		}
		fmt.Printf("uploaded %s -> %s\n", filepath.Base(medFile), id)
		return nil
	},
}

var medAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Upload a file attachment for App Review",
	Long: `Upload an attachment (screenshot, video, document) to the version's App
Review detail. The review detail must exist; create it with
"asc review-detail set" first.`,
	Example: `  asc review-detail attach --app 6790641087 --file app-store/demo-walkthrough.mp4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detail, err := medReviewDetail(ctx, c, medApp, medVersion)
		if err != nil {
			return err
		}
		id, err := uploadAsset(ctx, c, assetSpec{
			reserveType: "appStoreReviewAttachments",
			relName:     "appStoreReviewDetail",
			relType:     "appStoreReviewDetails",
			relID:       detail.ID,
			filePath:    medFile,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", medFile, err)
		}
		fmt.Printf("uploaded %s -> %s\n", filepath.Base(medFile), id)
		return nil
	},
}

var medAttachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "List App Review attachments for the editable version",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		detail, err := medReviewDetail(ctx, c, medApp, medVersion)
		if err != nil {
			return err
		}
		attachments, err := c.List(ctx, "/v1/appStoreReviewDetails/"+detail.ID+"/appStoreReviewAttachments?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "FILE\tSIZE\tSTATE\tID")
		for i := range attachments {
			a := &attachments[i]
			size := ""
			if v, ok := a.Attributes["fileSize"]; ok {
				size = fmt.Sprintf("%v", v)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Str("fileName"), size, medDeliveryState(a), a.ID)
		}
		return w.Flush()
	},
}

var medDeleteAttachmentCmd = &cobra.Command{
	Use:   "delete-attachment",
	Short: "Delete an App Review attachment by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appStoreReviewAttachments/"+medID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{medUploadPreviewCmd, medListPreviewsCmd, medUploadRoutingCoverageCmd, medAttachCmd, medAttachmentsCmd} {
		sub.Flags().StringVar(&medApp, "app", "", "app id or bundle id (required)")
		sub.Flags().StringVar(&medVersion, "version", "", "version string (default: the editable version)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{medUploadPreviewCmd, medListPreviewsCmd} {
		sub.Flags().StringVar(&medLocale, "locale", "ja", "locale, e.g. ja / en-US")
	}
	medUploadPreviewCmd.Flags().StringVar(&medDisplay, "display", "", "preview type, e.g. IPHONE_67 (required)")
	medUploadPreviewCmd.Flags().StringArrayVar(&medFiles, "file", nil, "preview video file (repeatable, applied in order)")
	medUploadPreviewCmd.Flags().StringVar(&medFrameTime, "frame-time-code", "", "poster frame time code, e.g. 00:00:05:00")
	_ = medUploadPreviewCmd.MarkFlagRequired("display")
	_ = medUploadPreviewCmd.MarkFlagRequired("file")

	medUploadRoutingCoverageCmd.Flags().StringVar(&medFile, "file", "", "GeoJSON coverage file (required)")
	_ = medUploadRoutingCoverageCmd.MarkFlagRequired("file")

	medAttachCmd.Flags().StringVar(&medFile, "file", "", "attachment file (required)")
	_ = medAttachCmd.MarkFlagRequired("file")

	medDeletePreviewCmd.Flags().StringVar(&medID, "id", "", "appPreview id (required)")
	medDeleteScreenshotSetCmd.Flags().StringVar(&medID, "id", "", "appScreenshotSet id (required)")
	medDeletePreviewSetCmd.Flags().StringVar(&medID, "id", "", "appPreviewSet id (required)")
	medDeleteAttachmentCmd.Flags().StringVar(&medID, "id", "", "appStoreReviewAttachment id (required)")
	for _, sub := range []*cobra.Command{medDeletePreviewCmd, medDeleteScreenshotSetCmd, medDeletePreviewSetCmd, medDeleteAttachmentCmd} {
		_ = sub.MarkFlagRequired("id")
	}

	assetsCmd.AddCommand(medUploadPreviewCmd, medListPreviewsCmd, medDeletePreviewCmd,
		medUploadRoutingCoverageCmd, medDeleteScreenshotSetCmd, medDeletePreviewSetCmd)
	reviewCmd.AddCommand(medAttachCmd, medAttachmentsCmd, medDeleteAttachmentCmd)
}
