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

package managedintegration

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// cloudProxyResponseHeaders are HTTP response headers injected by Elastic
// Cloud's edge proxy layer ("Found" is Elastic Cloud's legacy internal
// codename, retained in these header names) for every request it fronts on
// behalf of an Elastic Cloud Hosted or Serverless deployment. A self-managed
// (on-premises / customer-hosted) Kibana never produces these headers: they
// are added by infrastructure that sits in front of Kibana on Elastic's own
// cloud, not by Kibana's application code, so there is no kibana.yml setting
// a self-managed operator could enable to fabricate them.
var cloudProxyResponseHeaders = []string{
	"X-Found-Handling-Cluster",
	"X-Found-Handling-Instance",
}

// statusBodyDTO is a minimal DTO for the one field of Kibana's
// `GET /api/status` response body this preflight check needs. It
// deliberately duplicates (rather than imports) the shape used by
// internal/clients/kibanaoapi.GetKibanaStatus, whose DTO type is unexported.
type statusBodyDTO struct {
	Version struct {
		BuildFlavor *string `json:"build_flavor"`
	} `json:"version"`
}

// DetectCloudSignals implements Task 6.2 of the fleet-agentless-policy
// OpenSpec change (design.md Decision 7; specs/fleet-agentless-policy/spec.md
// "Deployment topology preflight check"). It makes a single call to Kibana's
// own `GET /api/status` endpoint and reports the two independent signals
// that distinguish an Elastic Cloud Hosted or Serverless deployment from a
// self-managed one, plus whether the probe itself completed successfully.
//
// This is the single shared implementation of the cloud/serverless probe in
// this package: checkDeploymentTopology below wraps it with a fail-open
// production policy, and acc_helpers_test.go's isConfirmedCloudOrServerless
// wraps it with the opposite, fail-closed policy for deciding whether to
// skip acceptance tests. It deliberately duplicates (rather than imports)
// the response-body shape used by internal/clients/kibanaoapi.GetKibanaStatus,
// whose DTO type is unexported and whose caching/version-gating machinery
// is out of scope for this narrowly-scoped preflight probe.
//
// Detection approach chosen, and why (see design.md Open Question 5 for the
// alternatives considered):
//
//  1. Serverless is already unambiguously reported by Kibana itself via
//     `version.build_flavor == "serverless"` -- the same signal
//     clients.KibanaScopedClient.EnforceMinVersion already relies on to
//     short-circuit version checks for serverless deployments (see
//     internal/clients/version_utils.go). This preflight reuses that same
//     signal rather than re-deriving it.
//  2. Cloud Hosted and self-managed both report build_flavor "traditional",
//     so flavor alone cannot tell them apart (confirmed empirically below).
//     The additional signal used here is the presence of the
//     `X-Found-Handling-Cluster` / `X-Found-Handling-Instance` HTTP response
//     headers on the `/api/status` response, injected by Elastic Cloud's
//     edge proxy in front of every Cloud Hosted (and Serverless) deployment.
//     This is an infrastructure-level signal external to Kibana's own
//     plugins, so -- unlike a heuristic based on which Kibana plugins happen
//     to be registered -- it is not sensitive to Kibana-version-specific
//     plugin churn and cannot be toggled via kibana.yml.
//  3. Two alternatives from Open Question 5 were considered and ruled out:
//     (a) reading the `xpack.fleet.agentless.enabled` kibana.yml flag --
//     there is no read API for it (it is not an Elasticsearch cluster
//     setting, and no Fleet settings endpoint echoes it); (b) a `dry_run`
//     preflight POST to `/api/fleet/managed_integrations` -- there is no
//     dry_run parameter on that endpoint in the generated kbapi client, and
//     empirically POSTing malformed bodies to it returns byte-for-byte
//     identical validation errors on a self-managed and a Cloud Hosted
//     Kibana (schema validation runs before any topology-aware business
//     logic), so the endpoint itself cannot be used as a non-destructive
//     topology probe.
//
// Empirical verification (2026-07-01, see the fleet-agentless-policy
// OpenSpec change's Task 6 report): against a live Kibana 9.4.3 Elastic
// Cloud Hosted deployment, `GET /api/status` returned build_flavor
// "traditional" with both X-Found-Handling-* headers present -- classified
// cloud, preflight passes. Against a self-managed docker-compose Kibana
// 9.4.0, the same call returned "traditional" with neither header present --
// classified self-managed, preflight fails closed.
//
// Returns:
//   - serverless: true if `version.build_flavor == "serverless"` was observed.
//   - cloudProxied: true if either X-Found-Handling-* header was observed.
//   - ok: false if the probe itself did not complete (network error,
//     non-200, or a malformed response body) -- callers should treat this as
//     "inconclusive", not as "self-managed", and apply their own fallback
//     policy. When ok is false, serverless and cloudProxied are always false.
func DetectCloudSignals(ctx context.Context, client *clients.KibanaScopedClient) (serverless bool, cloudProxied bool, ok bool) {
	oapi := client.GetKibanaOapiClient()
	if oapi == nil || oapi.API == nil {
		// No Kibana OpenAPI client to probe with. This should not happen in
		// practice (ProviderClientFactory.GetKibanaClient validates endpoint
		// presence before handing out a scoped client), but if it does, we
		// have no basis to classify the deployment.
		return false, false, false
	}

	resp, err := oapi.API.GetStatusWithResponse(ctx, &kbapi.GetStatusParams{})
	if err != nil || resp == nil || resp.HTTPResponse == nil || resp.StatusCode() != http.StatusOK {
		// Inconclusive: the status probe itself failed or did not complete.
		return false, false, false
	}

	var dto statusBodyDTO
	if jsonErr := json.Unmarshal(resp.Body, &dto); jsonErr != nil {
		// Inconclusive: malformed/unrecognized response body.
		return false, false, false
	}

	serverless = dto.Version.BuildFlavor != nil && *dto.Version.BuildFlavor == "serverless"

	for _, header := range cloudProxyResponseHeaders {
		if resp.HTTPResponse.Header.Get(header) != "" {
			cloudProxied = true
			break
		}
	}

	return serverless, cloudProxied, true
}

