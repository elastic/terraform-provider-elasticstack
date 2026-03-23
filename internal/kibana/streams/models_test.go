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

package streams

import (
	"context"
	"encoding/json"
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── streamType discriminator ──────────────────────────────────────────────────

func TestStreamType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		model    streamModel
		expected string
	}{
		{
			name:     "wired config set",
			model:    streamModel{WiredConfig: &wiredConfigModel{}},
			expected: streamTypeWired,
		},
		{
			name:     "classic config set",
			model:    streamModel{ClassicConfig: &classicConfigModel{}},
			expected: streamTypeClassic,
		},
		{
			name:     "query config set",
			model:    streamModel{QueryConfig: &queryConfigModel{}},
			expected: streamTypeQuery,
		},
		{
			name:     "no config set returns empty string",
			model:    streamModel{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.model.streamType())
		})
	}
}

// ── wiredConfigModel ─────────────────────────────────────────────────────────

func TestWiredConfigPopulateFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("full ingest populates all fields", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{
			Processing: kibanaoapi.StreamProcessing{
				Steps: json.RawMessage(`[{"grok":{"field":"message","patterns":["%%{GREEDYDATA:msg}"]}}]`),
			},
			Wired: &kibanaoapi.StreamIngestWired{
				Fields:  json.RawMessage(`{"host.name":{"type":"keyword"}}`),
				Routing: []kibanaoapi.StreamRoutingRule{{Destination: "logs.nginx.errors", Where: json.RawMessage(`{}`)}},
			},
			Lifecycle:    json.RawMessage(`{"dsl":{"data_retention":"30d"}}`),
			FailureStore: json.RawMessage(`{"inherit":{}}`),
			Settings: kibanaoapi.StreamIngestSettings{
				IndexNumberOfShards:   &kibanaoapi.StreamIngestSettingValue{Value: float64(2)},
				IndexNumberOfReplicas: &kibanaoapi.StreamIngestSettingValue{Value: float64(1)},
				IndexRefreshInterval:  &kibanaoapi.StreamIngestSettingValue{Value: "5s"},
			},
		}

		var m wiredConfigModel
		diags := m.populateFromAPI(ctx, ingest)
		require.False(t, diags.HasError())

		assert.False(t, m.ProcessingStepsJSON.IsNull())
		assert.False(t, m.FieldsJSON.IsNull())
		assert.False(t, m.RoutingJSON.IsNull())
		assert.False(t, m.LifecycleJSON.IsNull())
		assert.False(t, m.FailureStoreJSON.IsNull())
		assert.Equal(t, types.Int64Value(2), m.IndexNumberOfShards)
		assert.Equal(t, types.Int64Value(1), m.IndexNumberOfReplicas)
		assert.Equal(t, types.StringValue("5s"), m.IndexRefreshInterval)
	})

	t.Run("nil ingest is a no-op", func(t *testing.T) {
		t.Parallel()
		var m wiredConfigModel
		diags := m.populateFromAPI(ctx, nil)
		require.False(t, diags.HasError())
	})

	t.Run("nil Wired block sets fields and routing to null", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{}
		var m wiredConfigModel
		_ = m.populateFromAPI(ctx, ingest)
		assert.True(t, m.FieldsJSON.IsNull())
		assert.True(t, m.RoutingJSON.IsNull())
	})

	t.Run("nil index settings produce null Terraform values", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{}
		var m wiredConfigModel
		_ = m.populateFromAPI(ctx, ingest)
		assert.True(t, m.IndexNumberOfShards.IsNull())
		assert.True(t, m.IndexNumberOfReplicas.IsNull())
		assert.True(t, m.IndexRefreshInterval.IsNull())
	})

	t.Run("non-float64 index setting value produces null, not stale state", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{
			Settings: kibanaoapi.StreamIngestSettings{
				// API returned a string instead of a number — type assertion should fail gracefully.
				IndexNumberOfShards:   &kibanaoapi.StreamIngestSettingValue{Value: "not-a-number"},
				IndexNumberOfReplicas: &kibanaoapi.StreamIngestSettingValue{Value: true},
			},
		}
		var m wiredConfigModel
		_ = m.populateFromAPI(ctx, ingest)
		assert.True(t, m.IndexNumberOfShards.IsNull(), "expected null for non-float64 shards value")
		assert.True(t, m.IndexNumberOfReplicas.IsNull(), "expected null for non-float64 replicas value")
	})
}

