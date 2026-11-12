package syslog_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSyslogDestinationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSyslogDestinationConfig("udp4", "514", `["err", "crit"]`, `["filterlog"]`, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "transport", "udp4"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "hostname", "192.0.2.50"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "levels.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_syslog_destination.test", "levels.*", "crit"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "programs.#", "1"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "facilities.#", "0"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "rfc5424", "false"),
					resource.TestCheckResourceAttr("data.opnsense_syslog_destination.test", "levels.#", "2"),
				),
			},
			{
				ResourceName:      "opnsense_syslog_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
				// suppress is unmanaged when unset; import reads whatever the box has.
				ImportStateVerifyIgnore: []string{"suppress"},
			},
			{
				Config: testAccSyslogDestinationConfig("tcp4", "1514", `[]`, `[]`, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "transport", "tcp4"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "port", "1514"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "levels.#", "0"),
					resource.TestCheckResourceAttr("opnsense_syslog_destination.test", "rfc5424", "true"),
					resource.TestCheckNoResourceAttr("opnsense_syslog_destination.test", "suppress"),
				),
			},
		},
	})
}

func testAccSyslogDestinationConfig(transport, port, levels, programs string, rfc bool) string {
	return fmt.Sprintf(`
resource "opnsense_syslog_destination" "test" {
  enabled     = false
  transport   = %[1]q
  hostname    = "192.0.2.50"
  port        = %[2]q
  levels      = %[3]s
  programs    = %[4]s
  rfc5424     = %[5]t
  description = "tf acc"
}

data "opnsense_syslog_destination" "test" {
  id = opnsense_syslog_destination.test.id
}
`, transport, port, levels, programs, rfc)
}
