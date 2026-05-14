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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockElasticsearchServer returns an httptest.Server that responds to GET /
// with a minimal Elasticsearch info payload for the given version using the
// "default" build flavor. It sets the X-Elastic-Product header that the
// go-elasticsearch client requires for product-check validation.
func newMockElasticsearchServer(version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			payload := map[string]any{
				"cluster_uuid": "test-cluster-uuid",
				"version": map[string]any{
					"number":       version,
					"build_flavor": "default",
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// elasticsearchConnectionAttrTypes returns the attribute type map for
// config.ElasticsearchConnection so we can build framework type values in tests.
func elasticsearchConnectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username":                 types.StringType,
		"password":                 types.StringType,
		"api_key":                  types.StringType,
		"bearer_token":             types.StringType,
		"es_client_authentication": types.StringType,
		"endpoints":                types.ListType{ElemType: types.StringType},
		"headers":                  types.MapType{ElemType: types.StringType},
		"insecure":                 types.BoolType,
		"ca_file":                  types.StringType,
		"ca_data":                  types.StringType,
		"cert_file":                types.StringType,
		"key_file":                 types.StringType,
		"cert_data":                types.StringType,
		"key_data":                 types.StringType,
	}
}

// newScopedElasticsearchClientFromFactory creates an *ElasticsearchScopedClient
// via the factory pointing at the given endpoint.
func newScopedElasticsearchClientFromFactory(t *testing.T, endpoint string) *ElasticsearchScopedClient {
	t.Helper()

	ctx := context.Background()
	factory := newTestFactory(t)
	conn := config.ElasticsearchConnection{
		Username:               types.StringValue("elastic"),
		Password:               types.StringValue("changeme"),
		APIKey:                 types.StringValue(""),
		BearerToken:            types.StringValue(""),
		ESClientAuthentication: types.StringValue(""),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue(endpoint),
		}),
		Headers:  types.MapValueMust(types.StringType, map[string]attr.Value{}),
		Insecure: types.BoolValue(true),
		CAFile:   types.StringValue(""),
		CAData:   types.StringValue(""),
		CertFile: types.StringValue(""),
		KeyFile:  types.StringValue(""),
		CertData: types.StringValue(""),
		KeyData:  types.StringValue(""),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, list)
	require.False(t, diags.HasError())
	return scoped
}

// newMockScopedClient creates an ElasticsearchScopedClient whose typed
// elasticsearch transport points at the given HTTP server. It bypasses the
// provider factory to avoid environment-variable endpoint overrides in
// acceptance test environments.
func newMockScopedClient(t *testing.T, srv *httptest.Server) *ElasticsearchScopedClient {
	t.Helper()
	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
	})
	require.NoError(t, err)
	return &ElasticsearchScopedClient{typedClient: esClient, esEndpoints: []string{srv.URL}}
}

// --- GetESClient ---

func TestElasticsearchScopedClient_GetESClient_Nil(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{}
	_, err := sc.GetESClient()
	assert.Error(t, err, "GetESClient must return an error when typedClient is nil")
}

// --- Scenario 1: Missing ES endpoint returns actionable error ---

func TestElasticsearchScopedClient_GetESClient_MissingEndpoint(t *testing.T) {
	t.Parallel()
	// esEndpoints is empty: models a scoped client built with no endpoint configuration.
	sc := &ElasticsearchScopedClient{esEndpoints: []string{}}
	client, err := sc.GetESClient()
	assert.Nil(t, client, "GetESClient must return nil client when esEndpoints is empty")
	require.Error(t, err)
	assert.Equal(t, elasticsearchClientNotConfiguredError, err.Error())
}

func TestElasticsearchScopedClient_GetESClient_OnlyEmptyEndpoints(t *testing.T) {
	t.Parallel()
	// esEndpoints contains only empty strings: must be treated as unconfigured.
	sc := &ElasticsearchScopedClient{esEndpoints: []string{"", ""}}
	client, err := sc.GetESClient()
	assert.Nil(t, client, "GetESClient must return nil client when all esEndpoints are empty strings")
	require.Error(t, err)
	assert.Equal(t, elasticsearchClientNotConfiguredError, err.Error())
}

// --- Scenario 7 (ES): Endpoint present, auth empty → accessor succeeds ---

