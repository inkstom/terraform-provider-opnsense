import {
  to = opnsense_monit_settings.this
  id = "monit_settings"
}

# Turn Monit on and send alerts through the office relay.
resource "opnsense_monit_settings" "this" {
  enabled       = true
  interval      = 60
  mail_servers  = ["smtp.example.com"]
  mail_port     = 587
  mail_ssl      = true
  mail_username = "alerts@example.com"
  mail_password = var.smtp_password
}

variable "smtp_password" {
  type      = string
  sensitive = true
}
