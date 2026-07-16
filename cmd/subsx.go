package cmd

// subsx.go fills IAP/subscription API gaps: subscription images, subscription
// group submissions, draft versions (subscription / subscription group / IAP),
// offer code custom codes, and hosted IAP content info. All request shapes are
// verified against the App Store Connect OpenAPI spec 4.4.1.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	sbxApp        string
	sbxSub        string
	sbxProduct    string
	sbxFile       string
	sbxID         string
	sbxGroupID    string
	sbxVersionID  string
	sbxCode       string
	sbxCount      int
	sbxExpiration string
)

// --- shared output helpers ---------------------------------------------------

// sbxPrintJSON prints a resource as indented JSON.
func sbxPrintJSON(r *api.Resource) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// sbxPrintLocalizations lists a version's localizations as a table. All three
// localization resources (subscription, subscription group, IAP) share the
// locale/name/state attributes.
func sbxPrintLocalizations(ctx context.Context, c *api.Client, path string) error {
	locs, err := c.List(ctx, path)
	if err != nil {
		return err
	}
	fmt.Printf("Localizations (%d):\n", len(locs))
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "LOCALE\tNAME\tSTATE\tID")
	for i := range locs {
		l := &locs[i]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.Str("locale"), l.Str("name"), subsAttr(l, "state"), l.ID)
	}
	return w.Flush()
}

// sbxPrintImages lists a version's (or subscription's) images as a table.
func sbxPrintImages(ctx context.Context, c *api.Client, path string) error {
	imgs, err := c.List(ctx, path)
	if err != nil {
		return err
	}
	fmt.Printf("Images (%d):\n", len(imgs))
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "FILE_NAME\tSTATE\tID")
	for i := range imgs {
		img := &imgs[i]
		fmt.Fprintf(w, "%s\t%s\t%s\n", img.Str("fileName"), img.Str("state"), img.ID)
	}
	return w.Flush()
}

// --- subscriptions image -------------------------------------------------------

var sbxImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage a subscription's promotional image",
}

var sbxImageUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a promotional image for a subscription",
	Long: `Upload a promotional image via the reserve/upload/commit flow
(POST /v1/subscriptionImages relates directly to the subscription; the commit
sets uploaded plus sourceFileChecksum).`,
	Example: `  asc subscriptions image upload --app com.example.app --sub com.example.app.pro.monthly \
    --file app-store/sub-promo.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sbxApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, sbxSub)
		if err != nil {
			return err
		}
		id, err := uploadAsset(ctx, c, assetSpec{
			reserveType: "subscriptionImages",
			relName:     "subscription",
			relType:     "subscriptions",
			relID:       sub.ID,
			filePath:    sbxFile,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded subscription image for %s -> %s\n", sub.Str("productId"), id)
		return nil
	},
}

var sbxImageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's images",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sbxApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, sbxSub)
		if err != nil {
			return err
		}
		return sbxPrintImages(ctx, c, "/v1/subscriptions/"+sub.ID+"/images?limit=50")
	},
}

var sbxImageDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subscription image by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/subscriptionImages/"+sbxID); err != nil {
			return err
		}
		fmt.Println("Deleted subscription image.")
		return nil
	},
}

// --- subscriptions groups submit -------------------------------------------------

var sbxGroupSubmitCmd = &cobra.Command{
	Use:     "submit",
	Short:   "Submit a subscription group (its display names) for review",
	Example: `  asc subscriptions groups submit --group-id 21489000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Post(cmd.Context(), "/v1/subscriptionGroupSubmissions", api.Body{
			Data: api.Resource{
				Type:          "subscriptionGroupSubmissions",
				Relationships: map[string]json.RawMessage{"subscriptionGroup": api.Rel("subscriptionGroups", sbxGroupID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted subscription group %s for review.\n", sbxGroupID)
		return nil
	},
}

// --- subscriptions version -----------------------------------------------------

var sbxVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage subscription draft versions (to edit a live subscription)",
}

var sbxVersionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft version of a subscription",
	Long: `Create a subscriptionVersion, a draft for editing a subscription that is
already live. The create request carries no attributes, only the subscription
relationship.`,
	Example: `  asc subscriptions version create --app com.example.app --sub com.example.app.pro.monthly`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sbxApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, sbxSub)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/subscriptionVersions", api.Body{
			Data: api.Resource{
				Type:          "subscriptionVersions",
				Relationships: map[string]json.RawMessage{"subscription": api.Rel("subscriptions", sub.ID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created draft version for %s -> %s (state %s)\n",
			sub.Str("productId"), created.ID, subsAttr(created, "state"))
		return nil
	},
}

var sbxVersionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a subscription version with its localizations and images",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, _, err := c.Get(ctx, "/v1/subscriptionVersions/"+sbxVersionID)
		if err != nil {
			return err
		}
		if err := sbxPrintJSON(ver); err != nil {
			return err
		}
		if err := sbxPrintLocalizations(ctx, c, "/v1/subscriptionVersions/"+sbxVersionID+"/localizations?limit=50"); err != nil {
			return err
		}
		return sbxPrintImages(ctx, c, "/v1/subscriptionVersions/"+sbxVersionID+"/images?limit=50")
	},
}

