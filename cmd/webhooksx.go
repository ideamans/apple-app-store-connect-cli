package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	whxWebhookID  string
	whxDeliveryID string
)

var whxDeliveryCmd = &cobra.Command{
	Use:   "delivery",
	Short: "Show one delivery attempt of a webhook (full request/response details)",
	Long: `Show a single webhook delivery as JSON, including the request URL, the
response HTTP status code and body, the delivery state, and any error message.

The API has no per-delivery endpoint, so the delivery is located by scanning
the webhook's deliveries; both --webhook-id and --id are required.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		deliveries, err := c.List(cmd.Context(), "/v1/webhooks/"+whxWebhookID+"/deliveries?limit=200")
		if err != nil {
			return err
		}
		for i := range deliveries {
			if deliveries[i].ID == whxDeliveryID {
				b, err := json.MarshalIndent(&deliveries[i], "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
		}
		return fmt.Errorf("webhook %s has no delivery %s", whxWebhookID, whxDeliveryID)
	},
}

var whxRedeliverCmd = &cobra.Command{
	Use:   "redeliver",
	Short: "Redeliver a webhook event using a past delivery as the template",
	Long: `Ask App Store Connect to send the event of a previous delivery again. The
original delivery is referenced as the "template" of the new one (find
delivery ids with "asc webhooks deliveries --id <webhook-id>").`,
	Example: `  asc webhooks redeliver --delivery-id DELIVERY_ID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/webhookDeliveries", api.Body{
			Data: api.Resource{
				Type: "webhookDeliveries",
				Relationships: map[string]json.RawMessage{
					"template": api.Rel("webhookDeliveries", whxDeliveryID),
				},
			},
		})
		if err != nil {
			return err
		}
		if state := created.Str("deliveryState"); state != "" {
			fmt.Printf("Created redelivery %s (state %s).\n", created.ID, state)
		} else {
			fmt.Printf("Created redelivery %s.\n", created.ID)
		}
		return nil
	},
}

func init() {
	whxDeliveryCmd.Flags().StringVar(&whxWebhookID, "webhook-id", "", "webhook id the delivery belongs to (required)")
	whxDeliveryCmd.Flags().StringVar(&whxDeliveryID, "id", "", "webhookDelivery id (required)")
	_ = whxDeliveryCmd.MarkFlagRequired("webhook-id")
	_ = whxDeliveryCmd.MarkFlagRequired("id")

	whxRedeliverCmd.Flags().StringVar(&whxDeliveryID, "delivery-id", "", "webhookDelivery id to redeliver (required)")
	_ = whxRedeliverCmd.MarkFlagRequired("delivery-id")

	webhooksCmd.AddCommand(whxDeliveryCmd, whxRedeliverCmd)
}
