package monit_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// The test VM has Monit disabled, polling every 120s and mailing via
// 127.0.0.1:25 without credentials; the last step puts that back.
func TestAccMonitSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             `resource "opnsense_monit_settings" "test" {}`,
				ResourceName:       "opnsense_monit_settings.test",
				ImportState:        true,
				ImportStateId:      "monit_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccMonitSettingsConfig(true, 60, `["192.0.2.25", "192.0.2.26"]`, 587, "monit", "tf-acc-secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "interval", "60"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_servers.#", "2"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_port", "587"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_password", "tf-acc-secret"),
					// Unset attributes keep the firewall's value.
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_ssl_version", "auto"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "httpd_port", "2812"),
					resource.TestCheckResourceAttr("data.opnsense_monit_settings.test", "mail_port", "587"),
					resource.TestCheckNoResourceAttr("data.opnsense_monit_settings.test", "mail_password"),
					testAccCheckMonitStatus("running"),
				),
			},
			{
				Config: testAccMonitSettingsConfig(false, 120, `["127.0.0.1"]`, 25, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_servers.#", "1"),
					resource.TestCheckResourceAttr("opnsense_monit_settings.test", "mail_password", ""),
					testAccCheckMonitStatus("disabled"),
				),
			},
		},
	})
}

func TestAccMonitSettingsResource_CreateBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "opnsense_monit_settings" "test" {}`,
				ExpectError: regexp.MustCompile("Import Required"),
			},
		},
	})
}

func testAccMonitSettingsConfig(enabled bool, interval int, servers string, port int, user, password string) string {
	return fmt.Sprintf(`
resource "opnsense_monit_settings" "test" {
  enabled       = %[1]t
  interval      = %[2]d
  mail_servers  = %[3]s
  mail_port     = %[4]d
  mail_username = %[5]q
  mail_password = %[6]q
}

data "opnsense_monit_settings" "test" {
  depends_on = [opnsense_monit_settings.test]
}
`, enabled, interval, servers, port, user, password)
}

func testAccCheckMonitStatus(want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/monit/service/status", nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"},
		}}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var status struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return err
		}
		if status.Status != want {
			return fmt.Errorf("monit service status = %q, want %q", status.Status, want)
		}
		return nil
	}
}