// --- subscriptions groups version -------------------------------------------------

var sbxGroupVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage subscription group draft versions",
}

var sbxGroupVersionCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a draft version of a subscription group",
	Example: `  asc subscriptions groups version create --group-id 21489000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/subscriptionGroupVersions", api.Body{
			Data: api.Resource{
				Type:          "subscriptionGroupVersions",
				Relationships: map[string]json.RawMessage{"subscriptionGroup": api.Rel("subscriptionGroups", sbxGroupID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created draft version for group %s -> %s (state %s)\n",
			sbxGroupID, created.ID, subsAttr(created, "state"))
		return nil
	},
}

var sbxGroupVersionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a subscription group version with its localizations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, _, err := c.Get(ctx, "/v1/subscriptionGroupVersions/"+sbxVersionID)
		if err != nil {
			return err
		}
		if err := sbxPrintJSON(ver); err != nil {
			return err
		}
		return sbxPrintLocalizations(ctx, c, "/v1/subscriptionGroupVersions/"+sbxVersionID+"/localizations?limit=50")
	},
}

// --- subscriptions offer-codes custom-codes ------------------------------------------

var sbxCustomCodesCmd = &cobra.Command{
	Use:   "custom-codes",
	Short: "Manage custom (multi-redemption) codes for a subscription offer code campaign",
}

var sbxCustomCodesCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a custom code for an offer code campaign",
	Example: `  asc subscriptions offer-codes custom-codes create --id 21500000 --code SPRING24 --count 500 --expiration 2026-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"customCode":    sbxCode,
			"numberOfCodes": sbxCount,
		}
		if cmd.Flags().Changed("expiration") {
			attrs["expirationDate"] = sbxExpiration
		}
		created, err := c.Post(cmd.Context(), "/v1/subscriptionOfferCodeCustomCodes", api.Body{
			Data: api.Resource{
				Type:          "subscriptionOfferCodeCustomCodes",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"offerCode": api.Rel("subscriptionOfferCodes", sbxID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created custom code %s (%d redemptions) -> %s\n", sbxCode, sbxCount, created.ID)
		return nil
	},
}

var sbxCustomCodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List an offer code campaign's custom codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		codes, err := c.List(cmd.Context(), "/v1/subscriptionOfferCodes/"+sbxID+"/customCodes?limit=200")
		if err != nil {
			return err
		}
		return sbxPrintCustomCodes(codes)
	},
}

// sbxPrintCustomCodes prints subscription/IAP offer code custom codes (the two
// resources share the same attributes).
func sbxPrintCustomCodes(codes []api.Resource) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "CUSTOM_CODE\tACTIVE\tCODES\tCREATED\tEXPIRES\tID")
	for i := range codes {
		cc := &codes[i]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			cc.Str("customCode"), subsAttr(cc, "active"), subsAttr(cc, "numberOfCodes"),
			subsAttr(cc, "createdDate"), subsAttr(cc, "expirationDate"), cc.ID)
	}
	return w.Flush()
}

// --- iap version ------------------------------------------------------------------

var sbxIapVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage in-app purchase draft versions (to edit a live in-app purchase)",
}

var sbxIapVersionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft version of an in-app purchase",
	Long: `Create an inAppPurchaseVersion, a draft for editing an in-app purchase that
is already live. The create request carries no attributes, only the
inAppPurchase relationship.`,
	Example: `  asc iap version create --app com.example.app --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sbxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, sbxProduct)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/inAppPurchaseVersions", api.Body{
			Data: api.Resource{
				Type:          "inAppPurchaseVersions",
				Relationships: map[string]json.RawMessage{"inAppPurchase": api.Rel("inAppPurchases", iap.ID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created draft version for %s -> %s (state %s)\n",
			iap.Str("productId"), created.ID, subsAttr(created, "state"))
		return nil
	},
}

var sbxIapVersionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an in-app purchase version with its localizations and images",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		ver, _, err := c.Get(ctx, "/v1/inAppPurchaseVersions/"+sbxVersionID)
		if err != nil {
			return err
		}
		if err := sbxPrintJSON(ver); err != nil {
			return err
		}
		if err := sbxPrintLocalizations(ctx, c, "/v1/inAppPurchaseVersions/"+sbxVersionID+"/localizations?limit=50"); err != nil {
			return err
		}
		return sbxPrintImages(ctx, c, "/v1/inAppPurchaseVersions/"+sbxVersionID+"/images?limit=50")
	},
}

// --- iap offer-codes custom-codes ---------------------------------------------------

var sbxIapCustomCodesCmd = &cobra.Command{
	Use:   "custom-codes",
	Short: "Manage custom (multi-redemption) codes for an in-app purchase offer code",
}

var sbxIapCustomCodesCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a custom code for an in-app purchase offer code",
	Example: `  asc iap offer-codes custom-codes create --id 12345678-1234-1234-1234-123456789012 --code SPRING24 --count 500`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"customCode":    sbxCode,
			"numberOfCodes": sbxCount,
		}
		if cmd.Flags().Changed("expiration") {
			attrs["expirationDate"] = sbxExpiration
		}
		created, err := c.Post(cmd.Context(), "/v1/inAppPurchaseOfferCodeCustomCodes", api.Body{
			Data: api.Resource{
				Type:          "inAppPurchaseOfferCodeCustomCodes",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"offerCode": api.Rel("inAppPurchaseOfferCodes", sbxID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created custom code %s (%d redemptions) -> %s\n", sbxCode, sbxCount, created.ID)
		return nil
	},
}

var sbxIapCustomCodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List an in-app purchase offer code's custom codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		codes, err := c.List(cmd.Context(), "/v1/inAppPurchaseOfferCodes/"+sbxID+"/customCodes?limit=200")
		if err != nil {
			return err
		}
		return sbxPrintCustomCodes(codes)
	},
}

// --- iap content -----------------------------------------------------------------

var sbxIapContentCmd = &cobra.Command{
	Use:   "content",
	Short: "Show an in-app purchase's hosted content info (file name, size, download URL)",
	Long: `Show the Apple-hosted content of an in-app purchase
(GET /v2/inAppPurchases/{id}/content). Read-only: the spec exposes no create or
update operations for inAppPurchaseContents; hosted content is uploaded with
Xcode or Transporter.`,
	Example: `  asc iap content --app com.example.app --product com.example.app.premiumpack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, sbxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, sbxProduct)
		if err != nil {
			return err
		}
		content, err := c.GetOptional(ctx, "/v2/inAppPurchases/"+iap.ID+"/content")
		if err != nil {
			return err
		}
		if content == nil || content.ID == "" {
			fmt.Printf("In-app purchase %s has no hosted content.\n", iap.Str("productId"))
			return nil
		}
		fmt.Printf("id:               %s\n", content.ID)
		fmt.Printf("fileName:         %s\n", subsAttr(content, "fileName"))
		fmt.Printf("fileSize:         %s\n", subsAttr(content, "fileSize"))
		fmt.Printf("lastModifiedDate: %s\n", subsAttr(content, "lastModifiedDate"))
		fmt.Printf("url:              %s\n", subsAttr(content, "url"))
		return nil
	},
}

// --- wiring ------------------------------------------------------------------------

