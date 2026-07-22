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

// Package managedintegration_test implements acceptance tests for
// elasticstack_fleet_managed_integration (openspec/changes/fleet-managed-integration).
// All tests here require TF_ACC=1 and a live Kibana connection; per design.md
// and the change's own task description, this resource is only functional
// against Elastic Cloud Hosted or Serverless (Kibana >= 9.5.0 for
// managed_integrations; see managedintegration.MinVersion) -- the Fleet
// managed integrations API rejects self-managed stacks, and the resource's
// own topology preflight (topology.go) additionally refuses self-managed stacks
// it can positively identify. Positive acceptance tests gate on Kibana
// managedintegration.MinVersion via clients.KibanaScopedClient.EnforceMinVersion
// (acc_kibana_version_test.go), not the Elasticsearch cluster version. Topology
// (Cloud Hosted/Serverless) gating is separate: skipUnlessConfirmedCloud(t).
// Live-stack preconditions also call skipUnlessManagedIntegrationLiveStack for
// the pinned CSPM package version (acc_package_helpers_test.go).
//
// Every TestAcc* that needs a working managed integration calls
// skipUnlessManagedIntegrationLiveStack then skipUnlessConfirmedCloud right
// after acctest.PreCheck where applicable.
//
// The golden-path package is cloud_security_posture (CSPM), per design.md's
// Open Question 2. Every fixture here uses the "cspm" policy_template with
// the "cloudbeat/cis_aws" input (mapped input key
// "cspm-cloudbeat/cis_aws"). Two non-obvious, empirically-confirmed facts
// about this package's wire shape drove both the fixture shape below and two
// small fixes elsewhere in this change (see update.go... no, see the "genuine
// bugs found" note in this file's package comment continued in models_convert.go
// and schema.go):
//
//  1. CSPM's per-credential-type vars (role_arn, aws.credentials.type,
//     aws.account_type, etc.) are STREAM-level vars (under
//     inputs.<key>.streams."cloud_security_posture.findings".vars), not
//     input-level vars, even though the package's own registry metadata
//     lists them under the input's own `vars` array. Sending them under
//     inputs.<key>.vars directly is rejected by the create API with
//     "Variable cspm-cloudbeat/cis_aws:<name> not found".
//  2. Kibana enforces "required_vars" groups per credential type -- e.g.
//     aws.credentials.type = "assume_role" requires (only) aws.account_type;
//     "cloud_connectors" additionally requires aws.credentials.external_id.
//     Sending an incomplete group produces a 400 listing every group's
//     unmet requirements.
package managedintegration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// cspmPackageVersion pins the cloud_security_posture package version used by
// every fixture in this file. It was 3.4.0 (installed) at the time this
// change's acceptance tests were written and run against a live Kibana
// 9.4.3 Cloud Hosted deployment (see design.md's Decision 3 spike, which
// used the same version). Centralizing it here means a future package
// upgrade only needs one edit.
const cspmPackageVersion = "3.4.0"

const testResourceName = "elasticstack_fleet_managed_integration.test"

var regexpDefaultSpacePrefix = regexp.MustCompile(`^default/`)

