package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var prcPricingCmd = &cobra.Command{
	Use:   "pricing",
	Short: "Manage the app's price schedule and price points",
}

var prcAvailabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Manage the app's territory availability",
}

var (
	prcApp         string
	prcPrice       string
	prcBaseTerr    string
	prcStartDate   string
	prcTerr        string
	prcLimit       int
	prcAutomatic   bool
	prcTerritories string
	prcAvailNew    bool
	prcFree        bool
)

// prcListInc fetches a collection and its included resources, following
// pagination to the end (c.List drops the "included" member, so price and
// availability listings that need include= go through this instead).
func prcListInc(ctx context.Context, c *api.Client, path string) ([]api.Resource, []api.Resource, error) {
	var data, included []api.Resource
	next := path
	for next != "" {
		raw, err := c.Do(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, nil, err
		}
		var doc struct {
			Data     []api.Resource `json:"data"`
			Included []api.Resource `json:"included"`
			Links    struct {
				Next string `json:"next"`
			} `json:"links"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, nil, err
		}
		data = append(data, doc.Data...)
		included = append(included, doc.Included...)
		next = doc.Links.Next
	}
	return data, included, nil
}

// prcIndex maps included resources by "type/id" for quick lookup.
func prcIndex(included []api.Resource) map[string]*api.Resource {
	idx := make(map[string]*api.Resource, len(included))
	for i := range included {
		idx[included[i].Type+"/"+included[i].ID] = &included[i]
	}
	return idx
}

// prcRelID returns the id of a to-one relationship, or "" when the response
// carried no relationship data (relationship ids are only present when the
// related resource was requested via include=).
func prcRelID(r *api.Resource, name string) string {
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

// prcBool formats a bool attribute, or "" when absent.
func prcBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// prcDash returns s, or "-" when empty.
func prcDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// prcFindAppPricePoint resolves a customer price in a territory to an app
// price point id, mirroring the IAP price-point resolution in cmd/iap.go.
func prcFindAppPricePoint(ctx context.Context, c *api.Client, appID, territory, price string) (string, error) {
	points, err := c.List(ctx, "/v1/apps/"+appID+"/appPricePoints?filter[territory]="+territory+"&limit=200")
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
	return "", fmt.Errorf("no app price point with customerPrice=%s in %s.%s", price, territory, hint)
}

// prcPrintPrices prints price rows with their included price point and territory.
func prcPrintPrices(prices, included []api.Resource, pointRel, pointType string) error {
	idx := prcIndex(included)
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TERRITORY\tCURRENCY\tPRICE\tPROCEEDS\tSTART\tEND")
	for i := range prices {
		p := &prices[i]
		terrID := prcRelID(p, "territory")
		terr := idx["territories/"+terrID]
		point := idx[pointType+"/"+prcRelID(p, pointRel)]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			terrID, terr.Str("currency"),
			point.Str("customerPrice"), point.Str("proceeds"),
			prcDash(p.Str("startDate")), prcDash(p.Str("endDate")))
	}
	return w.Flush()
}

var prcShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show the app's price schedule (base territory and manual prices)",
	Example: `  asc pricing show --app 6790641087 --automatic`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		sched, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/appPriceSchedule")
		if err != nil {
			return err
		}
		if sched == nil || sched.ID == "" {
			fmt.Printf("App %s has no price schedule. Set one with \"asc pricing set\".\n", appID)
			return nil
		}
		base, err := c.GetOptional(ctx, "/v1/appPriceSchedules/"+sched.ID+"/baseTerritory")
		if err != nil {
			return err
		}
		if base != nil && base.ID != "" {
			fmt.Printf("Base territory: %s (%s)\n", base.ID, base.Str("currency"))
		}
		manual, inc, err := prcListInc(ctx, c, "/v1/appPriceSchedules/"+sched.ID+"/manualPrices?include=appPricePoint,territory&limit=200")
		if err != nil {
			// The live API returns 404 on the prices relationships for apps whose
			// price was never configured through the price-schedule flow.
			if api.IsNotFound(err) {
				fmt.Println("No prices configured yet. Set one with \"asc pricing set\".")
				return nil
			}
			return err
		}
		fmt.Printf("Manual prices (%d):\n", len(manual))
		if err := prcPrintPrices(manual, inc, "appPricePoint", "appPricePoints"); err != nil {
			return err
		}
		if prcAutomatic {
			auto, incA, err := prcListInc(ctx, c, "/v1/appPriceSchedules/"+sched.ID+"/automaticPrices?include=appPricePoint,territory&limit=200")
			if err != nil {
				if api.IsNotFound(err) {
					fmt.Println("\nNo automatic prices returned.")
					return nil
				}
				return err
			}
			fmt.Printf("\nAutomatic prices (%d):\n", len(auto))
			if err := prcPrintPrices(auto, incA, "appPricePoint", "appPricePoints"); err != nil {
				return err
			}
		}
		return nil
	},
}

var prcSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the app's price by customer price in a base territory",
	Long: `Set the price by matching a customer price in a base territory to an App
Store price point, then creating a price schedule. Apple derives every other
territory's price from the base territory's price point. --price is the
customer-facing price in the territory's currency (e.g. 300 for ¥300 with
--base-territory JPN). Without --start-date the price takes effect immediately.

Free apps also need a price schedule (a required submission item that is easy
to miss): pass --free to resolve the customerPrice "0" price point.`,
	Example: `  asc pricing set --app 6790641087 --free
  asc pricing set --app 6790641087 --price 300 --base-territory JPN --start-date 2026-08-01`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if prcFree {
			prcPrice = "0"
		}
		if prcPrice == "" {
			return fmt.Errorf("pass --price <amount> or --free")
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		pointID, err := prcFindAppPricePoint(ctx, c, appID, prcBaseTerr, prcPrice)
		if err != nil {
			return err
		}
		var startDate any
		if prcStartDate != "" {
			startDate = prcStartDate
		}
		const lid = "${price1}"
		manual, _ := json.Marshal(map[string]any{"data": []map[string]string{{"type": "appPrices", "id": lid}}})
		_, err = c.Post(ctx, "/v1/appPriceSchedules", api.Body{
			Data: api.Resource{
				Type: "appPriceSchedules",
				Relationships: map[string]json.RawMessage{
					"app":           api.Rel("apps", appID),
					"baseTerritory": api.Rel("territories", prcBaseTerr),
					"manualPrices":  manual,
				},
			},
			Included: []api.Resource{{
				Type:       "appPrices",
				ID:         lid,
				Attributes: map[string]any{"startDate": startDate},
				Relationships: map[string]json.RawMessage{
					"appPricePoint": api.Rel("appPricePoints", pointID),
				},
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("App %s price set to %s %s (price point %s).\n", appID, prcPrice, prcBaseTerr, pointID)
		return nil
	},
}

var prcPointsCmd = &cobra.Command{
	Use:     "points",
	Short:   "List app price points for a territory",
	Example: `  asc pricing points --app 6790641087 --territory JPN --limit 30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		points, err := c.List(ctx, "/v1/apps/"+appID+"/appPricePoints?filter[territory]="+prcTerr+"&limit=200")
		if err != nil {
			return err
		}
		if prcLimit > 0 && len(points) > prcLimit {
			points = points[:prcLimit]
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PRICE\tPROCEEDS\tID")
		for _, p := range points {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Str("customerPrice"), p.Str("proceeds"), p.ID)
		}
		return w.Flush()
	},
}

