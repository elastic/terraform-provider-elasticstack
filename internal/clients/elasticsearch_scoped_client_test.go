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
	"os"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockElasticsearchServer returns an httptest.Server that responds to GET /
// with a minimal Elasticsearch info payload for the given version using the
// "default" build flavor. It sets the X-Elastic-Product header that the
// go-elasticsearch client requires for product-check validation.
func newMockElasticsearchServer(version string) *httptest.Server {
	return newMockElasticsearchServerWithFlavor(version, "default")
}

// newMockElasticsearchServerWithFlavor returns an httptest.Server that responds to GET /
// with a minimal Elasticsearch info payload for the given version and build flavor.
func newMockElasticsearchServerWithFlavor(version, flavor string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			payload := map[string]any{
				"cluster_uuid": "test-cluster-uuid",
				"version": map[string]any{
					"number":       version,
					"build_flavor": flavor,
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
		"ca_fingerprint":           types.StringType,
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

	// Prevent ELASTICSEARCH_* env vars from overriding elasticsearch_connection endpoints.
	for _, key := range []string{
		"ELASTICSEARCH_ENDPOINTS",
		"ELASTICSEARCH_INSECURE",
		"ELASTICSEARCH_BEARER_TOKEN",
		"ELASTICSEARCH_ES_CLIENT_AUTHENTICATION",
		"ELASTICSEARCH_CA_FINGERPRINT",
	} {
		if val, ok := os.LookupEnv(key); ok {
			t.Setenv(key, val)
		}
		os.Unsetenv(key)
	}

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

func TestElasticsearchScopedClient_GetESClient_ReturnsTypedClient(t *testing.T) {
	t.Parallel()
	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch.example.com:9200"},
	})
	require.NoError(t, err)

	sc := &ElasticsearchScopedClient{
		typedClient: esClient,
		esEndpoints: []string{"http://elasticsearch.example.com:9200"},
	}
	require.Same(t, esClient, sc.GetESClient())
}

func TestElasticsearchScopedClient_GetESClient_NilWhenUnconfigured(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{esEndpoints: []string{}}
	assert.Nil(t, sc.GetESClient())
}

func TestElasticsearchScopedClient_GetESClient_Present(t *testing.T) {
	t.Parallel()
	factory := newTestFactory(t)
	// Build a scoped client from provider defaults and verify the factory path
	// succeeds without panic.
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

	esClient := scoped.GetESClient()
	require.NotNil(t, esClient)
}

// --- ElasticsearchScopedClient version / flavor routing ---

func TestElasticsearchScopedClient_ServerlessEnforceMinVersion(t *testing.T) {
	srv := newMockElasticsearchServerWithFlavor("8.19.0", ServerlessFlavor)
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

// --- IsServerless ---

func TestElasticsearchScopedClient_IsServerless_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{esEndpoints: []string{}}
	isServerless, diags := sc.IsServerless(context.Background())
	assert.False(t, isServerless)
	require.True(t, diags.HasError())
}

func TestElasticsearchScopedClient_IsServerless_InfoAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	isServerless, diags := scoped.IsServerless(context.Background())
	assert.False(t, isServerless)
	require.True(t, diags.HasError())
}

func TestElasticsearchScopedClient_IsServerless_Serverless(t *testing.T) {
	srv := newMockElasticsearchServerWithFlavor("8.19.0", ServerlessFlavor)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	isServerless, diags := scoped.IsServerless(context.Background())
	require.False(t, diags.HasError())
	assert.True(t, isServerless)
}

func TestElasticsearchScopedClient_IsServerless_Stateful(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	isServerless, diags := scoped.IsServerless(context.Background())
	require.False(t, diags.HasError())
	assert.False(t, isServerless)
}

func TestElasticsearchScopedClient_IsServerless_EmptyFlavor(t *testing.T) {
	srv := newMockElasticsearchServerWithFlavor("8.19.0", "")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	isServerless, diags := scoped.IsServerless(context.Background())
	require.False(t, diags.HasError())
	assert.False(t, isServerless, "empty build_flavor must not be treated as serverless")
}

