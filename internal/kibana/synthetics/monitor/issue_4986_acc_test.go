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

package monitor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

// stubSyntheticsServiceURL is the unreachable service URL of the us_west
// location served by the stub Synthetics Service manifest (see the Makefile).
const stubSyntheticsServiceURL = "http://127.0.0.1:1"

const issue4986ResourceName = "elasticstack_kibana_synthetics_monitor.issue_4986"

// TestAccReproduceIssue4986 covers https://github.com/elastic/terraform-provider-elasticstack/issues/4986.
//
// When the Synthetics Service reports push errors, Kibana's add-monitor API
// returns HTTP 200 with {"message":"error pushing monitor to the service",
// "attributes":{"errors":[...]},"id":"..."} instead of the monitor, and the
// edit-monitor API returns the same shape without the id. The stack's stub
// Synthetics Service makes every push fail, so both responses can be
// triggered against a live Kibana.
func TestAccReproduceIssue4986(t *testing.T) {
	acctest.PreCheck(t)
	versionutils.SkipIfUnsupported(t, minKibanaVersion, versionutils.FlavorAny)

	kibanaClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	oapiClient := kibanaClient.GetKibanaOapiClient()

	skipUnlessStubSyntheticsService(t, oapiClient)
	enableSyntheticsService(t, oapiClient)
	primeFailingPush(t, oapiClient)

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"name": config.StringVariable(name)}
	var monitorID string
	t.Cleanup(func() { deleteMonitorsMatching(t, oapiClient, name) })

	resource.Test(t, resource.TestCase{
		CheckDestroy: func(_ *terraform.State) error {
			monitor, diags := kibanaoapi.GetMonitor(context.Background(), oapiClient, "", monitorID)
			if diags.HasError() {
				return fmt.Errorf("failed to read monitor %s: %v", monitorID, diags)
			}
			if monitor != nil {
				return fmt.Errorf("monitor %s still exists", monitorID)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(issue4986ResourceName, "name", "Issue 4986 Monitor - "+name),
					resource.TestCheckResourceAttr(issue4986ResourceName, "http.url", "http://localhost:5601"),
					resource.TestCheckResourceAttr(issue4986ResourceName, "locations.#", "1"),
					resource.TestCheckResourceAttr(issue4986ResourceName, "locations.0", "us_west"),
					resource.TestCheckResourceAttr(issue4986ResourceName, "params", `{"foo":"bar"}`),
					checkMonitorMatchesState(oapiClient, &monitorID, "Issue 4986 Monitor - "+name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(issue4986ResourceName, "name", "Issue 4986 Monitor Updated - "+name),
					resource.TestCheckResourceAttr(issue4986ResourceName, "http.url", "http://localhost:5601/status"),
					resource.TestCheckResourceAttr(issue4986ResourceName, "params", `{"foo":"baz"}`),
					checkMonitorMatchesState(oapiClient, &monitorID, "Issue 4986 Monitor Updated - "+name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

// checkMonitorMatchesState verifies the monitor tracked in state exists in
// Kibana with the expected name, proving it was not orphaned.
func checkMonitorMatchesState(oapiClient *kibanaoapi.Client, monitorID *string, wantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[issue4986ResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", issue4986ResourceName)
		}
		compositeID, diags := synthetics.TryReadCompositeID(rs.Primary.ID)
		if diags.HasError() || compositeID == nil {
			return fmt.Errorf("unexpected monitor id %q", rs.Primary.ID)
		}
		*monitorID = compositeID.ResourceID

		monitor, getDiags := kibanaoapi.GetMonitor(context.Background(), oapiClient, compositeID.ClusterID, compositeID.ResourceID)
		if getDiags.HasError() {
			return fmt.Errorf("failed to read monitor %s: %v", rs.Primary.ID, getDiags)
		}
		if monitor == nil {
			return fmt.Errorf("monitor %s from state does not exist in Kibana", rs.Primary.ID)
		}
		if monitor.Name == nil || *monitor.Name != wantName {
			return fmt.Errorf("monitor %s has name %v, want %q", rs.Primary.ID, monitor.Name, wantName)
		}
		if monitor.Type == nil || *monitor.Type != kbapi.SyntheticsMonitorTypeHttp {
			return fmt.Errorf("monitor %s has type %v, want http", rs.Primary.ID, monitor.Type)
		}
		return nil
	}
}

// deleteMonitorsMatching removes monitors whose name contains the test's
// random name. Without the fix, create orphans the monitor, so the framework's
// destroy never removes it.
func deleteMonitorsMatching(t *testing.T, oapiClient *kibanaoapi.Client, name string) {
	t.Helper()
	ctx := context.Background()
	resp, err := oapiClient.API.GetSyntheticMonitorsWithResponse(ctx, &kbapi.GetSyntheticMonitorsParams{Query: &name})
	if err != nil {
		t.Logf("failed to list monitors matching %q for cleanup: %v", name, err)
		return
	}
	if resp.StatusCode() != http.StatusOK {
		t.Logf("failed to list monitors matching %q for cleanup: HTTP %d: %s", name, resp.StatusCode(), resp.Body)
		return
	}

	var parsed struct {
		Monitors []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"monitors"`
	}
	if err := json.Unmarshal(resp.Body, &parsed); err != nil {
		t.Logf("failed to parse monitors matching %q for cleanup: %v", name, err)
		return
	}
	for _, monitor := range parsed.Monitors {
		if !strings.Contains(monitor.Name, name) {
			continue
		}
		if diags := kibanaoapi.DeleteMonitor(ctx, oapiClient, "", monitor.ID); diags.HasError() {
			t.Logf("failed to delete monitor %s: %v", monitor.ID, diags)
		}
	}
}

func kibanaInternalRequest(t *testing.T, oapiClient *kibanaoapi.Client, method, path string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, strings.TrimRight(oapiClient.URL, "/")+path, nil)
	require.NoError(t, err)
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("x-elastic-internal-origin", "kibana")
	req.Header.Set("elastic-api-version", "1")

	resp, err := oapiClient.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, body
}

// skipUnlessStubSyntheticsService skips the test unless Kibana lists the stub
// Synthetics Service location configured by `make docker-fleet`.
func skipUnlessStubSyntheticsService(t *testing.T, oapiClient *kibanaoapi.Client) {
	t.Helper()
	status, body := kibanaInternalRequest(t, oapiClient, http.MethodGet, "/internal/uptime/service/locations")
	if status != http.StatusOK {
		t.Skipf("stub Synthetics Service not available: listing service locations returned HTTP %d", status)
	}

	var parsed struct {
		Locations []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"locations"`
	}
	require.NoError(t, json.Unmarshal(body, &parsed))
	for _, location := range parsed.Locations {
		if location.ID == "us_west" && location.URL == stubSyntheticsServiceURL {
			return
		}
	}
	t.Skip("stub Synthetics Service not available: Kibana does not list the stub us_west location")
}

// enableSyntheticsService creates the Synthetics Service API key Kibana needs
// before it pushes monitors to Elastic-managed locations.
func enableSyntheticsService(t *testing.T, oapiClient *kibanaoapi.Client) {
	t.Helper()
	status, body := kibanaInternalRequest(t, oapiClient, http.MethodPut, "/internal/synthetics/service/enablement")
	if status == http.StatusNotFound {
		status, body = kibanaInternalRequest(t, oapiClient, http.MethodPut, "/internal/uptime/service/enablement")
	}
	require.Equal(t, http.StatusOK, status, "failed to enable the Synthetics Service: %s", body)
}

// primeFailingPush creates monitors directly until Kibana returns the
// push-error body. Every push to the stub service fails, so once one create
// reports errors, every later create reports them too.
func primeFailingPush(t *testing.T, oapiClient *kibanaoapi.Client) {
	t.Helper()
	ctx := context.Background()
	for attempt := range 10 {
		req := kbapi.SyntheticsMonitorRequest{}
		require.NoError(t, req.FromSyntheticsHttpMonitorFields(kbapi.SyntheticsHttpMonitorFields{
			Name:      fmt.Sprintf("issue-4986-probe-%s-%d", sdkacctest.RandString(8), attempt),
			Type:      kbapi.SyntheticsHttpMonitorFieldsType(kbapi.SyntheticsMonitorTypeHttp),
			Url:       "http://localhost:5601",
			Locations: &[]string{"us_west"},
			Labels:    &map[string]string{},
		}))

		result, syncErrors, diags := kibanaoapi.CreateMonitor(ctx, oapiClient, "", req)
		require.False(t, diags.HasError(), "failed to create probe monitor: %v", diags)
		require.NotNil(t, result)
		require.NotNil(t, result.Id)
		probeID := *result.Id
		t.Cleanup(func() {
			if diags := kibanaoapi.DeleteMonitor(context.Background(), oapiClient, "", probeID); diags.HasError() {
				t.Logf("failed to delete probe monitor %s: %v", probeID, diags)
			}
		})

		if len(syncErrors) > 0 {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Kibana never reported Synthetics Service push errors for probe monitors")
}
