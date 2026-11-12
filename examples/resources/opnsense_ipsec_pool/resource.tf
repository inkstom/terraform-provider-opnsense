# Hand road-warrior clients an address and DNS servers.
resource "opnsense_ipsec_pool" "roadwarrior" {
  name      = "roadwarrior"
  addresses = "10.200.0.0/24"
  dns       = ["10.0.0.1"]
}
