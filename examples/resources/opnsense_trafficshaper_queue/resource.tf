resource "opnsense_trafficshaper_pipe" "wan_up" {
  bandwidth   = 20
  description = "WAN upload"
}

# Give VoIP a larger share of the upload pipe than bulk traffic.
resource "opnsense_trafficshaper_queue" "voip" {
  pipe        = opnsense_trafficshaper_pipe.wan_up.id
  weight      = 80
  description = "VoIP"
}

resource "opnsense_trafficshaper_queue" "bulk" {
  pipe        = opnsense_trafficshaper_pipe.wan_up.id
  weight      = 20
  description = "Bulk"
}
