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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func newIndexSettingsTestServer(t *testing.T, indexExists, putFails bool) (*clients.ElasticsearchScopedClient, *[]map[string]any) {
	t.Helper()

	indices := map[string]map[string]any{}
	if indexExists {
		indices["my-index"] = map[string]any{
			"index.number_of_shards": "1",
			"index.uuid":             "index-uuid",
		}
	}
	return newIndexSettingsServer(t, indices, putFails)
}

func TestCreateIndexSettings_IndexExistsSendsDeclaredSettings(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	plan := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	result, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, "test-cluster-uuid/"+indexName, result.Model.ID.ValueString())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.mapping.total_fields.limit": json.Number("5000"),
	}, (*putBodies)[0])
}

// REQ-002: an index-only create (no typed attribute, no settings_json) still
// verifies the target index exists and records the resource in state with the
// computed id, but issues no PUT /{index}/_settings call.
func TestCreateIndexSettings_IndexOnlyIssuesNoSettingsPut(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	plan := tfModel{
		Index: types.StringValue(indexName),
	}

	result, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, "test-cluster-uuid/"+indexName, result.Model.ID.ValueString())
	require.Empty(t, *putBodies, "index-only create must not issue a settings PUT")
}

// Settings_json numeric values must reach the wire payload as
// exact JSON tokens. Integers beyond float64 precision (2^53) must not be
// rounded, so distinct large-integer plans send distinct payloads.
func TestCreateIndexSettings_SettingsJSONNumbersSentWithExactTokens(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	plan := tfModel{
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.max_result_window": 9007199254740993, "index.max_terms_count": 1.5}`),
	}

	_, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.max_result_window": json.Number("9007199254740993"),
		"index.max_terms_count":   json.Number("1.5"),
	}, (*putBodies)[0], "numeric settings_json values must be sent with their exact JSON tokens")
}

// REQ-002: typed Set and settings_json array values serialize as real JSON
// arrays (not JSON-encoded strings) in the PUT /{index}/_settings payload.
func TestCreateIndexSettings_ArrayValuesSerializeAsJSONArrays(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	queryDefaultField, setDiags := types.SetValueFrom(context.Background(), types.StringType, []string{"title", "body"})
	require.False(t, setDiags.HasError(), "unexpected diagnostics: %v", setDiags.Errors())

	plan := tfModel{
		Index:             types.StringValue(indexName),
		QueryDefaultField: queryDefaultField,
		SettingsJSON:      jsontypes.NewNormalizedValue(`{"some.future.array.setting":["a","b"]}`),
	}

	_, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.query.default_field": []any{"title", "body"},
		"some.future.array.setting": []any{"a", "b"},
	}, (*putBodies)[0])
}

// REQ-003: changed array values serialize as real JSON arrays in the update
// payload, and a settings_json array key removed from the declaration is
// sent as null so Elasticsearch resets it.
func TestUpdateIndexSettings_ArrayValuesSerializeAsJSONArrays(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	priorQueryDefaultField, setDiags := types.SetValueFrom(context.Background(), types.StringType, []string{"title"})
	require.False(t, setDiags.HasError(), "unexpected diagnostics: %v", setDiags.Errors())

	prior := tfModel{
		Index:             types.StringValue(indexName),
		QueryDefaultField: priorQueryDefaultField,
		SettingsJSON:      jsontypes.NewNormalizedValue(`{"some.future.array.setting":["a"]}`),
	}

	planQueryDefaultField, planSetDiags := types.SetValueFrom(context.Background(), types.StringType, []string{"title", "body"})
	require.False(t, planSetDiags.HasError(), "unexpected diagnostics: %v", planSetDiags.Errors())

	plan := tfModel{
		Index:             types.StringValue(indexName),
		QueryDefaultField: planQueryDefaultField,
		SettingsJSON:      jsontypes.NewNormalizedValue(`{"other.future.array.setting":["x"]}`),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.query.default_field":  []any{"title", "body"},
		"other.future.array.setting": []any{"x"},
		"some.future.array.setting":  nil,
	}, (*putBodies)[0])
}

