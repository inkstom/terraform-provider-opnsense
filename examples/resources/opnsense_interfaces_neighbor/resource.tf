# Pin the printer's MAC to its address.
resource "opnsense_interfaces_neighbor" "printer" {
  mac_address = "00:11:22:33:44:55"
  ip_address  = "192.168.10.30"
  description = "Office printer"
}
