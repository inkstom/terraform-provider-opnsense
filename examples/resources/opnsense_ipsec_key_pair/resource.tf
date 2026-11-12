resource "tls_private_key" "ipsec" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "opnsense_ipsec_key_pair" "local" {
  name        = "site-a"
  key_type    = "ecdsa"
  public_key  = tls_private_key.ipsec.public_key_pem
  private_key = tls_private_key.ipsec.private_key_pem_pkcs8
}

# The peer's public key only.
resource "opnsense_ipsec_key_pair" "peer" {
  name       = "site-b"
  key_type   = "ecdsa"
  public_key = file("site-b.pub.pem")
}
