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

package script_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/TestAccResourceScriptFromSDK/upgrade/main.tf
var testAccResourceScriptFromSDKConfig string

func TestAccResourceScript(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_script.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_script.test", "params"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_script.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_script.test", "context"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// Ensure the provider doesn't panic if the script has been deleted outside of the Terraform flow
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					require.NoError(t, err)

					typedClient := client.GetESClient()

					_, err = typedClient.Core.DeleteScript(scriptID).Do(context.Background())
					require.NoError(t, err)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_script.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_script.test", "context"),
				),
			},
		},
	})
}

func TestAccResourceScriptImport(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_script.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				ResourceName:             "elasticstack_elasticsearch_script.test",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					if err != nil {
						return "", err
					}
					clusterID, diag := client.ClusterID(context.Background())
					if diag.HasError() {
						return "", fmt.Errorf("failed to get cluster uuid: %s", diag[0].Summary())
					}

					return fmt.Sprintf("%s/%s", *clusterID, scriptID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				// context is not returned by the Elasticsearch API so we cannot verify it
				ImportStateVerifyIgnore: []string{"context"},
			},
		},
	})
}

func TestAccResourceScriptSearchTemplate(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "lang", "mustache"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "source", `{"from":"{{from}}","query":{"match":{"message":"{{query_string}}"}},"size":"{{size}}"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "params", `{"query_string":"My query string"}`),
				),
			},
		},
	})
}

func TestAccResourceScriptFromSDK(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the script with the last provider version where the script resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				Config:          testAccResourceScriptFromSDKConfig,
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("upgrade"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
		},
	})
}

func TestAccResourceScriptParamsRemoval(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_script.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "params", `{"modifier":3}`),
					resource.TestCheckNoResourceAttr(resourceName, "context"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckNoResourceAttr(resourceName, "params"),
					resource.TestCheckNoResourceAttr(resourceName, "context"),
				),
			},
		},
	})
}

func TestAccResourceScriptForceNew(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_script.test"
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	replacementScriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
				),
			},
			{
				// Changing script_id must force replacement rather than an in-place update.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(replacementScriptID)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", replacementScriptID),
				),
			},
		},
	})
}

func TestAccResourceScriptContextUpdate(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_script.test"
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("score"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttr(resourceName, "context", "score"),
					resource.TestCheckResourceAttr(resourceName, "source", "_score * params['multiplier']"),
				),
			},
			{
				// context must be updatable in-place from one valid value to another,
				// not just set <-> unset.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("filter"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttr(resourceName, "context", "filter"),
					resource.TestCheckResourceAttr(resourceName, "source", "params['enabled'] == true"),
				),
			},
		},
	})
}

func TestAccResourceScriptLangUpdate(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_script.test"
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("painless"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttr(resourceName, "lang", "painless"),
					resource.TestCheckResourceAttr(resourceName, "source", "_score * params['multiplier']"),
				),
			},
			{
				// lang must be updatable in-place, and the change must be honored
				// rather than reusing a cached compiled script.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("expression"),
				ConfigVariables:          config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
					resource.TestCheckResourceAttr(resourceName, "lang", "expression"),
					resource.TestCheckResourceAttr(resourceName, "source", "_score * multiplier"),
				),
			},
		},
	})
}

func TestAccResourceScriptExplicitConnection(t *testing.T) {
	endpoints := scriptESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}

	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_script.test_conn"

	apiKey := os.Getenv("ELASTICSEARCH_API_KEY")
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "script_id", scriptID),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "lang", "painless"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#",
			fmt.Sprintf("%d", len(endpoints))),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
	}
	// The fixture mirrors the auth precedence used elsewhere: an api_key, when set,
	// is used exclusively; otherwise username/password are configured.
	if apiKey != "" {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.api_key", apiKey))
	} else {
		checks = append(checks,
			resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.username", username),
			resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.password", password),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"script_id": config.StringVariable(scriptID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(apiKey),
					"username":  config.StringVariable(username),
					"password":  config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func scriptESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}

func checkScriptDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_script" {
			continue
		}

		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
		typedClient := client.GetESClient()
		_, err = typedClient.Core.GetScript(compID.ResourceID).Do(context.Background())
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("script (%s) still exists", compID.ResourceID)
	}
	return nil
}
