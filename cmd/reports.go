package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

var rptReportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Download sales and finance reports",
}

var (
	rptVendor    string
	rptType      string
	rptSubtype   string
	rptFrequency string
	rptDate      string
	rptVersion   string
	rptFinType   string
	rptRegion    string
	rptOutput    string
	rptGzip      bool
)

var rptSalesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Download a sales and trends report (TSV)",
	Long: `Download a sales report from GET /v1/salesReports.

The report arrives gzip-compressed. By default it is decompressed and the
tab-separated content is written to stdout (or --output); pass --gzip to keep
the raw .gz bytes instead.

The vendor number is shown in App Store Connect under Payments and Financial
Reports (an 8-digit number starting with 8).`,
	Example: `  asc reports sales --vendor 8XXXXXXX --date 2026-07-15
  asc reports sales --vendor 8XXXXXXX --type SUBSCRIPTION --subtype SUMMARY --frequency DAILY --version 1_3 --output subs.tsv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		q := url.Values{}
		q.Set("filter[vendorNumber]", rptVendor)
		q.Set("filter[reportType]", rptType)
		q.Set("filter[reportSubType]", rptSubtype)
		q.Set("filter[frequency]", rptFrequency)
		if rptDate != "" {
			q.Set("filter[reportDate]", rptDate)
		}
		if rptVersion != "" {
			q.Set("filter[version]", rptVersion)
		}
		return rptFetchReport(cmd, "/v1/salesReports", q)
	},
}

var rptFinanceCmd = &cobra.Command{
	Use:   "finance",
	Short: "Download a finance report (TSV)",
	Long: `Download a finance report from GET /v1/financeReports.

--region is a two-character financial region code (e.g. JP, US, or ZZ for the
consolidated report; Z1 with --type FINANCE_DETAIL). --date is the fiscal
period, e.g. 2026-06. The report arrives gzip-compressed and is decompressed
by default; pass --gzip to keep the raw .gz bytes.

The vendor number is shown in App Store Connect under Payments and Financial
Reports.`,
	Example: `  asc reports finance --vendor 8XXXXXXX --region ZZ --date 2026-06
  asc reports finance --vendor 8XXXXXXX --region Z1 --type FINANCE_DETAIL --date 2026-06 --output detail.tsv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		q := url.Values{}
		q.Set("filter[vendorNumber]", rptVendor)
		q.Set("filter[reportType]", rptFinType)
		q.Set("filter[regionCode]", rptRegion)
		q.Set("filter[reportDate]", rptDate)
		return rptFetchReport(cmd, "/v1/financeReports", q)
	},
}

// rptFetchReport downloads a gzip report and emits it per --gzip/--output.
func rptFetchReport(cmd *cobra.Command, path string, q url.Values) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	data, err := c.Download(cmd.Context(), path+"?"+q.Encode(), "application/a-gzip")
	if err != nil {
		return err
	}
	if rptGzip {
		return rptEmit(data, rptOutput)
	}
	tsv, err := rptGunzip(data)
	if err != nil {
		return err
	}
	return rptEmit(tsv, rptOutput)
}

// rptGunzip decompresses gzip data; content that is not gzipped passes through.
func rptGunzip(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return data, nil
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func rptEmit(data []byte, output string) error {
	if output == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", output, len(data))
	return nil
}

// --- analytics -----------------------------------------------------------------

var rptAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Request and download App Store analytics reports",
	Long: `Work with the Analytics Reports API. The flow is:

  1. analytics request    -- create a report request for an app
  2. analytics requests   -- find the request id
  3. analytics reports    -- list reports available for a request
  4. analytics instances  -- list a report's instances (per processing date)
  5. analytics segments   -- list or download an instance's data segments (.csv.gz)`,
}

var (
	rptApp         string
	rptAccessType  string
	rptRequestID   string
	rptCategory    string
	rptReportID    string
	rptGranularity string
	rptInstanceID  string
	rptOutputDir   string
)

var rptAnalyticsRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Create an analytics report request for an app",
	Long: `Create an analytics report request (POST /v1/analyticsReportRequests).

--access-type ONGOING generates reports continuously; ONE_TIME_SNAPSHOT
generates historical data once. Reports become available asynchronously.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rptAccessType != "ONGOING" && rptAccessType != "ONE_TIME_SNAPSHOT" {
			return fmt.Errorf("--access-type must be ONGOING or ONE_TIME_SNAPSHOT, got %q", rptAccessType)
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, rptApp)
		if err != nil {
			return err
		}
		res, err := c.Post(ctx, "/v1/analyticsReportRequests", api.Body{
			Data: api.Resource{
				Type:          "analyticsReportRequests",
				Attributes:    map[string]any{"accessType": rptAccessType},
				Relationships: map[string]json.RawMessage{"app": api.Rel("apps", appID)},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created analytics report request %s (accessType=%s).\n", res.ID, rptAccessType)
		return nil
	},
}

var rptAnalyticsRequestsCmd = &cobra.Command{
	Use:   "requests",
	Short: "List an app's analytics report requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		appID, err := resolveAppID(ctx, c, rptApp)
		if err != nil {
			return err
		}
		reqs, err := c.List(ctx, "/v1/apps/"+appID+"/analyticsReportRequests?limit=200")
		if err != nil {
			return err
		}
		if len(reqs) == 0 {
			fmt.Println("No analytics report requests found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ACCESS TYPE\tSTOPPED\tID")
		for i := range reqs {
			r := &reqs[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Str("accessType"), rptBoolStr(r, "stoppedDueToInactivity"), r.ID)
		}
		return w.Flush()
	},
}

var rptAnalyticsReportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "List reports available for an analytics report request",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/analyticsReportRequests/" + rptRequestID + "/reports?limit=200"
		if rptCategory != "" {
			path += "&filter[category]=" + url.QueryEscape(rptCategory)
		}
		reports, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(reports) == 0 {
			fmt.Println("No reports found (report generation is asynchronous; try again later).")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tID")
		for i := range reports {
			r := &reports[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Str("name"), r.Str("category"), r.ID)
		}
		return w.Flush()
	},
}

var rptAnalyticsInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List instances of an analytics report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/analyticsReports/" + rptReportID + "/instances?limit=200"
		if rptGranularity != "" {
			path += "&filter[granularity]=" + url.QueryEscape(rptGranularity)
		}
		instances, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(instances) == 0 {
			fmt.Println("No report instances found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "GRANULARITY\tPROCESSING DATE\tID")
		for i := range instances {
			r := &instances[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Str("granularity"), r.Str("processingDate"), r.ID)
		}
		return w.Flush()
	},
}

var rptAnalyticsSegmentsCmd = &cobra.Command{
	Use:   "segments",
	Short: "List or download an analytics report instance's data segments",
	Long: `List the data segments of a report instance. Without --output-dir the
segment URLs and sizes are printed; with --output-dir each segment is
downloaded to <dir>/<checksum>.csv.gz (the files are gzip-compressed CSV).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		segments, err := c.List(ctx, "/v1/analyticsReportInstances/"+rptInstanceID+"/segments?limit=200")
		if err != nil {
			return err
		}
		if len(segments) == 0 {
			fmt.Println("No segments found.")
			return nil
		}
		if rptOutputDir == "" {
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "SIZE\tCHECKSUM\tURL")
			for i := range segments {
				s := &segments[i]
				fmt.Fprintf(w, "%d\t%s\t%s\n", rptInt(s, "sizeInBytes"), s.Str("checksum"), s.Str("url"))
			}
			return w.Flush()
		}
		if err := os.MkdirAll(rptOutputDir, 0o755); err != nil {
			return err
		}
		for i := range segments {
			s := &segments[i]
			segURL := s.Str("url")
			if segURL == "" {
				return fmt.Errorf("segment %s has no url attribute", s.ID)
			}
			data, err := c.Download(ctx, segURL, "")
			if err != nil {
				return fmt.Errorf("segment %d/%d: %w", i+1, len(segments), err)
			}
			name := s.Str("checksum")
			if name == "" {
				name = fmt.Sprintf("segment-%d", i+1)
			}
			path := filepath.Join(rptOutputDir, name+".csv.gz")
			if sum := s.Str("checksum"); sum != "" && api.MD5Hex(data) != sum {
				fmt.Fprintf(os.Stderr, "Warning: checksum mismatch for %s\n", path)
			}
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("Downloaded %s (%d bytes)\n", path, len(data))
		}
		return nil
	},
}

// rptInt returns a numeric attribute as int64 (JSON numbers decode as float64).
func rptInt(r *api.Resource, key string) int64 {
	if v, ok := r.Attributes[key].(float64); ok {
		return int64(v)
	}
	return 0
}

