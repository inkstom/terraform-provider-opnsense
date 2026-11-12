# Dynamic pool
resource "opnsense_dnsmasq_range" "iot" {
  interface     = "opt7"
  start_address = "192.168.60.100"
  end_address   = "192.168.60.199"
  lease_time    = 4000
  description   = "IoT dynamic DHCP"
}

# Reservation-only: hand out addresses only to hosts with a static mapping
resource "opnsense_dnsmasq_range" "mgmt" {
  interface     = "opt10"
  start_address = "192.168.99.0"
  mode          = "static"
  description   = "Management reservation-only DHCP"
}
