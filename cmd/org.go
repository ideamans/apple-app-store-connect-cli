package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// orgRolesHelp lists the UserRole enum values from the App Store Connect API spec.
const orgRolesHelp = "comma-separated roles: ADMIN, FINANCE, ACCOUNT_HOLDER, SALES, MARKETING, APP_MANAGER, DEVELOPER, ACCESS_TO_REPORTS, CUSTOMER_SUPPORT, CREATE_APPS, CLOUD_MANAGED_DEVELOPER_ID, CLOUD_MANAGED_APP_DISTRIBUTION, GENERATE_INDIVIDUAL_KEYS"

var orgValidRoles = map[string]bool{
	"ADMIN":                          true,
	"FINANCE":                        true,
	"ACCOUNT_HOLDER":                 true,
	"SALES":                          true,
	"MARKETING":                      true,
	"APP_MANAGER":                    true,
	"DEVELOPER":                      true,
	"ACCESS_TO_REPORTS":              true,
	"CUSTOMER_SUPPORT":               true,
	"CREATE_APPS":                    true,
	"CLOUD_MANAGED_DEVELOPER_ID":     true,
	"CLOUD_MANAGED_APP_DISTRIBUTION": true,
	"GENERATE_INDIVIDUAL_KEYS":       true,
}

var orgUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users on the team",
}

var orgInvitationsCmd = &cobra.Command{
	Use:   "invitations",
	Short: "Manage user invitations",
}

var (
	orgUserID      string
	orgUserEmail   string
	orgUserRoles   string
	orgUserAllApps bool
	orgUserProv    bool
	orgUserAppIDs  string

	orgInvID      string
	orgInvEmail   string
	orgInvFirst   string
	orgInvLast    string
	orgInvRoles   string
	orgInvAllApps bool
	orgInvApps    []string
)

// orgParseRoles splits and validates a comma-separated roles list.
func orgParseRoles(s string) ([]string, error) {
	var roles []string
	for _, p := range strings.Split(s, ",") {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if !orgValidRoles[p] {
			return nil, fmt.Errorf("invalid role %q (%s)", p, orgRolesHelp)
		}
		roles = append(roles, p)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles given (%s)", orgRolesHelp)
	}
	return roles, nil
}

// orgJoinedRoles returns the resource's roles attribute joined by commas.
func orgJoinedRoles(r *api.Resource) string {
	var roles []string
	if err := r.DecodeAttr("roles", &roles); err != nil {
		return ""
	}
	return strings.Join(roles, ",")
}

// orgFindUser resolves a user by --id or by --email (filter[username]).
func orgFindUser(ctx context.Context, c *api.Client, id, email string) (*api.Resource, error) {
	if id != "" {
		u, _, err := c.Get(ctx, "/v1/users/"+id)
		return u, err
	}
	if email == "" {
		return nil, fmt.Errorf("--id or --email is required")
	}
	list, err := c.List(ctx, "/v1/users?filter[username]="+url.QueryEscape(email)+"&limit=200")
	if err != nil {
		return nil, err
	}
	if u := findByAttr(list, "username", email); u != nil {
		return u, nil
	}
	if len(list) == 1 {
		return &list[0], nil
	}
	return nil, fmt.Errorf("no user found with email %q", email)
}

// orgFindInvitation resolves an invitation by --id or by --email (filter[email]).
func orgFindInvitation(ctx context.Context, c *api.Client, id, email string) (*api.Resource, error) {
	if id != "" {
		inv, _, err := c.Get(ctx, "/v1/userInvitations/"+id)
		return inv, err
	}
	if email == "" {
		return nil, fmt.Errorf("--id or --email is required")
	}
	list, err := c.List(ctx, "/v1/userInvitations?filter[email]="+url.QueryEscape(email)+"&limit=200")
	if err != nil {
		return nil, err
	}
	if inv := findByAttr(list, "email", email); inv != nil {
		return inv, nil
	}
	if len(list) == 1 {
		return &list[0], nil
	}
	return nil, fmt.Errorf("no invitation found for email %q", email)
}

var orgUsersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users on the team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		users, err := c.List(cmd.Context(), "/v1/users?limit=200")
		if err != nil {
			return err
		}
		if len(users) == 0 {
			fmt.Println("No users found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "USERNAME\tNAME\tROLES\tID")
		for i := range users {
			u := &users[i]
			name := strings.TrimSpace(u.Str("firstName") + " " + u.Str("lastName"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", u.Str("username"), name, orgJoinedRoles(u), u.ID)
		}
		return w.Flush()
	},
}

var orgUsersShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a user by id or email",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		u, err := orgFindUser(cmd.Context(), c, orgUserID, orgUserEmail)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(u, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var orgUsersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user's roles and permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		u, err := orgFindUser(ctx, c, orgUserID, orgUserEmail)
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("roles") {
			roles, err := orgParseRoles(orgUserRoles)
			if err != nil {
				return err
			}
			attrs["roles"] = roles
		}
		if cmd.Flags().Changed("all-apps-visible") {
			attrs["allAppsVisible"] = orgUserAllApps
		}
		if cmd.Flags().Changed("provisioning-allowed") {
			attrs["provisioningAllowed"] = orgUserProv
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update (use --roles, --all-apps-visible or --provisioning-allowed)")
		}
		if _, err := c.Patch(ctx, "/v1/users/"+u.ID, api.Body{
			Data: api.Resource{Type: "users", ID: u.ID, Attributes: attrs},
		}); err != nil {
			return err
		}
		fmt.Printf("User %s updated.\n", u.Str("username"))
		return nil
	},
}

var orgUsersRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a user from the team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		u, err := orgFindUser(ctx, c, orgUserID, orgUserEmail)
		if err != nil {
			return err
		}
		if err := c.Delete(ctx, "/v1/users/"+u.ID); err != nil {
			return err
		}
		fmt.Printf("User %s removed.\n", u.Str("username"))
		return nil
	},
}

var orgUsersAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List the apps visible to a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		u, err := orgFindUser(ctx, c, orgUserID, orgUserEmail)
		if err != nil {
			return err
		}
		apps, err := c.List(ctx, "/v1/users/"+u.ID+"/visibleApps?limit=200")
		if err != nil {
			return err
		}
		if len(apps) == 0 {
			fmt.Println("No visible apps (the user may have access to all apps).")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tBUNDLE ID\tID")
		for i := range apps {
			a := &apps[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.Str("name"), a.Str("bundleId"), a.ID)
		}
		return w.Flush()
	},
}

var orgUsersSetAppsCmd = &cobra.Command{
	Use:   "set-apps",
	Short: "Replace the set of apps visible to a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		u, err := orgFindUser(ctx, c, orgUserID, orgUserEmail)
		if err != nil {
			return err
		}
		var linkages []map[string]string
		for _, id := range strings.Split(orgUserAppIDs, ",") {
			if id = strings.TrimSpace(id); id != "" {
				linkages = append(linkages, map[string]string{"type": "apps", "id": id})
			}
		}
		if len(linkages) == 0 {
			return fmt.Errorf("--app-ids must list at least one app id")
		}
		payload, err := json.Marshal(map[string]any{"data": linkages})
		if err != nil {
			return err
		}
		path := "/v1/users/" + u.ID + "/relationships/visibleApps"
		if dryRun {
			fmt.Fprintf(os.Stderr, "DRY-RUN PATCH %s\n%s\n", path, payload)
			return nil
		}
		if _, err := c.Do(ctx, http.MethodPatch, path, bytes.NewReader(payload)); err != nil {
			return err
		}
		fmt.Printf("Visible apps for user %s set (%d apps).\n", u.Str("username"), len(linkages))
		return nil
	},
}

var orgInvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending user invitations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		invs, err := c.List(cmd.Context(), "/v1/userInvitations?limit=200")
		if err != nil {
			return err
		}
		if len(invs) == 0 {
			fmt.Println("No pending invitations.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "EMAIL\tNAME\tROLES\tEXPIRES\tID")
		for i := range invs {
			inv := &invs[i]
			name := strings.TrimSpace(inv.Str("firstName") + " " + inv.Str("lastName"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", inv.Str("email"), name, orgJoinedRoles(inv), inv.Str("expirationDate"), inv.ID)
		}
		return w.Flush()
	},
}

var orgInvCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Invite a new user to the team",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		roles, err := orgParseRoles(orgInvRoles)
		if err != nil {
			return err
		}
		attrs := map[string]any{
			"email":     orgInvEmail,
			"firstName": orgInvFirst,
			"lastName":  orgInvLast,
			"roles":     roles,
		}
		if cmd.Flags().Changed("all-apps-visible") {
			attrs["allAppsVisible"] = orgInvAllApps
		}
		var rels map[string]json.RawMessage
		if len(orgInvApps) > 0 {
			items := make([]map[string]string, 0, len(orgInvApps))
			for _, ref := range orgInvApps {
				appID, err := resolveAppID(cmd.Context(), c, strings.TrimSpace(ref))
				if err != nil {
					return err
				}
				items = append(items, map[string]string{"type": "apps", "id": appID})
			}
			linkage, err := json.Marshal(map[string]any{"data": items})
			if err != nil {
				return err
			}
			rels = map[string]json.RawMessage{"visibleApps": linkage}
		}
		inv, err := c.Post(cmd.Context(), "/v1/userInvitations", api.Body{
			Data: api.Resource{Type: "userInvitations", Attributes: attrs, Relationships: rels},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Invitation sent to %s (id %s).\n", orgInvEmail, inv.ID)
		return nil
	},
}

var orgInvCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a pending user invitation",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		inv, err := orgFindInvitation(ctx, c, orgInvID, orgInvEmail)
		if err != nil {
			return err
		}
		if err := c.Delete(ctx, "/v1/userInvitations/"+inv.ID); err != nil {
			return err
		}
		fmt.Printf("Invitation %s canceled.\n", inv.Str("email"))
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{orgUsersShowCmd, orgUsersUpdateCmd, orgUsersRemoveCmd, orgUsersAppsCmd, orgUsersSetAppsCmd} {
		sub.Flags().StringVar(&orgUserID, "id", "", "user id")
		sub.Flags().StringVar(&orgUserEmail, "email", "", "user email (Apple ID)")
	}
	orgUsersUpdateCmd.Flags().StringVar(&orgUserRoles, "roles", "", orgRolesHelp)
	orgUsersUpdateCmd.Flags().BoolVar(&orgUserAllApps, "all-apps-visible", false, "grant access to all apps")
	orgUsersUpdateCmd.Flags().BoolVar(&orgUserProv, "provisioning-allowed", false, "allow access to Certificates, Identifiers & Profiles")
	orgUsersSetAppsCmd.Flags().StringVar(&orgUserAppIDs, "app-ids", "", "comma-separated app ids to make visible (required)")
	_ = orgUsersSetAppsCmd.MarkFlagRequired("app-ids")
	orgUsersCmd.AddCommand(orgUsersListCmd, orgUsersShowCmd, orgUsersUpdateCmd, orgUsersRemoveCmd, orgUsersAppsCmd, orgUsersSetAppsCmd)

	orgInvCreateCmd.Flags().StringVar(&orgInvEmail, "email", "", "invitee email (required)")
	orgInvCreateCmd.Flags().StringVar(&orgInvFirst, "first-name", "", "invitee first name (required)")
	orgInvCreateCmd.Flags().StringVar(&orgInvLast, "last-name", "", "invitee last name (required)")
	orgInvCreateCmd.Flags().StringVar(&orgInvRoles, "roles", "", orgRolesHelp+" (required)")
	orgInvCreateCmd.Flags().BoolVar(&orgInvAllApps, "all-apps-visible", false, "grant access to all apps")
	orgInvCreateCmd.Flags().StringSliceVar(&orgInvApps, "app-ids", nil, "apps to make visible (app ids or bundle ids); required for app-scoped roles without --all-apps-visible")
	_ = orgInvCreateCmd.MarkFlagRequired("email")
	_ = orgInvCreateCmd.MarkFlagRequired("first-name")
	_ = orgInvCreateCmd.MarkFlagRequired("last-name")
	_ = orgInvCreateCmd.MarkFlagRequired("roles")
	orgInvCancelCmd.Flags().StringVar(&orgInvID, "id", "", "invitation id")
	orgInvCancelCmd.Flags().StringVar(&orgInvEmail, "email", "", "invitee email")
	orgInvitationsCmd.AddCommand(orgInvListCmd, orgInvCreateCmd, orgInvCancelCmd)

	rootCmd.AddCommand(orgUsersCmd, orgInvitationsCmd)
}
