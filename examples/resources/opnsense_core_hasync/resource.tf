import {
  to = opnsense_core_hasync.this
  id = "core_hasync"
}

# Sync states over a dedicated link and push configuration to the backup.
resource "opnsense_core_hasync" "this" {
  pfsync_interface  = "opt1"
  pfsync_peer_ip    = "10.255.255.2"
  synchronize_to_ip = "10.255.255.2"
  username          = "root"
  password          = var.backup_root_password
  sync_items        = ["aliases", "rules", "nat", "certs", "cron", "dhcpd", "ipsec", "wireguard"]
}

variable "backup_root_password" {
  type      = string
  sensitive = true
}
