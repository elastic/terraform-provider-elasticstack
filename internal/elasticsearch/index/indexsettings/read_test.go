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

package indexsettings

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func newIndexSettingsReadTestServer(t *testing.T, indexExists bool, flatSettings map[string]any) *clients.ElasticsearchScopedClient {
	t.Helper()

	indices := map[string]map[string]any{}
	if indexExists {
		indices["my-index"] = flatSettings
	}
	client, _ := newIndexSettingsServer(t, indices, false)
	return client
}

func TestReadIndexSettings_UnrelatedSettingsIgnored(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.number_of_shards":           "1",
		"index.uuid":                       "index-uuid",
		"index.number_of_replicas":         "2",
		"index.mapping.total_fields.limit": "5000",
	})

	state := tfModel{
		ID:                      types.StringValue("test-cluster-uuid/" + indexName),
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)
	require.Equal(t, types.Int64Value(5000), read.MappingTotalFieldsLimit)
	require.True(t, read.NumberOfReplicas.IsNull(), "unrelated setting must not be adopted into state")
	require.True(t, read.SettingsJSON.IsNull(), "settings_json must stay unset when not declared in state")
}
func TestReadIndexSettings_OutOfBandChangeSurfacesDrift(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.number_of_shards":           "1",
		"index.mapping.total_fields.limit": "2300",
	})

	state := tfModel{
		ID:                      types.StringValue("test-cluster-uuid/" + indexName),
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)
	require.Equal(t, types.Int64Value(2300), read.MappingTotalFieldsLimit, "out-of-band change must surface in state so the plan proposes restoring 5000")
}

// fakeImportPrivateState implements entitycore.PrivateStateStorage, keyed only
// on the import-hydration flag.
type fakeImportPrivateState struct {
	importFlag []byte
}

func (f *fakeImportPrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if key == importHydrationPrivateKey {
		return f.importFlag, nil
	}
	return nil, nil
}

func (f *fakeImportPrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if key == importHydrationPrivateKey {
		f.importFlag = value
	}
	return nil
}

func TestPostReadIndexSettings_ImportHydratesOnceViaFlag(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.number_of_shards":           "1",
		"index.uuid":                       "index-uuid",
		"index.number_of_replicas":         "2",
		"index.mapping.total_fields.limit": "5000",
		"index.refresh_interval":           "10s",
	})

	private := &fakeImportPrivateState{importFlag: []byte("true")}
	state := tfModel{
		ID:    types.StringValue("test-cluster-uuid/" + indexName),
		Index: types.StringValue(indexName),
	}
	req := entitycore.ElasticsearchPostReadRequest[tfModel]{
		Client:  client,
		Prior:   state,
		State:   state,
		Private: private,
	}

	hydrated, diags := postReadIndexSettings(context.Background(), req)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, types.Int64Value(5000), hydrated.MappingTotalFieldsLimit)
	require.Equal(t, types.Int64Value(2), hydrated.NumberOfReplicas)
	require.Equal(t, types.StringValue("10s"), hydrated.RefreshInterval)
	require.True(t, hydrated.SettingsJSON.IsNull(), "import hydration must leave settings_json unset")
	require.Nil(t, private.importFlag, "the import flag must be cleared so later reads do not hydrate")

	second, diags := postReadIndexSettings(context.Background(), entitycore.ElasticsearchPostReadRequest[tfModel]{
		Client:  client,
		Prior:   hydrated,
		State:   hydrated,
		Private: private,
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, hydrated, second, "later reads must not re-hydrate once the flag is cleared")
}

func TestReadIndexSettings_ResetSettingNulledThenNotAdopted(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.number_of_shards":           "1",
		"index.mapping.total_fields.limit": "5000",
	})

	state := tfModel{
		ID:               types.StringValue("test-cluster-uuid/" + indexName),
		Index:            types.StringValue(indexName),
		NumberOfReplicas: types.Int64Value(2),
	}

	first, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)
	require.True(t, first.NumberOfReplicas.IsNull(), "externally-reset tracked setting must be nulled so drift shows")

	second, found, diags := readIndexSettings(context.Background(), client, indexName, first)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)
	require.True(t, second.NumberOfReplicas.IsNull())
	require.True(t, second.MappingTotalFieldsLimit.IsNull(), "an unrelated setting must not be adopted once the tracked set is empty")
}

