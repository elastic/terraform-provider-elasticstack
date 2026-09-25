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

package agentbuilderworkflow_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaAgentBuilderAPIVersion = agentbuilder.MinExtendedAPIVersion
)

func TestAccResourceAgentBuilderWorkflow(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	// workflow IDs are workflow-<UUIDv4>
	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceID, "id", regexp.MustCompile(`^default/workflow-`)),
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "A test workflow for acceptance testing"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration_yaml"),
					resource.TestMatchResourceAttr(resourceID, "configuration_yaml", regexp.MustCompile(`name: Test Workflow`)),
				),
			},
			{
				// Verify whitespace/indentation/blank-line changes are detected as updates.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_whitespace"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
				),
			},
			{
				// Verify map key reordering is detected as an update.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_reordered"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
				),
			},
			{
				// Import by composite id: <space_id>/<workflow_id>
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				ResourceName: resourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Updated Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "An updated test workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				// Verify description is null in state when YAML omits the description field.
				// populateFromAPI maps a missing/empty description to types.StringNull(), which
				// in Plugin Framework state means the attribute is not present (not "").
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_no_description"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckNoResourceAttr(resourceID, "description"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_space"
	spaceResourceID := "elasticstack_kibana_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
					"space_id":    config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(spaceResourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "name", "Space Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				// Import by composite id: <space_id>/<workflow_id>
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
					"space_id":    config.StringVariable(spaceID),
				},
				ResourceName: resourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowInvalidCreate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_invalid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				ExpectError: regexp.MustCompile(`(?i)invalid workflow`),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowInvalidUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_invalid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_valid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_invalid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				ExpectError: regexp.MustCompile(`(?i)invalid workflow`),
			},
		},
	})
}

// TestAccResourceAgentBuilderWorkflowInvalidDrift verifies that `valid` is
// surfaced as `false` on a workflow that already exists in state, without an
// error diagnostic. This can't happen through Terraform's own Create/Update
// path: populateWrittenCreate/populateWrittenUpdate (create.go, update.go)
// deliberately turn `valid: false` from those responses into an error so the
// write fails outright, meaning the resource is never persisted to state with
// valid = false via apply. But the underlying Kibana API does accept and
// store such YAML (that's the response those callbacks are reacting to), so
// a workflow can go invalid out-of-band -- e.g. a later API/UI edit -- and
// Terraform's plain Read (populateFromAPI) has no such guard, so a refresh
// must surface valid = false cleanly.
func TestAccResourceAgentBuilderWorkflowInvalidDrift(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_invalid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_valid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				// Mutate the workflow's YAML directly via the API (bypassing
				// Terraform) to a configuration Kibana accepts but flags
				// invalid -- the same YAML TestAccResourceAgentBuilderWorkflowInvalidUpdate
				// uses to trigger the write-path error -- then refresh and
				// confirm the resource picks up valid = false with no error.
				// RefreshState reuses the prior step's config; it cannot be
				// combined with ConfigDirectory/ConfigVariables.
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					setWorkflowYamlDirect(t, "default", workflowID, "not_working: hello_world")
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "valid", "false"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowAutoGeneratedID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_auto"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceID, "workflow_id"),
					resource.TestCheckResourceAttr(resourceID, "name", "Auto ID Workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderWorkflow(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	const (
		resourceID   = "elasticstack_kibana_agentbuilder_workflow.test"
		dataSourceID = "data.elasticstack_kibana_agentbuilder_workflow.test"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", resourceID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "workflow_id", resourceID, "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "configuration_yaml", resourceID, "configuration_yaml"),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					// kibana_connection is not set on the data source block: confirm the default/absent state.
					resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.#", "0"),
					// Assert configuration_yaml directly on the data source (independent of the paired resource)
					// to exercise the data-source-side YAML normalization path on its own.
					resource.TestMatchResourceAttr(dataSourceID, "configuration_yaml", regexp.MustCompile(`name: Test Workflow`)),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaAgentBuilderWorkflowNotFound verifies that looking up
// a nonexistent workflow ID surfaces the "Workflow not found" diagnostic from
// readWorkflowDataSource, mirroring TestAccResourceAgentBuilderWorkflowInvalidCreate's
// resource-side error-path coverage.
func TestAccDataSourceKibanaAgentBuilderWorkflowNotFound(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	missingID := "workflow-does-not-exist-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_found"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(missingID),
				},
				ExpectError: regexp.MustCompile(`(?i)workflow not found`),
			},
		},
	})
}

