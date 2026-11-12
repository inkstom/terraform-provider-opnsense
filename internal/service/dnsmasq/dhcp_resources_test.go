package dnsmasq_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The CI VM only has `wan` (10.0.2.0/24 via QEMU user networking), so ranges
// and options are bound to wan with addresses from that subnet.

func TestAccDnsmasqRangeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_dnsmasq_range" "test" {
  interface     = "wan"
  start_address = "10.0.2.100"
  end_address   = "10.0.2.150"
  lease_time    = 4000
  description   = "tfacc dynamic range"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_dnsmasq_range.test", "id"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "interface", "wan"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "lease_time", "4000"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "domain_type", "range"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "mode", ""),
					resource.TestCheckNoResourceAttr("opnsense_dnsmasq_range.test", "prefix_length"),
				),
			},
			{
				ResourceName:      "opnsense_dnsmasq_range.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Reservation-only range, as used for infrastructure VLANs.
				Config: `
resource "opnsense_dnsmasq_range" "test" {
  interface     = "wan"
  start_address = "10.0.2.0"
  mode          = "static"
  domain_type   = "interface"
  domain        = "tfacc.internal"
  description   = "tfacc reservation-only"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "mode", "static"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "end_address", ""),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_range.test", "domain_type", "interface"),
					resource.TestCheckNoResourceAttr("opnsense_dnsmasq_range.test", "lease_time"),
				),
			},
		},
	})
}

func TestAccDnsmasqOptionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_dnsmasq_option" "test" {
  option      = "6"
  interface   = "wan"
  value       = "10.0.2.15"
  description = "tfacc dns option"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_dnsmasq_option.test", "id"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "type", "set"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "option", "6"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "force", "false"),
				),
			},
			{
				ResourceName:      "opnsense_dnsmasq_option.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
resource "opnsense_dnsmasq_option" "test" {
  option      = "42"
  interface   = "wan"
  value       = "10.0.2.15,10.0.2.2"
  force       = true
  description = "tfacc ntp option"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "option", "42"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "value", "10.0.2.15,10.0.2.2"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_option.test", "force", "true"),
				),
			},
		},
	})
}

const dnsmasqSettingsBase = `
resource "opnsense_dnsmasq_settings" "this" {
  enabled    = true
  interfaces = ["wan"]
  port       = 0
  %s
}`

func TestAccDnsmasqSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_dnsmasq_settings" "this" {}`,
				ResourceName:       "opnsense_dnsmasq_settings.this",
				ImportState:        true,
				ImportStateId:      "dnsmasq_settings",
				ImportStatePersist: true,
			},
			{
				Config: fmt.Sprintf(dnsmasqSettingsBase, `dhcp_log_dhcp = true
  cache_size    = 500`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "id", "dnsmasq_settings"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "port", "0"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "dhcp_log_dhcp", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "cache_size", "500"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "dhcp_host_ping", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "no_ident", "true"),
				),
			},
			{
				// Back to defaults; unset integers go back to "" on the firewall.
				Config: fmt.Sprintf(dnsmasqSettingsBase, `strict_interface_binding = false`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "strict_interface_binding", "false"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.this", "dhcp_log_dhcp", "false"),
					resource.TestCheckNoResourceAttr("opnsense_dnsmasq_settings.this", "cache_size"),
				),
			},
		},
	})
}

func TestAccDnsmasqSettingsResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      `resource "opnsense_dnsmasq_settings" "this" {}`,
			ExpectError: regexp.MustCompile("Import Required"),
		}},
	})
}