// TestAccResourceManagedIntegration covers the default-space full lifecycle (create,
// read, update, import, destroy) against live managed_integrations APIs.
func TestAccResourceManagedIntegration(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	baseVars := config.Variables{
		"policy_name":     config.StringVariable(policyName),
		"package_version": config.StringVariable(cspmPackageVersion),
	}

	// capturedUpdatedAt is populated by the "update_vars" step's Check below
	// and compared against in the "update_flag_only" step that follows it --
	// see that step's comment for what this proves.
	var capturedUpdatedAt string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				// Create.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "name", policyName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Managed integration CSPM Test Policy"),
					resource.TestCheckResourceAttr(testResourceName, "namespace", "default"),
					resource.TestCheckResourceAttr(testResourceName, "policy_template", "cspm"),
					resource.TestCheckResourceAttr(testResourceName, "package.name", "cloud_security_posture"),
					resource.TestCheckResourceAttr(testResourceName, "package.version", cspmPackageVersion),
					resource.TestCheckResourceAttrSet(testResourceName, "package.title"),
					resource.TestCheckResourceAttrSet(testResourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
					resource.TestMatchResourceAttr(testResourceName, "id", regexpDefaultSpacePrefix),
					testCheckJSONSubset("vars_json", map[string]any{
						"posture":    "cspm",
						"deployment": "aws",
					}),
					resource.TestCheckResourceAttr(testResourceName, "inputs.cspm-cloudbeat/cis_aws.enabled", "true"),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"role_arn":             "arn:aws:iam::123456789012:role/tf-acc-test-role",
						"aws.credentials.type": "assume_role",
						"aws.account_type":     "single-account",
					}),
					// Other package inputs (kspm-*, cspm-cloudbeat/cis_gcp,
					// cspm-cloudbeat/cis_azure, vuln_mgmt-*) must NOT leak into
					// state -- see this file's package comment and
					// models_convert.go's inputsKnownKeySet/populateInputsModel
					// fix for why the raw API response otherwise includes all
					// of them.
					resource.TestCheckResourceAttr(testResourceName, "inputs.%", "1"),
					// "Extras" and top-level variables coverage: these three
					// attributes are spec'd as Optional, updatable in-place
					// (spec.md's "Extras" and "Variables and inputs" sections)
					// but were previously only ever set in
					// ImportStateVerifyIgnore lists, never actually configured
					// or asserted by value.
					resource.TestCheckResourceAttr(testResourceName, "var_group_selections.deployment", "aws"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.%", "1"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.env.string_value", "test"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.0", "logs-custom-*"),
				),
			},
			{
				// Update description in-place.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_description"),
				ConfigVariables:          baseVars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated managed integration CSPM Test Policy"),
				),
			},
			{
				// Update an input's stream vars in-place (also exercises Task
				// 8.7's assertion via the ConfigPlanChecks below; see
				// TestAccResourceManagedIntegration_InputsUpdateInPlace for a
				// standalone, minimal repro of just this scenario).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_vars"),
				ConfigVariables:          baseVars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated managed integration CSPM Test Policy"),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"role_arn":             "arn:aws:iam::123456789012:role/tf-acc-test-role-updated",
						"aws.credentials.type": "assume_role",
						"aws.account_type":     "organization-account",
					}),
					// This step also adds a second global_data_tags entry (see
					// testdata/.../update_vars/main.tf), exercising the
					// in-place update path for global_data_tags (spec.md:
					// "updatable in-place"), not just create-time coverage.
					testCheckJSONSubset("vars_json", map[string]any{
						"posture":    "cspm",
						"deployment": "gcp",
					}),
					resource.TestCheckResourceAttr(testResourceName, "var_group_selections.deployment", "aws"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.%", "2"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.env.string_value", "test"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.team.string_value", "security"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.0", "logs-custom-*"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.1", "metrics-acc-*"),
					testCheckManagedIntegrationUpdateExtrasPersisted(
						testResourceName,
						map[string]any{"posture": "cspm", "deployment": "gcp"},
						nil,
						[]string{"logs-custom-*", "metrics-acc-*"},
					),
					resource.TestCheckResourceAttrWith(testResourceName, "updated_at", func(value string) error {
						capturedUpdatedAt = value
						return nil
					}),
				),
			},
			{
				// Real-world regression proof for the update.go
				// onlyCreateOnlyFlagsChanged short-circuit (see that function's
				// doc comment): this config is byte-for-byte identical to the
				// "update_vars" step's above except for adding
				// skip_topology_check = true, a create-only flag that
				// updateAgentlessPolicy never sends to the Fleet API. Terraform
				// still invokes Update (skip_topology_check is not
				// RequiresReplace, so this step still shows a plan diff), but
				// updateAgentlessPolicy itself must recognize the change is
				// confined to a create-only flag and skip the GET+PUT round
				// trip entirely.
				//
				// onlyCreateOnlyFlagsChanged (update.go) never compares
				// created_at/updated_at at all -- both are purely
				// server-Computed, so a naive prior.UpdatedAt.Equal(plan.UpdatedAt)
				// comparison would be permanently defeated here: updated_at
				// has no UseStateForUnknown plan modifier (schema.go -- it
				// legitimately changes on every real Update, so promising it
				// won't would be wrong) and so is Unknown in this plan
				// regardless of whether this change is create-only-flags-only.
				// See update_test.go's TestOnlyCreateOnlyFlagsChanged for the
				// equivalent unit-level proof.
				//
				// Kibana bumps managed_integrations updated_at on every real PUT, so
				// an unchanged updated_at after this apply is direct empirical
				// evidence against a live deployment that no API call was made
				// -- not just that the provider's local bookkeeping thinks so.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_flag_only"),
				ConfigVariables:          baseVars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "skip_topology_check", "true"),
					resource.TestCheckResourceAttrWith(testResourceName, "updated_at", func(value string) error {
						if value != capturedUpdatedAt {
							return fmt.Errorf(
								"updated_at changed from %q to %q after a create-only-flag-only change; "+
									"this means updateAgentlessPolicy's onlyCreateOnlyFlagsChanged short-circuit "+
									"did not fire and a real GET+PUT round trip was made against the Fleet API",
								capturedUpdatedAt, value,
							)
						}
						return nil
					}),
				),
			},
			{
				// Import by composite ID (default-space case): the resource's own
				// `id` is already "default/<policy_id>".
				// ConfigDirectory matches the immediately preceding
				// "update_flag_only" step (not "update_vars") so this step's
				// implicit re-apply is a no-op and doesn't reintroduce a real
				// GET+PUT before the import itself.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_flag_only"),
				ConfigVariables:          baseVars,
				ResourceName:             testResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"kibana_connection",
					"force",
					"force_delete",
					"create_dataset_templates",
					"skip_topology_check",
					// policy_template is create-only (absent from GET); import starts
					// without prior config so it remains null until set — see
					// spec.md "policy_template create-only preservation".
					"policy_template",
					// vars_json carries an internal contextual-normalization
					// marker (see policyshape's JSONWithContextualDefaultsType)
					// that legitimately differs between a freshly-imported read
					// and the live-managed resource's own state; the same
					// attribute is ignored for the same reason in
					// integration_policy's own import tests (see
					// TestAccResourceIntegrationPolicySecrets in
					// internal/fleet/integration_policy/acc_test.go).
					"vars_json",
					// inputs may differ after import when the API returns more keys
					// than config tracks (inputsKnownKeySet); see package comment.
					"inputs",
				},
			},
		},
	})
}

