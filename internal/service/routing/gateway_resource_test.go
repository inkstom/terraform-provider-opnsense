package routing_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The CI VM only has a `wan` interface; far_gateway lets us use an address outside its subnet.

func testAccGatewayConfig(name, ip, extra string) string {
	return fmt.Sprintf(`
resource "opnsense_routing_gateway" %[1]q {
  name        = %[1]q
  interface   = "wan"
  gateway     = %[2]q
  far_gateway = true
  %[3]s
}
`, name, ip, extra)
}

func TestAccRoutingGatewayResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig("TFACC_GW1", "203.0.113.1", `priority = 200`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "priority", "200"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "ip_protocol", "inet"),
					resource.TestCheckResourceAttrSet("opnsense_routing_gateway.TFACC_GW1", "id"),
				),
			},
			{
				ResourceName:      "opnsense_routing_gateway.TFACC_GW1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayConfig("TFACC_GW1", "203.0.113.1", `
  description     = "updated"
  monitor_disable = false
  monitor         = "203.0.113.1"
  latency_low     = 200
  latency_high    = 500`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "description", "updated"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "monitor_disable", "false"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.TFACC_GW1", "latency_high", "500"),
				),
			},
		},
	})
}

func TestAccRoutingGatewayGroupResource(t *testing.T) {
	gateways := testAccGatewayConfig("TFACC_GGW1", "203.0.113.11", "") + testAccGatewayConfig("TFACC_GGW2", "203.0.113.12", "")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.MinVersionPreCheck(t, "26.7") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: gateways + `
resource "opnsense_routing_gateway_group" "test" {
  name  = "TFACC_GROUP"
  tiers = [[opnsense_routing_gateway.TFACC_GGW1.name], [opnsense_routing_gateway.TFACC_GGW2.name]]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "tiers.#", "2"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "trigger", "down"),
				),
			},
			{
				ResourceName:      "opnsense_routing_gateway_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: gateways + `
resource "opnsense_routing_gateway_group" "test" {
  name         = "TFACC_GROUP"
  tiers        = [[opnsense_routing_gateway.TFACC_GGW1.name, opnsense_routing_gateway.TFACC_GGW2.name]]
  trigger      = "downlosslatency"
  pool_options = "round-robin"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "tiers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.test", "tiers.0.#", "2"),
				),
			},
		},
	})
}
