import {
  to = opnsense_ipsec_settings.this
  id = "ipsec_settings"
}

resource "opnsense_ipsec_settings" "this" {
  enabled              = true
  passthrough_networks = ["192.168.10.0/24"] # LAN traffic to itself stays out of the tunnels
}