// TestAccResourceManagedIntegration_NonDefaultSpace covers import via composite
// "<space_id>/<policy_id>" in a non-default Kibana space.
// spec.md's "Import by composite ID" scenario is explicitly stated in terms
// of a non-"default" space, so this creates a real Kibana space and confirms
// both that space_ids round-trips through Create and that the composite ID
// is exactly "<space_id>/<policy_id>".
func TestAccResourceManagedIntegration_NonDefaultSpace(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	policyName := fmt.Sprintf("managed-integration-%s", suffix)
	spaceID := fmt.Sprintf("managed-integration-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(cspmPackageVersion),
					"space_id":        config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(testResourceName, "space_ids.*", spaceID),
					resource.TestMatchResourceAttr(testResourceName, "id", regexp.MustCompile(fmt.Sprintf(`^%s/`, regexp.QuoteMeta(spaceID)))),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(cspmPackageVersion),
					"space_id":        config.StringVariable(spaceID),
				},
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"kibana_connection",
					"force",
					"force_delete",
					"create_dataset_templates",
					"skip_topology_check",
					"policy_template",
					"vars_json",
					"inputs",
				},
			},
		},
	})
}

// TestAccResourceManagedIntegration_ForceDelete verifies force_delete round-trips and
// that ?force=true delete succeeds end-to-end against live managed_integrations.
// was investigated and found impractical: the DELETE endpoint's conflict
// path is documented (design.md Decision 5) as triggering when the hidden
// managed agent policy is "still provisioning, or has associated agents" --
// both require either winning a real race against Fleet's own agentless
// provisioning workflow (not something a test can reliably trigger on
// demand) or actually enrolling a live Elastic Agent against the policy
// (out of scope: agentless policies exist precisely so no agent host is
// needed, and this repo's acceptance tests don't provision arbitrary compute
// to enroll one). Absent a reliable trigger, this test instead verifies the
// two things that ARE practically verifiable end-to-end against a real
// deployment:
//  1. force_delete = true does not break the normal (non-conflict) delete
//     path -- ?force=true is accepted by a real Kibana and the policy is
//     genuinely removed (checkResourceManagedIntegrationDestroy would fail the
//     test otherwise).
//  2. The attribute itself round-trips through state.
//
// The conflict-diagnostic-hint logic itself (conflictHintDiagnostics,
// delete.go) already has unit-level coverage independent of this test: see
// TestConflictHintDiagnostics in delete_test.go.
func TestAccResourceManagedIntegration_ForceDelete(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(cspmPackageVersion),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "force_delete", "true"),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_CloudConnector wires a real cloud connector
// (POST /api/fleet/cloud_connectors) and configures a plaintext stream
// aws.credentials.external_id. Terraform state must retain the plaintext after
// create/read even when the Fleet API stores a secret reference server-side.
func TestAccResourceManagedIntegration_CloudConnector(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)
	acctest.PreCheck(t)

	externalID := fmt.Sprintf("tf-acc-ext-%s", sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum))
	connectorID, cleanupConnector := createTestCloudConnector(t, externalID)
	t.Cleanup(cleanupConnector)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":           config.StringVariable(policyName),
					"package_version":       config.StringVariable(cspmPackageVersion),
					"cloud_connector_id":    config.StringVariable(connectorID),
					"external_id_plaintext": config.StringVariable(externalID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.cloud_connector_id", connectorID),
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.target_csp", "aws"),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"aws.credentials.external_id": externalID,
					}),
					testCheckCloudConnectorPersisted(testResourceName, connectorID),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_NameUpdateInPlace verifies in-place `name`
