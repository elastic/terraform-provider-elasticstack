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
	"context"
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minSupportedDestAliasesVersion = version.Must(version.NewSemver("8.8.0"))
var minSupportedAdvancedSettingsVersion = version.Must(version.NewSemver("8.5.0"))

func TestAccResourceTransformWithPivot(t *testing.T) {

	transformNamePivot := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNamePivot),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "pivot"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.time.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.time.delay", "20s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "timeout", "1m"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNamePivot),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "yet another test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.indices.1", "additional_index"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.index", "dest_index_for_transform_v2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_pivot", "pivot"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.time.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.time.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "timeout", "1m"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNamePivot),
				},
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformNameLatest),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "name", transformNameLatest),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_latest", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "description", "test description (latest)"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "source.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "destination.index", "dest_index_for_transform"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAdvancedSettingsVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
					"pipeline_name":  config.StringVariable(pipelineName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "name", transformName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "source.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "source.query", `{"term":{"status":"active"}}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_transform.test_advanced", "source.runtime_mappings"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "destination.index", "dest_index_for_transform_advanced"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_advanced", "destination.pipeline", pipelineName),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
					"index_name":     config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.indices.0", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.index", "dest_index_for_transform"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.0.move_on_creation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.1.move_on_creation", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.0.move_on_creation", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.aliases.1.move_on_creation", "true"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceTransformFromSDK/main.tf
var transformFromSDKConfig string

// TestAccResourceTransformFromSDK verifies that state created by the last
// SDK-based provider release (v0.14.5 — where source/destination/sync/
// retention_policy were ListNestedBlocks with singleton lists) can be read and
// re-applied without changes by the current Plugin Framework implementation,
// which uses SingleNestedBlock for the same blocks. This guards both the
// schema-shape change and the v0→v1 state upgrader.
func TestAccResourceTransformFromSDK(t *testing.T) {
	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceTransformDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source: "elastic/elasticstack",
						// last SDK-backed release of the transform resource —
						// do not bump without re-checking upgrade compatibility
						VersionConstraint: "0.14.5",
					},
				},
				Config: transformFromSDKConfig,
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "destination.0.index", "dest_index_for_transform_from_sdk"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "sync.0.time.0.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "retention_policy.0.time.0.field", "order_date"),
				),
			},
			{
				// Re-apply with the in-tree PF provider — the v0→v1 state
				// upgrader must convert singleton-list state into single
				// objects, and the resulting plan must be a no-op.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sdk"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "source.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "destination.index", "dest_index_for_transform_from_sdk"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "sync.time.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "sync.time.delay", "20s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "retention_policy.time.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "retention_policy.time.max_age", "30d"),
				),
			},
			{
				// Import verification after the v0→v1 migration.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sdk"),
				ConfigVariables: config.Variables{
					"transform_name": config.StringVariable(transformName),
				},
				ResourceName:            "elasticstack_elasticsearch_transform.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"defer_validation", "elasticsearch_connection", "timeout"},
			},
		},
	})
}

// TestAccResourceTransformValidation verifies schema-level validation errors
// for pivot/latest mutual exclusion and nested required fields in sync.time
// and retention_policy.time.
func TestAccResourceTransformValidation(t *testing.T) {
	cases := []struct {
		name        string
		dir         string
		expectError *regexp.Regexp
	}{
		{
			name:        "both pivot and latest set",
			dir:         "both_pivot_latest",
			expectError: regexp.MustCompile(`(?i)one \(and only one\) of \[pivot,latest\]`),
		},
		{
			name:        "neither pivot nor latest set",
			dir:         "neither_pivot_latest",
			expectError: regexp.MustCompile(`(?i)one \(and only one\) of \[pivot,latest\]`),
		},
		{
			name:        "sync block without time",
			dir:         "sync_no_time",
			expectError: regexp.MustCompile(`(?is)must be specified when[\s\S]*is specified`),
		},
		{
			name:        "sync.time block without field",
			dir:         "sync_time_no_field",
			expectError: regexp.MustCompile(`(?is)must be specified when[\s\S]*is specified`),
		},
		{
			name:        "retention_policy block without time",
			dir:         "retention_no_time",
			expectError: regexp.MustCompile(`(?is)must be specified when[\s\S]*is specified`),
		},
		{
			name:        "retention_policy.time block without field",
			dir:         "retention_time_missing",
			expectError: regexp.MustCompile(`(?is)must be specified when[\s\S]*is specified`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.PreCheck(t) },
				Steps: []resource.TestStep{
					{
						ProtoV6ProviderFactories: acctest.Providers,
						ConfigDirectory:          acctest.NamedTestCaseDirectory(tc.dir),
						ConfigVariables: config.Variables{
							"transform_name": config.StringVariable(transformName),
						},
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

func checkResourceTransformDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
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
		res, err := esClient.Transform.GetTransform().TransformId(compID.ResourceID).Do(context.Background())
		if err != nil {
			var esErr *types.ElasticsearchError
			if errors.As(err, &esErr) && esErr.Status == 404 {
				continue
			}
			return err
		}

		if len(res.Transforms) > 0 {
			return fmt.Errorf("transform (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
