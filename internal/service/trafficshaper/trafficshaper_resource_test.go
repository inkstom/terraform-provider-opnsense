package trafficshaper_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTrafficshaperResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficshaperConfig(50, "fq_codel", 50, `["192.0.2.0/24", "198.51.100.0/24"]`, `["ef"]`, "queue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_trafficshaper_pipe.test", "bandwidth", "50"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_pipe.test", "scheduler", "fq_codel"),
					resource.TestCheckResourceAttrSet("opnsense_trafficshaper_pipe.test", "number"),
					resource.TestCheckNoResourceAttr("opnsense_trafficshaper_pipe.test", "queue_size"),
					resource.TestCheckResourceAttrPair("opnsense_trafficshaper_queue.test", "pipe", "opnsense_trafficshaper_pipe.test", "id"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_queue.test", "weight", "50"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_rule.test", "source.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_trafficshaper_rule.test", "source.*", "198.51.100.0/24"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_rule.test", "dscp.#", "1"),
					resource.TestCheckResourceAttrPair("opnsense_trafficshaper_rule.test", "target", "opnsense_trafficshaper_queue.test", "id"),
					resource.TestCheckResourceAttr("data.opnsense_trafficshaper_rule.test", "destination_port", "443"),
					resource.TestCheckResourceAttr("data.opnsense_trafficshaper_pipe.test", "bandwidth_metric", "Mbit"),
					resource.TestCheckResourceAttr("data.opnsense_trafficshaper_queue.test", "weight", "50"),
					testAccCheckPipeLive("opnsense_trafficshaper_pipe.test", "50.000 Mbit/s"),
				),
			},
			{ResourceName: "opnsense_trafficshaper_pipe.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_trafficshaper_queue.test", ImportState: true, ImportStateVerify: true},
			{ResourceName: "opnsense_trafficshaper_rule.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccTrafficshaperConfig(20, "", 80, `["any"]`, `[]`, "pipe"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_trafficshaper_pipe.test", "bandwidth", "20"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_pipe.test", "scheduler", ""),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_queue.test", "weight", "80"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_rule.test", "source.#", "1"),
					resource.TestCheckResourceAttr("opnsense_trafficshaper_rule.test", "dscp.#", "0"),
					resource.TestCheckResourceAttrPair("opnsense_trafficshaper_rule.test", "target", "opnsense_trafficshaper_pipe.test", "id"),
					testAccCheckPipeLive("opnsense_trafficshaper_pipe.test", "20.000 Mbit/s"),
				),
			},
		},
	})
}

func testAccTrafficshaperConfig(bw int, scheduler string, weight int, source, dscp, target string) string {
	return fmt.Sprintf(`
resource "opnsense_trafficshaper_pipe" "test" {
  bandwidth   = %[1]d
  scheduler   = %[2]q
  description = "tf acc pipe"
}

resource "opnsense_trafficshaper_queue" "test" {
  pipe        = opnsense_trafficshaper_pipe.test.id
  weight      = %[3]d
  description = "tf acc queue"
}

resource "opnsense_trafficshaper_rule" "test" {
  sequence         = 10
  interface        = "wan"
  protocol         = "tcp"
  source           = %[4]s
  destination_port = "443"
  dscp             = %[5]s
  direction        = "out"
  target           = opnsense_trafficshaper_%[6]s.test.id
  description      = "tf acc rule"
}

data "opnsense_trafficshaper_pipe" "test" {
  id = opnsense_trafficshaper_pipe.test.id
}

data "opnsense_trafficshaper_queue" "test" {
  id = opnsense_trafficshaper_queue.test.id
}

data "opnsense_trafficshaper_rule" "test" {
  id = opnsense_trafficshaper_rule.test.id
}
`, bw, scheduler, weight, source, dscp, target)
}

// testAccCheckPipeLive confirms the pipe was applied to the running shaper
// (ipfw/dummynet), not just saved to the config.
func testAccCheckPipeLive(name, wantBandwidth string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s not in state", name)
		}
		number := rs.Primary.Attributes["number"]

		req, err := http.NewRequest(http.MethodGet, os.Getenv("OPNSENSE_URI")+"/api/trafficshaper/service/statistics", nil)
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

		var stats struct {
			Items []struct {
				Pipe json.RawMessage `json:"pipe"` // a string for pipes, a number for queues
				Bw   string          `json:"bw"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("decode statistics: %w", err)
		}
		for _, p := range stats.Items {
			if strings.Trim(string(p.Pipe), `"`) == number && p.Bw != "" {
				if p.Bw != wantBandwidth {
					return fmt.Errorf("live pipe %s bandwidth = %q, want %q", number, p.Bw, wantBandwidth)
				}
				return nil
			}
		}
		return fmt.Errorf("pipe %s not in live shaper statistics", number)
	}
}
