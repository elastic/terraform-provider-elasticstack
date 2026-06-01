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

package connectors_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue634 is a regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/634
//
// Importing a .webhook connector that has no secrets sets "secrets" to null in
// state rather than "{}" (empty object). ImportStateVerify (without ignoring
// "secrets") detects this: the pre-import state has secrets="{}" but the
// imported state has secrets=null, causing a verification mismatch.
//
// The fix should make populateFromAPI preserve secrets="{}" after import so
// that ImportStateVerify passes without ignoring the secrets attribute.
func TestAccReproduceIssue634(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Create a webhook connector with secrets = "{}".
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".webhook"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "secrets", "{}"),
				),
			},
			{
				// Import and verify all attributes, ignoring the config field only
				// (Kibana adds __tf_provider_context which was not in the original
				// config and is expected). secrets is intentionally NOT ignored:
				// the bug causes secrets to be null after import (was "{}" before),
				// so the ImportStateVerify diff will mention "secrets".
				// ExpectError matches that diff, so the step passes when the bug is
				// present and fails (no matching error) when the bug is fixed.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
				ResourceName:            "elasticstack_kibana_action_connector.test",
				ExpectError:             regexp.MustCompile(`"secrets"`),
			},
		},
	})
}
