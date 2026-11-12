package interfaces_test

import (
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Requires OPNsense >= 26.7.7. The CI VM has a single NIC (vtnet0), so a VLAN is assigned.
const testAccAssignmentVlan = `
resource "opnsense_interfaces_vlan" "assign" {
  description = "tfacc assignment"
  tag         = 3902
  priority    = 0
  parent      = "vtnet0"
  device      = "vlan0.3902"
}
`

func TestAccInterfacesAssignmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAssignmentVlan + `
resource "opnsense_interfaces_assignment" "test" {
  device       = opnsense_interfaces_vlan.assign.device
  description  = "TFACC"
  ipv4_type    = "staticv4"
  ipv4_address = "192.0.2.1/24"
}

data "opnsense_interfaces_assignment" "wan" {
  id = "wan"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("opnsense_interfaces_assignment.test", "id", regexpOpt),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("data.opnsense_interfaces_assignment.wan", "device"),
				),
			},
			{
				ResourceName:      "opnsense_interfaces_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssignmentVlan + `
resource "opnsense_interfaces_assignment" "test" {
  device       = opnsense_interfaces_vlan.assign.device
  description  = "TFACC2"
  ipv4_type    = "staticv4"
  ipv4_address = "192.0.2.1/25"
  lock         = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "description", "TFACC2"),
					resource.TestCheckResourceAttr("opnsense_interfaces_assignment.test", "lock", "true"),
				),
			},
		},
	})
}

var regexpOpt = regexp.MustCompile(`^opt[0-9]+$`)
