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

package connector_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/connectorfieldtype"
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

const (
	contentConnectorResourceAddr   = "elasticstack_elasticsearch_connector.test"
	contentConnectorDataSourceAddr = "data.elasticstack_elasticsearch_connector.lookup"
)

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

func createConnectorViaAPI(t *testing.T, connectorID string) {
	t.Helper()
	accRequireConnectorSupported(t)
	ctx := context.Background()
	client := accConnectorClient(t)
	name := "TF acc ds api"
	description := "created via API for data source acceptance test"
	indexName := "content-connector-" + connectorID
	_, diags := esclient.CreateConnector(ctx, client, connectorID, esclient.CreateConnectorBody{
		Name:        &name,
		Description: &description,
		IndexName:   &indexName,
		ServiceType: "postgresql",
	})
	if diags.HasError() {
		t.Fatalf("create connector %q via API: %s", connectorID, diags[0].Summary())
	}
}

func testAccCheckContentConnectorDestroyByID(t *testing.T, connectorID string) func(*terraform.State) error {
	t.Helper()
	return func(*terraform.State) error {
		accRequireConnectorSupported(t)
		deleteConnectorAPI(t, connectorID)
		return nil
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

func registerConnectorSensitiveConfigurationSchema(t *testing.T, connectorID string) {
	t.Helper()
	registerConnectorConfigurationSchema(t, connectorID, map[string]map[string]any{
		"api_secret": connectorConfigurationSchemaField("API secret", connectorfieldtype.Str.Name, true),
		"endpoint":   connectorConfigurationSchemaField("Endpoint", connectorfieldtype.Str.Name, false),
	})
}

func putConnectorFilteringMarker(t *testing.T, connectorID, marker string) {
	t.Helper()
	accRequireConnectorSupported(t)
	ctx := context.Background()
	client := accConnectorClient(t)
	body, err := json.Marshal(map[string]any{
		"rules": []map[string]any{{
			"id":     "rule-acc",
			"order":  0,
			"field":  "title",
			"rule":   "contains",
			"policy": "include",
			"value":  marker,
		}},
	})
	if err != nil {
		t.Fatalf("marshal filtering rules: %v", err)
	}
	_, err = client.GetESClient().Connector.UpdateFiltering(connectorID).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		t.Fatalf("update filtering for %q: %v", connectorID, err)
	}
}

func testAccCheckDataSourceConfigurationContains(keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[contentConnectorDataSourceAddr]
		if !ok {
			return fmt.Errorf("data source %q not in state", contentConnectorDataSourceAddr)
		}
		configJSON, ok := rs.Primary.Attributes["configuration"]
		if !ok || configJSON == "" {
			return fmt.Errorf("data source %q has no configuration attribute", contentConnectorDataSourceAddr)
		}
		for _, key := range keys {
			if !strings.Contains(configJSON, fmt.Sprintf("%q", key)) {
				return fmt.Errorf("configuration JSON missing key %q: %s", key, configJSON)
			}
		}
		return nil
	}
}

func testAccCheckDataSourceJSONObjectEmpty(attr string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(contentConnectorDataSourceAddr, attr, func(value string) error {
		var m map[string]any
		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return fmt.Errorf("%s is not valid JSON object: %w", attr, err)
		}
		if len(m) != 0 {
			return fmt.Errorf("expected empty %s object, got %v", attr, m)
		}
		return nil
	})
}

func testAccCheckDataSourceJSONArrayNonEmpty(attr string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(contentConnectorDataSourceAddr, attr, func(value string) error {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(value), &arr); err != nil {
			return fmt.Errorf("%s is not valid JSON array: %w", attr, err)
		}
		if len(arr) == 0 {
			return fmt.Errorf("expected non-empty %s array", attr)
		}
		if !strings.Contains(value, `"active"`) {
			return fmt.Errorf("expected %s to contain filtering rule structure with \"active\"", attr)
		}
		return nil
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
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.name", "ent-search-generic-ingestion"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.extract_binary_content", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.reduce_whitespace", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.run_ml_inference", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.interval", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.incremental_sync.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.sync_rules.basic.enabled", "true"),
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
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "index_name", "content-connector-upd-"+connectorID),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.extract_binary_content", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "pipeline.run_ml_inference", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.full.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "scheduling.incremental.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorResourceAddr, "features.document_level_security.enabled", "true"),
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

