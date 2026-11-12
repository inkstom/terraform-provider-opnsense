resource "opnsense_interfaces_assignment" "lan_port2" {
  device      = "igc2"
  description = "LAN2"
}

resource "opnsense_interfaces_assignment" "lan_port3" {
  device      = "igc3"
  description = "LAN3"
}

# Bridge two spare ports into one switched segment.
resource "opnsense_interfaces_bridge" "lan" {
  members        = [opnsense_interfaces_assignment.lan_port2.id, opnsense_interfaces_assignment.lan_port3.id]
  stp_enabled    = true
  stp_interfaces = [opnsense_interfaces_assignment.lan_port2.id, opnsense_interfaces_assignment.lan_port3.id]
  description    = "LAN switch ports"
}
