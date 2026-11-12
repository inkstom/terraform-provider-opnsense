resource "opnsense_dhcrelay_destination" "dc" {
  name    = "domain-controllers"
  servers = ["192.168.10.5", "192.168.10.6"]
}
