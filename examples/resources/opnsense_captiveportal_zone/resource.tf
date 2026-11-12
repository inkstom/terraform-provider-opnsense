# Guest Wi-Fi portal: log in with a local account, idle out after an hour.
resource "opnsense_captiveportal_zone" "guest" {
  interfaces        = ["opt2"]
  auth_servers      = ["Local Database"]
  idle_timeout      = 60
  hard_timeout      = 480
  allowed_addresses = ["192.168.50.2"] # the guest network's printer
  description       = "Guest"
}
