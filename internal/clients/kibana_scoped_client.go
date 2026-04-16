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
	"errors"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// KibanaScopedClient is a typed client surface for Kibana and Fleet operations.
// It exposes: Kibana legacy client, Kibana OpenAPI client, SLO client, Fleet
// client, Kibana auth-context helpers, and Kibana-derived version/flavor checks.
//
// It deliberately does NOT expose provider-level Elasticsearch identity so that
// version and identity checks always resolve against the scoped Kibana
// connection rather than the provider-level Elasticsearch cluster.
type KibanaScopedClient struct {
	kibana       *kibana.Client
	kibanaOapi   *kibanaoapi.Client
	sloAPI       slo.SloAPI
	kibanaConfig kibana.Config
	fleet        *fleetclient.Client
	// version is the provider version string used to tag API user-agent headers.
	version string
}

// GetKibanaClient returns the Kibana legacy client.
func (k *KibanaScopedClient) GetKibanaClient() (*kibana.Client, error) {
	if k.kibana == nil {
		return nil, errors.New("kibana client not found")
	}
	return k.kibana, nil
}

// GetKibanaOapiClient returns the Kibana OpenAPI client.
func (k *KibanaScopedClient) GetKibanaOapiClient() (*kibanaoapi.Client, error) {
	if k.kibanaOapi == nil {
		return nil, errors.New("kibanaoapi client not found")
	}
	return k.kibanaOapi, nil
}

// GetSloClient returns the SLO client.
func (k *KibanaScopedClient) GetSloClient() (slo.SloAPI, error) {
	if k.sloAPI == nil {
		return nil, errors.New("slo client not found")
	}
	return k.sloAPI, nil
}

// GetFleetClient returns the Fleet client.
func (k *KibanaScopedClient) GetFleetClient() (*fleetclient.Client, error) {
	if k.fleet == nil {
		return nil, errors.New("fleet client not found")
	}
	return k.fleet, nil
}

// SetSloAuthContext injects authentication credentials into the context for SLO
// API calls. Credentials are derived from the scoped Kibana connection.
func (k *KibanaScopedClient) SetSloAuthContext(ctx context.Context) context.Context {
	if k.kibanaConfig.ApiKey != "" {
		return context.WithValue(ctx, slo.ContextAPIKeys, map[string]slo.APIKey{
			"apiKeyAuth": {
				Prefix: "ApiKey",
				Key:    k.kibanaConfig.ApiKey,
			}})
	}

	return context.WithValue(ctx, slo.ContextBasicAuth, slo.BasicAuth{
		UserName: k.kibanaConfig.Username,
		Password: k.kibanaConfig.Password,
	})
}

// ServerVersion returns the version of the Kibana server. Version is always
// derived from the Kibana status API; there is no Elasticsearch fallback.
func (k *KibanaScopedClient) ServerVersion(_ context.Context) (*version.Version, diag.Diagnostics) {
	kibClient, err := k.GetKibanaClient()
	if err != nil {
		return nil, diag.Errorf("failed to get version from Kibana API: %s, "+
			"please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	status, err := kibClient.KibanaStatus.Get()
	if err != nil {
		return nil, diag.Errorf("failed to get version from Kibana API: %s, "+
			"Please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	vMap, ok := status["version"].(map[string]any)
	if !ok {
		return nil, diag.Errorf("failed to get version from Kibana API")
	}

	rawVersion, ok := vMap["number"].(string)
	if !ok {
		return nil, diag.Errorf("failed to get version number from Kibana status")
	}

	serverVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return serverVersion, nil
}

// ServerFlavor returns the flavor (e.g. "serverless", "default") of the Kibana
// server. Flavor is always derived from the Kibana status API.
func (k *KibanaScopedClient) ServerFlavor(_ context.Context) (string, diag.Diagnostics) {
	kibClient, err := k.GetKibanaClient()
	if err != nil {
		return "", diag.Errorf("failed to get flavor from Kibana API: %s, "+
			"please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	status, err := kibClient.KibanaStatus.Get()
	if err != nil {
		return "", diag.Errorf("failed to get flavor from Kibana API: %s, "+
			"Please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	vMap, ok := status["version"].(map[string]any)
	if !ok {
		return "", diag.Errorf("failed to get flavor from Kibana API")
	}

	serverFlavor, ok := vMap["build_flavor"].(string)
	if !ok {
		// build_flavor field is not present in older Kibana versions (pre-serverless)
		// Default to empty string to indicate traditional/stateful deployment
		return "", nil
	}

	return serverFlavor, nil
}

// EnforceMinVersion returns true when the Kibana server version is greater than
// or equal to minVersion, or when the server is running in serverless mode.
func (k *KibanaScopedClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	flavor, diags := k.ServerFlavor(ctx)
	if diags.HasError() {
		return false, diags
	}

	if flavor == ServerlessFlavor {
		return true, nil
	}

	serverVersion, diags := k.ServerVersion(ctx)
	if diags.HasError() {
		return false, diags
	}

	return serverVersion.GreaterThanOrEqual(minVersion), nil
}

// kibanaScopedClientFromAPIClient constructs a KibanaScopedClient from the
// Kibana-related fields of an *apiClient. This is the canonical adapter used by
// the factory and by NewAcceptanceTestingKibanaScopedClient.
func kibanaScopedClientFromAPIClient(a *apiClient) *KibanaScopedClient {
	return &KibanaScopedClient{
		kibana:       a.kibana,
		kibanaOapi:   a.kibanaOapi,
		sloAPI:       a.slo,
		kibanaConfig: a.kibanaConfig,
		fleet:        a.fleet,
		version:      a.version,
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
