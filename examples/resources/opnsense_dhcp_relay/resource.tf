resource "opnsense_dhcp_relay_destination" "dc" {
  name    = "domain-controllers"
  servers = ["192.168.10.5"]
}

# Relay DHCP from the guest VLAN (opt2) to the domain controllers.
resource "opnsense_dhcp_relay" "guest" {
  interface   = "opt2"
  destination = opnsense_dhcp_relay_destination.dc.id
  agent_info  = true
}