// checkDeploymentTopology is the production, fail-open policy wrapper around
// DetectCloudSignals. It returns error diagnostics ONLY when the probe
// completed successfully AND positively confirms the connected Kibana is
// self-managed (non-cloud). Every other outcome -- confirmed
// cloud/serverless, or a probe that could not be completed or parsed --
// returns nil (fail open), per Decision 7's explicit "fail open when
// inconclusive" fallback. This function is called from createAgentlessPolicy
// only (Create-time preflight; Read/Update/Delete are unaffected, per
// tasks.md section 6).
func checkDeploymentTopology(ctx context.Context, client *clients.KibanaScopedClient) diag.Diagnostics {
	serverless, cloudProxied, ok := DetectCloudSignals(ctx, client)
	if !ok || serverless || cloudProxied {
		// Inconclusive, or confirmed cloud/serverless: fail open per
		// Decision 7. Do not block a potentially-legitimate cloud-hosted
		// setup on a transient or unrelated status-endpoint error. If the
		// deployment genuinely is unsupported, the subsequent POST to
		// /api/fleet/managed_integrations will surface its own error.
		return nil
	}

	// Neither a cloud/serverless flavor nor an Elastic Cloud proxy header was
	// observed on an otherwise well-formed status response: positively
	// classify this as a self-managed deployment and fail closed.
	var diags diag.Diagnostics
	diags.AddError(
		"Unsupported deployment topology",
		"Fleet managed integrations require Elastic Cloud Hosted or Serverless; this Kibana deployment appears to be "+
			"self-managed. The Fleet managed integrations API provisions agent runtime capacity in Elastic's own cloud "+
			"infrastructure, so it is only functional on Elastic Cloud Hosted and Serverless (Security or "+
			"Observability) deployments -- see the Kibana 9.5.0+ Fleet managed integrations documentation. "+
			"Self-managed (on-premises) Kibana is not supported.",
	)
	return diags
}
