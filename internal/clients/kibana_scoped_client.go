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

package clients

import (
	"context"
	"sync"

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// KibanaScopedClient is a typed client surface for Kibana and Fleet operations.
// It exposes: Kibana OpenAPI client, Fleet client, and Kibana-derived
// version/flavor checks.
//
// It deliberately does NOT expose provider-level Elasticsearch identity so that
// version and identity checks always resolve against the scoped Kibana
// connection rather than the provider-level Elasticsearch cluster.
type KibanaScopedClient struct {
	kibanaOapi *kibanaoapi.Client
	fleet      *fleetclient.Client
	// version is the provider version string used to tag API user-agent headers.
	version string
	// kibanaEndpoint holds the resolved Kibana endpoint URL captured after
	// provider configuration, entity-local overrides, and environment overrides
	// have been applied. It is used by factory endpoint validation.
	kibanaEndpoint string
	// fleetEndpoint holds the resolved Fleet endpoint URL captured after
	// provider configuration, entity-local overrides, and environment overrides
	// have been applied. For Fleet, the value reflects the already-resolved
	// cfg.Fleet endpoint which may have been inherited from the Kibana-derived
	// config path.
	fleetEndpoint string

	// statusMu guards the cached server status fields. The cache is bounded to
	// the lifetime of a single KibanaScopedClient instance (the factory builds
	// one per Create/Read/Update), which is the natural scope for "the server
	// version cannot change underneath us".
	statusMu sync.Mutex
	// statusCached is true once getServerStatusRaw has produced a successful
	// (rawVersion, flavor) pair for this client instance. Errors are never
	// cached so transient failures recover on the next call.
	statusCached bool
	statusVer    string
	statusFlavor string
}

// GetKibanaOapiClient returns the Kibana OpenAPI client.
//
// Endpoint presence is validated by ProviderClientFactory.GetKibanaClient before
// a scoped client is returned.
func (k *KibanaScopedClient) GetKibanaOapiClient() *kibanaoapi.Client {
	return k.kibanaOapi
}

// GetFleetClient returns the Fleet client.
//
// Endpoint presence is validated by ProviderClientFactory.GetKibanaClient before
// a scoped client is returned.
func (k *KibanaScopedClient) GetFleetClient() *fleetclient.Client {
	return k.fleet
}

// getServerStatusRaw fetches the Kibana server status, returning the raw version
// string and build flavor. The successful result is cached for the lifetime of
// this client instance so that callers performing multiple version-gated
// decisions during one resource operation share a single `/api/status` round
// trip. Errors are not cached; a subsequent call will re-attempt the fetch so
// transient failures recover naturally.
//
// The mutex is held for the duration of the fetch so concurrent callers wait
// for the in-flight request and then observe the cached result instead of
// issuing parallel requests.
func (k *KibanaScopedClient) getServerStatusRaw(ctx context.Context) (rawVersion string, flavor string, diags fwdiag.Diagnostics) {
	k.statusMu.Lock()
	defer k.statusMu.Unlock()

	if k.statusCached {
		return k.statusVer, k.statusFlavor, nil
	}

	if k.GetKibanaOapiClient() == nil {
		return "", "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"Kibana OpenAPI client not configured",
			"the scoped client was not produced by ProviderClientFactory.GetKibanaClient; this is a provider bug — please report it",
		)}
	}

	rawVersion, flavor, diags = kibanaoapi.GetKibanaStatus(ctx, k.GetKibanaOapiClient().API)
	if diags.HasError() {
		return "", "", diags
	}

	k.statusCached = true
	k.statusVer = rawVersion
	k.statusFlavor = flavor
	return rawVersion, flavor, nil
}

// EnforceMinVersion returns true when the Kibana server version is greater than
// or equal to minVersion, or when the server is running in serverless mode.
// Same-core Kibana -SNAPSHOT builds satisfy a release minimum (e.g. 9.5.0-SNAPSHOT
// meets floor 9.5.0); see kibanaVersionAtLeastRelease in version_utils.go.
// If minVersion is nil, no minimum is enforced and the method returns true.
func (k *KibanaScopedClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, fwdiag.Diagnostics) {
	return enforceMinVersion(ctx, minVersion, k.fetchVersion, kibanaVersionAtLeastRelease)
}

// EnforceVersionCheck returns true when the given version check function
// returns true, or when the server is running in serverless mode.
func (k *KibanaScopedClient) EnforceVersionCheck(ctx context.Context, check func(*version.Version) bool) (bool, fwdiag.Diagnostics) {
	return enforceVersionCheck(ctx, check, k.fetchVersion)
}

// fetchVersion adapts getServerStatusRaw into the versionFetcher signature
// expected by the shared enforceMinVersion and enforceVersionCheck helpers.
func (k *KibanaScopedClient) fetchVersion(ctx context.Context) (rawVersion, flavor string, diags fwdiag.Diagnostics) {
	return k.getServerStatusRaw(ctx)
}

// kibanaScopedClientFromAPIClient constructs a KibanaScopedClient from the
// Kibana-related fields of an *apiClient. This is the canonical adapter used by
// the factory and by NewAcceptanceTestingKibanaScopedClient.
func kibanaScopedClientFromAPIClient(a *apiClient) *KibanaScopedClient {
	return &KibanaScopedClient{
		kibanaOapi:     a.kibanaOapi,
		fleet:          a.fleet,
		version:        a.version,
		kibanaEndpoint: a.kibanaEndpoint,
		fleetEndpoint:  a.fleetEndpoint,
	}
}

// NewAcceptanceTestingKibanaScopedClient builds a KibanaScopedClient for
// acceptance tests by reusing the acceptance testing APIClient.
func NewAcceptanceTestingKibanaScopedClient() (*KibanaScopedClient, error) {
	apiClient, err := newAcceptanceTestingClient()
	if err != nil {
		return nil, err
	}
	return kibanaScopedClientFromAPIClient(apiClient), nil
}
