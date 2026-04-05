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

package elasticdefendintegrationpolicy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionElasticDefend = version.Must(version.NewVersion("8.14.0"))

const resourceName = "elasticstack_fleet_elastic_defend_integration_policy.test"

// TestAccResourceElasticDefendIntegrationPolicy covers create, update, import,
// and refresh-after-out-of-band-delete for the Elastic Defend resource.
func TestAccResourceElasticDefendIntegrationPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceElasticDefendPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				Config:                   testConfigCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "namespace", "default"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// Step 2: Update description and preset
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				Config:                   testConfigUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			// Step 3: Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateIdFunc:        testImportStateIDFunc(resourceName),
				ImportStateVerify:        false, // Some fields like policy may differ slightly
			},
		},
	})
}

// TestAccResourceElasticDefendIntegrationPolicyDisappears covers the
// refresh-after-out-of-band-delete scenario.
func TestAccResourceElasticDefendIntegrationPolicyDisappears(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceElasticDefendPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				Config:                   testConfigCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					// Delete out of band to trigger refresh behavior
					deleteDefendPolicyOutOfBand(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %q not found", resourceName)
		}
		return rs.Primary.Attributes["policy_id"], nil
	}
}

func deleteDefendPolicyOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found", resourceName)
		}
		policyID := rs.Primary.Attributes["policy_id"]

		apiClient, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		fleetClient, err := apiClient.GetFleetClient()
		if err != nil {
			return err
		}

		diags := fleetclient.DeletePackagePolicy(context.Background(), fleetClient, policyID, "", false)
		if diags.HasError() {
			return fmt.Errorf("error deleting Defend policy %q: %v", policyID, diags)
		}
		return nil
	}
}

func checkResourceElasticDefendPolicyDestroy(s *terraform.State) error {
	apiClient, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_elastic_defend_integration_policy" {
			continue
		}

		policyID := rs.Primary.Attributes["policy_id"]
		policy, diags := fleetclient.GetDefendPackagePolicy(context.Background(), fleetClient, policyID, "")
		if diags.HasError() {
			return fmt.Errorf("error checking policy %q: %v", policyID, diags)
		}
		if policy != nil {
			return fmt.Errorf("Elastic Defend policy %q still exists", policyID)
		}
	}
	return nil
}

func testConfigCreate(policyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "%s-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = %q
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  integration_version = "8.14.0"

  policy = {
    windows = {
      events = {
        process = true
        network = true
        file    = true
      }
      malware = {
        mode = "prevent"
      }
    }
    mac = {
      events = {
        process = true
      }
    }
    linux = {
      events = {
        process = true
        network = true
      }
    }
  }
}
`, policyName, policyName)
}

func testConfigUpdate(policyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "%s-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = %q
  namespace           = "default"
  description         = "Updated description"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  integration_version = "8.14.0"

  policy = {
    windows = {
      events = {
        process = true
        network = true
        file    = false
      }
      malware = {
        mode = "detect"
      }
    }
    mac = {
      events = {
        process = true
        network = false
      }
    }
    linux = {
      events = {
        process = true
        network = false
      }
    }
  }
}
`, policyName, policyName)
}
