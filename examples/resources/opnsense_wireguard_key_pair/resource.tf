resource "opnsense_wireguard_key_pair" "site" {
  # Change the value to rotate the keys
  keepers = {
    rotation = "2026-11"
  }
}

resource "opnsense_wireguard_server" "site" {
  name           = "site"
  public_key     = opnsense_wireguard_key_pair.site.public_key
  private_key    = opnsense_wireguard_key_pair.site.private_key
  port           = 51820
  tunnel_address = ["10.99.0.1/24"]
}

output "site_public_key" {
  value = opnsense_wireguard_key_pair.site.public_key
}
