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

package resource

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPrivateData map[string][]byte

func (p testPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if val, ok := p[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (p testPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	p[key] = value
	return nil
}

func newMockElasticsearchClient(t *testing.T, version, flavor string) *clients.ElasticsearchScopedClient {
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

func TestApikeyCapabilitiesOfLastRead_LegacyVersionBlob(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty private state returns nil", func(t *testing.T) {
		t.Parallel()
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, testPrivateData{})
		require.False(t, diags.HasError())
		require.Nil(t, caps)
	})

	t.Run("legacy 7.0.0 blob yields all false capabilities", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{
			clusterVersionPrivateDataKey: []byte(`{"Version":"7.0.0"}`),
		}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.False(t, caps.SupportsUpdate)
		assert.False(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("legacy 8.20.0 blob yields all true capabilities", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{
			clusterVersionPrivateDataKey: []byte(`{"Version":"8.20.0"}`),
		}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.True(t, caps.SupportsRoleDescriptors)
		assert.True(t, caps.SupportsRestriction)
	})

	t.Run("new capability blob is returned as-is", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{
			clusterVersionPrivateDataKey: []byte(`{"SupportsUpdate":true,"SupportsRoleDescriptors":false,"SupportsRestriction":true}`),
		}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.False(t, caps.SupportsRoleDescriptors)
		assert.True(t, caps.SupportsRestriction)
	})

	t.Run("legacy empty version falls through to nil", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{
			clusterVersionPrivateDataKey: []byte(`{"Version":""}`),
		}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.Nil(t, caps)
	})

	t.Run("all-false new-format blob is preserved", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{
			clusterVersionPrivateDataKey: []byte(`{"SupportsUpdate":false,"SupportsRoleDescriptors":false,"SupportsRestriction":false}`),
		}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.False(t, caps.SupportsUpdate)
		assert.False(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("legacy 8.4.0 enables update only", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{clusterVersionPrivateDataKey: []byte(`{"Version":"8.4.0"}`)}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.False(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("legacy 8.5.0 enables update and role descriptors", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{clusterVersionPrivateDataKey: []byte(`{"Version":"8.5.0"}`)}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.True(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("legacy 8.8.9 enables update and role descriptors but not restriction", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{clusterVersionPrivateDataKey: []byte(`{"Version":"8.8.9"}`)}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.True(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("legacy 8.9.0 enables all capabilities", func(t *testing.T) {
		t.Parallel()
		priv := testPrivateData{clusterVersionPrivateDataKey: []byte(`{"Version":"8.9.0"}`)}
		caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, diags.HasError())
		require.NotNil(t, caps)
		assert.True(t, caps.SupportsUpdate)
		assert.True(t, caps.SupportsRoleDescriptors)
		assert.True(t, caps.SupportsRestriction)
	})
}

func TestSaveAPIKeyCapabilities(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("serverless persists all true", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClient(t, "8.19.0", clients.ServerlessFlavor)
		priv := testPrivateData{}

		diags := saveAPIKeyCapabilities(ctx, client, priv)
		require.False(t, diags.HasError())

		var stored apikey.APIKeyCapabilities
		require.NoError(t, json.Unmarshal(priv[clusterVersionPrivateDataKey], &stored))
		assert.True(t, stored.SupportsUpdate)
		assert.True(t, stored.SupportsRoleDescriptors)
		assert.True(t, stored.SupportsRestriction)
	})

	t.Run("stateful 8.0.0 persists and round-trips all-false capabilities", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClient(t, "8.0.0", "default")
		priv := testPrivateData{}

		diags := saveAPIKeyCapabilities(ctx, client, priv)
		require.False(t, diags.HasError())

		caps, readDiags := apikeyCapabilitiesOfLastRead(ctx, priv)
		require.False(t, readDiags.HasError())
		require.NotNil(t, caps)
		assert.False(t, caps.SupportsUpdate)
		assert.False(t, caps.SupportsRoleDescriptors)
		assert.False(t, caps.SupportsRestriction)
	})

	t.Run("stateful mixed version persists expected flags", func(t *testing.T) {
		t.Parallel()
		client := newMockElasticsearchClient(t, "8.6.0", "default")
		priv := testPrivateData{}

		diags := saveAPIKeyCapabilities(ctx, client, priv)
		require.False(t, diags.HasError())

		var stored apikey.APIKeyCapabilities
		require.NoError(t, json.Unmarshal(priv[clusterVersionPrivateDataKey], &stored))
		assert.True(t, stored.SupportsUpdate)
		assert.True(t, stored.SupportsRoleDescriptors)
		assert.False(t, stored.SupportsRestriction)
	})
}
