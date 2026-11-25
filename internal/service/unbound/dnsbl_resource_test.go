package unbound_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUnboundDnsblResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.MinVersionPreCheck(t, "26.1") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Only the required field, so every default is exercised.
			{
				Config: `
resource "opnsense_unbound_dnsbl" "test" {
  description = "tf-acc dnsbl"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "type.#", "0"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "cache_ttl", "72000"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "address", ""),
				),
			},
			// Local lists only, so the apply doesn't download anything.
			{
				Config: testAccUnboundDnsblConfig(true, `["192.0.2.0/24", "198.51.100.0/24"]`, "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "blocklists.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_unbound_dnsbl.test", "source_nets.*", "198.51.100.0/24"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "address", "192.0.2.1"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "cache_ttl", "3600"),
					resource.TestCheckResourceAttrPair("data.opnsense_unbound_dnsbl.test", "blocklists.#", "opnsense_unbound_dnsbl.test", "blocklists.#"),
					resource.TestCheckResourceAttrPair("data.opnsense_unbound_dnsbl.test", "description", "opnsense_unbound_dnsbl.test", "description"),
				),
			},
			{
				ResourceName:      "opnsense_unbound_dnsbl.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Disabled, so selecting predefined lists doesn't download them.
			{
				Config: testAccUnboundDnsblConfig(false, "[]", `["hgz002", "oisd0"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "source_nets.#", "0"),
					resource.TestCheckTypeSetElemAttr("opnsense_unbound_dnsbl.test", "type.*", "hgz002"),
					resource.TestCheckResourceAttr("opnsense_unbound_dnsbl.test", "nxdomain", "true"),
				),
			},
		},
	})
}

func testAccUnboundDnsblConfig(enabled bool, sourceNets, types string) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_dnsbl" "test" {
  enabled     = %[1]t
  type        = %[3]s
  blocklists  = ["bad.example.com", "worse.example.org"]
  allowlists  = ["good.example.com"]
  wildcards   = ["ads.example.net"]
  source_nets = %[2]s
  address     = "192.0.2.1"
  nxdomain    = %[4]t
  cache_ttl   = 3600
  description = "tf-acc dnsbl"
}

data "opnsense_unbound_dnsbl" "test" {
  id = opnsense_unbound_dnsbl.test.id
}
`, enabled, sourceNets, types, !enabled)
}
