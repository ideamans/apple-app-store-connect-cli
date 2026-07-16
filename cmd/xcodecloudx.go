package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	xczID           string
	xczRepositoryID string
	xczLimit        int
)

// --- environment -----------------------------------------------------------

var xczEnvCmd = &cobra.Command{
	Use:   "environment",
	Short: "Browse the build environments Xcode Cloud offers (macOS and Xcode versions)",
}

var xczEnvMacOSVersionsCmd = &cobra.Command{
	Use:   "macos-versions",
	Short: "List macOS versions available to Xcode Cloud builds",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		versions, err := c.List(cmd.Context(), "/v1/ciMacOsVersions?limit=200")
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No macOS versions found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tNAME\tID")
		for i := range versions {
			v := &versions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Str("version"), v.Str("name"), v.ID)
		}
		return w.Flush()
	},
}

var xczEnvXcodeVersionsCmd = &cobra.Command{
	Use:   "xcode-versions",
	Short: "List Xcode versions available to Xcode Cloud builds",
	Long: `List the Xcode versions Xcode Cloud can build with. With --id a single
Xcode version is shown as JSON, including its testDestinations (simulator
devices and runtimes available for testing).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if xczID != "" {
			version, _, err := c.Get(cmd.Context(), "/v1/ciXcodeVersions/"+xczID)
			if err != nil {
				return err
			}
			return xccPrintJSON(version)
		}
		versions, err := c.List(cmd.Context(), "/v1/ciXcodeVersions?limit=200")
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No Xcode versions found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tNAME\tID")
		for i := range versions {
			v := &versions[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Str("version"), v.Str("name"), v.ID)
		}
		return w.Flush()
	},
}

// --- scm pull-requests -----------------------------------------------------

var xczScmPullRequestsCmd = &cobra.Command{
	Use:   "pull-requests",
	Short: "List pull requests of a repository",
	Long: `List the pull requests Xcode Cloud knows about for a repository (find
repository ids with "asc xcode-cloud scm repositories list").`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/v1/scmRepositories/%s/pullRequests?limit=%d", xczRepositoryID, xczLimit)
		prs, err := xccListPage(cmd.Context(), c, path)
		if err != nil {
			return err
		}
		if len(prs) == 0 {
			fmt.Println("No pull requests found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NUMBER\tTITLE\tSOURCE\tDESTINATION\tCLOSED\tID")
		for i := range prs {
			pr := &prs[i]
			source := pr.Str("sourceBranchName")
			if v, ok := pr.Attributes["isCrossRepository"].(bool); ok && v {
				source = pr.Str("sourceRepositoryOwner") + "/" + pr.Str("sourceRepositoryName") + ":" + source
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				xccInt(pr, "number"), pr.Str("title"), source, pr.Str("destinationBranchName"), xccBool(pr, "isClosed"), pr.ID)
		}
		return w.Flush()
	},
}

func init() {
	xczEnvXcodeVersionsCmd.Flags().StringVar(&xczID, "id", "", "ciXcodeVersion id (show one version, incl. test destinations)")
	xczEnvCmd.AddCommand(xczEnvMacOSVersionsCmd, xczEnvXcodeVersionsCmd)

	xczScmPullRequestsCmd.Flags().StringVar(&xczRepositoryID, "repository-id", "", "scmRepository id (required)")
	xczScmPullRequestsCmd.Flags().IntVar(&xczLimit, "limit", 50, "maximum number of pull requests to list (max 200)")
	_ = xczScmPullRequestsCmd.MarkFlagRequired("repository-id")

	xccCmd.AddCommand(xczEnvCmd)
	xccScmCmd.AddCommand(xczScmPullRequestsCmd)
}
