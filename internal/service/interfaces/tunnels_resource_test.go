package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Everything here works on the single-NIC test VM: tunnels use wan as the
// local endpoint and documentation addresses as the remote end.
func TestAccInterfacesTunnelsResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInterfacesTunnelsConfig("first", "192.0.2.1", "100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_loopback.test", "description", "tf-acc first"),
					resource.TestMatchResourceAttr("opnsense_interfaces_loopback.test", "device", regexpLoopback),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "local_address", "wan"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "remote_address", "192.0.2.1"),
					resource.TestMatchResourceAttr("opnsense_interfaces_gif.test", "device", regexpGif),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "no_clamp", "false"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "tunnel_remote_net", "30"),
					resource.TestMatchResourceAttr("opnsense_interfaces_gre.test", "device", regexpGre),
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "vni", "100"),
					resource.TestMatchResourceAttr("opnsense_interfaces_vxlan.test", "device", regexpVxlan),
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "mac_address", "00:11:22:33:44:55"),
					resource.TestCheckResourceAttrPair("data.opnsense_interfaces_gif.test", "device", "opnsense_interfaces_gif.test", "device"),
					resource.TestCheckResourceAttrPair("data.opnsense_interfaces_vxlan.test", "device", "opnsense_interfaces_vxlan.test", "device"),
				),
			},
			{ResourceName: "opnsense_interfaces_loopback.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_interfaces_gif.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_interfaces_gre.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_interfaces_vxlan.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_interfaces_neighbor.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccInterfacesTunnelsConfig("second", "198.51.100.1", "200"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_loopback.test", "description", "tf-acc second"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gif.test", "remote_address", "198.51.100.1"),
					resource.TestCheckResourceAttr("opnsense_interfaces_gre.test", "remote_address", "198.51.100.1"),
					resource.TestCheckResourceAttr("opnsense_interfaces_vxlan.test", "vni", "200"),
					resource.TestCheckResourceAttr("opnsense_interfaces_neighbor.test", "description", "tf-acc second"),
				),
			},
		},
	})
}

func testAccInterfacesTunnelsConfig(desc, remote, vni string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_loopback" "test" {
  description = "tf-acc %[1]s"
}

resource "opnsense_interfaces_gif" "test" {
  local_address         = "wan"
  remote_address        = %[2]q
  tunnel_local_address  = "10.99.0.1"
  tunnel_remote_address = "10.99.0.2"
  tunnel_remote_net     = "30"
  description           = "tf-acc %[1]s"
}

resource "opnsense_interfaces_gre" "test" {
  local_address         = "wan"
  remote_address        = %[2]q
  tunnel_local_address  = "10.98.0.1"
  tunnel_remote_address = "10.98.0.2"
  tunnel_remote_net     = "30"
  description           = "tf-acc %[1]s"
}

resource "opnsense_interfaces_vxlan" "test" {
  vni            = %[3]q
  local_address  = "10.0.2.15"
  remote_address = %[2]q
}

resource "opnsense_interfaces_neighbor" "test" {
  mac_address = "00:11:22:33:44:55"
  ip_address  = "10.0.2.200"
  description = "tf-acc %[1]s"
}

data "opnsense_interfaces_gif" "test" {
  id = opnsense_interfaces_gif.test.id
}

data "opnsense_interfaces_vxlan" "test" {
  id = opnsense_interfaces_vxlan.test.id
}
`, desc, remote, vni)
}