func init() {
	// --app
	for _, sub := range []*cobra.Command{
		sbxImageUploadCmd, sbxImageListCmd, sbxVersionCreateCmd,
		sbxIapVersionCreateCmd, sbxIapContentCmd,
	} {
		sub.Flags().StringVar(&sbxApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	// --sub
	for _, sub := range []*cobra.Command{sbxImageUploadCmd, sbxImageListCmd, sbxVersionCreateCmd} {
		sub.Flags().StringVar(&sbxSub, "sub", "", "subscription id or productId (required)")
		_ = sub.MarkFlagRequired("sub")
	}
	// --product
	for _, sub := range []*cobra.Command{sbxIapVersionCreateCmd, sbxIapContentCmd} {
		sub.Flags().StringVar(&sbxProduct, "product", "", "product id or ASC in-app purchase id (required)")
		_ = sub.MarkFlagRequired("product")
	}
	// --group-id
	for _, sub := range []*cobra.Command{sbxGroupSubmitCmd, sbxGroupVersionCreateCmd} {
		sub.Flags().StringVar(&sbxGroupID, "group-id", "", "subscription group id (required)")
		_ = sub.MarkFlagRequired("group-id")
	}
	// --version-id
	sbxVersionShowCmd.Flags().StringVar(&sbxVersionID, "version-id", "", "subscriptionVersion id (required)")
	_ = sbxVersionShowCmd.MarkFlagRequired("version-id")
	sbxGroupVersionShowCmd.Flags().StringVar(&sbxVersionID, "version-id", "", "subscriptionGroupVersion id (required)")
	_ = sbxGroupVersionShowCmd.MarkFlagRequired("version-id")
	sbxIapVersionShowCmd.Flags().StringVar(&sbxVersionID, "version-id", "", "inAppPurchaseVersion id (required)")
	_ = sbxIapVersionShowCmd.MarkFlagRequired("version-id")

	// image
	sbxImageUploadCmd.Flags().StringVar(&sbxFile, "file", "", "image file (required)")
	_ = sbxImageUploadCmd.MarkFlagRequired("file")
	sbxImageDeleteCmd.Flags().StringVar(&sbxID, "id", "", "subscriptionImage id (required)")
	_ = sbxImageDeleteCmd.MarkFlagRequired("id")

	// custom-codes (subscription)
	sbxCustomCodesCreateCmd.Flags().StringVar(&sbxID, "id", "", "subscriptionOfferCode id (required)")
	sbxCustomCodesListCmd.Flags().StringVar(&sbxID, "id", "", "subscriptionOfferCode id (required)")
	// custom-codes (IAP)
	sbxIapCustomCodesCreateCmd.Flags().StringVar(&sbxID, "id", "", "inAppPurchaseOfferCode id (required)")
	sbxIapCustomCodesListCmd.Flags().StringVar(&sbxID, "id", "", "inAppPurchaseOfferCode id (required)")
	for _, sub := range []*cobra.Command{
		sbxCustomCodesCreateCmd, sbxCustomCodesListCmd,
		sbxIapCustomCodesCreateCmd, sbxIapCustomCodesListCmd,
	} {
		_ = sub.MarkFlagRequired("id")
	}
	for _, sub := range []*cobra.Command{sbxCustomCodesCreateCmd, sbxIapCustomCodesCreateCmd} {
		sub.Flags().StringVar(&sbxCode, "code", "", "custom code value, e.g. SPRING24 (required)")
		sub.Flags().IntVar(&sbxCount, "count", 0, "number of redemptions the code allows (required)")
		sub.Flags().StringVar(&sbxExpiration, "expiration", "", "expiration date YYYY-MM-DD (optional)")
		_ = sub.MarkFlagRequired("code")
		_ = sub.MarkFlagRequired("count")
	}

	sbxImageCmd.AddCommand(sbxImageUploadCmd, sbxImageListCmd, sbxImageDeleteCmd)
	sbxVersionCmd.AddCommand(sbxVersionCreateCmd, sbxVersionShowCmd)
	sbxGroupVersionCmd.AddCommand(sbxGroupVersionCreateCmd, sbxGroupVersionShowCmd)
	sbxCustomCodesCmd.AddCommand(sbxCustomCodesCreateCmd, sbxCustomCodesListCmd)
	sbxIapVersionCmd.AddCommand(sbxIapVersionCreateCmd, sbxIapVersionShowCmd)
	sbxIapCustomCodesCmd.AddCommand(sbxIapCustomCodesCreateCmd, sbxIapCustomCodesListCmd)

	subsCmd.AddCommand(sbxImageCmd, sbxVersionCmd)
	subsGroupsCmd.AddCommand(sbxGroupSubmitCmd, sbxGroupVersionCmd)
	subsCodesCmd.AddCommand(sbxCustomCodesCmd)
	iapCmd.AddCommand(sbxIapVersionCmd, sbxIapContentCmd)
	ipxOfferCmd.AddCommand(sbxIapCustomCodesCmd)
}
