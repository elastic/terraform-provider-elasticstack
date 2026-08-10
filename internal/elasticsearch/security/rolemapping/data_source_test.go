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

package rolemapping_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const roleMappingDataSourceName = "data.elasticstack_elasticsearch_security_role_mapping.test"

func TestAccDataSourceSecurityRoleMapping(t *testing.T) {
	names := []string{
		sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum),
		sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum) + " " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum),
	}

	for _, roleMappingName := range names {
		idPattern := regexp.MustCompile(`^[^/]*/` + regexp.QuoteMeta(roleMappingName) + `$`)

		resource.Test(t, resource.TestCase{
			PreCheck: func() { acctest.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(roleMappingDataSourceName, "id", idPattern),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "enabled", "true"),
						checks.TestCheckResourceListAttr(roleMappingDataSourceName, "roles", []string{"admin"}),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"rules",
							`{"any":[{"field":{"username":["esadmin"]}},{"field":{"groups":["cn=admins,dc=example,dc=com"]}}]}`,
						),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "metadata", `{"version":1}`),
						resource.TestCheckNoResourceAttr(roleMappingDataSourceName, "role_templates"),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingDataSourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "enabled", "false"),
						checks.TestCheckResourceListAttr(roleMappingDataSourceName, "roles", []string{"admin", "user"}),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"rules",
							`{"any":[{"field":{"username":["esadmin"]}},{"field":{"groups":["cn=admins,dc=example,dc=com"]}}]}`,
						),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "metadata", `{}`),
						resource.TestCheckNoResourceAttr(roleMappingDataSourceName, "role_templates"),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("rules_all"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingDataSourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "enabled", "true"),
						checks.TestCheckResourceListAttr(roleMappingDataSourceName, "roles", []string{"admin"}),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"rules",
							`{"all":[{"field":{"username":["poweruser"]}}]}`,
						),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("single_element_array"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingDataSourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"rules",
							`{"field":{"groups":["project1"]}}`,
						),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("role_templates"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingDataSourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "enabled", "false"),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "roles.#", "0"),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"role_templates",
							`[{"format":"json","template":"{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"}]`,
						),
						resource.TestCheckResourceAttr(
							roleMappingDataSourceName,
							"rules",
							`{"any":[{"field":{"username":["esadmin"]}},{"field":{"groups":["cn=admins,dc=example,dc=com"]}}]}`,
						),
						resource.TestCheckResourceAttr(roleMappingDataSourceName, "metadata", `{}`),
					),
				},
			},
		})
	}
}
