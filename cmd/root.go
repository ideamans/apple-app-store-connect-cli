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

// PluginVersion is the released version of this CLI. It is also the version
// recorded in plugins/apple-app-store-connect-cli/.claude-plugin/plugin.json —
// a test enforces that the two agree, and the release workflow enforces that
// both agree with the git tag. Bump it in the same commit as the tag.
const PluginVersion = "0.5.0"

// Root returns the assembled command tree without executing it, so the catalog
// generator works from the same definition main dispatches on.
func Root() *cobra.Command { return rootCmd }

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "profile name (env: ASC_PROFILE)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print mutating requests instead of sending them")
}
