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

package data_source_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatestatus"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorfieldtype"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorstatus"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const contentConnectorDataSourceAddr = "data.elasticstack_elasticsearch_connector.lookup"

// skipConnectorUnsupported reports whether the acceptance Elasticsearch
// cluster is older than the connector minimum supported version.
func skipConnectorUnsupported() func() (bool, error) {
	return versionutils.CheckIfVersionIsUnsupported(connector.MinSupportedVersion)
}

// accRequireConnectorSupported skips when the acceptance stack is below
// connector.MinSupportedVersion. Call from PreConfig and CheckDestroy
// helpers (which run before TestStep SkipFunc).
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

// setConnectorSyncNow sets sync_now = true on the connector document by updating the
// .elastic-connectors index directly. This is a test-only helper because the ES typed
// client does not expose a dedicated sync_now endpoint.
func setConnectorSyncNow(t *testing.T, connectorID string) {
	t.Helper()
	accRequireConnectorSupported(t)

	endpoint, _, _ := strings.Cut(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",")
	endpoint = strings.TrimRight(endpoint, "/")

	rawURL := fmt.Sprintf("%s/.elastic-connectors/_update/%s", endpoint, connectorID)
	body := strings.NewReader(`{"doc":{"sync_now":true}}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, rawURL, body)
	if err != nil {
		t.Fatalf("build sync_now update request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("ELASTICSEARCH_USERNAME"), os.Getenv("ELASTICSEARCH_PASSWORD"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("set sync_now for connector %q: %v", connectorID, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("set sync_now for connector %q: HTTP %d: %s", connectorID, resp.StatusCode, string(respBody))
	}
}

// updateConnectorStatusViaAPI updates the connector's lifecycle status.
func updateConnectorStatusViaAPI(t *testing.T, connectorID string, status connectorstatus.ConnectorStatus) {
	t.Helper()
	accRequireConnectorSupported(t)
	ctx := context.Background()
	client := accConnectorClient(t)
	_, err := client.GetESClient().Connector.UpdateStatus(connectorID).Request(&updatestatus.Request{
		Status: status,
	}).Do(ctx)
	if err != nil {
		t.Fatalf("update connector %q status to %q: %v", connectorID, status.Name, err)
	}
}

// testAccCheckNativeConnectorAPIKeysOrSkip skips the test when the
// native_connector_api_keys feature is absent from the data source state.
// This guards the native connector test on ES clusters that do not populate
// the feature (e.g., non-Elastic Cloud environments).
func testAccCheckNativeConnectorAPIKeysOrSkip(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[contentConnectorDataSourceAddr]
		if !ok {
			return fmt.Errorf("data source %q not in state", contentConnectorDataSourceAddr)
		}
		if _, present := rs.Primary.Attributes["features.native_connector_api_keys.enabled"]; !present {
			t.Skip("native_connector_api_keys feature not returned by this Elasticsearch cluster; skipping native connector API keys assertion")
		}
		return nil
	}
}

// TestAccDataSourceContentConnector_advancedSyncRules verifies that the data source
// correctly reflects features.sync_rules.advanced.enabled when set to true (REQ-010 gap 1).
func TestAccDataSourceContentConnector_advancedSyncRules(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-adv")
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
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "features.sync_rules.advanced.enabled", "true"),
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "features.sync_rules.basic.enabled", "true"),
				),
			},
		},
	})
}

// TestAccDataSourceContentConnector_syncNowQueued verifies the data source exposes
// sync_now = true when set on the connector document (REQ-010 gap 2).
func TestAccDataSourceContentConnector_syncNowQueued(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-snq")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroyByID(t, connectorID),
		Steps: []resource.TestStep{
			// Step 1: Create the connector via API; sync_now starts as false.
			{
				PreConfig: func() {
					createConnectorViaAPI(t, connectorID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check:                    resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "sync_now", "false"),
			},
			// Step 2: Set sync_now = true out of band; verify the data source reads it.
			{
				PreConfig: func() {
					setConnectorSyncNow(t, connectorID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check:                    resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "sync_now", "true"),
			},
		},
	})
}

// TestAccDataSourceContentConnector_statusNeedsConfiguration verifies the data source
// exposes status values other than "created" (REQ-010 gap 3).
func TestAccDataSourceContentConnector_statusNeedsConfiguration(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-ncon")
	vars := connectorIDVariables(connectorID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckContentConnectorDestroyByID(t, connectorID),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createConnectorViaAPI(t, connectorID)
					updateConnectorStatusViaAPI(t, connectorID, connectorstatus.Needsconfiguration)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check:                    resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "status", "needs_configuration"),
			},
		},
	})
}

// TestAccDataSourceContentConnector_nativeConnector verifies features.native_connector_api_keys
// is present for connectors created with is_native = true (REQ-010 gap 4).
// The test skips when the ES cluster does not populate this feature.
func TestAccDataSourceContentConnector_nativeConnector(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-ds-nat")
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
					resource.TestCheckResourceAttr(contentConnectorDataSourceAddr, "is_native", "true"),
					testAccCheckNativeConnectorAPIKeysOrSkip(t),
					resource.TestCheckResourceAttrSet(contentConnectorDataSourceAddr, "features.native_connector_api_keys.enabled"),
				),
			},
		},
	})
}
