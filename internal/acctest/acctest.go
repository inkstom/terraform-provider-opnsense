package acctest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	ProviderName = "opnsense"
)

var (
	ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error) = protoV6ProviderFactoriesInit(context.Background(), ProviderName)
)

func protoV6ProviderFactoriesInit(ctx context.Context, providerNames ...string) map[string]func() (tfprotov6.ProviderServer, error) {
	factories := make(map[string]func() (tfprotov6.ProviderServer, error))

	for _, name := range providerNames {
		if name == ProviderName {
			serverFactory, _, err := provider.ProtoV6ProviderServerFactory(ctx)
			if err != nil {
				log.Fatal(err)
			}
			factories[name] = func() (tfprotov6.ProviderServer, error) {
				return serverFactory(), nil
			}
		}
	}

	return factories
}

func AccPreCheck(t *testing.T) {
	if v := os.Getenv("OPNSENSE_URI"); v == "" {
		t.Fatal("OPNSENSE_URI must be set for acceptance tests")
	}
	if v := os.Getenv("OPNSENSE_API_KEY"); v == "" {
		t.Fatal("OPNSENSE_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("OPNSENSE_API_SECRET"); v == "" {
		t.Fatal("OPNSENSE_API_SECRET must be set for acceptance tests")
	}
}

// DomainOverridePreCheck skips the test unless OPNSENSE_TEST_DOMAIN_OVERRIDE=1.
// The addDomainOverride endpoint was removed in OPNsense 25+ and returns an empty
// body on newer versions. Set this variable only on OPNsense 24.x or earlier.
func DomainOverridePreCheck(t *testing.T) {
	t.Helper()
	AccPreCheck(t)
	if os.Getenv("OPNSENSE_TEST_DOMAIN_OVERRIDE") != "1" {
		t.Skip("OPNSENSE_TEST_DOMAIN_OVERRIDE=1 required: addDomainOverride endpoint removed in OPNsense 25+")
	}
}

// KeaDhcpv6PreCheck skips the test unless OPNSENSE_KEA_DHCPV6_IFACE is set to a
// valid interface name (e.g. "wan"). When set, it configures that interface in the
// Kea DHCPv6 general settings before the test runs and resets it to empty on cleanup.
func KeaDhcpv6PreCheck(t *testing.T) {
	t.Helper()
	AccPreCheck(t)

	iface := os.Getenv("OPNSENSE_KEA_DHCPV6_IFACE")
	if iface == "" {
		t.Skip("OPNSENSE_KEA_DHCPV6_IFACE must be set to a valid interface (e.g. 'wan') for Kea DHCPv6 tests")
	}

	uri := os.Getenv("OPNSENSE_URI")
	key := os.Getenv("OPNSENSE_API_KEY")
	secret := os.Getenv("OPNSENSE_API_SECRET")

	if err := setKeaDhcpv6Interface(t, uri, key, secret, iface); err != nil {
		t.Fatalf("KeaDhcpv6PreCheck: failed to configure interface %q: %s", iface, err)
	}

	t.Cleanup(func() {
		if err := setKeaDhcpv6Interface(t, uri, key, secret, ""); err != nil {
			t.Logf("KeaDhcpv6PreCheck cleanup: failed to reset interface: %s", err)
		}
	})
}

func setKeaDhcpv6Interface(t *testing.T, uri, key, secret, iface string) error {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"dhcpv6": map[string]any{
			"general": map[string]any{
				"interfaces": iface,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri+"/api/kea/dhcpv6/set", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(key+":"+secret)))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
		},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if result.Result != "saved" {
		return fmt.Errorf("unexpected result: %q", result.Result)
	}
	return nil
}

// MinVersionPreCheck skips the test when the target OPNsense is older than
// minVersion (e.g. "26.7"), for features whose API does not exist there.
// CI runs OPNsense 26.1; some resources need 26.7 or later.
func MinVersionPreCheck(t *testing.T, minVersion string) {
	t.Helper()
	AccPreCheck(t)

	have, err := opnsenseVersion()
	if err != nil {
		t.Fatalf("MinVersionPreCheck: %s", err)
	}
	if compareVersions(have, minVersion) < 0 {
		t.Skipf("requires OPNsense %s or later; target runs %s", minVersion, have)
	}
}

func opnsenseVersion() (string, error) {
	req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/core/firmware/status", nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(os.Getenv("OPNSENSE_API_KEY")+":"+os.Getenv("OPNSENSE_API_SECRET"))))

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"},
	}}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("firmware status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("firmware status: unexpected status %d", resp.StatusCode)
	}

	var status struct {
		Product struct {
			Version string `json:"product_version"`
		} `json:"product"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", fmt.Errorf("firmware status: decode: %w", err)
	}
	if status.Product.Version == "" {
		return "", fmt.Errorf("firmware status: no product_version")
	}
	return status.Product.Version, nil
}

// compareVersions compares dotted numeric versions such as "26.1.6" and
// "26.7", ignoring suffixes like "_1". Missing components count as 0.
func compareVersions(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(v string) []int {
	v, _, _ = strings.Cut(v, "_")
	v, _, _ = strings.Cut(v, "-")
	var out []int
	for _, s := range strings.Split(v, ".") {
		n, err := strconv.Atoi(s)
		if err != nil {
			break
		}
		out = append(out, n)
	}
	return out
}
