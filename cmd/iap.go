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

var iapCmd = &cobra.Command{
	Use:   "iap",
	Short: "Manage in-app purchases (localizations, price, review screenshot, submit)",
}

var (
	iapApp     string
	iapProduct string
	iapLocale  string
	iapName    string
	iapDesc    string
	iapTerr    string
	iapPrice   string
	iapFile    string
	iapAutoFit bool
)

// resolveIAP resolves a product id (numeric ASC id) or a productId string to the IAP resource.
func resolveIAP(ctx context.Context, c *api.Client, appID, ref string) (*api.Resource, error) {
	iaps, err := c.List(ctx, "/v1/apps/"+appID+"/inAppPurchasesV2?limit=200")
	if err != nil {
		return nil, err
	}
	for i := range iaps {
		if iaps[i].ID == ref || iaps[i].Str("productId") == ref {
			return &iaps[i], nil
		}
	}
	return nil, fmt.Errorf("no in-app purchase %q on app %s", ref, appID)
}

var iapListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's in-app purchases",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, iapApp)
		if err != nil {
			return err
		}
		iaps, err := c.List(ctx, "/v1/apps/"+appID+"/inAppPurchasesV2?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PRODUCT ID\tTYPE\tSTATE\tID")
		for _, p := range iaps {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Str("productId"), p.Str("inAppPurchaseType"), p.Str("state"), p.ID)
		}
		return w.Flush()
	},
}

var iapLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set an in-app purchase's localized display name and description",
	Example: `  asc iap localize --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --locale ja --name "AI解析チケット（100枚）" \
    --description "領収書のAI解析に使えるチケット。1枚の解析につき1枚消費します。"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, iapApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, iapProduct)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = iapName
		}
		if cmd.Flags().Changed("description") {
			v, err := valueOrFile(iapDesc)
			if err != nil {
				return err
			}
			attrs["description"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name and/or --description")
		}
		locs, err := c.List(ctx, "/v2/inAppPurchases/"+iap.ID+"/inAppPurchaseLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", iapLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/inAppPurchaseLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "inAppPurchaseLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			attrs["locale"] = iapLocale
			_, err = c.Post(ctx, "/v1/inAppPurchaseLocalizations", api.Body{
				Data: api.Resource{
					Type:          "inAppPurchaseLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"inAppPurchaseV2": api.Rel("inAppPurchases", iap.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("IAP %s localization (%s) updated.\n", iap.Str("productId"), iapLocale)
		return nil
	},
}

var iapPriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Set an in-app purchase's price by customer price in a base territory",
	Long: `Set the price by matching a customer price in a base territory to an App
Store price point, then creating a price schedule effective immediately.

Apple derives every other territory's price from the base territory's price
point. --price is the customer-facing price in the territory's currency
(e.g. 150 for ¥150 with --territory JPN).`,
	Example: `  asc iap price --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --territory JPN --price 150`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, iapApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, iapProduct)
		if err != nil {
			return err
		}
		points, err := c.List(ctx, "/v2/inAppPurchases/"+iap.ID+"/pricePoints?filter[territory]="+iapTerr+"&limit=200")
		if err != nil {
			return err
		}
		want, err := strconv.ParseFloat(iapPrice, 64)
		if err != nil {
			return fmt.Errorf("invalid --price %q: %w", iapPrice, err)
		}
		var pointID string
		var near []string
		for _, p := range points {
			cp := p.Str("customerPrice")
			if f, err := strconv.ParseFloat(cp, 64); err == nil {
				if f == want {
					pointID = p.ID
					break
				}
				near = append(near, cp)
			}
		}
		if pointID == "" {
			hint := ""
			if len(near) > 0 {
				if len(near) > 12 {
					near = near[:12]
				}
				hint = " Available nearby: " + strings.Join(near, ", ")
			}
			return fmt.Errorf("no price point with customerPrice=%s in %s.%s", iapPrice, iapTerr, hint)
		}
		const lid = "${price1}"
		manual, _ := json.Marshal(map[string]any{"data": []map[string]string{{"type": "inAppPurchasePrices", "id": lid}}})
		_, err = c.Post(ctx, "/v1/inAppPurchasePriceSchedules", api.Body{
			Data: api.Resource{
				Type: "inAppPurchasePriceSchedules",
				Relationships: map[string]json.RawMessage{
					"inAppPurchase": api.Rel("inAppPurchases", iap.ID),
					"baseTerritory": api.Rel("territories", iapTerr),
					"manualPrices":  manual,
				},
			},
			Included: []api.Resource{{
				Type:       "inAppPurchasePrices",
				ID:         lid,
				Attributes: map[string]any{"startDate": nil},
				Relationships: map[string]json.RawMessage{
					"inAppPurchasePricePoint": api.Rel("inAppPurchasePricePoints", pointID),
				},
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("IAP %s price set to %s %s (price point %s).\n", iap.Str("productId"), iapPrice, iapTerr, pointID)
		return nil
	},
}

var iapScreenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Upload the review screenshot for an in-app purchase",
	Long: `Upload the App Review screenshot for an in-app purchase.

