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

package resource_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorfieldtype"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const contentConnectorResourceAddr = "elasticstack_elasticsearch_connector.test"

func skipConnectorUnsupported() func() (bool, error) {
	return versionutils.CheckIfVersionIsUnsupported(connector.MinSupportedVersion)
}

// accRequireConnectorSupported skips when the acceptance stack is below MinSupportedVersion.
// Call from PreConfig and CheckDestroy helpers: PreConfig runs before TestStep SkipFunc.
func accRequireConnectorSupported(t *testing.T) {
	t.Helper()
	versionutils.SkipIfUnsupported(t, connector.MinSupportedVersion, versionutils.FlavorAny)
}

func connectorIDVariables(connectorID string) config.Variables {
	return config.Variables{
		"connector_id": config.StringVariable(connectorID),
	}
}

func connectorCompositeIDRegexp(connectorID string) *regexp.Regexp {
	return regexp.MustCompile(`^[^/]+/` + regexp.QuoteMeta(connectorID) + `$`)
}

func accConnectorClient(t *testing.T) *clients.ElasticsearchScopedClient {
	t.Helper()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	return client
}

func deleteConnectorAPI(t *testing.T, connectorID string) {
	t.Helper()
	accRequireConnectorSupported(t)
	ctx := context.Background()
	client := accConnectorClient(t)
	if diags := esclient.DeleteConnector(ctx, client, connectorID); diags.HasError() {
		t.Fatalf("delete connector %q: %s", connectorID, diags[0].Summary())
	}
}

func connectorConfigurationSchemaField(label, fieldType string, sensitive bool) map[string]any {
	return map[string]any{
		"label":     label,
		"type":      fieldType,
		"display":   "text",
		"required":  false,
		"sensitive": sensitive,
	}
}

func registerConnectorConfigurationSchema(t *testing.T, connectorID string, schema map[string]map[string]any) {
	t.Helper()
	accRequireConnectorSupported(t)
	ctx := context.Background()
	client := accConnectorClient(t)
	body, err := json.Marshal(map[string]any{"configuration": schema})
	if err != nil {
		t.Fatalf("marshal configuration schema: %v", err)
	}
	_, err = client.GetESClient().Connector.UpdateConfiguration(connectorID).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		t.Fatalf("register configuration schema for %q: %v", connectorID, err)
	}
}

func registerConnectorConfigurationBranchesSchema(t *testing.T, connectorID string) {
	t.Helper()
	registerConnectorConfigurationSchema(t, connectorID, map[string]map[string]any{
		"s_branch": connectorConfigurationSchemaField("String branch", connectorfieldtype.Str.Name, false),
		"n_branch": connectorConfigurationSchemaField("Number branch", connectorfieldtype.Int.Name, false),
		"b_branch": connectorConfigurationSchemaField("Bool branch", connectorfieldtype.Bool.Name, false),
		"j_branch": connectorConfigurationSchemaField("JSON branch", connectorfieldtype.List.Name, false),
	})
}

func registerConnectorPasswordSchema(t *testing.T, connectorID string) {
	t.Helper()
	registerConnectorPasswordSchemaSensitive(t, connectorID, true)
}

func registerConnectorPasswordSchemaSensitive(t *testing.T, connectorID string, sensitive bool) {
	t.Helper()
	registerConnectorConfigurationSchema(t, connectorID, map[string]map[string]any{
		"password": connectorConfigurationSchemaField("Password", connectorfieldtype.Str.Name, sensitive),
	})
}

func testAccCheckConnectorConfigurationEmpty(connectorID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		resp, diags := esclient.GetConnector(ctx, client, connectorID)
		if diags.HasError() {
			return fmt.Errorf("get connector %q: %s", connectorID, diags[0].Summary())
		}
		if resp == nil {
			return fmt.Errorf("connector %q not found", connectorID)
		}
		if len(resp.Configuration) > 0 {
			return fmt.Errorf("expected empty configuration schema on %q, got %d keys", connectorID, len(resp.Configuration))
		}
		return nil
	}
}

