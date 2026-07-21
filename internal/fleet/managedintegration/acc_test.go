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
// managed integrations API rejects self-managed stacks ("supports_agentless
// is only allowed in serverless and cloud environments"), and the resource's
// own topology preflight (topology.go) additionally refuses self-managed stacks
// it can positively identify. versionutils.SkipIfUnsupported still provides the
// Kibana-version part of the gate (see TestAccResourceManagedIntegration_VersionSkipGating
// below). Every TestAcc* function below that requires a working managed integration
// therefore calls skipUnlessConfirmedCloud(t) (acc_helpers_test.go) right after its
// versionutils.SkipIfUnsupported call: it makes the same GET /api/status
// probe as topology.go, but -- unlike topology.go, which fails open on
// ambiguity to protect a real cloud user's apply -- fails closed, skipping
// the test unless Cloud Hosted/Serverless is positively confirmed. This
// resolves what was previously an open gap here (no environment signal for
// "this is definitely cloud" existed in the repo's acctest package): the
// gap is closed by a test-only, fail-closed mirror of topology.go's
// detection signal, not by new CI infrastructure or a manual opt-in env
// var (both considered and rejected -- see PR #4034).
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
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/managedintegration"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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
					resource.TestCheckResourceAttr(testResourceName, "description", "Agentless CSPM Test Policy"),
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
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Agentless CSPM Test Policy"),
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
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Agentless CSPM Test Policy"),
					testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
						"role_arn":             "arn:aws:iam::123456789012:role/tf-acc-test-role-updated",
						"aws.credentials.type": "assume_role",
						"aws.account_type":     "organization-account",
					}),
					// This step also adds a second global_data_tags entry (see
					// testdata/.../update_vars/main.tf), exercising the
					// in-place update path for global_data_tags (spec.md:
					// "updatable in-place"), not just create-time coverage.
					resource.TestCheckResourceAttr(testResourceName, "var_group_selections.deployment", "aws"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.%", "2"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.env.string_value", "test"),
					resource.TestCheckResourceAttr(testResourceName, "global_data_tags.team.string_value", "security"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "additional_datastreams_permissions.0", "logs-custom-*"),
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
					// policy_template is create-only in practice (not returned on
					// managed_integrations GET) and isn't Computed in schema.go, so a
					// fresh import -- which starts from a blank model with no prior
					// config to preserve it from -- has no way to know it. See spec.md's
					// Import requirement: "Read SHALL populate all state attributes from
					// the Fleet API", which policy_template structurally cannot satisfy.
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
					// inputs legitimately differs after a fresh import: the
					// live-managed resource's state is filtered down to just
					// the keys the config tracks (see
					// models_convert.go's inputsKnownKeySet), but an import
					// starts with no such reference and so faithfully surfaces
					// every input CSPM's create/read responses include across
					// all of its policy templates, per spec.md's Import
					// requirement ("Read SHALL populate all state attributes
					// from the Fleet API"). This is expected, not a bug -- see
					// this file's package comment.
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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
	skipUnlessConfirmedCloud(t)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	policyName := fmt.Sprintf("agentless-policy-%s", suffix)
	spaceID := fmt.Sprintf("agentless-policy-%s", suffix)

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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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

