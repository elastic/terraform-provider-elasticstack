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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionElasticDefend = version.Must(version.NewVersion("8.14.0"))
var minVersionElasticDefendSpaceIDs = version.Must(version.NewVersion("9.1.0"))

const (
	resourceName            = "elasticstack_fleet_elastic_defend_integration_policy.test"
	agentPolicyResourceName = "elasticstack_fleet_agent_policy.test"
)

// TestAccResourceElasticDefendIntegrationPolicy covers create, update, import,
// and description round-trip behavior for the Elastic Defend resource.
func TestAccResourceElasticDefendIntegrationPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceElasticDefendPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with full policy settings including events, popups,
			// antivirus_registration, and attack_surface_reduction.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "namespace", "default"),
					resource.TestCheckResourceAttrPair(resourceName, "agent_policy_id", agentPolicyResourceName, "policy_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "integration_version", "8.14.0"),
					resource.TestCheckResourceAttr(resourceName, "preset", "EDRComplete"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testCheckSpaceIDsIfSupported(resourceName),
					// Windows events
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.network", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.file", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.dns", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.dll_and_driver_load", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.registry", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.security", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.authentication", "false"),
					// Windows malware
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.blocklist", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.notify_user", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.on_write_scan", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.ransomware.mode", "prevent"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.windows.ransomware.supported"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.memory_protection.mode", "detect"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.windows.memory_protection.supported"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.behavior_protection.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.behavior_protection.reputation_service", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.windows.behavior_protection.supported"),
					// Windows popup
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.message", "Malware detected"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.enabled", "false"),
					// Windows antivirus_registration
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.mode", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.enabled", "true"),
					// Windows attack_surface_reduction
					resource.TestCheckResourceAttr(resourceName, "policy.windows.attack_surface_reduction.credential_hardening.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.logging.file", "info"),
					// Mac events
					resource.TestCheckResourceAttr(resourceName, "policy.mac.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.events.file", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.malware.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.malware.blocklist", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.malware.on_write_scan", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.malware.notify_user", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.memory_protection.mode", "prevent"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.mac.memory_protection.supported"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.behavior_protection.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.behavior_protection.reputation_service", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.mac.behavior_protection.supported"),
					// Mac popup
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.message", "Mac malware alert"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.logging.file", "warning"),
					// Linux events
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.network", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.file", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.session_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.tty_io", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.malware.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.malware.blocklist", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.memory_protection.mode", "prevent"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.linux.memory_protection.supported"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.behavior_protection.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.behavior_protection.reputation_service", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "policy.linux.behavior_protection.supported"),
					// Linux popup
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.message", "Linux malware alert"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.logging.file", "warning"),
				),
			},
			// Step 2: Update description and nested policy settings
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttrPair(resourceName, "agent_policy_id", agentPolicyResourceName, "policy_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "integration_version", "8.14.0"),
					resource.TestCheckResourceAttr(resourceName, "preset", "EDRComplete"),
					testCheckSpaceIDsIfSupported(resourceName),
					// Windows events
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.network", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.file", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.dns", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.dll_and_driver_load", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.registry", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.security", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.events.authentication", "true"),
					// Windows malware
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.blocklist", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.notify_user", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.malware.on_write_scan", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.ransomware.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.memory_protection.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.behavior_protection.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.behavior_protection.reputation_service", "false"),
					// Windows popup
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.message", "Ransomware blocked"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.message", "Memory alert"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.enabled", "false"),
					// Windows antivirus_registration
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.mode", "sync_with_malware_prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.enabled", "false"),
					// Windows attack_surface_reduction
					resource.TestCheckResourceAttr(resourceName, "policy.windows.attack_surface_reduction.credential_hardening.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.logging.file", "error"),
					// Mac events
					resource.TestCheckResourceAttr(resourceName, "policy.mac.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.events.network", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.events.file", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.malware.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.memory_protection.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.behavior_protection.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.behavior_protection.reputation_service", "false"),
					// Mac popup
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.message", "Mac memory alert"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.logging.file", "error"),
					// Linux events
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.process", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.network", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.file", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.session_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.events.tty_io", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.malware.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.malware.blocklist", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.memory_protection.mode", "detect"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.behavior_protection.mode", "prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.behavior_protection.reputation_service", "false"),
					// Linux popup
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.message", "Linux memory alert"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.logging.file", "error"),
				),
			},
			// Step 3: Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testImportStateIDFunc(resourceName),
				ImportStateVerify: true,
				// description is Optional-only (unmanaged): import starts with a blank
				// model (all-null), so it stays null even when Kibana has a value.
				ImportStateVerifyIgnore: []string{"force", "description"},
			},
			// Step 4: Remove description and omit enabled to verify defaults are restored.
			// description is Optional-only (unmanaged): omitting it keeps null in state
			// regardless of the server value, matching the repo pattern for unmanaged fields.
			// Also exercises the force attribute.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_removed"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "force", "true"),
					testCheckSpaceIDsIfSupported(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.message", "Ransomware blocked"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.mode", "sync_with_malware_prevent"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.attack_surface_reduction.credential_hardening.enabled", "false"),
				),
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
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

