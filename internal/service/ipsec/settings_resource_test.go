package ipsec_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The test VM has automatic VPN rules on and no passthrough networks; the
// last step puts that back.
func TestAccIpsecSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_ipsec_settings" "test" {}`,
				ResourceName:       "opnsense_ipsec_settings.test",
				ImportState:        true,
				ImportStateId:      "ipsec_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccIpsecSettingsConfig(true, `["192.0.2.0/24", "198.51.100.0/24"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_settings.test", "disable_vpn_rules", "true"),
					resource.TestCheckResourceAttr("opnsense_ipsec_settings.test", "passthrough_networks.#", "2"),
					// Unset attributes keep the firewall's value.
					resource.TestCheckResourceAttr("opnsense_ipsec_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("data.opnsense_ipsec_settings.test", "passthrough_networks.#", "2"),
				),
			},
			{
				Config: testAccIpsecSettingsConfig(false, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_settings.test", "disable_vpn_rules", "false"),
					resource.TestCheckResourceAttr("opnsense_ipsec_settings.test", "passthrough_networks.#", "0"),
				),
			},
		},
	})
}

func TestAccIpsecSettingsResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "opnsense_ipsec_settings" "test" {}`,
				ExpectError: regexp.MustCompile("Import Required"),
			},
		},
	})
}

func testAccIpsecSettingsConfig(noRules bool, passthrough string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_settings" "test" {
  disable_vpn_rules    = %[1]t
  passthrough_networks = %[2]s
}

data "opnsense_ipsec_settings" "test" {
  depends_on = [opnsense_ipsec_settings.test]
}
`, noRules, passthrough)
}