// TestAccResourceManagedIntegration_CloudConnector covers Task 8.5: creating a
// policy with cloud_connector.cloud_connector_id set to a real, pre-existing
// cloud connector (created out-of-band via a raw Fleet API call in this test
// -- there is no elasticstack_fleet_cloud_connector resource yet, per this
// change's design.md "no hard dependency on fleet-cloud-connector" decision)
// and verifying it round-trips.
//
// Per design.md / models_convert.go's populateFromManagedIntegration
// comment, cloud_connector is intentionally NOT re-read from the API on
// Read (it isn't Computed in schema.go, and the read response only exposes a
// bare cloud_connector_id, not the full object) -- so state round-tripping
// it is a Terraform-side (config-preservation) guarantee, not proof the
// value actually reached Kibana. This test additionally reads GET
// /api/fleet/managed_integrations/{id} to confirm cloud_connector_id was
// persisted server-side, which is the part that actually proves the create
// request carried it (see testCheckCloudConnectorPersisted).
//
// Actually wiring the connector so Kibana persists cloud_connector_id
// server-side (empirically confirmed: it does NOT persist the field at all
// when cloud_connector.enabled = false, only when true) requires the
// input's own aws.credentials.type = "cloud_connectors", which in turn
// requires aws.credentials.external_id to be a *valid* secret reference
// (a bare string is rejected outright: "Package policy must contain valid
// external_id secret reference" once a cloud_connector block is present).
// This resource does not implement policyshape's secret-masking
// reconciliation (HandleRespSecrets/HandleReqRespSecrets -- wired up for
// integration_policy but not for managedintegration; see this file's package
// comment), so a bare string for a password-type var would normally trip
// "Provider produced inconsistent result after apply" once Kibana echoes it
// back as a {id,isSecretRef} object. This test sidesteps that gap (which is
// real, but out of scope for Task 8 -- see design.md's Non-Goals) by
// configuring aws.credentials.external_id as the *already-secret-ref-shaped*
// value up front (`{isSecretRef: true, id: <secret-id>}`), obtained from a
// throwaway probe policy exactly as createTestCloudConnector's own fixture
// does. Since the configured value already matches what Kibana will echo
// back byte-for-byte, there is no drift to reconcile. This is not a
// realistic end-user config (a real user would just supply a plaintext
// external_id learned from their cloud provider and let Fleet mint the
// secret), but it faithfully exercises cloud_connector_id round-tripping,
// which is this test's actual point.
func TestAccResourceManagedIntegration_CloudConnector(t *testing.T) {
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
	skipUnlessConfirmedCloud(t)
	acctest.PreCheck(t)

	secretRefID := mintExternalIDSecretRef(context.Background(), t, mustFleetClient(t))
	connectorID, cleanupConnector := createTestCloudConnector(t, secretRefID)
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
					"external_id_secret_id": config.StringVariable(secretRefID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.cloud_connector_id", connectorID),
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cloud_connector.target_csp", "aws"),
					testCheckCloudConnectorPersisted(testResourceName, connectorID),
				),
			},
		},
	})
}

// mustFleetClient is a small helper so TestAccResourceManagedIntegration_CloudConnector
// can obtain a *fleetclient.Client for mintExternalIDSecretRef without
// duplicating clients.NewAcceptanceTestingKibanaScopedClient's error handling
// inline.
func mustFleetClient(t *testing.T) *fleetclient.Client {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}
	return client.GetFleetClient()
}

// TestAccResourceManagedIntegration_NameUpdateInPlace verifies in-place `name`
// updates plan as Update and persist in Terraform state and on the Fleet API.
func TestAccResourceManagedIntegration_NameUpdateInPlace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
	skipUnlessConfirmedCloud(t)

	suffix := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	firstName := "agentless-" + suffix
	renamedName := "agentless-" + suffix + "-renamed"

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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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
				Check: testCheckJSONSubset("inputs.cspm-cloudbeat/cis_aws.streams.cloud_security_posture.findings.vars", map[string]any{
					"aws.account_type": "organization-account",
				}),
			},
		},
	})
}

// TestAccResourceManagedIntegration_PackageVersionUpdate verifies package.version
// updates in-place via managed_integrations PUT against a live stack.
func TestAccResourceManagedIntegration_PackageVersionUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_VersionSkipGating verifies apply fails on
// Kibana stacks older than managedintegration.MinVersion (9.5.0). Runs only
// when the acceptance Elasticsearch version is below that floor.
func TestAccResourceManagedIntegration_VersionSkipGating(t *testing.T) {
	constraints, err := version.NewConstraint("< " + managedintegration.MinVersion.String())
	if err != nil {
		t.Fatal(err)
	}
	versionutils.SkipIfUnsupportedConstraints(t, constraints, versionutils.FlavorStateful)

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
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
	skipUnlessConfirmedCloud(t)
	acctest.PreCheck(t)

	secretRefID := mintExternalIDSecretRef(context.Background(), t, mustFleetClient(t))
	connectorID, cleanupConnector := createTestCloudConnector(t, secretRefID)
	t.Cleanup(cleanupConnector)

	policyName := sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)
	baseVars := config.Variables{
		"policy_name":           config.StringVariable(policyName),
		"package_version":       config.StringVariable(cspmPackageVersion),
		"cloud_connector_id":    config.StringVariable(connectorID),
		"external_id_secret_id": config.StringVariable(secretRefID),
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
					testCheckCloudConnectorPersisted(testResourceName, connectorID),
				),
			},
		},
	})
}

