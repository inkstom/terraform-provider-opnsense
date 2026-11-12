resource "opnsense_trafficshaper_pipe" "guest" {
  bandwidth   = 50
  scheduler   = "fq_codel"
  description = "Guest download"
}

# Shape everything sent to the guest network.
resource "opnsense_trafficshaper_rule" "guest_down" {
  interface   = "wan"
  destination = ["192.168.50.0/24"]
  direction   = "in"
  target      = opnsense_trafficshaper_pipe.guest.id
  description = "Guest download"
}
