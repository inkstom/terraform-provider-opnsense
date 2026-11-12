package dhcrelay_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDhcrelayResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDhcrelayConfig(`["192.0.2.10"]`, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dhcrelay_destination.test", "name", "tf-acc-dest"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_destination.test", "servers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_relay.test", "interface", "wan"),
					resource.TestCheckResourceAttrPair("opnsense_dhcrelay_relay.test", "destination", "opnsense_dhcrelay_destination.test", "id"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_relay.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_relay.test", "agent_info", "false"),
					resource.TestCheckResourceAttr("data.opnsense_dhcrelay_destination.test", "servers.#", "1"),
					resource.TestCheckResourceAttr("data.opnsense_dhcrelay_relay.test", "interface", "wan"),
				),
			},
			{
				ResourceName:      "opnsense_dhcrelay_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "opnsense_dhcrelay_relay.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDhcrelayConfig(`["192.0.2.10", "192.0.2.11"]`, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dhcrelay_destination.test", "servers.#", "2"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_relay.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dhcrelay_relay.test", "agent_info", "true"),
				),
			},
		},
	})
}

func testAccDhcrelayConfig(servers string, enabled, agentInfo bool) string {
	return fmt.Sprintf(`
resource "opnsense_dhcrelay_destination" "test" {
  name    = "tf-acc-dest"
  servers = %[1]s
}

resource "opnsense_dhcrelay_relay" "test" {
  enabled     = %[2]t
  interface   = "wan"
  destination = opnsense_dhcrelay_destination.test.id
  agent_info  = %[3]t
}

data "opnsense_dhcrelay_destination" "test" {
  id = opnsense_dhcrelay_destination.test.id
}

data "opnsense_dhcrelay_relay" "test" {
  id = opnsense_dhcrelay_relay.test.id
}
`, servers, enabled, agentInfo)
}
