package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	ipxApp         string
	ipxProduct     string
	ipxProductID   string
	ipxName        string
	ipxType        string
	ipxFamilyShare bool
	ipxReviewNote  string
	ipxTerritories string
	ipxAvailNew    bool
	ipxFile        string
	ipxImageID     string
	ipxOfferName   string
	ipxEligibility []string
	ipxOfferTerr   string
	ipxOfferPrice  string
	ipxOfferID     string
	ipxNumCodes    int
	ipxExpiration  string
	ipxOutput      string
)

// ipxFindIAPPricePoint resolves a customer price in a territory to an IAP
// price point id, mirroring the resolution in "asc iap price".
func ipxFindIAPPricePoint(ctx context.Context, c *api.Client, iapID, territory, price string) (string, error) {
	points, err := c.List(ctx, "/v2/inAppPurchases/"+iapID+"/pricePoints?filter[territory]="+territory+"&limit=200")
	if err != nil {
		return "", err
	}
	want, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return "", fmt.Errorf("invalid --price %q: %w", price, err)
	}
	var near []string
	for _, p := range points {
		cp := p.Str("customerPrice")
		if f, err := strconv.ParseFloat(cp, 64); err == nil {
			if f == want {
				return p.ID, nil
			}
			near = append(near, cp)
		}
	}
	hint := ""
	if len(near) > 0 {
		if len(near) > 12 {
			near = near[:12]
		}
		hint = " Available nearby: " + strings.Join(near, ", ")
	}
	return "", fmt.Errorf("no price point with customerPrice=%s in %s.%s", price, territory, hint)
}

// ipxStrList formats a string-array attribute as a comma-joined string.
func ipxStrList(r *api.Resource, key string) string {
	raw, ok := r.Attributes[key].([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ",")
}

var ipxCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an in-app purchase",
	Example: `  asc iap create --app 6790641087 --product-id com.example.app.credits100 \
    --name "100 Credits" --type CONSUMABLE --review-note "Unlocks 100 analysis credits."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		switch ipxType {
		case "CONSUMABLE", "NON_CONSUMABLE", "NON_RENEWING_SUBSCRIPTION":
		default:
			return fmt.Errorf("--type must be CONSUMABLE, NON_CONSUMABLE or NON_RENEWING_SUBSCRIPTION")
		}
		attrs := map[string]any{
			"productId":         ipxProductID,
			"name":              ipxName,
			"inAppPurchaseType": ipxType,
		}
		if cmd.Flags().Changed("family-sharable") {
			attrs["familySharable"] = ipxFamilyShare
		}
		if ipxReviewNote != "" {
			v, err := valueOrFile(ipxReviewNote)
			if err != nil {
				return err
			}
			attrs["reviewNote"] = v
		}
		created, err := c.Post(ctx, "/v2/inAppPurchases", api.Body{
			Data: api.Resource{
				Type:          "inAppPurchases",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created IAP %s (%s).\n", ipxProductID, created.ID)
		return nil
	},
}

var ipxShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show an in-app purchase's attributes and state",
	Example: `  asc iap show --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		out := map[string]any{"id": iap.ID}
		for k, v := range iap.Attributes {
			out[k] = v
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var ipxDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an in-app purchase",
	Example: `  asc iap delete --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		if err := c.Delete(ctx, "/v2/inAppPurchases/"+iap.ID); err != nil {
			return err
		}
		fmt.Printf("Deleted IAP %s (%s).\n", iap.Str("productId"), iap.ID)
		return nil
	},
}

var ipxAvailCmd = &cobra.Command{
	Use:   "availability",
	Short: "Show or set the territories where an in-app purchase is available",
}

var ipxAvailShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show an in-app purchase's territory availability",
	Example: `  asc iap availability show --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		avail, err := c.GetOptional(ctx, "/v2/inAppPurchases/"+iap.ID+"/inAppPurchaseAvailability")
		if err != nil {
			return err
		}
		if avail == nil || avail.ID == "" {
			fmt.Printf("IAP %s has no availability configuration (available everywhere by default).\n", iap.Str("productId"))
			return nil
		}
		fmt.Printf("Available in new territories: %s\n", prcDash(prcBool(avail, "availableInNewTerritories")))
		terrs, err := c.List(ctx, "/v1/inAppPurchaseAvailabilities/"+avail.ID+"/availableTerritories?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TERRITORY\tCURRENCY")
		for _, t := range terrs {
			fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Str("currency"))
		}
		return w.Flush()
	},
}

var ipxAvailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the territories where an in-app purchase is available",
	Example: `  asc iap availability set --app 6790641087 --product com.example.app.credits100 \
    --territories JPN,USA --available-in-new-territories`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		var refs []map[string]string
		for _, t := range strings.Split(ipxTerritories, ",") {
			if t = strings.ToUpper(strings.TrimSpace(t)); t != "" {
				refs = append(refs, map[string]string{"type": "territories", "id": t})
			}
		}
		if len(refs) == 0 {
			return fmt.Errorf("--territories must list at least one territory code, e.g. JPN,USA")
		}
		terrRel, _ := json.Marshal(map[string]any{"data": refs})
		_, err = c.Post(ctx, "/v1/inAppPurchaseAvailabilities", api.Body{
			Data: api.Resource{
				Type:       "inAppPurchaseAvailabilities",
				Attributes: map[string]any{"availableInNewTerritories": ipxAvailNew},
				Relationships: map[string]json.RawMessage{
					"inAppPurchase":        api.Rel("inAppPurchases", iap.ID),
					"availableTerritories": terrRel,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("IAP %s availability set: %d territories (availableInNewTerritories=%v).\n", iap.Str("productId"), len(refs), ipxAvailNew)
		return nil
	},
}

var ipxScheduleCmd = &cobra.Command{
	Use:     "schedule",
	Short:   "Show an in-app purchase's price schedule (base territory and manual prices)",
	Example: `  asc iap schedule --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		sched, err := c.GetOptional(ctx, "/v2/inAppPurchases/"+iap.ID+"/iapPriceSchedule")
		if err != nil {
			return err
		}
		if sched == nil || sched.ID == "" {
			fmt.Printf("IAP %s has no price schedule. Set one with \"asc iap price\".\n", iap.Str("productId"))
			return nil
		}
		base, err := c.GetOptional(ctx, "/v1/inAppPurchasePriceSchedules/"+sched.ID+"/baseTerritory")
		if err != nil {
			return err
		}
		if base != nil && base.ID != "" {
			fmt.Printf("Base territory: %s (%s)\n", base.ID, base.Str("currency"))
		}
		manual, inc, err := prcListInc(ctx, c, "/v1/inAppPurchasePriceSchedules/"+sched.ID+"/manualPrices?include=inAppPurchasePricePoint,territory&limit=200")
		if err != nil {
			// The live API 404s on the prices sub-paths when the schedule was
			// never configured through this flow.
			if api.IsNotFound(err) {
				fmt.Printf("No prices configured yet. Set one with \"asc iap price\".\n")
				return nil
			}
			return err
		}
		fmt.Printf("Manual prices (%d):\n", len(manual))
		return prcPrintPrices(manual, inc, "inAppPurchasePricePoint", "inAppPurchasePricePoints")
	},
}

var ipxUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "Upload a promotional image for an in-app purchase",
	Example: `  asc iap upload-image --app 6790641087 --product com.example.app.credits100 \
    --file app-store/iap-promo.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		id, err := uploadAsset(ctx, c, assetSpec{
			reserveType: "inAppPurchaseImages",
			relName:     "inAppPurchase",
			relType:     "inAppPurchases",
			relID:       iap.ID,
			filePath:    ipxFile,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded IAP image -> %s\n", id)
		return nil
	},
}

var ipxListImagesCmd = &cobra.Command{
	Use:     "list-images",
	Short:   "List an in-app purchase's promotional images",
	Example: `  asc iap list-images --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		images, err := c.List(ctx, "/v2/inAppPurchases/"+iap.ID+"/images?limit=50")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tFILE NAME\tSTATE")
		for _, img := range images {
			fmt.Fprintf(w, "%s\t%s\t%s\n", img.ID, img.Str("fileName"), img.Str("state"))
		}
		return w.Flush()
	},
}

var ipxDeleteImageCmd = &cobra.Command{
	Use:     "delete-image",
	Short:   "Delete an in-app purchase promotional image by id",
	Example: `  asc iap delete-image --id 12345678-1234-1234-1234-123456789012`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/inAppPurchaseImages/"+ipxImageID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var ipxOfferCmd = &cobra.Command{
	Use:   "offer-codes",
	Short: "Manage in-app purchase offer codes",
}

var ipxOfferListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List an in-app purchase's offer codes",
	Example: `  asc iap offer-codes list --app 6790641087 --product com.example.app.credits100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		codes, err := c.List(ctx, "/v2/inAppPurchases/"+iap.ID+"/offerCodes?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tACTIVE\tELIGIBILITIES\tID")
		for i := range codes {
			oc := &codes[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", oc.Str("name"), prcDash(prcBool(oc, "active")), ipxStrList(oc, "customerEligibilities"), oc.ID)
		}
		return w.Flush()
	},
}

var ipxOfferCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an offer code for an in-app purchase",
	Long: `Create an offer code. The offer price is resolved from --price (customer
price) and --territory to an in-app purchase price point; Apple derives the
other territories' offer prices from it. After creating, generate redeemable
codes with "asc iap offer-codes create-codes".`,
	Example: `  asc iap offer-codes create --app 6790641087 --product com.example.app.credits100 \
    --name "Summer Promo" --territory JPN --price 100 --eligibility NON_SPENDER,CHURNED_SPENDER`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, ipxApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, ipxProduct)
		if err != nil {
			return err
		}
		for _, e := range ipxEligibility {
			switch e {
			case "NON_SPENDER", "ACTIVE_SPENDER", "CHURNED_SPENDER":
			default:
				return fmt.Errorf("--eligibility values must be NON_SPENDER, ACTIVE_SPENDER or CHURNED_SPENDER (got %q)", e)
			}
		}
		pointID, err := ipxFindIAPPricePoint(ctx, c, iap.ID, ipxOfferTerr, ipxOfferPrice)
		if err != nil {
			return err
		}
		const lid = "offer-price-1"
		pricesRel, _ := json.Marshal(map[string]any{"data": []map[string]string{{"type": "inAppPurchaseOfferPrices", "id": lid}}})
		created, err := c.Post(ctx, "/v1/inAppPurchaseOfferCodes", api.Body{
			Data: api.Resource{
				Type: "inAppPurchaseOfferCodes",
				Attributes: map[string]any{
					"name":                  ipxOfferName,
					"customerEligibilities": ipxEligibility,
				},
				Relationships: map[string]json.RawMessage{
					"inAppPurchase": api.Rel("inAppPurchases", iap.ID),
					"prices":        pricesRel,
				},
			},
			Included: []api.Resource{{
				Type: "inAppPurchaseOfferPrices",
				ID:   lid,
				Relationships: map[string]json.RawMessage{
					"territory":  api.Rel("territories", ipxOfferTerr),
					"pricePoint": api.Rel("inAppPurchasePricePoints", pointID),
				},
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created offer code %q (%s).\n", ipxOfferName, created.ID)
		return nil
	},
}

var ipxOfferDeactivateCmd = &cobra.Command{
	Use:     "deactivate",
	Short:   "Deactivate an offer code",
	Example: `  asc iap offer-codes deactivate --id 12345678-1234-1234-1234-123456789012`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/inAppPurchaseOfferCodes/"+ipxOfferID, api.Body{
			Data: api.Resource{
				Type:       "inAppPurchaseOfferCodes",
				ID:         ipxOfferID,
				Attributes: map[string]any{"active": false},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Offer code %s deactivated.\n", ipxOfferID)
		return nil
	},
}

var ipxOfferCreateCodesCmd = &cobra.Command{
	Use:   "create-codes",
	Short: "Generate one-time-use codes for an offer code",
	Example: `  asc iap offer-codes create-codes --id 12345678-1234-1234-1234-123456789012 \
    --number 100 --expiration-date 2026-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/inAppPurchaseOfferCodeOneTimeUseCodes", api.Body{
			Data: api.Resource{
				Type: "inAppPurchaseOfferCodeOneTimeUseCodes",
				Attributes: map[string]any{
					"numberOfCodes":  ipxNumCodes,
					"expirationDate": ipxExpiration,
				},
				Relationships: map[string]json.RawMessage{"offerCode": api.Rel("inAppPurchaseOfferCodes", ipxOfferID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created %d one-time-use codes (batch %s). Download with \"asc iap offer-codes download-codes --id %s\".\n",
			ipxNumCodes, created.ID, created.ID)
		return nil
	},
}

var ipxOfferDownloadCodesCmd = &cobra.Command{
	Use:   "download-codes",
	Short: "Download a one-time-use code batch as CSV",
	Long: `Download the code values of a one-time-use code batch (created with
"create-codes") as CSV. Code generation is asynchronous; retry shortly after
creation if the values are not ready yet.`,
	Example: `  asc iap offer-codes download-codes --id 12345678-1234-1234-1234-123456789012 --output codes.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		data, err := c.Download(cmd.Context(), "/v1/inAppPurchaseOfferCodeOneTimeUseCodes/"+ipxOfferID+"/values", "text/csv")
		if err != nil {
			return err
		}
		if ipxOutput != "" {
			if err := os.WriteFile(ipxOutput, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("Wrote %d bytes to %s.\n", len(data), ipxOutput)
			return nil
		}
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	appCmds := []*cobra.Command{
		ipxCreateCmd, ipxShowCmd, ipxDeleteCmd, ipxAvailShowCmd, ipxAvailSetCmd,
		ipxScheduleCmd, ipxUploadImageCmd, ipxListImagesCmd, ipxOfferListCmd, ipxOfferCreateCmd,
	}
	for _, sub := range appCmds {
		sub.Flags().StringVar(&ipxApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	prodCmds := []*cobra.Command{
		ipxShowCmd, ipxDeleteCmd, ipxAvailShowCmd, ipxAvailSetCmd,
		ipxScheduleCmd, ipxUploadImageCmd, ipxListImagesCmd, ipxOfferListCmd, ipxOfferCreateCmd,
	}
	for _, sub := range prodCmds {
		sub.Flags().StringVar(&ipxProduct, "product", "", "product id or ASC in-app purchase id (required)")
		_ = sub.MarkFlagRequired("product")
	}

	ipxCreateCmd.Flags().StringVar(&ipxProductID, "product-id", "", "new product id, e.g. com.example.app.credits100 (required)")
	ipxCreateCmd.Flags().StringVar(&ipxName, "name", "", "reference name shown in App Store Connect (required)")
	ipxCreateCmd.Flags().StringVar(&ipxType, "type", "", "CONSUMABLE, NON_CONSUMABLE or NON_RENEWING_SUBSCRIPTION (required)")
	ipxCreateCmd.Flags().BoolVar(&ipxFamilyShare, "family-sharable", false, "allow Family Sharing")
	ipxCreateCmd.Flags().StringVar(&ipxReviewNote, "review-note", "", "note for App Review (@file allowed)")
	_ = ipxCreateCmd.MarkFlagRequired("product-id")
	_ = ipxCreateCmd.MarkFlagRequired("name")
	_ = ipxCreateCmd.MarkFlagRequired("type")

	ipxAvailSetCmd.Flags().StringVar(&ipxTerritories, "territories", "", "comma-separated territory codes, e.g. JPN,USA (required)")
	ipxAvailSetCmd.Flags().BoolVar(&ipxAvailNew, "available-in-new-territories", false, "automatically become available in future new territories")
	_ = ipxAvailSetCmd.MarkFlagRequired("territories")

	ipxUploadImageCmd.Flags().StringVar(&ipxFile, "file", "", "image file (required)")
	_ = ipxUploadImageCmd.MarkFlagRequired("file")

	ipxDeleteImageCmd.Flags().StringVar(&ipxImageID, "id", "", "inAppPurchaseImage id (required)")
	_ = ipxDeleteImageCmd.MarkFlagRequired("id")

	ipxOfferCreateCmd.Flags().StringVar(&ipxOfferName, "name", "", "offer code reference name (required)")
	ipxOfferCreateCmd.Flags().StringSliceVar(&ipxEligibility, "eligibility", []string{"NON_SPENDER", "ACTIVE_SPENDER", "CHURNED_SPENDER"},
		"customer eligibilities: NON_SPENDER, ACTIVE_SPENDER, CHURNED_SPENDER")
	ipxOfferCreateCmd.Flags().StringVar(&ipxOfferTerr, "territory", "JPN", "territory code the offer price is defined in")
	ipxOfferCreateCmd.Flags().StringVar(&ipxOfferPrice, "price", "", "offer customer price in the territory currency (required)")
	_ = ipxOfferCreateCmd.MarkFlagRequired("name")
	_ = ipxOfferCreateCmd.MarkFlagRequired("price")

	for _, sub := range []*cobra.Command{ipxOfferDeactivateCmd, ipxOfferCreateCodesCmd, ipxOfferDownloadCodesCmd} {
		sub.Flags().StringVar(&ipxOfferID, "id", "", "resource id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	ipxOfferDeactivateCmd.Flags().Lookup("id").Usage = "inAppPurchaseOfferCode id (required)"
	ipxOfferCreateCodesCmd.Flags().Lookup("id").Usage = "inAppPurchaseOfferCode id (required)"
	ipxOfferDownloadCodesCmd.Flags().Lookup("id").Usage = "one-time-use code batch id from create-codes (required)"

	ipxOfferCreateCodesCmd.Flags().IntVar(&ipxNumCodes, "number", 0, "number of codes to generate (required)")
	ipxOfferCreateCodesCmd.Flags().StringVar(&ipxExpiration, "expiration-date", "", "expiration date YYYY-MM-DD (required)")
	_ = ipxOfferCreateCodesCmd.MarkFlagRequired("number")
	_ = ipxOfferCreateCodesCmd.MarkFlagRequired("expiration-date")

	ipxOfferDownloadCodesCmd.Flags().StringVar(&ipxOutput, "output", "", "write CSV to this file (default: stdout)")

	ipxAvailCmd.AddCommand(ipxAvailShowCmd, ipxAvailSetCmd)
	ipxOfferCmd.AddCommand(ipxOfferListCmd, ipxOfferCreateCmd, ipxOfferDeactivateCmd, ipxOfferCreateCodesCmd, ipxOfferDownloadCodesCmd)
	iapCmd.AddCommand(ipxCreateCmd, ipxShowCmd, ipxDeleteCmd, ipxAvailCmd, ipxScheduleCmd,
		ipxUploadImageCmd, ipxListImagesCmd, ipxDeleteImageCmd, ipxOfferCmd)
}
