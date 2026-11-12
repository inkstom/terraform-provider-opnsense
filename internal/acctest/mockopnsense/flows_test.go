package mockopnsense_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/browningluke/terraform-provider-opnsense/internal/acctest/mockopnsense"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real Terraform plan/apply/import cycles against an in-memory imitation of the
// OPNsense API. They verify provider wiring, state handling and plan stability (the framework
// fails a step if a plan is non-empty after apply). They do not replace live acceptance tests.

func TestMockGatewayLifecycle(t *testing.T) {
	m := mockopnsense.Start(t, mockopnsense.GatewayRoute())

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_routing_gateway" "wan2" {
  name        = "WAN2_GW"
  interface   = "wan"
  gateway     = "203.0.113.1"
  far_gateway = true
  upstream    = true
  priority    = 100
  latency_low = 200
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "ip_protocol", "inet"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "monitor_disable", "true"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "latency_high", "-1"),
					resource.TestCheckResourceAttrSet("opnsense_routing_gateway.wan2", "id"),
					func(s *terraform.State) error {
						for _, it := range m.Items("/routing/settings/getGateway") {
							if it["defaultgw"] != "1" || it["fargw"] != "1" || it["priority"] != "100" ||
								it["latencylow"] != "200" || it["latencyhigh"] != "" || it["disabled"] != "0" {
								return fmt.Errorf("unexpected API payload: %v", it)
							}
						}
						return nil
					},
				),
			},
			{
				ResourceName:      "opnsense_routing_gateway.wan2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
resource "opnsense_routing_gateway" "wan2" {
  name        = "WAN2_GW"
  description = "backup uplink"
  interface   = "wan"
  gateway     = "203.0.113.1"
  far_gateway = true
  enabled     = false
}

data "opnsense_routing_gateway" "wan2" {
  id = opnsense_routing_gateway.wan2.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "upstream", "false"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wan2", "priority", "255"),
					resource.TestCheckResourceAttr("data.opnsense_routing_gateway.wan2", "description", "backup uplink"),
					resource.TestCheckResourceAttr("data.opnsense_routing_gateway.wan2", "enabled", "false"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			if n := len(m.Items("/routing/settings/getGateway")); n != 0 {
				return fmt.Errorf("%d gateways left after destroy", n)
			}
			return nil
		},
	})
}

func TestMockGatewayGroupTiers(t *testing.T) {
	m := mockopnsense.Start(t, mockopnsense.GatewayRoute(), mockopnsense.GatewayGroupRoute())

	gateways := `
resource "opnsense_routing_gateway" "a" {
  name        = "GW_A"
  interface   = "wan"
  gateway     = "203.0.113.1"
  far_gateway = true
}
resource "opnsense_routing_gateway" "b" {
  name        = "GW_B"
  interface   = "wan"
  gateway     = "203.0.113.2"
  far_gateway = true
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: gateways + `
resource "opnsense_routing_gateway_group" "failover" {
  name    = "WAN_FAILOVER"
  tiers   = [[opnsense_routing_gateway.a.name], [opnsense_routing_gateway.b.name]]
  trigger = "downloss"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.failover", "tiers.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_routing_gateway_group.failover", "tiers.0.*", "GW_A"),
					resource.TestCheckTypeSetElemAttr("opnsense_routing_gateway_group.failover", "tiers.1.*", "GW_B"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.failover", "pool_options", ""),
					func(s *terraform.State) error {
						for _, it := range m.Items("/routing/group_settings/get") {
							if it["item"] != "GW_A" || it["item2"] != "GW_B" || it["item3"] != "" || it["trigger"] != "downloss" {
								return fmt.Errorf("unexpected API payload: %v", it)
							}
						}
						return nil
					},
				),
			},
			{
				ResourceName:      "opnsense_routing_gateway_group.failover",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Collapse to a single load-balanced tier: tier 2 must be cleared upstream.
				Config: gateways + `
resource "opnsense_routing_gateway_group" "failover" {
  name         = "WAN_FAILOVER"
  tiers        = [[opnsense_routing_gateway.a.name, opnsense_routing_gateway.b.name]]
  pool_options = "round-robin"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.failover", "tiers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.failover", "tiers.0.#", "2"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway_group.failover", "trigger", "down"),
					func(s *terraform.State) error {
						for _, it := range m.Items("/routing/group_settings/get") {
							if it["item"] != "GW_A,GW_B" || it["item2"] != "" || it["poolopts"] != "round-robin" {
								return fmt.Errorf("unexpected API payload: %v", it)
							}
						}
						return nil
					},
				),
			},
		},
	})
}

func TestMockInterfaceAssignment(t *testing.T) {
	m := mockopnsense.Start(t, mockopnsense.AssignmentRoute())
	const get = "/interfaces/assignment/getItem"
	m.Seed(get, "wan", map[string]string{"identifier": "wan", "if": "vtnet0", "descr": "WAN", "type4": "dhcp"})
	m.Seed(get, "lan", map[string]string{"identifier": "lan", "if": "vtnet1", "descr": "LAN", "type4": "staticv4", "ipaddr": "192.168.1.1/24"})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_interfaces_assignment" "iot" {
  device       = "vlan0.30"
  description  = "IOT"
  ipv4_type    = "staticv4"
  ipv4_address = "10.30.0.1/24"
  block_bogons = true
  mtu          = 1500
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.iot", "id", "opt1"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.iot", "lock", "false"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.iot", "mss", "-1"),
					func(s *terraform.State) error {
						it := m.Items(get)["opt1"]
						if it["if"] != "vlan0.30" || it["lock"] != "0" || it["type4"] != "staticv4" || it["mtu"] != "1500" || it["mss"] != "" {
							return fmt.Errorf("unexpected API payload: %v", it)
						}
						return nil
					},
				),
			},
			{
				ResourceName:      "opnsense_interfaces_assignment.iot",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Lock it; destroy at the end must clear the lock before deleting.
				Config: `
resource "opnsense_interfaces_assignment" "iot" {
  device       = "vlan0.30"
  description  = "IOT_DEVICES"
  ipv4_type    = "staticv4"
  ipv4_address = "10.30.0.1/23"
  lock         = true
}

data "opnsense_interfaces_assignment" "lan" {
  id = "lan"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.iot", "lock", "true"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.iot", "block_bogons", "false"),
					resource.TestCheckResourceAttr("data.opnsense_interfaces_assignment.lan", "device", "vtnet1"),
					resource.TestCheckResourceAttr("data.opnsense_interfaces_assignment.lan", "ipv4_address", "192.168.1.1/24"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			items := m.Items(get)
			if _, ok := items["opt1"]; ok {
				return fmt.Errorf("opt1 still assigned after destroy")
			}
			if len(items) != 2 {
				return fmt.Errorf("expected only wan/lan to remain, got %v", items)
			}
			unlocked := false
			for i, c := range m.Calls {
				if c == "POST /interfaces/assignment/delItem/opt1" {
					for _, prev := range m.Calls[:i] {
						if prev == "POST /interfaces/assignment/setItem/opt1" {
							unlocked = true
						}
					}
				}
			}
			if !unlocked {
				return fmt.Errorf("locked interface was not unlocked before delete: %v", m.Calls)
			}
			return nil
		},
	})
}

