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

var (
	dsxApp          string
	dsxID           string
	dsxPackageID    string
	dsxVersionID    string
	dsxLimit        int
	dsxJSON         bool
	dsxPackageName  string
	dsxFingerprints string
)

// --- alt-distribution package versions / variants / deltas ------------------

var dsxVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List versions of an alternative distribution package (newest first)",
	Long: `List the versions of an alternative distribution package. Find package ids
with "asc alt-distribution packages show --version-id <appStoreVersion-id>".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/v1/alternativeDistributionPackages/%s/versions?limit=%d", dsxPackageID, dsxLimit)
		versions, err := xccListPage(cmd.Context(), c, path)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No package versions found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STATE\tVERSION\tURL EXPIRES\tID")
		for i := range versions {
			v := &versions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Str("state"), v.Str("version"), v.Str("urlExpirationDate"), v.ID)
		}
		return w.Flush()
	},
}

// dsxListFiles lists the variants or deltas of a package version. Both resources
// share the same attributes (url, urlExpirationDate, alternativeDistributionKeyBlob,
// fileChecksum).
func dsxListFiles(ctx context.Context, kind string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	files, err := c.List(ctx, "/v1/alternativeDistributionPackageVersions/"+dsxVersionID+"/"+kind+"?limit=200")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Printf("No %s found.\n", kind)
		return nil
	}
	if dsxJSON {
		b, err := json.MarshalIndent(files, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "CHECKSUM\tKEY BLOB\tURL EXPIRES\tID\tURL")
	for i := range files {
		f := &files[i]
		hasBlob := ""
		if f.Str("alternativeDistributionKeyBlob") != "" {
			hasBlob = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			f.Str("fileChecksum"), hasBlob, f.Str("urlExpirationDate"), f.ID, f.Str("url"))
	}
	return w.Flush()
}

var dsxVariantsCmd = &cobra.Command{
	Use:   "variants",
	Short: "List variants of an alternative distribution package version",
	Long: `List the variants of a package version with their download URL, expiration
date, and file checksum. Use --json to include the full resources (e.g. the
alternativeDistributionKeyBlob, which is too large for the table).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dsxListFiles(cmd.Context(), "variants")
	},
}

