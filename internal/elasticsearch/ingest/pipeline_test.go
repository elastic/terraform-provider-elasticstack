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

package ingest_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceIngestPipeline(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIngestPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIngestPipelineCreate(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "name", pipelineName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "description", "Test Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "processors.#", "2"),
				),
			},
			{
				Config: testAccResourceIngestPipelineUpdate(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "name", pipelineName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "description", "Test Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "processors.#", "1"),
				),
			},
		},
	})
}

func testAccResourceIngestPipelineCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = "%s"
  description = "Test Pipeline"

  processors = [
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    }),
    <<EOF
    {"json": {
      "field": "data",
      "target_field": "parsed_data"
    }}
EOF
    ,
  ]
}
	`, name)
}

func testAccResourceIngestPipelineUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = "%s"
  description = "Test Pipeline"

  processors = [
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    })
  ]
}
	`, name)
}

func checkResourceIngestPipelineDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_ingest_pipeline" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.Indices.Get([]string{compID.ResourceID})
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Ingest pipeline (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
