package kea_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The test VM's agent is disabled on 127.0.0.1:8000; the last step puts that back.
func TestAccKeaCtrlAgentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_kea_ctrl_agent" "test" {}`,
				ResourceName:       "opnsense_kea_ctrl_agent.test",
				ImportState:        true,
				ImportStateId:      "kea_ctrl_agent",
				ImportStatePersist: true,
			},
			{
				Config: testAccKeaCtrlAgentConfig(8001),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_kea_ctrl_agent.test", "http_port", "8001"),
					// Unset attributes keep the firewall's value.
					resource.TestCheckResourceAttr("opnsense_kea_ctrl_agent.test", "http_host", "127.0.0.1"),
					resource.TestCheckResourceAttr("opnsense_kea_ctrl_agent.test", "enabled", "false"),
					resource.TestCheckResourceAttr("data.opnsense_kea_ctrl_agent.test", "http_port", "8001"),
				),
			},
			{
				Config: testAccKeaCtrlAgentConfig(8000),
				Check:  resource.TestCheckResourceAttr("opnsense_kea_ctrl_agent.test", "http_port", "8000"),
			},
		},
	})
}

func TestAccKeaCtrlAgentResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "opnsense_kea_ctrl_agent" "test" {}`,
				ExpectError: regexp.MustCompile("Import Required"),
			},
		},
	})
}

func testAccKeaCtrlAgentConfig(port int) string {
	return fmt.Sprintf(`
resource "opnsense_kea_ctrl_agent" "test" {
  http_port = %d
}

data "opnsense_kea_ctrl_agent" "test" {
  depends_on = [opnsense_kea_ctrl_agent.test]
}
`, port)
}
