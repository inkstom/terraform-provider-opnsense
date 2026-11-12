resource "opnsense_dnsmasq_tag" "pxe" {
  tag = "pxe"
}

# Hand clients tagged "pxe" the iPXE loader from the TFTP server.
resource "opnsense_dnsmasq_boot" "ipxe" {
  filename       = "ipxe.efi"
  server_address = "192.168.10.5"
  tags           = [opnsense_dnsmasq_tag.pxe.id]
  description    = "iPXE"
}
