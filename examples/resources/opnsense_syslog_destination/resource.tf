# Send firewall logs and errors to a central syslog server over TCP.
resource "opnsense_syslog_destination" "siem" {
  transport   = "tcp4"
  hostname    = "192.168.10.20"
  port        = "514"
  programs    = ["filterlog", "sshd", "audit"]
  levels      = ["info", "notice", "warn", "err", "crit", "alert", "emerg"]
  description = "SIEM"
}
