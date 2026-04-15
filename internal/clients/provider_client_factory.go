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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderClientFactory is the provider-scoped client-resolution surface
// injected into Plugin Framework ProviderData and SDK meta. Resources and data
// sources use the factory to obtain typed clients rather than consuming a broad
// *APIClient directly.
//
// During the Kibana/Fleet typed-client phase the factory exposes:
//   - Typed Kibana/Fleet resolution methods returning *KibanaScopedClient
//   - Transitional legacy Elasticsearch resolution methods returning *APIClient
//     so unconverted Elasticsearch entities continue to work unchanged
type ProviderClientFactory struct {
	// defaultClient holds provider-level clients built from the provider
	// configuration block. It is used as the fallback when an entity does not
	// configure a resource-local connection block.
	defaultClient *APIClient
}

// NewProviderClientFactory constructs a ProviderClientFactory wrapping the
// provided default client. This is called by the provider Configure method.
func NewProviderClientFactory(defaultClient *APIClient) *ProviderClientFactory {
	return &ProviderClientFactory{defaultClient: defaultClient}
}

// --- Typed Kibana / Fleet resolution methods ---

// GetKibanaClient resolves the effective *KibanaScopedClient for a Plugin
// Framework Kibana or Fleet entity. When kibanaConnList is empty or null the
// factory returns a typed client built from provider-level defaults. When the
// list contains a connection block, the factory returns a new typed scoped
// client whose Kibana legacy client, Kibana OpenAPI client, SLO client, and
// Fleet client are rebuilt from that scoped connection.
func (f *ProviderClientFactory) GetKibanaClient(ctx context.Context, kibanaConnList types.List) (*KibanaScopedClient, fwdiags.Diagnostics) {
	if f == nil || f.defaultClient == nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(
			"Provider not configured",
			"Expected configured provider client factory. Please report this issue to the provider developers.",
		)}
	}

	var kibConns []config.KibanaConnection
	if diags := kibanaConnList.ElementsAs(ctx, &kibConns, true); diags.HasError() {
		return nil, diags
	}

	if len(kibConns) == 0 {
		return kibanaScopedClientFromAPIClient(f.defaultClient), nil
	}

	cfg, diags := config.NewFromFrameworkKibanaResource(ctx, kibConns, f.defaultClient.version)
	if diags.HasError() {
		return nil, diags
	}

	return buildKibanaScopedClientFromConfig(*cfg, f.defaultClient.version)
}

// GetKibanaClientFromSDK resolves the effective *KibanaScopedClient for an SDK
// Kibana or Fleet entity. When the kibana_connection block is absent from d the
// factory returns a typed client built from provider-level defaults. When the
// block is configured a new typed scoped client is returned with all
// Kibana-derived clients rebuilt from the scoped connection.
func (f *ProviderClientFactory) GetKibanaClientFromSDK(d *schema.ResourceData) (*KibanaScopedClient, diag.Diagnostics) {
	if f == nil || f.defaultClient == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Provider not configured",
			Detail:   "Expected configured provider client factory. Please report this issue to the provider developers.",
		}}
	}

	resourceConfig, diags := config.NewFromSDKKibanaResource(d, f.defaultClient.version)
	if diags.HasError() {
		return nil, diags
	}

	if resourceConfig == nil {
		return kibanaScopedClientFromAPIClient(f.defaultClient), nil
	}

	scoped, fwDiags := buildKibanaScopedClientFromConfig(*resourceConfig, f.defaultClient.version)
	if fwDiags.HasError() {
		var sdkDiags diag.Diagnostics
		for _, d := range fwDiags {
			severity := diag.Error
			if d.Severity() == fwdiags.SeverityWarning {
				severity = diag.Warning
			}
			sdkDiags = append(sdkDiags, diag.Diagnostic{
				Severity: severity,
				Summary:  d.Summary(),
				Detail:   d.Detail(),
			})
		}
		return nil, sdkDiags
	}
	return scoped, nil
}

// --- Transitional legacy Elasticsearch resolution methods ---
//
// These methods preserve the existing broad *APIClient behavior for
// unconverted Elasticsearch entities. They are intentionally transitional and
// will be replaced in the follow-up Elasticsearch typed-client phase.

