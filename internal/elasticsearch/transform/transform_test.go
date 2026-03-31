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

package transform_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minSupportedDestAliasesVersion = version.Must(version.NewSemver("8.8.0"))
var minSupportedAdvancedSettingsVersion = version.Must(version.NewSemver("8.5.0"))

func TestAccResourceTransformWithPivot(t *testing.T) {

	transformNamePivot := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNamePivot),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "pivot"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.0.time.0.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.0.time.0.delay", "20s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "timeout", "1m"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNamePivot),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "yet another test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.1", "additional_index"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform_v2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "pivot"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.0.time.0.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.0.time.0.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "timeout", "1m"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
			{
				ResourceName:            "elasticstack_elasticsearch_transform.test_pivot",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"defer_validation", "elasticsearch_connection", "timeout"},
			},
		},
	})
}

func TestAccResourceTransformWithLatest(t *testing.T) {

	transformNameLatest := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNameLatest),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "name", transformNameLatest),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_latest", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "description", "test description (latest)"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_latest", "latest"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "frequency", "2m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "defer_validation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "timeout", "1m"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_latest", "pivot"),
				),
			},
		},
	})
}

func TestAccResourceTransformWithAdvancedSettings(t *testing.T) {
	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	pipelineName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedAdvancedSettingsVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
					"pipeline_name":  config.StringVariable(pipelineName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "name", transformName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "source.0.query", `{"term":{"status":"active"}}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "source.0.runtime_mappings"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "destination.0.index", "dest_index_for_transform_advanced"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "destination.0.pipeline", pipelineName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "metadata"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "pivot"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "align_checkpoints", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "dates_as_epoch_millis", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "deduce_mappings", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "docs_per_second", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "num_failure_retries", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "unattended", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "timeout", "1m"),
				),
			},
		},
	})
}

func TestAccResourceTransformNoDefer(t *testing.T) {

	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
					"index_name":     config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "false"),
				),
			},
		},
	})
}

func TestAccResourceTransformWithAliases(t *testing.T) {
	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.move_on_creation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.move_on_creation", "false"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.move_on_creation", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.move_on_creation", "true"),
				),
			},
		},
	})
}

func checkResourceTransformDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_transform" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.TransformGetTransform.WithTransformID(compID.ResourceID)
		res, err := esClient.TransformGetTransform(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("transform (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
