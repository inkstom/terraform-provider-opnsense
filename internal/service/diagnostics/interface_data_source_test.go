package diagnostics_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Read-only: every box has at least one interface. The single-interface data
// source is looked up by the device the list reports first.
func TestAccInterfaceDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
data "opnsense_interface_all" "test" {}

data "opnsense_interface" "test" {
  device = data.opnsense_interface_all.test.interfaces[0].device
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.opnsense_interface_all.test", "interfaces.0.device"),
				resource.TestCheckResourceAttrPair("data.opnsense_interface.test", "device", "data.opnsense_interface_all.test", "interfaces.0.device"),
				resource.TestCheckResourceAttrPair("data.opnsense_interface.test", "macaddr", "data.opnsense_interface_all.test", "interfaces.0.macaddr"),
				resource.TestCheckResourceAttrPair("data.opnsense_interface.test", "mtu", "data.opnsense_interface_all.test", "interfaces.0.mtu"),
				resource.TestCheckResourceAttrPair("data.opnsense_interface.test", "is_physical", "data.opnsense_interface_all.test", "interfaces.0.is_physical"),
			),
		}},
	})
}
