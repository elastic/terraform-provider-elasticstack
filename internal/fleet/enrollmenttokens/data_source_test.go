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

package enrollmenttokens_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionEnrollmentTokens = version.Must(version.NewVersion("8.6.0"))
var minVersionEnrollmentTokensSpaceID = version.Must(version.NewVersion("9.1.0"))

func TestAccDataSourceEnrollmentTokens(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEnrollmentTokens, versionutils.FlavorAny)

	policyID := uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_id": config.StringVariable(policyID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "policy_id", policyID),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "id", policyID),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.policy_id", policyID),
					testCheckTokensMinCount("data.elasticstack_fleet_enrollment_tokens.test", 1),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.api_key"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.api_key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.created_at"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.name"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.active", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrollmentTokensNoPolicyID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEnrollmentTokens, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.all", "id"),
					testCheckTokensMinCount("data.elasticstack_fleet_enrollment_tokens.all", 1),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.all", "tokens.0.key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.all", "tokens.0.api_key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.all", "tokens.0.created_at"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.all", "tokens.0.policy_id"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrollmentTokensSpaceID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEnrollmentTokensSpaceID, versionutils.FlavorAny)

	spaceID := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceName := "Test Space " + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_id"),
				ConfigVariables: config.Variables{
					"space_id":   config.StringVariable(spaceID),
					"space_name": config.StringVariable(spaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "space_id", spaceID),
					testCheckTokensMinCount("data.elasticstack_fleet_enrollment_tokens.test", 1),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.policy_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.api_key_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.created_at"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.name"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.active", "true"),
				),
			},
		},
	})
}

func testCheckTokensMinCount(resourceName string, minCount int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		rawCount, ok := rs.Primary.Attributes["tokens.#"]
		if !ok {
			return fmt.Errorf("resource %q has no tokens count in state", resourceName)
		}

		count, err := strconv.Atoi(rawCount)
		if err != nil {
			return fmt.Errorf("resource %q has invalid tokens count %q: %w", resourceName, rawCount, err)
		}

		if count < minCount {
			return fmt.Errorf("resource %q expected at least %d tokens, got %d", resourceName, minCount, count)
		}

		return nil
	}
}

func checkResourceAgentPolicyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_policy" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID, "")
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy != nil {
			return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
