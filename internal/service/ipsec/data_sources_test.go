package ipsec_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Every IPsec data source read by ID must match its resource attribute for attribute.
func TestAccIpsecDataSources(t *testing.T) {
	pairs := map[string][]string{
		"ipsec_connection":  {"enabled", "proposals.#", "version", "local_addresses.#", "remote_addresses.#", "rekey_time", "description"},
		"ipsec_auth_local":  {"ipsec_connection", "round", "authentication", "auth_id", "description"},
		"ipsec_auth_remote": {"ipsec_connection", "round", "authentication", "auth_id", "description"},
		"ipsec_child":       {"ipsec_connection", "mode", "local_networks.#", "remote_networks.#", "request_id", "description"},
		"ipsec_psk":         {"identity_local", "identity_remote", "pre_shared_key", "type", "description"},
		"ipsec_vti":         {"local_ip", "remote_ip", "tunnel_local_ip", "tunnel_remote_ip", "request_id"},
	}
	var checks []resource.TestCheckFunc
	for typ, attrs := range pairs {
		r, d := "opnsense_"+typ+".test", "data.opnsense_"+typ+".test"
		checks = append(checks, resource.TestCheckResourceAttrPair(d, "id", r, "id"))
		for _, a := range attrs {
			checks = append(checks, resource.TestCheckResourceAttrPair(d, a, r, a))
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccIpsecDataSourcesConfig,
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		}},
	})
}

const testAccIpsecDataSourcesConfig = `
resource "opnsense_ipsec_connection" "test" {
  enabled                  = "0"
  proposals                = ["aes128-sha256-modp2048"]
  unique                   = "no"
  aggressive               = "0"
  version                  = "2"
  mobike                   = "1"
  local_addresses          = ["192.0.2.1"]
  remote_addresses         = ["198.51.100.1"]
  local_port               = ""
  remote_port              = ""
  udp_encapsulation        = "0"
  reauthentication_time    = "3600"
  rekey_time               = "1800"
  ike_lifetime             = "3600"
  dpd_delay                = "10"
  dpd_timeout              = "60"
  send_certificate_request = "1"
  send_certificate         = "ifasked"
  keying_tries             = "1"
  description              = "tfacc ds connection"
}

resource "opnsense_ipsec_auth_local" "test" {
  enabled          = "1"
  ipsec_connection = opnsense_ipsec_connection.test.id
  round            = "0"
  authentication   = "psk"
  auth_id          = "local.example"
  description      = "tfacc ds local"
}

resource "opnsense_ipsec_auth_remote" "test" {
  enabled          = "1"
  ipsec_connection = opnsense_ipsec_connection.test.id
  round            = "0"
  authentication   = "psk"
  auth_id          = "remote.example"
  description      = "tfacc ds remote"
}

resource "opnsense_ipsec_child" "test" {
  enabled          = "1"
  ipsec_connection = opnsense_ipsec_connection.test.id
  proposals        = ["aes256-sha256-modp2048"]
  mode             = "tunnel"
  local_networks   = ["192.0.2.0/24"]
  remote_networks  = ["198.51.100.0/24"]
  request_id       = "4321"
  description      = "tfacc ds child"
}

resource "opnsense_ipsec_psk" "test" {
  identity_local  = "local.example"
  identity_remote = "remote.example"
  pre_shared_key  = "tfacc-data-source-psk"
  type            = "PSK"
  description     = "tfacc ds psk"
}

resource "opnsense_ipsec_vti" "test" {
  local_ip         = "192.0.2.1"
  remote_ip        = "198.51.100.1"
  tunnel_local_ip  = "10.255.0.1"
  tunnel_remote_ip = "10.255.0.2"
  request_id       = "4322"
}

data "opnsense_ipsec_connection" "test" {
  id = opnsense_ipsec_connection.test.id
}

data "opnsense_ipsec_auth_local" "test" {
  id = opnsense_ipsec_auth_local.test.id
}

data "opnsense_ipsec_auth_remote" "test" {
  id = opnsense_ipsec_auth_remote.test.id
}

data "opnsense_ipsec_child" "test" {
  id = opnsense_ipsec_child.test.id
}

data "opnsense_ipsec_psk" "test" {
  id = opnsense_ipsec_psk.test.id
}

data "opnsense_ipsec_vti" "test" {
  id = opnsense_ipsec_vti.test.id
}
`
