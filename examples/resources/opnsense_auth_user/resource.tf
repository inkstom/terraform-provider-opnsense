variable "alice_password" {
  type      = string
  sensitive = true
}

resource "opnsense_auth_group" "helpdesk" {
  name       = "helpdesk"
  privileges = ["page-dashboard-all", "page-diagnostics-arptable"]
}

resource "opnsense_auth_user" "alice" {
  name              = "alice"
  password          = var.alice_password
  full_name         = "Alice Example"
  email             = "alice@example.com"
  group_memberships = [opnsense_auth_group.helpdesk.gid]
}

# An account for SSH keys only; OPNsense sets a random password.
resource "opnsense_auth_user" "backup" {
  name            = "backup"
  shell           = "/bin/sh"
  authorized_keys = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleOnlyNotARealKey backup@example"
}