var dsxDeltasCmd = &cobra.Command{
	Use:   "deltas",
	Short: "List deltas of an alternative distribution package version",
	Long: `List the deltas of a package version with their download URL, expiration
date, and file checksum. Use --json to include the full resources (e.g. the
alternativeDistributionKeyBlob, which is too large for the table).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dsxListFiles(cmd.Context(), "deltas")
	},
}

// --- android-mapping ---------------------------------------------------------

var dsxAndroidMappingCmd = &cobra.Command{
	Use:   "android-mapping",
	Short: "Map Android apps to an iOS app (androidToIosAppMappingDetails)",
	Long: `Manage the Android-to-iOS app mapping details of an app. Each mapping names
an Android package and the SHA-256 fingerprints of its app signing key's
public certificates. An app can have multiple mappings (one per Android package).`,
}

// dsxAppMappings resolves the app and returns its androidToIosAppMappingDetails.
func dsxAppMappings(ctx context.Context, c *api.Client) (string, []api.Resource, error) {
	appID, err := resolveAppID(ctx, c, dsxApp)
	if err != nil {
		return "", nil, err
	}
	mappings, err := c.List(ctx, "/v1/apps/"+appID+"/androidToIosAppMappingDetails?limit=200")
	if err != nil {
		return "", nil, err
	}
	return appID, mappings, nil
}

var dsxAndroidMappingShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the Android-to-iOS app mappings of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, mappings, err := dsxAppMappings(cmd.Context(), c)
		if err != nil {
			return err
		}
		if len(mappings) == 0 {
			fmt.Println("No Android-to-iOS app mappings set.")
			return nil
		}
		b, err := json.MarshalIndent(mappings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

// dsxParseFingerprints accepts a comma- or newline-separated list, or @file.
func dsxParseFingerprints(s string) ([]string, error) {
	raw, err := valueOrFile(s)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' }) {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--fingerprints is empty")
	}
	return out, nil
}

var dsxAndroidMappingSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update the mapping for an Android package name",
	Long: `Set the Android-to-iOS mapping of an app for one Android package. If the app
already has a mapping with the same package name its fingerprints are updated;
otherwise a new mapping is created.`,
	Example: `  asc android-mapping set --app com.example.app \
    --package-name com.example.android \
    --fingerprints AA:BB:...,CC:DD:...   # or --fingerprints @fingerprints.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		fingerprints, err := dsxParseFingerprints(dsxFingerprints)
		if err != nil {
			return err
		}
		appID, mappings, err := dsxAppMappings(ctx, c)
		if err != nil {
			return err
		}
		if existing := findByAttr(mappings, "packageName", dsxPackageName); existing != nil {
			_, err = c.Patch(ctx, "/v1/androidToIosAppMappingDetails/"+existing.ID, api.Body{
				Data: api.Resource{
					Type: "androidToIosAppMappingDetails",
					ID:   existing.ID,
					Attributes: map[string]any{
						"appSigningKeyPublicCertificateSha256Fingerprints": fingerprints,
					},
				},
			})
			if err != nil {
				return err
			}
			fmt.Printf("Updated mapping %s for package %s.\n", existing.ID, dsxPackageName)
			return nil
		}
		created, err := c.Post(ctx, "/v1/androidToIosAppMappingDetails", api.Body{
			Data: api.Resource{
				Type: "androidToIosAppMappingDetails",
				Attributes: map[string]any{
					"packageName": dsxPackageName,
					"appSigningKeyPublicCertificateSha256Fingerprints": fingerprints,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created mapping %s for package %s.\n", created.ID, dsxPackageName)
		return nil
	},
}

var dsxAndroidMappingDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an Android-to-iOS app mapping",
	Long: `Delete a mapping by --id, or by --app (plus --package-name when the app has
more than one mapping).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if (dsxID == "") == (dsxApp == "") {
			return fmt.Errorf("exactly one of --id or --app is required")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		id := dsxID
		if id == "" {
			_, mappings, err := dsxAppMappings(ctx, c)
			if err != nil {
				return err
			}
			if dsxPackageName != "" {
				m := findByAttr(mappings, "packageName", dsxPackageName)
				if m == nil {
					return fmt.Errorf("no mapping with package name %q", dsxPackageName)
				}
				id = m.ID
			} else {
				switch len(mappings) {
				case 0:
					return fmt.Errorf("the app has no Android-to-iOS app mappings")
				case 1:
					id = mappings[0].ID
				default:
					var names []string
					for i := range mappings {
						names = append(names, mappings[i].Str("packageName"))
					}
					return fmt.Errorf("the app has %d mappings (%s); pass --package-name", len(mappings), strings.Join(names, ", "))
				}
			}
		}
		if err := c.Delete(ctx, "/v1/androidToIosAppMappingDetails/"+id); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

func init() {
	dsxVersionsCmd.Flags().StringVar(&dsxPackageID, "package-id", "", "alternativeDistributionPackage id (required)")
	dsxVersionsCmd.Flags().IntVar(&dsxLimit, "limit", 50, "maximum number of versions to list (max 200)")
	_ = dsxVersionsCmd.MarkFlagRequired("package-id")

	for _, sub := range []*cobra.Command{dsxVariantsCmd, dsxDeltasCmd} {
		sub.Flags().StringVar(&dsxVersionID, "version-id", "", "alternativeDistributionPackageVersion id (required)")
		sub.Flags().BoolVar(&dsxJSON, "json", false, "print the full resources as JSON instead of a table")
		_ = sub.MarkFlagRequired("version-id")
	}

	dstAltCmd.AddCommand(dsxVersionsCmd, dsxVariantsCmd, dsxDeltasCmd)

	dsxAndroidMappingShowCmd.Flags().StringVar(&dsxApp, "app", "", "app id or bundle id (required)")
	_ = dsxAndroidMappingShowCmd.MarkFlagRequired("app")

	dsxAndroidMappingSetCmd.Flags().StringVar(&dsxApp, "app", "", "app id or bundle id (required)")
	dsxAndroidMappingSetCmd.Flags().StringVar(&dsxPackageName, "package-name", "", "Android package name (required)")
	dsxAndroidMappingSetCmd.Flags().StringVar(&dsxFingerprints, "fingerprints", "",
		"SHA-256 fingerprints of the Android app signing key's public certificates, comma-separated or @file (required)")
	for _, f := range []string{"app", "package-name", "fingerprints"} {
		_ = dsxAndroidMappingSetCmd.MarkFlagRequired(f)
	}

	dsxAndroidMappingDeleteCmd.Flags().StringVar(&dsxID, "id", "", "androidToIosAppMappingDetail id")
	dsxAndroidMappingDeleteCmd.Flags().StringVar(&dsxApp, "app", "", "app id or bundle id")
	dsxAndroidMappingDeleteCmd.Flags().StringVar(&dsxPackageName, "package-name", "", "Android package name to delete (when the app has several mappings)")

	dsxAndroidMappingCmd.AddCommand(dsxAndroidMappingShowCmd, dsxAndroidMappingSetCmd, dsxAndroidMappingDeleteCmd)
	rootCmd.AddCommand(dsxAndroidMappingCmd)
}
