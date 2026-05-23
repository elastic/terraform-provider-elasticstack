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

package dashboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccResourceDashboardUnknownPanel_lensDashboardApp verifies read-time unknown-panel
// fallback when Kibana still returns a lens-dashboard-app panel (no registered handler).
func TestAccResourceDashboardUnknownPanel_lensDashboardApp(t *testing.T) {
	dashboardTitle := "Test Dashboard lens-dashboard-app fallback " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	var dashboardID string

	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["elasticstack_kibana_dashboard.test"]
						if !ok {
							return fmt.Errorf("resource not found in state")
						}
						parsedID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
						if diags.HasError() {
							return fmt.Errorf("could not parse dashboard composite ID %q", rs.Primary.ID)
						}
						dashboardID = parsedID.ResourceID
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					if dashboardID == "" {
						t.Fatal("dashboardID not set from step 1")
					}
					if err := replaceDashboardPanelWithLensDashboardApp(t, dashboardID); err != nil {
						t.Fatalf("failed to replace dashboard panels: %v", err)
					}
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "48"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.config_json"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config"),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "panels.0.config_json", `"type":"metric"`),
				),
			},
		},
	})
}

func replaceDashboardPanelWithLensDashboardApp(t *testing.T, dashboardID string) error {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return fmt.Errorf("failed to create Kibana scoped client: %w", err)
	}

	kibanaClient, diags := client.GetKibanaOapiClient()
	if diags.HasError() {
		return fmt.Errorf("failed to get Kibana OAPI client: %s", diags[0].Summary())
	}

	getURL := fmt.Sprintf("%s/api/dashboards/%s", kibanaClient.URL, dashboardID)
	getReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, getURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %w", err)
	}
	getReq.Header.Set("kbn-xsrf", "true")

	getResp, err := kibanaClient.HTTP.Do(getReq)
	if err != nil {
		return fmt.Errorf("failed to GET dashboard: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(getResp.Body, 4096))
		return fmt.Errorf("GET dashboard returned status %d: %s", getResp.StatusCode, string(body))
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to decode GET response: %w", err)
	}

	data := body.Data
	if data == nil {
		data = map[string]any{}
	}
	data["panels"] = []map[string]any{
		{
			"id":   "tf-acc-lens-app-1",
			"type": "lens-dashboard-app",
			"grid": map[string]any{
				"x": 0,
				"y": 0,
				"w": 48,
				"h": 15,
			},
			"config": map[string]any{
				"time_range": map[string]any{"from": "now-15m", "to": "now"},
				"title":      "API lens-dashboard-app metric",
				"type":       "metric",
				"data_source": map[string]any{
					"type":          "data_view_spec",
					"index_pattern": "metrics-*",
					"time_field":    "@timestamp",
				},
				"query": map[string]any{"language": "kql", "expression": ""},
				"metrics": []map[string]any{
					{
						"type":      "primary",
						"operation": "count",
						"format":    map[string]any{"type": "number"},
					},
				},
			},
		},
	}

	putBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal PUT body: %w", err)
	}

	putURL := fmt.Sprintf("%s/api/dashboards/%s", kibanaClient.URL, dashboardID)
	putReq, err := http.NewRequestWithContext(context.Background(), http.MethodPut, putURL, bytes.NewReader(putBody))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("kbn-xsrf", "true")

	putResp, err := kibanaClient.HTTP.Do(putReq)
	if err != nil {
		return fmt.Errorf("failed to PUT dashboard: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(putResp.Body, 4096))
		return fmt.Errorf("PUT dashboard returned status %d: %s", putResp.StatusCode, string(body))
	}

	return nil
}