// updates plan as Update and persist in Terraform state and on the Fleet API.
func TestAccResourceManagedIntegration_NameUpdateInPlace(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	suffix := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	firstName := "mi-" + suffix
	renamedName := "mi-" + suffix + "-renamed"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(firstName),
					"package_version": config.StringVariable(cspmPackageVersion),
				},
				Check: resource.TestCheckResourceAttr(testResourceName, "name", firstName),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(renamedName),
					"package_version": config.StringVariable(cspmPackageVersion),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "name", renamedName),
					testCheckManagedIntegrationNamePersisted(testResourceName, renamedName),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_InputsUpdateInPlace verifies input stream vars
// update in-place via managed_integrations PUT (full-replace body from plan).
func TestAccResourceManagedIntegration_InputsUpdateInPlace(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	vars := func() config.Variables {
		return config.Variables{
			"policy_name":     config.StringVariable(policyName),
			"package_version": config.StringVariable(cspmPackageVersion),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars(),
				Check: testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
					"aws.account_type": "single-account",
				}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"aws.account_type": "organization-account",
					}),
					testCheckManagedIntegrationStreamVarString(testResourceName, "aws.account_type", "organization-account"),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_PackageVersionUpdate verifies package.version
// updates in-place via managed_integrations PUT against a live stack.
func TestAccResourceManagedIntegration_PackageVersionUpdate(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	fromVersion, toVersion := resolveCSPMInPlaceVersionUpgradePair(t)
	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	var capturedPolicyID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(fromVersion),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "package.version", fromVersion),
					resource.TestCheckResourceAttrSet(testResourceName, "policy_id"),
					resource.TestCheckResourceAttrWith(testResourceName, "policy_id", func(value string) error {
						capturedPolicyID = value
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(toVersion),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "package.version", toVersion),
					resource.TestCheckResourceAttr(testResourceName, "policy_id", capturedPolicyID),
					testCheckManagedIntegrationPackageVersionPersisted(testResourceName, toVersion),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_GlobalDataTagsNumber verifies number_value
// tags round-trip through create against a live stack.
func TestAccResourceManagedIntegration_GlobalDataTagsNumber(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(cspmPackageVersion),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.%", "2"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.env.string_value", "test"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.priority.number_value", "42"),
					testCheckManagedIntegrationGlobalDataTagsPersisted(testResourceName,
						map[string]string{"env": "test"},
						map[string]float64{"priority": 42},
					),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_VersionSkipGating verifies apply fails on
// Kibana stacks older than managedintegration.MinVersion (9.5.0). Runs only when
// Kibana /api/status reports a version below that floor.
func TestAccResourceManagedIntegration_VersionSkipGating(t *testing.T) {
	skipIfKibanaMeetsManagedIntegrationMinVersion(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"package_version": config.StringVariable(cspmPackageVersion),
				},
				ExpectError: regexp.MustCompile(`Fleet managed integrations require Elastic Stack v9\.5\.0`),
			},
		},
	})
}