func testAccCheckConnectorConfigurationStringValue(connectorID, key, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		resp, diags := esclient.GetConnector(ctx, client, connectorID)
		if diags.HasError() {
			return fmt.Errorf("get connector %q: %s", connectorID, diags[0].Summary())
		}
		if resp == nil {
			return fmt.Errorf("connector %q not found", connectorID)
		}
		props, ok := resp.Configuration[key]
		if !ok {
			return fmt.Errorf("configuration key %q not found on connector %q", key, connectorID)
		}
		if len(props.Value) == 0 || string(props.Value) == "null" {
			return fmt.Errorf("configuration key %q has no value on connector %q", key, connectorID)
		}
		var got string
		if err := json.Unmarshal(props.Value, &got); err != nil {
			return fmt.Errorf("decode configuration value for %q: %w", key, err)
		}
		if got != want {
			return fmt.Errorf("configuration %q value: got %q, want %q", key, got, want)
		}
		return nil
	}
}

func registerConnectorAPIKeySchema(t *testing.T, connectorID string) {
	t.Helper()
	registerConnectorConfigurationSchema(t, connectorID, map[string]map[string]any{
		"api_key": connectorConfigurationSchemaField("API key", connectorfieldtype.Str.Name, true),
	})
}

func testAccCheckContentConnectorAbsentFromState(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[addr]; ok {
			return fmt.Errorf("expected %q to be absent from state after refresh (connector deleted out-of-band)", addr)
		}
		return nil
	}
}

func testAccCheckContentConnectorDestroy(connectorID string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_connector" {
				continue
			}
			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
			if compID.ResourceID != connectorID {
				continue
			}
			resp, diags := esclient.GetConnector(ctx, client, connectorID)
			if diags.HasError() {
				return fmt.Errorf("checking connector deletion: %s", diags[0].Summary())
			}
			if resp != nil {
				return fmt.Errorf("connector %q still exists", connectorID)
			}
		}
		return nil
	}
}

func contentConnectorImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return rs.Primary.ID, nil
	}
}

// TestAccResourceContentConnector_minimal covers create, composite import, and destroy.
func TestAccResourceContentConnector_minimal(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-min")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(contentConnectorResourceAddr, "id", connectorCompositeIDRegexp(connectorID)),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "connector_id", connectorID),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "service_type", "postgresql"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "name", "TF acc minimal"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "description", "acceptance test connector"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "is_native", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             contentConnectorResourceAddr,
				ImportState:              true,
				ImportStateIdFunc:        contentConnectorImportID(contentConnectorResourceAddr),
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_fullEnvelope verifies pipeline, scheduling, and features land in state.
func TestAccResourceContentConnector_fullEnvelope(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-full")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "connector_id", connectorID),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "service_type", "postgresql"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "name", "TF acc full envelope"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "index_name", "content-connector-"+connectorID),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "language", "en"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.name", "ent-search-generic-ingestion"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.extract_binary_content", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.reduce_whitespace", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.run_ml_inference", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.interval", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.interval", "0 30 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.access_control.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.access_control.interval", "0 0 0 * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.incremental_sync.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.sync_rules.basic.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.native_connector_api_keys.enabled", "false"),
				),
			},
		},
	})
}

