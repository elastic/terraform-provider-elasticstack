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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestFactory constructs a *ProviderClientFactory backed by a minimal
// test APIClient for use in unit tests.
func newTestFactory(t *testing.T) *ProviderClientFactory {
	t.Helper()
	return NewProviderClientFactory(newTestAPIClient(t))
}

// --- ConvertProviderDataToFactory ---

func TestConvertProviderDataToFactory_Nil(t *testing.T) {
	t.Parallel()
	factory, diags := ConvertProviderDataToFactory(nil)
	require.False(t, diags.HasError())
	assert.Nil(t, factory)
}

func TestConvertProviderDataToFactory_WrongType(t *testing.T) {
	t.Parallel()
	_, diags := ConvertProviderDataToFactory("unexpected-string")
	assert.True(t, diags.HasError())
}

func TestConvertProviderDataToFactory_Valid(t *testing.T) {
	t.Parallel()
	f := newTestFactory(t)
	result, diags := ConvertProviderDataToFactory(f)
	require.False(t, diags.HasError())
	assert.Same(t, f, result)
}

// --- ProviderClientFactory.GetKibanaClient (Framework) ---

// TestGetKibanaClient_EmptyList verifies that when kibana_connection is absent
// the factory returns a typed client derived from provider-level defaults.
func TestGetKibanaClient_EmptyList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)

	// The scoped client must expose both Kibana OpenAPI and Fleet clients.
	require.NotNil(t, scoped.GetKibanaOapiClient())
	require.NotNil(t, scoped.GetFleetClient())
}

// TestGetKibanaClient_NullList verifies that a null kibana_connection is treated
// the same as an empty list (provider defaults are returned).
func TestGetKibanaClient_NullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	nullList := types.ListNull(types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()})

	scoped, diags := factory.GetKibanaClient(ctx, nullList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
}

// TestGetKibanaClient_WithConnection verifies that when a kibana_connection block
// is present the factory returns a new typed scoped client with rebuilt clients.
func TestGetKibanaClient_WithConnection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	conn := config.KibanaConnection{
		Username:    types.StringValue("kibana-user"),
		Password:    types.StringValue("kibana-pass"),
		APIKey:      types.StringValue(""),
		BearerToken: types.StringValue(""),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("http://kibana.example.com:5601"),
		}),
		CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
		Insecure: types.BoolValue(false),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, list)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)

	// The scoped client must expose Kibana-derived surfaces.
	require.NotNil(t, scoped.GetKibanaOapiClient())
	require.NotNil(t, scoped.GetFleetClient())
}

// --- KibanaScopedClient version / flavor routing ---

// newScopedClientFromFactory creates a *KibanaScopedClient via the factory
// pointing at the given endpoint. TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT
// is set so environment variables cannot override the URL.
func newScopedClientFromFactory(t *testing.T, endpoint string) *KibanaScopedClient {
	t.Helper()
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")

	ctx := context.Background()
	factory := newTestFactory(t)
	conn := config.KibanaConnection{
		Username:    types.StringValue("kibana-user"),
		Password:    types.StringValue("kibana-pass"),
		APIKey:      types.StringValue(""),
		BearerToken: types.StringValue(""),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue(endpoint),
		}),
		CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
		Insecure: types.BoolValue(false),
	}
	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, list)
	require.False(t, diags.HasError())
	return scoped
}

// TestKibanaScopedClient_EnforceMinVersion_ViaFactory_Stateful verifies that
// EnforceMinVersion on a scoped client obtained from the factory routes through
// the Kibana status API (not Elasticsearch) and evaluates stateful version gates.
func TestKibanaScopedClient_EnforceMinVersion_ViaFactory_Stateful(t *testing.T) {
	const serverVersion = "8.19.0"
	srv := newMockKibanaServer(serverVersion)
	defer srv.Close()

	scoped := newScopedClientFromFactory(t, srv.URL)

	minBelow, err := goversion.NewVersion("8.0.0")
	require.NoError(t, err)
	ok, diags := scoped.EnforceMinVersion(context.Background(), minBelow)
	require.False(t, diags.HasError())
	assert.True(t, ok, "8.19.0 must satisfy min 8.0.0")

	minAbove, err := goversion.NewVersion("9.0.0")
	require.NoError(t, err)
	ok, diags = scoped.EnforceMinVersion(context.Background(), minAbove)
	require.False(t, diags.HasError())
	assert.False(t, ok, "8.19.0 must not satisfy min 9.0.0")
}

// TestKibanaScopedClient_EnforceMinVersion_ViaFactory_ServerlessShortCircuit verifies that
// EnforceMinVersion always returns true for serverless Kibana obtained via the factory.
func TestKibanaScopedClient_EnforceMinVersion_ViaFactory_ServerlessShortCircuit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":"8.19.0","build_flavor":"serverless"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")
	scoped := newScopedClientFromFactory(t, srv.URL)

	require.NotNil(t, scoped)
	// Any version gate must pass for serverless.
	ver, _ := goversion.NewVersion("99.0.0")
	ok, diags := scoped.EnforceMinVersion(context.Background(), ver)
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must always satisfy any version gate")
}

