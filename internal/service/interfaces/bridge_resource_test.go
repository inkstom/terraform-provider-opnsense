package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Bridge members must be assigned interfaces. The test VM has one NIC, so the
// test bridges two assigned VLANs; interface assignment needs OPNsense 26.7.7.
func TestAccInterfacesBridgeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.MinVersionPreCheck(t, "26.7.7") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInterfacesBridgeConfig(true, "[opnsense_interfaces_assignment.br_a.id]", "[opnsense_interfaces_assignment.br_b.id]", "tf-acc bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "members.#", "2"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "stp_enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "stp_protocol", "rstp"),
					resource.TestCheckResourceAttrPair("opnsense_interfaces_bridge.test", "stp_interfaces.0", "opnsense_interfaces_assignment.br_a", "id"),
					resource.TestCheckResourceAttrPair("opnsense_interfaces_bridge.test", "edge_ports.0", "opnsense_interfaces_assignment.br_b", "id"),
					resource.TestMatchResourceAttr("opnsense_interfaces_bridge.test", "device", regexpBridge),
					resource.TestCheckNoResourceAttr("opnsense_interfaces_bridge.test", "max_age"),
					resource.TestCheckResourceAttrPair("data.opnsense_interfaces_bridge.test", "device", "opnsense_interfaces_bridge.test", "device"),
				),
			},
			{ResourceName: "opnsense_interfaces_bridge.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccInterfacesBridgeConfig(false, "[]", "[]", "tf-acc bridge 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "stp_enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "stp_interfaces.#", "0"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "edge_ports.#", "0"),
					resource.TestCheckResourceAttr("opnsense_interfaces_bridge.test", "description", "tf-acc bridge 2"),
				),
			},
		},
	})
}

func testAccInterfacesBridgeConfig(stp bool, stpIfs, edge, desc string) string {
	return fmt.Sprintf(`
resource "opnsense_interfaces_vlan" "br_a" {
  description = "tfacc bridge a"
  tag         = 3908
  priority    = 0
  parent      = "vtnet0"
  device      = "vlan0.3908"
}

resource "opnsense_interfaces_vlan" "br_b" {
  description = "tfacc bridge b"
  tag         = 3909
  priority    = 0
  parent      = "vtnet0"
  device      = "vlan0.3909"
}

resource "opnsense_interfaces_assignment" "br_a" {
  device      = opnsense_interfaces_vlan.br_a.device
  description = "TFACCBRA"
}

resource "opnsense_interfaces_assignment" "br_b" {
  device      = opnsense_interfaces_vlan.br_b.device
  description = "TFACCBRB"
}

resource "opnsense_interfaces_bridge" "test" {
  members        = [opnsense_interfaces_assignment.br_a.id, opnsense_interfaces_assignment.br_b.id]
  stp_enabled    = %[1]t
  stp_interfaces = %[2]s
  edge_ports     = %[3]s
  description    = %[4]q
}

data "opnsense_interfaces_bridge" "test" {
  id = opnsense_interfaces_bridge.test.id
}
`, stp, stpIfs, edge, desc)
}
