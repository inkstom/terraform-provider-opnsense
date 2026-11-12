package captiveportal_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A zone needs a non-WAN interface, so the test assigns a VLAN on the VM's
// only NIC; interface assignment needs OPNsense 26.7.7.
func TestAccCaptiveportalZoneResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.MinVersionPreCheck(t, "26.7.7") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCaptiveportalZoneConfig(true, 30, `["192.0.2.10", "198.51.100.0/24"]`, `["00:11:22:33:44:55"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("opnsense_captiveportal_zone.test", "interfaces.0", "opnsense_interfaces_assignment.cp", "id"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "auth_servers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "idle_timeout", "30"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "allowed_addresses.#", "2"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "allowed_mac_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("opnsense_captiveportal_zone.test", "zone_id"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "roaming", "true"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "cors_policy", "any"),
					resource.TestCheckResourceAttr("data.opnsense_captiveportal_zone.test", "idle_timeout", "30"),
					testAccCheckPortalStatus("running"),
				),
			},
			{
				ResourceName:      "opnsense_captiveportal_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCaptiveportalZoneConfig(false, 60, `[]`, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "idle_timeout", "60"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "allowed_addresses.#", "0"),
					resource.TestCheckResourceAttr("opnsense_captiveportal_zone.test", "allowed_mac_addresses.#", "0"),
					testAccCheckPortalStatus("disabled"),
				),
			},
		},
	})
}

func testAccCaptiveportalZoneConfig(enabled bool, idle int, addrs, macs string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vlan" "cp" {
  description = "tfacc captive portal"
  tag         = 3905
  priority    = 0
  parent      = "vtnet0"
  device      = "vlan0.3905"
}

resource "opnsense_interfaces_assignment" "cp" {
  device       = opnsense_interfaces_vlan.cp.device
  description  = "TFACCCP"
  ipv4_type    = "staticv4"
  ipv4_address = "192.0.2.1/24"
}

resource "opnsense_captiveportal_zone" "test" {
  enabled               = %[1]t
  interfaces            = [opnsense_interfaces_assignment.cp.id]
  auth_servers          = ["Local Database"]
  idle_timeout          = %[2]d
  hard_timeout          = 120
  allowed_addresses     = %[3]s
  allowed_mac_addresses = %[4]s
  description           = "tfacc"
}

data "opnsense_captiveportal_zone" "test" {
  id = opnsense_captiveportal_zone.test.id
}
`, enabled, idle, addrs, macs)
}

// testAccCheckPortalStatus checks the captive portal service itself, which
// runs only while at least one zone is enabled.
func testAccCheckPortalStatus(want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/captiveportal/service/status", nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"},
		}}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var status struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return err
		}
		if status.Status != want {
			return fmt.Errorf("captive portal service status = %q, want %q", status.Status, want)
		}
		return nil
	}
}
