package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var xccCmd = &cobra.Command{
	Use:     "xcode-cloud",
	Aliases: []string{"ci"},
	Short:   "Work with Xcode Cloud products, workflows, and build runs",
	Long: `Inspect Xcode Cloud CI products and workflows, start build runs, and
retrieve run results (actions, artifacts, issues, test results).`,
}

var (
	xccID          string
	xccProductID   string
	xccWorkflowID  string
	xccBranch      string
	xccTag         string
	xccClean       bool
	xccLimit       int
	xccActionID    string
	xccDownloadDir string
	xccProviderID  string
)

// xccInt formats an integer attribute (JSON numbers arrive as float64).
func xccInt(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(float64); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	return ""
}

// xccBool formats a boolean attribute, or "" when absent.
func xccBool(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

// xccStatus summarizes a build run or action: the completionStatus when
// finished, otherwise the executionProgress (PENDING/RUNNING/COMPLETE).
func xccStatus(r *api.Resource) string {
	if s := r.Str("completionStatus"); s != "" {
		return s
	}
	return r.Str("executionProgress")
}

// xccListPage fetches a single page of resources. Unlike c.List it does not
// follow pagination, so ?limit= caps the result.
func xccListPage(ctx context.Context, c *api.Client, path string) ([]api.Resource, error) {
	data, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Data []api.Resource `json:"data"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc.Data, nil
}

// xccPrintJSON prints a resource as indented JSON.
func xccPrintJSON(r *api.Resource) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// --- products ------------------------------------------------------------

var xccProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "List, show, and delete Xcode Cloud products",
}

var xccProductsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Xcode Cloud products",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		products, err := c.List(cmd.Context(), "/v1/ciProducts?limit=200")
		if err != nil {
			return err
		}
		if len(products) == 0 {
			fmt.Println("No Xcode Cloud products found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tCREATED\tID")
		for i := range products {
			p := &products[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Str("name"), p.Str("productType"), p.Str("createdDate"), p.ID)
		}
		return w.Flush()
	},
}

var xccProductsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an Xcode Cloud product",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		product, _, err := c.Get(cmd.Context(), "/v1/ciProducts/"+xccID)
		if err != nil {
			return err
		}
		return xccPrintJSON(product)
	},
}

var xccProductsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an Xcode Cloud product (removes all its workflows and build data)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/ciProducts/"+xccID); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

// --- workflows -----------------------------------------------------------

var xccWorkflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "List, show, enable, and disable Xcode Cloud workflows",
}

var xccWorkflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows of a product",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		workflows, err := c.List(cmd.Context(), "/v1/ciProducts/"+xccProductID+"/workflows?limit=200")
		if err != nil {
			return err
		}
		if len(workflows) == 0 {
			fmt.Println("No workflows found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tENABLED\tID")
		for i := range workflows {
			wf := &workflows[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", wf.Str("name"), xccBool(wf, "isEnabled"), wf.ID)
		}
		return w.Flush()
	},
}

var xccWorkflowsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a workflow (start conditions, actions, environment)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		wf, _, err := c.Get(cmd.Context(), "/v1/ciWorkflows/"+xccID)
		if err != nil {
			return err
		}
		return xccPrintJSON(wf)
	},
}

// xccSetWorkflowEnabled PATCHes the workflow's isEnabled attribute.
func xccSetWorkflowEnabled(cmd *cobra.Command, enabled bool) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	_, err = c.Patch(cmd.Context(), "/v1/ciWorkflows/"+xccID, api.Body{
		Data: api.Resource{
			Type:       "ciWorkflows",
			ID:         xccID,
			Attributes: map[string]any{"isEnabled": enabled},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("Workflow %s isEnabled=%t.\n", xccID, enabled)
	return nil
}

var xccWorkflowsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return xccSetWorkflowEnabled(cmd, true)
	},
}

var xccWorkflowsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return xccSetWorkflowEnabled(cmd, false)
	},
}

// --- run -----------------------------------------------------------------

// xccResolveGitRef resolves a branch or tag name to an scmGitReference id via
// the workflow's repository. kind is BRANCH or TAG.
func xccResolveGitRef(ctx context.Context, c *api.Client, workflowID, name, kind string) (string, error) {
	repo, _, err := c.Get(ctx, "/v1/ciWorkflows/"+workflowID+"/repository")
	if err != nil {
		return "", err
	}
	if repo == nil || repo.ID == "" {
		return "", fmt.Errorf("workflow %s has no repository", workflowID)
	}
	refs, err := c.List(ctx, "/v1/scmRepositories/"+repo.ID+"/gitReferences?limit=200")
	if err != nil {
		return "", err
	}
	for i := range refs {
		if refs[i].Str("name") == name && refs[i].Str("kind") == kind {
			return refs[i].ID, nil
		}
	}
	return "", fmt.Errorf("no git reference of kind %s named %q in repository %s", kind, name, repo.ID)
}

var xccRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a new build run for a workflow",
	Long: `Start an Xcode Cloud build run. With --branch or --tag the name is resolved
to a git reference of the workflow's repository and used as the source;
otherwise Xcode Cloud builds the workflow's default reference.`,
	Example: `  asc xcode-cloud run --workflow-id WORKFLOW_ID --branch main`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if xccBranch != "" && xccTag != "" {
			return fmt.Errorf("--branch and --tag are mutually exclusive")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		rels := map[string]json.RawMessage{"workflow": api.Rel("ciWorkflows", xccWorkflowID)}
		if xccBranch != "" || xccTag != "" {
			name, kind := xccBranch, "BRANCH"
			if xccTag != "" {
				name, kind = xccTag, "TAG"
			}
			refID, err := xccResolveGitRef(ctx, c, xccWorkflowID, name, kind)
			if err != nil {
				return err
			}
			rels["sourceBranchOrTag"] = api.Rel("scmGitReferences", refID)
		}
		var attrs map[string]any
		if cmd.Flags().Changed("clean") {
			attrs = map[string]any{"clean": xccClean}
		}
		created, err := c.Post(ctx, "/v1/ciBuildRuns", api.Body{
			Data: api.Resource{
				Type:          "ciBuildRuns",
				Attributes:    attrs,
				Relationships: rels,
			},
		})
		if err != nil {
			return err
		}
		if number := xccInt(created, "number"); number != "" {
			fmt.Printf("Started build run %s (number %s).\n", created.ID, number)
		} else {
			fmt.Printf("Started build run %s.\n", created.ID)
		}
		return nil
	},
}

// --- runs ----------------------------------------------------------------

var xccRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "List and inspect build runs",
}

var xccRunsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List build runs of a workflow or product (newest first)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if (xccWorkflowID == "") == (xccProductID == "") {
			return fmt.Errorf("exactly one of --workflow-id or --product-id is required")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/ciWorkflows/" + xccWorkflowID + "/buildRuns"
		if xccProductID != "" {
			path = "/v1/ciProducts/" + xccProductID + "/buildRuns"
		}
		runs, err := xccListPage(cmd.Context(), c, fmt.Sprintf("%s?sort=-number&limit=%d", path, xccLimit))
		if err != nil {
			return err
		}
		if len(runs) == 0 {
			fmt.Println("No build runs found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NUMBER\tSTATUS\tSTARTED\tID")
		for i := range runs {
			r := &runs[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", xccInt(r, "number"), xccStatus(r), r.Str("startedDate"), r.ID)
		}
		return w.Flush()
	},
}

var xccRunsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a build run and its actions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		run, _, err := c.Get(ctx, "/v1/ciBuildRuns/"+xccID)
		if err != nil {
			return err
		}
		if err := xccPrintJSON(run); err != nil {
			return err
		}
		actions, err := c.List(ctx, "/v1/ciBuildRuns/"+xccID+"/actions?limit=200")
		if err != nil {
			return err
		}
		if len(actions) == 0 {
			return nil
		}
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ACTION\tTYPE\tSTATUS\tID")
		for i := range actions {
			a := &actions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Str("name"), a.Str("actionType"), xccStatus(a), a.ID)
		}
		return w.Flush()
	},
}

