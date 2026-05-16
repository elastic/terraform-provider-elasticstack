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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceComponentTemplateBooleanMappingNoDrift verifies that a component template
// whose template.mappings contains boolean scalar values does not produce spurious plan changes
// after the initial apply. The regression trigger is dynamic: false — Elasticsearch echoes it as
// the JSON string "false" (because DynamicMapping uses encoding.TextMarshaler), so the mappings
// custom type must compare that semantically equal to the original boolean false to prevent drift
// (REQ-022–REQ-025). date_detection and numeric_detection are included as additional stability
// anchors but do not trigger the stringification regression themselves.
func TestAccResourceComponentTemplateBooleanMappingNoDrift(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             "elasticstack_elasticsearch_component_template.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}

// TestAccResourceComponentTemplateNullSettingsNoDrift verifies that a component template
// whose template.settings contains a JSON null value does not produce spurious plan changes
// after the initial apply. Elasticsearch echoes null as the string "null"; the settings custom
// type must canonicalize that consistently to prevent drift.
func TestAccResourceComponentTemplateNullSettingsNoDrift(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             "elasticstack_elasticsearch_component_template.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}