// --- NewKibanaScopedClientFromFactory ---

func TestNewKibanaScopedClientFromFactory_NilFactory(t *testing.T) {
	t.Parallel()
	scoped, diags := NewKibanaScopedClientFromFactory(nil)
	assert.Nil(t, scoped)
	assert.False(t, diags.HasError())
}

func TestNewKibanaScopedClientFromFactory_Valid(t *testing.T) {
	t.Parallel()
	f := newTestFactory(t)
	scoped, diags := NewKibanaScopedClientFromFactory(f)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
	require.NotNil(t, scoped.GetKibanaOapiClient())
	require.NotNil(t, scoped.GetFleetClient())
}

func TestGetElasticsearchClient_ProviderDefaultEndpointsWithoutTypedClient(t *testing.T) {
	ctx := context.Background()
	factory := NewProviderClientFactory(&apiClient{
		version:     "unit-testing",
		esEndpoints: []string{"http://localhost:9200"},
	})

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, emptyList)
	require.True(t, diags.HasError(), "factory must fail when ES endpoint is set but typed client is nil")
	assert.Nil(t, scoped)
	assert.Equal(t, "Elasticsearch client not found", diags[0].Summary())
}

func TestGetKibanaClient_ProviderDefaultKibanaEndpointWithoutInnerClient(t *testing.T) {
	ctx := context.Background()
	factory := NewProviderClientFactory(&apiClient{
		version:        "unit-testing",
		kibanaEndpoint: "http://localhost:5601",
	})

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.True(t, diags.HasError(), "factory must fail when Kibana endpoint is set but OpenAPI client is nil")
	assert.Nil(t, scoped)
	assert.Equal(t, "kibanaoapi client not found", diags[0].Summary())
}

func TestGetKibanaClient_ProviderDefaultFleetEndpointWithoutInnerClient(t *testing.T) {
	ctx := context.Background()
	factory := NewProviderClientFactory(&apiClient{
		version:       "unit-testing",
		fleetEndpoint: "http://localhost:5601",
	})

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.True(t, diags.HasError(), "factory must fail when Fleet endpoint is set but Fleet client is nil")
	assert.Nil(t, scoped)
	assert.Equal(t, "Fleet client not found", diags[0].Summary())
}

func TestNewKibanaScopedClientFromFactory_MissingEndpoint(t *testing.T) {
	t.Parallel()
	factory := NewProviderClientFactory(&apiClient{version: "unit-testing"})
	scoped, diags := NewKibanaScopedClientFromFactory(factory)
	require.True(t, diags.HasError(), "factory helper must fail when neither Kibana nor Fleet endpoint is configured")
	assert.Nil(t, scoped)
	assert.Equal(t, kibanaFleetClientNotConfiguredError, diags[0].Detail())
}

// --- Resource-level path regression tests (Tasks 2.3 / 2.4) ---

func TestGetElasticsearchClient_MultipleConnectionBlocks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	conn := config.ElasticsearchConnection{
		Username: types.StringValue("elastic"),
		Password: types.StringValue("changeme"),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("http://elasticsearch.example.com:9200"),
		}),
		Headers:  types.MapValueMust(types.StringType, map[string]attr.Value{}),
		Insecure: types.BoolValue(false),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{conn, conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, list)
	require.True(t, diags.HasError(), "factory must reject multiple elasticsearch_connection blocks")
	assert.Nil(t, scoped)
	assert.Equal(t, "Invalid elasticsearch_connection", diags[0].Summary())
	assert.Equal(t, "At most one elasticsearch_connection block is allowed.", diags[0].Detail())
}

func TestGetKibanaClient_NilFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var f *ProviderClientFactory

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	_, diags = f.GetKibanaClient(ctx, emptyList)
	assert.True(t, diags.HasError(), "GetKibanaClient on a nil factory must return an error diagnostic")
}

func TestGetKibanaClient_ResourceLevelBuildError(t *testing.T) {
	ctx := context.Background()
	factory := newTestFactory(t)

	conn := config.KibanaConnection{
		Username: types.StringValue("elastic"),
		Password: types.StringValue("changeme"),
		APIKey:   types.StringValue(""),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("http://kibana.example.com:5601"),
		}),
		CACerts: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("/nonexistent/ca.pem"),
		}),
		Insecure: types.BoolValue(false),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, list)
	require.True(t, diags.HasError(), "factory must surface kibana_connection build errors before endpoint validation")
	assert.Nil(t, scoped)
	assert.Equal(t, "Failed to build Kibana OpenAPI client", diags[0].Summary())
	assert.NotEqual(t, kibanaFleetClientNotConfiguredError, diags[0].Detail())
}

// --- Entity-local elasticsearch_connection with missing endpoint ---
// GetElasticsearchClient must fail at factory resolution when the connection
// block has no endpoints, matching the provider-default missing-endpoint path.

