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

package osquerypack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKibanaScopedClient(t *testing.T, server *httptest.Server) *clients.KibanaScopedClient {
	t.Helper()
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	scopedClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	require.NotNil(t, scopedClient.GetKibanaOapiClient())
	return scopedClient
}

func osqueryPackFindResponseBody(t *testing.T, name, savedObjectID string, readOnly *bool) []byte {
	t.Helper()
	data := map[string]any{
		"name":            name,
		"saved_object_id": savedObjectID,
	}
	if readOnly != nil {
		data["read_only"] = *readOnly
	}
	body, err := json.Marshal(map[string]any{"data": data})
	require.NoError(t, err)
	return body
}

func TestReadOsqueryPackDataSource_invalidPackID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &clients.KibanaScopedClient{}

	tests := []struct {
		name   string
		packID types.String
	}{
		{name: "null", packID: types.StringNull()},
		{name: "unknown", packID: types.StringUnknown()},
		{name: "empty string", packID: types.StringValue("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := readOsqueryPackDataSource(ctx, client, dataSourceModel{
				osqueryPackBaseModel: osqueryPackBaseModel{PackID: tt.packID},
			})
			require.True(t, diags.HasError())
			require.Equal(t, "Invalid configuration", diags.Errors()[0].Summary())
			assert.Contains(t, diags.Errors()[0].Detail(), "pack_id must be set")
		})
	}
}

func TestReadOsqueryPackDataSource_notFound(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/osquery/packs/missing-pack", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := newTestKibanaScopedClient(t, server)
	_, diags := readOsqueryPackDataSource(ctx, client, dataSourceModel{
		osqueryPackBaseModel: osqueryPackBaseModel{PackID: types.StringValue("missing-pack")},
	})

	require.True(t, diags.HasError())
	require.Equal(t, "Osquery pack not found", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "missing-pack")
	assert.Contains(t, diags.Errors()[0].Detail(), "default")
}

func TestReadOsqueryPackDataSource_prebuiltPack(t *testing.T) {
	ctx := context.Background()

	readOnly := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/osquery/packs/prebuilt-pack-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(osqueryPackFindResponseBody(t, "Prebuilt pack", "prebuilt-pack-id", &readOnly))
	}))
	t.Cleanup(server.Close)

	client := newTestKibanaScopedClient(t, server)
	result, diags := readOsqueryPackDataSource(ctx, client, dataSourceModel{
		osqueryPackBaseModel: osqueryPackBaseModel{PackID: types.StringValue("prebuilt-pack-id")},
	})

	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, result.ReadOnly.ValueBool())
	require.Equal(t, "Prebuilt pack", result.Name.ValueString())
	require.Equal(t, "default/prebuilt-pack-id", result.ID.ValueString())
}

func TestReadOsqueryPackDataSource_nonDefaultSpace(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/s/staging/api/osquery/packs/staging-pack-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(osqueryPackFindResponseBody(t, "Staging pack", "staging-pack-id", nil))
	}))
	t.Cleanup(server.Close)

	client := newTestKibanaScopedClient(t, server)
	result, diags := readOsqueryPackDataSource(ctx, client, dataSourceModel{
		osqueryPackBaseModel: osqueryPackBaseModel{
			PackID:  types.StringValue("staging-pack-id"),
			SpaceID: types.StringValue("staging"),
		},
	})

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, "staging", result.SpaceID.ValueString())
	require.Equal(t, "staging/staging-pack-id", result.ID.ValueString())
}

func TestReadOsqueryPackDataSource_defaultSpaceWhenOmitted(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/osquery/packs/pack-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(osqueryPackFindResponseBody(t, "Default space pack", "pack-id", nil))
	}))
	t.Cleanup(server.Close)

	client := newTestKibanaScopedClient(t, server)
	result, diags := readOsqueryPackDataSource(ctx, client, dataSourceModel{
		osqueryPackBaseModel: osqueryPackBaseModel{
			PackID:  types.StringValue("pack-id"),
			SpaceID: types.StringNull(),
		},
	})

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, "default", result.SpaceID.ValueString())
}
