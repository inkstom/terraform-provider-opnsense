import {
  to = opnsense_kea_control_agent.this
  id = "kea_control_agent"
}

# Let the HA peer reach this firewall's control agent over the sync link.
resource "opnsense_kea_control_agent" "this" {
  enabled   = true
  http_host = "10.255.255.1"
  http_port = 8000
}
