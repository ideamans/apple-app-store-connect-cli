package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var bldCmd = &cobra.Command{
	Use:   "builds",
	Short: "Manage builds (list, expire, encryption, notify, upload binaries)",
}

var (
	bldApp         string
	bldBuild       string
	bldVersion     string
	bldLimit       int
	bldFile        string
	bldPlatform    string
	bldBuildNumber string
	bldEncryption  string
)

// bldResolveBuild accepts a build id (UUID form, returned as-is) or a build
// version string (the CFBundleVersion), which is looked up for the app.
func bldResolveBuild(ctx context.Context, c *api.Client, appRef, buildRef string) (string, error) {
	if buildRef == "" {
		return "", fmt.Errorf("--build is required (build id or build version string)")
	}
	if len(buildRef) == 36 && strings.Count(buildRef, "-") == 4 {
		return buildRef, nil // already a build id
	}
	if appRef == "" {
		return "", fmt.Errorf("--app is required when --build is a version string")
	}
	appID, err := resolveAppID(ctx, c, appRef)
	if err != nil {
		return "", err
	}
	builds, err := c.List(ctx, "/v1/builds?filter[app]="+appID+"&filter[version]="+url.QueryEscape(buildRef)+"&limit=1")
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", fmt.Errorf("no build with version %q for app %s (is it uploaded and processed?)", buildRef, appID)
	}
	return builds[0].ID, nil
}

// bldRelID returns the id of a to-one relationship on a fetched resource.
func bldRelID(r *api.Resource, name string) string {
	raw, ok := r.Relationships[name]
	if !ok {
		return ""
	}
	var rel struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &rel); err != nil {
		return ""
	}
	return rel.Data.ID
}

func bldBoolAttr(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		if v {
			return "true"
		}
		return "false"
	}
	return ""
}

var bldListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's builds, newest first",
	Example: `  asc builds list --app 6790641087
  asc builds list --app 6790641087 --version 1.2.0 --limit 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, bldApp)
		if err != nil {
			return err
		}
		if bldLimit < 1 {
			return fmt.Errorf("--limit must be at least 1")
		}
		pageLimit := bldLimit
		if pageLimit > 200 {
			pageLimit = 200
		}
		path := "/v1/builds?filter[app]=" + appID + "&sort=-uploadedDate&include=preReleaseVersion&limit=" + strconv.Itoa(pageLimit)
		if bldVersion != "" {
			path += "&filter[preReleaseVersion.version]=" + url.QueryEscape(bldVersion)
		}
		// Decode manually to keep the "included" preReleaseVersions that
		// c.List would drop.
		var builds []api.Resource
		pre := map[string]string{}
		next := path
		for next != "" && len(builds) < bldLimit {
			raw, err := c.Do(ctx, http.MethodGet, next, nil)
			if err != nil {
				return err
			}
			var doc struct {
				Data     []api.Resource `json:"data"`
				Included []api.Resource `json:"included"`
				Links    struct {
					Next string `json:"next"`
				} `json:"links"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				return err
			}
			for i := range doc.Included {
				if doc.Included[i].Type == "preReleaseVersions" {
					pre[doc.Included[i].ID] = doc.Included[i].Str("version")
				}
			}
			builds = append(builds, doc.Data...)
			next = doc.Links.Next
		}
		if len(builds) > bldLimit {
			builds = builds[:bldLimit]
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "BUILD\tPRERELEASE\tSTATE\tEXPIRED\tUPLOADED\tID")
		for i := range builds {
			b := &builds[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				b.Str("version"), pre[bldRelID(b, "preReleaseVersion")], b.Str("processingState"),
				bldBoolAttr(b, "expired"), b.Str("uploadedDate"), b.ID)
		}
		return w.Flush()
	},
}

var bldShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a build's full details",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := bldResolveBuild(ctx, c, bldApp, bldBuild)
		if err != nil {
			return err
		}
		build, _, err := c.Get(ctx, "/v1/builds/"+buildID)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(build, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var bldExpireCmd = &cobra.Command{
	Use:   "expire",
	Short: "Expire a build (removes it from TestFlight)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := bldResolveBuild(ctx, c, bldApp, bldBuild)
		if err != nil {
			return err
		}
		_, err = c.Patch(ctx, "/v1/builds/"+buildID, api.Body{
			Data: api.Resource{Type: "builds", ID: buildID, Attributes: map[string]any{"expired": true}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Build %s expired.\n", buildID)
		return nil
	},
}

var bldEncryptionCmd = &cobra.Command{
	Use:     "encryption",
	Short:   "Set a build's export compliance (usesNonExemptEncryption)",
	Example: `  asc builds encryption --app 6790641087 --build 42 --uses-non-exempt-encryption false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		val, err := strconv.ParseBool(bldEncryption)
		if err != nil {
			return fmt.Errorf("--uses-non-exempt-encryption must be true or false")
		}
		ctx := cmd.Context()
		buildID, err := bldResolveBuild(ctx, c, bldApp, bldBuild)
		if err != nil {
			return err
		}
		_, err = c.Patch(ctx, "/v1/builds/"+buildID, api.Body{
			Data: api.Resource{Type: "builds", ID: buildID, Attributes: map[string]any{"usesNonExemptEncryption": val}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Build %s usesNonExemptEncryption set to %v.\n", buildID, val)
		return nil
	},
}

var bldNotifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Notify TestFlight testers that a build is available",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		buildID, err := bldResolveBuild(ctx, c, bldApp, bldBuild)
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/buildBetaNotifications", api.Body{
			Data: api.Resource{
				Type:          "buildBetaNotifications",
				Relationships: map[string]json.RawMessage{"build": api.Rel("builds", buildID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Notified testers about build %s.\n", buildID)
		return nil
	},
}

var bldPrereleaseCmd = &cobra.Command{
	Use:   "prerelease-versions",
	Short: "List the app's prerelease (TestFlight) versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, bldApp)
		if err != nil {
			return err
		}
		versions, err := c.List(ctx, "/v1/preReleaseVersions?filter[app]="+appID+"&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tPLATFORM\tID")
		for _, v := range versions {
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Str("version"), v.Str("platform"), v.ID)
		}
		return w.Flush()
	},
}

// bldUTIByExt maps binary file extensions to the UTIs accepted by buildUploadFiles.
var bldUTIByExt = map[string]string{
	".ipa": "com.apple.ipa",
	".pkg": "com.apple.pkg",
	".zip": "com.pkware.zip-archive",
}

var bldUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a build binary via the buildUploads API (no Transporter needed)",
	Long: `Upload an .ipa/.pkg/.zip build binary directly through the App Store
Connect API: a buildUpload is created for the app/version, a buildUploadFile is
reserved, the bytes are PUT to Apple's pre-signed URLs, and the upload is
committed with an MD5 checksum. Apple then processes the binary into a build.`,
	Example: `  asc builds upload --app 6790641087 --file build/App.ipa \
    --platform IOS --version 1.2.0 --build-number 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, bldApp)
		if err != nil {
			return err
		}
		uti, ok := bldUTIByExt[strings.ToLower(filepath.Ext(bldFile))]
		if !ok {
			return fmt.Errorf("unsupported file extension %q (supported: .ipa, .pkg, .zip)", filepath.Ext(bldFile))
		}
		data, err := os.ReadFile(bldFile)
		if err != nil {
			return err
		}
		upload, err := c.Post(ctx, "/v1/buildUploads", api.Body{
			Data: api.Resource{
				Type: "buildUploads",
				Attributes: map[string]any{
					"cfBundleShortVersionString": bldVersion,
					"cfBundleVersion":            bldBuildNumber,
					"platform":                   bldPlatform,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("1/4 Created build upload %s (%s %s build %s).\n", upload.ID, bldPlatform, bldVersion, bldBuildNumber)
		file, err := c.Post(ctx, "/v1/buildUploadFiles", api.Body{
			Data: api.Resource{
				Type: "buildUploadFiles",
				Attributes: map[string]any{
					"assetType": "ASSET",
					"fileName":  filepath.Base(bldFile),
					"fileSize":  len(data),
					"uti":       uti,
				},
				Relationships: map[string]json.RawMessage{"buildUpload": api.Rel("buildUploads", upload.ID)},
			},
		})
		if err != nil {
			return err
		}
		if c.DryRun {
			fmt.Println("Dry run: skipping byte upload and commit.")
			return nil
		}
		var ops []api.UploadOperation
		if err := file.DecodeAttr("uploadOperations", &ops); err != nil {
			return fmt.Errorf("buildUploadFile %s returned no upload operations: %w", file.ID, err)
		}
		fmt.Printf("2/4 Reserved upload file %s (%d bytes, %d operation(s)).\n", file.ID, len(data), len(ops))
		if err := c.Upload(ctx, ops, data); err != nil {
			return err
		}
		fmt.Println("3/4 Uploaded binary bytes.")
		_, err = c.Patch(ctx, "/v1/buildUploadFiles/"+file.ID, api.Body{
			Data: api.Resource{
				Type: "buildUploadFiles",
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
			return err
		}
		fmt.Printf("4/4 Committed upload %s.\n", upload.ID)
		fmt.Println("Apple is now processing the binary; this can take a while.")
		fmt.Printf("Watch for the new build with: asc builds list --app %s\n", bldApp)
		fmt.Printf("Check upload state with:     asc api \"/v1/buildUploads/%s\"\n", upload.ID)
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{bldListCmd, bldPrereleaseCmd, bldUploadCmd} {
		sub.Flags().StringVar(&bldApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{bldShowCmd, bldExpireCmd, bldEncryptionCmd, bldNotifyCmd} {
		sub.Flags().StringVar(&bldApp, "app", "", "app id or bundle id (required when --build is a version string)")
		sub.Flags().StringVar(&bldBuild, "build", "", "build id or build version string (required)")
		_ = sub.MarkFlagRequired("build")
	}

	bldListCmd.Flags().StringVar(&bldVersion, "version", "", "filter by marketing (prerelease) version, e.g. 1.2.0")
	bldListCmd.Flags().IntVar(&bldLimit, "limit", 50, "maximum number of builds to list")

	bldEncryptionCmd.Flags().StringVar(&bldEncryption, "uses-non-exempt-encryption", "", "true or false (required)")
	_ = bldEncryptionCmd.MarkFlagRequired("uses-non-exempt-encryption")

	bldUploadCmd.Flags().StringVar(&bldFile, "file", "", "build binary: .ipa, .pkg, or .zip (required)")
	bldUploadCmd.Flags().StringVar(&bldPlatform, "platform", "IOS", "platform: IOS, MAC_OS, TV_OS, VISION_OS")
	bldUploadCmd.Flags().StringVar(&bldVersion, "version", "", "CFBundleShortVersionString, e.g. 1.2.0 (required)")
	bldUploadCmd.Flags().StringVar(&bldBuildNumber, "build-number", "", "CFBundleVersion, e.g. 42 (required)")
	_ = bldUploadCmd.MarkFlagRequired("file")
	_ = bldUploadCmd.MarkFlagRequired("version")
	_ = bldUploadCmd.MarkFlagRequired("build-number")

	bldCmd.AddCommand(bldListCmd, bldShowCmd, bldExpireCmd, bldEncryptionCmd, bldNotifyCmd, bldPrereleaseCmd, bldUploadCmd)
	rootCmd.AddCommand(bldCmd)
}
