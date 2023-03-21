package watcher_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceWatch(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceWatchDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWatchCreate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "trigger", `{"schedule":{"cron":"0 0/1 * * * ?"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "input", `{"none":{}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "condition", `{"always":{}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "actions", `{}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "metadata", `{}`),
				),
			},
			{
				Config: testAccResourceWatchUpdate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "trigger", `{"schedule":{"cron":"0 0/2 * * * ?"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "input", `{"simple":{"name":"example"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "condition", `{"never":{}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "actions", `{"log":{"logging":{"level":"info","text":"example logging text"}}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "metadata", `{"example_key":"example_value"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "throttle_period_in_millis", "10000"),
				),
			},
		},
	})
}

func testAccResourceWatchCreate(watchID string) string {
	return fmt.Sprintf(`
 provider "elasticstack" {
   elasticsearch {}
 }

 resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = "%s"
	active = false
 	
	trigger = <<EOF
	{
		"schedule" : { "cron" : "0 0/1 * * * ?" }
	}
EOF
 }
 	`, watchID)
}

func testAccResourceWatchUpdate(watchID string) string {
	return fmt.Sprintf(`
 provider "elasticstack" {
   elasticsearch {}
 }

 resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = "%s"
	active = true
	
	trigger = <<EOF
	{
		"schedule" : { "cron" : "0 0/2 * * * ?" }
	}
EOF

	input = <<EOF
	{
		"simple" : {
			"name" : "example"
		}
	}
EOF

	condition = <<EOF
	{
		"never" : {}
	}
EOF

	actions = <<EOF
	{
		"log" : {
			"logging" : {
				"level" : "info",
				"text" : "example logging text"
			}
		}
	}
EOF

	metadata = <<EOF
	{
		"example_key" : "example_value"
	}
EOF

	throttle_period_in_millis = 10000
 }
 	`, watchID)
}

func checkResourceWatchDestroy(s *terraform.State) error {

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_watch" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Watcher.GetWatch(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("watch (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
