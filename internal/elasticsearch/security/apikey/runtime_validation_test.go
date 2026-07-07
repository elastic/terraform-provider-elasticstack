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

package apikey

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockElasticsearchClientForAPIKeyTests(t *testing.T, version, flavor string) *clients.ElasticsearchScopedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	t.Cleanup(srv.Close)

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
	})
	require.NoError(t, err)

	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

func TestValidateRestrictionSupport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	noRestrictionModel := TfModel{
		RoleDescriptors: customtypes.NewJSONWithDefaultsValue(
			`{"role-a":{"cluster":["all"]}}`,
			PopulateRoleDescriptorsDefaults,
		),
	}
	withRestrictionModel := TfModel{
		RoleDescriptors: customtypes.NewJSONWithDefaultsValue(
			`{"role-a":{"cluster":["all"],"restriction":{"workflows":["search-application"]}}}`,
			PopulateRoleDescriptorsDefaults,
		),
	}
	malformedModel := TfModel{
		RoleDescriptors: customtypes.NewJSONWithDefaultsValue(
			`{invalid`,
			PopulateRoleDescriptorsDefaults,
		),
	}

	t.Run("no restrictions skips capability lookup", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClientForAPIKeyTests(t, "7.0.0", "default")
		diags := ValidateRestrictionSupport(ctx, client, noRestrictionModel)
		assert.False(t, diags.HasError())
	})

	t.Run("serverless allows restrictions", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClientForAPIKeyTests(t, "8.6.0", clients.ServerlessFlavor)
		diags := ValidateRestrictionSupport(ctx, client, withRestrictionModel)
		assert.False(t, diags.HasError())
	})

	t.Run("stateful below min version rejects restrictions", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClientForAPIKeyTests(t, "8.6.0", "default")
		diags := ValidateRestrictionSupport(ctx, client, withRestrictionModel)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "not supported in this version of Elasticsearch")
		assert.Contains(t, diags.Errors()[0].Detail(), "role-a")
	})

	t.Run("stateful at min version allows restrictions", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClientForAPIKeyTests(t, "8.9.0", "default")
		diags := ValidateRestrictionSupport(ctx, client, withRestrictionModel)
		assert.False(t, diags.HasError())
	})

	t.Run("malformed role descriptors propagates error", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClientForAPIKeyTests(t, "8.9.0", "default")
		diags := ValidateRestrictionSupport(ctx, client, malformedModel)
		require.True(t, diags.HasError())
	})
}

func TestSynthesizeAPIKeyCapabilitiesFromVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version             string
		wantUpdate          bool
		wantRoleDescriptors bool
		wantRestriction     bool
	}{
		{version: "7.0.0", wantUpdate: false, wantRoleDescriptors: false, wantRestriction: false},
		{version: "8.4.0", wantUpdate: true, wantRoleDescriptors: false, wantRestriction: false},
		{version: "8.5.0", wantUpdate: true, wantRoleDescriptors: true, wantRestriction: false},
		{version: "8.8.9", wantUpdate: true, wantRoleDescriptors: true, wantRestriction: false},
		{version: "8.9.0", wantUpdate: true, wantRoleDescriptors: true, wantRestriction: true},
		{version: "8.20.0", wantUpdate: true, wantRoleDescriptors: true, wantRestriction: true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			ver := version.Must(version.NewVersion(tt.version))
			caps := SynthesizeAPIKeyCapabilitiesFromVersion(ver)
			assert.Equal(t, tt.wantUpdate, caps.SupportsUpdate)
			assert.Equal(t, tt.wantRoleDescriptors, caps.SupportsRoleDescriptors)
			assert.Equal(t, tt.wantRestriction, caps.SupportsRestriction)
		})
	}
}
