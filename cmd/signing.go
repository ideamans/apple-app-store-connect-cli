package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// Enum values from the App Store Connect API spec (4.4.1).
const (
	sgnCertTypeHelp = "certificate type: IOS_DEVELOPMENT, IOS_DISTRIBUTION, MAC_APP_DEVELOPMENT, MAC_APP_DISTRIBUTION, MAC_INSTALLER_DISTRIBUTION, DEVELOPMENT, DISTRIBUTION, DEVELOPER_ID_APPLICATION, DEVELOPER_ID_APPLICATION_G2, DEVELOPER_ID_KEXT, DEVELOPER_ID_KEXT_G2, APPLE_PAY, APPLE_PAY_MERCHANT_IDENTITY, APPLE_PAY_PSP_IDENTITY, APPLE_PAY_RSA, IDENTITY_ACCESS, PASS_TYPE_ID, PASS_TYPE_ID_WITH_NFC"

	sgnPlatformHelp = "platform: IOS, MAC_OS or UNIVERSAL"

	sgnCapabilityHelp = "capability: ICLOUD, IN_APP_PURCHASE, GAME_CENTER, PUSH_NOTIFICATIONS, WALLET, INTER_APP_AUDIO, MAPS, ASSOCIATED_DOMAINS, PERSONAL_VPN, APP_GROUPS, HEALTHKIT, HOMEKIT, WIRELESS_ACCESSORY_CONFIGURATION, APPLE_PAY, DATA_PROTECTION, SIRIKIT, NETWORK_EXTENSIONS, MULTIPATH, HOT_SPOT, NFC_TAG_READING, CLASSKIT, AUTOFILL_CREDENTIAL_PROVIDER, ACCESS_WIFI_INFORMATION, NETWORK_CUSTOM_PROTOCOL, COREMEDIA_HLS_LOW_LATENCY, SYSTEM_EXTENSION_INSTALL, USER_MANAGEMENT, APPLE_ID_AUTH"

	sgnProfileTypeHelp = "profile type: IOS_APP_DEVELOPMENT, IOS_APP_STORE, IOS_APP_ADHOC, IOS_APP_INHOUSE, MAC_APP_DEVELOPMENT, MAC_APP_STORE, MAC_APP_DIRECT, TVOS_APP_DEVELOPMENT, TVOS_APP_STORE, TVOS_APP_ADHOC, TVOS_APP_INHOUSE, MAC_CATALYST_APP_DEVELOPMENT, MAC_CATALYST_APP_STORE, MAC_CATALYST_APP_DIRECT"
)

var sgnCertificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Manage signing certificates",
}

var sgnBundleIDsCmd = &cobra.Command{
	Use:   "bundle-ids",
	Short: "Manage bundle IDs and their capabilities",
}

var sgnDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage registered devices",
}

var sgnProvProfilesCmd = &cobra.Command{
	Use:   "provisioning-profiles",
	Short: "Manage provisioning profiles",
}

var sgnPassTypeIDsCmd = &cobra.Command{
	Use:   "pass-type-ids",
	Short: "Manage Wallet pass type IDs",
}

var sgnMerchantIDsCmd = &cobra.Command{
	Use:   "merchant-ids",
	Short: "Manage Apple Pay merchant IDs",
}

var (
	sgnCertType       string
	sgnCertCSR        string
	sgnCertOutput     string
	sgnCertMerchantID string
	sgnCertPassTypeID string
	sgnCertID         string

	sgnBidIdentifier string
	sgnBidName       string
	sgnBidPlatform   string
	sgnBidID         string
	sgnBidCapability string

	sgnDevName     string
	sgnDevUDID     string
	sgnDevPlatform string
	sgnDevID       string
	sgnDevStatus   string

	sgnProfName      string
	sgnProfType      string
	sgnProfBundleID  string
	sgnProfCertIDs   string
	sgnProfDeviceIDs string
	sgnProfID        string
	sgnProfOutput    string

	sgnPassIdentifier string
	sgnPassName       string
	sgnPassID         string

	sgnMerchIdentifier string
	sgnMerchName       string
	sgnMerchID         string
)

// sgnSplitList splits a comma-separated flag value, trimming blanks.
func sgnSplitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// sgnRelList builds a to-many relationship value: {"data":[{"type":..,"id":..},...]}.
func sgnRelList(typ string, ids []string) json.RawMessage {
	items := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]string{"type": typ, "id": id})
	}
	b, _ := json.Marshal(map[string]any{"data": items})
	return b
}

