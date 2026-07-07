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

package data_source

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorstatus"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncstatus"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-010 runtime telemetry attributes must appear on the data source schema.
var req010RuntimeTelemetryAttrs = []string{
	"status",
	"last_seen",
	"last_synced",
	"last_sync_status",
	"last_indexed_document_count",
	"last_deleted_document_count",
	"last_sync_scheduled_at",
	"last_sync_error",
	"last_access_control_sync_status",
	"last_access_control_sync_error",
	"last_access_control_sync_scheduled_at",
	"last_incremental_sync_scheduled_at",
	"error",
	"filtering",
	"custom_scheduling",
	"configuration",
	"sync_cursor",
	"sync_now",
}

func TestDataSourceSchemaFactory_containsREQ010RuntimeTelemetryAttributes(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchemaFactory(context.Background())
	attrs := schema.GetAttributes()

	for _, name := range req010RuntimeTelemetryAttrs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, ok := attrs[name]
			require.True(t, ok, "data source schema missing REQ-010 attribute %q", name)
		})
	}
}

func TestMarshalConnectorJSONField(t *testing.T) {
	t.Parallel()

	t.Run("nil any", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("filtering", nil, &diags)
		require.True(t, got.IsNull())
		require.False(t, diags.HasError())
	})

	t.Run("typed nil map", func(t *testing.T) {
		t.Parallel()
		var m map[string]any
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("configuration", m, &diags)
		require.True(t, got.IsNull(), "typed nil map must not become NormalizedValue(\"null\")")
		require.False(t, diags.HasError())
	})

	t.Run("typed nil slice", func(t *testing.T) {
		t.Parallel()
		var s []estypes.FilteringConfig
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("filtering", s, &diags)
		require.True(t, got.IsNull())
		require.False(t, diags.HasError())
	})

	t.Run("non-empty map", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("configuration", map[string]int{"a": 1}, &diags)
		require.False(t, got.IsNull())
		require.Equal(t, `{"a":1}`, got.ValueString())
		require.False(t, diags.HasError())
	})

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("filtering", []int{1, 2}, &diags)
		require.False(t, got.IsNull())
		require.Equal(t, "[1,2]", got.ValueString())
		require.False(t, diags.HasError())
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		got := marshalConnectorJSONField("filtering", make(chan int), &diags)
		require.True(t, got.IsNull())
		require.True(t, diags.HasError())
	})
}

