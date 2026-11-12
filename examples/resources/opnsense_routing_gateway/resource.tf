# Static gateway on a second WAN, used as a default-route candidate
resource "opnsense_routing_gateway" "wan2" {
  name        = "WAN2_GW"
  description = "Backup uplink"
  interface   = "opt2"
  gateway     = "198.51.100.1"
  upstream    = true
  priority    = 254

  # Enable dpinger monitoring
  monitor_disable = false
  monitor         = "1.1.1.1"
  latency_low     = 200
  latency_high    = 500
  loss_low        = 10
  loss_high       = 20
}

# Static route through it
resource "opnsense_route" "backup_mgmt" {
  network = "192.0.2.0/24"
  gateway = opnsense_routing_gateway.wan2.name
}
