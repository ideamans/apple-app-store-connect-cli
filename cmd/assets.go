package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Upload and manage media assets (screenshots) via the reserve→upload→commit flow",
	Long: `Upload media assets to App Store Connect. Uploading is the one part of the
submission flow that "asc api" cannot do on its own: each asset is reserved
(POST), its bytes are PUT to a per-asset pre-signed URL, then the upload is
committed (PATCH uploaded=true with an MD5 checksum). These commands do all
three steps.

A 2xx commit does NOT mean the asset was accepted: Apple validates
asynchronously and rejections (e.g. IMAGE_INCORRECT_DIMENSIONS) only appear in
assetDeliveryState afterwards. These commands therefore poll until validation
reaches COMPLETE and exit non-zero on FAILED with the error codes.`,
}

// assetSpec describes one asset upload target.
type assetSpec struct {
	reserveType string // e.g. "appScreenshots"
	relName     string // relationship name, e.g. "appScreenshotSet"
	relType     string // relationship resource type, e.g. "appScreenshotSets"
	relID       string
	filePath    string
}

// uploadAsset runs reserve → upload → commit → validation wait and returns the
// new asset id.
func uploadAsset(ctx context.Context, c *api.Client, spec assetSpec) (string, error) {
	data, err := os.ReadFile(spec.filePath)
	if err != nil {
		return "", err
	}
	return uploadAssetBytes(ctx, c, spec, data, filepath.Base(spec.filePath))
}

