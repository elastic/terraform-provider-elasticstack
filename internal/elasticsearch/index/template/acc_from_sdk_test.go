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

package template_test

import (
	_ "embed"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed testdata/TestAccResourceIndexTemplateFromSDK/step1_sdk_no_dso/main.tf
var sdkIndexTemplateFromSDKStep1Compat string

// TestAccResourceIndexTemplateFromSDK upgrades state authored by the last Plugin SDK v2 release
// (REQ-042). Step 1 uses registry provider 0.14.5 without template.data_stream_options (not in the
// SDK schema); step 2 applies the Plugin Framework configuration that adds data_stream_options; step 3
// asserts a no-op plan. Requires Elasticsearch >= 9.1.0 for data_stream_options.
func TestAccResourceIndexTemplateFromSDK(t *testing.T) {
	acctest.PreCheck(t)

	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	skipBelowDSO := versionutils.CheckIfVersionIsUnsupported(index.MinSupportedDataStreamOptionsVersion)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: skipBelowDSO,
				// 0.14.5 is the latest registry release at the time of this pin; the resource was still
				// on Plugin SDK v2 through that line. The in-tree provider is Plugin Framework.
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.5",
					},
				},
				Config: sdkIndexTemplateFromSDKStep1Compat,
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "priority", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "data_stream.0.hidden", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "data_stream.0.allow_custom_routing", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.0.mappings", `{"properties":{"from_sdk":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.upgrade",
						"template.0.alias.*",
						map[string]string{
							"name":           "routing_only_alias",
							"index_routing":  "shard-a",
							"search_routing": "shard-a",
						},
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.upgrade",
						"template.0.lifecycle.*",
						map[string]string{"data_retention": "7d"},
					),
				),
			},
			{
				SkipFunc:                 skipBelowDSO,
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2_pf"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "priority", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "data_stream.hidden", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "data_stream.allow_custom_routing", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.mappings", `{"properties":{"from_sdk":{"type":"keyword"}}}`),
					testAccCheckResourceAttrIndexSettingsSemantic("elasticstack_elasticsearch_index_template.upgrade", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.upgrade",
						"template.alias.*",
						map[string]string{
							"name":    "routing_only_alias",
							"routing": "shard-a",
						},
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.lifecycle.data_retention", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.data_stream_options.failure_store.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.upgrade", "template.data_stream_options.failure_store.lifecycle.data_retention", "30d"),
				),
			},
			{
				SkipFunc:                 skipBelowDSO,
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2_pf"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