// rptBoolStr returns a boolean attribute as "true"/"false", or "" when absent.
func rptBoolStr(r *api.Resource, key string) string {
	if v, ok := r.Attributes[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

func init() {
	rptSalesCmd.Flags().StringVar(&rptVendor, "vendor", "", "vendor number from Payments and Financial Reports (required)")
	_ = rptSalesCmd.MarkFlagRequired("vendor")
	rptSalesCmd.Flags().StringVar(&rptType, "type", "SALES", "report type: SALES, PRE_ORDER, NEWSSTAND, SUBSCRIPTION, SUBSCRIPTION_EVENT, SUBSCRIBER, SUBSCRIPTION_OFFER_CODE_REDEMPTION, INSTALLS, FIRST_ANNUAL, WIN_BACK_ELIGIBILITY")
	rptSalesCmd.Flags().StringVar(&rptSubtype, "subtype", "SUMMARY", "report subtype: SUMMARY, DETAILED, SUMMARY_INSTALL_TYPE, SUMMARY_TERRITORY, SUMMARY_CHANNEL")
	rptSalesCmd.Flags().StringVar(&rptFrequency, "frequency", "DAILY", "frequency: DAILY, WEEKLY, MONTHLY, YEARLY")
	rptSalesCmd.Flags().StringVar(&rptDate, "date", "", "report date, e.g. 2026-07-15 (daily), 2026-07 (monthly); latest when omitted")
	rptSalesCmd.Flags().StringVar(&rptVersion, "version", "", "report version, e.g. 1_0, 1_1")
	rptSalesCmd.Flags().StringVar(&rptOutput, "output", "", "write to file instead of stdout")
	rptSalesCmd.Flags().BoolVar(&rptGzip, "gzip", false, "keep the raw gzip bytes instead of decompressing")

	rptFinanceCmd.Flags().StringVar(&rptVendor, "vendor", "", "vendor number from Payments and Financial Reports (required)")
	_ = rptFinanceCmd.MarkFlagRequired("vendor")
	rptFinanceCmd.Flags().StringVar(&rptFinType, "type", "FINANCIAL", "report type: FINANCIAL or FINANCE_DETAIL")
	rptFinanceCmd.Flags().StringVar(&rptRegion, "region", "", "financial region code, e.g. JP, US, ZZ (required)")
	_ = rptFinanceCmd.MarkFlagRequired("region")
	rptFinanceCmd.Flags().StringVar(&rptDate, "date", "", "fiscal period, e.g. 2026-06 (required)")
	_ = rptFinanceCmd.MarkFlagRequired("date")
	rptFinanceCmd.Flags().StringVar(&rptOutput, "output", "", "write to file instead of stdout")
	rptFinanceCmd.Flags().BoolVar(&rptGzip, "gzip", false, "keep the raw gzip bytes instead of decompressing")

	rptReportsCmd.AddCommand(rptSalesCmd, rptFinanceCmd)
	rootCmd.AddCommand(rptReportsCmd)

	for _, sub := range []*cobra.Command{rptAnalyticsRequestCmd, rptAnalyticsRequestsCmd} {
		sub.Flags().StringVar(&rptApp, "app", "", "app id or bundle id (required)")
		_ = sub.MarkFlagRequired("app")
	}
	rptAnalyticsRequestCmd.Flags().StringVar(&rptAccessType, "access-type", "ONGOING", "ONGOING or ONE_TIME_SNAPSHOT")

	rptAnalyticsReportsCmd.Flags().StringVar(&rptRequestID, "request-id", "", "analytics report request id (required)")
	_ = rptAnalyticsReportsCmd.MarkFlagRequired("request-id")
	rptAnalyticsReportsCmd.Flags().StringVar(&rptCategory, "category", "", "filter by category: APP_USAGE, APP_STORE_ENGAGEMENT, COMMERCE, FRAMEWORK_USAGE, PERFORMANCE")

	rptAnalyticsInstancesCmd.Flags().StringVar(&rptReportID, "report-id", "", "analytics report id (required)")
	_ = rptAnalyticsInstancesCmd.MarkFlagRequired("report-id")
	rptAnalyticsInstancesCmd.Flags().StringVar(&rptGranularity, "granularity", "", "filter by granularity: DAILY, WEEKLY, MONTHLY")

	rptAnalyticsSegmentsCmd.Flags().StringVar(&rptInstanceID, "instance-id", "", "analytics report instance id (required)")
	_ = rptAnalyticsSegmentsCmd.MarkFlagRequired("instance-id")
	rptAnalyticsSegmentsCmd.Flags().StringVar(&rptOutputDir, "output-dir", "", "download segments into this directory (list only when omitted)")

	rptAnalyticsCmd.AddCommand(rptAnalyticsRequestCmd, rptAnalyticsRequestsCmd, rptAnalyticsReportsCmd, rptAnalyticsInstancesCmd, rptAnalyticsSegmentsCmd)
	rootCmd.AddCommand(rptAnalyticsCmd)
}
