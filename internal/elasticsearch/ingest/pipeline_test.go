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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceIngestPipeline(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_ingest_pipeline.test_pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIngestPipelineDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", pipelineName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "processors.#", "2"),
					CheckResourceJSON(resourceName, "processors.0", `{"set":{"description":"My set processor description","field":"_meta","value":"indexed"}}`),
					CheckResourceJSON(resourceName, "processors.1", `{"json":{"field":"data","target_field":"parsed_data"}}`),
					resource.TestCheckResourceAttr(resourceName, "on_failure.#", "1"),
					CheckResourceJSON(resourceName, "on_failure.0", `{"set":{"field":"_index","value":"failed-{{ _index }}"}}`),
					CheckResourceJSON(resourceName, "metadata", `{"owner":"test"}`),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", pipelineName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "processors.#", "1"),
					CheckResourceJSON(resourceName, "processors.0", `{"set":{"description":"Updated set processor","field":"_meta","value":"reindexed"}}`),
					CheckResourceJSON(resourceName, "metadata", `{"owner":"updated"}`),
					resource.TestCheckNoResourceAttr(resourceName, "on_failure.#"),
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