// TestAccResourceManagedIntegration_CloudConnectorUpdate verifies a non-connector
// field update preserves cloud_connector association on the Fleet API.
func TestAccResourceManagedIntegration_CloudConnectorUpdate(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)
	acctest.PreCheck(t)

	externalID := fmt.Sprintf("tf-acc-ext-%s", sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum))
	connectorID, cleanupConnector := createTestCloudConnector(t, externalID)
	t.Cleanup(cleanupConnector)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	baseVars := config.Variables{
		"policy_name":           config.StringVariable(policyName),
		"package_version":       config.StringVariable(cspmPackageVersion),
		"cloud_connector_id":    config.StringVariable(connectorID),
		"external_id_plaintext": config.StringVariable(externalID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars,
				Check: resource.ComposeTestCheckFunc(
					testCheckCloudConnectorPersisted(testResourceName, connectorID),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"aws.credentials.external_id": externalID,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_description"),
				ConfigVariables:          baseVars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated description after cloud_connector association"),
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.cloud_connector_id", connectorID),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"aws.credentials.external_id": externalID,
					}),
					testCheckCloudConnectorPersisted(testResourceName, connectorID),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_CloudConnectorRequiresReplace verifies the
// object-level RequiresReplace plan modifier on cloud_connector (any sub-field
// change forces destroy-before-create).
func TestAccResourceManagedIntegration_CloudConnectorRequiresReplace(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)
	acctest.PreCheck(t)

	externalID := fmt.Sprintf("tf-acc-ext-%s", sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum))
	connectorID, cleanupConnector := createTestCloudConnector(t, externalID)
	t.Cleanup(cleanupConnector)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	baseVars := config.Variables{
		"policy_name":           config.StringVariable(policyName),
		"package_version":       config.StringVariable(cspmPackageVersion),
		"cloud_connector_id":    config.StringVariable(connectorID),
		"external_id_plaintext": config.StringVariable(externalID),
		"cloud_connector_name":  config.StringVariable("tf-acc-connector-name-a"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_cloud_connector_name"),
				ConfigVariables: config.Variables{
					"policy_name":           config.StringVariable(policyName),
					"package_version":       config.StringVariable(cspmPackageVersion),
					"cloud_connector_id":    config.StringVariable(connectorID),
					"external_id_plaintext": config.StringVariable(externalID),
					"cloud_connector_name":  config.StringVariable("tf-acc-connector-name-b"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

// TestAccResourceManagedIntegration_ConditionRoundTrip verifies input/stream
// condition expressions round-trip through create, read, and in-place update.
func TestAccResourceManagedIntegration_ConditionRoundTrip(t *testing.T) {
	skipUnlessManagedIntegrationLiveStack(t)
	skipUnlessConfirmedCloud(t)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	initialInputCondition := "host.os.family == 'linux'"
	initialStreamCondition := "data_stream.dataset == 'audit'"
	updatedInputCondition := "host.os.family == 'windows'"
	updatedStreamCondition := "data_stream.dataset == 'logs'"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceManagedIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":      config.StringVariable(policyName),
					"package_version":  config.StringVariable(cspmPackageVersion),
					"input_condition":  config.StringVariable(initialInputCondition),
					"stream_condition": config.StringVariable(initialStreamCondition),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "inputs.cspm-cloudbeat/cis_aws.condition", initialInputCondition),
					resource.TestCheckResourceAttr(testResourceName, "inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.condition", initialStreamCondition),
					testCheckManagedIntegrationConditionsPersisted(testResourceName, initialInputCondition, initialStreamCondition),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":      config.StringVariable(policyName),
					"package_version":  config.StringVariable(cspmPackageVersion),
					"input_condition":  config.StringVariable(updatedInputCondition),
					"stream_condition": config.StringVariable(updatedStreamCondition),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "inputs.cspm-cloudbeat/cis_aws.condition", updatedInputCondition),
					resource.TestCheckResourceAttr(testResourceName, "inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.condition", updatedStreamCondition),
					testCheckManagedIntegrationConditionsPersisted(testResourceName, updatedInputCondition, updatedStreamCondition),
				),
			},
		},
	})
}

func checkResourceManagedIntegrationDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_managed_integration" {
			continue
		}

		fc := client.GetFleetClient()

		policyID := rs.Primary.Attributes["policy_id"]
		spaceID := "default"
		if id, diags := clients.CompositeIDFromStr(rs.Primary.ID); !diags.HasError() && id != nil {
			spaceID = id.ClusterID
		}

		item, diags := fleetclient.ReadManagedIntegration(context.Background(), fc, spaceID, policyID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if item != nil {
			return fmt.Errorf("managed integration id=%v still exists, but it should have been removed", policyID)
		}
	}
	return nil
}

// testCheckJSONSubset asserts that the JSON object stored in resourceName's
// attr attribute contains at least the keys/values in expected (extra keys
// -- e.g. package-populated defaults -- are ignored). This is used instead
// of an exact-string TestCheckResourceAttr for vars_json/inputs.*.vars
// attributes because Fleet fills in additional package-declared vars (with
// package-specific default values) alongside whatever the config sets, and
// hand-deriving the exact expected canonical JSON for every var CSPM
// declares would be brittle against package updates.
func testCheckJSONSubset(attr string, expected map[string]any) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testResourceName]
		if !ok || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", testResourceName)
		}
		raw, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %s not set on %s", attr, testResourceName)
		}
		var actual map[string]any
		if err := json.Unmarshal([]byte(raw), &actual); err != nil {
			return fmt.Errorf("failed to parse %s.%s as JSON: %w (value: %s)", testResourceName, attr, err, raw)
		}
		for k, want := range expected {
			got, ok := actual[k]
			if !ok {
				return fmt.Errorf("key %q missing from %s.%s; got: %v", k, testResourceName, attr, actual)
			}
			gotBytes, _ := json.Marshal(got)
			wantBytes, _ := json.Marshal(want)
			if string(gotBytes) != string(wantBytes) {
				return fmt.Errorf("key %q in %s.%s: got %s, want %s", k, testResourceName, attr, gotBytes, wantBytes)
			}
		}
		return nil
	}
}