func TestWiredConfigToAPIIngest(t *testing.T) {
	t.Parallel()

	t.Run("nil fields and routing produce nil Wired block", func(t *testing.T) {
		t.Parallel()
		m := wiredConfigModel{
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			FieldsJSON:            jsontypes.NewNormalizedNull(),
			RoutingJSON:           jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		assert.Nil(t, ingest.Wired, "Wired should be nil when neither fields_json nor routing_json are set")
	})

	t.Run("only fields_json set populates Wired with fields only", func(t *testing.T) {
		t.Parallel()
		m := wiredConfigModel{
			FieldsJSON:            jsontypes.NewNormalizedValue(`{"host.name":{"type":"keyword"}}`),
			RoutingJSON:           jsontypes.NewNormalizedNull(),
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		require.NotNil(t, ingest.Wired)
		assert.NotNil(t, ingest.Wired.Fields)
		assert.Nil(t, ingest.Wired.Routing)
	})

	t.Run("only routing_json set populates Wired with routing only", func(t *testing.T) {
		t.Parallel()
		m := wiredConfigModel{
			FieldsJSON:            jsontypes.NewNormalizedNull(),
			RoutingJSON:           jsontypes.NewNormalizedValue(`[{"destination":"logs.errors","where":{}}]`),
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		require.NotNil(t, ingest.Wired)
		assert.Nil(t, ingest.Wired.Fields)
		assert.Len(t, ingest.Wired.Routing, 1)
		assert.Equal(t, "logs.errors", ingest.Wired.Routing[0].Destination)
	})

	t.Run("invalid routing_json adds error diagnostic", func(t *testing.T) {
		t.Parallel()
		m := wiredConfigModel{
			RoutingJSON:           jsontypes.NewNormalizedValue(`not-valid-json`),
			FieldsJSON:            jsontypes.NewNormalizedNull(),
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		_ = m.toAPIIngest(&diags)
		assert.True(t, diags.HasError())
	})

	t.Run("index settings serialised correctly", func(t *testing.T) {
		t.Parallel()
		m := wiredConfigModel{
			FieldsJSON:            jsontypes.NewNormalizedNull(),
			RoutingJSON:           jsontypes.NewNormalizedNull(),
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Value(3),
			IndexNumberOfReplicas: types.Int64Value(1),
			IndexRefreshInterval:  types.StringValue("1s"),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		require.NotNil(t, ingest.Settings.IndexNumberOfShards)
		assert.InDelta(t, float64(3), ingest.Settings.IndexNumberOfShards.Value, 0)
		require.NotNil(t, ingest.Settings.IndexNumberOfReplicas)
		assert.InDelta(t, float64(1), ingest.Settings.IndexNumberOfReplicas.Value, 0)
		require.NotNil(t, ingest.Settings.IndexRefreshInterval)
		assert.Equal(t, "1s", ingest.Settings.IndexRefreshInterval.Value)
	})
}

// ── classicConfigModel ────────────────────────────────────────────────────────

func TestClassicConfigPopulateFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("non-float64 index setting value produces null", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{
			Settings: kibanaoapi.StreamIngestSettings{
				IndexNumberOfShards:   &kibanaoapi.StreamIngestSettingValue{Value: "bad-value"},
				IndexNumberOfReplicas: &kibanaoapi.StreamIngestSettingValue{Value: struct{}{}},
			},
		}
		var m classicConfigModel
		m.populateFromAPI(ctx, ingest)
		assert.True(t, m.IndexNumberOfShards.IsNull())
		assert.True(t, m.IndexNumberOfReplicas.IsNull())
	})

	t.Run("field overrides populated from classic block", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{
			Classic: &kibanaoapi.StreamIngestClassic{
				FieldOverrides: json.RawMessage(`{"host.name":{"type":"keyword"}}`),
			},
		}
		var m classicConfigModel
		m.populateFromAPI(ctx, ingest)
		assert.False(t, m.FieldOverridesJSON.IsNull())
	})

	t.Run("nil classic block sets field overrides to null", func(t *testing.T) {
		t.Parallel()
		ingest := &kibanaoapi.StreamIngest{}
		var m classicConfigModel
		m.populateFromAPI(ctx, ingest)
		assert.True(t, m.FieldOverridesJSON.IsNull())
	})
}

func TestClassicConfigToAPIIngest(t *testing.T) {
	t.Parallel()

	t.Run("field overrides set populates Classic block", func(t *testing.T) {
		t.Parallel()
		m := classicConfigModel{
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			FieldOverridesJSON:    jsontypes.NewNormalizedValue(`{"host.name":{"type":"keyword"}}`),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		require.NotNil(t, ingest.Classic)
		assert.NotNil(t, ingest.Classic.FieldOverrides)
	})

	t.Run("nil field overrides does not populate Classic block", func(t *testing.T) {
		t.Parallel()
		m := classicConfigModel{
			ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
			FieldOverridesJSON:    jsontypes.NewNormalizedNull(),
			LifecycleJSON:         jsontypes.NewNormalizedNull(),
			FailureStoreJSON:      jsontypes.NewNormalizedNull(),
			IndexNumberOfShards:   types.Int64Null(),
			IndexNumberOfReplicas: types.Int64Null(),
			IndexRefreshInterval:  types.StringNull(),
		}
		var diags diag.Diagnostics
		ingest := m.toAPIIngest(&diags)
		require.False(t, diags.HasError())
		assert.Nil(t, ingest.Classic)
	})
}

// ── queryConfigModel ──────────────────────────────────────────────────────────

func TestQueryConfigRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("populateFromAPI with view", func(t *testing.T) {
		t.Parallel()
		q := &kibanaoapi.StreamQueryESQLDef{Esql: "FROM logs* | LIMIT 10", View: "my-view"}
		var m queryConfigModel
		m.populateFromAPI(q)
		assert.Equal(t, types.StringValue("FROM logs* | LIMIT 10"), m.Esql)
		assert.Equal(t, types.StringValue("my-view"), m.View)
	})

	t.Run("populateFromAPI without view produces null", func(t *testing.T) {
		t.Parallel()
		q := &kibanaoapi.StreamQueryESQLDef{Esql: "FROM logs*"}
		var m queryConfigModel
		m.populateFromAPI(q)
		assert.Equal(t, types.StringValue("FROM logs*"), m.Esql)
		assert.True(t, m.View.IsNull())
	})

	t.Run("populateFromAPI nil is no-op", func(t *testing.T) {
		t.Parallel()
		var m queryConfigModel
		m.populateFromAPI(nil)
		assert.True(t, m.Esql.IsNull() || m.Esql.IsUnknown() || m.Esql.ValueString() == "")
	})

	t.Run("toAPI round-trips esql and view", func(t *testing.T) {
		t.Parallel()
		m := queryConfigModel{
			Esql: types.StringValue("FROM logs*"),
			View: types.StringValue("my-view"),
		}
		q := m.toAPI()
		require.NotNil(t, q)
		assert.Equal(t, "FROM logs*", q.Esql)
		assert.Equal(t, "my-view", q.View)
	})

	t.Run("toAPI with null view omits view", func(t *testing.T) {
		t.Parallel()
		m := queryConfigModel{
			Esql: types.StringValue("FROM logs*"),
			View: types.StringNull(),
		}
		q := m.toAPI()
		require.NotNil(t, q)
		assert.Equal(t, "FROM logs*", q.Esql)
		assert.Empty(t, q.View)
	})
}

// ── streamModel.toAPIUpsertRequest ────────────────────────────────────────────

func TestToAPIUpsertRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("wired stream type and description set correctly", func(t *testing.T) {
		t.Parallel()
		m := streamModel{
			Name:        types.StringValue("logs.nginx"),
			SpaceID:     types.StringValue("default"),
			Description: types.StringValue("Nginx logs"),
			WiredConfig: &wiredConfigModel{
				ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
				FieldsJSON:            jsontypes.NewNormalizedNull(),
				RoutingJSON:           jsontypes.NewNormalizedNull(),
				LifecycleJSON:         jsontypes.NewNormalizedNull(),
				FailureStoreJSON:      jsontypes.NewNormalizedNull(),
				IndexNumberOfShards:   types.Int64Null(),
				IndexNumberOfReplicas: types.Int64Null(),
				IndexRefreshInterval:  types.StringNull(),
			},
			Dashboards: types.ListNull(types.StringType),
		}
		var diags diag.Diagnostics
		req := m.toAPIUpsertRequest(ctx, &diags)
		require.False(t, diags.HasError())
		assert.Equal(t, streamTypeWired, req.Stream.Type)
		assert.Equal(t, "Nginx logs", req.Stream.Description)
		assert.NotNil(t, req.Stream.Ingest)
		assert.Nil(t, req.Stream.Query)
	})

	t.Run("query stream type set correctly", func(t *testing.T) {
		t.Parallel()
		m := streamModel{
			Name:        types.StringValue("logs.nginx.errors-view"),
			SpaceID:     types.StringValue("default"),
			Description: types.StringValue(""),
			QueryConfig: &queryConfigModel{
				Esql: types.StringValue("FROM logs.nginx | WHERE http.response.status_code >= 400"),
				View: types.StringNull(),
			},
			Dashboards: types.ListNull(types.StringType),
		}
		var diags diag.Diagnostics
		req := m.toAPIUpsertRequest(ctx, &diags)
		require.False(t, diags.HasError())
		assert.Equal(t, streamTypeQuery, req.Stream.Type)
		assert.Nil(t, req.Stream.Ingest)
		require.NotNil(t, req.Stream.Query)
		assert.Equal(t, "FROM logs.nginx | WHERE http.response.status_code >= 400", req.Stream.Query.Esql)
	})

	t.Run("attached queries include all fields", func(t *testing.T) {
		t.Parallel()
		score := float64(70)
		m := streamModel{
			Name:        types.StringValue("logs.nginx"),
			SpaceID:     types.StringValue("default"),
			Description: types.StringValue(""),
			WiredConfig: &wiredConfigModel{
				ProcessingStepsJSON:   jsontypes.NewNormalizedNull(),
				FieldsJSON:            jsontypes.NewNormalizedNull(),
				RoutingJSON:           jsontypes.NewNormalizedNull(),
				LifecycleJSON:         jsontypes.NewNormalizedNull(),
				FailureStoreJSON:      jsontypes.NewNormalizedNull(),
				IndexNumberOfShards:   types.Int64Null(),
				IndexNumberOfReplicas: types.Int64Null(),
				IndexRefreshInterval:  types.StringNull(),
			},
			Dashboards: types.ListNull(types.StringType),
			Queries: []streamQueryModel{
				{
					ID:            types.StringValue("q1"),
					Title:         types.StringValue("High errors"),
					Description:   types.StringValue("Detects 5xx rates"),
					Esql:          types.StringValue("FROM logs.nginx | WHERE http.status >= 500"),
					SeverityScore: types.Float64Value(score),
					Evidence:      types.ListNull(types.StringType),
				},
			},
		}
		var diags diag.Diagnostics
		req := m.toAPIUpsertRequest(ctx, &diags)
		require.False(t, diags.HasError())
		require.Len(t, req.Queries, 1)
		q := req.Queries[0]
		assert.Equal(t, "q1", q.ID)
		assert.Equal(t, "High errors", q.Title)
		assert.Equal(t, "Detects 5xx rates", q.Description)
		assert.Equal(t, "FROM logs.nginx | WHERE http.status >= 500", q.Esql.Query)
		require.NotNil(t, q.SeverityScore)
		assert.InDelta(t, float32(70), *q.SeverityScore, 0.001)
	})
}
