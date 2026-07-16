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

var bgaCmd = &cobra.Command{
	Use:   "background-assets",
	Short: "Manage Background Assets asset packs (list, create, archive, versions, upload, release status)",
	Long: `Manage Background Assets asset packs through the App Store Connect API.

An asset pack (backgroundAsset) belongs to an app and is identified by its
assetPackIdentifier. Each upload creates a new backgroundAssetVersion whose
version number Apple assigns automatically. Releases to internal beta,
external beta, and the App Store are separate read-only resources; App Store
and external beta releases are initiated by adding the backgroundAssetVersion
to a review submission (reviewSubmissionItems), while internal beta releases
happen automatically when a version finishes processing.`,
}

var (
	bgaApp       string
	bgaPackID    string
	bgaID        string
	bgaAssetID   string
	bgaVersionID string
	bgaFile      string
	bgaManifest  string
)

// bgaBoolAttr formats a boolean attribute, or "" when absent.
func bgaBoolAttr(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		if v {
			return "true"
		}
		return "false"
	}
	return ""
}

// bgaPlatforms joins the platforms array attribute of a backgroundAssetVersion.
func bgaPlatforms(r *api.Resource) string {
	var ps []string
	if err := r.DecodeAttr("platforms", &ps); err != nil {
		return ""
	}
	return strings.Join(ps, ",")
}

var bgaListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List the app's background asset packs",
	Example: `  asc background-assets list --app 6790641087`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, bgaApp)
		if err != nil {
			return err
		}
		assets, err := c.List(ctx, "/v1/apps/"+appID+"/backgroundAssets?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PACK_IDENTIFIER\tARCHIVED\tCREATED\tID")
		for i := range assets {
			a := &assets[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				a.Str("assetPackIdentifier"), bgaBoolAttr(a, "archived"), a.Str("createdDate"), a.ID)
		}
		return w.Flush()
	},
}

var bgaCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a background asset pack for an app",
	Example: `  asc background-assets create --app 6790641087 --pack-id com.example.app.levels`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, bgaApp)
		if err != nil {
			return err
		}
		asset, err := c.Post(ctx, "/v1/backgroundAssets", api.Body{
			Data: api.Resource{
				Type:          "backgroundAssets",
				Attributes:    map[string]any{"assetPackIdentifier": bgaPackID},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created background asset %s (%s).\n", asset.ID, bgaPackID)
		return nil
	},
}

// bgaSetArchived PATCHes the archived attribute of a background asset.
func bgaSetArchived(ctx context.Context, c *api.Client, id string, archived bool) error {
	_, err := c.Patch(ctx, "/v1/backgroundAssets/"+id, api.Body{
		Data: api.Resource{Type: "backgroundAssets", ID: id, Attributes: map[string]any{"archived": archived}},
	})
	return err
}

var bgaArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a background asset pack",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := bgaSetArchived(cmd.Context(), c, bgaID, true); err != nil {
			return err
		}
		fmt.Printf("Background asset %s archived.\n", bgaID)
		return nil
	},
}

var bgaUnarchiveCmd = &cobra.Command{
	Use:   "unarchive",
	Short: "Unarchive a background asset pack",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := bgaSetArchived(cmd.Context(), c, bgaID, false); err != nil {
			return err
		}
		fmt.Printf("Background asset %s unarchived.\n", bgaID)
		return nil
	},
}

