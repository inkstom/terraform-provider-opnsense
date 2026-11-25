# Block ads, trackers and threats with HaGeZi Multi NORMAL for every client.
resource "opnsense_unbound_dnsbl" "example" {
  type        = ["hgz002"]
  allowlists  = ["example\\.com"]
  description = "HaGeZi Multi NORMAL"
}
