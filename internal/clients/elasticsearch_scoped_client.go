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
	"fmt"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/info"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ElasticsearchScopedClient is a typed client surface for Elasticsearch
// operations. It exposes the underlying go-elasticsearch client plus all
// Elasticsearch-derived helper behavior that resources need: composite ID
// generation, cluster identity lookup, version checks, flavor checks, and
// minimum-version enforcement.
//
// It deliberately does not expose Kibana or Fleet state so that all version and
// identity checks always resolve against the scoped Elasticsearch connection.
type ElasticsearchScopedClient struct {
	elasticsearch            *elasticsearch.Client
	elasticsearchClusterInfo *info.Response
	mu                       sync.Mutex
	// esEndpoints holds the resolved Elasticsearch endpoint addresses captured
	// after provider configuration, entity-local overrides, and environment
	// overrides have been applied. It is used by accessor validation to
	// distinguish missing endpoint configuration from unexpected nil states.
	esEndpoints []string
	// typedClient is the lazily-initialized strongly-typed Elasticsearch client.
	typedClient *elasticsearch.TypedClient
	// typedClientOnce ensures typedClient is created exactly once.
	typedClientOnce sync.Once
}

// GetESTypedClient returns the strongly-typed Elasticsearch client.
//
// The typed client is lazily initialized on the first call by converting the
// underlying *elasticsearch.Client via ToTyped(). The result is cached so that
// subsequent calls return the same *elasticsearch.TypedClient without repeated
// conversion. Initialization is safe for concurrent use by multiple goroutines.
//
// The returned typed client shares the same underlying transport, endpoints,
// and configuration as the untyped client returned by GetESClient(). A product
// check may run on the typed client's first request, adding marginal latency
// on first use.
func (e *ElasticsearchScopedClient) GetESTypedClient() (*elasticsearch.TypedClient, error) {
	esClient, err := e.GetESClient()
	if err != nil {
		return nil, err
	}
	e.typedClientOnce.Do(func() {
		e.typedClient = esClient.ToTyped()
	})
	return e.typedClient, nil
}

// GetESClient returns the underlying go-elasticsearch client. It satisfies the
// ESClient sink interface used by internal/clients/elasticsearch/ helpers.
func (e *ElasticsearchScopedClient) GetESClient() (*elasticsearch.Client, error) {
	hasEndpoint := false
	for _, ep := range e.esEndpoints {
		if ep != "" {
			hasEndpoint = true
			break
		}
	}
	if !hasEndpoint {
		return nil, errors.New("elasticsearch client is not configured: set elasticsearch.endpoints, elasticsearch_connection.endpoints, or ELASTICSEARCH_ENDPOINTS")
	}
	if e.elasticsearch == nil {
		return nil, errors.New("elasticsearch client not found")
	}
	return e.elasticsearch, nil
}

// serverInfo fetches and caches the Elasticsearch cluster info.
// It is safe for concurrent use: the mutex ensures only one goroutine fetches
// the info from the server, and subsequent callers use the cached result.
func (e *ElasticsearchScopedClient) serverInfo(ctx context.Context) (*info.Response, diag.Diagnostics) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.elasticsearchClusterInfo != nil {
		return e.elasticsearchClusterInfo, nil
	}

	typedClient, err := e.GetESTypedClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	res, err := typedClient.Core.Info().Do(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	// cache info
	e.elasticsearchClusterInfo = res

	return res, nil
}

// ClusterID returns the UUID of the connected Elasticsearch cluster. It is
// cached after the first call.
func (e *ElasticsearchScopedClient) ClusterID(ctx context.Context) (*string, diag.Diagnostics) {
	info, diags := e.serverInfo(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if uuid := info.ClusterUuid; uuid != "" && uuid != "_na_" {
		tflog.Trace(ctx, fmt.Sprintf("cluster UUID: %s", uuid))
		return &uuid, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to get cluster UUID",
		Detail: `Unable to get cluster UUID.
		There might be a problem with permissions or cluster is still starting up and UUID has not been populated yet.`,
	})
	return nil, diags
}

// ID returns a CompositeID combining the cluster UUID and the given resourceID.
func (e *ElasticsearchScopedClient) ID(ctx context.Context, resourceID string) (*CompositeID, diag.Diagnostics) {
	clusterID, diags := e.ClusterID(ctx)
	if diags.HasError() {
		return nil, diags
	}
	return &CompositeID{*clusterID, resourceID}, diags
}

// ServerVersion returns the version of the Elasticsearch server, derived from
// the cluster Info API.
func (e *ElasticsearchScopedClient) ServerVersion(ctx context.Context) (*version.Version, diag.Diagnostics) {
	info, diags := e.serverInfo(ctx)
	if diags.HasError() {
		return nil, diags
	}

	rawVersion := info.Version.Int
	serverVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return serverVersion, nil
}

// ServerFlavor returns the build flavor (e.g. "serverless", "default") of the
// Elasticsearch server, derived from the cluster Info API.
func (e *ElasticsearchScopedClient) ServerFlavor(ctx context.Context) (string, diag.Diagnostics) {
	info, diags := e.serverInfo(ctx)
	if diags.HasError() {
		return "", diags
	}
	return info.Version.BuildFlavor, nil
}

// EnforceMinVersion returns true when the server version is greater than or
// equal to minVersion, or when the server is running in serverless mode.
// If minVersion is nil, no minimum is enforced and the method returns true.
func (e *ElasticsearchScopedClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	if minVersion == nil {
		return true, nil
	}

	flavor, diags := e.ServerFlavor(ctx)
	if diags.HasError() {
		return false, diags
	}

	if flavor == ServerlessFlavor {
		return true, nil
	}

	serverVersion, diags := e.ServerVersion(ctx)
	if diags.HasError() {
		return false, diags
	}

	return serverVersion.GreaterThanOrEqual(minVersion), nil
}

// elasticsearchScopedClientFromAPIClient constructs an ElasticsearchScopedClient
// from the Elasticsearch-related fields of an *apiClient. This is the canonical
// adapter used by the factory and by NewAcceptanceTestingElasticsearchScopedClient.
func elasticsearchScopedClientFromAPIClient(a *apiClient) *ElasticsearchScopedClient {
	return &ElasticsearchScopedClient{
		elasticsearch:            a.elasticsearch,
		elasticsearchClusterInfo: a.elasticsearchClusterInfo,
		esEndpoints:              a.esEndpoints,
	}
}

// NewAcceptanceTestingElasticsearchScopedClient builds an
// ElasticsearchScopedClient for acceptance tests by reusing the acceptance
// testing APIClient.
func NewAcceptanceTestingElasticsearchScopedClient() (*ElasticsearchScopedClient, error) {
	apiClient, err := newAcceptanceTestingClient()
	if err != nil {
		return nil, err
	}
	return elasticsearchScopedClientFromAPIClient(apiClient), nil
}
