# Resolve the corporate domain through the VPN's DNS server.
resource "opnsense_dnsmasq_domain" "corp" {
  domain      = "corp.example.com"
  server      = "10.20.0.53"
  description = "Corporate DNS over the site-to-site VPN"
}
