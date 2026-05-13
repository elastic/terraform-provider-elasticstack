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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// createSearchSavedObjectForDiscoverRef provisions a `search` saved object used as the Discover
// artifact for discover_session by_reference panels (discover-session SO type is not supported on 9.4).
func createSearchSavedObjectForDiscoverRef(t *testing.T, id string) error {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return fmt.Errorf("kibana client: %w", err)
	}
	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return fmt.Errorf("kibana oapi client: %w", err)
	}

	searchSource := map[string]any{
		"query": map[string]any{
			"language": "kuery",
			"query":    "",
		},
		"filter": []any{},
	}
	searchSourceJSON, err := json.Marshal(searchSource)
	if err != nil {
		return fmt.Errorf("marshal searchSource for saved search: %w", err)
	}

	payload := map[string]any{
		"attributes": map[string]any{
			"title":   "tf-acc-discover-ref-" + id,
			"columns": []string{"@timestamp", "message"},
			"sort":    [][]string{{"@timestamp", "desc"}},
			"kibanaSavedObjectMeta": map[string]any{
				"searchSourceJSON": string(searchSourceJSON),
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/api/saved_objects/search/%s?overwrite=true", kibanaClient.URL, id)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("POST saved_objects/search: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("POST %s returned %d: %s", url, resp.StatusCode, string(body))
	}

	t.Cleanup(func() {
		deleteSearchSavedObjectForDiscoverRef(t, id)
	})
	return nil
}

// deleteSearchSavedObjectForDiscoverRef removes the `search` saved object created for discover_session
// by_reference acceptance tests. 404 is ignored (idempotent cleanup).
func deleteSearchSavedObjectForDiscoverRef(t *testing.T, id string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Logf("discover_session acc cleanup: kibana client: %v", err)
		return
	}
	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Logf("discover_session acc cleanup: kibana oapi client: %v", err)
		return
	}

	delURL := fmt.Sprintf("%s/api/saved_objects/search/%s", kibanaClient.URL, id)
	delReq, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, delURL, nil)
	if err != nil {
		t.Logf("discover_session acc cleanup: new DELETE request: %v", err)
		return
	}
	delReq.Header.Set("kbn-xsrf", "true")

	delResp, err := kibanaClient.HTTP.Do(delReq)
	if err != nil {
		t.Logf("discover_session acc cleanup: DELETE %s: %v", delURL, err)
		return
	}
	defer delResp.Body.Close()

	if delResp.StatusCode == http.StatusNotFound {
		return
	}
	if delResp.StatusCode != http.StatusOK && delResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(io.LimitReader(delResp.Body, 4096))
		t.Logf("discover_session acc cleanup: DELETE %s returned %d: %s", delURL, delResp.StatusCode, string(body))
	}
}

func TestAccResourceDashboardDiscoverSession_by_value_dsl(t *testing.T) {
	dashboardTitle := "Acc disc dsl " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("dsl"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "discover_session"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.discover_session_config.by_value.tab.dsl.query.expression"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.discover_session_config.by_value.tab.dsl.query.language", "kql"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("dsl"),
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

func TestAccResourceDashboardDiscoverSession_by_value_esql(t *testing.T) {
	dashboardTitle := "Acc disc esql " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "discover_session"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.discover_session_config.by_value.tab.esql.data_source_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
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

func TestAccResourceDashboardDiscoverSession_by_reference(t *testing.T) {
	dashboardTitle := "Acc disc ref " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	refID := "acc-disc-ref-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				PreConfig: func() {
					if err := createSearchSavedObjectForDiscoverRef(t, refID); err != nil {
						t.Fatalf("create search saved object: %v", err)
					}
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
					"discover_ref_id": config.StringVariable(refID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "discover_session"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.discover_session_config.by_reference.ref_id", refID),
					// Kibana may omit selected_tab_id on read for some references (e.g. legacy search SO).
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
					"discover_ref_id": config.StringVariable(refID),
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
