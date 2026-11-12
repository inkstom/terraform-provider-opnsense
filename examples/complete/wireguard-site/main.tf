# Site-to-site WireGuard on OPNsense, end to end:
# keys -> instance -> interface assignment -> gateway -> policy route.
# Requires OPNsense 26.7.7+ for interface assignments.

terraform {
  required_providers {
    opnsense = {
      source = "browningluke/opnsense"
    }
  }
}

provider "opnsense" {}

variable "remote_public_key" {
  type = string
}

variable "remote_endpoint" {
  type    = string
  default = "vpn.example.net"
}

resource "opnsense_wireguard_key_pair" "site" {}
resource "opnsense_wireguard_psk" "site" {}

resource "opnsense_wireguard_client" "remote" {
  name           = "remote-site"
  public_key     = var.remote_public_key
  psk            = opnsense_wireguard_psk.site.psk
  server_address = var.remote_endpoint
  server_port    = 51820
  tunnel_address = ["10.99.0.2/32", "172.16.0.0/16"]
  keep_alive     = 25
}

resource "opnsense_wireguard_server" "site" {
  name           = "site"
  public_key     = opnsense_wireguard_key_pair.site.public_key
  private_key    = opnsense_wireguard_key_pair.site.private_key
  port           = 51820
  tunnel_address = ["10.99.0.1/24"]
  peers          = [opnsense_wireguard_client.remote.id]

  # Route via an explicit gateway instead of auto-installed routes
  disable_routes = true
  gateway        = "10.99.0.2"
}

resource "opnsense_interfaces_assignment" "wg" {
  device            = opnsense_wireguard_server.site.device # "wgN"
  description       = "WG_SITE"
  gateway_interface = true
}

resource "opnsense_routing_gateway" "wg" {
  name            = "WG_SITE_GW"
  interface       = opnsense_interfaces_assignment.wg.id
  gateway         = opnsense_wireguard_server.site.gateway
  monitor_disable = false
  monitor         = "10.99.0.2"
}

resource "opnsense_route" "remote_lan" {
  network = "172.16.0.0/16"
  gateway = opnsense_routing_gateway.wg.name
}

output "public_key_for_remote" {
  value = opnsense_wireguard_key_pair.site.public_key
}

output "psk_for_remote" {
  value     = opnsense_wireguard_psk.site.psk
  sensitive = true
}
