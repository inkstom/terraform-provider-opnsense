variable "ddns_password" {
  type      = string
  sensitive = true
}

# A dyndns2-compatible provider through the custom service.
resource "opnsense_dyndns_account" "example" {
  service     = "custom"
  protocol    = "dyndns2"
  server      = "members.example-ddns.net"
  username    = "myuser"
  password    = var.ddns_password
  hostnames   = ["home.example-ddns.net"]
  checkip     = "web_icanhazip"
  description = "Home WAN address"
}