var bgaVersionsCmd = &cobra.Command{
	Use:     "versions",
	Short:   "List a background asset pack's versions, newest first",
	Example: `  asc background-assets versions --asset-id 1234abcd-...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		versions, err := c.List(cmd.Context(), "/v1/backgroundAssets/"+bgaAssetID+"/versions?sort=-version&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tSTATE\tPLATFORMS\tCREATED\tID")
		for i := range versions {
			v := &versions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				v.Str("version"), v.Str("state"), bgaPlatforms(v), v.Str("createdDate"), v.ID)
		}
		return w.Flush()
	},
}

// bgaUploadFile runs the backgroundAssetUploadFiles reserve → PUT → commit flow
// for one file and returns the upload file id.
func bgaUploadFile(ctx context.Context, c *api.Client, versionID, path, assetType string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	file, err := c.Post(ctx, "/v1/backgroundAssetUploadFiles", api.Body{
		Data: api.Resource{
			Type: "backgroundAssetUploadFiles",
			Attributes: map[string]any{
				"assetType": assetType,
				"fileName":  filepath.Base(path),
				"fileSize":  len(data),
			},
			Relationships: map[string]json.RawMessage{
				"backgroundAssetVersion": api.Rel("backgroundAssetVersions", versionID),
			},
		},
	})
	if err != nil {
		return "", err
	}
	if c.DryRun {
		fmt.Printf("Dry run: skipping byte upload and commit for %s.\n", filepath.Base(path))
		return "dry-run", nil
	}
	var ops []api.UploadOperation
	if err := file.DecodeAttr("uploadOperations", &ops); err != nil {
		return "", fmt.Errorf("backgroundAssetUploadFile %s returned no upload operations: %w", file.ID, err)
	}
	if err := c.Upload(ctx, ops, data); err != nil {
		return "", err
	}
	_, err = c.Patch(ctx, "/v1/backgroundAssetUploadFiles/"+file.ID, api.Body{
		Data: api.Resource{
			Type: "backgroundAssetUploadFiles",
			ID:   file.ID,
			Attributes: map[string]any{
				"uploaded": true,
				"sourceFileChecksums": map[string]any{
					"file": map[string]any{"hash": api.MD5Hex(data), "algorithm": "MD5"},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	return file.ID, nil
}

var bgaUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a new version of an asset pack (create version, then reserve→PUT→commit)",
	Long: `Upload a new background asset version: a backgroundAssetVersion is created
for the asset pack (Apple assigns the version number), the asset pack file is
reserved as a backgroundAssetUploadFile, its bytes are PUT to Apple's
pre-signed URLs, and the upload is committed with an MD5 checksum. Pass
--manifest to also upload a MANIFEST file for the same version. Apple then
processes the version (see "background-assets versions").`,
	Example: `  asc background-assets upload --asset-id 1234abcd-... --file pack.aar`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		version, err := c.Post(ctx, "/v1/backgroundAssetVersions", api.Body{
			Data: api.Resource{
				Type: "backgroundAssetVersions",
				Relationships: map[string]json.RawMessage{
					"backgroundAsset": api.Rel("backgroundAssets", bgaAssetID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created background asset version %s (version %s).\n", version.ID, version.Str("version"))
		fileID, err := bgaUploadFile(ctx, c, version.ID, bgaFile, "ASSET")
		if err != nil {
			return fmt.Errorf("%s: %w", bgaFile, err)
		}
		fmt.Printf("Uploaded asset file %s -> %s.\n", filepath.Base(bgaFile), fileID)
		if bgaManifest != "" {
			manifestID, err := bgaUploadFile(ctx, c, version.ID, bgaManifest, "MANIFEST")
			if err != nil {
				return fmt.Errorf("%s: %w", bgaManifest, err)
			}
			fmt.Printf("Uploaded manifest file %s -> %s.\n", filepath.Base(bgaManifest), manifestID)
		}
		if c.DryRun {
			return nil
		}
		fmt.Println("Apple is now processing the version.")
		fmt.Printf("Check its state with: asc background-assets versions --asset-id %s\n", bgaAssetID)
		return nil
	},
}

// bgaTrackByType maps release resource types to the track names shown to users.
var bgaTrackByType = map[string]string{
	"backgroundAssetVersionInternalBetaReleases": "internal-beta",
	"backgroundAssetVersionExternalBetaReleases": "external-beta",
	"backgroundAssetVersionAppStoreReleases":     "app-store",
}

var bgaReleasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Show a version's release state on each track (internal-beta, external-beta, app-store)",
	Long: `Show the release resources attached to a backgroundAssetVersion. The App
Store Connect API exposes releases as read-only state: internal beta releases
are created automatically when a version finishes processing, and external
beta / App Store releases are created by submitting the backgroundAssetVersion
for review (reviewSubmissionItems with a backgroundAssetVersion relationship).`,
	Example: `  asc background-assets releases --version-id 5678efgh-...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, included, err := c.Get(cmd.Context(),
			"/v1/backgroundAssetVersions/"+bgaVersionID+"?include=internalBetaRelease,externalBetaRelease,appStoreRelease")
		if err != nil {
			return err
		}
		releases := map[string]*api.Resource{}
		for i := range included {
			if track, ok := bgaTrackByType[included[i].Type]; ok {
				releases[track] = &included[i]
			}
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TRACK\tSTATE\tID")
		for _, track := range []string{"internal-beta", "external-beta", "app-store"} {
			if r := releases[track]; r != nil {
				fmt.Fprintf(w, "%s\t%s\t%s\n", track, r.Str("state"), r.ID)
			} else {
				fmt.Fprintf(w, "%s\t-\t-\n", track)
			}
		}
		return w.Flush()
	},
}

func init() {
	for _, sub := range []*cobra.Command{bgaListCmd, bgaCreateCmd} {
		sub.Flags().StringVar(&bgaApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	bgaCreateCmd.Flags().StringVar(&bgaPackID, "pack-id", "", "assetPackIdentifier, e.g. com.example.app.levels (required)")
	_ = bgaCreateCmd.MarkFlagRequired("pack-id")

	for _, sub := range []*cobra.Command{bgaArchiveCmd, bgaUnarchiveCmd} {
		sub.Flags().StringVar(&bgaID, "id", "", "backgroundAsset id (required)")
		_ = sub.MarkFlagRequired("id")
	}

	for _, sub := range []*cobra.Command{bgaVersionsCmd, bgaUploadCmd} {
		sub.Flags().StringVar(&bgaAssetID, "asset-id", "", "backgroundAsset id (required)")
		_ = sub.MarkFlagRequired("asset-id")
	}
	bgaUploadCmd.Flags().StringVar(&bgaFile, "file", "", "asset pack file to upload (required)")
	bgaUploadCmd.Flags().StringVar(&bgaManifest, "manifest", "", "optional MANIFEST file to upload for the same version")
	_ = bgaUploadCmd.MarkFlagRequired("file")

	bgaReleasesCmd.Flags().StringVar(&bgaVersionID, "version-id", "", "backgroundAssetVersion id (required)")
	_ = bgaReleasesCmd.MarkFlagRequired("version-id")

	bgaCmd.AddCommand(bgaListCmd, bgaCreateCmd, bgaArchiveCmd, bgaUnarchiveCmd, bgaVersionsCmd, bgaUploadCmd, bgaReleasesCmd)
	rootCmd.AddCommand(bgaCmd)
}