func TestElasticsearchScopedClient_GetESClient_EndpointPresentNoAuth(t *testing.T) {
	t.Parallel()
	// Build a real ES client pointing at a dummy address but with no credentials.
	// The accessor validates endpoint presence only, so it must not reject this.
	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch.example.com:9200"},
	})
	require.NoError(t, err)

	sc := &ElasticsearchScopedClient{
		typedClient: esClient,
		esEndpoints: []string{"http://elasticsearch.example.com:9200"},
	}
	client, err := sc.GetESClient()
	require.NoError(t, err,
		"GetESClient must not fail when endpoint is present but auth fields are empty")
	assert.NotNil(t, client)
}

func TestElasticsearchScopedClient_GetESClient_Present(t *testing.T) {
	t.Parallel()
	factory := newTestFactory(t)
	// Build a scoped client from provider defaults (newTestAPIClient has no ES client set,
	// but the factory wraps the APIClient fields). Here we verify the method exists and
	// the scoped client built from a real factory does not panic.
	ctx := context.Background()
	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, emptyList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
}

// --- GetElasticsearchClient (Framework) ---

func TestGetElasticsearchClient_EmptyList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, emptyList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
}

func TestGetElasticsearchClient_NullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	nullList := types.ListNull(types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()})

	scoped, diags := factory.GetElasticsearchClient(ctx, nullList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
}

func TestGetElasticsearchClient_WithConnection(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	// Use the factory path to verify wiring; do not assert specific server
	// responses because the factory may pick up ELASTICSEARCH_ENDPOINTS from
	// the environment in CI.
	scoped := newScopedElasticsearchClientFromFactory(t, srv.URL)
	require.NotNil(t, scoped)

	esClient, err := scoped.GetESClient()
	require.NoError(t, err, "GetESClient must return a valid client when connection is configured")
	require.NotNil(t, esClient)
}

// --- GetElasticsearchClientFromSDK ---

func TestGetElasticsearchClientFromSDK_AbsentBlock(t *testing.T) {
	t.Parallel()
	factory := newTestFactory(t)

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"elasticsearch_connection": providerschema.GetEsConnectionSchema("elasticsearch_connection", false),
	}, map[string]any{})

	scoped, diags := factory.GetElasticsearchClientFromSDK(rd)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
}

func TestGetElasticsearchClientFromSDK_WithBlock(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	factory := newTestFactory(t)

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"elasticsearch_connection": providerschema.GetEsConnectionSchema("elasticsearch_connection", false),
	}, map[string]any{
		"elasticsearch_connection": []any{
			map[string]any{
				"username":                 "elastic",
				"password":                 "changeme",
				"api_key":                  "",
				"bearer_token":             "",
				"es_client_authentication": "",
				"endpoints":                []any{srv.URL},
				"insecure":                 true,
				"ca_file":                  "",
				"ca_data":                  "",
				"cert_file":                "",
				"key_file":                 "",
				"cert_data":                "",
				"key_data":                 "",
			},
		},
	})

	scoped, diags := factory.GetElasticsearchClientFromSDK(rd)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)

	esClient, err := scoped.GetESClient()
	require.NoError(t, err, "GetESClient must return a valid client")
	require.NotNil(t, esClient)
}

// --- ElasticsearchScopedClient version / flavor routing ---

func TestElasticsearchScopedClient_ServerVersion(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockElasticsearchServer(wantVersion)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ver, diags := scoped.ServerVersion(context.Background())
	require.False(t, diags.HasError())
	require.NotNil(t, ver)
	assert.Equal(t, wantVersion, ver.Original())
}

func TestElasticsearchScopedClient_ServerFlavor(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockElasticsearchServer(wantVersion)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	flavor, diags := scoped.ServerFlavor(context.Background())
	require.False(t, diags.HasError())
	assert.Equal(t, "default", flavor)
}

func TestElasticsearchScopedClient_ServerlessEnforceMinVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"serverless-uuid","version":{"number":"8.19.0","build_flavor":"serverless"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)
	require.NotNil(t, scoped)

	// Any version gate must pass for serverless.
	ver, _ := goversion.NewVersion("99.0.0")
	ok, diags := scoped.EnforceMinVersion(context.Background(), ver)
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must always satisfy any version gate")
}

