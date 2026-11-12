# Singleton: import it first, then manage it.
import {
  to = opnsense_dnsmasq_settings.this
  id = "dnsmasq_settings"
}

# DHCP-only dnsmasq next to Unbound: port 0 turns dnsmasq's DNS off.
resource "opnsense_dnsmasq_settings" "this" {
  enabled    = true
  interfaces = ["lan", "opt1", "opt10"]
  port       = 0
}
