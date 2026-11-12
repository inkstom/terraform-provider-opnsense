package wireguard_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func pairs(typ string, attrs ...string) []resource.TestCheckFunc {
	r, d := typ+".test", "data."+typ+".test"
	out := []resource.TestCheckFunc{resource.TestCheckResourceAttrPair(d, "id", r, "id")}
	for _, a := range attrs {
		out = append(out, resource.TestCheckResourceAttrPair(d, a, r, a))
	}
	return out
}

func TestAccWireguardDataSources(t *testing.T) {
	checks := pairs("opnsense_wireguard_server", "name", "enabled", "public_key", "port", "tunnel_address.#", "instance", "peers.#")
	checks = append(checks, pairs("opnsense_wireguard_client", "name", "enabled", "public_key", "tunnel_address.#", "server_address", "server_port")...)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "opnsense_wireguard_client" "test" {
  name           = "tfacc-ds-client"
  public_key     = "ng28gpVw6F9YYsXTdPtfglAsGU1ET+ePrTFDfOzGniw="
  tunnel_address = ["10.0.9.2/32"]
  server_address = "198.51.100.9"
  server_port    = 51821
}

resource "opnsense_wireguard_server" "test" {
  name           = "tfacc-ds-server"
  private_key    = "iD8W5IupKNnbt18tFt9PxpTLlBeBu+yxEUseLQT3OVQ="
  public_key     = "N6z94WfQPp2uLMs2uPA2CASRZBdPXUCJcZyrFufTkVE="
  tunnel_address = ["10.0.9.1/24"]
  port           = 51821
  peers          = [opnsense_wireguard_client.test.id]
}

data "opnsense_wireguard_server" "test" {
  id = opnsense_wireguard_server.test.id
}

data "opnsense_wireguard_client" "test" {
  id = opnsense_wireguard_client.test.id
}`,
			Check: resource.ComposeAggregateTestCheckFunc(checks...),
		}},
	})
}
