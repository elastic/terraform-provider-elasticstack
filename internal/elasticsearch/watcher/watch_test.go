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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "active", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "body", `{"trigger":{"schedule":{"cron":"0 0/1 * * * ?"}},"input":{"none":{}},"condition":{"always":{}},"actions":{}}`),
				),
			},
			{
				Config: testAccResourceWatchUpdate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "active", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "body", `{"trigger":{"schedule":{"cron":"0 0/1 * * * ?"}},"input":{"search":{"request":{"indices":["logstash*"],"body":{"query":{"bool":{"must":{"match":{"response":404}},"filter":{"range":{"@timestamp":{"from":"{{ctx.trigger.scheduled_time}}||-5m","to":"{{ctx.trigger.triggered_time}}"}}}}}}}}},"condition":{"always":{}},"actions":{},"metadata":{"example_key":"example_value"},"throttle_period_in_millis":10000}`),
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

 resource "elasticstack_elasticsearch_watcher_watch" "test" {
  watch_id = "%s"
	active = false
 	body = <<EOF
	{
		"trigger" : {
			"schedule" : { "cron" : "0 0/1 * * * ?" }
		},
		"input" : {
			"none" : {}
		},
		"condition" : {
			"always" : {}
		},
		"actions" : {}
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

 resource "elasticstack_elasticsearch_watcher_watch" "test" {
  watch_id = "%s"
	active = true
 	body = <<EOF
	{
		"trigger" : {
			"schedule" : { "cron" : "0 0/1 * * * ?" }
		},
		"input" : {
			"search" : {
				"request" : {
					"indices" : [
						"logstash*"
					],
					"body" : {
						"query" : {
							"bool" : {
								"must" : {
									"match": {
										 "response": 404
									}
								},
								"filter" : {
									"range": {
										"@timestamp": {
											"from": "{{ctx.trigger.scheduled_time}}||-5m",
											"to": "{{ctx.trigger.triggered_time}}"
										}
									}
								}
							}
						}
					}
				}
			}
		},
		"condition" : {
			"compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
		},
		"actions" : {
			"email_admin" : {
				"email" : {
					"to" : "admin@domain.host.com",
					"subject" : "404 recently encountered"
				}
			}
		},
		"metadata" : {
			"example_key" : "example_value"
		},
		"throttle_period_in_millis" : 10000
	}
EOF
 }
 	`, watchID)
}

func checkResourceWatchDestroy(s *terraform.State) error {

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_watcher_watch" {
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
