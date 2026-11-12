import {
  to = opnsense_syslog_settings.this
  id = "syslog_settings"
}

# Keep 90 days of logs locally, rotating files at 50 MB.
resource "opnsense_syslog_settings" "this" {
  max_preserve  = 90
  max_file_size = 52428800
}