// TestAccDataSourceContentConnector_basic verifies envelope, aspects, and runtime telemetry (REQ-010).
func TestAccDataSourceContentConnector_basic(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-basic")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroy(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(contentConnectorDataSourceAddr, "id", connectorCompositeIDRegexp(connectorID)),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "connector_id", connectorID),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "service_type", "postgresql"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "name", "TF acc ds basic"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "description", "data source basic acceptance test"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "index_name", "content-connector-"+connectorID),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "language", "en"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "is_native", "false"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "api_key_id"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "api_key_secret_id"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "pipeline.name", "ent-search-generic-ingestion"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "pipeline.extract_binary_content", "true"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "pipeline.reduce_whitespace", "true"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "pipeline.run_ml_inference", "false"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.full.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.full.interval", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.incremental.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.incremental.interval", "0 30 * * * ?"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.access_control.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "scheduling.access_control.interval", "0 0 0 * * ?"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "features.document_level_security.enabled", "false"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "features.incremental_sync.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "features.sync_rules.basic.enabled", "true"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "features.native_connector_api_keys.enabled"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "features.sync_rules.advanced.enabled"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "status", "created"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_seen"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_synced"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_sync_status"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_indexed_document_count"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_deleted_document_count"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_sync_scheduled_at"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_sync_error"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_access_control_sync_status"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_access_control_sync_error"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_access_control_sync_scheduled_at"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "last_incremental_sync_scheduled_at"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "error"),
					testAccCheckDataSourceJSONObjectEmpty("configuration"),
					testAccCheckDataSourceJSONArrayNonEmpty("filtering"),
					testAccCheckDataSourceJSONObjectEmpty("custom_scheduling"),
					resource.TestCheckNoResourceAttr(contentConnectorDataSourceAddr, "sync_cursor"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "sync_now", "false"),
				),
			},
		},
	})
}

// TestAccDataSourceContentConnector_filteringAndCustomScheduling verifies filtering is exposed after API update.
// custom_scheduling has no PUT endpoint (connector service only); empty {} exposure is asserted here.
func TestAccDataSourceContentConnector_filteringAndCustomScheduling(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-filt")
	marker := "tf-acc-filter-marker"
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroyByID(t, connectorID),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createConnectorViaAPI(t, connectorID)
					putConnectorFilteringMarker(t, connectorID, marker)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(contentConnectorDataSourceAddr, "filtering", regexp.MustCompile(regexp.QuoteMeta(marker))),
					testAccCheckDataSourceJSONObjectEmpty("custom_scheduling"),
				),
			},
		},
	})
}

// TestAccDataSourceContentConnector_notFound verifies a 404 surfaces as a diagnostic error.
func TestAccDataSourceContentConnector_notFound(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-nonexistent")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`(?s)Connector not found.*` + regexp.QuoteMeta(connectorID)),
			},
		},
	})
}

// TestAccDataSourceContentConnector_configurationWithSensitiveFields verifies the data source exposes the full configuration schema including sensitive fields.
func TestAccDataSourceContentConnector_configurationWithSensitiveFields(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-cfg")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroyByID(t, connectorID),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createConnectorViaAPI(t, connectorID)
					registerConnectorSensitiveConfigurationSchema(t, connectorID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceConfigurationContains("api_secret", "endpoint"),
					resource.TestMatchResourceAttr(contentConnectorDataSourceAddr, "configuration", regexp.MustCompile(`"api_secret"[^}]*"sensitive":\s*true`)),
				),
			},
		},
	})
}

// TestAccResourceContentConnector_versionGate verifies apply fails on Elasticsearch versions below 8.12.
func TestAccResourceContentConnector_versionGate(t *testing.T) {
	isUnsupported, err := skipConnectorUnsupported()()
	if err != nil {
		t.Fatalf("could not determine server version: %v", err)
	}
	if !isUnsupported {
		t.Skip("requires Elasticsearch < 8.12.0")
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
				ExpectError:              regexp.MustCompile(`8\.12\.0`),
				Destroy:                  false,
			},
		},
	})
}
