package cmd

import (
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

// whkEventTypeEnum lists the WebhookEventType values accepted by the API.
const whkEventTypeEnum = `Event types:
  ALTERNATIVE_DISTRIBUTION_PACKAGE_AVAILABLE_UPDATED
  ALTERNATIVE_DISTRIBUTION_PACKAGE_VERSION_CREATED
  ALTERNATIVE_DISTRIBUTION_TERRITORY_AVAILABILITY_UPDATED
  APP_STORE_VERSION_APP_VERSION_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_APP_STORE_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_EXTERNAL_BETA_RELEASE_STATE_UPDATED
  BACKGROUND_ASSET_VERSION_INTERNAL_BETA_RELEASE_CREATED
  BACKGROUND_ASSET_VERSION_STATE_UPDATED
  BETA_FEEDBACK_CRASH_SUBMISSION_CREATED
  BETA_FEEDBACK_SCREENSHOT_SUBMISSION_CREATED
  BUILD_BETA_DETAIL_EXTERNAL_BUILD_STATE_UPDATED
  BUILD_UPLOAD_STATE_UPDATED`

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage App Store Connect webhooks and inspect their deliveries",
	Long:  "Create, update, ping, and delete webhooks for an app, and inspect delivery attempts.\n\n" + whkEventTypeEnum,
}

var (
	whkApp        string
	whkID         string
	whkName       string
	whkURL        string
	whkSecret     string
	whkEventTypes []string
	whkEnabled    bool
	whkState      string
	whkLimit      int
)

// whkBool formats a boolean attribute, or "" when absent.
func whkBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// whkEventTypesOf joins the eventTypes array attribute.
func whkEventTypesOf(r *api.Resource) string {
	raw, ok := r.Attributes["eventTypes"].([]any)
	if !ok {
		return ""
	}
	var out []string
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return strings.Join(out, ",")
}

// whkRelID extracts the id of a to-one relationship from the raw JSON.
func whkRelID(r *api.Resource, name string) string {
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

var whkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, whkApp)
		if err != nil {
			return err
		}
		hooks, err := c.List(ctx, "/v1/apps/"+appID+"/webhooks?limit=200")
		if err != nil {
			return err
		}
		if len(hooks) == 0 {
			fmt.Println("No webhooks found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tENABLED\tEVENT TYPES\tID")
		for i := range hooks {
			h := &hooks[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", h.Str("name"), h.Str("url"), whkBool(h, "enabled"), whkEventTypesOf(h), h.ID)
		}
		return w.Flush()
	},
}

var whkCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook for an app",
	Long:  "Create a webhook. The secret signs each delivery payload (X-Apple-CK-Signature).\n\n" + whkEventTypeEnum,
	Example: `  asc webhooks create --app com.example.app --name ci-hook \
    --url https://example.com/hook --secret @secret.txt \
    --event-types APP_STORE_VERSION_APP_VERSION_STATE_UPDATED,BUILD_UPLOAD_STATE_UPDATED`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, whkApp)
		if err != nil {
			return err
		}
		secret, err := valueOrFile(whkSecret)
		if err != nil {
			return err
		}
		created, err := c.Post(ctx, "/v1/webhooks", api.Body{
			Data: api.Resource{
				Type: "webhooks",
				Attributes: map[string]any{
					"name":       whkName,
					"url":        whkURL,
					"secret":     secret,
					"eventTypes": whkEventTypes,
					"enabled":    whkEnabled,
				},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created webhook %s.\n", created.ID)
		return nil
	},
}

var whkUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a webhook (only the flags you pass are changed)",
	Long:  "Update a webhook's name, URL, secret, enabled state, or event types.\n\n" + whkEventTypeEnum,
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = whkName
		}
		if cmd.Flags().Changed("url") {
			attrs["url"] = whkURL
		}
		if cmd.Flags().Changed("secret") {
			secret, err := valueOrFile(whkSecret)
			if err != nil {
				return err
			}
			attrs["secret"] = secret
		}
		if cmd.Flags().Changed("enabled") {
			attrs["enabled"] = whkEnabled
		}
		if cmd.Flags().Changed("event-types") {
			attrs["eventTypes"] = whkEventTypes
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass at least one of --name, --url, --secret, --enabled, --event-types")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/webhooks/"+whkID, api.Body{
			Data: api.Resource{
				Type:       "webhooks",
				ID:         whkID,
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

var whkDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/webhooks/"+whkID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var whkPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send a test ping event to a webhook",
	Long: `Send a ping event to the webhook's URL. The result of the attempt shows up as
a delivery; check it with "asc webhooks deliveries --id <webhook-id>".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/webhookPings", api.Body{
			Data: api.Resource{
				Type:          "webhookPings",
				Relationships: map[string]json.RawMessage{"webhook": api.Rel("webhooks", whkID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Ping %s sent. Check the result with: asc webhooks deliveries --id %s\n", created.ID, whkID)
		return nil
	},
}

var whkDeliveriesCmd = &cobra.Command{
	Use:   "deliveries",
	Short: "List delivery attempts of a webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/v1/webhooks/%s/deliveries?include=event&limit=%d", whkID, whkLimit)
		if whkState != "" {
			path += "&filter[deliveryState]=" + whkState
		}
		data, err := c.Do(cmd.Context(), http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		var doc struct {
			Data     []api.Resource `json:"data"`
			Included []api.Resource `json:"included"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			return err
		}
		if len(doc.Data) == 0 {
			fmt.Println("No deliveries found.")
			return nil
		}
		// Map included webhookEvents to their eventType.
		eventTypes := map[string]string{}
		for i := range doc.Included {
			inc := &doc.Included[i]
			if inc.Type == "webhookEvents" {
				eventTypes[inc.ID] = inc.Str("eventType")
			}
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "EVENT\tSTATE\tHTTP\tCREATED\tID")
		for i := range doc.Data {
			d := &doc.Data[i]
			httpCode := ""
			if resp, ok := d.Attributes["response"].(map[string]any); ok {
				if code, ok := resp["httpStatusCode"].(float64); ok {
					httpCode = strconv.FormatInt(int64(code), 10)
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				eventTypes[whkRelID(d, "event")], d.Str("deliveryState"), httpCode, d.Str("createdDate"), d.ID)
		}
		return w.Flush()
	},
}

func init() {
	whkListCmd.Flags().StringVar(&whkApp, "app", "", "app id or bundle id (required)")
	_ = whkListCmd.MarkFlagRequired("app")

	whkCreateCmd.Flags().StringVar(&whkApp, "app", "", "app id or bundle id (required)")
	whkCreateCmd.Flags().StringVar(&whkName, "name", "", "webhook name (required)")
	whkCreateCmd.Flags().StringVar(&whkURL, "url", "", "HTTPS endpoint URL (required)")
	whkCreateCmd.Flags().StringVar(&whkSecret, "secret", "", "signing secret, or @file to read it from a file (required)")
	whkCreateCmd.Flags().StringSliceVar(&whkEventTypes, "event-types", nil, "comma-separated event types (required); see command help for values")
	whkCreateCmd.Flags().BoolVar(&whkEnabled, "enabled", true, "create the webhook enabled")
	for _, f := range []string{"app", "name", "url", "secret", "event-types"} {
		_ = whkCreateCmd.MarkFlagRequired(f)
	}

	whkUpdateCmd.Flags().StringVar(&whkID, "id", "", "webhook id (required)")
	whkUpdateCmd.Flags().StringVar(&whkName, "name", "", "new name")
	whkUpdateCmd.Flags().StringVar(&whkURL, "url", "", "new endpoint URL")
	whkUpdateCmd.Flags().StringVar(&whkSecret, "secret", "", "new signing secret, or @file")
	whkUpdateCmd.Flags().StringSliceVar(&whkEventTypes, "event-types", nil, "new comma-separated event types")
	whkUpdateCmd.Flags().BoolVar(&whkEnabled, "enabled", true, "enable (true) or disable (false) the webhook")
	_ = whkUpdateCmd.MarkFlagRequired("id")

	for _, sub := range []*cobra.Command{whkDeleteCmd, whkPingCmd, whkDeliveriesCmd} {
		sub.Flags().StringVar(&whkID, "id", "", "webhook id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	whkDeliveriesCmd.Flags().StringVar(&whkState, "state", "", "filter by delivery state: SUCCEEDED, FAILED, or PENDING")
	whkDeliveriesCmd.Flags().IntVar(&whkLimit, "limit", 50, "maximum number of deliveries to list (max 200)")

	webhooksCmd.AddCommand(whkListCmd, whkCreateCmd, whkUpdateCmd, whkDeleteCmd, whkPingCmd, whkDeliveriesCmd)
	rootCmd.AddCommand(webhooksCmd)
}