func TestGetElasticsearchClient_EntityLocalMissingEndpoint(t *testing.T) {
	// Prevent ELASTICSEARCH_ENDPOINTS env var from supplying a fallback endpoint.
	t.Setenv("ELASTICSEARCH_ENDPOINTS", "")

	ctx := context.Background()
	factory := newTestFactory(t)

	conn := config.ElasticsearchConnection{
		Username:               types.StringValue("elastic"),
		Password:               types.StringValue("changeme"),
		APIKey:                 types.StringValue(""),
		BearerToken:            types.StringValue(""),
		ESClientAuthentication: types.StringValue(""),
		// Endpoints intentionally empty.
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{}),
		Headers:   types.MapValueMust(types.StringType, map[string]attr.Value{}),
		Insecure:  types.BoolValue(false),
		CAFile:    types.StringValue(""),
		CAData:    types.StringValue(""),
		CertFile:  types.StringValue(""),
		KeyFile:   types.StringValue(""),
		CertData:  types.StringValue(""),
		KeyData:   types.StringValue(""),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, list)
	require.True(t, diags.HasError(), "factory must fail when elasticsearch_connection has no endpoints")
	assert.Nil(t, scoped)
	assert.Equal(t, elasticsearchClientNotConfiguredError, diags[0].Detail())
}

// --- Entity-local kibana_connection with missing endpoint ---
// GetKibanaClient must fail at factory resolution when the connection block
// has no endpoints, matching the provider-default missing-endpoint path.

func TestGetKibanaClient_EntityLocalMissingEndpoint(t *testing.T) {
	// Prevent KIBANA_ENDPOINT env var from supplying a fallback endpoint.
	t.Setenv("KIBANA_ENDPOINT", "")

	ctx := context.Background()
	factory := newTestFactory(t)

	conn := config.KibanaConnection{
		Username:    types.StringValue("elastic"),
		Password:    types.StringValue("changeme"),
		APIKey:      types.StringValue(""),
		BearerToken: types.StringValue(""),
		// Endpoints intentionally empty.
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{}),
		CACerts:   types.ListValueMust(types.StringType, []attr.Value{}),
		Insecure:  types.BoolValue(false),
	}

	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, list)
	require.True(t, diags.HasError(), "factory must fail when kibana_connection has no endpoints")
	assert.Nil(t, scoped)
	assert.Equal(t, kibanaFleetClientNotConfiguredError, diags[0].Detail())
}

// --- Task 4.3: factory-level endpoint validation and accessor smoke tests ---

func TestGetElasticsearchClient_ProviderDefaultMissingEndpoint(t *testing.T) {
	t.Setenv("ELASTICSEARCH_ENDPOINTS", "")

	ctx := context.Background()
	factory := NewProviderClientFactory(&apiClient{
		version:     "unit-testing",
		esEndpoints: []string{},
	})

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: elasticsearchConnectionAttrTypes()},
		[]config.ElasticsearchConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetElasticsearchClient(ctx, emptyList)
	require.True(t, diags.HasError(), "factory must fail when no ES endpoint is configured")
	assert.Nil(t, scoped)
	assert.Equal(t, elasticsearchClientNotConfiguredError, diags[0].Detail())
}

func TestGetKibanaClient_ProviderDefaultMissingEndpoint(t *testing.T) {
	t.Setenv("KIBANA_ENDPOINT", "")
	t.Setenv("FLEET_ENDPOINT", "")

	ctx := context.Background()
	factory := NewProviderClientFactory(&apiClient{version: "unit-testing"})

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.True(t, diags.HasError(), "factory must fail when neither Kibana nor Fleet endpoint is configured")
	assert.Nil(t, scoped)
	assert.Equal(t, kibanaFleetClientNotConfiguredError, diags[0].Detail())
}

func TestGetKibanaClient_ProviderFleetEndpointOnly(t *testing.T) {
	t.Setenv("KIBANA_ENDPOINT", "")
	t.Setenv("FLEET_ENDPOINT", "")

	const fleetURL = "https://fleet-only.example.com"
	ctx := context.Background()
	factory, diags := NewProviderClientFactoryFromFramework(ctx, config.ProviderConfiguration{
		Fleet: []config.FleetConnection{
			{
				Endpoint: types.StringValue(fleetURL),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure: types.BoolValue(false),
			},
		},
	}, "test-version")
	require.False(t, diags.HasError(), "factory construction must succeed with fleet-only config: %v", diags)

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.False(t, diags.HasError(), "GetKibanaClient must succeed with fleet-only provider config: %v", diags)
	require.NotNil(t, scoped)

	oapi := scoped.GetKibanaOapiClient()
	require.NotNil(t, oapi)
	require.NotNil(t, scoped.GetFleetClient())
	assert.Equal(t, fleetURL, scoped.kibanaEndpoint)
}

func TestGetElasticsearchClient_SuccessfulFactoryReturnsNonNilAccessor(t *testing.T) {
	t.Parallel()
	srv := newMockElasticsearchServer("8.19.0")
	defer srv.Close()

	scoped := newScopedElasticsearchClientFromFactory(t, srv.URL)
	require.NotNil(t, scoped.GetESClient())
}

func TestGetKibanaClient_SuccessfulFactoryReturnsNonNilAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestFactory(t)

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, emptyList)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
	require.NotNil(t, scoped.GetKibanaOapiClient())
	require.NotNil(t, scoped.GetFleetClient())
}
