resource "opnsense_interfaces_vxlan" "dc" {
  vni            = "100"
  local_address  = "198.51.100.10"
  remote_address = "198.51.100.20"
}
