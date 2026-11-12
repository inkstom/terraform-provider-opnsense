package captiveportal_test

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func templateZip(t *testing.T, body string) string {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("index.html")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// Runs on 26.1 (captiveportal/service/*Template) and 26.7 (captiveportal/template/*).
func TestAccCaptiveportalTemplateResource(t *testing.T) {
	one := templateZip(t, "<html><body>tf acc one</body></html>")
	two := templateZip(t, "<html><body>tf acc two</body></html>")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCaptiveportalTemplateConfig("tfacc", one),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_captiveportal_template.test", "name", "tfacc"),
					resource.TestCheckResourceAttrSet("opnsense_captiveportal_template.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_captiveportal_template.test", "file_id"),
					resource.TestCheckResourceAttrSet("opnsense_captiveportal_template.test", "content_sha256"),
					resource.TestCheckResourceAttr("data.opnsense_captiveportal_template.test", "name", "tfacc"),
					resource.TestCheckResourceAttrSet("data.opnsense_captiveportal_template.test", "content_base64"),
				),
			},
			{
				ResourceName:            "opnsense_captiveportal_template.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content_base64", "content_sha256"},
			},
			{
				Config: testAccCaptiveportalTemplateConfig("tfacc2", two),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_captiveportal_template.test", "name", "tfacc2"),
					resource.TestCheckResourceAttr("data.opnsense_captiveportal_template.test", "name", "tfacc2"),
				),
			},
		},
	})
}

func testAccCaptiveportalTemplateConfig(name, content string) string {
	return fmt.Sprintf(`
resource "opnsense_captiveportal_template" "test" {
  name           = %[1]q
  content_base64 = %[2]q
}

data "opnsense_captiveportal_template" "test" {
  id         = opnsense_captiveportal_template.test.id
  depends_on = [opnsense_captiveportal_template.test]
}
`, name, content)
}
