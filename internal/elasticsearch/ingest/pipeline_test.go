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
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const ingestPipelineResourceName = "elasticstack_elasticsearch_ingest_pipeline.test_pipeline"

func TestAccResourceIngestPipeline(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := ingestPipelineResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
		},
	})
}

// TestAccResourceIngestPipeline_unsetOptionals exercises the "set then clear" path
// for every optional attribute (description, metadata, on_failure) and asserts
// they are absent from state after removal.
func TestAccResourceIngestPipeline_unsetOptionals(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := ingestPipelineResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_set"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "All optionals set"),
					CheckResourceJSON(resourceName, "metadata", `{"owner":"test"}`),
					resource.TestCheckResourceAttr(resourceName, "on_failure.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_optionals"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", pipelineName),
					resource.TestCheckResourceAttr(resourceName, "processors.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckNoResourceAttr(resourceName, "metadata"),
					resource.TestCheckNoResourceAttr(resourceName, "on_failure.#"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_optionals"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceIngestPipeline_multiOnFailure verifies coverage for on_failure with
// more than one element and update-in-place of existing elements (not just removal).
func TestAccResourceIngestPipeline_multiOnFailure(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := ingestPipelineResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("two"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "on_failure.#", "2"),
					CheckResourceJSON(resourceName, "on_failure.0", `{"set":{"field":"_index","value":"failed-{{ _index }}"}}`),
					CheckResourceJSON(resourceName, "on_failure.1", `{"set":{"field":"error_message","value":"{{ _ingest.on_failure_message }}"}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("two_mutated"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "on_failure.#", "2"),
					CheckResourceJSON(resourceName, "on_failure.0", `{"set":{"field":"_index","value":"dlq-{{ _index }}"}}`),
					CheckResourceJSON(resourceName, "on_failure.1", `{"set":{"field":"error_reason","value":"{{ _ingest.on_failure_message }}"}}`),
				),
			},
		},
	})
}

// TestAccResourceIngestPipeline_validators locks in the SizeAtLeast(1) validators
// on processors and on_failure.
func TestAccResourceIngestPipeline_validators(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_processors"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				ExpectError:              regexp.MustCompile(`(?s)processors.*at least 1`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_on_failure"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				ExpectError:              regexp.MustCompile(`(?s)on_failure.*at least 1`),
			},
		},
	})
}

// TestAccResourceIngestPipeline_forceNew verifies that changing `name` triggers
// a destroy-and-recreate, locking in the RequiresReplace plan modifier.
func TestAccResourceIngestPipeline_forceNew(t *testing.T) {
	initialName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	renamedName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := ingestPipelineResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(initialName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", initialName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("renamed"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(renamedName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", renamedName),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceIngestPipelineFromSDK/create/main.tf
var ingestPipelineSDKCreateConfig string

// TestAccResourceIngestPipelineFromSDK upgrades state authored by the last Plugin SDK v2
// release (v0.14.5) to the in-tree Plugin Framework implementation. Step 1 creates the
// pipeline using the registry SDK provider; step 2 applies the same configuration with
// the local PF provider and asserts state is preserved; step 3 asserts a no-op plan.
func TestAccResourceIngestPipelineFromSDK(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := ingestPipelineResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.5",
					},
				},
				Config:          ingestPipelineSDKCreateConfig,
				ConfigVariables: config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", pipelineName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "processors.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "on_failure.#", "1"),
					CheckResourceJSON(resourceName, "metadata", `{"owner":"test"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", pipelineName),
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func checkResourceIngestPipelineDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_ingest_pipeline" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		_, err = typedClient.Ingest.GetPipeline().Id(compID.ResourceID).Do(context.Background())
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}
		return fmt.Errorf("Ingest pipeline (%s) still exists", compID.ResourceID)
	}
	return nil
}