func TestReadIndexSettings_SettingsJSONScalarsRoundTrip(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.max_result_window": "20000",
		"index.blocks.read_only":  "false",
		"index.refresh_interval":  "10s",
	})

	state := tfModel{
		ID:           types.StringValue("test-cluster-uuid/" + indexName),
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.max_result_window":20000,"index.blocks.read_only":false,"index.refresh_interval":"10s"}`),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)

	var stateSettings, readSettings map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(state.SettingsJSON.ValueString()), &stateSettings))
	require.NoError(t, json.Unmarshal([]byte(read.SettingsJSON.ValueString()), &readSettings))
	require.Equal(t, stateSettings, readSettings, "reconciled settings_json keys must round-trip without false drift")
}

// REQ-004: array-valued settings_json keys round-trip without drift when
// Elasticsearch returns the array as a JSON-encoded string in the flat
// settings response; element order is preserved and a tracked array is
// never silently dropped, even when it is empty.
func TestReadIndexSettings_SettingsJSONEncodedArrayRoundTrip(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.query.default_field": `["title","body"]`,
		"some.future.array.setting": `[]`,
		"some.number.array":         `[1,2]`,
		"some.bool.array":           `[true]`,
	})

	state := tfModel{
		ID:           types.StringValue("test-cluster-uuid/" + indexName),
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.query.default_field":["title","body"],"some.future.array.setting":[],"some.number.array":[1,2],"some.bool.array":[true]}`),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)

	var readSettings map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(read.SettingsJSON.ValueString()), &readSettings))
	require.JSONEq(t, `["title","body"]`, string(readSettings["index.query.default_field"]), "tracked string array must round-trip with element order preserved")
	require.JSONEq(t, `[]`, string(readSettings["some.future.array.setting"]), "tracked empty array must round-trip")
	require.JSONEq(t, `[1,2]`, string(readSettings["some.number.array"]), "tracked number array must round-trip")
	require.JSONEq(t, `[true]`, string(readSettings["some.bool.array"]), "tracked bool array must round-trip")
}

// Numeric settings_json values reconcile with their exact JSON tokens, so integers beyond float64 precision (2^53) round-trip from
// string API values without false drift, as scalars and inside arrays.
func TestReadIndexSettings_SettingsJSONExactNumbersRoundTrip(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.max_result_window": "9007199254740993",
		"some.number.array":       `[9007199254740993,2]`,
	})

	state := tfModel{
		ID:           types.StringValue("test-cluster-uuid/" + indexName),
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.max_result_window":9007199254740993,"some.number.array":[9007199254740993,2]}`),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)

	var readSettings map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(read.SettingsJSON.ValueString()), &readSettings))
	require.Equal(t, `9007199254740993`, string(readSettings["index.max_result_window"]), "exact integer token must be preserved, not rounded through float64")
	require.Equal(t, `[9007199254740993,2]`, string(readSettings["some.number.array"]), "exact integer tokens inside arrays must be preserved, not rounded through float64")
}

// REQ-004: array-valued settings_json keys arriving as real JSON arrays in
// the API response reconcile against the declared element types; a changed
// array surfaces as drift taken from the API, preserving element order.
func TestReadIndexSettings_SettingsJSONRealArrayReconciles(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.query.default_field": []any{"body", "title"},
	})

	state := tfModel{
		ID:           types.StringValue("test-cluster-uuid/" + indexName),
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.query.default_field":["title","body"]}`),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)

	var readSettings map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(read.SettingsJSON.ValueString()), &readSettings))
	require.JSONEq(t, `["body","title"]`, string(readSettings["index.query.default_field"]), "API array order must be preserved as drift")
}

// REQ-004: the typed query_default_field attribute decodes a JSON-encoded
// string array from the flat API response, mirroring the index resource's
// stringSliceFromAny decoding, so an unchanged typed set shows no drift.
func TestReadIndexSettings_TypedQueryDefaultFieldDecodesEncodedArray(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.query.default_field": `["title","body"]`,
	})

	state := tfModel{
		ID:    types.StringValue("test-cluster-uuid/" + indexName),
		Index: types.StringValue(indexName),
	}
	stateSet, setDiags := types.SetValueFrom(context.Background(), types.StringType, []string{"title", "body"})
	require.False(t, setDiags.HasError(), "unexpected diagnostics: %v", setDiags.Errors())
	state.QueryDefaultField = stateSet

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, found)
	require.Equal(t, state.QueryDefaultField, read.QueryDefaultField, "encoded string array must hydrate the typed set without drift")
}