func TestElasticsearchScopedClient_EnforceMinVersion_Satisfied(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	minVer, err := goversion.NewVersion("8.0.0")
	require.NoError(t, err)

	ok, diags := scoped.EnforceMinVersion(context.Background(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok, "8.19.0 must satisfy min version 8.0.0")
}

func TestElasticsearchScopedClient_EnforceMinVersion_NotSatisfied(t *testing.T) {
	srv := newMockElasticsearchServer("7.17.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	minVer, err := goversion.NewVersion("8.0.0")
	require.NoError(t, err)

	ok, diags := scoped.EnforceMinVersion(context.Background(), minVer)
	require.False(t, diags.HasError())
	assert.False(t, ok, "7.17.0 must not satisfy min version 8.0.0")
}

// --- ClusterID ---

func TestElasticsearchScopedClient_ClusterID_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"abc-123","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	id, diags := scoped.ClusterID(context.Background())
	require.False(t, diags.HasError())
	require.NotNil(t, id)
	assert.Equal(t, "abc-123", *id)
}

func TestElasticsearchScopedClient_ClusterID_NA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"_na_","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	id, diags := scoped.ClusterID(context.Background())
	assert.True(t, diags.HasError(), "ClusterID must return an error when cluster_uuid is '_na_'")
	assert.Nil(t, id)
}

func TestElasticsearchScopedClient_ClusterID_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	id, diags := scoped.ClusterID(context.Background())
	assert.True(t, diags.HasError(), "ClusterID must return an error when cluster_uuid is empty")
	assert.Nil(t, id)
}

// --- ID ---

func TestElasticsearchScopedClient_ID_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"abc-123","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	composite, diags := scoped.ID(context.Background(), "my-resource")
	require.False(t, diags.HasError())
	require.NotNil(t, composite)
	assert.Equal(t, "abc-123", composite.ClusterID)
	assert.Equal(t, "my-resource", composite.ResourceID)
}

// --- serverInfo cache ---

func TestElasticsearchScopedClient_ServerInfo_IsCached(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"abc-123","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	_, diags := scoped.ServerVersion(context.Background())
	require.False(t, diags.HasError())

	_, diags = scoped.ServerVersion(context.Background())
	require.False(t, diags.HasError())

	assert.Equal(t, 1, callCount, "serverInfo must only call the Elasticsearch Info API once (result must be cached)")
}

// --- ServerVersion error paths ---

func TestElasticsearchScopedClient_ServerVersion_InvalidVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprintf(w, `{"cluster_uuid":"abc-123","version":{"number":"not-a-version","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	_, diags := scoped.ServerVersion(context.Background())
	assert.True(t, diags.HasError(), "ServerVersion must return an error for an unparseable version string")
}

// --- Nil-factory guard tests ---

func TestGetElasticsearchClient_NilFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var f *ProviderClientFactory

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{},
	)
	require.False(t, diags.HasError())

	_, diags = f.GetElasticsearchClient(ctx, emptyList)
	assert.True(t, diags.HasError(), "GetElasticsearchClient on a nil factory must return an error diagnostic")
}

func TestGetElasticsearchClientFromSDK_NilFactory(t *testing.T) {
	t.Parallel()
	var f *ProviderClientFactory

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"elasticsearch_connection": providerschema.GetEsConnectionSchema("elasticsearch_connection", false),
	}, map[string]any{})

	_, diags := f.GetElasticsearchClientFromSDK(rd)
	assert.True(t, diags.HasError(), "GetElasticsearchClientFromSDK on a nil factory must return an error diagnostic")
}

// --- GetESClient typed-client tests ---

func TestGetESClient_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{Addresses: []string{"http://localhost:9200"}})
	if err != nil {
		t.Fatalf("failed to create elasticsearch client: %v", err)
	}

	scoped := &ElasticsearchScopedClient{
		typedClient: esClient,
		esEndpoints: []string{"http://localhost:9200"},
	}

	typedClient, err := scoped.GetESClient()
	require.NoError(t, err, "GetESClient must not return an error when configured")
	require.NotNil(t, typedClient, "expected non-nil typed client, got nil")
}

func TestGetESClient_ReturnsErrorWhenUnconfigured(t *testing.T) {
	t.Parallel()

	scoped := &ElasticsearchScopedClient{
		esEndpoints: []string{},
	}

	typedClient, err := scoped.GetESClient()
	assert.Nil(t, typedClient, "expected nil typed client when unconfigured")
	assert.Error(t, err, "expected error when calling GetESClient on unconfigured client")
}
