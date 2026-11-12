# Failover: use WAN_DHCP; fall back to WAN2_GW when it's down or lossy
resource "opnsense_routing_gateway_group" "failover" {
  name    = "WAN_FAILOVER"
  trigger = "downloss"
  tiers = [
    ["WAN_DHCP"],
    [opnsense_routing_gateway.wan2.name],
  ]
}

# Load balancing: both gateways in the same tier
resource "opnsense_routing_gateway_group" "balanced" {
  name         = "WAN_LB"
  pool_options = "round-robin sticky-address"
  tiers        = [["WAN_DHCP", opnsense_routing_gateway.wan2.name]]
}
