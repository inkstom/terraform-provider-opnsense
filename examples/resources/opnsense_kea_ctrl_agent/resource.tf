import {
  to = opnsense_kea_ctrl_agent.this
  id = "kea_ctrl_agent"
}

# Let the HA peer reach this firewall's control agent over the sync link.
resource "opnsense_kea_ctrl_agent" "this" {
  enabled   = true
  http_host = "10.255.255.1"
  http_port = 8000
}
