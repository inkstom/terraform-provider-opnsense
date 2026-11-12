resource "opnsense_interfaces_gre" "branch" {
  local_address         = "wan"
  remote_address        = "203.0.113.20"
  tunnel_local_address  = "10.98.0.1"
  tunnel_remote_address = "10.98.0.2"
  tunnel_remote_net     = "30"
  description           = "GRE to branch office"
}
