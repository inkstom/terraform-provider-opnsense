resource "opnsense_dnsmasq_option" "mgmt_dns" {
  interface   = "opt10"
  option      = "6" # dns-server
  value       = "192.168.99.1"
  description = "DNS for the management VLAN"
}
