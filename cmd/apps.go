package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Work with apps",
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all apps in the team",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Resolve(profileFlag)
		if err != nil {
			return err
		}
		client := api.New(creds)

		type app struct {
			ID         string `json:"id"`
			Attributes struct {
				Name     string `json:"name"`
				BundleID string `json:"bundleId"`
				SKU      string `json:"sku"`
			} `json:"attributes"`
		}
		var apps []app

		next := "/v1/apps?limit=200&sort=name"
		for next != "" {
			data, err := client.Do(cmd.Context(), http.MethodGet, next, nil)
			if err != nil {
				return err
			}
			var page struct {
				Data  []app `json:"data"`
				Links struct {
					Next string `json:"next"`
				} `json:"links"`
			}
			if err := json.Unmarshal(data, &page); err != nil {
				return err
			}
			apps = append(apps, page.Data...)
			next = page.Links.Next
		}

		if len(apps) == 0 {
			fmt.Println("No apps found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tBUNDLE ID\tSKU\tID")
		for _, a := range apps {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Attributes.Name, a.Attributes.BundleID, a.Attributes.SKU, a.ID)
		}
		return w.Flush()
	},
}

func init() {
	appsCmd.AddCommand(appsListCmd)
	rootCmd.AddCommand(appsCmd)
}
