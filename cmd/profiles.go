package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage credential profiles",
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			fmt.Println(`No profiles configured. Run "asc configure --issuer-id <ID> --key <AuthKey_XXX.p8>".`)
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tKEY ID\tISSUER ID\tPRIVATE KEY")
		for _, name := range cfg.ProfileNames() {
			p := cfg.Profiles[name]
			marker := ""
			if name == cfg.DefaultProfile {
				marker = " (default)"
			}
			fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n", name, marker, p.KeyID, p.IssuerID, p.PrivateKey)
		}
		return w.Flush()
	},
}

var profilesUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		name := args[0]
		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("profile %q not found", name)
		}
		cfg.DefaultProfile = name
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Default profile set to %q.\n", name)
		return nil
	},
}

var profilesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a profile (the key file is kept)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		name := args[0]
		profile, ok := cfg.Profiles[name]
		if !ok {
			return fmt.Errorf("profile %q not found", name)
		}
		delete(cfg.Profiles, name)
		if cfg.DefaultProfile == name {
			cfg.DefaultProfile = ""
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Profile %q removed. The key file %s was kept; delete it manually if no other profile uses it.\n", name, profile.PrivateKey)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesListCmd, profilesUseCmd, profilesRemoveCmd)
	rootCmd.AddCommand(profilesCmd)
}
