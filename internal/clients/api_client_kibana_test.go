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

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAPIClient returns a minimal *APIClient suitable for unit tests.
// It sets up Kibana clients with a fake endpoint so that actual network
// calls are never made.
func newTestAPIClient(t *testing.T) *APIClient {
	t.Helper()
	kib, err := kibana.NewClient(kibana.Config{
		Address:  "http://localhost:5601",
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{
		URL:      "http://localhost:5601",
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	return &APIClient{
		kibana:       kib,
		kibanaOapi:   kibOapi,
		kibanaConfig: kibana.Config{Address: "http://localhost:5601", Username: "elastic", Password: "changeme"},
		version:      "unit-testing",
	}
}

// TestMaybeNewKibanaAPIClientFromFrameworkResource_EmptyList verifies that
// when kibana_connection is absent the default client is returned unchanged.
func TestMaybeNewKibanaAPIClientFromFrameworkResource_EmptyList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	defaultClient := newTestAPIClient(t)

	emptyList, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())

	result, diags := MaybeNewKibanaAPIClientFromFrameworkResource(ctx, emptyList, defaultClient)
	require.False(t, diags.HasError())
	assert.Same(t, defaultClient, result, "should return the same default client when kibana_connection is absent")
}

// TestMaybeNewKibanaAPIClientFromFrameworkResource_NullList verifies that a
// null list is treated the same as an empty list.
func TestMaybeNewKibanaAPIClientFromFrameworkResource_NullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	defaultClient := newTestAPIClient(t)

	nullList := types.ListNull(types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()})

	result, diags := MaybeNewKibanaAPIClientFromFrameworkResource(ctx, nullList, defaultClient)
	require.False(t, diags.HasError())
	assert.Same(t, defaultClient, result, "should return the same default client when kibana_connection is null")
}

// TestMaybeNewKibanaAPIClientFromFrameworkResource_WithConnection verifies
// that when a kibana_connection block is present a new scoped client is
// returned and it does NOT carry an Elasticsearch client (so that version
// and identity checks resolve against the scoped Kibana endpoint).
func TestMaybeNewKibanaAPIClientFromFrameworkResource_WithConnection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	defaultClient := newTestAPIClient(t)

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

	result, diags := MaybeNewKibanaAPIClientFromFrameworkResource(ctx, list, defaultClient)
	require.False(t, diags.HasError())
	require.NotNil(t, result)

	// The scoped client must not be the same object as the default client.
	assert.NotSame(t, defaultClient, result, "should return a new client when kibana_connection is set")

	// The scoped client must NOT have an Elasticsearch client so version/flavor
	// checks fall back to the Kibana path.
	assert.Nil(t, result.elasticsearch, "scoped Kibana client must not carry an Elasticsearch client")

	// Kibana-derived client surfaces must be populated.
	_, err := result.GetKibanaClient()
	require.NoError(t, err, "Kibana legacy client must be present")

	_, err = result.GetKibanaOapiClient()
	require.NoError(t, err, "Kibana OpenAPI client must be present")

	_, err = result.GetFleetClient()
	assert.NoError(t, err, "Fleet client must be present")
}

// TestNewKibanaAPIClientFromSDKResource_AbsentBlock verifies that when the
// kibana_connection block is absent the provider-level default client is
// returned unchanged.
func TestNewKibanaAPIClientFromSDKResource_AbsentBlock(t *testing.T) {
	t.Parallel()
	defaultClient := newTestAPIClient(t)

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"kibana_connection": providerschema.GetKibanaEntityConnectionSchema(),
	}, map[string]any{})

	result, diags := NewKibanaAPIClientFromSDKResource(rd, defaultClient)
	require.False(t, diags.HasError())
	assert.Same(t, defaultClient, result, "should return default client when kibana_connection block is absent")
}

