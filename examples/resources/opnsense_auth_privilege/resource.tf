resource "opnsense_auth_group" "helpdesk" {
  name = "helpdesk"
}

# Authoritative: only the helpdesk group holds this privilege.
resource "opnsense_auth_privilege" "arp_table" {
  privilege = "page-diagnostics-arptable"
  groups    = [opnsense_auth_group.helpdesk.id]
}
