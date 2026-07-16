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

var subsCmd = &cobra.Command{
	Use:     "subscriptions",
	Aliases: []string{"subs"},
	Short:   "Manage auto-renewable subscriptions (groups, prices, offers, availability, review)",
}

var (
	subsApp           string
	subsSub           string
	subsGroupID       string
	subsName          string
	subsProductID     string
	subsPeriodFlag    string
	subsGroupLevel    int
	subsFamShare      bool
	subsReviewNote    string
	subsLocale        string
	subsDesc          string
	subsCustomAppName string
	subsTerr          string
	subsPriceFlag     string
	subsStartDate     string
	subsEndDate       string
	subsPreserve      bool
	subsDuration      string
	subsMode          string
	subsPeriods       int
	subsOfferCode     string
	subsEligibility   string
	subsCustElig      []string
	subsID            string
	subsCount         int
	subsExpiration    string
	subsOutput        string
	subsTerritories   []string
	subsAvailNew      bool
	subsAvailPlanType string
	subsFile          string
	subsOptIn         bool
	subsSandboxOptIn  bool
	subsRenewalType   string
	subsVisibleAll    bool
	subsEnabled       bool
	subsIAP           string
)

// --- enum helpers ---------------------------------------------------------

// subsPeriodEnum maps ISO-8601 shorthand and enum names to the API's
// subscriptionPeriod enum (verified against the OpenAPI spec).
var subsPeriodEnum = map[string]string{
	"P1W": "ONE_WEEK", "P1M": "ONE_MONTH", "P2M": "TWO_MONTHS",
	"P3M": "THREE_MONTHS", "P6M": "SIX_MONTHS", "P1Y": "ONE_YEAR",
	"ONE_WEEK": "ONE_WEEK", "ONE_MONTH": "ONE_MONTH", "TWO_MONTHS": "TWO_MONTHS",
	"THREE_MONTHS": "THREE_MONTHS", "SIX_MONTHS": "SIX_MONTHS", "ONE_YEAR": "ONE_YEAR",
}

// subsDurationEnum maps shorthand to the SubscriptionOfferDuration enum.
var subsDurationEnum = map[string]string{
	"P3D": "THREE_DAYS", "P1W": "ONE_WEEK", "P2W": "TWO_WEEKS", "P1M": "ONE_MONTH",
	"P2M": "TWO_MONTHS", "P3M": "THREE_MONTHS", "P6M": "SIX_MONTHS", "P1Y": "ONE_YEAR",
	"THREE_DAYS": "THREE_DAYS", "ONE_WEEK": "ONE_WEEK", "TWO_WEEKS": "TWO_WEEKS",
	"ONE_MONTH": "ONE_MONTH", "TWO_MONTHS": "TWO_MONTHS", "THREE_MONTHS": "THREE_MONTHS",
	"SIX_MONTHS": "SIX_MONTHS", "ONE_YEAR": "ONE_YEAR",
}

// subsGraceDurationEnum maps shorthand to the SubscriptionGracePeriodDuration enum.
var subsGraceDurationEnum = map[string]string{
	"P3D": "THREE_DAYS", "P16D": "SIXTEEN_DAYS", "P28D": "TWENTY_EIGHT_DAYS",
	"THREE_DAYS": "THREE_DAYS", "SIXTEEN_DAYS": "SIXTEEN_DAYS", "TWENTY_EIGHT_DAYS": "TWENTY_EIGHT_DAYS",
}

func subsEnum(m map[string]string, v, what, valid string) (string, error) {
	if e, ok := m[strings.ToUpper(v)]; ok {
		return e, nil
	}
	return "", fmt.Errorf("invalid %s %q (valid: %s)", what, v, valid)
}

func subsOfferModeValue(v string) (string, error) {
	u := strings.ToUpper(v)
	switch u {
	case "FREE_TRIAL", "PAY_AS_YOU_GO", "PAY_UP_FRONT":
		return u, nil
	}
	return "", fmt.Errorf("invalid --offer-mode %q (valid: FREE_TRIAL, PAY_AS_YOU_GO, PAY_UP_FRONT)", v)
}

// --- JSON:API helpers ------------------------------------------------------

// subsRelMany builds a to-many relationship value: {"data":[{"type":..,"id":..},...]}.
func subsRelMany(typ string, ids []string) json.RawMessage {
	items := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]string{"type": typ, "id": id})
	}
	b, _ := json.Marshal(map[string]any{"data": items})
	return b
}

// subsRelID extracts a to-one relationship's resource id from a response
// resource (present when the relationship was requested via include).
func subsRelID(r *api.Resource, name string) string {
	raw, ok := r.Relationships[name]
	if !ok {
		return ""
	}
	var v struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	return v.Data.ID
}

// subsListWithIncluded fetches a collection with its included resources,
// following pagination (api.Client.List drops "included").
func subsListWithIncluded(ctx context.Context, c *api.Client, path string) ([]api.Resource, []api.Resource, error) {
	var out, inc []api.Resource
	next := path
	for next != "" {
		data, err := c.Do(ctx, http.MethodGet, next, nil)
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
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, nil, err
		}
		out = append(out, doc.Data...)
		inc = append(inc, doc.Included...)
		next = doc.Links.Next
	}
	return out, inc, nil
}

func subsFindIncluded(inc []api.Resource, typ, id string) *api.Resource {
	for i := range inc {
		if inc[i].Type == typ && inc[i].ID == id {
			return &inc[i]
		}
	}
	return nil
}

// subsAttr formats an attribute for table output ("-" when absent or null).
func subsAttr(r *api.Resource, key string) string {
	v, ok := r.Attributes[key]
	if !ok || v == nil {
		return "-"
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return "-"
		}
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// --- resolution helpers ----------------------------------------------------

// subsGroups lists the app's subscription groups.
func subsGroups(ctx context.Context, c *api.Client, appID string) ([]api.Resource, error) {
	return c.List(ctx, "/v1/apps/"+appID+"/subscriptionGroups?limit=200")
}

// subsResolve resolves --sub (an ASC subscription id or a productId) to the
// subscription resource by scanning the app's groups; a numeric ref that is
// not found among them is tried as a direct id.
func subsResolve(ctx context.Context, c *api.Client, appID, ref string) (*api.Resource, error) {
	if ref == "" {
		return nil, fmt.Errorf("--sub is required (subscription id or productId)")
	}
	groups, err := subsGroups(ctx, c, appID)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		list, err := c.List(ctx, "/v1/subscriptionGroups/"+groups[i].ID+"/subscriptions?limit=200")
		if err != nil {
			return nil, err
		}
		for j := range list {
			if list[j].ID == ref || list[j].Str("productId") == ref {
				return &list[j], nil
			}
		}
	}
	if isDigits(ref) {
		if r, _, err := c.Get(ctx, "/v1/subscriptions/"+ref); err == nil {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no subscription %q on app %s", ref, appID)
}

// subsPricePoint resolves a customer price in a territory to a subscription
// price point id, mirroring the IAP price-point matching.
func subsPricePoint(ctx context.Context, c *api.Client, subID, territory, price string) (string, error) {
	points, err := c.List(ctx, "/v1/subscriptions/"+subID+"/pricePoints?filter[territory]="+territory+"&limit=200")
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
	return "", fmt.Errorf("no subscription price point with customerPrice=%s in %s.%s", price, territory, hint)
}

// --- groups ----------------------------------------------------------------

var subsGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage subscription groups and their localizations",
}

var subsGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's subscription groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		groups, err := subsGroups(ctx, c, appID)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REFERENCE_NAME\tID")
		for _, g := range groups {
			fmt.Fprintf(w, "%s\t%s\n", g.Str("referenceName"), g.ID)
		}
		return w.Flush()
	},
}

var subsGroupsCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a subscription group",
	Example: `  asc subscriptions groups create --app com.example.app --name "Premium Plans"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/subscriptionGroups", api.Body{
			Data: api.Resource{
				Type:          "subscriptionGroups",
				Attributes:    map[string]any{"referenceName": subsName},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created subscription group %q -> %s\n", subsName, created.ID)
		return nil
	},
}

var subsGroupsLocalizeCmd = &cobra.Command{
	Use:     "localize",
	Short:   "Set a subscription group's localized display name",
	Example: `  asc subscriptions groups localize --group-id 21489000 --locale ja --name "プレミアムプラン"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = subsName
		}
		if cmd.Flags().Changed("custom-app-name") {
			attrs["customAppName"] = subsCustomAppName
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name and/or --custom-app-name")
		}
		locs, err := c.List(ctx, "/v1/subscriptionGroups/"+subsGroupID+"/subscriptionGroupLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", subsLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/subscriptionGroupLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "subscriptionGroupLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			if _, ok := attrs["name"]; !ok {
				return fmt.Errorf("--name is required when creating a new localization")
			}
			attrs["locale"] = subsLocale
			_, err = c.Post(ctx, "/v1/subscriptionGroupLocalizations", api.Body{
				Data: api.Resource{
					Type:          "subscriptionGroupLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"subscriptionGroup": api.Rel("subscriptionGroups", subsGroupID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Subscription group %s localization (%s) updated.\n", subsGroupID, subsLocale)
		return nil
	},
}

var subsGroupsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subscription group",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/subscriptionGroups/"+subsGroupID); err != nil {
			return err
		}
		fmt.Printf("Deleted subscription group %s.\n", subsGroupID)
		return nil
	},
}

// --- subscriptions ----------------------------------------------------------

var subsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscriptions across the app's subscription groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		groups, err := subsGroups(ctx, c, appID)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PRODUCT_ID\tNAME\tPERIOD\tSTATE\tGROUP\tID")
		for _, g := range groups {
			list, err := c.List(ctx, "/v1/subscriptionGroups/"+g.ID+"/subscriptions?limit=200")
			if err != nil {
				return err
			}
			for _, s := range list {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					s.Str("productId"), s.Str("name"), subsAttr(&s, "subscriptionPeriod"),
					s.Str("state"), g.Str("referenceName"), s.ID)
			}
		}
		return w.Flush()
	},
}

var subsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an auto-renewable subscription in a group",
	Example: `  asc subscriptions create --app com.example.app --group-id 21489000 \
    --product-id com.example.app.pro.monthly --name "Pro Monthly" --period P1M`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		if _, err := resolveAppID(ctx, c, subsApp); err != nil {
			return err
		}
		period, err := subsEnum(subsPeriodEnum, subsPeriodFlag, "--period",
			"P1W/ONE_WEEK, P1M/ONE_MONTH, P2M/TWO_MONTHS, P3M/THREE_MONTHS, P6M/SIX_MONTHS, P1Y/ONE_YEAR")
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"name":               subsName,
			"productId":          subsProductID,
			"subscriptionPeriod": period,
		}
		if cmd.Flags().Changed("group-level") {
			attrs["groupLevel"] = subsGroupLevel
		}
		if cmd.Flags().Changed("family-sharable") {
			attrs["familySharable"] = subsFamShare
		}
		if cmd.Flags().Changed("review-note") {
			v, err := valueOrFile(subsReviewNote)
			if err != nil {
				return err
			}
			attrs["reviewNote"] = v
		}
		created, err := c.Post(ctx, "/v1/subscriptions", api.Body{
			Data: api.Resource{
				Type:          "subscriptions",
				Attributes:    attrs,
				Relationships: map[string]json.RawMessage{"group": api.Rel("subscriptionGroups", subsGroupID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created subscription %s -> %s\n", subsProductID, created.ID)
		return nil
	},
}

var subsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a subscription as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(sub, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var subsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a subscription's name, period, group level, family sharing or review note",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = subsName
		}
		if cmd.Flags().Changed("period") {
			period, err := subsEnum(subsPeriodEnum, subsPeriodFlag, "--period",
				"P1W/ONE_WEEK, P1M/ONE_MONTH, P2M/TWO_MONTHS, P3M/THREE_MONTHS, P6M/SIX_MONTHS, P1Y/ONE_YEAR")
			if err != nil {
				return err
			}
			attrs["subscriptionPeriod"] = period
		}
		if cmd.Flags().Changed("group-level") {
			attrs["groupLevel"] = subsGroupLevel
		}
		if cmd.Flags().Changed("family-sharable") {
			attrs["familySharable"] = subsFamShare
		}
		if cmd.Flags().Changed("review-note") {
			v, err := valueOrFile(subsReviewNote)
			if err != nil {
				return err
			}
			attrs["reviewNote"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass --name, --period, --group-level, --family-sharable and/or --review-note")
		}
		_, err = c.Patch(ctx, "/v1/subscriptions/"+sub.ID, api.Body{
			Data: api.Resource{Type: "subscriptions", ID: sub.ID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Updated subscription %s.\n", sub.Str("productId"))
		return nil
	},
}

var subsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subscription (only possible before it is approved)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		if err := c.Delete(ctx, "/v1/subscriptions/"+sub.ID); err != nil {
			return err
		}
		fmt.Printf("Deleted subscription %s (%s).\n", sub.Str("productId"), sub.ID)
		return nil
	},
}

var subsLocalizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "Set a subscription's localized display name and description",
	Example: `  asc subscriptions localize --app com.example.app --sub com.example.app.pro.monthly \
    --locale ja --name "プロ（月額）" --description "全機能が使える月額プラン"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = subsName
		}
		if cmd.Flags().Changed("description") {
			v, err := valueOrFile(subsDesc)
			if err != nil {
				return err
			}
			attrs["description"] = v
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --name and/or --description")
		}
		locs, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/subscriptionLocalizations?limit=50")
		if err != nil {
			return err
		}
		if existing := findByAttr(locs, "locale", subsLocale); existing != nil {
			_, err = c.Patch(ctx, "/v1/subscriptionLocalizations/"+existing.ID, api.Body{
				Data: api.Resource{Type: "subscriptionLocalizations", ID: existing.ID, Attributes: attrs},
			})
		} else {
			if _, ok := attrs["name"]; !ok {
				return fmt.Errorf("--name is required when creating a new localization")
			}
			attrs["locale"] = subsLocale
			_, err = c.Post(ctx, "/v1/subscriptionLocalizations", api.Body{
				Data: api.Resource{
					Type:          "subscriptionLocalizations",
					Attributes:    attrs,
					Relationships: map[string]json.RawMessage{"subscription": api.Rel("subscriptions", sub.ID)},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Subscription %s localization (%s) updated.\n", sub.Str("productId"), subsLocale)
		return nil
	},
}

// --- price -------------------------------------------------------------------

var subsPriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Manage subscription prices",
}

var subsPriceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a subscription's price by customer price in a territory",
	Long: `Set the price by matching a customer price in a territory to a subscription
price point, then creating a subscriptionPrice. Apple derives the other
territories' prices from this price point. Without --start-date the change is
scheduled immediately; --preserve-current keeps the current price for existing
subscribers.`,
	Example: `  asc subscriptions price set --app com.example.app --sub com.example.app.pro.monthly \
    --territory JPN --price 600`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		pointID, err := subsPricePoint(ctx, c, sub.ID, subsTerr, subsPriceFlag)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("start-date") {
			attrs["startDate"] = subsStartDate
		}
		if cmd.Flags().Changed("preserve-current") {
			attrs["preserveCurrentPrice"] = subsPreserve
		}
		body := api.Body{
			Data: api.Resource{
				Type:       "subscriptionPrices",
				Attributes: attrs,
				Relationships: map[string]json.RawMessage{
					"subscription":           api.Rel("subscriptions", sub.ID),
					"subscriptionPricePoint": api.Rel("subscriptionPricePoints", pointID),
					"territory":              api.Rel("territories", subsTerr),
				},
			},
		}
		if len(attrs) == 0 {
			body.Data.Attributes = nil
		}
		if _, err := c.Post(ctx, "/v1/subscriptionPrices", body); err != nil {
			return err
		}
		fmt.Printf("Subscription %s price set to %s %s (price point %s).\n",
			sub.Str("productId"), subsPriceFlag, subsTerr, pointID)
		return nil
	},
}

var subsPriceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's scheduled and current prices",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		prices, included, err := subsListWithIncluded(ctx, c,
			"/v1/subscriptions/"+sub.ID+"/prices?include=subscriptionPricePoint,territory&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TERRITORY\tPRICE\tSTART_DATE\tPRESERVED\tID")
		for i := range prices {
			p := &prices[i]
			terr := subsRelID(p, "territory")
			price := "-"
			if pp := subsFindIncluded(included, "subscriptionPricePoints", subsRelID(p, "subscriptionPricePoint")); pp != nil {
				price = pp.Str("customerPrice")
			}
			start := subsAttr(p, "startDate")
			if start == "-" {
				start = "(current)"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", terr, price, start, subsAttr(p, "preserved"), p.ID)
		}
		return w.Flush()
	},
}

// --- introductory offers -----------------------------------------------------

var subsIntroCmd = &cobra.Command{
	Use:   "intro-offer",
	Short: "Manage introductory offers (free trial, pay as you go, pay up front)",
}

var subsIntroSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create an introductory offer",
	Long: `Create an introductory offer. FREE_TRIAL needs no price; PAY_AS_YOU_GO and
PAY_UP_FRONT require --territory and --price to resolve a subscription price
point. Without --territory a FREE_TRIAL applies to all territories.`,
	Example: `  asc subscriptions intro-offer set --app com.example.app --sub com.example.app.pro.monthly \
    --duration P1W --offer-mode FREE_TRIAL`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		duration, err := subsEnum(subsDurationEnum, subsDuration, "--duration",
			"P3D, P1W, P2W, P1M, P2M, P3M, P6M, P1Y (or the enum names)")
		if err != nil {
			return err
		}
		mode, err := subsOfferModeValue(subsMode)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"duration":        duration,
			"offerMode":       mode,
			"numberOfPeriods": subsPeriods,
		}
		if cmd.Flags().Changed("start-date") {
			attrs["startDate"] = subsStartDate
		}
		if cmd.Flags().Changed("end-date") {
			attrs["endDate"] = subsEndDate
		}
		rels := map[string]json.RawMessage{"subscription": api.Rel("subscriptions", sub.ID)}
		if mode == "FREE_TRIAL" {
			if cmd.Flags().Changed("territory") {
				rels["territory"] = api.Rel("territories", subsTerr)
			}
		} else {
			if !cmd.Flags().Changed("price") {
				return fmt.Errorf("--price is required for offer mode %s", mode)
			}
			pointID, err := subsPricePoint(ctx, c, sub.ID, subsTerr, subsPriceFlag)
			if err != nil {
				return err
			}
			rels["territory"] = api.Rel("territories", subsTerr)
			rels["subscriptionPricePoint"] = api.Rel("subscriptionPricePoints", pointID)
		}
		_, err = c.Post(ctx, "/v1/subscriptionIntroductoryOffers", api.Body{
			Data: api.Resource{Type: "subscriptionIntroductoryOffers", Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created introductory offer (%s %s) for %s.\n", mode, duration, sub.Str("productId"))
		return nil
	},
}

var subsIntroListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's introductory offers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		offers, included, err := subsListWithIncluded(ctx, c,
			"/v1/subscriptions/"+sub.ID+"/introductoryOffers?include=territory,subscriptionPricePoint&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "MODE\tDURATION\tPERIODS\tTERRITORY\tPRICE\tSTART\tEND\tID")
		for i := range offers {
			o := &offers[i]
			terr := subsRelID(o, "territory")
			if terr == "" {
				terr = "(all)"
			}
			price := "-"
			if pp := subsFindIncluded(included, "subscriptionPricePoints", subsRelID(o, "subscriptionPricePoint")); pp != nil {
				price = pp.Str("customerPrice")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				o.Str("offerMode"), o.Str("duration"), subsAttr(o, "numberOfPeriods"),
				terr, price, subsAttr(o, "startDate"), subsAttr(o, "endDate"), o.ID)
		}
		return w.Flush()
	},
}

var subsIntroDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an introductory offer by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/subscriptionIntroductoryOffers/"+subsID); err != nil {
			return err
		}
		fmt.Println("Deleted introductory offer.")
		return nil
	},
}

// --- promotional offers ------------------------------------------------------

var subsPromoCmd = &cobra.Command{
	Use:   "promo-offer",
	Short: "Manage promotional offers for existing subscribers",
}

var subsPromoCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a promotional offer",
	Long: `Create a promotional offer with a price for one territory (Apple equalizes
the remaining territories from that price point). FREE_TRIAL needs no --price.`,
	Example: `  asc subscriptions promo-offer create --app com.example.app --sub com.example.app.pro.monthly \
    --name "Win back 50%" --offer-code WINBACK50 --duration P1M --offer-mode PAY_AS_YOU_GO \
    --periods 3 --territory JPN --price 300`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		duration, err := subsEnum(subsDurationEnum, subsDuration, "--duration",
			"P3D, P1W, P2W, P1M, P2M, P3M, P6M, P1Y (or the enum names)")
		if err != nil {
			return err
		}
		mode, err := subsOfferModeValue(subsMode)
		if err != nil {
			return err
		}
		const lid = "price-1"
		priceRels := map[string]json.RawMessage{"territory": api.Rel("territories", subsTerr)}
		if mode != "FREE_TRIAL" {
			if !cmd.Flags().Changed("price") {
				return fmt.Errorf("--price is required for offer mode %s", mode)
			}
			pointID, err := subsPricePoint(ctx, c, sub.ID, subsTerr, subsPriceFlag)
			if err != nil {
				return err
			}
			priceRels["subscriptionPricePoint"] = api.Rel("subscriptionPricePoints", pointID)
		}
		_, err = c.Post(ctx, "/v1/subscriptionPromotionalOffers", api.Body{
			Data: api.Resource{
				Type: "subscriptionPromotionalOffers",
				Attributes: map[string]any{
					"name":            subsName,
					"offerCode":       subsOfferCode,
					"duration":        duration,
					"offerMode":       mode,
					"numberOfPeriods": subsPeriods,
				},
				Relationships: map[string]json.RawMessage{
					"subscription": api.Rel("subscriptions", sub.ID),
					"prices":       subsRelMany("subscriptionPromotionalOfferPrices", []string{lid}),
				},
			},
			Included: []api.Resource{{
				Type:          "subscriptionPromotionalOfferPrices",
				ID:            lid,
				Relationships: priceRels,
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created promotional offer %q (code %s) for %s.\n", subsName, subsOfferCode, sub.Str("productId"))
		return nil
	},
}

var subsPromoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's promotional offers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		offers, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/promotionalOffers?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tOFFER_CODE\tMODE\tDURATION\tPERIODS\tID")
		for i := range offers {
			o := &offers[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				o.Str("name"), o.Str("offerCode"), o.Str("offerMode"), o.Str("duration"),
				subsAttr(o, "numberOfPeriods"), o.ID)
		}
		return w.Flush()
	},
}

var subsPromoDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a promotional offer by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/subscriptionPromotionalOffers/"+subsID); err != nil {
			return err
		}
		fmt.Println("Deleted promotional offer.")
		return nil
	},
}

// --- offer codes ---------------------------------------------------------------

var subsCodesCmd = &cobra.Command{
	Use:   "offer-codes",
	Short: "Manage subscription offer codes and their one-time-use codes",
}

var subsCodesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an offer code campaign",
	Long: `Create an offer code with a price for one territory (Apple equalizes the
remaining territories from that price point). FREE_TRIAL needs no --price.`,
	Example: `  asc subscriptions offer-codes create --app com.example.app --sub com.example.app.pro.monthly \
    --name "Launch trial" --duration P1M --offer-mode FREE_TRIAL --periods 1 --territory JPN`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		duration, err := subsEnum(subsDurationEnum, subsDuration, "--duration",
			"P3D, P1W, P2W, P1M, P2M, P3M, P6M, P1Y (or the enum names)")
		if err != nil {
			return err
		}
		mode, err := subsOfferModeValue(subsMode)
		if err != nil {
			return err
		}
		elig := strings.ToUpper(subsEligibility)
		if elig != "STACK_WITH_INTRO_OFFERS" && elig != "REPLACE_INTRO_OFFERS" {
			return fmt.Errorf("invalid --eligibility %q (valid: STACK_WITH_INTRO_OFFERS, REPLACE_INTRO_OFFERS)", subsEligibility)
		}
		var custElig []string
		for _, e := range subsCustElig {
			u := strings.ToUpper(strings.TrimSpace(e))
			if u != "NEW" && u != "EXISTING" && u != "EXPIRED" {
				return fmt.Errorf("invalid --customer-eligibility %q (valid: NEW, EXISTING, EXPIRED)", e)
			}
			custElig = append(custElig, u)
		}
		const lid = "price-1"
		priceRels := map[string]json.RawMessage{"territory": api.Rel("territories", subsTerr)}
		if mode != "FREE_TRIAL" {
			if !cmd.Flags().Changed("price") {
				return fmt.Errorf("--price is required for offer mode %s", mode)
			}
			pointID, err := subsPricePoint(ctx, c, sub.ID, subsTerr, subsPriceFlag)
			if err != nil {
				return err
			}
			priceRels["subscriptionPricePoint"] = api.Rel("subscriptionPricePoints", pointID)
		}
		created, err := c.Post(ctx, "/v1/subscriptionOfferCodes", api.Body{
			Data: api.Resource{
				Type: "subscriptionOfferCodes",
				Attributes: map[string]any{
					"name":                  subsName,
					"duration":              duration,
					"offerMode":             mode,
					"numberOfPeriods":       subsPeriods,
					"offerEligibility":      elig,
					"customerEligibilities": custElig,
				},
				Relationships: map[string]json.RawMessage{
					"subscription": api.Rel("subscriptions", sub.ID),
					"prices":       subsRelMany("subscriptionOfferCodePrices", []string{lid}),
				},
			},
			Included: []api.Resource{{
				Type:          "subscriptionOfferCodePrices",
				ID:            lid,
				Relationships: priceRels,
			}},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created offer code campaign %q -> %s\n", subsName, created.ID)
		return nil
	},
}

var subsCodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's offer code campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		codes, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/offerCodes?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tACTIVE\tMODE\tDURATION\tPERIODS\tTOTAL_CODES\tID")
		for i := range codes {
			o := &codes[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				o.Str("name"), subsAttr(o, "active"), o.Str("offerMode"), o.Str("duration"),
				subsAttr(o, "numberOfPeriods"), subsAttr(o, "totalNumberOfCodes"), o.ID)
		}
		return w.Flush()
	},
}

var subsCodesDeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate an offer code campaign by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/subscriptionOfferCodes/"+subsID, api.Body{
			Data: api.Resource{Type: "subscriptionOfferCodes", ID: subsID, Attributes: map[string]any{"active": false}},
		})
		if err != nil {
			return err
		}
		fmt.Println("Deactivated offer code campaign.")
		return nil
	},
}

var subsCodesCreateCodesCmd = &cobra.Command{
	Use:     "create-codes",
	Short:   "Generate a batch of one-time-use codes for an offer code campaign",
	Example: `  asc subscriptions offer-codes create-codes --id 21500000 --count 100 --expiration 2026-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/subscriptionOfferCodeOneTimeUseCodes", api.Body{
			Data: api.Resource{
				Type: "subscriptionOfferCodeOneTimeUseCodes",
				Attributes: map[string]any{
					"numberOfCodes":  subsCount,
					"expirationDate": subsExpiration,
				},
				Relationships: map[string]json.RawMessage{"offerCode": api.Rel("subscriptionOfferCodes", subsID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created %d one-time-use codes (batch %s, expires %s).\n", subsCount, created.ID, subsExpiration)
		return nil
	},
}

var subsCodesDownloadCmd = &cobra.Command{
	Use:   "download-codes",
	Short: "Download a one-time-use code batch as CSV",
	Long: `Download the code values of a one-time-use code batch (the id printed by
create-codes, or listed via
"asc api /v1/subscriptionOfferCodes/{id}/oneTimeUseCodes"). Prints CSV to
stdout, or writes to --output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		data, err := c.Download(cmd.Context(), "/v1/subscriptionOfferCodeOneTimeUseCodes/"+subsID+"/values", "text/csv")
		if err != nil {
			return err
		}
		if subsOutput != "" {
			if err := os.WriteFile(subsOutput, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("Wrote %d bytes to %s\n", len(data), subsOutput)
			return nil
		}
		_, err = os.Stdout.Write(data)
		return err
	},
}

// --- availability --------------------------------------------------------------

var subsAvailCmd = &cobra.Command{
	Use:   "availability",
	Short: "Show or set the territories where a subscription is available",
}

var subsAvailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a subscription's territory availability",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		avails, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/planAvailabilities?limit=200")
		if err != nil {
			return err
		}
		if len(avails) == 0 {
			fmt.Printf("Subscription %s has no availability set.\n", sub.Str("productId"))
			return nil
		}
		for _, avail := range avails {
			fmt.Printf("planType: %s  availableInNewTerritories: %s\n",
				avail.Str("planType"), subsAttr(&avail, "availableInNewTerritories"))
			terrs, err := c.List(ctx, "/v1/subscriptionPlanAvailabilities/"+avail.ID+"/availableTerritories?limit=200")
			if err != nil {
				return err
			}
			codes := make([]string, 0, len(terrs))
			for _, t := range terrs {
				codes = append(codes, t.ID)
			}
			fmt.Printf("territories (%d): %s\n", len(codes), strings.Join(codes, ", "))
		}
		return nil
	},
}

var subsAvailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the territories where a subscription is available",
	Example: `  asc subscriptions availability set --app com.example.app --sub com.example.app.pro.monthly \
    --territories JPN,USA --available-in-new-territories`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		terrs := make([]string, 0, len(subsTerritories))
		for _, t := range subsTerritories {
			t = strings.ToUpper(strings.TrimSpace(t))
			if t != "" {
				terrs = append(terrs, t)
			}
		}
		if len(terrs) == 0 {
			return fmt.Errorf("--territories must list at least one territory code")
		}
		planType := strings.ToUpper(subsAvailPlanType)
		if planType != "MONTHLY" && planType != "UPFRONT" {
			return fmt.Errorf("--plan-type must be MONTHLY or UPFRONT")
		}
		// One planAvailability exists per plan type; PATCH it when present, POST otherwise.
		existing, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/planAvailabilities?limit=200")
		if err != nil {
			return err
		}
		if cur := findByAttr(existing, "planType", planType); cur != nil {
			_, err = c.Patch(ctx, "/v1/subscriptionPlanAvailabilities/"+cur.ID, api.Body{
				Data: api.Resource{
					Type:       "subscriptionPlanAvailabilities",
					ID:         cur.ID,
					Attributes: map[string]any{"availableInNewTerritories": subsAvailNew},
					Relationships: map[string]json.RawMessage{
						"availableTerritories": subsRelMany("territories", terrs),
					},
				},
			})
		} else {
			_, err = c.Post(ctx, "/v1/subscriptionPlanAvailabilities", api.Body{
				Data: api.Resource{
					Type: "subscriptionPlanAvailabilities",
					Attributes: map[string]any{
						"planType":                  planType,
						"availableInNewTerritories": subsAvailNew,
					},
					Relationships: map[string]json.RawMessage{
						"subscription":         api.Rel("subscriptions", sub.ID),
						"availableTerritories": subsRelMany("territories", terrs),
					},
				},
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Set %s availability of %s to %d territories.\n", planType, sub.Str("productId"), len(terrs))
		return nil
	},
}

// --- screenshot / submit ---------------------------------------------------------

var subsScreenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Upload the review screenshot for a subscription",
	Example: `  asc subscriptions screenshot --app com.example.app --sub com.example.app.pro.monthly \
    --file app-store/sub-review.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		id, err := uploadAsset(ctx, c, assetSpec{
			reserveType: "subscriptionAppStoreReviewScreenshots",
			relName:     "subscription",
			relType:     "subscriptions",
			relID:       sub.ID,
			filePath:    subsFile,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded subscription review screenshot -> %s\n", id)
		return nil
	},
}

var subsSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a subscription for review",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		_, err = c.Post(ctx, "/v1/subscriptionSubmissions", api.Body{
			Data: api.Resource{
				Type:          "subscriptionSubmissions",
				Relationships: map[string]json.RawMessage{"subscription": api.Rel("subscriptions", sub.ID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Submitted subscription %s for review.\n", sub.Str("productId"))
		return nil
	},
}

// --- grace period ------------------------------------------------------------------

var subsGraceCmd = &cobra.Command{
	Use:   "grace-period",
	Short: "Show or configure the app's billing grace period",
}

var subsGraceShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the app's billing grace period settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		gp, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/subscriptionGracePeriod")
		if err != nil {
			return err
		}
		if gp == nil || gp.ID == "" {
			fmt.Println("No grace period resource on this app.")
			return nil
		}
		b, err := json.MarshalIndent(gp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var subsGraceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update the app's billing grace period settings",
	Example: `  asc subscriptions grace-period set --app com.example.app \
    --opt-in --duration P16D --renewal-type ALL_RENEWALS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		gp, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/subscriptionGracePeriod")
		if err != nil {
			return err
		}
		if gp == nil || gp.ID == "" {
			return fmt.Errorf("app %s has no grace period resource to update", appID)
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("opt-in") {
			attrs["optIn"] = subsOptIn
		}
		if cmd.Flags().Changed("sandbox-opt-in") {
			attrs["sandboxOptIn"] = subsSandboxOptIn
		}
		if cmd.Flags().Changed("duration") {
			d, err := subsEnum(subsGraceDurationEnum, subsDuration, "--duration",
				"P3D/THREE_DAYS, P16D/SIXTEEN_DAYS, P28D/TWENTY_EIGHT_DAYS")
			if err != nil {
				return err
			}
			attrs["duration"] = d
		}
		if cmd.Flags().Changed("renewal-type") {
			rt := strings.ToUpper(subsRenewalType)
			if rt != "ALL_RENEWALS" && rt != "PAID_TO_PAID_ONLY" {
				return fmt.Errorf("invalid --renewal-type %q (valid: ALL_RENEWALS, PAID_TO_PAID_ONLY)", subsRenewalType)
			}
			attrs["renewalType"] = rt
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to set: pass --opt-in, --sandbox-opt-in, --duration and/or --renewal-type")
		}
		_, err = c.Patch(ctx, "/v1/subscriptionGracePeriods/"+gp.ID, api.Body{
			Data: api.Resource{Type: "subscriptionGracePeriods", ID: gp.ID, Attributes: attrs},
		})
		if err != nil {
			return err
		}
		fmt.Println("Updated grace period settings.")
		return nil
	},
}

// --- win-back offers -----------------------------------------------------------------

var subsWinBackCmd = &cobra.Command{
	Use:   "win-back",
	Short: "List or delete win-back offers (create is not supported; see help)",
	Long: `List or delete a subscription's win-back offers. Creating win-back offers is
not supported here: the API's inline price schema for winBackOffers carries no
territory or price-point relationships, so a reliable create cannot be built
from the published spec. Create them in App Store Connect, or use "asc api"
with a hand-crafted body.`,
}

var subsWinBackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a subscription's win-back offers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		sub, err := subsResolve(ctx, c, appID, subsSub)
		if err != nil {
			return err
		}
		offers, err := c.List(ctx, "/v1/subscriptions/"+sub.ID+"/winBackOffers?limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "REFERENCE_NAME\tOFFER_ID\tMODE\tDURATION\tPRIORITY\tSTART\tEND\tID")
		for i := range offers {
			o := &offers[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				o.Str("referenceName"), o.Str("offerId"), o.Str("offerMode"), o.Str("duration"),
				o.Str("priority"), subsAttr(o, "startDate"), subsAttr(o, "endDate"), o.ID)
		}
		return w.Flush()
	},
}

var subsWinBackDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a win-back offer by id",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/winBackOffers/"+subsID); err != nil {
			return err
		}
		fmt.Println("Deleted win-back offer.")
		return nil
	},
}

// --- promoted purchases ----------------------------------------------------------------

var subsPromotedCmd = &cobra.Command{
	Use:   "promoted",
	Short: "Manage promoted in-app purchases and subscriptions on the App Store",
}

var subsPromotedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the app's promoted purchases",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		promos, included, err := subsListWithIncluded(ctx, c,
			"/v1/apps/"+appID+"/promotedPurchases?include=inAppPurchaseV2,subscription&limit=200")
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PRODUCT_ID\tKIND\tVISIBLE_ALL\tENABLED\tSTATE\tID")
		for i := range promos {
			p := &promos[i]
			kind, productID := "-", "-"
			if id := subsRelID(p, "subscription"); id != "" {
				kind = "subscription"
				if r := subsFindIncluded(included, "subscriptions", id); r != nil {
					productID = r.Str("productId")
				}
			} else if id := subsRelID(p, "inAppPurchaseV2"); id != "" {
				kind = "iap"
				if r := subsFindIncluded(included, "inAppPurchases", id); r != nil {
					productID = r.Str("productId")
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				productID, kind, subsAttr(p, "visibleForAllUsers"), subsAttr(p, "enabled"), p.Str("state"), p.ID)
		}
		return w.Flush()
	},
}

var subsPromotedSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Promote a subscription or in-app purchase (creates or updates the promoted purchase)",
	Example: `  asc subscriptions promoted set --app com.example.app --sub com.example.app.pro.monthly \
    --visible-for-all-users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, subsApp)
		if err != nil {
			return err
		}
		if (subsSub == "") == (subsIAP == "") {
			return fmt.Errorf("pass exactly one of --sub or --iap")
		}
		relName, relType, targetID, label := "", "", "", ""
		if subsSub != "" {
			sub, err := subsResolve(ctx, c, appID, subsSub)
			if err != nil {
				return err
			}
			relName, relType, targetID, label = "subscription", "subscriptions", sub.ID, sub.Str("productId")
		} else {
			iap, err := resolveIAP(ctx, c, appID, subsIAP)
			if err != nil {
				return err
			}
			relName, relType, targetID, label = "inAppPurchaseV2", "inAppPurchases", iap.ID, iap.Str("productId")
		}
		existing, _, err := subsListWithIncluded(ctx, c,
			"/v1/apps/"+appID+"/promotedPurchases?include=inAppPurchaseV2,subscription&limit=200")
		if err != nil {
			return err
		}
		var existingID string
		for i := range existing {
			if subsRelID(&existing[i], relName) == targetID {
				existingID = existing[i].ID
				break
			}
		}
		if existingID != "" {
			attrs := map[string]any{}
			if cmd.Flags().Changed("visible-for-all-users") {
				attrs["visibleForAllUsers"] = subsVisibleAll
			}
			if cmd.Flags().Changed("enabled") {
				attrs["enabled"] = subsEnabled
			}
			if len(attrs) == 0 {
				return fmt.Errorf("%s is already promoted; pass --visible-for-all-users and/or --enabled to change it", label)
			}
			_, err = c.Patch(ctx, "/v1/promotedPurchases/"+existingID, api.Body{
				Data: api.Resource{Type: "promotedPurchases", ID: existingID, Attributes: attrs},
			})
			if err != nil {
				return err
			}
			fmt.Printf("Updated promoted purchase for %s.\n", label)
			return nil
		}
		attrs := map[string]any{"visibleForAllUsers": subsVisibleAll}
		if cmd.Flags().Changed("enabled") {
			attrs["enabled"] = subsEnabled
		}
		_, err = c.Post(ctx, "/v1/promotedPurchases", api.Body{
			Data: api.Resource{
				Type:       "promotedPurchases",
				Attributes: attrs,
				Relationships: map[string]json.RawMessage{
					"app":   api.Rel("apps", appID),
					relName: api.Rel(relType, targetID),
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Promoted %s on the App Store.\n", label)
		return nil
	},
}

// --- wiring ------------------------------------------------------------------------

func init() {
	appRequired := []*cobra.Command{
		subsGroupsListCmd, subsGroupsCreateCmd,
		subsListCmd, subsCreateCmd, subsShowCmd, subsUpdateCmd, subsDeleteCmd, subsLocalizeCmd,
		subsPriceSetCmd, subsPriceListCmd,
		subsIntroSetCmd, subsIntroListCmd,
		subsPromoCreateCmd, subsPromoListCmd,
		subsCodesCreateCmd, subsCodesListCmd,
		subsAvailShowCmd, subsAvailSetCmd,
		subsScreenshotCmd, subsSubmitCmd,
		subsGraceShowCmd, subsGraceSetCmd,
		subsWinBackListCmd,
		subsPromotedListCmd, subsPromotedSetCmd,
	}
	for _, sub := range appRequired {
		sub.Flags().StringVar(&subsApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	subRequired := []*cobra.Command{
		subsShowCmd, subsUpdateCmd, subsDeleteCmd, subsLocalizeCmd,
		subsPriceSetCmd, subsPriceListCmd,
		subsIntroSetCmd, subsIntroListCmd,
		subsPromoCreateCmd, subsPromoListCmd,
		subsCodesCreateCmd, subsCodesListCmd,
		subsAvailShowCmd, subsAvailSetCmd,
		subsScreenshotCmd, subsSubmitCmd,
		subsWinBackListCmd,
	}
	for _, sub := range subRequired {
		sub.Flags().StringVar(&subsSub, "sub", "", "subscription id or productId (required)")
		_ = sub.MarkFlagRequired("sub")
	}

	// groups
	subsGroupsCreateCmd.Flags().StringVar(&subsName, "name", "", "group reference name (required)")
	_ = subsGroupsCreateCmd.MarkFlagRequired("name")
	subsGroupsLocalizeCmd.Flags().StringVar(&subsGroupID, "group-id", "", "subscription group id (required)")
	_ = subsGroupsLocalizeCmd.MarkFlagRequired("group-id")
	subsGroupsLocalizeCmd.Flags().StringVar(&subsLocale, "locale", "ja", "locale, e.g. ja / en-US")
	subsGroupsLocalizeCmd.Flags().StringVar(&subsName, "name", "", "localized group display name")
	subsGroupsLocalizeCmd.Flags().StringVar(&subsCustomAppName, "custom-app-name", "", "custom app name shown with the group")
	subsGroupsDeleteCmd.Flags().StringVar(&subsGroupID, "group-id", "", "subscription group id (required)")
	_ = subsGroupsDeleteCmd.MarkFlagRequired("group-id")

	// create / update
	subsCreateCmd.Flags().StringVar(&subsGroupID, "group-id", "", "subscription group id (required)")
	subsCreateCmd.Flags().StringVar(&subsProductID, "product-id", "", "product id, e.g. com.example.app.pro.monthly (required)")
	subsCreateCmd.Flags().StringVar(&subsName, "name", "", "reference name (required)")
	subsCreateCmd.Flags().StringVar(&subsPeriodFlag, "period", "", "subscription period: P1W, P1M, P2M, P3M, P6M or P1Y (required)")
	subsCreateCmd.Flags().IntVar(&subsGroupLevel, "group-level", 0, "group level (1 = highest tier)")
	subsCreateCmd.Flags().BoolVar(&subsFamShare, "family-sharable", false, "enable Family Sharing")
	subsCreateCmd.Flags().StringVar(&subsReviewNote, "review-note", "", "review note (@file allowed)")
	_ = subsCreateCmd.MarkFlagRequired("group-id")
	_ = subsCreateCmd.MarkFlagRequired("product-id")
	_ = subsCreateCmd.MarkFlagRequired("name")
	_ = subsCreateCmd.MarkFlagRequired("period")

	subsUpdateCmd.Flags().StringVar(&subsName, "name", "", "reference name")
	subsUpdateCmd.Flags().StringVar(&subsPeriodFlag, "period", "", "subscription period: P1W, P1M, P2M, P3M, P6M or P1Y")
	subsUpdateCmd.Flags().IntVar(&subsGroupLevel, "group-level", 0, "group level (1 = highest tier)")
	subsUpdateCmd.Flags().BoolVar(&subsFamShare, "family-sharable", false, "enable Family Sharing")
	subsUpdateCmd.Flags().StringVar(&subsReviewNote, "review-note", "", "review note (@file allowed)")

	// localize
	subsLocalizeCmd.Flags().StringVar(&subsLocale, "locale", "ja", "locale, e.g. ja / en-US")
	subsLocalizeCmd.Flags().StringVar(&subsName, "name", "", "localized display name")
	subsLocalizeCmd.Flags().StringVar(&subsDesc, "description", "", "localized description (@file allowed)")

	// price
	subsPriceSetCmd.Flags().StringVar(&subsTerr, "territory", "JPN", "territory code, e.g. JPN, USA")
	subsPriceSetCmd.Flags().StringVar(&subsPriceFlag, "price", "", "customer price in the territory currency, e.g. 600 (required)")
	subsPriceSetCmd.Flags().StringVar(&subsStartDate, "start-date", "", "start date YYYY-MM-DD (default: immediately)")
	subsPriceSetCmd.Flags().BoolVar(&subsPreserve, "preserve-current", false, "keep the current price for existing subscribers")
	_ = subsPriceSetCmd.MarkFlagRequired("price")

	// intro-offer
	subsIntroSetCmd.Flags().StringVar(&subsDuration, "duration", "", "offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required)")
	subsIntroSetCmd.Flags().StringVar(&subsMode, "offer-mode", "", "FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required)")
	subsIntroSetCmd.Flags().IntVar(&subsPeriods, "periods", 1, "number of periods")
	subsIntroSetCmd.Flags().StringVar(&subsTerr, "territory", "JPN", "territory code (paid modes; optional for FREE_TRIAL)")
	subsIntroSetCmd.Flags().StringVar(&subsPriceFlag, "price", "", "customer price in the territory currency (paid modes)")
	subsIntroSetCmd.Flags().StringVar(&subsStartDate, "start-date", "", "start date YYYY-MM-DD")
	subsIntroSetCmd.Flags().StringVar(&subsEndDate, "end-date", "", "end date YYYY-MM-DD")
	_ = subsIntroSetCmd.MarkFlagRequired("duration")
	_ = subsIntroSetCmd.MarkFlagRequired("offer-mode")
	subsIntroDeleteCmd.Flags().StringVar(&subsID, "id", "", "subscriptionIntroductoryOffer id (required)")
	_ = subsIntroDeleteCmd.MarkFlagRequired("id")

	// promo-offer
	subsPromoCreateCmd.Flags().StringVar(&subsName, "name", "", "offer reference name (required)")
	subsPromoCreateCmd.Flags().StringVar(&subsOfferCode, "offer-code", "", "offer identifier used by your app, e.g. WINBACK50 (required)")
	subsPromoCreateCmd.Flags().StringVar(&subsDuration, "duration", "", "offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required)")
	subsPromoCreateCmd.Flags().StringVar(&subsMode, "offer-mode", "", "FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required)")
	subsPromoCreateCmd.Flags().IntVar(&subsPeriods, "periods", 1, "number of periods")
	subsPromoCreateCmd.Flags().StringVar(&subsTerr, "territory", "JPN", "territory code for the price")
	subsPromoCreateCmd.Flags().StringVar(&subsPriceFlag, "price", "", "customer price in the territory currency (paid modes)")
	_ = subsPromoCreateCmd.MarkFlagRequired("name")
	_ = subsPromoCreateCmd.MarkFlagRequired("offer-code")
	_ = subsPromoCreateCmd.MarkFlagRequired("duration")
	_ = subsPromoCreateCmd.MarkFlagRequired("offer-mode")
	subsPromoDeleteCmd.Flags().StringVar(&subsID, "id", "", "subscriptionPromotionalOffer id (required)")
	_ = subsPromoDeleteCmd.MarkFlagRequired("id")

	// offer-codes
	subsCodesCreateCmd.Flags().StringVar(&subsName, "name", "", "campaign name (required)")
	subsCodesCreateCmd.Flags().StringVar(&subsDuration, "duration", "", "offer duration: P3D, P1W, P2W, P1M, P2M, P3M, P6M or P1Y (required)")
	subsCodesCreateCmd.Flags().StringVar(&subsMode, "offer-mode", "", "FREE_TRIAL, PAY_AS_YOU_GO or PAY_UP_FRONT (required)")
	subsCodesCreateCmd.Flags().IntVar(&subsPeriods, "periods", 1, "number of periods")
	subsCodesCreateCmd.Flags().StringVar(&subsEligibility, "eligibility", "STACK_WITH_INTRO_OFFERS", "STACK_WITH_INTRO_OFFERS or REPLACE_INTRO_OFFERS")
	subsCodesCreateCmd.Flags().StringSliceVar(&subsCustElig, "customer-eligibility", []string{"NEW", "EXISTING", "EXPIRED"}, "customer eligibility: NEW, EXISTING, EXPIRED")
	subsCodesCreateCmd.Flags().StringVar(&subsTerr, "territory", "JPN", "territory code for the price")
	subsCodesCreateCmd.Flags().StringVar(&subsPriceFlag, "price", "", "customer price in the territory currency (paid modes)")
	_ = subsCodesCreateCmd.MarkFlagRequired("name")
	_ = subsCodesCreateCmd.MarkFlagRequired("duration")
	_ = subsCodesCreateCmd.MarkFlagRequired("offer-mode")
	subsCodesDeactivateCmd.Flags().StringVar(&subsID, "id", "", "subscriptionOfferCode id (required)")
	_ = subsCodesDeactivateCmd.MarkFlagRequired("id")
	subsCodesCreateCodesCmd.Flags().StringVar(&subsID, "id", "", "subscriptionOfferCode id (required)")
	subsCodesCreateCodesCmd.Flags().IntVar(&subsCount, "count", 0, "number of codes to generate (required)")
	subsCodesCreateCodesCmd.Flags().StringVar(&subsExpiration, "expiration", "", "expiration date YYYY-MM-DD (required)")
	_ = subsCodesCreateCodesCmd.MarkFlagRequired("id")
	_ = subsCodesCreateCodesCmd.MarkFlagRequired("count")
	_ = subsCodesCreateCodesCmd.MarkFlagRequired("expiration")
	subsCodesDownloadCmd.Flags().StringVar(&subsID, "id", "", "subscriptionOfferCodeOneTimeUseCodes batch id (required)")
	subsCodesDownloadCmd.Flags().StringVar(&subsOutput, "output", "", "write CSV to this file instead of stdout")
	_ = subsCodesDownloadCmd.MarkFlagRequired("id")

	// availability
	subsAvailSetCmd.Flags().StringSliceVar(&subsTerritories, "territories", nil, "territory codes, e.g. JPN,USA (required)")
	subsAvailSetCmd.Flags().BoolVar(&subsAvailNew, "available-in-new-territories", true, "automatically include future App Store territories")
	subsAvailSetCmd.Flags().StringVar(&subsAvailPlanType, "plan-type", "MONTHLY", "subscription plan type: MONTHLY or UPFRONT")
	_ = subsAvailSetCmd.MarkFlagRequired("territories")

	// screenshot
	subsScreenshotCmd.Flags().StringVar(&subsFile, "file", "", "review screenshot file (required)")
	_ = subsScreenshotCmd.MarkFlagRequired("file")

	// grace period
	subsGraceSetCmd.Flags().BoolVar(&subsOptIn, "opt-in", false, "enable the grace period in production")
	subsGraceSetCmd.Flags().BoolVar(&subsSandboxOptIn, "sandbox-opt-in", false, "enable the grace period in sandbox")
	subsGraceSetCmd.Flags().StringVar(&subsDuration, "duration", "", "grace period duration: P3D, P16D or P28D")
	subsGraceSetCmd.Flags().StringVar(&subsRenewalType, "renewal-type", "", "ALL_RENEWALS or PAID_TO_PAID_ONLY")

	// win-back
	subsWinBackDeleteCmd.Flags().StringVar(&subsID, "id", "", "winBackOffer id (required)")
	_ = subsWinBackDeleteCmd.MarkFlagRequired("id")

	// promoted
	subsPromotedSetCmd.Flags().StringVar(&subsSub, "sub", "", "subscription id or productId to promote")
	subsPromotedSetCmd.Flags().StringVar(&subsIAP, "iap", "", "in-app purchase product id or ASC id to promote")
	subsPromotedSetCmd.Flags().BoolVar(&subsVisibleAll, "visible-for-all-users", true, "visible to all users (not only past purchasers)")
	subsPromotedSetCmd.Flags().BoolVar(&subsEnabled, "enabled", false, "enable or disable the promotion")

	subsGroupsCmd.AddCommand(subsGroupsListCmd, subsGroupsCreateCmd, subsGroupsLocalizeCmd, subsGroupsDeleteCmd)
	subsPriceCmd.AddCommand(subsPriceSetCmd, subsPriceListCmd)
	subsIntroCmd.AddCommand(subsIntroSetCmd, subsIntroListCmd, subsIntroDeleteCmd)
	subsPromoCmd.AddCommand(subsPromoCreateCmd, subsPromoListCmd, subsPromoDeleteCmd)
	subsCodesCmd.AddCommand(subsCodesCreateCmd, subsCodesListCmd, subsCodesDeactivateCmd, subsCodesCreateCodesCmd, subsCodesDownloadCmd)
	subsAvailCmd.AddCommand(subsAvailShowCmd, subsAvailSetCmd)
	subsGraceCmd.AddCommand(subsGraceShowCmd, subsGraceSetCmd)
	subsWinBackCmd.AddCommand(subsWinBackListCmd, subsWinBackDeleteCmd)
	subsPromotedCmd.AddCommand(subsPromotedListCmd, subsPromotedSetCmd)

	subsCmd.AddCommand(
		subsGroupsCmd,
		subsListCmd, subsCreateCmd, subsShowCmd, subsUpdateCmd, subsDeleteCmd, subsLocalizeCmd,
		subsPriceCmd, subsIntroCmd, subsPromoCmd, subsCodesCmd, subsAvailCmd,
		subsScreenshotCmd, subsSubmitCmd, subsGraceCmd, subsWinBackCmd, subsPromotedCmd,
	)
	rootCmd.AddCommand(subsCmd)
}
