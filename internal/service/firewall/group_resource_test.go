package firewall_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig("tfaccgrp", `["wan", "lo0"]`, true, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "name", "tfaccgrp"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "members.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_firewall_group.test", "members.*", "lo0"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "no_group", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "sequence", "5"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_group.test", "members.#", "2"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_group.test", "name", "tfaccgrp"),
				),
			},
			{
				ResourceName:      "opnsense_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallGroupConfig("tfaccgrpx", `["wan"]`, false, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "name", "tfaccgrpx"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "members.#", "1"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "no_group", "false"),
					resource.TestCheckResourceAttr("opnsense_firewall_group.test", "sequence", "0"),
				),
			},
		},
	})
}

func testAccFirewallGroupConfig(name, members string, noGroup bool, seq int) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_group" "test" {
  name        = %[1]q
  members     = %[2]s
  no_group    = %[3]t
  sequence    = %[4]d
  description = "tf acc"
}

data "opnsense_firewall_group" "test" {
  id = opnsense_firewall_group.test.id
}
`, name, members, noGroup, seq)
}
