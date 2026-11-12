# Also send 10.10.0.0/16 into the child SA with request ID 100.
resource "opnsense_ipsec_manual_spd" "extra" {
  request_id  = 100
  source      = "10.10.0.0/16"
  destination = "172.16.0.0/16"
  description = "Extra networks for site B"
}
