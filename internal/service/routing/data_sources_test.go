package routing_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func pairs(typ string, attrs ...string) []resource.TestCheckFunc {
	r, d := typ+".test", "data."+typ+".test"
	out := []resource.TestCheckFunc{resource.TestCheckResourceAttrPair(d, "id", r, "id")}
	for _, a := range attrs {
		out = append(out, resource.TestCheckResourceAttrPair(d, a, r, a))
	}
	return out
}

func TestAccRoutingGatewayDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "opnsense_routing_gateway" "test" {
  name        = "TFACC_DSGW"
  interface   = "wan"
  gateway     = "203.0.113.21"
  far_gateway = true
  priority    = 210
  description = "tfacc ds"
}

data "opnsense_routing_gateway" "test" {
  id = opnsense_routing_gateway.test.id
}`,
			Check: resource.ComposeAggregateTestCheckFunc(pairs("opnsense_routing_gateway",
				"name", "interface", "gateway", "far_gateway", "priority", "ip_protocol", "description", "enabled")...),
		}},
	})
}

func TestAccRoutingGatewayGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.MinVersionPreCheck(t, "26.7") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "opnsense_routing_gateway" "a" {
  name        = "TFACC_DSGGW1"
  interface   = "wan"
  gateway     = "203.0.113.22"
  far_gateway = true
}

resource "opnsense_routing_gateway" "b" {
  name        = "TFACC_DSGGW2"
  interface   = "wan"
  gateway     = "203.0.113.23"
  far_gateway = true
}

resource "opnsense_routing_gateway_group" "test" {
  name        = "TFACC_DSGROUP"
  tiers       = [[opnsense_routing_gateway.a.name], [opnsense_routing_gateway.b.name]]
  description = "tfacc ds"
}

data "opnsense_routing_gateway_group" "test" {
  id = opnsense_routing_gateway_group.test.id
}`,
			Check: resource.ComposeAggregateTestCheckFunc(pairs("opnsense_routing_gateway_group",
				"name", "tiers.#", "tiers.0.0", "tiers.1.0", "trigger", "pool_options", "description")...),
		}},
	})
}
