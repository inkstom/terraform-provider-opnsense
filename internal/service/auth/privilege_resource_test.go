package auth_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAuthPrivilegeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthPrivilegeConfig(`[opnsense_auth_user.u.id]`, `[opnsense_auth_group.g.id]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_privilege.test", "id", "page-diagnostics-arptable"),
					resource.TestCheckResourceAttrPair("opnsense_auth_privilege.test", "users.0", "opnsense_auth_user.u", "id"),
					resource.TestCheckResourceAttrPair("opnsense_auth_privilege.test", "groups.0", "opnsense_auth_group.g", "id"),
					resource.TestCheckResourceAttr("data.opnsense_auth_privilege.test", "groups.#", "1"),
				),
			},
			{
				ResourceName:      "opnsense_auth_privilege.test",
				ImportState:       true,
				ImportStateId:     "page-diagnostics-arptable",
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthPrivilegeConfig(`[]`, `[opnsense_auth_group.g.id]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_privilege.test", "users.#", "0"),
					resource.TestCheckResourceAttr("opnsense_auth_privilege.test", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccAuthPrivilegeConfig(users, groups string) string {
	return `
resource "opnsense_auth_user" "u" {
  name = "tf-acc-priv-user"
}

resource "opnsense_auth_group" "g" {
  name = "tf-acc-priv-group"
}

resource "opnsense_auth_privilege" "test" {
  privilege = "page-diagnostics-arptable"
  users     = ` + users + `
  groups    = ` + groups + `
}

data "opnsense_auth_privilege" "test" {
  privilege  = "page-diagnostics-arptable"
  depends_on = [opnsense_auth_privilege.test]
}
`
}
