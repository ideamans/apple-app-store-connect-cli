package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	secApp          string
	secID           string
	secBuildID      string
	secFile         string
	secDescription  string
	secProprietary  bool
	secThirdParty   bool
	secFrenchStore  bool
	secDeviceFamily string
	secPublish      bool
)

// secBool formats a boolean attribute, or "" when absent.
func secBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// secPrintJSON prints a resource as indented JSON.
func secPrintJSON(r *api.Resource) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// --- encryption ------------------------------------------------------------

var secEncryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Manage app encryption declarations (export compliance)",
}

var secEncryptionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List encryption declarations of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, secApp)
		if err != nil {
			return err
		}
		decls, err := c.List(ctx, "/v1/appEncryptionDeclarations?filter[app]="+appID+"&limit=200")
		if err != nil {
			return err
		}
		if len(decls) == 0 {
			fmt.Println("No encryption declarations found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STATE\tEXEMPT\tPROPRIETARY\tTHIRD PARTY\tFRENCH STORE\tCREATED\tID")
		for i := range decls {
			d := &decls[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				d.Str("appEncryptionDeclarationState"),
				secBool(d, "exempt"),
				secBool(d, "containsProprietaryCryptography"),
				secBool(d, "containsThirdPartyCryptography"),
				secBool(d, "availableOnFrenchStore"),
				d.Str("createdDate"),
				d.ID)
		}
		return w.Flush()
	},
}

var secEncryptionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an encryption declaration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		decl, _, err := c.Get(cmd.Context(), "/v1/appEncryptionDeclarations/"+secID)
		if err != nil {
			return err
		}
		return secPrintJSON(decl)
	},
}

var secEncryptionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an encryption declaration for an app",
	Long: `Create an app encryption declaration. --description explains how the app uses
encryption. Pass --proprietary and/or --third-party when the app contains
proprietary or third-party cryptography (beyond Apple's OS crypto), and
--available-on-french-store when the app is distributed in France.`,
	Example: `  asc encryption create --app com.example.app \
    --description "Uses HTTPS only" --available-on-french-store`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, secApp)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/appEncryptionDeclarations", api.Body{
			Data: api.Resource{
				Type: "appEncryptionDeclarations",
				Attributes: map[string]any{
					"appDescription":                  secDescription,
					"containsProprietaryCryptography": secProprietary,
					"containsThirdPartyCryptography":  secThirdParty,
					"availableOnFrenchStore":          secFrenchStore,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created encryption declaration %s.\n", created.ID)
		return nil
	},
}

