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
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

//go:embed testdata/TestAccResourceSecurityRoleMapping/create/main.tf
var roleMappingCreateConfig string

const roleMappingResourceName = "elasticstack_elasticsearch_security_role_mapping.test"

func TestAccResourceSecurityRoleMapping(t *testing.T) {
	names := []string{
		sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum),
		sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum) + " " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum),
	}

	for _, roleMappingName := range names {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkResourceSecurityRoleMappingDestroy,
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingResourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "true"),
						checks.TestCheckResourceListAttr(roleMappingResourceName, "roles", []string{"admin"}),
						resource.TestCheckResourceAttr(roleMappingResourceName, "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "metadata", `{"version":1}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingResourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "false"),
						checks.TestCheckResourceListAttr(roleMappingResourceName, "roles", []string{"admin", "user"}),
						resource.TestCheckResourceAttr(roleMappingResourceName, "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "metadata", `{}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("role_templates"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingResourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "false"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "roles.#", "0"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "role_templates", `[{"format":"json","template":"{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"}]`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "metadata", `{}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("role_templates_updated"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(roleMappingResourceName, "id"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
						resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "roles.#", "0"),
						resource.TestCheckResourceAttr(roleMappingResourceName, "role_templates", `[{"format":"json","template":"{\"source\":\"{{#tojson}}roles{{/tojson}}\"}"}]`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "rules", `{"any":[{"field":{"username":"poweruser"}},{"field":{"groups":"cn=operators,dc=example,dc=com"}}]}`),
						resource.TestCheckResourceAttr(roleMappingResourceName, "metadata", `{}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("role_templates_updated"),
					ConfigVariables: config.Variables{
						"name": config.StringVariable(roleMappingName),
					},
					ResourceName:      roleMappingResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	}
}

func TestAccResourceSecurityRoleMappingFromSDK(t *testing.T) {
	roleMappingName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the role mapping with the last provider version where the role mapping resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				Config: roleMappingCreateConfig,
				ConfigVariables: config.Variables{
					"name": config.StringVariable(roleMappingName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
					resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "true"),
					checks.TestCheckResourceListAttr(roleMappingResourceName, "roles", []string{"admin"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(roleMappingName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(roleMappingResourceName, "name", roleMappingName),
					resource.TestCheckResourceAttr(roleMappingResourceName, "enabled", "true"),
					checks.TestCheckResourceListAttr(roleMappingResourceName, "roles", []string{"admin"}),
				),
			},
		},
	})
}

func checkResourceSecurityRoleMappingDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role_mapping" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		typedClient, err := client.GetESTypedClient()
		if err != nil {
			return err
		}
		_, err = typedClient.Security.GetRoleMapping().Name(compID.ResourceID).Do(context.Background())
		if err != nil {
			if acctest.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("Role mapping (%s) still exists", compID.ResourceID)
	}
	return nil
}
