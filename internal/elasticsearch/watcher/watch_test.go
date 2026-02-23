// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package watcher_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	watchTriggerCreateExpected = `{"schedule":{"cron":"0 0/1 * * * ?"}}`
	watchTriggerUpdateExpected = `{"schedule":{"cron":"0 0/2 * * * ?"}}`
	watchInputNoneExpected     = `{"none":{}}`
	watchConditionAlways       = `{"always":{}}`
	watchActionsEmpty          = `{}`
	watchMetadataEmpty         = `{}`
	watchInputSimpleExpected   = `{"simple":{"name":"example"}}`
	watchConditionNever        = `{"never":{}}`
	watchActionsLogExpected    = `{"log":{"logging":{"level":"info","text":"example logging text"}}}`
	watchMetadataExample       = `{"example_key":"example_value"}`
	watchTransformExpected     = `{"search":{"request":{"body":{"query":{"match_all":{}}},"indices":[],"rest_total_hits_as_int":true,` +
		`"search_type":"query_then_fetch"}}}`
)

func TestResourceWatch(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceWatchDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWatchCreate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "trigger", watchTriggerCreateExpected),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "input", watchInputNoneExpected),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "condition", watchConditionAlways),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "actions", watchActionsEmpty),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "metadata", watchMetadataEmpty),
				),
			},
			{
				Config: testAccResourceWatchUpdate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "trigger", watchTriggerUpdateExpected),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "input", watchInputSimpleExpected),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "condition", watchConditionNever),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "actions", watchActionsLogExpected),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "metadata", watchMetadataExample),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "transform", watchTransformExpected),
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

	transform = <<EOF
	{
		"search" : {
			"request" : {
				"body" : { 
					"query" : { 
						"match_all" : {} 
					}
				},
				"indices": [],
				"rest_total_hits_as_int" : true,
				"search_type": "query_then_fetch"
			}
		}
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
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Watcher.GetWatch(compID.ResourceID)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("watch (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
