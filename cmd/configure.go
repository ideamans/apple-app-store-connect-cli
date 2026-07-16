package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/auth"
	"github.com/ideamans/apple-app-store-connect-cli/internal/config"
)

var (
	configureIssuerID string
	configureKeyPath  string
	configureKeyID    string
	configureForce    bool
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Add a profile from an issuer ID and a downloaded .p8 key",
	Long: `Copies the .p8 private key into ~/.config/apple-app-store-connect/keys/
(with 0600 permissions) and registers a profile in config.toml.

The key ID is derived from the AuthKey_XXXXXXXXXX.p8 filename unless --key-id
is given. The first registered profile becomes the default profile.

Use an App Store Connect API key WITH A ROLE (Users and Access > Integrations >
App Store Connect API; Admin or App Manager for write access). In-App Purchase
keys for the App Store Server API look similar but get 401 on this API.`,
	Example: `  asc configure --issuer-id 12345678-aaaa-bbbb-cccc-1234567890ab --key ~/Downloads/AuthKey_ABC123DEF4.p8
  asc configure --profile client-a --issuer-id ... --key ... --key-id ABC123DEF4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := profileFlag
		if name == "" {
			name = "default"
		}

		keyID := configureKeyID
		if keyID == "" {
			keyID = config.KeyIDFromFilename(configureKeyPath)
		}
		if keyID == "" {
			return fmt.Errorf("cannot derive key ID from filename %q; pass --key-id", filepath.Base(configureKeyPath))
		}

		pemData, err := os.ReadFile(configureKeyPath)
		if err != nil {
			return err
		}
		if _, err := auth.ParsePrivateKey(pemData); err != nil {
			return fmt.Errorf("%s does not look like a valid App Store Connect key: %w", configureKeyPath, err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, exists := cfg.Profiles[name]; exists && !configureForce {
			return fmt.Errorf("profile %q already exists; use --force to overwrite", name)
		}

		keysDir, err := config.KeysDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(keysDir, 0o700); err != nil {
			return err
		}
		destName := fmt.Sprintf("AuthKey_%s.p8", keyID)
		destPath := filepath.Join(keysDir, destName)
		if err := os.WriteFile(destPath, pemData, 0o600); err != nil {
			return err
		}

		cfg.Profiles[name] = config.Profile{
			IssuerID:   configureIssuerID,
			KeyID:      keyID,
			PrivateKey: filepath.Join("keys", destName),
		}
		madeDefault := false
		if cfg.DefaultProfile == "" {
			cfg.DefaultProfile = name
			madeDefault = true
		}
		if err := cfg.Save(); err != nil {
			return err
		}

		configPath, _ := config.FilePath()
		fmt.Printf("Profile %q saved to %s\n", name, configPath)
		fmt.Printf("  issuer_id:   %s\n", configureIssuerID)
		fmt.Printf("  key_id:      %s\n", keyID)
		fmt.Printf("  private_key: %s\n", destPath)
		if madeDefault {
			fmt.Printf("Set %q as the default profile.\n", name)
		}
		fmt.Println("\nVerify with: asc apps list")
		fmt.Printf("The original key file %s is no longer needed by this CLI.\n", configureKeyPath)
		return nil
	},
}

func init() {
	configureCmd.Flags().StringVar(&configureIssuerID, "issuer-id", "", "App Store Connect API issuer ID (required)")
	configureCmd.Flags().StringVar(&configureKeyPath, "key", "", "path to the downloaded AuthKey_XXXXXXXXXX.p8 file (required)")
	configureCmd.Flags().StringVar(&configureKeyID, "key-id", "", "key ID (default: derived from the .p8 filename)")
	configureCmd.Flags().BoolVar(&configureForce, "force", false, "overwrite an existing profile")
	_ = configureCmd.MarkFlagRequired("issuer-id")
	_ = configureCmd.MarkFlagRequired("key")
	rootCmd.AddCommand(configureCmd)
}
