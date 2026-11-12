# Cap guest traffic at 50 Mbit/s, shared fairly between guests.
resource "opnsense_trafficshaper_pipe" "guest" {
  bandwidth   = 50
  scheduler   = "fq_codel"
  description = "Guest download"
}