// sgnDate trims an ISO 8601 timestamp to its date part for table output.
func sgnDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// sgnCSRBase64 reads a certificate signing request file and returns the base64
// body the API expects in csrContent: PEM armor lines and all whitespace are
// stripped, leaving only the base64-encoded DER.
func sgnCSRBase64(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var body []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-----") {
			continue
		}
		body = append(body, line)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("no CSR content found in %s", path)
	}
	return strings.Join(body, ""), nil
}

// sgnWriteCertFile decodes certificateContent (base64 DER) and writes a .cer file.
func sgnWriteCertFile(cert *api.Resource, output string) error {
	content := cert.Str("certificateContent")
	if content == "" {
		return fmt.Errorf("certificate %s has no certificateContent", cert.ID)
	}
	der, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("decode certificateContent: %w", err)
	}
	if output == "" {
		if sn := cert.Str("serialNumber"); sn != "" {
			output = sn + ".cer"
		} else {
			output = cert.ID + ".cer"
		}
	}
	if err := os.WriteFile(output, der, 0o644); err != nil {
		return err
	}
	fmt.Printf("Wrote certificate to %s.\n", output)
	return nil
}

// --- certificates ---------------------------------------------------------

var sgnCertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List signing certificates",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/certificates?limit=200"
		if sgnCertType != "" {
			path += "&filter[certificateType]=" + url.QueryEscape(sgnCertType)
		}
		certs, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(certs) == 0 {
			fmt.Println("No certificates found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "SERIAL\tTYPE\tNAME\tEXPIRES\tID")
		for i := range certs {
			ct := &certs[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ct.Str("serialNumber"), ct.Str("certificateType"), ct.Str("name"), sgnDate(ct.Str("expirationDate")), ct.ID)
		}
		return w.Flush()
	},
}

var sgnCertCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a signing certificate from a CSR file",
	Example: `  asc certificates create --type IOS_DISTRIBUTION --csr CertificateSigningRequest.certSigningRequest
  asc certificates create --type DEVELOPMENT --csr request.csr --output development.cer`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		csr, err := sgnCSRBase64(sgnCertCSR)
		if err != nil {
			return err
		}
		rels := map[string]json.RawMessage{}
		if sgnCertMerchantID != "" {
			rels["merchantId"] = api.Rel("merchantIds", sgnCertMerchantID)
		}
		if sgnCertPassTypeID != "" {
			rels["passTypeId"] = api.Rel("passTypeIds", sgnCertPassTypeID)
		}
		if len(rels) == 0 {
			rels = nil
		}
		cert, err := c.Post(cmd.Context(), "/v1/certificates", api.Body{
			Data: api.Resource{
				Type: "certificates",
				Attributes: map[string]any{
					"csrContent":      csr,
					"certificateType": sgnCertType,
				},
				Relationships: rels,
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Certificate created: %s (serial %s, id %s).\n", cert.Str("name"), cert.Str("serialNumber"), cert.ID)
		if cert.Str("certificateContent") == "" {
			// Dry-run stub or content withheld; nothing to write.
			return nil
		}
		return sgnWriteCertFile(cert, sgnCertOutput)
	},
}

var sgnCertDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a certificate as a .cer file",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		cert, _, err := c.Get(cmd.Context(), "/v1/certificates/"+sgnCertID)
		if err != nil {
			return err
		}
		return sgnWriteCertFile(cert, sgnCertOutput)
	},
}

var sgnCertRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/certificates/"+sgnCertID); err != nil {
			return err
		}
		fmt.Printf("Certificate %s revoked.\n", sgnCertID)
		return nil
	},
}

// --- bundle-ids -----------------------------------------------------------

var sgnBidListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bundle IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		bids, err := c.List(cmd.Context(), "/v1/bundleIds?limit=200")
		if err != nil {
			return err
		}
		if len(bids) == 0 {
			fmt.Println("No bundle IDs found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "IDENTIFIER\tNAME\tPLATFORM\tID")
		for i := range bids {
			b := &bids[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.Str("identifier"), b.Str("name"), b.Str("platform"), b.ID)
		}
		return w.Flush()
	},
}

var sgnBidCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a new bundle ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		bid, err := c.Post(cmd.Context(), "/v1/bundleIds", api.Body{
			Data: api.Resource{
				Type: "bundleIds",
				Attributes: map[string]any{
					"identifier": sgnBidIdentifier,
					"name":       sgnBidName,
					"platform":   sgnBidPlatform,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Bundle ID %s created (id %s).\n", sgnBidIdentifier, bid.ID)
		return nil
	},
}

var sgnBidDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a bundle ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/bundleIds/"+sgnBidID); err != nil {
			return err
		}
		fmt.Printf("Bundle ID %s deleted.\n", sgnBidID)
		return nil
	},
}

var sgnBidCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "List the capabilities enabled on a bundle ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		caps, err := c.List(cmd.Context(), "/v1/bundleIds/"+sgnBidID+"/bundleIdCapabilities?limit=200")
		if err != nil {
			return err
		}
		if len(caps) == 0 {
			fmt.Println("No capabilities enabled.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "CAPABILITY\tID")
		for i := range caps {
			fmt.Fprintf(w, "%s\t%s\n", caps[i].Str("capabilityType"), caps[i].ID)
		}
		return w.Flush()
	},
}

var sgnBidEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a capability on a bundle ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		capType := strings.ToUpper(strings.TrimSpace(sgnBidCapability))
		if _, err := c.Post(cmd.Context(), "/v1/bundleIdCapabilities", api.Body{
			Data: api.Resource{
				Type:          "bundleIdCapabilities",
				Attributes:    map[string]any{"capabilityType": capType},
				Relationships: map[string]json.RawMessage{"bundleId": api.Rel("bundleIds", sgnBidID)},
			},
		}); err != nil {
			return err
		}
		fmt.Printf("Capability %s enabled on bundle ID %s.\n", capType, sgnBidID)
		return nil
	},
}

var sgnBidDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a capability on a bundle ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		capType := strings.ToUpper(strings.TrimSpace(sgnBidCapability))
		caps, err := c.List(ctx, "/v1/bundleIds/"+sgnBidID+"/bundleIdCapabilities?limit=200")
		if err != nil {
			return err
		}
		found := findByAttr(caps, "capabilityType", capType)
		if found == nil {
			return fmt.Errorf("capability %s is not enabled on bundle ID %s", capType, sgnBidID)
		}
		if err := c.Delete(ctx, "/v1/bundleIdCapabilities/"+found.ID); err != nil {
			return err
		}
		fmt.Printf("Capability %s disabled on bundle ID %s.\n", capType, sgnBidID)
		return nil
	},
}

// --- devices --------------------------------------------------------------

var sgnDevListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/devices?limit=200"
		if sgnDevPlatform != "" {
			path += "&filter[platform]=" + url.QueryEscape(sgnDevPlatform)
		}
		devs, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(devs) == 0 {
			fmt.Println("No devices found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUDID\tPLATFORM\tSTATUS\tID")
		for i := range devs {
			d := &devs[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Str("name"), d.Str("udid"), d.Str("platform"), d.Str("status"), d.ID)
		}
		return w.Flush()
	},
}

var sgnDevRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new device",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		dev, err := c.Post(cmd.Context(), "/v1/devices", api.Body{
			Data: api.Resource{
				Type: "devices",
				Attributes: map[string]any{
					"name":     sgnDevName,
					"udid":     sgnDevUDID,
					"platform": sgnDevPlatform,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Device %s registered (id %s).\n", sgnDevName, dev.ID)
		return nil
	},
}

var sgnDevUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Rename a device or change its status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		attrs := map[string]any{}
		if cmd.Flags().Changed("name") {
			attrs["name"] = sgnDevName
		}
		if cmd.Flags().Changed("status") {
			attrs["status"] = strings.ToUpper(strings.TrimSpace(sgnDevStatus))
		}
		if len(attrs) == 0 {
			return fmt.Errorf("nothing to update (use --name and/or --status)")
		}
		if _, err := c.Patch(cmd.Context(), "/v1/devices/"+sgnDevID, api.Body{
			Data: api.Resource{Type: "devices", ID: sgnDevID, Attributes: attrs},
		}); err != nil {
			return err
		}
		fmt.Printf("Device %s updated.\n", sgnDevID)
		return nil
	},
}

// --- provisioning-profiles --------------------------------------------------

