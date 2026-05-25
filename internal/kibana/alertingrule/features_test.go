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

package alertingrule

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockKibanaStatusServer(versionStr, buildFlavor string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":%q}}`, versionStr, buildFlavor)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

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

func newKibanaScopedClientForTest(t *testing.T, serverURL string) *clients.KibanaScopedClient {
	t.Helper()
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")
	t.Setenv("KIBANA_ENDPOINT", "")

	ctx := context.Background()
	cfg := config.ProviderConfiguration{
		Kibana: []config.KibanaConnection{
			{
				Username:    types.StringValue("elastic"),
				Password:    types.StringValue("changeme"),
				APIKey:      types.StringValue(""),
				BearerToken: types.StringValue(""),
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue(serverURL),
				}),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure: types.BoolValue(false),
			},
		},
	}

	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, cfg, "test-version")
	require.False(t, diags.HasError(), "factory construction must not fail: %v", diags)

	conn := cfg.Kibana[0]
	list, diags := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()},
		[]config.KibanaConnection{conn},
	)
	require.False(t, diags.HasError())

	scoped, diags := factory.GetKibanaClient(ctx, list)
	require.False(t, diags.HasError())
	return scoped
}

func TestResolveAlertingRuleFeatures(t *testing.T) {
	t.Run("server below all thresholds", func(t *testing.T) {
		srv := newMockKibanaStatusServer("8.0.0", "default")
		t.Cleanup(srv.Close)

		client := newKibanaScopedClientForTest(t, srv.URL)
		features, diags := resolveAlertingRuleFeatures(t.Context(), client)
		require.False(t, diags.HasError())
		assert.Equal(t, alertingRuleFeatures{}, features)
	})

	t.Run("server above all thresholds", func(t *testing.T) {
		srv := newMockKibanaStatusServer("9.3.0", "default")
		t.Cleanup(srv.Close)

		client := newKibanaScopedClientForTest(t, srv.URL)
		features, diags := resolveAlertingRuleFeatures(t.Context(), client)
		require.False(t, diags.HasError())
		assert.Equal(t, alertingRuleFeatures{
			SupportsFrequency:       true,
			SupportsAlertsFilter:    true,
			SupportsAlertDelay:      true,
			SupportsFlapping:        true,
			SupportsFlappingEnabled: true,
		}, features)
	})

	t.Run("serverless", func(t *testing.T) {
		srv := newMockKibanaStatusServer("8.0.0", "serverless")
		t.Cleanup(srv.Close)

		client := newKibanaScopedClientForTest(t, srv.URL)
		features, diags := resolveAlertingRuleFeatures(t.Context(), client)
		require.False(t, diags.HasError())
		assert.Equal(t, alertingRuleFeatures{
			SupportsFrequency:       true,
			SupportsAlertsFilter:    true,
			SupportsAlertDelay:      true,
			SupportsFlapping:        true,
			SupportsFlappingEnabled: true,
		}, features)
	})
}
