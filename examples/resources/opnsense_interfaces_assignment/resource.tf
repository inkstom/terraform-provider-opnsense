resource "opnsense_interfaces_vlan" "iot" {
  description = "IoT VLAN"
  tag         = 30
  parent      = "igc1"
}

# Assign the VLAN as an interface with a static address
resource "opnsense_interfaces_assignment" "iot" {
  device       = opnsense_interfaces_vlan.iot.device
  description  = "IOT"
  ipv4_type    = "staticv4"
  ipv4_address = "10.30.0.1/24"
}

# The identifier (e.g. "opt3") is what other resources expect as an interface
resource "opnsense_firewall_filter" "iot_dns" {
  interface = {
    interface = [opnsense_interfaces_assignment.iot.id]
  }
  filter = {
    action    = "pass"
    direction = "in"
    protocol  = "TCP/UDP"
    source = {
      net = "${opnsense_interfaces_assignment.iot.id}"
    }
    destination = {
      net  = "(self)"
      port = "53"
    }
  }
  description = "IoT may use firewall DNS"
}