func TestMockWireguardEndToEnd(t *testing.T) {
	m := mockopnsense.Start(t, mockopnsense.WireguardServerRoute(), mockopnsense.AssignmentRoute(),
		mockopnsense.GatewayRoute())

	config := func(keeper string) string {
		return fmt.Sprintf(`
resource "opnsense_wireguard_key_pair" "site" {
  keepers = { rotation = %q }
}

resource "opnsense_wireguard_psk" "peer" {}

resource "opnsense_wireguard_server" "site" {
  name           = "site"
  public_key     = opnsense_wireguard_key_pair.site.public_key
  private_key    = opnsense_wireguard_key_pair.site.private_key
  port           = 51820
  tunnel_address = ["10.99.0.1/24"]
  disable_routes = true
  gateway        = "10.99.0.2"
}

resource "opnsense_interfaces_assignment" "wg" {
  device            = opnsense_wireguard_server.site.device
  description       = "WG_SITE"
  gateway_interface = true
}

resource "opnsense_routing_gateway" "wg" {
  name      = "WG_SITE_GW"
  interface = opnsense_interfaces_assignment.wg.id
  gateway   = opnsense_wireguard_server.site.gateway
}
`, keeper)
	}

	var firstPub string
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("2026-11"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_wireguard_server.site", "device", "wg0"),
					resource.TestCheckResourceAttr("opnsense_wireguard_server.site", "debug", "false"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.wg", "device", "wg0"),
					resource.TestCheckResourceAttr("opnsense_routing_gateway.wg", "interface", "opt1"),
					resource.TestMatchResourceAttr("opnsense_wireguard_psk.peer", "psk", regexp.MustCompile(`^[A-Za-z0-9+/]{43}=$`)),
					resource.TestMatchResourceAttr("opnsense_wireguard_psk.peer", "id", regexp.MustCompile(`^psk-[0-9a-f]+$`)),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["opnsense_wireguard_key_pair.site"]
						firstPub = rs.Primary.Attributes["public_key"]
						for _, it := range m.Items("/wireguard/server/getServer") {
							if it["pubkey"] != firstPub || it["privkey"] == "" {
								return fmt.Errorf("server not wired to generated keys: %v", it)
							}
						}
						return nil
					},
				),
			},
			{
				// Rotating the keeper regenerates keys and updates the server in place.
				Config: config("2026-12"),
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["opnsense_wireguard_key_pair.site"]
					if rs.Primary.Attributes["public_key"] == firstPub {
						return fmt.Errorf("key pair was not regenerated")
					}
					if s.RootModule().Resources["opnsense_wireguard_server.site"].Primary.Attributes["device"] != "wg0" {
						return fmt.Errorf("server was replaced instead of updated")
					}
					return nil
				},
			},
			{
				ResourceName:            "opnsense_wireguard_server.site",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestMockValidation(t *testing.T) {
	mockopnsense.Start(t)
	for name, tc := range map[string]struct{ config, err string }{
		"empty tier": {`
resource "opnsense_routing_gateway_group" "g" {
  name  = "G"
  tiers = [[], ["GW_A"]]
}`, `(?s)tiers.*at least 1`},
		"too many tiers": {`
resource "opnsense_routing_gateway_group" "g" {
  name  = "G"
  tiers = [["a"], ["b"], ["c"], ["d"], ["e"], ["f"]]
}`, `(?s)tiers.*must contain at least 1 elements and at most 5`},
		"bad gateway name": {`
resource "opnsense_routing_gateway" "g" {
  name      = "has space"
  interface = "wan"
}`, `must match`},
		"static without address": {`
resource "opnsense_interfaces_assignment" "a" {
  device    = "vlan0.31"
  ipv4_type = "staticv4"
}`, `ipv4_address`},
		"track6 without parent": {`
resource "opnsense_interfaces_assignment" "a" {
  device    = "vlan0.31"
  ipv6_type = "track6"
}`, `track6_interface`},
		"bad description": {`
resource "opnsense_interfaces_assignment" "a" {
  device      = "vlan0.31"
  description = "IOT-NET"
}`, `letters, digits`},
	} {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      tc.config,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.err),
				}},
			})
		})
	}
}
