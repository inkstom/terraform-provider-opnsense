package wireguard_test

import (
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var base64Key = regexp.MustCompile(`^[A-Za-z0-9+/]{43}=$`)

func TestAccWireguardKeysResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_wireguard_key_pair" "test" {}
resource "opnsense_wireguard_psk" "test" {}

resource "opnsense_wireguard_server" "test" {
  name        = "tfacc-keys"
  public_key  = opnsense_wireguard_key_pair.test.public_key
  private_key = opnsense_wireguard_key_pair.test.private_key
}

resource "opnsense_wireguard_client" "test" {
  name           = "tfacc-keys-peer"
  public_key     = opnsense_wireguard_key_pair.test.public_key
  psk            = opnsense_wireguard_psk.test.psk
  tunnel_address = ["10.0.9.2/32"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("opnsense_wireguard_key_pair.test", "public_key", base64Key),
					resource.TestMatchResourceAttr("opnsense_wireguard_key_pair.test", "private_key", base64Key),
					resource.TestMatchResourceAttr("opnsense_wireguard_psk.test", "psk", base64Key),
					resource.TestMatchResourceAttr("opnsense_wireguard_server.test", "device", regexp.MustCompile(`^wg[0-9]+$`)),
				),
			},
		},
	})
}
