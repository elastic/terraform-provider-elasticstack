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

	kibana "github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	// The scoped client must expose a Kibana legacy client.
	_, err := scoped.GetKibanaClient()
	require.NoError(t, err, "Kibana legacy client must be present on provider-default scoped client")
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
	_, err := scoped.GetKibanaClient()
	require.NoError(t, err)

	_, err = scoped.GetKibanaOapiClient()
	require.NoError(t, err)

	_, err = scoped.GetFleetClient()
	require.NoError(t, err)
}

// --- ProviderClientFactory.GetKibanaClientFromSDK ---

// TestGetKibanaClientFromSDK_AbsentBlock verifies that the factory returns a
// provider-default scoped client when kibana_connection is absent.
func TestGetKibanaClientFromSDK_AbsentBlock(t *testing.T) {
	t.Parallel()
	factory := newTestFactory(t)

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"kibana_connection": providerschema.GetKibanaEntityConnectionSchema(),
	}, map[string]any{})

	scoped, diags := factory.GetKibanaClientFromSDK(rd)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)

	// Must expose a Kibana legacy client.
	_, err := scoped.GetKibanaClient()
	require.NoError(t, err)
}

// TestGetKibanaClientFromSDK_WithBlock verifies that when kibana_connection is
// set the factory returns a new scoped client with rebuilt clients.
func TestGetKibanaClientFromSDK_WithBlock(t *testing.T) {
	t.Parallel()
	factory := newTestFactory(t)

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

	scoped, diags := factory.GetKibanaClientFromSDK(rd)
	require.False(t, diags.HasError())
	require.NotNil(t, scoped)

	_, err := scoped.GetKibanaClient()
	require.NoError(t, err)

	_, err = scoped.GetKibanaOapiClient()
	require.NoError(t, err)

	_, err = scoped.GetFleetClient()
	require.NoError(t, err)
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

// TestKibanaScopedClient_ServerVersion_ViaFactory verifies that
// ServerVersion on a scoped client obtained from the factory routes through
// the Kibana status API and not Elasticsearch.
func TestKibanaScopedClient_ServerVersion_ViaFactory(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockKibanaServer(wantVersion)
	defer srv.Close()

	scoped := newScopedClientFromFactory(t, srv.URL)

	ver, diags := scoped.ServerVersion(context.Background())
	require.False(t, diags.HasError())
	require.NotNil(t, ver)
	assert.Equal(t, wantVersion, ver.Original())
}

// TestKibanaScopedClient_ServerFlavor_ViaFactory verifies that
// ServerFlavor on a factory-obtained scoped client routes through Kibana.
func TestKibanaScopedClient_ServerFlavor_ViaFactory(t *testing.T) {
	const wantVersion = "8.19.0"
	srv := newMockKibanaServer(wantVersion)
	defer srv.Close()

	scoped := newScopedClientFromFactory(t, srv.URL)

	flavor, diags := scoped.ServerFlavor(context.Background())
	require.False(t, diags.HasError())
	assert.Equal(t, "default", flavor)
}

// TestKibanaScopedClient_ServerlessEnforceMinVersion verifies that
// EnforceMinVersion always returns true for serverless Kibana.
func TestKibanaScopedClient_ServerlessEnforceMinVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
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

// --- ConvertMetaToFactory ---

func TestConvertMetaToFactory_NilMeta(t *testing.T) {
	t.Parallel()
	factory, diags := ConvertMetaToFactory(nil)
	assert.True(t, diags.HasError(), "nil meta must return an error diagnostic")
	assert.Nil(t, factory)
}

func TestConvertMetaToFactory_WrongType(t *testing.T) {
	t.Parallel()
	_, diags := ConvertMetaToFactory("unexpected-string")
	assert.True(t, diags.HasError())
}

func TestConvertMetaToFactory_Valid(t *testing.T) {
	t.Parallel()
	f := newTestFactory(t)
	result, diags := ConvertMetaToFactory(f)
	require.False(t, diags.HasError())
	assert.Same(t, f, result)
}

// --- NewKibanaScopedClientFromFactory ---

func TestNewKibanaScopedClientFromFactory_NilFactory(t *testing.T) {
	t.Parallel()
	result := NewKibanaScopedClientFromFactory(nil)
	assert.Nil(t, result)
}

func TestNewKibanaScopedClientFromFactory_Valid(t *testing.T) {
	t.Parallel()
	f := newTestFactory(t)
	result := NewKibanaScopedClientFromFactory(f)
	require.NotNil(t, result)
	_, err := result.GetKibanaClient()
	require.NoError(t, err)
}

// --- KibanaScopedClient.SetSloAuthContext ---

func TestSetSloAuthContext_ApiKey(t *testing.T) {
	t.Parallel()
	scoped := &KibanaScopedClient{
		kibanaConfig: kibana.Config{ApiKey: "my-api-key"},
	}

	ctx := scoped.SetSloAuthContext(context.Background())
	keys, ok := ctx.Value(slo.ContextAPIKeys).(map[string]slo.APIKey)
	require.True(t, ok, "expected ContextAPIKeys in context")
	key, exists := keys["apiKeyAuth"]
	require.True(t, exists, "expected apiKeyAuth key")
	assert.Equal(t, "ApiKey", key.Prefix)
	assert.Equal(t, "my-api-key", key.Key)
}

func TestSetSloAuthContext_BasicAuth(t *testing.T) {
	t.Parallel()
	scoped := &KibanaScopedClient{
		kibanaConfig: kibana.Config{Username: "user", Password: "pass"},
	}

	ctx := scoped.SetSloAuthContext(context.Background())
	auth, ok := ctx.Value(slo.ContextBasicAuth).(slo.BasicAuth)
	require.True(t, ok, "expected ContextBasicAuth in context")
	assert.Equal(t, "user", auth.UserName)
	assert.Equal(t, "pass", auth.Password)
}
