package firewall_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallCategoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallCategoryConfig("tfacc-category", "ff0000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "name", "tfacc-category"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "color", "ff0000"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "auto", "false"),
					resource.TestCheckResourceAttrPair("data.opnsense_firewall_category.test", "id", "opnsense_firewall_category.test", "id"),
					resource.TestCheckResourceAttrPair("data.opnsense_firewall_category.test", "name", "opnsense_firewall_category.test", "name"),
					resource.TestCheckResourceAttrPair("data.opnsense_firewall_category.test", "color", "opnsense_firewall_category.test", "color"),
					resource.TestCheckResourceAttrPair("data.opnsense_firewall_category.test", "auto", "opnsense_firewall_category.test", "auto"),
				),
			},
			{
				ResourceName:      "opnsense_firewall_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallCategoryConfig("tfacc-category-2", "00ff00"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "name", "tfacc-category-2"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "color", "00ff00"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_category.test", "color", "00ff00"),
				),
			},
		},
	})
}

func testAccFirewallCategoryConfig(name, color string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_category" "test" {
  name  = %q
  color = %q
}

data "opnsense_firewall_category" "test" {
  id = opnsense_firewall_category.test.id
}
`, name, color)
}