var sgnProfListCmd = &cobra.Command{
	Use:   "list",
	Short: "List provisioning profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		path := "/v1/profiles?limit=200"
		if sgnProfType != "" {
			path += "&filter[profileType]=" + url.QueryEscape(sgnProfType)
		}
		profs, err := c.List(cmd.Context(), path)
		if err != nil {
			return err
		}
		if len(profs) == 0 {
			fmt.Println("No provisioning profiles found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATE\tEXPIRES\tID")
		for i := range profs {
			p := &profs[i]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.Str("name"), p.Str("profileType"), p.Str("profileState"), sgnDate(p.Str("expirationDate")), p.ID)
		}
		return w.Flush()
	},
}

var sgnProfCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a provisioning profile",
	Example: `  asc provisioning-profiles create --name "My App Store" --type IOS_APP_STORE \
    --bundle-id-id ABC123 --certificate-ids CERT1
  asc provisioning-profiles create --name "My Dev" --type IOS_APP_DEVELOPMENT \
    --bundle-id-id ABC123 --certificate-ids CERT1,CERT2 --device-ids DEV1,DEV2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		certIDs := sgnSplitList(sgnProfCertIDs)
		if len(certIDs) == 0 {
			return fmt.Errorf("--certificate-ids must list at least one certificate id")
		}
		rels := map[string]json.RawMessage{
			"bundleId":     api.Rel("bundleIds", sgnProfBundleID),
			"certificates": sgnRelList("certificates", certIDs),
		}
		if deviceIDs := sgnSplitList(sgnProfDeviceIDs); len(deviceIDs) > 0 {
			rels["devices"] = sgnRelList("devices", deviceIDs)
		}
		prof, err := c.Post(cmd.Context(), "/v1/profiles", api.Body{
			Data: api.Resource{
				Type: "profiles",
				Attributes: map[string]any{
					"name":        sgnProfName,
					"profileType": sgnProfType,
				},
				Relationships: rels,
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Provisioning profile %s created (id %s).\n", sgnProfName, prof.ID)
		return nil
	},
}

var sgnProfDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a provisioning profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/profiles/"+sgnProfID); err != nil {
			return err
		}
		fmt.Printf("Provisioning profile %s deleted.\n", sgnProfID)
		return nil
	},
}

var sgnProfDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a provisioning profile as a .mobileprovision file",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		prof, _, err := c.Get(cmd.Context(), "/v1/profiles/"+sgnProfID)
		if err != nil {
			return err
		}
		content := prof.Str("profileContent")
		if content == "" {
			return fmt.Errorf("profile %s has no profileContent", sgnProfID)
		}
		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return fmt.Errorf("decode profileContent: %w", err)
		}
		output := sgnProfOutput
		if output == "" {
			if name := prof.Str("name"); name != "" {
				output = name + ".mobileprovision"
			} else {
				output = prof.ID + ".mobileprovision"
			}
		}
		if err := os.WriteFile(output, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("Wrote provisioning profile to %s.\n", output)
		return nil
	},
}

// --- pass-type-ids ----------------------------------------------------------

var sgnPassListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Wallet pass type IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		items, err := c.List(cmd.Context(), "/v1/passTypeIds?limit=200")
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("No pass type IDs found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "IDENTIFIER\tNAME\tID")
		for i := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\n", items[i].Str("identifier"), items[i].Str("name"), items[i].ID)
		}
		return w.Flush()
	},
}

var sgnPassCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a new Wallet pass type ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		item, err := c.Post(cmd.Context(), "/v1/passTypeIds", api.Body{
			Data: api.Resource{
				Type: "passTypeIds",
				Attributes: map[string]any{
					"identifier": sgnPassIdentifier,
					"name":       sgnPassName,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Pass type ID %s created (id %s).\n", sgnPassIdentifier, item.ID)
		return nil
	},
}

var sgnPassDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Wallet pass type ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/passTypeIds/"+sgnPassID); err != nil {
			return err
		}
		fmt.Printf("Pass type ID %s deleted.\n", sgnPassID)
		return nil
	},
}

// --- merchant-ids -----------------------------------------------------------

var sgnMerchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Apple Pay merchant IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		items, err := c.List(cmd.Context(), "/v1/merchantIds?limit=200")
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("No merchant IDs found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "IDENTIFIER\tNAME\tID")
		for i := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\n", items[i].Str("identifier"), items[i].Str("name"), items[i].ID)
		}
		return w.Flush()
	},
}

var sgnMerchCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a new Apple Pay merchant ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		item, err := c.Post(cmd.Context(), "/v1/merchantIds", api.Body{
			Data: api.Resource{
				Type: "merchantIds",
				Attributes: map[string]any{
					"identifier": sgnMerchIdentifier,
					"name":       sgnMerchName,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Merchant ID %s created (id %s).\n", sgnMerchIdentifier, item.ID)
		return nil
	},
}

var sgnMerchDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an Apple Pay merchant ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}
		if err := c.Delete(cmd.Context(), "/v1/merchantIds/"+sgnMerchID); err != nil {
			return err
		}
		fmt.Printf("Merchant ID %s deleted.\n", sgnMerchID)
		return nil
	},
}

func init() {
	// certificates
	sgnCertListCmd.Flags().StringVar(&sgnCertType, "type", "", "filter by "+sgnCertTypeHelp)
	sgnCertCreateCmd.Flags().StringVar(&sgnCertType, "type", "", sgnCertTypeHelp+" (required)")
	sgnCertCreateCmd.Flags().StringVar(&sgnCertCSR, "csr", "", "path to a .certSigningRequest / PEM CSR file (required)")
	sgnCertCreateCmd.Flags().StringVar(&sgnCertOutput, "output", "", "output .cer path (default: <serial>.cer)")
	sgnCertCreateCmd.Flags().StringVar(&sgnCertMerchantID, "merchant-id-id", "", "merchantIds resource id (required for APPLE_PAY* types)")
	sgnCertCreateCmd.Flags().StringVar(&sgnCertPassTypeID, "pass-type-id-id", "", "passTypeIds resource id (required for PASS_TYPE_ID* types)")
	_ = sgnCertCreateCmd.MarkFlagRequired("type")
	_ = sgnCertCreateCmd.MarkFlagRequired("csr")
	sgnCertDownloadCmd.Flags().StringVar(&sgnCertID, "id", "", "certificate id (required)")
	sgnCertDownloadCmd.Flags().StringVar(&sgnCertOutput, "output", "", "output .cer path (default: <serial>.cer)")
	_ = sgnCertDownloadCmd.MarkFlagRequired("id")
	sgnCertRevokeCmd.Flags().StringVar(&sgnCertID, "id", "", "certificate id (required)")
	_ = sgnCertRevokeCmd.MarkFlagRequired("id")
	sgnCertificatesCmd.AddCommand(sgnCertListCmd, sgnCertCreateCmd, sgnCertDownloadCmd, sgnCertRevokeCmd)

	// bundle-ids
	sgnBidCreateCmd.Flags().StringVar(&sgnBidIdentifier, "identifier", "", "bundle identifier, e.g. com.example.app (required)")
	sgnBidCreateCmd.Flags().StringVar(&sgnBidName, "name", "", "bundle ID name (required)")
	sgnBidCreateCmd.Flags().StringVar(&sgnBidPlatform, "platform", "", sgnPlatformHelp+" (required)")
	_ = sgnBidCreateCmd.MarkFlagRequired("identifier")
	_ = sgnBidCreateCmd.MarkFlagRequired("name")
	_ = sgnBidCreateCmd.MarkFlagRequired("platform")
	for _, sub := range []*cobra.Command{sgnBidDeleteCmd, sgnBidCapabilitiesCmd, sgnBidEnableCmd, sgnBidDisableCmd} {
		sub.Flags().StringVar(&sgnBidID, "id", "", "bundle ID resource id (required)")
		_ = sub.MarkFlagRequired("id")
	}
	for _, sub := range []*cobra.Command{sgnBidEnableCmd, sgnBidDisableCmd} {
		sub.Flags().StringVar(&sgnBidCapability, "capability", "", sgnCapabilityHelp+" (required)")
		_ = sub.MarkFlagRequired("capability")
	}
	sgnBundleIDsCmd.AddCommand(sgnBidListCmd, sgnBidCreateCmd, sgnBidDeleteCmd, sgnBidCapabilitiesCmd, sgnBidEnableCmd, sgnBidDisableCmd)

	// devices
	sgnDevListCmd.Flags().StringVar(&sgnDevPlatform, "platform", "", "filter by "+sgnPlatformHelp)
	sgnDevRegisterCmd.Flags().StringVar(&sgnDevName, "name", "", "device name (required)")
	sgnDevRegisterCmd.Flags().StringVar(&sgnDevUDID, "udid", "", "device UDID (required)")
	sgnDevRegisterCmd.Flags().StringVar(&sgnDevPlatform, "platform", "", sgnPlatformHelp+" (required)")
	_ = sgnDevRegisterCmd.MarkFlagRequired("name")
	_ = sgnDevRegisterCmd.MarkFlagRequired("udid")
	_ = sgnDevRegisterCmd.MarkFlagRequired("platform")
	sgnDevUpdateCmd.Flags().StringVar(&sgnDevID, "id", "", "device id (required)")
	sgnDevUpdateCmd.Flags().StringVar(&sgnDevName, "name", "", "new device name")
	sgnDevUpdateCmd.Flags().StringVar(&sgnDevStatus, "status", "", "ENABLED or DISABLED")
	_ = sgnDevUpdateCmd.MarkFlagRequired("id")
	sgnDevicesCmd.AddCommand(sgnDevListCmd, sgnDevRegisterCmd, sgnDevUpdateCmd)

	// provisioning-profiles
	sgnProfListCmd.Flags().StringVar(&sgnProfType, "type", "", "filter by "+sgnProfileTypeHelp)
	sgnProfCreateCmd.Flags().StringVar(&sgnProfName, "name", "", "profile name (required)")
	sgnProfCreateCmd.Flags().StringVar(&sgnProfType, "type", "", sgnProfileTypeHelp+" (required)")
	sgnProfCreateCmd.Flags().StringVar(&sgnProfBundleID, "bundle-id-id", "", "bundle ID resource id (required)")
	sgnProfCreateCmd.Flags().StringVar(&sgnProfCertIDs, "certificate-ids", "", "comma-separated certificate ids (required)")
	sgnProfCreateCmd.Flags().StringVar(&sgnProfDeviceIDs, "device-ids", "", "comma-separated device ids (for development/ad hoc profiles)")
	_ = sgnProfCreateCmd.MarkFlagRequired("name")
	_ = sgnProfCreateCmd.MarkFlagRequired("type")
	_ = sgnProfCreateCmd.MarkFlagRequired("bundle-id-id")
	_ = sgnProfCreateCmd.MarkFlagRequired("certificate-ids")
	sgnProfDeleteCmd.Flags().StringVar(&sgnProfID, "id", "", "profile id (required)")
	_ = sgnProfDeleteCmd.MarkFlagRequired("id")
	sgnProfDownloadCmd.Flags().StringVar(&sgnProfID, "id", "", "profile id (required)")
	sgnProfDownloadCmd.Flags().StringVar(&sgnProfOutput, "output", "", "output path (default: <name>.mobileprovision)")
	_ = sgnProfDownloadCmd.MarkFlagRequired("id")
	sgnProvProfilesCmd.AddCommand(sgnProfListCmd, sgnProfCreateCmd, sgnProfDeleteCmd, sgnProfDownloadCmd)

	// pass-type-ids
	sgnPassCreateCmd.Flags().StringVar(&sgnPassIdentifier, "identifier", "", "pass type identifier, e.g. pass.com.example.coupon (required)")
	sgnPassCreateCmd.Flags().StringVar(&sgnPassName, "name", "", "pass type name (required)")
	_ = sgnPassCreateCmd.MarkFlagRequired("identifier")
	_ = sgnPassCreateCmd.MarkFlagRequired("name")
	sgnPassDeleteCmd.Flags().StringVar(&sgnPassID, "id", "", "pass type ID resource id (required)")
	_ = sgnPassDeleteCmd.MarkFlagRequired("id")
	sgnPassTypeIDsCmd.AddCommand(sgnPassListCmd, sgnPassCreateCmd, sgnPassDeleteCmd)

	// merchant-ids
	sgnMerchCreateCmd.Flags().StringVar(&sgnMerchIdentifier, "identifier", "", "merchant identifier, e.g. merchant.com.example (required)")
	sgnMerchCreateCmd.Flags().StringVar(&sgnMerchName, "name", "", "merchant ID name (required)")
	_ = sgnMerchCreateCmd.MarkFlagRequired("identifier")
	_ = sgnMerchCreateCmd.MarkFlagRequired("name")
	sgnMerchDeleteCmd.Flags().StringVar(&sgnMerchID, "id", "", "merchant ID resource id (required)")
	_ = sgnMerchDeleteCmd.MarkFlagRequired("id")
	sgnMerchantIDsCmd.AddCommand(sgnMerchListCmd, sgnMerchCreateCmd, sgnMerchDeleteCmd)

	rootCmd.AddCommand(sgnCertificatesCmd, sgnBundleIDsCmd, sgnDevicesCmd, sgnProvProfilesCmd, sgnPassTypeIDsCmd, sgnMerchantIDsCmd)
}