// The update diff compares exact numeric tokens after key
// canonicalization, so integers differing beyond float64 precision (2^53)
// trigger a settings update instead of comparing equal.
func TestUpdateIndexSettings_DistinctLargeIntegersAreNotEqual(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	prior := tfModel{
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.max_result_window": 9007199254740993}`),
	}
	plan := tfModel{
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"max_result_window": 9007199254740994}`),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1, "distinct integers beyond 2^53 must not compare equal in the canonical diff")
	require.Equal(t, map[string]any{
		"max_result_window": json.Number("9007199254740994"),
	}, (*putBodies)[0])
}

func TestUpdateIndexSettings_ChangedSettingSendsUpdate(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	prior := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(2300),
	}
	plan := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.mapping.total_fields.limit": json.Number("5000"),
	}, (*putBodies)[0])
}

// The canonical diff stays independent of `index.` prefix
// spellings and of the numeric source (typed Int64 attribute vs settings_json
// token, integers and decimals alike), so a semantically unchanged
// declaration never issues a PUT.
func TestUpdateIndexSettings_CanonicalEquivalenceSkipsAPICall(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	prior := tfModel{
		Index:           types.StringValue(indexName),
		MaxResultWindow: types.Int64Value(20000),
		SettingsJSON:    jsontypes.NewNormalizedValue(`{"index.max_terms_count": 1.5, "some.setting": "x"}`),
	}
	plan := tfModel{
		Index:           types.StringValue(indexName),
		MaxResultWindow: types.Int64Value(20000),
		SettingsJSON:    jsontypes.NewNormalizedValue(`{"max_terms_count": 1.5, "index.some.setting": "x"}`),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Empty(t, *putBodies, "prefix-spelling and numeric-source differences must not issue a PUT when the canonical values are unchanged")
}

func TestUpdateIndexSettings_UnchangedSkipsAPICall(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	model := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(2300),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  model,
		Prior: &model,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Empty(t, *putBodies, "no PUT /_settings call may be issued when the declared settings are unchanged")
}

func TestUpdateIndexSettings_APIErrorSurfacesDiagnostics(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, true)

	prior := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(2300),
	}
	plan := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	result, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.True(t, diags.HasError(), "expected an error diagnostic when PUT /_settings fails")
	require.Len(t, *putBodies, 1)
	require.Contains(t, diags.Errors()[0].Summary(), "closed index")
	require.Equal(t, plan, result.Model)
}

func TestUpdateIndexSettings_RemovedSettingsJSONKeySendsNull(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	prior := tfModel{
		Index:        types.StringValue(indexName),
		SettingsJSON: jsontypes.NewNormalizedValue(`{"index.refresh_interval": "5s"}`),
	}
	plan := tfModel{
		Index:            types.StringValue(indexName),
		NumberOfReplicas: types.Int64Value(2),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.number_of_replicas": json.Number("2"),
		"index.refresh_interval":   nil,
	}, (*putBodies)[0])
}

func TestUpdateIndexSettings_RemovedSettingSendsNull(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	prior := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}
	plan := tfModel{
		Index: types.StringValue(indexName),
	}

	_, diags := updateIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:  plan,
		Prior: &prior,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.mapping.total_fields.limit": nil,
	}, (*putBodies)[0])
}

func TestCreateIndexSettings_MergesSettingsJSONWithTypedAttributes(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, true, false)

	plan := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
		SettingsJSON:            jsontypes.NewNormalizedValue(`{"index.refresh_interval": "5s", "index.search.slowlog.threshold.query.warn": "10s"}`),
	}

	_, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, *putBodies, 1)
	require.Equal(t, map[string]any{
		"index.mapping.total_fields.limit":          json.Number("5000"),
		"index.refresh_interval":                    "5s",
		"index.search.slowlog.threshold.query.warn": "10s",
	}, (*putBodies)[0])
}

func TestCreateIndexSettings_IndexNotFound(t *testing.T) {
	t.Parallel()

	const indexName = "my-index"
	client, putBodies := newIndexSettingsTestServer(t, false, false)

	plan := tfModel{
		Index:                   types.StringValue(indexName),
		MappingTotalFieldsLimit: types.Int64Value(5000),
	}

	_, diags := createIndexSettings(context.Background(), client, entitycore.WriteRequest[tfModel]{Plan: plan})

	require.True(t, diags.HasError(), "expected an error diagnostic when the index does not exist")
	require.Contains(t, diags.Errors()[0].Detail(), indexName)
	require.Empty(t, *putBodies, "no PUT /_settings call may be issued when the index does not exist")
}
