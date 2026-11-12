resource "opnsense_monit_alert" "ops" {
  recipient = "ops@example.com"
  events    = ["nonexist", "resource", "status"]
  reminder  = "10"
}
