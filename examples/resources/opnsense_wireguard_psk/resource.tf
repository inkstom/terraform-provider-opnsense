resource "opnsense_wireguard_psk" "laptop" {}

resource "opnsense_wireguard_client" "laptop" {
  name           = "laptop"
  public_key     = var.laptop_public_key
  psk            = opnsense_wireguard_psk.laptop.psk
  tunnel_address = ["10.99.0.2/32"]
}
