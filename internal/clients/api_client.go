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
	"strings"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/info"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

type CompositeID struct {
	ClusterID  string
	ResourceID string
}

const ServerlessFlavor = "serverless"

// CompositeIDFromStr parses an ID as <cluster_uuid>/<resource_identifier>. Only the first "/"
// separates cluster from resource, so resource_identifier may contain further slashes (for example
// ML calendar events "<calendar_id>/<event_id>" after the cluster segment).
//
// For backward compatibility, an ID with an empty cluster segment and a non-empty resource
// segment (for example "/<synthetics_monitor_id>" from legacy [CompositeID.String] formatting) is
// accepted; an empty resource segment (including a trailing slash after the cluster) is rejected.
func CompositeIDFromStr(id string) (*CompositeID, fwdiags.Diagnostics) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Wrong resource ID.",
				"Resource ID must have following format: <cluster_uuid>/<resource identifier>",
			),
		}
	}
	if parts[0] == "" {
		return &CompositeID{
			ClusterID:  "",
			ResourceID: parts[1],
		}, nil
	}
	return &CompositeID{
		ClusterID:  parts[0],
		ResourceID: parts[1],
	}, nil
}

func (c *CompositeID) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterID, c.ResourceID)
}

// apiClient is the internal broad client that holds all configured service
// clients built from the provider configuration block. It is unexported;
// external code must use scoped clients (KibanaScopedClient,
// ElasticsearchScopedClient) obtained through ProviderClientFactory or the
// acceptance-testing helper constructors.
type apiClient struct {
	elasticsearch            *elasticsearch.TypedClient
	elasticsearchClusterInfo *info.Response
	kibanaOapi               *kibanaoapi.Client
	fleet                    *fleet.Client
	version                  string
	// esEndpoints holds the resolved Elasticsearch endpoint addresses from
	// provider configuration plus environment overrides. Entity-local overrides
	// are applied later in ProviderClientFactory and stored on scoped clients.
	// Carried through to ElasticsearchScopedClient for factory endpoint validation.
	esEndpoints []string
	// kibanaEndpoint holds the resolved Kibana endpoint URL from provider
	// configuration plus environment overrides. Entity-local overrides are
	// applied later in ProviderClientFactory and stored on scoped clients.
	// Carried through to KibanaScopedClient for factory endpoint validation.
	kibanaEndpoint string
	// fleetEndpoint holds the resolved Fleet endpoint URL from provider
	// configuration plus environment overrides, including any inheritance from
	// the Kibana-derived config path. Entity-local overrides are applied later
	// in ProviderClientFactory and stored on scoped clients. Carried through to
	// KibanaScopedClient for factory endpoint validation.
	fleetEndpoint string
}

func newAcceptanceTestingClient() (*apiClient, error) {
	version := "tf-acceptance-testing"
	cfg := config.NewFromEnv(version)
	return newAPIClientFromConfig(cfg, version)
}

func newAPIClientFromFramework(ctx context.Context, cfg config.ProviderConfiguration, version string) (*apiClient, fwdiags.Diagnostics) {
	clientCfg, diags := config.NewFromFramework(ctx, cfg, version)
	if diags.HasError() {
		return nil, diags
	}

	client, err := newAPIClientFromConfig(clientCfg, version)
	if err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic("Failed to create API client", err.Error()),
		}
	}

	return client, nil
}

type MinVersionEnforceable interface {
	EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, fwdiags.Diagnostics)
}

func buildEsClient(cfg config.Client) (*elasticsearch.TypedClient, error) {
	if cfg.Elasticsearch == nil {
		return nil, nil
	}

	es, err := elasticsearch.NewTypedClient(*cfg.Elasticsearch)
	if err != nil {
		return nil, fmt.Errorf("unable to create Elasticsearch client: %w", err)
	}

	return es, nil
}

func buildKibanaOapiClient(cfg config.Client) (*kibanaoapi.Client, error) {
	client, err := kibanaoapi.NewClient(*cfg.KibanaOapi)
	if err != nil {
		return nil, fmt.Errorf("unable to create KibanaOapi client: %w", err)
	}

	return client, nil
}

func buildFleetClient(cfg config.Client) (*fleet.Client, error) {
	client, err := fleet.NewClient(*cfg.Fleet)
	if err != nil {
		return nil, fmt.Errorf("unable to create Fleet client: %w", err)
	}

	return client, nil
}

func newAPIClientFromConfig(cfg config.Client, version string) (*apiClient, error) {
	client := &apiClient{
		version: version,
	}

	if cfg.Elasticsearch != nil {
		esClient, err := buildEsClient(cfg)
		if err != nil {
			return nil, err
		}
		client.elasticsearch = esClient
		client.esEndpoints = cfg.Elasticsearch.Addresses
	}

	if cfg.KibanaOapi != nil {
		kibanaOapiClient, err := buildKibanaOapiClient(cfg)
		if err != nil {
			return nil, err
		}
		client.kibanaOapi = kibanaOapiClient

		if cfg.KibanaOapi != nil {
			client.kibanaEndpoint = cfg.KibanaOapi.URL
		}
	}

	if cfg.Fleet != nil {
		fleetClient, err := buildFleetClient(cfg)
		if err != nil {
			return nil, err
		}

		client.fleet = fleetClient
		client.fleetEndpoint = cfg.Fleet.URL
	}

	return client, nil
}
