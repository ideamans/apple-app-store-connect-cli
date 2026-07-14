package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

var (
	apiMethod string
	apiData   string
)

var apiCmd = &cobra.Command{
	Use:   "api <path>",
	Short: "Send a raw request to the App Store Connect API",
	Long: `Sends an authenticated request to https://api.appstoreconnect.apple.com
and prints the JSON response. Use this for endpoints not yet covered by a
dedicated subcommand.`,
	Example: `  asc api /v1/apps
  asc api "/v1/apps?filter[bundleId]=com.example.app"
  asc api -X POST /v1/betaGroups -d '{"data": {...}}'
  asc api -X POST /v1/betaGroups -d @payload.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		creds, err := config.Resolve(profileFlag)
		if err != nil {
			return err
		}

		var body io.Reader
		if apiData != "" {
			payload := []byte(apiData)
			if strings.HasPrefix(apiData, "@") {
				payload, err = os.ReadFile(apiData[1:])
				if err != nil {
					return err
				}
			}
			body = bytes.NewReader(payload)
		}

		data, err := api.New(creds).Do(cmd.Context(), strings.ToUpper(apiMethod), path, body)
		if err != nil {
			return err
		}
		var pretty bytes.Buffer
		if json.Indent(&pretty, data, "", "  ") == nil {
			fmt.Println(pretty.String())
		} else if len(data) > 0 {
			fmt.Println(string(data))
		}
		return nil
	},
}

func init() {
	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", "GET", "HTTP method")
	apiCmd.Flags().StringVarP(&apiData, "data", "d", "", "JSON request body, or @file to read from a file")
	rootCmd.AddCommand(apiCmd)
}
