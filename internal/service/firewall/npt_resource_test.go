package firewall_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallNptResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNptConfig("2001:db8:1::/64", true, "[opnsense_firewall_category.test.id]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "interface", "wan"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "internal_prefix", "fd00:1::/64"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "external_prefix", "2001:db8:1::/64"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "log", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "categories.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("opnsense_firewall_npt.test", "categories.*", "opnsense_firewall_category.test", "id"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_npt.test", "external_prefix", "2001:db8:1::/64"),
				),
			},
			{
				ResourceName:      "opnsense_firewall_npt.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallNptConfig("2001:db8:2::/64", false, "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "external_prefix", "2001:db8:2::/64"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "log", "false"),
					resource.TestCheckResourceAttr("opnsense_firewall_npt.test", "categories.#", "0"),
				),
			},
		},
	})
}

func testAccFirewallNptConfig(external string, log bool, categories string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_category" "test" {
  name = "tf-acc-npt"
}

resource "opnsense_firewall_npt" "test" {
  enabled         = false
  log             = %[2]t
  sequence        = 10
  interface       = "wan"
  internal_prefix = "fd00:1::/64"
  external_prefix = %[1]q
  categories      = %[3]s
  description     = "tf acc"
}

data "opnsense_firewall_npt" "test" {
  id = opnsense_firewall_npt.test.id
}
`, external, log, categories)
}