func TestMarshalConnectorRawJSONField(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     json.RawMessage
		wantNil bool
		wantVal string
		wantErr bool
	}{
		{name: "nil bytes", raw: nil, wantNil: true},
		{name: "empty bytes", raw: []byte{}, wantNil: true},
		{name: "literal null", raw: []byte("null"), wantNil: true},
		{name: "whitespace null", raw: []byte(" null "), wantNil: true},
		{name: "json string", raw: []byte(`"value"`), wantVal: `"value"`},
		{name: "json object", raw: []byte(`{"a":1}`), wantVal: `{"a":1}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			got := marshalConnectorRawJSONField("sync_cursor", tc.raw, &diags)
			if tc.wantNil {
				require.True(t, got.IsNull())
			} else {
				require.False(t, got.IsNull())
				require.Equal(t, tc.wantVal, got.ValueString())
			}
			require.Equal(t, tc.wantErr, diags.HasError())
		})
	}
}

func TestConnectorDateTimeToString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   estypes.DateTime
		wantNil bool
		want    string
	}{
		{name: "nil", input: nil, wantNil: true},
		{name: "zero millis", input: estypes.DateTime(float64(0)), wantNil: true},
		{name: "RFC3339", input: estypes.DateTime("2024-06-01T12:00:00Z"), want: "2024-06-01T12:00:00.000Z"},
		{name: "epoch millis", input: estypes.DateTime(float64(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC).UnixMilli())), want: "2024-06-01T12:00:00.000Z"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := connectorDateTimeToString(tc.input)
			if tc.wantNil {
				require.True(t, got.IsNull())
				return
			}
			require.False(t, got.IsNull())
			require.Equal(t, tc.want, got.ValueString())
		})
	}
}

func TestPopulateContentConnectorDataSourceFromAPI_populated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := clients.NewElasticsearchScopedClientForTest(nil, []string{"http://localhost:9200"})

	serviceType := "postgresql"
	name := "Test Connector"
	lastSyncErr := "sync failed"
	syncCompleted := syncstatus.SyncStatus{Name: "completed"}
	indexed := int64(42)
	deleted := int64(3)
	lastSeen := estypes.DateTime("2024-06-01T10:00:00Z")
	lastSynced := estypes.DateTime("2024-06-01T11:00:00Z")
	lastSyncScheduled := estypes.DateTime("2024-06-01T09:00:00Z")

	resp := &getconnector.Response{
		ServiceType:              &serviceType,
		Name:                     &name,
		IsNative:                 false,
		Status:                   connectorstatus.Connected,
		LastSeen:                 lastSeen,
		LastSynced:               lastSynced,
		LastSyncStatus:           &syncCompleted,
		LastIndexedDocumentCount: &indexed,
		LastDeletedDocumentCount: &deleted,
		LastSyncScheduledAt:      lastSyncScheduled,
		LastSyncError:            &lastSyncErr,
		Filtering: []estypes.FilteringConfig{{
			Active: estypes.FilteringRules{},
			Draft:  estypes.FilteringRules{},
		}},
		CustomScheduling: estypes.ConnectorCustomScheduling{
			"full": estypes.CustomScheduling{
				Enabled:  true,
				Interval: "0 0 * * *",
				Name:     "full",
			},
		},
		Configuration: estypes.ConnectorConfiguration{
			"host": {Value: json.RawMessage(`"localhost"`)},
		},
		SyncCursor: json.RawMessage(`{"page":1}`),
		SyncNow:    true,
		Pipeline: &estypes.IngestPipelineParams{
			Name:                 "search-default-ingestion",
			ExtractBinaryContent: true,
			ReduceWhitespace:     true,
			RunMlInference:       false,
		},
	}

	var diags diag.Diagnostics
	model := ContentConnectorDataSourceModel{
		ConnectorID: fwtypes.StringValue("music"),
	}
	populateContentConnectorDataSourceFromAPI(ctx, client, "music", resp, &model, &diags)

	require.False(t, model.Status.IsNull())
	assert.Equal(t, "connected", model.Status.ValueString())
	require.False(t, model.LastSeen.IsNull())
	require.False(t, model.LastSynced.IsNull())
	require.False(t, model.LastSyncStatus.IsNull())
	assert.Equal(t, "completed", model.LastSyncStatus.ValueString())
	require.False(t, model.LastIndexedDocumentCount.IsNull())
	assert.Equal(t, int64(42), model.LastIndexedDocumentCount.ValueInt64())
	require.False(t, model.LastDeletedDocumentCount.IsNull())
	assert.Equal(t, int64(3), model.LastDeletedDocumentCount.ValueInt64())
	require.False(t, model.LastSyncScheduledAt.IsNull())
	require.False(t, model.LastSyncError.IsNull())
	assert.Equal(t, lastSyncErr, model.LastSyncError.ValueString())

	require.False(t, model.Filtering.IsNull())
	assert.Contains(t, model.Filtering.ValueString(), `"active"`)
	require.False(t, model.CustomScheduling.IsNull())
	assert.Contains(t, model.CustomScheduling.ValueString(), `"full"`)
	require.False(t, model.Configuration.IsNull())
	assert.Contains(t, model.Configuration.ValueString(), `"host"`)
	require.False(t, model.SyncCursor.IsNull())
	assert.JSONEq(t, `{"page":1}`, model.SyncCursor.ValueString())
	assert.True(t, model.SyncNow.ValueBool())
}

func TestPopulateContentConnectorDataSourceFromAPI_sparseAPIValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := clients.NewElasticsearchScopedClientForTest(nil, []string{"http://localhost:9200"})

	resp := &getconnector.Response{
		Status:           connectorstatus.Created,
		IsNative:         false,
		Filtering:        nil,
		CustomScheduling: nil,
		Configuration:    nil,
		SyncCursor:       nil,
	}

	var diags diag.Diagnostics
	model := ContentConnectorDataSourceModel{
		ConnectorID: fwtypes.StringValue("sparse"),
	}
	populateContentConnectorDataSourceFromAPI(ctx, client, "sparse", resp, &model, &diags)

	assert.Equal(t, "created", model.Status.ValueString())
	assert.True(t, model.ServiceType.IsNull())
	assert.True(t, model.Name.IsNull())
	assert.True(t, model.LastSeen.IsNull())
	assert.True(t, model.LastSynced.IsNull())
	assert.True(t, model.LastSyncStatus.IsNull())
	assert.True(t, model.LastIndexedDocumentCount.IsNull())
	assert.True(t, model.LastSyncError.IsNull())
	assert.True(t, model.Filtering.IsNull(), "nil filtering must be Terraform null, not JSON literal null")
	assert.True(t, model.CustomScheduling.IsNull())
	assert.True(t, model.Configuration.IsNull())
	assert.True(t, model.SyncCursor.IsNull())
}
