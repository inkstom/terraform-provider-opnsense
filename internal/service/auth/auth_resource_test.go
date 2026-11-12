package auth_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var csrfInput = regexp.MustCompile(`<input type="hidden" name="([^"]+)" value="([^"]+)"`)

// testAccCheckWebLogin logs in to the web UI as user/password and checks
// whether it succeeds, proving the password was really set.
func testAccCheckWebLogin(user, password string, want bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		jar, _ := cookiejar.New(nil)
		c := &http.Client{Jar: jar, Timeout: 30 * time.Second, Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}, //nolint:gosec // test VMs use self-signed certs
		}}
		base := strings.TrimSuffix(os.Getenv("OPNSENSE_URI"), "/") + "/"
		resp, err := c.Get(base)
		if err != nil {
			return err
		}
		page, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		form := url.Values{"usernamefld": {user}, "passwordfld": {password}, "login": {"1"}}
		if m := csrfInput.FindSubmatch(page); m != nil {
			form.Set(string(m[1]), string(m[2]))
		}
		resp, err = c.PostForm(base, form)
		if err != nil {
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		got := !strings.Contains(string(body), "usernamefld")
		if got != want {
			return fmt.Errorf("web login as %s: success=%v, want %v", user, got, want)
		}
		return nil
	}
}

func TestAccAuthGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthGroupConfig("tf-acc-group", "first", `["page-diagnostics-arptable", "page-diagnostics-authentication"]`, `["0"]`, `["192.0.2.0/24"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "name", "tf-acc-group"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "description", "first"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "privileges.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_auth_group.test", "privileges.*", "page-diagnostics-arptable"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "members.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_auth_group.test", "members.*", "0"),
					resource.TestCheckTypeSetElemAttr("opnsense_auth_group.test", "source_networks.*", "192.0.2.0/24"),
					resource.TestCheckResourceAttrSet("opnsense_auth_group.test", "gid"),
					resource.TestCheckResourceAttrSet("opnsense_auth_group.test", "id"),
					resource.TestCheckResourceAttrPair("data.opnsense_auth_group.test", "gid", "opnsense_auth_group.test", "gid"),
					resource.TestCheckResourceAttr("data.opnsense_auth_group.test", "privileges.#", "2"),
				),
			},
			{
				ResourceName:      "opnsense_auth_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthGroupConfig("tf-acc-group-renamed", "second", `["page-diagnostics-arptable"]`, `[]`, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "name", "tf-acc-group-renamed"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "description", "second"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "privileges.#", "1"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "members.#", "0"),
					resource.TestCheckResourceAttr("opnsense_auth_group.test", "source_networks.#", "0"),
				),
			},
		},
	})
}

func testAccAuthGroupConfig(name, description, privileges, members, networks string) string {
	return fmt.Sprintf(`
resource "opnsense_auth_group" "test" {
  name            = %[1]q
  description     = %[2]q
  privileges      = %[3]s
  members         = %[4]s
  source_networks = %[5]s
}

data "opnsense_auth_group" "test" {
  id = opnsense_auth_group.test.id
}
`, name, description, privileges, members, networks)
}

func TestAccAuthUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthUserConfig("tf-acc-user", "First-pass-1!", "Terraform Test", false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "name", "tf-acc-user"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "full_name", "Terraform Test"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "email", "tf-acc@example.com"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "shell", "/bin/sh"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "disabled", "false"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "privileges.#", "1"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "group_memberships.#", "1"),
					resource.TestCheckResourceAttrPair("opnsense_auth_user.test", "group_memberships.0", "opnsense_auth_group.users", "gid"),
					resource.TestCheckResourceAttrSet("opnsense_auth_user.test", "uid"),
					testAccCheckWebLogin("tf-acc-user", "First-pass-1!", true),
					resource.TestCheckResourceAttrPair("data.opnsense_auth_user.test", "uid", "opnsense_auth_user.test", "uid"),
					resource.TestCheckNoResourceAttr("data.opnsense_auth_user.test", "password"),
					// Membership shows up on the group side too.
					resource.TestCheckResourceAttrPair("data.opnsense_auth_group.users", "members.0", "opnsense_auth_user.test", "uid"),
				),
			},
			{
				ResourceName:            "opnsense_auth_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				// New password, disabled, no group, no direct privilege.
				Config: testAccAuthUserConfig("tf-acc-user", "Second-pass-2!", "Renamed", true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "full_name", "Renamed"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "disabled", "true"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "group_memberships.#", "0"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "privileges.#", "0"),
				),
			},
			{
				// Re-enable with the same password: an unchanged password must
				// keep working, and the old one must not.
				Config: testAccAuthUserConfig("tf-acc-user", "Second-pass-2!", "Renamed", false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebLogin("tf-acc-user", "Second-pass-2!", true),
					testAccCheckWebLogin("tf-acc-user", "First-pass-1!", false),
				),
			},
		},
	})
}

func testAccAuthUserConfig(name, password, fullName string, disabled, inGroup bool) string {
	groups, privs := "[]", "[]"
	if inGroup {
		groups = "[opnsense_auth_group.users.gid]"
		privs = `["page-diagnostics-arptable"]`
	}
	return fmt.Sprintf(`
resource "opnsense_auth_group" "users" {
  name = "tf-acc-users"
}

resource "opnsense_auth_user" "test" {
  name              = %[1]q
  password          = %[2]q
  full_name         = %[3]q
  email             = "tf-acc@example.com"
  shell             = "/bin/sh"
  disabled          = %[4]t
  privileges        = %[6]s
  group_memberships = %[5]s
}

data "opnsense_auth_user" "test" {
  id = opnsense_auth_user.test.id
}

data "opnsense_auth_group" "users" {
  id         = opnsense_auth_group.users.id
  depends_on = [opnsense_auth_user.test]
}
`, name, password, fullName, disabled, groups, privs)
}

// A user without a password gets a random one from OPNsense; updates must
// not fail or try to send an empty password.
func TestAccAuthUserResource_NoPassword(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_auth_user" "test" {
  name    = "tf-acc-nopass"
  comment = "api only"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("opnsense_auth_user.test", "password"),
					resource.TestCheckResourceAttr("opnsense_auth_user.test", "comment", "api only"),
				),
			},
			{
				Config: `
resource "opnsense_auth_user" "test" {
  name    = "tf-acc-nopass"
  comment = "still api only"
}`,
				Check: resource.TestCheckResourceAttr("opnsense_auth_user.test", "comment", "still api only"),
			},
			{
				ResourceName:      "opnsense_auth_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
