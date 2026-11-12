package syslog_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The test VM keeps 31 days with no size limit; the last step puts that back.
func TestAccSyslogSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_syslog_settings" "test" {}`,
				ResourceName:       "opnsense_syslog_settings.test",
				ImportState:        true,
				ImportStateId:      "syslog_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccSyslogSettingsConfig(20, 1000000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "max_preserve", "20"),
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "max_file_size", "1000000"),
					// Unset attributes keep the firewall's value.
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "log_local", "true"),
					resource.TestCheckResourceAttr("data.opnsense_syslog_settings.test", "max_preserve", "20"),
				),
			},
			{
				Config: testAccSyslogSettingsConfig(31, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "max_preserve", "31"),
					resource.TestCheckResourceAttr("opnsense_syslog_settings.test", "max_file_size", "0"),
				),
			},
		},
	})
}

func TestAccSyslogSettingsResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "opnsense_syslog_settings" "test" {}`,
				ExpectError: regexp.MustCompile("Import Required"),
			},
		},
	})
}

func testAccSyslogSettingsConfig(days, size int) string {
	return fmt.Sprintf(`
resource "opnsense_syslog_settings" "test" {
  max_preserve  = %[1]d
  max_file_size = %[2]d
}

data "opnsense_syslog_settings" "test" {
  depends_on = [opnsense_syslog_settings.test]
}
`, days, size)
}
