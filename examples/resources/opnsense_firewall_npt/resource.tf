# Translate the internal ULA prefix to the provider-assigned prefix on WAN.
resource "opnsense_firewall_npt" "lan" {
  interface       = "wan"
  internal_prefix = "fd00:10::/64"
  external_prefix = "2001:db8:10::/64"
  description     = "LAN NPTv6"
}