// TestAccResourceContentConnector_update verifies envelope aspect updates and a clean replan.
func TestAccResourceContentConnector_update(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-upd")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "name", "TF acc full envelope"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.run_ml_inference", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "name", "TF acc updated"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "description", "updated description"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "language", "en"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "index_name", "content-connector-upd-"+connectorID),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.extract_binary_content", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.reduce_whitespace", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.run_ml_inference", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.interval", "0 15 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.interval", "0 45 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.document_level_security.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.incremental_sync.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.sync_rules.basic.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.sync_rules.advanced.enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_configurationValues_branches round-trips each configuration_values branch.
func TestAccResourceContentConnector_configurationValues_branches(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-cfg")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: func(_ *terraform.State) error {
					registerConnectorConfigurationBranchesSchema(t, connectorID)
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_values"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "configuration_values.s_branch.string", "x"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "configuration_values.n_branch.number", "42"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "configuration_values.b_branch.bool", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "configuration_values.j_branch.json", `{"a":1}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_values"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_configurationValues_secret verifies write-only secret_value behaviour.
func TestAccResourceContentConnector_configurationValues_secret(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-sec")
	baseVars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars,
				Check: func(_ *terraform.State) error {
					registerConnectorPasswordSchema(t, connectorID)
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables: config.Variables{
					"connector_id": config.StringVariable(connectorID),
					"secret_value": config.StringVariable("pw1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(contentConnectorResourceAddr, "configuration_values.password.secret_value"),
					testAccCheckConnectorConfigurationStringValue(connectorID, "password", "pw1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables: config.Variables{
					"connector_id": config.StringVariable(connectorID),
					"secret_value": config.StringVariable("pw1"),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				// REQ-011: hash mismatch marks id unknown → non-empty plan (warning not assertable in acc tests).
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables: config.Variables{
					"connector_id": config.StringVariable(connectorID),
					"secret_value": config.StringVariable("pw2"),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables: config.Variables{
					"connector_id": config.StringVariable(connectorID),
					"secret_value": config.StringVariable("pw2"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(contentConnectorResourceAddr, plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccCheckConnectorConfigurationStringValue(connectorID, "password", "pw2"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_secret"),
				ConfigVariables:          baseVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_secret"),
				ConfigVariables:          baseVars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				// Hash cleared after removal: re-applying the same secret should schedule a fresh baseline update.
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables: config.Variables{
					"connector_id": config.StringVariable(connectorID),
					"secret_value": config.StringVariable("pw1"),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccResourceContentConnector_configurationValues_branchValidator verifies plan-time branch validation.
func TestAccResourceContentConnector_configurationValues_branchValidator(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-val")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_branch"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`(?i)Exactly one of`),
				Destroy:                  false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi_branch"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`(?i)Exactly one of`),
				Destroy:                  false,
			},
		},
	})
}

// TestAccResourceContentConnector_configurationValues_sensitiveFieldWarning preserves sensitive string values across refresh.
// REQ-008 warning diagnostic is covered at unit level in TestPopulateConfigurationValuesFromAPI/sensitive_non-secret_branch_warns
// (terraform-plugin-testing cannot assert provider warning diagnostics during acc tests).
func TestAccResourceContentConnector_configurationValues_sensitiveFieldWarning(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-warn")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: func(_ *terraform.State) error {
					registerConnectorAPIKeySchema(t, connectorID)
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_api_key"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "configuration_values.api_key.string", "x"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_api_key"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_configurationValues_preflight rejects configuration_values before schema registration.
func TestAccResourceContentConnector_configurationValues_preflight(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-pre")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check:                    testAccCheckConnectorConfigurationEmpty(connectorID),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_values"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile("Connector configuration schema not yet registered"),
				Destroy:                  false,
				Check:                    testAccCheckConnectorConfigurationEmpty(connectorID),
			},
		},
	})
}

// TestAccResourceContentConnector_import verifies bare and composite import forms.
func TestAccResourceContentConnector_import(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-imp")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             contentConnectorResourceAddr,
				ImportState:              true,
				ImportStateId:            connectorID,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             contentConnectorResourceAddr,
				ImportState:              true,
				ImportStateIdFunc:        contentConnectorImportID(contentConnectorResourceAddr),
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_externalDelete verifies refresh drops a connector deleted out-of-band.
func TestAccResourceContentConnector_externalDelete(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-del")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "connector_id", connectorID),
				),
			},
			{
				PreConfig: func() {
					deleteConnectorAPI(t, connectorID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				Check:                    testAccCheckContentConnectorAbsentFromState(contentConnectorResourceAddr),
			},
		},
	})
}

// TestAccResourceContentConnector_deleteIdempotent verifies destroy succeeds when the connector was already deleted.
func TestAccResourceContentConnector_deleteIdempotent(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-404")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				PreConfig: func() {
					deleteConnectorAPI(t, connectorID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Destroy:         true,
			},
		},
	})
}

// TestAccResourceContentConnector_importSecretBaseline verifies REQ-011 post-import secret behaviour.
func TestAccResourceContentConnector_importSecretBaseline(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-impsec")
	secretVars := config.Variables{
		"connector_id": config.StringVariable(connectorID),
		"secret_value": config.StringVariable("pw1"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          connectorIDVariables(connectorID),
				Check: func(_ *terraform.State) error {
					registerConnectorPasswordSchema(t, connectorID)
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables:          secretVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables:          secretVars,
				ResourceName:             contentConnectorResourceAddr,
				ImportState:              true,
				ImportStateId:            connectorID,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"configuration_values", "id"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables:          secretVars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables:          secretVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secret"),
				ConfigVariables:          secretVars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceContentConnector_versionGate verifies apply fails on Elasticsearch versions below 8.16.
func TestAccResourceContentConnector_versionGate(t *testing.T) {
	isUnsupported, err := skipConnectorUnsupported()()
	if err != nil {
		t.Fatalf("could not determine server version: %v", err)
	}
	if !isUnsupported {
		t.Skip("requires Elasticsearch < 8.16.0")
	}

	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ver")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`8\.16\.0`),
				Destroy:                  false,
			},
		},
	})
}