var secEncryptionAssignBuildCmd = &cobra.Command{
	Use:   "assign-build",
	Short: "Assign a build to an encryption declaration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		// The relationships/builds linkage endpoint is deprecated; the current way
		// to associate a build is PATCHing the build's appEncryptionDeclaration.
		_, err = c.Patch(cmd.Context(), "/v1/builds/"+secBuildID, api.Body{
			Data: api.Resource{
				Type: "builds",
				ID:   secBuildID,
				Relationships: map[string]json.RawMessage{
					"appEncryptionDeclaration": api.Rel("appEncryptionDeclarations", secID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Assigned build %s to encryption declaration %s.\n", secBuildID, secID)
		return nil
	},
}

var secEncryptionUploadDocumentCmd = &cobra.Command{
	Use:     "upload-document",
	Short:   "Upload a compliance document (e.g. PDF) for an encryption declaration",
	Example: `  asc encryption upload-document --id DECLARATION_ID --file compliance.pdf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		id, err := uploadAsset(cmd.Context(), c, assetSpec{
			reserveType: "appEncryptionDeclarationDocuments",
			relName:     "appEncryptionDeclaration",
			relType:     "appEncryptionDeclarations",
			relID:       secID,
			filePath:    secFile,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded document %s.\n", id)
		return nil
	},
}

// --- accessibility -----------------------------------------------------------

// secA11ySupports maps the supports* attributes to CLI flags; the same set is
// registered on both create and update, and only flags the user passed are sent.
var secA11ySupports = []struct {
	flag string
	attr string
	val  *bool
}{
	{"audio-descriptions", "supportsAudioDescriptions", new(bool)},
	{"captions", "supportsCaptions", new(bool)},
	{"dark-interface", "supportsDarkInterface", new(bool)},
	{"differentiate-without-color", "supportsDifferentiateWithoutColorAlone", new(bool)},
	{"larger-text", "supportsLargerText", new(bool)},
	{"reduced-motion", "supportsReducedMotion", new(bool)},
	{"sufficient-contrast", "supportsSufficientContrast", new(bool)},
	{"voice-control", "supportsVoiceControl", new(bool)},
	{"voiceover", "supportsVoiceover", new(bool)},
}

// secA11yAttrs collects the supports* attributes from flags the user passed.
func secA11yAttrs(cmd *cobra.Command) map[string]any {
	attrs := map[string]any{}
	for _, s := range secA11ySupports {
		if cmd.Flags().Changed(s.flag) {
			attrs[s.attr] = *s.val
		}
	}
	return attrs
}

var secAccessibilityCmd = &cobra.Command{
	Use:   "accessibility",
	Short: "Manage accessibility declarations (Accessibility Nutrition Labels)",
}

var secAccessibilityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessibility declarations of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, secApp)
		if err != nil {
			return err
		}
		decls, err := c.List(ctx, "/v1/apps/"+appID+"/accessibilityDeclarations?limit=200")
		if err != nil {
			return err
		}
		if len(decls) == 0 {
			fmt.Println("No accessibility declarations found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "DEVICE FAMILY\tSTATE\tVOICEOVER\tLARGER TEXT\tID")
		for i := range decls {
			d := &decls[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.Str("deviceFamily"), d.Str("state"), secBool(d, "supportsVoiceover"), secBool(d, "supportsLargerText"), d.ID)
		}
		return w.Flush()
	},
}

var secAccessibilityShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an accessibility declaration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		decl, _, err := c.Get(cmd.Context(), "/v1/accessibilityDeclarations/"+secID)
		if err != nil {
			return err
		}
		return secPrintJSON(decl)
	},
}

var secAccessibilityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an accessibility declaration for an app and device family",
	Long: `Create an accessibility declaration. --device-family is one of IPHONE, IPAD,
APPLE_TV, APPLE_WATCH, MAC, VISION. Pass any of the supports flags (e.g.
--voiceover=true) to declare supported accessibility features; omitted flags
are left unset. Publish it afterwards with "asc accessibility update --publish".`,
	Example: `  asc accessibility create --app com.example.app --device-family IPHONE \
    --voiceover=true --larger-text=true --dark-interface=true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, secApp)
		if err != nil {
			return err
		}
		attrs := secA11yAttrs(cmd)
		attrs["deviceFamily"] = secDeviceFamily
		created, err := c.Post(ctx, "/v1/accessibilityDeclarations", api.Body{
			Data: api.Resource{
				Type:          "accessibilityDeclarations",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created accessibility declaration %s.\n", created.ID)
		return nil
	},
}

var secAccessibilityUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update (and optionally publish) an accessibility declaration",
	Long: `Update the supports flags of a DRAFT accessibility declaration. Only the flags
you pass are changed. Pass --publish to publish the declaration to the App Store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := secA11yAttrs(cmd)
		if cmd.Flags().Changed("publish") {
			attrs["publish"] = secPublish
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass at least one supports flag or --publish")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/accessibilityDeclarations/"+secID, api.Body{
			Data: api.Resource{
				Type:       "accessibilityDeclarations",
				ID:         secID,
				Attributes: attrs,
			},
		})
		if err != nil {
			return err
		}
		fmt.Println("Updated.")
		return nil
	},
}

var secAccessibilityDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an accessibility declaration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/accessibilityDeclarations/"+secID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

func init() {
	secEncryptionListCmd.Flags().StringVar(&secApp, "app", "", "app id or bundle id (required)")
	_ = secEncryptionListCmd.MarkFlagRequired("app")

	secEncryptionShowCmd.Flags().StringVar(&secID, "id", "", "appEncryptionDeclaration id (required)")
	_ = secEncryptionShowCmd.MarkFlagRequired("id")

	secEncryptionCreateCmd.Flags().StringVar(&secApp, "app", "", "app id or bundle id (required)")
	secEncryptionCreateCmd.Flags().StringVar(&secDescription, "description", "", "description of how the app uses encryption (required)")
	secEncryptionCreateCmd.Flags().BoolVar(&secProprietary, "proprietary", false, "contains proprietary cryptography")
	secEncryptionCreateCmd.Flags().BoolVar(&secThirdParty, "third-party", false, "contains third-party cryptography")
	secEncryptionCreateCmd.Flags().BoolVar(&secFrenchStore, "available-on-french-store", false, "the app is available on the French store")
	_ = secEncryptionCreateCmd.MarkFlagRequired("app")
	_ = secEncryptionCreateCmd.MarkFlagRequired("description")

	secEncryptionAssignBuildCmd.Flags().StringVar(&secID, "id", "", "appEncryptionDeclaration id (required)")
	secEncryptionAssignBuildCmd.Flags().StringVar(&secBuildID, "build-id", "", "build id (required)")
	_ = secEncryptionAssignBuildCmd.MarkFlagRequired("id")
	_ = secEncryptionAssignBuildCmd.MarkFlagRequired("build-id")

	secEncryptionUploadDocumentCmd.Flags().StringVar(&secID, "id", "", "appEncryptionDeclaration id (required)")
	secEncryptionUploadDocumentCmd.Flags().StringVar(&secFile, "file", "", "document file, e.g. a PDF (required)")
	_ = secEncryptionUploadDocumentCmd.MarkFlagRequired("id")
	_ = secEncryptionUploadDocumentCmd.MarkFlagRequired("file")

	secEncryptionCmd.AddCommand(secEncryptionListCmd, secEncryptionShowCmd, secEncryptionCreateCmd,
		secEncryptionAssignBuildCmd, secEncryptionUploadDocumentCmd)

	secAccessibilityListCmd.Flags().StringVar(&secApp, "app", "", "app id or bundle id (required)")
	_ = secAccessibilityListCmd.MarkFlagRequired("app")

	secAccessibilityShowCmd.Flags().StringVar(&secID, "id", "", "accessibilityDeclaration id (required)")
	_ = secAccessibilityShowCmd.MarkFlagRequired("id")

	secAccessibilityCreateCmd.Flags().StringVar(&secApp, "app", "", "app id or bundle id (required)")
	secAccessibilityCreateCmd.Flags().StringVar(&secDeviceFamily, "device-family", "", "IPHONE, IPAD, APPLE_TV, APPLE_WATCH, MAC, or VISION (required)")
	_ = secAccessibilityCreateCmd.MarkFlagRequired("app")
	_ = secAccessibilityCreateCmd.MarkFlagRequired("device-family")

	secAccessibilityUpdateCmd.Flags().StringVar(&secID, "id", "", "accessibilityDeclaration id (required)")
	secAccessibilityUpdateCmd.Flags().BoolVar(&secPublish, "publish", false, "publish the declaration to the App Store")
	_ = secAccessibilityUpdateCmd.MarkFlagRequired("id")

	for _, sub := range []*cobra.Command{secAccessibilityCreateCmd, secAccessibilityUpdateCmd} {
		for _, s := range secA11ySupports {
			sub.Flags().BoolVar(s.val, s.flag, false, "declare support: "+s.attr)
		}
	}

	secAccessibilityDeleteCmd.Flags().StringVar(&secID, "id", "", "accessibilityDeclaration id (required)")
	_ = secAccessibilityDeleteCmd.MarkFlagRequired("id")

	secAccessibilityCmd.AddCommand(secAccessibilityListCmd, secAccessibilityShowCmd, secAccessibilityCreateCmd,
		secAccessibilityUpdateCmd, secAccessibilityDeleteCmd)

	rootCmd.AddCommand(secEncryptionCmd, secAccessibilityCmd)
}