// GetElasticsearchClient resolves the effective *APIClient for a Plugin
// Framework Elasticsearch entity, applying resource-local elasticsearch_connection
// if present. Mirrors the behavior of the previous MaybeNewAPIClientFromFrameworkResource.
func (f *ProviderClientFactory) GetElasticsearchClient(ctx context.Context, esConnList types.List) (*APIClient, fwdiags.Diagnostics) {
	return MaybeNewAPIClientFromFrameworkResource(ctx, esConnList, f.defaultClient)
}

// GetElasticsearchClientFromSDK resolves the effective *APIClient for an SDK
// Elasticsearch entity. Mirrors the behavior of the previous NewAPIClientFromSDKResource.
func (f *ProviderClientFactory) GetElasticsearchClientFromSDK(d *schema.ResourceData, meta any) (*APIClient, diag.Diagnostics) {
	return NewAPIClientFromSDKResource(d, meta)
}

// GetDefaultClient returns the provider-level default *APIClient.
// This method is for transitional use only; prefer the typed resolution methods above.
func (f *ProviderClientFactory) GetDefaultClient() *APIClient {
	return f.defaultClient
}

// --- Helper constructors ---

// buildKibanaScopedClientFromConfig builds a *KibanaScopedClient from a
// config.Client that has already been populated from a scoped kibana_connection.
func buildKibanaScopedClientFromConfig(cfg config.Client, version string) (*KibanaScopedClient, fwdiags.Diagnostics) {
	if cfg.Kibana == nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(
			"Missing Kibana config",
			"kibana_connection is required but the Kibana configuration was not set",
		)}
	}

	kibanaClient, err := buildKibanaClient(cfg)
	if err != nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic("Failed to build Kibana client", err.Error())}
	}

	kibanaOapiClient, err := buildKibanaOapiClient(cfg)
	if err != nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic("Failed to build Kibana OpenAPI client", err.Error())}
	}

	fleetClient, err := buildFleetClient(cfg)
	if err != nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic("Failed to build Fleet client", err.Error())}
	}

	var sloAPI slo.SloAPI
	if kibanaClient != nil {
		kibanaHTTPClient := kibanaClient.Client.GetClient()
		sloAPI = buildSloClient(cfg, kibanaHTTPClient).SloAPI
	}

	return &KibanaScopedClient{
		kibana:       kibanaClient,
		kibanaOapi:   kibanaOapiClient,
		sloAPI:       sloAPI,
		kibanaConfig: *cfg.Kibana,
		fleet:        fleetClient,
		version:      version,
	}, nil
}

// ConvertProviderDataToFactory converts the providerData value injected by
// Framework into a *ProviderClientFactory. It returns an error diagnostic when
// providerData is set but is not the expected type.
func ConvertProviderDataToFactory(providerData any) (*ProviderClientFactory, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if providerData == nil {
		return nil, diags
	}

	factory, ok := providerData.(*ProviderClientFactory)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ProviderClientFactory, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, diags
	}
	if factory == nil {
		diags.AddError(
			"Unconfigured Client Factory",
			"Expected configured client factory. Please report this issue to the provider developers.",
		)
	}
	return factory, diags
}

// NewKibanaScopedClientFromFactory returns a *KibanaScopedClient built from the
// factory's provider-level defaults. This is the typed Kibana surface
// equivalent of calling GetKibanaClient with an empty connection list.
func NewKibanaScopedClientFromFactory(f *ProviderClientFactory) *KibanaScopedClient {
	if f == nil || f.defaultClient == nil {
		return nil
	}
	return kibanaScopedClientFromAPIClient(f.defaultClient)
}

// ConvertMetaToFactory converts the SDK meta value into a *ProviderClientFactory.
func ConvertMetaToFactory(meta any) (*ProviderClientFactory, diag.Diagnostics) {
	if meta == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unconfigured Client Factory",
			Detail:   "Expected configured provider client factory, got nil. Report this issue to the provider developers.",
		}}
	}

	factory, ok := meta.(*ProviderClientFactory)
	if !ok {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unexpected meta type",
			Detail:   fmt.Sprintf("Expected *ProviderClientFactory, got: %T. Please report this issue to the provider developers.", meta),
		}}
	}
	if factory == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unconfigured Client Factory",
			Detail:   "Expected configured provider client factory, got nil. Report this issue to the provider developers.",
		}}
	}
	return factory, nil
}
