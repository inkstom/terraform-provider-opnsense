package core_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The test VM starts with HA sync unconfigured; the last step puts it back.
func TestAccCoreHasyncResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_high_availability" "test" {}`,
				ResourceName:       "opnsense_high_availability.test",
				ImportState:        true,
				ImportStateId:      "high_availability",
				ImportStatePersist: true,
			},
			{
				Config: testAccCoreHasyncConfig("192.0.2.10", "root", "tf-acc-secret", `["aliases", "cron"]`, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "synchronize_to_ip", "192.0.2.10"),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "username", "root"),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "password", "tf-acc-secret"),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "sync_items.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_high_availability.test", "sync_items.*", "cron"),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "disable_preempt", "true"),
					// Unset attributes keep the firewall's value.
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "pfsync_version", "1400"),
					resource.TestCheckResourceAttr("data.opnsense_high_availability.test", "synchronize_to_ip", "192.0.2.10"),
					resource.TestCheckNoResourceAttr("data.opnsense_high_availability.test", "password"),
				),
			},
			{
				Config: testAccCoreHasyncConfig("", "", "", `[]`, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "synchronize_to_ip", ""),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "password", ""),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "sync_items.#", "0"),
					resource.TestCheckResourceAttr("opnsense_high_availability.test", "disable_preempt", "false"),
				),
			},
		},
	})
}

func TestAccCoreHasyncResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "opnsense_high_availability" "test" {}`,
				ExpectError: regexp.MustCompile("Import Required"),
			},
		},
	})
}

func testAccCoreHasyncConfig(syncIP, user, password, items string, noPreempt bool) string {
	return fmt.Sprintf(`
resource "opnsense_high_availability" "test" {
  synchronize_to_ip = %[1]q
  username          = %[2]q
  password          = %[3]q
  sync_items        = %[4]s
  disable_preempt   = %[5]t
}

data "opnsense_high_availability" "test" {
  depends_on = [opnsense_high_availability.test]
}
`, syncIP, user, password, items, noPreempt)
}