// uploadAssetBytes is uploadAsset for in-memory data (e.g. auto-fitted images).
// After committing it waits for Apple's asynchronous validation: a 2xx commit
// alone does not mean the asset was accepted (see waitAssetDelivery).
func uploadAssetBytes(ctx context.Context, c *api.Client, spec assetSpec, data []byte, fileName string) (string, error) {
	reserved, err := c.Post(ctx, "/v1/"+spec.reserveType, api.Body{
		Data: api.Resource{
			Type:          spec.reserveType,
			Attributes:    map[string]any{"fileName": fileName, "fileSize": len(data)},
			Relationships: map[string]json.RawMessage{spec.relName: api.Rel(spec.relType, spec.relID)},
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
	_, err = c.Patch(ctx, "/v1/"+spec.reserveType+"/"+reserved.ID, api.Body{
		Data: api.Resource{
			Type: spec.reserveType,
			ID:   reserved.ID,
			Attributes: map[string]any{
				"uploaded":           true,
				"sourceFileChecksum": api.MD5Hex(data),
			},
		},
	})
	if err != nil {
		return "", err
	}
	if err := waitAssetDelivery(ctx, c, "/v1/"+spec.reserveType+"/"+reserved.ID); err != nil {
		return "", fmt.Errorf("%s: %w", fileName, err)
	}
	return reserved.ID, nil
}

// versionLocalizationID returns the localization id for locale on the editable
// version, creating an empty one if it does not exist.
func versionLocalizationID(ctx context.Context, c *api.Client, verID, locale string) (string, error) {
	locs, err := c.List(ctx, "/v1/appStoreVersions/"+verID+"/appStoreVersionLocalizations?limit=50")
	if err != nil {
		return "", err
	}
	if loc := findByAttr(locs, "locale", locale); loc != nil {
		return loc.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appStoreVersionLocalizations", api.Body{
		Data: api.Resource{
			Type:          "appStoreVersionLocalizations",
			Attributes:    map[string]any{"locale": locale},
			Relationships: map[string]json.RawMessage{"appStoreVersion": api.Rel("appStoreVersions", verID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

var (
	asApp     string
	asVersion string
	asLocale  string
	asDisplay string
	asFiles   []string
	asDelete  string
)

var assetsUploadScreenshotCmd = &cobra.Command{
	Use:   "upload-screenshot",
	Short: "Upload one or more screenshots for a locale and display type",
	Long: `Upload screenshots to the editable version. Files are appended to the
screenshot set for the given display type in filename order, so pass them in the
order you want them shown.

Common --display values: APP_IPHONE_67 (6.5"/6.7"/6.9"), APP_IPHONE_61,
APP_IPHONE_55, APP_IPAD_PRO_129, APP_IPAD_PRO_3GEN_129. Full list:
asc api "/v1/appScreenshotSets?..." or Apple's ScreenshotDisplayType docs.`,
	Example: `  asc assets upload-screenshot --app 6790641087 --locale ja --display APP_IPHONE_67 \
    --file app-store/screenshots/01-hero.png --file app-store/screenshots/02.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, asApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, asVersion)
		if err != nil {
			return err
		}
		locID, err := versionLocalizationID(ctx, c, ver.ID, asLocale)
		if err != nil {
			return err
		}
		setID, err := screenshotSetID(ctx, c, locID, asDisplay)
		if err != nil {
			return err
		}
		for _, f := range asFiles {
			id, err := uploadAsset(ctx, c, assetSpec{
				reserveType: "appScreenshots",
				relName:     "appScreenshotSet",
				relType:     "appScreenshotSets",
				relID:       setID,
				filePath:    f,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			fmt.Printf("uploaded %s -> %s\n", filepath.Base(f), id)
		}
		return nil
	},
}

// screenshotSetID finds or creates the screenshot set for a localization + display type.
func screenshotSetID(ctx context.Context, c *api.Client, locID, display string) (string, error) {
	sets, err := c.List(ctx, "/v1/appStoreVersionLocalizations/"+locID+"/appScreenshotSets?limit=50")
	if err != nil {
		return "", err
	}
	if set := findByAttr(sets, "screenshotDisplayType", display); set != nil {
		return set.ID, nil
	}
	created, err := c.Post(ctx, "/v1/appScreenshotSets", api.Body{
		Data: api.Resource{
			Type:          "appScreenshotSets",
			Attributes:    map[string]any{"screenshotDisplayType": display},
			Relationships: map[string]json.RawMessage{"appStoreVersionLocalization": api.Rel("appStoreVersionLocalizations", locID)},
		},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

var assetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List screenshot sets and screenshots for a locale",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, asApp)
		if err != nil {
			return err
		}
		ver, err := editableVersion(ctx, c, appID, asVersion)
		if err != nil {
			return err
		}
		locs, err := c.List(ctx, "/v1/appStoreVersions/"+ver.ID+"/appStoreVersionLocalizations?limit=50")
		if err != nil {
			return err
		}
		loc := findByAttr(locs, "locale", asLocale)
		if loc == nil {
			fmt.Printf("No localization for %s.\n", asLocale)
			return nil
		}
		sets, err := c.List(ctx, "/v1/appStoreVersionLocalizations/"+loc.ID+"/appScreenshotSets?limit=50")
		if err != nil {
			return err
		}
		for _, set := range sets {
			shots, err := c.List(ctx, "/v1/appScreenshotSets/"+set.ID+"/appScreenshots?limit=50")
			if err != nil {
				return err
			}
			fmt.Printf("%s (%s)  %d screenshot(s)\n", set.Str("screenshotDisplayType"), set.ID, len(shots))
			for _, s := range shots {
				state := ""
				if ds, ok := s.Attributes["assetDeliveryState"].(map[string]any); ok {
					if st, ok := ds["state"].(string); ok {
						state = st
					}
				}
				fmt.Printf("  %s  %s  %s\n", s.ID, s.Str("fileName"), state)
			}
		}
		return nil
	},
}

var assetsDeleteScreenshotCmd = &cobra.Command{
	Use:   "delete-screenshot",
	Short: "Delete a screenshot by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/appScreenshots/"+asDelete); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{assetsUploadScreenshotCmd, assetsListCmd} {
		sub.Flags().StringVar(&asApp, "app", "", "app id or bundle id (required)")
		sub.Flags().StringVar(&asVersion, "version", "", "version string (default: the editable version)")
		sub.Flags().StringVar(&asLocale, "locale", "ja", "locale, e.g. ja / en-US")
		_ = sub.MarkFlagRequired("app")
	}
	assetsUploadScreenshotCmd.Flags().StringVar(&asDisplay, "display", "", "screenshot display type, e.g. APP_IPHONE_67 (required)")
	assetsUploadScreenshotCmd.Flags().StringArrayVar(&asFiles, "file", nil, "screenshot file (repeatable, applied in order)")
	_ = assetsUploadScreenshotCmd.MarkFlagRequired("display")
	_ = assetsUploadScreenshotCmd.MarkFlagRequired("file")

	assetsDeleteScreenshotCmd.Flags().StringVar(&asDelete, "id", "", "appScreenshot id (required)")
	_ = assetsDeleteScreenshotCmd.MarkFlagRequired("id")

	assetsCmd.AddCommand(assetsUploadScreenshotCmd, assetsListCmd, assetsDeleteScreenshotCmd)
	rootCmd.AddCommand(assetsCmd)
}
