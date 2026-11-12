resource "opnsense_interfaces_gif" "he_ipv6" {
  local_address         = "wan"
  remote_address        = "203.0.113.1"
  tunnel_local_address  = "10.99.0.1"
  tunnel_remote_address = "10.99.0.2"
  tunnel_remote_net     = "30"
  description           = "IPv6 tunnel broker"
}
