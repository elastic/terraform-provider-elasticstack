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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccResourceDashboardUnknownPanel tests that unknown panel types (e.g. `discover_session`)
// are preserved during import and produce a no-op plan.
//
// The dashboard is created via the Kibana REST API (not Terraform config) because
// unknown panel types cannot be authored in Terraform.
func TestAccResourceDashboardUnknownPanel(t *testing.T) {
	dashboardTitle := "Test Dashboard Unknown Panel " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	var dashboardID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				PreConfig: func() {
					id, err := createDashboardWithUnknownPanel(dashboardTitle)
					if err != nil {
						t.Fatalf("Failed to create dashboard with unknown panel: %v", err)
					}
					dashboardID = id
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName: "elasticstack_kibana_dashboard.test",
				ImportState:  true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					if dashboardID == "" {
						return "", fmt.Errorf("dashboardID not set; PreConfig may not have run")
					}
					return dashboardID, nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"time_range.mode",
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "discover_session"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "48"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.config_json"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.config_json", regexp.MustCompile(`"timeRange"\s*:\s*\{`)),
				),
			},
			// Verify no-op plan after import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// createDashboardWithUnknownPanel creates a dashboard via the Kibana REST API containing a
// `discover_session` panel (an unknown panel type that the provider doesn't have a typed
// config block for).
func createDashboardWithUnknownPanel(title string) (string, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return "", fmt.Errorf("failed to create Kibana scoped client: %w", err)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return "", fmt.Errorf("failed to get Kibana OAPI client: %w", err)
	}

	// Build the dashboard payload with an unknown panel type (discover_session)
	body := map[string]any{
		"title":       title,
		"description": "Acceptance test dashboard for unknown panel type preservation",
		"timeRange": map[string]any{
			"from": "now-15m",
			"to":   "now",
		},
		"refreshInterval": map[string]any{
			"pause": true,
			"value": 0,
		},
		"query": map[string]any{
			"language":   "kql",
			"expression": "",
		},
		"panels": []map[string]any{
			{
				"id":   "tf-acc-discover-1",
				"type": "discover_session",
				"grid": map[string]any{
					"x": 0,
					"y": 0,
					"w": 48,
					"h": 15,
				},
				"config": map[string]any{
					"timeRange": map[string]any{
						"from": "now-30d",
						"to":   "now",
					},
					"columns": []string{"_source"},
					"sort":    [][]any{{"@timestamp", "desc"}},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal dashboard body: %w", err)
	}

	url := fmt.Sprintf("%sapi/dashboards/dashboard?allowUnmappedKeys=true", kibanaClient.URL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("unexpected status %d creating dashboard: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode dashboard response: %w", err)
	}

	if result.ID == "" {
		return "", fmt.Errorf("dashboard created but returned empty ID")
	}

	return result.ID, nil
}