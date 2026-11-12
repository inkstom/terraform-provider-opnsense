package core_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCoreTunableResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreTunableConfig("2", "drop RST to closed ports"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_core_tunable.test", "tunable", "net.inet.tcp.blackhole"),
					resource.TestCheckResourceAttr("opnsense_core_tunable.test", "value", "2"),
					resource.TestCheckResourceAttr("opnsense_core_tunable.test", "description", "drop RST to closed ports"),
					resource.TestCheckResourceAttrSet("opnsense_core_tunable.test", "id"),
					resource.TestCheckResourceAttr("data.opnsense_core_tunable.test", "value", "2"),
				),
			},
			{
				ResourceName:      "opnsense_core_tunable.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreTunableConfig("1", "changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_core_tunable.test", "value", "1"),
					resource.TestCheckResourceAttr("opnsense_core_tunable.test", "description", "changed"),
				),
			},
		},
	})
}

func testAccCoreTunableConfig(value, description string) string {
	return fmt.Sprintf(`
resource "opnsense_core_tunable" "test" {
  tunable     = "net.inet.tcp.blackhole"
  value       = %[1]q
  description = %[2]q
}

data "opnsense_core_tunable" "test" {
  id = opnsense_core_tunable.test.id
}
`, value, description)
}
