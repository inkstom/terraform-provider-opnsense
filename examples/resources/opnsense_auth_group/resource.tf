# A read-only helpdesk group, allowed to log in from the management network only.
resource "opnsense_auth_group" "helpdesk" {
  name        = "helpdesk"
  description = "Read-only diagnostics"
  privileges = [
    "page-dashboard-all",
    "page-diagnostics-arptable",
    "page-diagnostics-logs-system",
  ]
  source_networks = ["192.168.10.0/24"]
}
