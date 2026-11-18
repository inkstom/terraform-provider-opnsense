package dyndns_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDyndnsAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PluginPreCheck(t, "os-ddclient", "/api/dyndns/settings/get") },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Only the required fields, so every default is exercised.
			{
				Config: `
resource "opnsense_dyndns_account" "test" {
  enabled   = false
  service   = "duckdns"
  hostnames = ["tf-acc.duckdns.org"]
  checkip   = "web_icanhazip"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "service", "duckdns"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "protocol", ""),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "ttl", "300"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "force_ssl", "true"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "hostnames.#", "1"),
				),
			},
			{
				Config: testAccDyndnsAccountConfig("first-secret", `["tf-acc.example.com", "tf-acc2.example.com"]`, 600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "protocol", "dyndns2"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "password", "first-secret"),
					resource.TestCheckTypeSetElemAttr("opnsense_dyndns_account.test", "hostnames.*", "tf-acc2.example.com"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "ttl", "600"),
					resource.TestCheckResourceAttrPair("data.opnsense_dyndns_account.test", "hostnames.#", "opnsense_dyndns_account.test", "hostnames.#"),
					resource.TestCheckResourceAttrPair("data.opnsense_dyndns_account.test", "server", "opnsense_dyndns_account.test", "server"),
				),
			},
			{
				ResourceName:      "opnsense_dyndns_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				// OPNsense never returns the password.
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccDyndnsAccountConfig("second-secret", `["tf-acc3.example.com"]`, 900),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "password", "second-secret"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "hostnames.#", "1"),
					resource.TestCheckResourceAttr("opnsense_dyndns_account.test", "ttl", "900"),
				),
			},
		},
	})
}

func testAccDyndnsAccountConfig(password, hostnames string, ttl int) string {
	return fmt.Sprintf(`
resource "opnsense_dyndns_account" "test" {
  enabled     = false
  service     = "custom"
  protocol    = "dyndns2"
  server      = "dyndns.example.com"
  username    = "tf-acc"
  password    = %[1]q
  hostnames   = %[2]s
  checkip     = "web_icanhazip"
  ttl         = %[3]d
  description = "tf-acc dyndns"
}

data "opnsense_dyndns_account" "test" {
  id = opnsense_dyndns_account.test.id
}
`, password, hostnames, ttl)
}
