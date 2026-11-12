# Put the guest and IoT VLANs in one group so a single rule set covers both.
resource "opnsense_firewall_group" "untrusted" {
  name        = "untrusted"
  members     = ["opt2", "opt3"]
  description = "Guest and IoT"
}