// TestAccDataSourceKibanaAgentBuilderWorkflowSpacePrecedence verifies that an
// explicit space_id set directly on the data source block wins over the space
// segment embedded in a composite id, per clients.ResolveCompositeSpaceAndID.
// The id passed in embeds a space that doesn't exist; the explicit space_id
// (matching the space the workflow actually lives in) must take precedence for
// the lookup to succeed.
func TestAccDataSourceKibanaAgentBuilderWorkflowSpacePrecedence(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	const (
		resourceID   = "elasticstack_kibana_agentbuilder_workflow.test"
		dataSourceID = "data.elasticstack_kibana_agentbuilder_workflow.test"
	)

	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "workflow_id", resourceID, "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "configuration_yaml", resourceID, "configuration_yaml"),
					// The explicit space_id must win, not the "does-not-exist" space embedded in id.
					resource.TestCheckResourceAttr(dataSourceID, "space_id", spaceID),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaAgentBuilderWorkflowBareID verifies that a bare
// (non-composite) workflow_id combined with an explicit space_id resolves to
// the same workflow as the composite-id lookup form, per
// clients.ResolveCompositeSpaceAndID.
func TestAccDataSourceKibanaAgentBuilderWorkflowBareID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	const (
		resourceID   = "elasticstack_kibana_agentbuilder_workflow.test"
		dataSourceID = "data.elasticstack_kibana_agentbuilder_workflow.test"
	)

	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", resourceID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "workflow_id", resourceID, "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "configuration_yaml", resourceID, "configuration_yaml"),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", spaceID),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderWorkflowSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	const (
		resourceID   = "elasticstack_kibana_agentbuilder_workflow.test"
		dataSourceID = "data.elasticstack_kibana_agentbuilder_workflow.test"
	)

	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", resourceID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "workflow_id", resourceID, "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "configuration_yaml", resourceID, "configuration_yaml"),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", spaceID),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderWorkflowKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	const (
		resourceID   = "elasticstack_kibana_agentbuilder_workflow.test"
		dataSourceID = "data.elasticstack_kibana_agentbuilder_workflow.test"
	)

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair(dataSourceID, "id", resourceID, "id"),
		resource.TestCheckResourceAttrPair(dataSourceID, "workflow_id", resourceID, "workflow_id"),
		resource.TestCheckResourceAttrPair(dataSourceID, "configuration_yaml", resourceID, "configuration_yaml"),
		resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(dataSourceID)...)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          acctest.KibanaConnectionVariables(config.Variables{}),
				Check:                    resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

// TestAccResourceAgentBuilderWorkflowKibanaConnection exercises the kibana_connection block
// on the workflow resource, covering all sub-attributes and verifying import round-trip.
func TestAccResourceAgentBuilderWorkflowKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	workflowID := "workflow-" + uuid.New().String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttrSet(resourceID, "kibana_connection.0.endpoints.0"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.insecure", "false"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				// Import ignores kibana_connection (same pattern as tool tests).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
			{
				// insecure = true variant, matching the tool resource's
				// kibana_connection coverage (agentbuildertool/acc_test.go).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.insecure", "true"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
		},
	})
}

// setWorkflowYamlDirect updates a workflow's YAML directly through the Kibana
// API, bypassing Terraform entirely, so tests can put a workflow into a state
// (e.g. valid = false) that the resource's own Create/Update callbacks would
// otherwise turn into an apply-time error.
func setWorkflowYamlDirect(t *testing.T, spaceID, workflowID, yaml string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}

	_, diags := kibanaoapi.UpdateWorkflow(context.Background(), client.GetKibanaOapiClient(), spaceID, workflowID, kbapi.PutWorkflowsWorkflowIdJSONRequestBody{
		Yaml: &yaml,
	})
	if diags.HasError() {
		t.Fatalf("failed to update workflow directly: %v", diags)
	}
}
