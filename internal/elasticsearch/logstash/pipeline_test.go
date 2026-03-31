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
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceLogstashPipeline(t *testing.T) {
	// Pipelines must start with a letter or underscore
	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceLogstashPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"pipeline_id": config.StringVariable(pipelineID)},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"pipeline_id": config.StringVariable(pipelineID)},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_metadata"),
				ConfigVariables:          config.Variables{"pipeline_id": config.StringVariable(pipelineID)},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_metadata"),
				ConfigVariables:          config.Variables{"pipeline_id": config.StringVariable(pipelineID)},
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
		},
	})
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

// TestAccResourceLogstashPipelinePersistedQueue tests queue_type = "persisted" and
// associated persisted-queue settings that have no coverage in the main test.
func TestAccResourceLogstashPipelinePersistedQueue(t *testing.T) {
	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test_persisted"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelinePersistedQueue(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "queue_type", "persisted"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_bytes", "512mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_events", "2000"),
					resource.TestCheckResourceAttr(resourceName, "queue_page_capacity", "128mb"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_acks", "512"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_writes", "1024"),
					resource.TestCheckResourceAttr(resourceName, "queue_checkpoint_retry", "false"),
					resource.TestCheckResourceAttr(resourceName, "queue_drain", "true"),
				),
			},
		},
	})
}

func testAccResourceLogstashPipelinePersistedQueue(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_persisted" {
  pipeline_id = "%s"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"

  queue_type               = "persisted"
  queue_max_bytes          = "512mb"
  queue_max_events         = 2000
  queue_page_capacity      = "128mb"
  queue_checkpoint_acks    = 512
  queue_checkpoint_writes  = 1024
  queue_checkpoint_retry   = false
  queue_drain              = true
}
`, pipelineID)
}

// TestAccResourceLogstashPipelineEnumVariants verifies that the enum attributes
// pipeline_ecs_compatibility and pipeline_ordered accept values other than the ones
// exercised by the main test ("disabled" and "auto").
func TestAccResourceLogstashPipelineEnumVariants(t *testing.T) {
	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test_enums"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Step 1: set ecs_compatibility = "v1", pipeline_ordered = "false"
			{
				Config: testAccResourceLogstashPipelineEnumVariants(pipelineID, "v1", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ecs_compatibility", "v1"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ordered", "false"),
				),
			},
			// Step 2: switch to ecs_compatibility = "v8", pipeline_ordered = "true"
			{
				Config: testAccResourceLogstashPipelineEnumVariants(pipelineID, "v8", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ecs_compatibility", "v8"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ordered", "true"),
				),
			},
		},
	})
}

func testAccResourceLogstashPipelineEnumVariants(pipelineID, ecsCompat, ordered string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_enums" {
  pipeline_id                = "%s"
  pipeline                   = "input{} filter{} output{}"
  username                   = "test_user"
  pipeline_ecs_compatibility = "%s"
  pipeline_ordered           = "%s"
}
`, pipelineID, ecsCompat, ordered)
}

// TestAccResourceLogstashPipelineOptionalUnset verifies that optional attributes can be
// set in one step and then removed (unset) in a subsequent step without error,
// demonstrating proper optional-attribute lifecycle management.
//
// Note: Elasticsearch retains pipeline settings (batch delay, batch size, queue type, etc.)
// server-side after a PUT that omits them. The provider reads them back on every Read, so
// those attributes will remain populated in state even when removed from config. Only simple
// top-level fields like `description` that the provider explicitly sends as empty are cleared.
func TestAccResourceLogstashPipelineOptionalUnset(t *testing.T) {
	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test_optional"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Step 1: create with a selection of optional attributes set.
			{
				Config: testAccResourceLogstashPipelineOptionalSet(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "description", "Optional attrs set"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_delay", "75"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_batch_size", "150"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_workers", "1"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ecs_compatibility", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_ordered", "auto"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_unsafe_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "queue_type", "memory"),
					resource.TestCheckResourceAttr(resourceName, "queue_max_bytes", "256mb"),
				),
			},
			// Step 2: remove optional attributes from config; the resource must apply without
			// error. Elasticsearch retains the previously-sent pipeline/queue settings
			// server-side, so only the description (explicitly set to "" by the provider) can
			// be asserted as cleared.
			{
				Config: testAccResourceLogstashPipelineOptionalUnset(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func testAccResourceLogstashPipelineOptionalSet(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_optional" {
  pipeline_id              = "%s"
  pipeline                 = "input{} filter{} output{}"
  username                 = "test_user"
  description              = "Optional attrs set"
  pipeline_batch_delay     = 75
  pipeline_batch_size      = 150
  pipeline_workers         = 1
  pipeline_ecs_compatibility = "disabled"
  pipeline_ordered         = "auto"
  pipeline_unsafe_shutdown = true
  queue_type               = "memory"
  queue_max_bytes          = "256mb"
}
`, pipelineID)
}

func testAccResourceLogstashPipelineOptionalUnset(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_optional" {
  pipeline_id = "%s"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"
}
`, pipelineID)
}

// TestAccResourceLogstashPipelineForceNew verifies that changing pipeline_id (a ForceNew
// attribute) destroys the old resource and creates a new one.
func TestAccResourceLogstashPipelineForceNew(t *testing.T) {
	firstID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	secondID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test_forcenew"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelineForceNew(firstID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", firstID),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// Changing pipeline_id must trigger a destroy + create (ForceNew).
			{
				Config: testAccResourceLogstashPipelineForceNew(secondID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", secondID),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccResourceLogstashPipelineForceNew(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_forcenew" {
  pipeline_id = "%s"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"
}
`, pipelineID)
}

// TestAccResourceLogstashPipelineExplicitConnection tests the elasticsearch_connection
// block by supplying explicit endpoint credentials from the test environment.
func TestAccResourceLogstashPipelineExplicitConnection(t *testing.T) {
	endpoints := logstashPipelineESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}

	pipelineID := "pipeline-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_logstash_pipeline.test_conn"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceLogstashPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelineExplicitConnection(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(resourceName, "pipeline", "input{} filter{} output{}"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#",
						fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
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

func testAccResourceLogstashPipelineExplicitConnection(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_conn" {
  pipeline_id = "%s"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"

  elasticsearch_connection {
    %s
    insecure = true
  }
}
`, pipelineID, buildLogstashPipelineESConnectionBlock())
}

func buildLogstashPipelineESConnectionBlock() string {
	endpoints := logstashPipelineESEndpoints()
	quoted := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		quoted = append(quoted, fmt.Sprintf("%q", ep))
	}
	endpointList := strings.Join(quoted, ", ")

	if apiKey := os.Getenv("ELASTICSEARCH_API_KEY"); apiKey != "" {
		return fmt.Sprintf(`endpoints = [%s]
    api_key   = %q`, endpointList, apiKey)
	}

	return fmt.Sprintf(`endpoints = [%s]
    username  = %q
    password  = %q`, endpointList, os.Getenv("ELASTICSEARCH_USERNAME"), os.Getenv("ELASTICSEARCH_PASSWORD"))
}

func logstashPipelineESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}