// TestNewKibanaAPIClientFromSDKResource_WithBlock verifies that when a
// kibana_connection block is present a new scoped client is returned with
// Kibana-derived clients rebuilt and no Elasticsearch client set.
func TestNewKibanaAPIClientFromSDKResource_WithBlock(t *testing.T) {
	t.Parallel()
	defaultClient := newTestAPIClient(t)

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"kibana_connection": providerschema.GetKibanaEntityConnectionSchema(),
	}, map[string]any{
		"kibana_connection": []any{
			map[string]any{
				"username":     "kibana-user",
				"password":     "kibana-pass",
				"endpoints":    []any{"http://kibana.example.com:5601"},
				"ca_certs":     []any{},
				"insecure":     false,
				"api_key":      "",
				"bearer_token": "",
			},
		},
	})

	result, diags := NewKibanaAPIClientFromSDKResource(rd, defaultClient)
	require.False(t, diags.HasError())
	require.NotNil(t, result)

	// Must not be the same client as the default.
	assert.NotSame(t, defaultClient, result)

	// Must NOT carry an Elasticsearch client.
	assert.Nil(t, result.elasticsearch, "scoped Kibana client must not carry an Elasticsearch client")

	// Kibana-derived surfaces must be populated.
	_, err := result.GetKibanaClient()
	require.NoError(t, err, "Kibana legacy client must be present")

	_, err = result.GetKibanaOapiClient()
	require.NoError(t, err, "Kibana OpenAPI client must be present")

	_, err = result.GetFleetClient()
	assert.NoError(t, err, "Fleet client must be present")
}

// newMockKibanaServer returns an httptest.Server that responds to GET /api/status
// with a minimal Kibana status payload for the given version string.
// The caller must call Close() on the returned server.
func newMockKibanaServer(version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":"default"}}`, version)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// newScopedKibanaClient creates a scoped *APIClient via
// MaybeNewKibanaAPIClientFromFrameworkResource pointing at the given endpoint.
// TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT is set via t.Setenv so
// that environment-variable endpoint overrides do not redirect the client.
func newScopedKibanaClient(t *testing.T, endpoint string) *APIClient {
	t.Helper()
	// Lock the endpoint so the KIBANA_ENDPOINT env var (if set in CI) cannot
	// override the explicitly configured URL.
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")

	ctx := context.Background()
	defaultClient := newTestAPIClient(t)
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

	scoped, diags := MaybeNewKibanaAPIClientFromFrameworkResource(ctx, list, defaultClient)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)
	return scoped
}

// TestScopedKibanaClient_ServerVersion_RoutesFromKibana verifies that
// ServerVersion on a scoped Kibana client (no Elasticsearch) routes through
// the Kibana status API rather than the Elasticsearch info API.
func TestScopedKibanaClient_ServerVersion_RoutesFromKibana(t *testing.T) {
	const wantVersion = "8.18.0"
	srv := newMockKibanaServer(wantVersion)
	defer srv.Close()

	scoped := newScopedKibanaClient(t, srv.URL)
	require.Nil(t, scoped.elasticsearch, "precondition: scoped client must have no Elasticsearch client")

	ver, diags := scoped.ServerVersion(context.Background())
	require.False(t, diags.HasError(), "ServerVersion must succeed against the mock Kibana server")
	require.NotNil(t, ver)
	assert.Equal(t, wantVersion, ver.Original(),
		"version must come from the mock Kibana server, not from an Elasticsearch path")
}

// TestScopedKibanaClient_ServerFlavor_RoutesFromKibana verifies that
// ServerFlavor on a scoped Kibana client routes through the Kibana status API.
func TestScopedKibanaClient_ServerFlavor_RoutesFromKibana(t *testing.T) {
	const wantVersion = "8.18.0"
	srv := newMockKibanaServer(wantVersion)
	defer srv.Close()

	scoped := newScopedKibanaClient(t, srv.URL)
	require.Nil(t, scoped.elasticsearch, "precondition: scoped client must have no Elasticsearch client")

	flavor, diags := scoped.ServerFlavor(context.Background())
	require.False(t, diags.HasError(), "ServerFlavor must succeed against the mock Kibana server")
	assert.Equal(t, "default", flavor,
		"flavor must come from the mock Kibana server, not from an Elasticsearch path")
}

// kibanaConnectionAttrTypes returns the attribute type map for
// config.KibanaConnection so we can build framework type values in tests.
func kibanaConnectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username":     types.StringType,
		"password":     types.StringType,
		"api_key":      types.StringType,
		"bearer_token": types.StringType,
		"endpoints":    types.ListType{ElemType: types.StringType},
		"ca_certs":     types.ListType{ElemType: types.StringType},
		"insecure":     types.BoolType,
	}
}