// --- AcceptanceServerInfo ---

func TestAcceptanceServerInfo_Stateful(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockElasticsearchServer(wantVersion)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ver, isServerless, diags := AcceptanceServerInfo(context.Background(), scoped)
	require.False(t, diags.HasError())
	require.NotNil(t, ver)
	assert.Equal(t, wantVersion, ver.Original())
	assert.False(t, isServerless)
}

func TestAcceptanceServerInfo_Serverless(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockElasticsearchServerWithFlavor(wantVersion, ServerlessFlavor)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ver, isServerless, diags := AcceptanceServerInfo(context.Background(), scoped)
	require.False(t, diags.HasError())
	require.NotNil(t, ver)
	assert.Equal(t, wantVersion, ver.Original())
	assert.True(t, isServerless)
}

func TestAcceptanceServerInfo_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{esEndpoints: []string{}}

	ver, isServerless, diags := AcceptanceServerInfo(context.Background(), sc)
	assert.Nil(t, ver)
	assert.False(t, isServerless)
	require.True(t, diags.HasError())
}

func TestAcceptanceServerInfo_InfoAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ver, isServerless, diags := AcceptanceServerInfo(context.Background(), scoped)
	assert.Nil(t, ver)
	assert.False(t, isServerless)
	require.True(t, diags.HasError())
}

// --- EnforceVersionCheck ---

func TestElasticsearchScopedClient_EnforceVersionCheck_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{esEndpoints: []string{}}
	ok, diags := sc.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestElasticsearchScopedClient_EnforceVersionCheck_InfoAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestElasticsearchScopedClient_EnforceVersionCheck_MalformedVersionResponse(t *testing.T) {
	srv := newMockElasticsearchServerWithFlavor("not-a-version", "default")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestElasticsearchScopedClient_EnforceVersionCheck_StatefulBelowMin(t *testing.T) {
	srv := newMockElasticsearchServer("8.10.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)
	minVer, err := goversion.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(v *goversion.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestElasticsearchScopedClient_EnforceVersionCheck_StatefulAtMin(t *testing.T) {
	srv := newMockElasticsearchServer("8.15.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)
	minVer, err := goversion.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(v *goversion.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestElasticsearchScopedClient_EnforceVersionCheck_StatefulAboveMin(t *testing.T) {
	srv := newMockElasticsearchServer("9.0.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)
	minVer, err := goversion.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(v *goversion.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestElasticsearchScopedClient_EnforceVersionCheck_ServerlessShortCircuit(t *testing.T) {
	srv := newMockElasticsearchServerWithFlavor("8.10.0", ServerlessFlavor)
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool { return false })
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must short-circuit to true even when check returns false")
}

func TestElasticsearchScopedClient_EnforceVersionCheck_PredicateFalseOnStateful(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool {
		return false
	})
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestElasticsearchScopedClient_EnforceVersionCheck_PredicateTrueOnStateful(t *testing.T) {
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	scoped := newMockScopedClient(t, srv)

	ok, diags := scoped.EnforceVersionCheck(context.Background(), func(_ *goversion.Version) bool {
		return true
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
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

	_, diags := scoped.IsServerless(context.Background())
	require.False(t, diags.HasError())

	_, diags = scoped.IsServerless(context.Background())
	require.False(t, diags.HasError())

	assert.Equal(t, 1, callCount, "serverInfo must only call the Elasticsearch Info API once (result must be cached)")
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

// --- GetESClient typed-client tests ---

func TestGetESClient_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{Addresses: []string{"http://localhost:9200"}})
	require.NoError(t, err)

	scoped := &ElasticsearchScopedClient{
		typedClient: esClient,
		esEndpoints: []string{"http://localhost:9200"},
	}

	require.NotNil(t, scoped.GetESClient())
}
