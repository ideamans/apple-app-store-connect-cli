package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var profileFlag string

var rootCmd = &cobra.Command{
	Use:           "asc",
	Short:         "Apple App Store Connect API CLI",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "profile name (env: ASC_PROFILE)")
}
