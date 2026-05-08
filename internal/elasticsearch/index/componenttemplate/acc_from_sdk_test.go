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

package componenttemplate_test

import (
	_ "embed"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed testdata/TestAccResourceComponentTemplateFromSDK/config/main.tf
var sdkComponentTemplateFromSDKConfig string

// TestAccResourceComponentTemplateFromSDK upgrades state authored by the last Plugin SDK v2 release
// (REQ-042). Step 1 uses registry provider 0.14.5; step 2 applies the Plugin Framework configuration;
// step 3 asserts a no-op plan.
func TestAccResourceComponentTemplateFromSDK(t *testing.T) {
	acctest.PreCheck(t)

	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				// 0.14.5 is the latest registry release at the time of this pin; the resource was still
				// on Plugin SDK v2 through that line. The in-tree provider is Plugin Framework.
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.5",
					},
				},
				Config: sdkComponentTemplateFromSDKConfig,
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "version", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "template.0.mappings", `{"properties":{"from_sdk":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_component_template.upgrade",
						"template.0.alias.*",
						map[string]string{"name": "my_alias"},
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("config"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "version", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "template.mappings", `{"properties":{"from_sdk":{"type":"keyword"}}}`),
					testAccCheckResourceAttrIndexSettingsSemantic("elasticstack_elasticsearch_component_template.upgrade", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.upgrade", "template.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_component_template.upgrade",
						"template.alias.*",
						map[string]string{"name": "my_alias"},
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("config"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