// --- artifacts / issues / tests -------------------------------------------

var xccArtifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "List (and optionally download) artifacts of a build action",
	Long: `List the artifacts produced by a build action (use "asc xcode-cloud runs show"
to find action ids). With --download-dir each artifact's download URL is
fetched and the file is saved into that directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		artifacts, err := c.List(ctx, "/v1/ciBuildActions/"+xccActionID+"/artifacts?limit=200")
		if err != nil {
			return err
		}
		if len(artifacts) == 0 {
			fmt.Println("No artifacts found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "FILE\tTYPE\tSIZE\tID")
		for i := range artifacts {
			a := &artifacts[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Str("fileName"), a.Str("fileType"), xccInt(a, "fileSize"), a.ID)
		}
		if err := w.Flush(); err != nil {
			return err
		}
		if xccDownloadDir == "" {
			return nil
		}
		if err := os.MkdirAll(xccDownloadDir, 0o755); err != nil {
			return err
		}
		for i := range artifacts {
			// downloadUrl is only populated on the individual artifact resource.
			full, _, err := c.Get(ctx, "/v1/ciArtifacts/"+artifacts[i].ID)
			if err != nil {
				return err
			}
			url := full.Str("downloadUrl")
			if url == "" {
				fmt.Fprintf(os.Stderr, "skipping %s: no download URL\n", full.Str("fileName"))
				continue
			}
			data, err := c.Download(ctx, url, "")
			if err != nil {
				return fmt.Errorf("%s: %w", full.Str("fileName"), err)
			}
			dest := filepath.Join(xccDownloadDir, full.Str("fileName"))
			if err := os.WriteFile(dest, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("downloaded %s (%d bytes)\n", dest, len(data))
		}
		return nil
	},
}

var xccIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "List issues (errors, warnings, test failures) of a build action",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		issues, err := c.List(cmd.Context(), "/v1/ciBuildActions/"+xccActionID+"/issues?limit=200")
		if err != nil {
			return err
		}
		if len(issues) == 0 {
			fmt.Println("No issues found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tCATEGORY\tMESSAGE\tID")
		for i := range issues {
			is := &issues[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", is.Str("issueType"), is.Str("category"), is.Str("message"), is.ID)
		}
		return w.Flush()
	},
}

var xccTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "List test results of a build action",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		results, err := c.List(cmd.Context(), "/v1/ciBuildActions/"+xccActionID+"/testResults?limit=200")
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("No test results found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STATUS\tCLASS\tNAME\tID")
		for i := range results {
			t := &results[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Str("status"), t.Str("className"), t.Str("name"), t.ID)
		}
		return w.Flush()
	},
}

// --- scm -----------------------------------------------------------------

var xccScmCmd = &cobra.Command{
	Use:   "scm",
	Short: "Browse Git providers and repositories connected to Xcode Cloud",
}

var xccScmProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Source control providers",
}

var xccScmProvidersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List source control providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		providers, err := c.List(cmd.Context(), "/v1/scmProviders?limit=200")
		if err != nil {
			return err
		}
		if len(providers) == 0 {
			fmt.Println("No SCM providers found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "KIND\tNAME\tURL\tID")
		for i := range providers {
			p := &providers[i]
			kind, name := "", ""
			if pt, ok := p.Attributes["scmProviderType"].(map[string]any); ok {
				kind, _ = pt["kind"].(string)
				name, _ = pt["displayName"].(string)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", kind, name, p.Str("url"), p.ID)
		}
		return w.Flush()
	},
}

var xccScmRepositoriesCmd = &cobra.Command{
	Use:   "repositories",
	Short: "Source control repositories",
}

var xccScmRepositoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories (all, or those of one provider)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/scmRepositories?limit=200"
		if xccProviderID != "" {
			path = "/v1/scmProviders/" + xccProviderID + "/repositories?limit=200"
		}
		repos, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			fmt.Println("No repositories found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "OWNER\tNAME\tCLONE URL\tID")
		for i := range repos {
			r := &repos[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Str("ownerName"), r.Str("repositoryName"), r.Str("httpCloneUrl"), r.ID)
		}
		return w.Flush()
	},
}

func init() {
	for _, sub := range []*cobra.Command{xccProductsShowCmd, xccProductsDeleteCmd} {
		sub.Flags().StringVar(&xccID, "id", "", "ciProduct id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	xccProductsCmd.AddCommand(xccProductsListCmd, xccProductsShowCmd, xccProductsDeleteCmd)

	xccWorkflowsListCmd.Flags().StringVar(&xccProductID, "product-id", "", "ciProduct id (required)")
	_ = xccWorkflowsListCmd.MarkFlagRequired("product-id")
	for _, sub := range []*cobra.Command{xccWorkflowsShowCmd, xccWorkflowsEnableCmd, xccWorkflowsDisableCmd} {
		sub.Flags().StringVar(&xccID, "id", "", "ciWorkflow id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	xccWorkflowsCmd.AddCommand(xccWorkflowsListCmd, xccWorkflowsShowCmd, xccWorkflowsEnableCmd, xccWorkflowsDisableCmd)

	xccRunCmd.Flags().StringVar(&xccWorkflowID, "workflow-id", "", "ciWorkflow id (required)")
	xccRunCmd.Flags().StringVar(&xccBranch, "branch", "", "branch name to build (default: the workflow's default reference)")
	xccRunCmd.Flags().StringVar(&xccTag, "tag", "", "tag name to build")
	xccRunCmd.Flags().BoolVar(&xccClean, "clean", false, "perform a clean build (no derived data cache)")
	_ = xccRunCmd.MarkFlagRequired("workflow-id")

	xccRunsListCmd.Flags().StringVar(&xccWorkflowID, "workflow-id", "", "ciWorkflow id")
	xccRunsListCmd.Flags().StringVar(&xccProductID, "product-id", "", "ciProduct id")
	xccRunsListCmd.Flags().IntVar(&xccLimit, "limit", 25, "maximum number of runs to list (max 200)")
	xccRunsShowCmd.Flags().StringVar(&xccID, "id", "", "ciBuildRun id (required)")
	_ = xccRunsShowCmd.MarkFlagRequired("id")
	xccRunsCmd.AddCommand(xccRunsListCmd, xccRunsShowCmd)

	xccArtifactsCmd.Flags().StringVar(&xccActionID, "action-id", "", "ciBuildAction id (required)")
	xccArtifactsCmd.Flags().StringVar(&xccDownloadDir, "download-dir", "", "directory to download artifact files into")
	_ = xccArtifactsCmd.MarkFlagRequired("action-id")
	for _, sub := range []*cobra.Command{xccIssuesCmd, xccTestsCmd} {
		sub.Flags().StringVar(&xccActionID, "action-id", "", "ciBuildAction id (required)")
		_ = sub.MarkFlagRequired("action-id")
	}

	xccScmProvidersCmd.AddCommand(xccScmProvidersListCmd)
	xccScmRepositoriesListCmd.Flags().StringVar(&xccProviderID, "provider-id", "", "scmProvider id (default: all repositories)")
	xccScmRepositoriesCmd.AddCommand(xccScmRepositoriesListCmd)
	xccScmCmd.AddCommand(xccScmProvidersCmd, xccScmRepositoriesCmd)

	xccCmd.AddCommand(xccProductsCmd, xccWorkflowsCmd, xccRunCmd, xccRunsCmd, xccArtifactsCmd, xccIssuesCmd, xccTestsCmd, xccScmCmd)
	rootCmd.AddCommand(xccCmd)
}
