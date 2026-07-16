package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var (
	fbkApp         string
	fbkID          string
	fbkTerritory   string
	fbkRating      int
	fbkSort        string
	fbkBody        string
	fbkBuild       string
	fbkPlatform    string
	fbkMetricType  string
	fbkDeviceType  string
	fbkDiagType    string
	fbkSignatureID string
	fbkOutput      string
)

// --- customer reviews ------------------------------------------------------------

var fbkReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Work with App Store customer reviews and responses",
}

var fbkReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List an app's customer reviews",
	Example: `  asc reviews list --app com.example.app
  asc reviews list --app 6790641087 --territory JPN --rating 1 --sort createdDate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, fbkApp)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("limit", "200")
		q.Set("sort", fbkSort)
		if fbkTerritory != "" {
			q.Set("filter[territory]", fbkTerritory)
		}
		if fbkRating > 0 {
			q.Set("filter[rating]", strconv.Itoa(fbkRating))
		}
		reviews, err := c.List(ctx, "/v1/apps/"+appID+"/customerReviews?"+q.Encode())
		if err != nil {
			return err
		}
		if len(reviews) == 0 {
			fmt.Println("No customer reviews found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "RATING\tTITLE\tTERRITORY\tDATE\tID")
		for i := range reviews {
			r := &reviews[i]
			date := r.Str("createdDate")
			if len(date) > 10 {
				date = date[:10]
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				fbkInt(r, "rating"), fbkTruncate(r.Str("title"), 40), r.Str("territory"), date, r.ID)
		}
		return w.Flush()
	},
}

var fbkReviewsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a customer review's full body and any developer response",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		review, _, err := c.Get(ctx, "/v1/customerReviews/"+fbkID)
		if err != nil {
			return err
		}
		out, err := json.MarshalIndent(review, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		resp, err := c.GetOptional(ctx, "/v1/customerReviews/"+fbkID+"/response")
		if err != nil {
			return err
		}
		if resp == nil || resp.ID == "" {
			fmt.Println("No developer response.")
			return nil
		}
		out, err = json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println("Response:")
		fmt.Println(string(out))
		return nil
	},
}

var fbkReviewsRespondCmd = &cobra.Command{
	Use:   "respond",
	Short: "Create or replace the developer response to a customer review",
	Example: `  asc reviews respond --id <review id> --body "Thank you for the feedback!"
  asc reviews respond --id <review id> --body @response.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := valueOrFile(fbkBody)
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		res, err := c.Post(cmd.Context(), "/v1/customerReviewResponses", api.Body{
			Data: api.Resource{
				Type:          "customerReviewResponses",
				Attributes:    map[string]any{"responseBody": body},
				Relationships: map[string]json.RawMessage{"review": api.Rel("customerReviews", fbkID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Response %s created for review %s.\n", res.ID, fbkID)
		return nil
	},
}

var fbkReviewsDeleteRespCmd = &cobra.Command{
	Use:   "delete-response",
	Short: "Delete the developer response to a customer review",
	Long: `Delete a customer review response. --id accepts either the review id (its
response is looked up) or the response id itself.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		respID := fbkID
		resp, err := c.GetOptional(ctx, "/v1/customerReviews/"+fbkID+"/response")
		if err != nil {
			return err
		}
		if resp != nil && resp.ID != "" {
			respID = resp.ID
		}
		if err := c.Delete(ctx, "/v1/customerReviewResponses/"+respID); err != nil {
			if api.IsNotFound(err) {
				return fmt.Errorf("no response found for %q (tried as review id and as response id)", fbkID)
			}
			return err
		}
		fmt.Printf("Deleted review response %s.\n", respID)
		return nil
	},
}

// --- performance metrics ------------------------------------------------------

var fbkMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Download Xcode performance and power metrics",
}

var fbkMetricsPowerCmd = &cobra.Command{
	Use:   "power",
	Short: "Download perfPowerMetrics for an app or a build",
	Long: `Download performance and power metrics (perfPowerMetrics) for an app or a
specific build. The response is Apple's xcode-metrics JSON (not JSON:API) and
is pretty-printed to stdout.`,
	Example: `  asc metrics power --app com.example.app
  asc metrics power --build <build id> --metric-type LAUNCH --device-type iPhone15,2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fbkApp != "" && fbkBuild != "" {
			return fmt.Errorf("pass --app or --build, not both")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		var path string
		switch {
		case fbkBuild != "":
			path = "/v1/builds/" + fbkBuild + "/perfPowerMetrics"
		case fbkApp != "":
			appID, err := resolveAppID(ctx, c, fbkApp)
			if err != nil {
				return err
			}
			path = "/v1/apps/" + appID + "/perfPowerMetrics"
		default:
			return fmt.Errorf("--app or --build is required")
		}
		q := url.Values{}
		if fbkPlatform != "" {
			q.Set("filter[platform]", fbkPlatform)
		}
		if fbkMetricType != "" {
			q.Set("filter[metricType]", fbkMetricType)
		}
		if fbkDeviceType != "" {
			q.Set("filter[deviceType]", fbkDeviceType)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Download(ctx, path, "application/vnd.apple.xcode-metrics+json")
		if err != nil {
			return err
		}
		return fbkEmitJSON(data, "")
	},
}

// --- diagnostics ---------------------------------------------------------------

var fbkDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Inspect build diagnostic signatures and logs",
}

var fbkDiagSignaturesCmd = &cobra.Command{
	Use:   "signatures",
	Short: "List a build's diagnostic signatures",
	Example: `  asc diagnostics signatures --build <build id>
  asc diagnostics signatures --build <build id> --type HANGS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("limit", "200")
		if fbkDiagType != "" {
			q.Set("filter[diagnosticType]", fbkDiagType)
		}
		sigs, err := c.List(cmd.Context(), "/v1/builds/"+fbkBuild+"/diagnosticSignatures?"+q.Encode())
		if err != nil {
			return err
		}
		if len(sigs) == 0 {
			fmt.Println("No diagnostic signatures found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tWEIGHT\tSIGNATURE\tID")
		for i := range sigs {
			s := &sigs[i]
			fmt.Fprintf(w, "%s\t%.4f\t%s\t%s\n",
				s.Str("diagnosticType"), fbkFloat(s, "weight"), fbkTruncate(s.Str("signature"), 60), s.ID)
		}
		return w.Flush()
	},
}

var fbkDiagLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Download the diagnostic logs for a diagnostic signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		data, err := c.Download(cmd.Context(), "/v1/diagnosticSignatures/"+fbkSignatureID+"/logs",
			"application/vnd.apple.diagnostic-logs+json")
		if err != nil {
			return err
		}
		return fbkEmitJSON(data, fbkOutput)
	},
}

// --- TestFlight beta feedback ----------------------------------------------------

var fbkFeedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Inspect TestFlight beta feedback (crashes and screenshots)",
}

var fbkFeedbackCrashesCmd = &cobra.Command{
	Use:   "crashes",
	Short: "List an app's beta feedback crash submissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fbkListSubmissions(cmd, "betaFeedbackCrashSubmissions")
	},
}

var fbkFeedbackScreenshotsCmd = &cobra.Command{
	Use:   "screenshots",
	Short: "List an app's beta feedback screenshot submissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fbkListSubmissions(cmd, "betaFeedbackScreenshotSubmissions")
	},
}

// fbkListSubmissions lists crash or screenshot beta feedback submissions.
func fbkListSubmissions(cmd *cobra.Command, kind string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	appID, err := resolveAppID(ctx, c, fbkApp)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("limit", "200")
	q.Set("sort", "-createdDate")
	if fbkBuild != "" {
		q.Set("filter[build]", fbkBuild)
	}
	subs, err := c.List(ctx, "/v1/apps/"+appID+"/"+kind+"?"+q.Encode())
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		fmt.Println("No submissions found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tDEVICE\tOS\tEMAIL\tCOMMENT\tID")
	for i := range subs {
		s := &subs[i]
		date := s.Str("createdDate")
		if len(date) > 10 {
			date = date[:10]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			date, s.Str("deviceModel"), s.Str("osVersion"), s.Str("email"),
			fbkTruncate(s.Str("comment"), 40), s.ID)
	}
	return w.Flush()
}

var fbkFeedbackCrashLogCmd = &cobra.Command{
	Use:   "crash-log",
	Short: "Download the crash log of a beta feedback crash submission",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		log, _, err := c.Get(cmd.Context(), "/v1/betaFeedbackCrashSubmissions/"+fbkID+"/crashLog")
		if err != nil {
			return err
		}
		text := log.Str("logText")
		if text == "" {
			return fmt.Errorf("crash log %s has no logText", log.ID)
		}
		if fbkOutput == "" {
			fmt.Println(text)
			return nil
		}
		if err := os.WriteFile(fbkOutput, []byte(text), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", fbkOutput, len(text))
		return nil
	},
}

var fbkFeedbackDeleteCrashCmd = &cobra.Command{
	Use:   "delete-crash",
	Short: "Delete a beta feedback crash submission",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/betaFeedbackCrashSubmissions/"+fbkID); err != nil {
			return err
		}
		fmt.Printf("Deleted crash submission %s.\n", fbkID)
		return nil
	},
}

var fbkFeedbackDeleteScreenshotCmd = &cobra.Command{
	Use:   "delete-screenshot",
	Short: "Delete a beta feedback screenshot submission",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/betaFeedbackScreenshotSubmissions/"+fbkID); err != nil {
			return err
		}
		fmt.Printf("Deleted screenshot submission %s.\n", fbkID)
		return nil
	},
}

// --- helpers ---------------------------------------------------------------------

// fbkEmitJSON pretty-prints JSON to stdout or writes it to a file.
func fbkEmitJSON(data []byte, output string) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err == nil {
		data = buf.Bytes()
	}
	if output == "" {
		fmt.Println(string(data))
		return nil
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", output, len(data))
	return nil
}

// fbkTruncate collapses whitespace and truncates to max runes.
func fbkTruncate(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// fbkInt returns a numeric attribute as int64 (JSON numbers decode as float64).
func fbkInt(r *api.Resource, key string) int64 {
	if v, ok := r.Attributes[key].(float64); ok {
		return int64(v)
	}
	return 0
}

// fbkFloat returns a numeric attribute as float64.
func fbkFloat(r *api.Resource, key string) float64 {
	if v, ok := r.Attributes[key].(float64); ok {
		return v
	}
	return 0
}

func init() {
	fbkReviewsListCmd.Flags().StringVar(&fbkApp, "app", "", "app id or bundle id (required)")
	_ = fbkReviewsListCmd.MarkFlagRequired("app")
	fbkReviewsListCmd.Flags().StringVar(&fbkTerritory, "territory", "", "filter by territory code, e.g. JPN, USA")
	fbkReviewsListCmd.Flags().IntVar(&fbkRating, "rating", 0, "filter by star rating (1-5)")
	fbkReviewsListCmd.Flags().StringVar(&fbkSort, "sort", "-createdDate", "sort: createdDate, -createdDate, rating, -rating")

	for _, sub := range []*cobra.Command{fbkReviewsShowCmd, fbkReviewsRespondCmd} {
		sub.Flags().StringVar(&fbkID, "id", "", "customer review id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	fbkReviewsRespondCmd.Flags().StringVar(&fbkBody, "body", "", "response body (@file allowed) (required)")
	_ = fbkReviewsRespondCmd.MarkFlagRequired("body")
	fbkReviewsDeleteRespCmd.Flags().StringVar(&fbkID, "id", "", "review id or response id (required)")
	_ = fbkReviewsDeleteRespCmd.MarkFlagRequired("id")

	fbkReviewsCmd.AddCommand(fbkReviewsListCmd, fbkReviewsShowCmd, fbkReviewsRespondCmd, fbkReviewsDeleteRespCmd)
	rootCmd.AddCommand(fbkReviewsCmd)

	fbkMetricsPowerCmd.Flags().StringVar(&fbkApp, "app", "", "app id or bundle id")
	fbkMetricsPowerCmd.Flags().StringVar(&fbkBuild, "build", "", "build id (instead of --app)")
	fbkMetricsPowerCmd.Flags().StringVar(&fbkPlatform, "platform", "", "filter by platform: IOS")
	fbkMetricsPowerCmd.Flags().StringVar(&fbkMetricType, "metric-type", "", "filter: DISK, HANG, BATTERY, LAUNCH, MEMORY, ANIMATION, TERMINATION, STORAGE")
	fbkMetricsPowerCmd.Flags().StringVar(&fbkDeviceType, "device-type", "", "filter by device type, e.g. iPhone15,2")

	fbkMetricsCmd.AddCommand(fbkMetricsPowerCmd)
	rootCmd.AddCommand(fbkMetricsCmd)

	fbkDiagSignaturesCmd.Flags().StringVar(&fbkBuild, "build", "", "build id (required)")
	_ = fbkDiagSignaturesCmd.MarkFlagRequired("build")
	fbkDiagSignaturesCmd.Flags().StringVar(&fbkDiagType, "type", "", "filter by diagnostic type: DISK_WRITES, HANGS, LAUNCHES")

	fbkDiagLogsCmd.Flags().StringVar(&fbkSignatureID, "signature-id", "", "diagnostic signature id (required)")
	_ = fbkDiagLogsCmd.MarkFlagRequired("signature-id")
	fbkDiagLogsCmd.Flags().StringVar(&fbkOutput, "output", "", "write to file instead of stdout")

	fbkDiagnosticsCmd.AddCommand(fbkDiagSignaturesCmd, fbkDiagLogsCmd)
	rootCmd.AddCommand(fbkDiagnosticsCmd)

	for _, sub := range []*cobra.Command{fbkFeedbackCrashesCmd, fbkFeedbackScreenshotsCmd} {
		sub.Flags().StringVar(&fbkApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
		sub.Flags().StringVar(&fbkBuild, "build", "", "filter by build id")
	}
	fbkFeedbackCrashLogCmd.Flags().StringVar(&fbkID, "id", "", "crash submission id (required)")
	_ = fbkFeedbackCrashLogCmd.MarkFlagRequired("id")
	fbkFeedbackCrashLogCmd.Flags().StringVar(&fbkOutput, "output", "", "write to file instead of stdout")
	fbkFeedbackDeleteCrashCmd.Flags().StringVar(&fbkID, "id", "", "crash submission id (required)")
	_ = fbkFeedbackDeleteCrashCmd.MarkFlagRequired("id")
	fbkFeedbackDeleteScreenshotCmd.Flags().StringVar(&fbkID, "id", "", "screenshot submission id (required)")
	_ = fbkFeedbackDeleteScreenshotCmd.MarkFlagRequired("id")

	fbkFeedbackCmd.AddCommand(fbkFeedbackCrashesCmd, fbkFeedbackScreenshotsCmd, fbkFeedbackCrashLogCmd,
		fbkFeedbackDeleteCrashCmd, fbkFeedbackDeleteScreenshotCmd)
	rootCmd.AddCommand(fbkFeedbackCmd)
}