var prcTerritoriesCmd = &cobra.Command{
	Use:   "territories",
	Short: "List App Store territories and their currencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		terrs, err := c.List(cmd.Context(), "/v1/territories?limit=200")
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

var prcAvailShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show the app's territory availability and pre-order status",
	Example: `  asc availability show --app 6790641087`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		avail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/appAvailabilityV2")
		if err != nil {
			return err
		}
		if avail == nil || avail.ID == "" {
			fmt.Printf("App %s has no availability configuration. Set one with \"asc availability set\".\n", appID)
			return nil
		}
		fmt.Printf("Available in new territories: %s\n", prcDash(prcBool(avail, "availableInNewTerritories")))
		tas, _, err := prcListInc(ctx, c, "/v2/appAvailabilities/"+avail.ID+"/territoryAvailabilities?include=territory&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TERRITORY\tAVAILABLE\tRELEASE DATE\tPREORDER\tPREORDER PUBLISH")
		for i := range tas {
			ta := &tas[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				prcRelID(ta, "territory"),
				prcDash(prcBool(ta, "available")),
				prcDash(ta.Str("releaseDate")),
				prcDash(prcBool(ta, "preOrderEnabled")),
				prcDash(ta.Str("preOrderPublishDate")))
		}
		return w.Flush()
	},
}

var prcAvailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the territories where the app is available",
	Long: `Replace the app's availability with the given territory list. Territories
not listed become unavailable. The API requires a territoryAvailability entry
for EVERY App Store territory (~175), so the command fetches the full territory
list and marks the ones you pass available and all others unavailable. Pass
--available-in-new-territories to opt in to territories Apple adds later.`,
	Example: `  asc availability set --app 6790641087 --territories JPN,USA,GBR
  asc availability set --app 6790641087 --territories JPN --available-in-new-territories`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		wanted := map[string]bool{}
		for _, t := range strings.Split(prcTerritories, ",") {
			if t = strings.ToUpper(strings.TrimSpace(t)); t != "" {
				wanted[t] = true
			}
		}
		if len(wanted) == 0 {
			return fmt.Errorf("--territories must list at least one territory code, e.g. JPN,USA")
		}
		// The API rejects partial lists ("expected included territoryAvailability
		// for id 'SVN'..."), so every territory must be sent explicitly.
		all, err := c.List(ctx, "/v1/territories?limit=200")
		if err != nil {
			return err
		}
		known := map[string]bool{}
		for _, t := range all {
			known[t.ID] = true
		}
		for t := range wanted {
			if !known[t] {
				return fmt.Errorf("unknown territory %q (see: asc territories)", t)
			}
		}
		refs := make([]map[string]string, 0, len(all))
		included := make([]api.Resource, 0, len(all))
		available := 0
		for i, t := range all {
			lid := fmt.Sprintf("${ta%d}", i+1)
			refs = append(refs, map[string]string{"type": "territoryAvailabilities", "id": lid})
			if wanted[t.ID] {
				available++
			}
			included = append(included, api.Resource{
				Type:          "territoryAvailabilities",
				ID:            lid,
				Attributes:    map[string]any{"available": wanted[t.ID]},
				Relationships: map[string]json.RawMessage{"territory": api.Rel("territories", t.ID)},
			})
		}
		taRel, _ := json.Marshal(map[string]any{"data": refs})
		_, err = c.Post(ctx, "/v2/appAvailabilities", api.Body{
			Data: api.Resource{
				Type:       "appAvailabilities",
				Attributes: map[string]any{"availableInNewTerritories": prcAvailNew},
				Relationships: map[string]json.RawMessage{
					"app":                     api.Rel("apps", appID),
					"territoryAvailabilities": taRel,
				},
			},
			Included: included,
		})
		if err != nil {
			return err
		}
		fmt.Printf("App %s availability set: available in %d of %d territories (availableInNewTerritories=%v).\n",
			appID, available, len(all), prcAvailNew)
		return nil
	},
}

var prcAvailEndPreorderCmd = &cobra.Command{
	Use:     "end-preorder",
	Short:   "End the app's pre-order in every territory where it is enabled",
	Example: `  asc availability end-preorder --app 6790641087`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, prcApp)
		if err != nil {
			return err
		}
		avail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/appAvailabilityV2")
		if err != nil {
			return err
		}
		if avail == nil || avail.ID == "" {
			return fmt.Errorf("app %s has no availability configuration", appID)
		}
		tas, err := c.List(ctx, "/v2/appAvailabilities/"+avail.ID+"/territoryAvailabilities?limit=200")
		if err != nil {
			return err
		}
		var refs []map[string]string
		for i := range tas {
			if v, ok := tas[i].Attributes["preOrderEnabled"].(bool); ok && v {
				refs = append(refs, map[string]string{"type": "territoryAvailabilities", "id": tas[i].ID})
			}
		}
		if len(refs) == 0 {
			fmt.Printf("App %s has no territories with an active pre-order.\n", appID)
			return nil
		}
		taRel, _ := json.Marshal(map[string]any{"data": refs})
		_, err = c.Post(ctx, "/v1/endAppAvailabilityPreOrders", api.Body{
			Data: api.Resource{
				Type:          "endAppAvailabilityPreOrders",
				Relationships: map[string]json.RawMessage{"territoryAvailabilities": taRel},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Ended pre-order in %d territories.\n", len(refs))
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{prcShowCmd, prcSetCmd, prcPointsCmd, prcAvailShowCmd, prcAvailSetCmd, prcAvailEndPreorderCmd} {
		sub.Flags().StringVar(&prcApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	prcShowCmd.Flags().BoolVar(&prcAutomatic, "automatic", false, "also list automatic (derived) prices for every territory")

	prcSetCmd.Flags().StringVar(&prcPrice, "price", "", "customer price in the base territory currency, e.g. 0.99 or 300")
	prcSetCmd.Flags().BoolVar(&prcFree, "free", false, "make the app free (shorthand for --price 0)")
	prcSetCmd.Flags().StringVar(&prcBaseTerr, "base-territory", "USA", "base territory code, e.g. USA, JPN")
	prcSetCmd.Flags().StringVar(&prcStartDate, "start-date", "", "start date YYYY-MM-DD (default: effective immediately)")

	prcPointsCmd.Flags().StringVar(&prcTerr, "territory", "JPN", "territory code, e.g. JPN, USA")
	prcPointsCmd.Flags().IntVar(&prcLimit, "limit", 0, "maximum number of price points to print (default: all)")

	prcAvailSetCmd.Flags().StringVar(&prcTerritories, "territories", "", "comma-separated territory codes, e.g. JPN,USA,GBR (required)")
	prcAvailSetCmd.Flags().BoolVar(&prcAvailNew, "available-in-new-territories", false, "automatically become available in future new territories")
	_ = prcAvailSetCmd.MarkFlagRequired("territories")

	prcPricingCmd.AddCommand(prcShowCmd, prcSetCmd, prcPointsCmd)
	prcAvailabilityCmd.AddCommand(prcAvailShowCmd, prcAvailSetCmd, prcAvailEndPreorderCmd)
	rootCmd.AddCommand(prcPricingCmd, prcAvailabilityCmd, prcTerritoriesCmd)
}
