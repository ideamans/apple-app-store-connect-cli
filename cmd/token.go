package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/auth"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

var tokenTTL time.Duration

var tokenCmd = &cobra.Command{
	Use:     "token",
	Short:   "Print a signed JWT for use with curl etc.",
	Example: `  curl -H "Authorization: Bearer $(asc token)" https://api.appstoreconnect.apple.com/v1/apps`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Resolve(profileFlag)
		if err != nil {
			return err
		}
		token, err := auth.Token(creds, tokenTTL)
		if err != nil {
			return err
		}
		fmt.Println(token)
		return nil
	},
}

func init() {
	tokenCmd.Flags().DurationVar(&tokenTTL, "ttl", auth.DefaultTTL, "token lifetime (max 20m)")
	rootCmd.AddCommand(tokenCmd)
}