// An empty typed query_default_field set is valid and must
// round-trip as an empty set, never nulled, whether Elasticsearch reports it
// as a JSON-encoded array string or as a real JSON array.
func TestReadIndexSettings_TypedQueryDefaultFieldEmptyArrayStaysEmpty(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	for _, apiValue := range []any{`[]`, []any{}} {
		client := newIndexSettingsReadTestServer(t, true, map[string]any{
			"index.query.default_field": apiValue,
		})

		state := tfModel{
			ID:    types.StringValue("test-cluster-uuid/" + indexName),
			Index: types.StringValue(indexName),
		}
		emptySet, setDiags := types.SetValueFrom(context.Background(), types.StringType, []string{})
		require.False(t, setDiags.HasError(), "unexpected diagnostics: %v", setDiags.Errors())
		state.QueryDefaultField = emptySet

		read, found, diags := readIndexSettings(context.Background(), client, indexName, state)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
		require.True(t, found)
		require.Equal(t, emptySet, read.QueryDefaultField, "an empty tracked array must stay an empty set, not be nulled (%v)", apiValue)
	}
}

// Import hydration populates the typed query_default_field
// from an empty JSON-encoded array as an empty set, not as null, so a narrowed
// configuration declaring an empty tracked array converges without drift.
func TestPostReadIndexSettings_ImportHydratesEmptyTypedArray(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, true, map[string]any{
		"index.query.default_field": `[]`,
	})

	private := &fakeImportPrivateState{importFlag: []byte("true")}
	state := tfModel{
		ID:    types.StringValue("test-cluster-uuid/" + indexName),
		Index: types.StringValue(indexName),
	}
	req := entitycore.ElasticsearchPostReadRequest[tfModel]{
		Client:  client,
		Prior:   state,
		State:   state,
		Private: private,
	}

	hydrated, diags := postReadIndexSettings(context.Background(), req)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())

	emptySet, setDiags := types.SetValueFrom(context.Background(), types.StringType, []string{})
	require.False(t, setDiags.HasError(), "unexpected diagnostics: %v", setDiags.Errors())
	require.Equal(t, emptySet, hydrated.QueryDefaultField, "an empty encoded array must hydrate as an empty set, not stay null")
}

func TestReadIndexSettings_NotFoundReportsNotFound(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client := newIndexSettingsReadTestServer(t, false, nil)

	state := tfModel{
		ID:               types.StringValue("test-cluster-uuid/" + indexName),
		Index:            types.StringValue(indexName),
		NumberOfReplicas: types.Int64Value(2),
	}

	read, found, diags := readIndexSettings(context.Background(), client, indexName, state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.False(t, found, "a missing index must report not-found so the envelope removes the resource from state")
	require.Equal(t, state, read)
}

func TestImportState_ParsesCompositeIDIntoIDAndIndex(t *testing.T) {
	t.Parallel()

	res, ok := NewIndexSettingsResource().(resource.ResourceWithImportState)
	require.True(t, ok, "the resource must implement ResourceWithImportState")

	ctx := context.Background()
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(ctx, resource.ImportStateRequest{ID: "test-cluster-uuid/my-index"}, &imported)

	require.False(t, imported.Diagnostics.HasError(), "unexpected diagnostics: %v", imported.Diagnostics.Errors())

	var id, index types.String
	require.False(t, imported.State.GetAttribute(ctx, path.Root("id"), &id).HasError(), "failed to read imported id")
	require.False(t, imported.State.GetAttribute(ctx, path.Root("index"), &index).HasError(), "failed to read imported index")
	require.Equal(t, "test-cluster-uuid/my-index", id.ValueString())
	require.Equal(t, "my-index", index.ValueString())
}

func TestImportState_RejectsNonCompositeID(t *testing.T) {
	t.Parallel()

	res, ok := NewIndexSettingsResource().(resource.ResourceWithImportState)
	require.True(t, ok, "the resource must implement ResourceWithImportState")

	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "not-a-composite-id"}, &imported)

	require.True(t, imported.Diagnostics.HasError(), "a non-composite import ID must be rejected")
}