IMPORTANT: the IAP review-screenshot validator only accepts LEGACY device
sizes — current App Store screenshot sizes (1290×2796, 1320×2868, 1284×2778,
1170×2532, ...) are all rejected asynchronously with IMAGE_INCORRECT_DIMENSIONS.
Known-accepted sizes: 1242×2208, 2208×1242, 2048×2732, 2732×2048. The command
validates dimensions before uploading; pass --auto-fit to scale and pad the
image to an accepted size automatically. After upload it waits for Apple's
validation and reports the IAP state (READY_TO_SUBMIT when metadata is done).`,
	Example: `  asc iap screenshot --app 6790641087 --product com.ideamans.JapanReceiptScan.credits100 \
    --file app-store/iap-review.png --auto-fit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, iapApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, iapProduct)
		if err != nil {
			return err
		}
		data, fileName, err := prepareReviewScreenshot(iapFile, iapAutoFit)
		if err != nil {
			return err
		}
		id, err := uploadAssetBytes(ctx, c, assetSpec{
			reserveType: "inAppPurchaseAppStoreReviewScreenshots",
			relName:     "inAppPurchaseV2",
			relType:     "inAppPurchases",
			relID:       iap.ID,
		}, data, fileName)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded IAP review screenshot -> %s\n", id)
		if !c.DryRun {
			if after, err := resolveIAP(ctx, c, appID, iapProduct); err == nil {
				fmt.Printf("IAP state: %s\n", after.Str("state"))
			}
		}
		return nil
	},
}

var iapSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit an in-app purchase for review",
	Long: `Submit an in-app purchase for review on its own (create an
inAppPurchaseSubmission). For a first app submission, IAPs are usually reviewed
together with the app version; use this for IAPs added or changed later.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, iapApp)
		if err != nil {
			return err
		}
		iap, err := resolveIAP(ctx, c, appID, iapProduct)
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/inAppPurchaseSubmissions", api.Body{
			Data: api.Resource{
				Type:          "inAppPurchaseSubmissions",
				Relationships: map[string]json.RawMessage{"inAppPurchaseV2": api.Rel("inAppPurchases", iap.ID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted IAP %s for review.\n", iap.Str("productId"))
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{iapListCmd, iapLocalizeCmd, iapPriceCmd, iapScreenshotCmd, iapSubmitCmd} {
		sub.Flags().StringVar(&iapApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	for _, sub := range []*cobra.Command{iapLocalizeCmd, iapPriceCmd, iapScreenshotCmd, iapSubmitCmd} {
		sub.Flags().StringVar(&iapProduct, "product", "", "product id or ASC in-app purchase id (required)")
		_ = sub.MarkFlagRequired("product")
	}
	iapLocalizeCmd.Flags().StringVar(&iapLocale, "locale", "ja", "locale, e.g. ja / en-US")
	iapLocalizeCmd.Flags().StringVar(&iapName, "name", "", "display name (max 30 chars)")
	iapLocalizeCmd.Flags().StringVar(&iapDesc, "description", "", "description (max 45 chars, @file allowed)")

	iapPriceCmd.Flags().StringVar(&iapTerr, "territory", "JPN", "base territory code, e.g. JPN, USA")
	iapPriceCmd.Flags().StringVar(&iapPrice, "price", "", "customer price in the territory currency, e.g. 150 (required)")
	_ = iapPriceCmd.MarkFlagRequired("price")

	iapScreenshotCmd.Flags().StringVar(&iapFile, "file", "", "review screenshot file (required)")
	iapScreenshotCmd.Flags().BoolVar(&iapAutoFit, "auto-fit", false, "scale and pad the image to an accepted review-screenshot size (keeps aspect ratio, white padding)")
	_ = iapScreenshotCmd.MarkFlagRequired("file")

	iapCmd.AddCommand(iapListCmd, iapLocalizeCmd, iapPriceCmd, iapScreenshotCmd, iapSubmitCmd)
	rootCmd.AddCommand(iapCmd)
}
