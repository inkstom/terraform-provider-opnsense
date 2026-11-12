package monit_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMonitResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitConfig("75", "alert", `["resource", "status"]`, "every 2 cycles"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_monit_test.cpu", "condition", "cpu is greater than 75% for 3 cycles"),
					resource.TestCheckResourceAttr("opnsense_monit_test.cpu", "type", "ProcessResource"),
					resource.TestCheckResourceAttr("opnsense_monit_test.cpu", "action", "alert"),
					resource.TestCheckResourceAttr("opnsense_monit_service.unbound", "type", "process"),
					resource.TestCheckResourceAttr("opnsense_monit_service.unbound", "tests.#", "1"),
					resource.TestCheckResourceAttrPair("opnsense_monit_service.unbound", "tests.0", "opnsense_monit_test.cpu", "id"),
					resource.TestCheckResourceAttr("opnsense_monit_service.unbound", "poll_time", "every 2 cycles"),
					resource.TestCheckResourceAttr("opnsense_monit_alert.ops", "events.#", "2"),
					resource.TestCheckResourceAttr("data.opnsense_monit_service.unbound", "pidfile", "/var/run/unbound.pid"),
					resource.TestCheckResourceAttr("data.opnsense_monit_test.cpu", "action", "alert"),
					resource.TestCheckResourceAttr("data.opnsense_monit_alert.ops", "recipient", "ops@example.com"),
				),
			},
			{ResourceName: "opnsense_monit_test.cpu", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_monit_service.unbound", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_monit_alert.ops", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccMonitConfig("90", "restart", `[]`, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_monit_test.cpu", "condition", "cpu is greater than 90% for 3 cycles"),
					resource.TestCheckResourceAttr("opnsense_monit_test.cpu", "action", "restart"),
					resource.TestCheckResourceAttr("opnsense_monit_service.unbound", "poll_time", ""),
					resource.TestCheckResourceAttr("opnsense_monit_alert.ops", "events.#", "0"),
				),
			},
		},
	})
}

func testAccMonitConfig(cpu, action, events, poll string) string {
	return fmt.Sprintf(`
resource "opnsense_monit_test" "cpu" {
  name      = "tf-acc-cpu"
  condition = "cpu is greater than %[1]s%% for 3 cycles"
  action    = %[2]q
}

resource "opnsense_monit_service" "unbound" {
  enabled   = false
  name      = "tf_acc_unbound"
  type      = "process"
  pidfile   = "/var/run/unbound.pid"
  start     = "/usr/local/sbin/configctl unbound start"
  stop      = "/usr/local/sbin/configctl unbound stop"
  tests     = [opnsense_monit_test.cpu.id]
  poll_time = %[4]q
}

resource "opnsense_monit_alert" "ops" {
  enabled   = false
  recipient = "ops@example.com"
  events    = %[3]s
}

data "opnsense_monit_service" "unbound" {
  id = opnsense_monit_service.unbound.id
}

data "opnsense_monit_test" "cpu" {
  id = opnsense_monit_test.cpu.id
}

data "opnsense_monit_alert" "ops" {
  id = opnsense_monit_alert.ops.id
}
`, cpu, action, events, poll)
}
