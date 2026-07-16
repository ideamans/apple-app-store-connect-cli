package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	dstApp       string
	dstID        string
	dstPublicKey string
	dstDomain    string
	dstRefName   string
	dstVersionID string
	dstURL       string
	dstSecret    string
	dstIDs       []string
)

// dstBool formats a boolean attribute, or "" when absent.
func dstBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// dstPrintJSON prints a resource as indented JSON.
func dstPrintJSON(r *api.Resource) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// --- alt-distribution --------------------------------------------------------

var dstAltCmd = &cobra.Command{
	Use:   "alt-distribution",
	Short: "Manage EU alternative distribution keys, domains, and packages",
}

var dstKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Alternative distribution keys (marketplace/web distribution public keys)",
}

var dstKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alternative distribution keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		keys, err := c.List(cmd.Context(), "/v1/alternativeDistributionKeys?limit=200")
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Println("No alternative distribution keys found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPUBLIC KEY")
		for i := range keys {
			k := &keys[i]
			fmt.Fprintf(w, "%s\t%s\n", k.ID, strings.SplitN(k.Str("publicKey"), "\n", 2)[0])
		}
		return w.Flush()
	},
}

var dstKeysCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Register an alternative distribution public key",
	Example: `  asc alt-distribution keys create --public-key @public_key.pem --app com.example.app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		key, err := valueOrFile(dstPublicKey)
		if err != nil {
			return err
		}
		res := api.Resource{
			Type:       "alternativeDistributionKeys",
			Attributes: map[string]any{"publicKey": key},
		}
		if dstApp != "" {
			appID, err := resolveAppID(ctx, c, dstApp)
			if err != nil {
				return err
			}
			res.Relationships = map[string]json.RawMessage{"app": api.Rel("apps", appID)}
		}
		created, err := c.Post(ctx, "/v1/alternativeDistributionKeys", api.Body{Data: res})
		if err != nil {
			return err
		}
		fmt.Printf("Created alternative distribution key %s.\n", created.ID)
		return nil
	},
}

var dstKeysDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an alternative distribution key",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/alternativeDistributionKeys/"+dstID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var dstDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Alternative distribution domains (web distribution)",
}

var dstDomainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alternative distribution domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		domains, err := c.List(cmd.Context(), "/v1/alternativeDistributionDomains?limit=200")
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			fmt.Println("No alternative distribution domains found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tREFERENCE NAME\tCREATED\tID")
		for i := range domains {
			d := &domains[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Str("domain"), d.Str("referenceName"), d.Str("createdDate"), d.ID)
		}
		return w.Flush()
	},
}

var dstDomainsCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Register an alternative distribution domain",
	Example: `  asc alt-distribution domains create --domain example.com --reference-name production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/alternativeDistributionDomains", api.Body{
			Data: api.Resource{
				Type: "alternativeDistributionDomains",
				Attributes: map[string]any{
					"domain":        dstDomain,
					"referenceName": dstRefName,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created alternative distribution domain %s.\n", created.ID)
		return nil
	},
}

var dstDomainsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an alternative distribution domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/alternativeDistributionDomains/"+dstID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var dstPackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Alternative distribution packages of App Store versions",
}

var dstPackagesShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the alternative distribution package of an App Store version (and its versions)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		pkg, err := c.GetOptional(ctx, "/v1/appStoreVersions/"+dstVersionID+"/alternativeDistributionPackage")
		if err != nil {
			return err
		}
		if pkg == nil || pkg.ID == "" {
			fmt.Println("No alternative distribution package for this version.")
			return nil
		}
		if err := dstPrintJSON(pkg); err != nil {
			return err
		}
		versions, err := c.List(ctx, "/v1/alternativeDistributionPackages/"+pkg.ID+"/versions?limit=50")
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return nil
		}
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tSTATE\tURL EXPIRES\tID")
		for i := range versions {
			v := &versions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Str("version"), v.Str("state"), v.Str("urlExpirationDate"), v.ID)
		}
		return w.Flush()
	},
}

// --- marketplace ---------------------------------------------------------------

var dstMarketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Manage alternative app marketplace search details and webhooks",
}

var dstSearchDetailCmd = &cobra.Command{
	Use:   "search-detail",
	Short: "Marketplace search detail (catalog URL) of a marketplace app",
}

var dstSearchDetailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the marketplace search detail of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, dstApp)
		if err != nil {
			return err
		}
		detail, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/marketplaceSearchDetail")
		if err != nil {
			return err
		}
		if detail == nil || detail.ID == "" {
			fmt.Println("No marketplace search detail set.")
			return nil
		}
		return dstPrintJSON(detail)
	},
}

var dstSearchDetailSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the marketplace catalog URL of an app (creates or updates the search detail)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, dstApp)
		if err != nil {
			return err
		}
		existing, err := c.GetOptional(ctx, "/v1/apps/"+appID+"/marketplaceSearchDetail")
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != "" {
			_, err = c.Patch(ctx, "/v1/marketplaceSearchDetails/"+existing.ID, api.Body{
				Data: api.Resource{
					Type:       "marketplaceSearchDetails",
					ID:         existing.ID,
					Attributes: map[string]any{"catalogUrl": dstURL},
				},
			})
			if err != nil {
				return err
			}
			fmt.Println("Updated catalog URL.")
			return nil
		}
		created, err := c.Post(ctx, "/v1/marketplaceSearchDetails", api.Body{
			Data: api.Resource{
				Type:          "marketplaceSearchDetails",
				Attributes:    map[string]any{"catalogUrl": dstURL},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created marketplace search detail %s.\n", created.ID)
		return nil
	},
}

var dstMarketplaceWebhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Marketplace webhooks (deprecated by Apple in favor of the webhooks API)",
}

var dstMarketplaceWebhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List marketplace webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		hooks, err := c.List(cmd.Context(), "/v1/marketplaceWebhooks?limit=200")
		if err != nil {
			return err
		}
		if len(hooks) == 0 {
			fmt.Println("No marketplace webhooks found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ENDPOINT URL\tID")
		for i := range hooks {
			h := &hooks[i]
			fmt.Fprintf(w, "%s\t%s\n", h.Str("endpointUrl"), h.ID)
		}
		return w.Flush()
	},
}

var dstMarketplaceWebhookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a marketplace webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		secret, err := valueOrFile(dstSecret)
		if err != nil {
			return err
		}
		created, err := c.Post(cmd.Context(), "/v1/marketplaceWebhooks", api.Body{
			Data: api.Resource{
				Type: "marketplaceWebhooks",
				Attributes: map[string]any{
					"endpointUrl": dstURL,
					"secret":      secret,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created marketplace webhook %s.\n", created.ID)
		return nil
	},
}

var dstMarketplaceWebhookUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a marketplace webhook (only the flags you pass are changed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := map[string]any{}
		if cmd.Flags().Changed("url") {
			attrs["endpointUrl"] = dstURL
		}
		if cmd.Flags().Changed("secret") {
			secret, err := valueOrFile(dstSecret)
			if err != nil {
				return err
			}
			attrs["secret"] = secret
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update: pass --url and/or --secret")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		_, err = c.Patch(cmd.Context(), "/v1/marketplaceWebhooks/"+dstID, api.Body{
			Data: api.Resource{
				Type:       "marketplaceWebhooks",
				ID:         dstID,
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

var dstMarketplaceWebhookDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a marketplace webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/marketplaceWebhooks/"+dstID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// --- app-tags -------------------------------------------------------------------

var dstAppTagsCmd = &cobra.Command{
	Use:   "app-tags",
	Short: "App Store tags assigned to an app",
}

var dstAppTagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List App Store tags of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, dstApp)
		if err != nil {
			return err
		}
		tags, err := c.List(ctx, "/v1/apps/"+appID+"/appTags?limit=200&sort=name")
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			fmt.Println("No app tags found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVISIBLE IN APP STORE\tID")
		for i := range tags {
			t := &tags[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.Str("name"), dstBool(t, "visibleInAppStore"), t.ID)
		}
		return w.Flush()
	},
}

// --- actors ---------------------------------------------------------------------

var dstActorsCmd = &cobra.Command{
	Use:   "actors",
	Short: "Look up actors (users, API keys, Xcode Cloud, Apple) referenced by other resources",
	Long: `Actors identify who performed an action (e.g. in app store version release
history). The list endpoint requires explicit ids, so pass the actor ids you
found on other resources.`,
}

var dstActorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List actors by id (the API requires an id filter)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		actors, err := c.List(cmd.Context(), "/v1/actors?filter[id]="+strings.Join(dstIDs, ",")+"&limit=200")
		if err != nil {
			return err
		}
		if len(actors) == 0 {
			fmt.Println("No actors found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tNAME\tEMAIL\tAPI KEY\tID")
		for i := range actors {
			a := &actors[i]
			name := strings.TrimSpace(a.Str("userFirstName") + " " + a.Str("userLastName"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", a.Str("actorType"), name, a.Str("userEmail"), a.Str("apiKeyId"), a.ID)
		}
		return w.Flush()
	},
}

var dstActorsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an actor",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		actor, _, err := c.Get(cmd.Context(), "/v1/actors/"+dstID)
		if err != nil {
			return err
		}
		return dstPrintJSON(actor)
	},
}

func init() {
	dstKeysCreateCmd.Flags().StringVar(&dstPublicKey, "public-key", "", "PEM public key, or @file to read it from a file (required)")
	dstKeysCreateCmd.Flags().StringVar(&dstApp, "app", "", "app id or bundle id to associate the key with (optional)")
	_ = dstKeysCreateCmd.MarkFlagRequired("public-key")
	dstKeysDeleteCmd.Flags().StringVar(&dstID, "id", "", "alternativeDistributionKey id (required)")
	_ = dstKeysDeleteCmd.MarkFlagRequired("id")
	dstKeysCmd.AddCommand(dstKeysListCmd, dstKeysCreateCmd, dstKeysDeleteCmd)

	dstDomainsCreateCmd.Flags().StringVar(&dstDomain, "domain", "", "domain name, e.g. example.com (required)")
	dstDomainsCreateCmd.Flags().StringVar(&dstRefName, "reference-name", "", "reference name for the domain (required)")
	_ = dstDomainsCreateCmd.MarkFlagRequired("domain")
	_ = dstDomainsCreateCmd.MarkFlagRequired("reference-name")
	dstDomainsDeleteCmd.Flags().StringVar(&dstID, "id", "", "alternativeDistributionDomain id (required)")
	_ = dstDomainsDeleteCmd.MarkFlagRequired("id")
	dstDomainsCmd.AddCommand(dstDomainsListCmd, dstDomainsCreateCmd, dstDomainsDeleteCmd)

	dstPackagesShowCmd.Flags().StringVar(&dstVersionID, "version-id", "", "appStoreVersion id (required)")
	_ = dstPackagesShowCmd.MarkFlagRequired("version-id")
	dstPackagesCmd.AddCommand(dstPackagesShowCmd)

	dstAltCmd.AddCommand(dstKeysCmd, dstDomainsCmd, dstPackagesCmd)

	dstSearchDetailShowCmd.Flags().StringVar(&dstApp, "app", "", "app id or bundle id (required)")
	_ = dstSearchDetailShowCmd.MarkFlagRequired("app")
	dstSearchDetailSetCmd.Flags().StringVar(&dstApp, "app", "", "app id or bundle id (required)")
	dstSearchDetailSetCmd.Flags().StringVar(&dstURL, "url", "", "catalog URL (required)")
	_ = dstSearchDetailSetCmd.MarkFlagRequired("app")
	_ = dstSearchDetailSetCmd.MarkFlagRequired("url")
	dstSearchDetailCmd.AddCommand(dstSearchDetailShowCmd, dstSearchDetailSetCmd)

	dstMarketplaceWebhookCreateCmd.Flags().StringVar(&dstURL, "url", "", "endpoint URL (required)")
	dstMarketplaceWebhookCreateCmd.Flags().StringVar(&dstSecret, "secret", "", "signing secret, or @file (required)")
	_ = dstMarketplaceWebhookCreateCmd.MarkFlagRequired("url")
	_ = dstMarketplaceWebhookCreateCmd.MarkFlagRequired("secret")
	dstMarketplaceWebhookUpdateCmd.Flags().StringVar(&dstID, "id", "", "marketplaceWebhook id (required)")
	dstMarketplaceWebhookUpdateCmd.Flags().StringVar(&dstURL, "url", "", "new endpoint URL")
	dstMarketplaceWebhookUpdateCmd.Flags().StringVar(&dstSecret, "secret", "", "new signing secret, or @file")
	_ = dstMarketplaceWebhookUpdateCmd.MarkFlagRequired("id")
	dstMarketplaceWebhookDeleteCmd.Flags().StringVar(&dstID, "id", "", "marketplaceWebhook id (required)")
	_ = dstMarketplaceWebhookDeleteCmd.MarkFlagRequired("id")
	dstMarketplaceWebhookCmd.AddCommand(dstMarketplaceWebhookListCmd, dstMarketplaceWebhookCreateCmd,
		dstMarketplaceWebhookUpdateCmd, dstMarketplaceWebhookDeleteCmd)

	dstMarketplaceCmd.AddCommand(dstSearchDetailCmd, dstMarketplaceWebhookCmd)

	dstAppTagsListCmd.Flags().StringVar(&dstApp, "app", "", "app id or bundle id (required)")
	_ = dstAppTagsListCmd.MarkFlagRequired("app")
	dstAppTagsCmd.AddCommand(dstAppTagsListCmd)

	dstActorsListCmd.Flags().StringSliceVar(&dstIDs, "ids", nil, "comma-separated actor ids (required)")
	_ = dstActorsListCmd.MarkFlagRequired("ids")
	dstActorsShowCmd.Flags().StringVar(&dstID, "id", "", "actor id (required)")
	_ = dstActorsShowCmd.MarkFlagRequired("id")
	dstActorsCmd.AddCommand(dstActorsListCmd, dstActorsShowCmd)

	rootCmd.AddCommand(dstAltCmd, dstMarketplaceCmd, dstAppTagsCmd, dstActorsCmd)
}