// createTestCloudConnector creates a real AWS cloud connector via POST
// /api/fleet/cloud_connectors (no Terraform resource yet). externalIDPlaintext
// is stored as a password/text var on the connector and must match the managed
// integration stream var when using aws.credentials.type = cloud_connectors.
func createTestCloudConnector(t *testing.T, externalIDPlaintext string) (string, func()) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client for cloud connector fixture: %v", err)
	}
	fc := client.GetFleetClient()
	ctx := context.Background()

	externalIDVar := kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties{}
	if err := externalIDVar.FromPostFleetCloudConnectorsJSONBodyVars3(kbapi.PostFleetCloudConnectorsJSONBodyVars3{
		Type: "password",
		Value: func() kbapi.PostFleetCloudConnectorsJSONBody_Vars_3_Value {
			var v kbapi.PostFleetCloudConnectorsJSONBody_Vars_3_Value
			if err := v.FromPostFleetCloudConnectorsJSONBodyVars3Value0(externalIDPlaintext); err != nil {
				t.Fatalf("failed to build cloud connector external_id plaintext: %v", err)
			}
			return v
		}(),
	}); err != nil {
		t.Fatalf("failed to build cloud connector external_id var: %v", err)
	}

	// A bare-string arm (FromPostFleetCloudConnectorsJSONBodyVars0) is
	// accepted by the JSON schema but is NOT recognized by Kibana's
	// cloud-connector validation -- empirically confirmed: it fails with
	// "Package policy must contain role_arn variable" even though the key is
	// present. The structured {type, value} arm (matching what
	// POST /api/fleet/cloud_connectors actually round-trips) is required.
	roleArnVar := kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties{}
	if err := roleArnVar.FromPostFleetCloudConnectorsJSONBodyVars3(kbapi.PostFleetCloudConnectorsJSONBodyVars3{
		Type: "text",
		Value: func() kbapi.PostFleetCloudConnectorsJSONBody_Vars_3_Value {
			var v kbapi.PostFleetCloudConnectorsJSONBody_Vars_3_Value
			if err := v.FromPostFleetCloudConnectorsJSONBodyVars3Value0("arn:aws:iam::123456789012:role/tf-acc-test-role"); err != nil {
				t.Fatalf("failed to build cloud connector role_arn value: %v", err)
			}
			return v
		}(),
	}); err != nil {
		t.Fatalf("failed to build cloud connector role_arn var: %v", err)
	}

	accountType := kbapi.PostFleetCloudConnectorsJSONBodyAccountTypeSingleAccount
	name := fmt.Sprintf("tf-acc-managed-integration-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	body := kbapi.PostFleetCloudConnectorsJSONRequestBody{
		Name:          name,
		CloudProvider: kbapi.Aws,
		AccountType:   &accountType,
		Vars: map[string]kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties{
			"role_arn":    roleArnVar,
			"external_id": externalIDVar,
		},
	}

	resp, err := fc.API.PostFleetCloudConnectorsWithResponse(ctx, body)
	if err != nil {
		t.Fatalf("failed to create test cloud connector fixture: %v", err)
	}
	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		t.Fatalf("failed to create test cloud connector fixture: status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	connectorID := resp.JSON200.Item.Id

	cleanup := func() {
		delResp, err := fc.API.DeleteFleetCloudConnectorsCloudconnectoridWithResponse(context.Background(), connectorID, nil)
		if err != nil {
			t.Logf("failed to delete test cloud connector fixture %s: %v", connectorID, err)
			return
		}
		if delResp.StatusCode() >= 300 && delResp.StatusCode() != 404 {
			t.Logf("failed to delete test cloud connector fixture %s: status %d: %s", connectorID, delResp.StatusCode(), string(delResp.Body))
		}
	}
	return connectorID, cleanup
}
