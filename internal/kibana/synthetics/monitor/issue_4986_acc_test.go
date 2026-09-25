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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue4986 reproduces https://github.com/elastic/terraform-provider-elasticstack/issues/4986:
//
// Kibana can return a synthetics monitor CREATE response whose `type` field is
// omitted/null even though the monitor was created successfully server-side.
// `toModelV0` in schema.go switches on that response type, falls through to
// the default case, and returns "unsupported monitor type: ". Because the
// error surfaces after the monitor already exists in Kibana, Terraform never
// records it in state -- the monitor is orphaned.
//
// Real Kibana always echoes `type` back deterministically, so the omission
// can't be triggered directly against the live stack. Instead, this test
// puts a reverse proxy in front of the real Kibana endpoint that forwards
// every request unmodified except the monitor CREATE response, from which it
// strips the `type` field before the provider parses it. Every other call
// (Fleet agent policy, private location, and the monitor CREATE request
// itself) still goes to the real, live Kibana/Fleet stack.
func TestAccReproduceIssue4986(t *testing.T) {
	acctest.PreCheck(t)
	versionutils.SkipIfUnsupported(t, minKibanaVersion, versionutils.FlavorAny)

	realEndpoint := strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))
	if realEndpoint == "" {
		t.Fatal("KIBANA_ENDPOINT must be set for this test to run")
	}

	// Resolve a client against the *real* Kibana endpoint before KIBANA_ENDPOINT
	// is overridden below, so the orphaned monitor can be deleted directly.
	cleanupClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to build cleanup Kibana client: %s", err)
	}

	target, err := url.Parse(realEndpoint)
	if err != nil {
		t.Fatalf("failed to parse KIBANA_ENDPOINT %q: %s", realEndpoint, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	baseDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		baseDirector(req)
		req.Host = target.Host
		// Let net/http.Transport negotiate compression itself so it also
		// transparently decompresses the response before ModifyResponse runs.
		req.Header.Del("Accept-Encoding")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.Request == nil ||
			resp.Request.Method != http.MethodPost ||
			!strings.Contains(resp.Request.URL.Path, "/api/synthetics/monitors") ||
			(resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) {
			return nil
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}

		var payload map[string]any
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			// Not JSON (or not the shape we expect) -- pass through untouched.
			resp.Body = io.NopCloser(bytes.NewReader(body))
			resp.ContentLength = int64(len(body))
			return nil
		}

		if id, ok := payload["id"].(string); ok {
			// The provider fails to parse this mutated response, so it never
			// records the monitor in state -- exactly the orphaning described
			// in the issue. Delete it against the real Kibana endpoint right
			// away so it doesn't block the test framework's automatic destroy
			// of the private location that references it.
			if diags := kibanaoapi.DeleteMonitor(context.Background(), cleanupClient.GetKibanaOapiClient(), "default", id); diags.HasError() {
				t.Logf("failed to delete orphaned monitor %s created against the real Kibana endpoint: %v", id, diags)
			}
		}

		// Simulate the Kibana CREATE response omitting `type`, as reported in #4986.
		delete(payload, "type")

		mutated, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return marshalErr
		}

		resp.Body = io.NopCloser(bytes.NewReader(mutated))
		resp.ContentLength = int64(len(mutated))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mutated)))
		return nil
	}

	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	t.Setenv("KIBANA_ENDPOINT", proxyServer.URL)

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccIssue4986Config(name),
				ExpectError: regexp.MustCompile(`unsupported monitor type`),
			},
		},
	})
}

func testAccIssue4986Config(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "issue_4986" {
  name               = "Issue 4986 Agent Policy - %[1]s"
  namespace          = "testacc"
  description        = "Issue 4986 Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.issue_4986.source_id
}

resource "elasticstack_fleet_agent_download_source" "issue_4986" {
  name      = "Issue 4986 Agent Download Source - %[1]s"
  source_id = "agent-download-source-issue-4986-%[1]s"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_kibana_synthetics_private_location" "issue_4986" {
  label           = "issue-4986-pl-%[1]s"
  agent_policy_id = elasticstack_fleet_agent_policy.issue_4986.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "issue_4986" {
  name              = "Issue 4986 Monitor - %[1]s"
  private_locations = [elasticstack_kibana_synthetics_private_location.issue_4986.label]
  http = {
    url = "http://localhost:5601"
  }
}
`, name)
}
