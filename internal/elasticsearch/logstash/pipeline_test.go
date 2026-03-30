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

package logstash_test

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

func TestResourceLogstashPipeline(t *testing.T) {
	// Pipelines must start with a letter or underscore
	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelineCreate(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "Description of Logstash Pipeline"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_metadata", "{\"type\":\"logstash_pipeline\",\"version\":1}"),
					resource.TestCheckResourceAttr(resourceName, "pipeline", "input{} filter{} output{}"),
					resource.TestCheckResourceAttr(resourceName, "username", "test_user"),
				),
			},
			{
				Config: testAccResourceLogstashPipelineUpdate(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description of Logstash Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "pipeline", "input{} \nfilter{} \noutput{}"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_delay", "100"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_size", "250"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ecs_compatibility", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_metadata", "{\"type\":\"logstash_pipeline\",\"version\":2}"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ordered", "auto"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_plugin_classloaders", "false"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_unsafe_shutdown", "false"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_workers", "2"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_acks", "1024"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_retry", "true"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_writes", "2048"),
					resource.TestCheckResourceAttr(resourceName, "queue_drain", "false"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_bytes", "1mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_events", "0"),
					resource.TestCheckResourceAttr(resourceName, "queue_page_capacity", "64mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_type", "memory"),
					resource.TestCheckResourceAttr(resourceName, "username", "test_user"),
				),
			},
			{
				Config: testAccResourceLogstashPipelineUpdateWithMetadata(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description of Logstash Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "pipeline", "input{} \nfilter{} \noutput{}"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_delay", "100"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_size", "250"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ecs_compatibility", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_metadata", "{\"type\":\"logstash_pipeline\",\"version\":3}"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ordered", "auto"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_plugin_classloaders", "true"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_unsafe_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_workers", "2"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_acks", "1024"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_retry", "true"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_writes", "2048"),
					resource.TestCheckResourceAttr(resourceName, "queue_drain", "true"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_bytes", "2mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_events", "1"),
					resource.TestCheckResourceAttr(resourceName, "queue_page_capacity", "64mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_type", "memory"),
					resource.TestCheckResourceAttr(resourceName, "username", "test_user"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
			},
		},
	})
}

func testAccResourceLogstashPipelineCreate(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
  description = "Description of Logstash Pipeline"
  pipeline = "input{} filter{} output{}"
  username = "test_user"
}
  `, pipelineID)
}

func testAccResourceLogstashPipelineUpdate(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
  description = "Updated description of Logstash Pipeline"
  pipeline = "input{} \nfilter{} \noutput{}"
  username = "test_user"

  pipeline_batch_delay = 100
  pipeline_batch_size = 250
  pipeline_ecs_compatibility = "disabled"
  pipeline_metadata = jsonencode({
    type = "logstash_pipeline"
    version = 2
  })
  pipeline_ordered = "auto"
  pipeline_plugin_classloaders = false
  pipeline_unsafe_shutdown = false
  pipeline_workers = 2
  queue_checkpoint_acks = 1024
  queue_checkpoint_retry = true
  queue_checkpoint_writes = 2048
  queue_drain = false
  queue_max_bytes = "1mb"
  queue_max_events = 0
  queue_page_capacity = "64mb"
  queue_type = "memory"
}
  `, pipelineID)
}

func testAccResourceLogstashPipelineUpdateWithMetadata(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
  description = "Updated description of Logstash Pipeline"
  pipeline = "input{} \nfilter{} \noutput{}"
  username = "test_user"

  pipeline_batch_delay = 100
  pipeline_batch_size = 250
  pipeline_ecs_compatibility = "disabled"
  pipeline_metadata = jsonencode({
    type = "logstash_pipeline"
    version = 3
  })
  pipeline_ordered = "auto"
  pipeline_plugin_classloaders = true
  pipeline_unsafe_shutdown = true
  pipeline_workers = 2
  queue_checkpoint_acks = 1024
  queue_checkpoint_retry = true
  queue_checkpoint_writes = 2048
  queue_drain = true
  queue_max_bytes = "2mb"
  queue_max_events = 1
  queue_page_capacity = "64mb"
  queue_type = "memory"
}
  `, pipelineID)
}

func checkResourceLogstashPipelineDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_logstash_pipeline" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.LogstashGetPipeline(esClient.LogstashGetPipeline.WithDocumentID(compID.ResourceID))
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("logstash pipeline (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