// TestAccResourceManagedIntegration_ConditionRoundTrip verifies input/stream
// condition expressions round-trip through create, read, and in-place update.
func TestAccResourceManagedIntegration_ConditionRoundTrip(t *testing.T) {
	versionutils.SkipIfUnsupported(t, managedintegration.MinVersion, versionutils.FlavorAny)
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

// createTestCloudConnector creates a real AWS cloud connector via a raw
// POST /api/fleet/cloud_connectors call (there is no
// elasticstack_fleet_cloud_connector resource yet -- see this file's
// TestAccResourceManagedIntegration_CloudConnector doc comment) and returns its
// ID plus a cleanup function that deletes it. secretRefID must be a real,
// already-minted Fleet secret ID (see mintExternalIDSecretRef) -- the Kibana
// Fleet plugin requires `external_id` to be a pre-existing secret reference
// ({isSecretRef: true, id: ...}) rather than a bare string (empirically
// confirmed: a bare string is rejected with "External ID secret reference is
// not valid", even though `role_arn` accepts a bare string arm of the same
// union).
func createTestCloudConnector(t *testing.T, secretRefID string) (string, func()) {
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
			if err := v.FromPostFleetCloudConnectorsJSONBodyVars3Value1(kbapi.PostFleetCloudConnectorsJSONBodyVars3Value1{
				Id:          secretRefID,
				IsSecretRef: true,
			}); err != nil {
				t.Fatalf("failed to build cloud connector external_id secret ref: %v", err)
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
	name := fmt.Sprintf("tf-acc-agentless-policy-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
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

// mintExternalIDSecretRef creates and immediately deletes a throwaway
// agentless CSPM policy purely to get Fleet to convert a plain-string
// password-type var (aws.credentials.external_id) into a stored secret, then
// returns that secret's ref ID -- see createTestCloudConnector's doc
// comment for why this indirection is necessary (the cloud_connectors API
// has no direct "create a secret" endpoint of its own).
func mintExternalIDSecretRef(ctx context.Context, t *testing.T, fc *fleetclient.Client) string {
	t.Helper()

	probeName := fmt.Sprintf("tf-acc-secret-probe-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	createBody := kbapi.PostFleetAgentlessPoliciesJSONRequestBody{
		Name: probeName,
		Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{
			Name:    "cloud_security_posture",
			Version: cspmPackageVersion,
		},
	}
	policyTemplate := "cspm"
	createBody.PolicyTemplate = &policyTemplate
	if err := json.Unmarshal([]byte(`{"posture":"cspm","deployment":"aws"}`), &createBody.Vars); err != nil {
		t.Fatalf("failed to build secret-probe vars: %v", err)
	}
	if err := json.Unmarshal([]byte(`{
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"streams": {
				"cloud_security_posture.findings": {
					"enabled": true,
					"vars": {
						"aws.credentials.external_id": "tf-acc-secret-probe-value",
						"aws.credentials.type": "cloud_connectors",
						"aws.account_type": "single-account",
						"role_arn": "arn:aws:iam::123456789012:role/tf-acc-test-role"
					}
				}
			}
		}
	}`), &createBody.Inputs); err != nil {
		t.Fatalf("failed to build secret-probe inputs: %v", err)
	}

	createResp, err := fc.API.PostFleetAgentlessPoliciesWithResponse(ctx, createBody)
	if err != nil {
		t.Fatalf("failed to create secret-probe agentless policy: %v", err)
	}
	if createResp.StatusCode() != 200 || createResp.JSON200 == nil {
		t.Fatalf("failed to create secret-probe agentless policy: status %d: %s", createResp.StatusCode(), string(createResp.Body))
	}
	probeID := createResp.JSON200.Item.Id

	t.Cleanup(func() {
		force := true
		_, err := fc.API.DeleteFleetAgentlessPoliciesPolicyidWithResponse(ctx, probeID, &kbapi.DeleteFleetAgentlessPoliciesPolicyidParams{Force: &force})
		if err != nil {
			t.Logf("failed to delete secret-probe agentless policy %s: %v", probeID, err)
		}
	})

	getResp, err := fc.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, probeID, nil)
	if err != nil {
		t.Fatalf("failed to read back secret-probe agentless policy: %v", err)
	}
	if getResp.StatusCode() != 200 || getResp.JSON200 == nil {
		t.Fatalf("failed to read back secret-probe agentless policy: status %d: %s", getResp.StatusCode(), string(getResp.Body))
	}

	// Parsed directly from the raw response body (rather than through
	// kbapi's PackagePolicyMappedOrTypedInputs union accessors) to sidestep
	// ambiguity in which union arm GetFleetPackagePoliciesPackagepolicyidWithResponse's
	// generated accessors pick apart; this shape was confirmed directly
	// against a live Kibana 9.4.3 response body during development of this
	// fixture.
	// Value is decoded as json.RawMessage rather than a fixed struct because
	// its shape varies by var: most are a bare string (e.g. role_arn's
	// `{"value":"arn:...","type":"text"}`), while a secret-backed var's
	// value is an object (`{"value":{"id":"...","isSecretRef":true},...}`).
	var parsed struct {
		Item struct {
			Inputs []struct {
				Type    string `json:"type"`
				Streams []struct {
					Vars map[string]struct {
						Value json.RawMessage `json:"value"`
					} `json:"vars"`
				} `json:"streams"`
			} `json:"inputs"`
		} `json:"item"`
	}
	if err := json.Unmarshal(getResp.Body, &parsed); err != nil {
		t.Fatalf("failed to decode secret-probe policy response: %v", err)
	}
	for _, in := range parsed.Item.Inputs {
		if in.Type != "cloudbeat/cis_aws" {
			continue
		}
		for _, s := range in.Streams {
			v, ok := s.Vars["aws.credentials.external_id"]
			if !ok {
				continue
			}
			var secretRef struct {
				ID          string `json:"id"`
				IsSecretRef bool   `json:"isSecretRef"`
			}
			if err := json.Unmarshal(v.Value, &secretRef); err != nil {
				continue
			}
			if secretRef.IsSecretRef {
				return secretRef.ID
			}
		}
	}
	t.Fatalf("could not find a secret ref for aws.credentials.external_id in secret-probe policy %s", probeID)
	return ""
}

// testCheckManagedIntegrationNamePersisted reads GET /api/fleet/managed_integrations/{id}
// to confirm the integration name was persisted server-side after an in-place rename.
func testCheckManagedIntegrationNamePersisted(resourceName, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		policyID := rs.Primary.Attributes["policy_id"]
		spaceID := "default"
		if id, diags := clients.CompositeIDFromStr(rs.Primary.ID); !diags.HasError() && id != nil {
			spaceID = id.ClusterID
		}

		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		fc := client.GetFleetClient()

		item, diags := fleetclient.ReadManagedIntegration(context.Background(), fc, spaceID, policyID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if item == nil {
			return fmt.Errorf("managed integration %s not found when checking name", policyID)
		}
		if item.Name != expectedName {
			return fmt.Errorf("managed integration %s: expected name %q, got %q", policyID, expectedName, item.Name)
		}
		return nil
	}
}

// testCheckManagedIntegrationPackageVersionPersisted reads GET /api/fleet/managed_integrations/{id}
// to confirm package.version was persisted server-side after an in-place bump.
func testCheckManagedIntegrationPackageVersionPersisted(resourceName, expectedVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		policyID := rs.Primary.Attributes["policy_id"]
		spaceID := "default"
		if id, diags := clients.CompositeIDFromStr(rs.Primary.ID); !diags.HasError() && id != nil {
			spaceID = id.ClusterID
		}

		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		fc := client.GetFleetClient()

		item, diags := fleetclient.ReadManagedIntegration(context.Background(), fc, spaceID, policyID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if item == nil {
			return fmt.Errorf("managed integration %s not found when checking package.version", policyID)
		}
		if item.Package.Version != expectedVersion {
			return fmt.Errorf("managed integration %s: expected package.version %q, got %q", policyID, expectedVersion, item.Package.Version)
		}
		return nil
	}
}

// testCheckCloudConnectorPersisted reads GET /api/fleet/managed_integrations/{id}
// to confirm cloud_connector_id was persisted server-side (association fields are
// re-sent on PUT from prior state; name/target_csp are write-only in state).
func testCheckCloudConnectorPersisted(resourceName, expectedConnectorID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		policyID := rs.Primary.Attributes["policy_id"]
		spaceID := "default"
		if id, diags := clients.CompositeIDFromStr(rs.Primary.ID); !diags.HasError() && id != nil {
			spaceID = id.ClusterID
		}

		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		fc := client.GetFleetClient()

		item, diags := fleetclient.ReadManagedIntegration(context.Background(), fc, spaceID, policyID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if item == nil {
			return fmt.Errorf("managed integration %s not found when checking cloud_connector_id", policyID)
		}
		if item.CloudConnector == nil || item.CloudConnector.CloudConnectorId != expectedConnectorID {
			got := ""
			if item.CloudConnector != nil {
				got = item.CloudConnector.CloudConnectorId
			}
			return fmt.Errorf("managed integration %s: expected cloud_connector_id %q, got %q", policyID, expectedConnectorID, got)
		}
		return nil
	}
}
