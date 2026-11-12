package dnsmasq_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDnsmasqTagDomainBootResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsmasqTagDomainBootConfig("tfaccpxe", "192.0.2.53", "pxelinux.0", "[opnsense_dnsmasq_tag.test.id]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_tag.test", "tag", "tfaccpxe"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_domain.test", "domain", "tfacc.example.com"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_domain.test", "server", "192.0.2.53"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_domain.test", "sequence", "1"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_boot.test", "filename", "pxelinux.0"),
					resource.TestCheckTypeSetElemAttrPair("opnsense_dnsmasq_boot.test", "tags.*", "opnsense_dnsmasq_tag.test", "id"),
					resource.TestCheckResourceAttr("data.opnsense_dnsmasq_tag.test", "tag", "tfaccpxe"),
					resource.TestCheckResourceAttr("data.opnsense_dnsmasq_domain.test", "server", "192.0.2.53"),
					resource.TestCheckResourceAttr("data.opnsense_dnsmasq_boot.test", "tags.#", "1"),
				),
			},
			{ResourceName: "opnsense_dnsmasq_tag.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_dnsmasq_domain.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_dnsmasq_boot.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccDnsmasqTagDomainBootConfig("tfaccpxe2", "198.51.100.53", "ipxe.efi", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_dnsmasq_tag.test", "tag", "tfaccpxe2"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_domain.test", "server", "198.51.100.53"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_boot.test", "filename", "ipxe.efi"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_boot.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccDnsmasqTagDomainBootConfig(tag, server, file, tags string) string {
	return fmt.Sprintf(`
resource "opnsense_dnsmasq_tag" "test" {
  tag = %[1]q
}

resource "opnsense_dnsmasq_domain" "test" {
  domain      = "tfacc.example.com"
  server      = %[2]q
  description = "tf acc"
}

resource "opnsense_dnsmasq_boot" "test" {
  filename       = %[3]q
  server_address = "192.0.2.69"
  tags           = %[4]s
  description    = "tf acc"
}

data "opnsense_dnsmasq_tag" "test" {
  id = opnsense_dnsmasq_tag.test.id
}

data "opnsense_dnsmasq_domain" "test" {
  id = opnsense_dnsmasq_domain.test.id
}

data "opnsense_dnsmasq_boot" "test" {
  id = opnsense_dnsmasq_boot.test.id
}
`, tag, server, file, tags)
}