func newIndexSettingsMultiIndexReadTestServer(t *testing.T, indices map[string]map[string]any) *clients.ElasticsearchScopedClient {
	t.Helper()

	client, _ := newIndexSettingsServer(t, indices, false)
	return client
}

// REQ-006 for_each over multiple concrete indices: each instance manages its
// own index only. Fast-feedback coverage for the per-instance independence
// guarantees; the apply-level for_each scenario is covered end-to-end by
// TestAccResourceIndexSettings_forEachIndependently.
func TestReadIndexSettings_ForEachInstancesAreIndependent(t *testing.T) {
	t.Parallel()

	client := newIndexSettingsMultiIndexReadTestServer(t, map[string]map[string]any{
		"index-a": {
			"index.number_of_shards":           "1",
			"index.number_of_replicas":         "3",
			"index.mapping.total_fields.limit": "5000",
			"index.refresh_interval":           "10s",
		},
		"index-b": {
			"index.number_of_shards":           "1",
			"index.number_of_replicas":         "2",
			"index.mapping.total_fields.limit": "2000",
			"index.refresh_interval":           "30s",
		},
	})

	instanceA := tfModel{
		ID:                      types.StringValue("test-cluster-uuid/index-a"),
		Index:                   types.StringValue("index-a"),
		MappingTotalFieldsLimit: types.Int64Value(5000),
		SettingsJSON:            jsontypes.NewNormalizedValue(`{"index.refresh_interval":"10s"}`),
	}
	instanceB := tfModel{
		ID:                      types.StringValue("test-cluster-uuid/index-b"),
		Index:                   types.StringValue("index-b"),
		MappingTotalFieldsLimit: types.Int64Value(2000),
		SettingsJSON:            jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`),
	}

	readA, foundA, diagsA := readIndexSettings(context.Background(), client, "index-a", instanceA)
	require.False(t, diagsA.HasError(), "instance A read: %v", diagsA.Errors())
	require.True(t, foundA)

	readB, foundB, diagsB := readIndexSettings(context.Background(), client, "index-b", instanceB)
	require.False(t, diagsB.HasError(), "instance B read: %v", diagsB.Errors())
	require.True(t, foundB)

	// Each instance refreshes from its own index's settings only.
	require.Equal(t, types.Int64Value(5000), readA.MappingTotalFieldsLimit)
	require.Equal(t, types.Int64Value(2000), readB.MappingTotalFieldsLimit)
	require.JSONEq(t, `{"index.refresh_interval":"10s"}`, readA.SettingsJSON.ValueString())
	require.JSONEq(t, `{"index.refresh_interval":"30s"}`, readB.SettingsJSON.ValueString())

	// The per-instance independence guarantee: settings declared only by the
	// other index's instance must not bleed into this instance's state.
	require.True(t, readA.NumberOfReplicas.IsNull(), "instance A does not declare number_of_replicas")
	require.True(t, readB.NumberOfReplicas.IsNull(), "instance B does not declare number_of_replicas")
	require.Equal(t, "index-a", readA.Index.ValueString())
	require.Equal(t, "index-b", readB.Index.ValueString())

	// A write (update) on one instance must not send any change for the other:
	// reading back after the update keeps each instance scoped to its index.
	updatedA := instanceA
	updatedA.MappingTotalFieldsLimit = types.Int64Value(6000)
	updateResult, writeDiags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:    updatedA,
		Prior:   &instanceA,
		Config:  updatedA,
		WriteID: "test-cluster-uuid/index-a",
	})
	require.False(t, writeDiags.HasError(), "instance A update: %v", writeDiags.Errors())
	require.Equal(t, updatedA, updateResult.Model)

	readBack, found, diags := readIndexSettings(context.Background(), client, "index-a", updatedA)
	require.False(t, diags.HasError(), "instance A re-read: %v", diags.Errors())
	require.True(t, found)
	require.Equal(t, types.Int64Value(6000), readBack.MappingTotalFieldsLimit)
}
