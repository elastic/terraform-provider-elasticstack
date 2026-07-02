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

package agentlesspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

// cloudProxyResponseHeaders mirrors topology.go's unexported
// cloudProxyResponseHeaders var. It cannot be imported directly (this file
// is in the agentlesspolicy_test external test package, and the production
// var is unexported), and duplicating a two-string-literal slice is simpler
// and lower-risk than exporting an internal detail of topology.go purely for
// test consumption. See isConfirmedCloudOrServerless's doc comment for why
// this file mirrors, rather than reuses, topology.go's detection logic.
var cloudProxyResponseHeaders = []string{
	"X-Found-Handling-Cluster",
	"X-Found-Handling-Instance",
}

// isConfirmedCloudOrServerless makes the same GET /api/status call, and
// checks the same two signals, as topology.go's checkDeploymentTopology
// (build_flavor == "serverless", and the X-Found-Handling-* cloud-proxy
// headers) -- but is a deliberately separate, test-only mirror of that
// logic, not a shared code path. It exists so skipUnlessConfirmedCloud below
// can invert checkDeploymentTopology's risk tolerance for this call site:
//
//   - checkDeploymentTopology (topology.go) is the resource's own production
//     preflight. It fails OPEN on any ambiguity -- inconclusive probe,
//     network error, malformed body -- because wrongly blocking a legitimate
//     Cloud Hosted/Serverless user's apply is the worse outcome.
//   - isConfirmedCloudOrServerless is the inverse: it returns true ONLY when
//     one of the two signals is positively observed on a well-formed 200
//     response. Every other outcome (network error, non-200, missing
//     signals, malformed body) returns false, i.e. "not confirmed cloud".
//     This is deliberately fail-CLOSED, because the consumer
//     (skipUnlessConfirmedCloud) uses false to SKIP a test -- a cheap, safe
//     outcome -- rather than to block a real user's apply.
//
// These are two different call sites with intentionally opposite
// conservative defaults for the same underlying signal; see this repo's PR
// #4034 discussion for why that is correct, not a contradiction.
func isConfirmedCloudOrServerless(ctx context.Context, client *clients.KibanaScopedClient) bool {
	oapi := client.GetKibanaOapiClient()
	if oapi == nil || oapi.API == nil {
		return false
	}

	resp, err := oapi.API.GetStatusWithResponse(ctx, &kbapi.GetStatusParams{})
	if err != nil || resp == nil || resp.HTTPResponse == nil || resp.StatusCode() != http.StatusOK {
		return false
	}

	var dto struct {
		Version struct {
			BuildFlavor *string `json:"build_flavor"`
		} `json:"version"`
	}
	if jsonErr := json.Unmarshal(resp.Body, &dto); jsonErr != nil {
		return false
	}

	if dto.Version.BuildFlavor != nil && *dto.Version.BuildFlavor == "serverless" {
		return true
	}

	for _, header := range cloudProxyResponseHeaders {
		if resp.HTTPResponse.Header.Get(header) != "" {
			return true
		}
	}

	return false
}

// skipUnlessConfirmedCloud skips t unless isConfirmedCloudOrServerless
// positively confirms the acceptance-testing Kibana connection is Elastic
// Cloud Hosted or Serverless.
//
// Fleet agentless policies only function against Elastic Cloud Hosted or
// Serverless (Kibana 9.3.0+; see topology.go's checkDeploymentTopology), but
// this repo's CI (.github/workflows/provider.yml) runs every
// acceptance-test matrix job against a self-managed stack (`make
// docker-fleet`) -- there is no Cloud Hosted/Serverless CI lane. Without
// this check, checkDeploymentTopology's own preflight correctly rejects that
// self-managed stack, which would FAIL these tests in CI. Skipping instead
// is the correct outcome for an environment that structurally cannot run
// them -- the same conceptual situation as versionutils.SkipIfUnsupported
// skipping a test against a too-old Kibana rather than failing it.
//
// Deliberately not gated behind a manual opt-in environment variable: that
// alternative was considered and rejected (it would require new CI wiring
// -- or a human to remember to set it -- to ever exercise these tests for
// real; a positive, automatic environment probe needs neither).
func skipUnlessConfirmedCloud(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		// Not actually running acceptance tests -- resource.Test's own
		// TF_ACC guard will skip the rest of the test regardless. Don't
		// spend a real network call probing a connection that may not even
		// be configured in this environment.
		return
	}

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Skipf("skipping: could not establish a Kibana connection to check deployment topology (%v); "+
			"this environment is not confirmed to be Elastic Cloud Hosted or Serverless, and agentless "+
			"policies require Cloud Hosted/Serverless", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !isConfirmedCloudOrServerless(ctx, client) {
		t.Skip("skipping: this environment is not a confirmed Elastic Cloud Hosted or Serverless deployment; " +
			"agentless policies require Cloud Hosted/Serverless and this repo's CI has no such lane -- run " +
			"against a real cloud-hosted Kibana to exercise this test")
	}
}
