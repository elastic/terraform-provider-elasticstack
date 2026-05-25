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

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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
	// have been applied. It is used by accessor validation to distinguish missing
	// endpoint configuration from unexpected nil states.
	kibanaEndpoint string
	// fleetEndpoint holds the resolved Fleet endpoint URL captured after
	// provider configuration, entity-local overrides, and environment overrides
	// have been applied. For Fleet, the value reflects the already-resolved
	// cfg.Fleet endpoint which may have been inherited from the Kibana-derived
	// config path.
	fleetEndpoint string
}

// GetKibanaOapiClient returns the Kibana OpenAPI client.
func (k *KibanaScopedClient) GetKibanaOapiClient() (*kibanaoapi.Client, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	if k.kibanaEndpoint == "" {
		diags.AddError("kibana OpenAPI client is not configured: set kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT", "")
		return nil, diags
	}
	if k.kibanaOapi == nil {
		diags.AddError("kibanaoapi client not found", "")
		return nil, diags
	}
	return k.kibanaOapi, diags
}

// GetFleetClient returns the Fleet client.
func (k *KibanaScopedClient) GetFleetClient() (*fleetclient.Client, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	if k.fleetEndpoint == "" {
		const fleetMsg = "fleet client is not configured: set fleet.endpoint or FLEET_ENDPOINT, " +
			"or configure kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT " +
			"for inherited Fleet endpoint resolution"
		diags.AddError("Fleet client not configured", fleetMsg)
		return nil, diags
	}
	if k.fleet == nil {
		diags.AddError("Fleet client not found", "")
		return nil, diags
	}
	return k.fleet, diags
}

// getServerStatusRaw fetches the Kibana server status, returning the raw version
// string and build flavor in a single HTTP call. It centralises GetKibanaOapiClient
// setup and the GetKibanaStatus call so callers share one code path and one request.
func (k *KibanaScopedClient) getServerStatusRaw(ctx context.Context) (rawVersion string, flavor string, diags fwdiag.Diagnostics) {
	oapiClient, d := k.GetKibanaOapiClient()
	diags.Append(d...)
	if diags.HasError() {
		return "", "", diags
	}

	rawVersion, flavor, diags = kibanaoapi.GetKibanaStatus(ctx, oapiClient.API)
	return rawVersion, flavor, diags
}

// EnforceMinVersion returns true when the Kibana server version is greater than
// or equal to minVersion, or when the server is running in serverless mode.
func (k *KibanaScopedClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, fwdiag.Diagnostics) {
	rawVersion, flavor, diags := k.getServerStatusRaw(ctx)
	if diags.HasError() {
		return false, diags
	}

	if flavor == ServerlessFlavor {
		return true, nil
	}

	serverVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		return false, diagutil.FrameworkDiagFromError(err)
	}

	return serverVersion.GreaterThanOrEqual(minVersion), nil
}

// EnforceVersionCheck returns true when the given version check function
// returns true, or when the server is running in serverless mode.
func (k *KibanaScopedClient) EnforceVersionCheck(ctx context.Context, check func(*version.Version) bool) (bool, fwdiag.Diagnostics) {
	rawVersion, flavor, diags := k.getServerStatusRaw(ctx)
	if diags.HasError() {
		return false, diags
	}

	if flavor == ServerlessFlavor {
		return true, nil
	}

	sv, err := version.NewVersion(rawVersion)
	if err != nil {
		return false, diagutil.FrameworkDiagFromError(err)
	}

	return check(sv), nil
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