// TestAccResourceElasticDefendIntegrationPolicyPopupDefaults verifies that the
// Windows popup block is populated with computed default values (message="",
// enabled=false) when not explicitly configured in the Terraform config, and
// that antivirus_registration and attack_surface_reduction also apply their
// schema defaults.
func TestAccResourceElasticDefendIntegrationPolicyPopupDefaults(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceElasticDefendPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					// Windows popup is Computed+Default at the parent level; all sub-block
					// values default to message="" and enabled=false when not overridden.
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.ransomware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.popup.behavior_protection.enabled", "false"),
					// Windows antivirus_registration defaults
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.mode", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "policy.windows.antivirus_registration.enabled", "false"),
					// Windows attack_surface_reduction defaults
					resource.TestCheckResourceAttr(resourceName, "policy.windows.attack_surface_reduction.credential_hardening.enabled", "false"),
					// Mac popup defaults
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.mac.popup.behavior_protection.enabled", "false"),
					// Linux popup defaults
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.malware.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.memory_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.message", ""),
					resource.TestCheckResourceAttr(resourceName, "policy.linux.popup.behavior_protection.enabled", "false"),
				),
			},
		},
	})
}

// TestAccResourceElasticDefendIntegrationPolicyKibanaConnection verifies that
// the resource can be created when an entity-local kibana_connection block is
// supplied instead of relying on the provider-level Kibana configuration.
func TestAccResourceElasticDefendIntegrationPolicyKibanaConnection(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		CheckDestroy: checkResourceElasticDefendPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefend),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"policy_name": config.StringVariable(policyName),
				}),
				Check: resource.ComposeTestCheckFunc(
					append([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "name", policyName),
						resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
						resource.TestCheckResourceAttr(resourceName, "kibana_connection.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.endpoints.#", "1"),
						resource.TestCheckResourceAttrSet(resourceName, "kibana_connection.0.endpoints.0"),
					}, acctest.KibanaConnectionAuthChecks(resourceName)...)...,
				),
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

		apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}

		fleetClient, err := apiClient.GetFleetClient()
		if err != nil {
			return err
		}

		spaceID := rs.Primary.Attributes["space_ids.0"]
		diags := fleetclient.DeletePackagePolicy(context.Background(), fleetClient, policyID, spaceID, false)
		if diags.HasError() {
			return fmt.Errorf("error deleting Defend policy %q: %v", policyID, diags)
		}
		return nil
	}
}

func testCheckSpaceIDsIfSupported(resourceAddress string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		unsupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionElasticDefendSpaceIDs)()
		if err != nil {
			return err
		}
		if unsupported {
			return nil
		}

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceAddress, "space_ids.#", "1"),
			resource.TestCheckTypeSetElemAttr(resourceAddress, "space_ids.*", "default"),
		)(state)
	}
}

func checkResourceElasticDefendPolicyDestroy(s *terraform.State) error {
	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
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
