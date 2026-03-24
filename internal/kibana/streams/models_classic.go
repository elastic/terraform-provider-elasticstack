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

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// classicConfigModel is the Terraform model for a classic stream's configuration.
type classicConfigModel struct {
	ProcessingSteps       []processingStepModel `tfsdk:"processing_steps"`
	FieldOverridesJSON    jsontypes.Normalized  `tfsdk:"field_overrides_json"`
	LifecycleJSON         jsontypes.Normalized  `tfsdk:"lifecycle_json"`
	FailureStoreJSON      jsontypes.Normalized  `tfsdk:"failure_store_json"`
	IndexNumberOfShards   types.Int64           `tfsdk:"index_number_of_shards"`
	IndexNumberOfReplicas types.Int64           `tfsdk:"index_number_of_replicas"`
	IndexRefreshInterval  types.String          `tfsdk:"index_refresh_interval"`
}

// populateFromAPI populates the classic config model from an API ingest response.
func (m *classicConfigModel) populateFromAPI(_ context.Context, ingest *kibanaoapi.StreamIngest) {
	if ingest == nil {
		return
	}

	// Processing steps — split into individual step models for per-step plan diffs.
	// An empty array from the API is treated as null (no steps configured).
	if len(ingest.Processing.Steps) > 0 {
		var rawSteps []json.RawMessage
		if err := json.Unmarshal(ingest.Processing.Steps, &rawSteps); err != nil {
			m.ProcessingSteps = nil
		} else if len(rawSteps) > 0 {
			steps := make([]processingStepModel, len(rawSteps))
			for i, raw := range rawSteps {
				steps[i] = processingStepModel{JSON: jsontypes.NewNormalizedValue(string(raw))}
			}
			m.ProcessingSteps = steps
		} else {
			m.ProcessingSteps = nil
		}
	} else {
		m.ProcessingSteps = nil
	}

	// Classic-specific field overrides
	if ingest.Classic != nil && len(ingest.Classic.FieldOverrides) > 0 {
		m.FieldOverridesJSON = jsontypes.NewNormalizedValue(string(ingest.Classic.FieldOverrides))
	} else {
		m.FieldOverridesJSON = jsontypes.NewNormalizedNull()
	}

	// Lifecycle
	if len(ingest.Lifecycle) > 0 {
		m.LifecycleJSON = jsontypes.NewNormalizedValue(string(ingest.Lifecycle))
	} else {
		m.LifecycleJSON = jsontypes.NewNormalizedNull()
	}

	// Failure store
	if len(ingest.FailureStore) > 0 {
		m.FailureStoreJSON = jsontypes.NewNormalizedValue(string(ingest.FailureStore))
	} else {
		m.FailureStoreJSON = jsontypes.NewNormalizedNull()
	}

	// Index settings
	if ingest.Settings.IndexNumberOfShards != nil {
		if v, ok := ingest.Settings.IndexNumberOfShards.Value.(float64); ok {
			m.IndexNumberOfShards = types.Int64Value(int64(v))
		} else {
			m.IndexNumberOfShards = types.Int64Null()
		}
	} else {
		m.IndexNumberOfShards = types.Int64Null()
	}

	if ingest.Settings.IndexNumberOfReplicas != nil {
		if v, ok := ingest.Settings.IndexNumberOfReplicas.Value.(float64); ok {
			m.IndexNumberOfReplicas = types.Int64Value(int64(v))
		} else {
			m.IndexNumberOfReplicas = types.Int64Null()
		}
	} else {
		m.IndexNumberOfReplicas = types.Int64Null()
	}

	if ingest.Settings.IndexRefreshInterval != nil {
		switch v := ingest.Settings.IndexRefreshInterval.Value.(type) {
		case string:
			m.IndexRefreshInterval = types.StringValue(v)
		default:
			m.IndexRefreshInterval = types.StringNull()
		}
	} else {
		m.IndexRefreshInterval = types.StringNull()
	}
}

// toAPIIngest converts the classic config model to an API ingest object.
func (m *classicConfigModel) toAPIIngest(diags *diag.Diagnostics) *kibanaoapi.StreamIngest {
	ingest := &kibanaoapi.StreamIngest{}

	// Processing steps — the API requires the array to be present (even empty).
	rawSteps := make([]json.RawMessage, 0, len(m.ProcessingSteps))
	for _, step := range m.ProcessingSteps {
		if typeutils.IsKnown(step.JSON) {
			rawSteps = append(rawSteps, json.RawMessage(step.JSON.ValueString()))
		}
	}
	stepsJSON, err := json.Marshal(rawSteps)
	if err != nil {
		diags.AddError("Failed to marshal processing steps", err.Error())
		return ingest
	}
	ingest.Processing.Steps = stepsJSON

	// Classic field overrides
	if typeutils.IsKnown(m.FieldOverridesJSON) {
		ingest.Classic = &kibanaoapi.StreamIngestClassic{
			FieldOverrides: json.RawMessage(m.FieldOverridesJSON.ValueString()),
		}
	}

	// Lifecycle — required by API; default to inherit when not configured.
	if typeutils.IsKnown(m.LifecycleJSON) {
		ingest.Lifecycle = json.RawMessage(m.LifecycleJSON.ValueString())
	} else {
		ingest.Lifecycle = json.RawMessage(`{"inherit":{}}`)
	}

	// Failure store — required by API; default to disabled when not configured.
	if typeutils.IsKnown(m.FailureStoreJSON) {
		ingest.FailureStore = json.RawMessage(m.FailureStoreJSON.ValueString())
	} else {
		ingest.FailureStore = json.RawMessage(`{"disabled":{}}`)
	}

	// Index settings
	if typeutils.IsKnown(m.IndexNumberOfShards) {
		ingest.Settings.IndexNumberOfShards = &kibanaoapi.StreamIngestSettingValue{
			Value: float64(m.IndexNumberOfShards.ValueInt64()),
		}
	}
	if typeutils.IsKnown(m.IndexNumberOfReplicas) {
		ingest.Settings.IndexNumberOfReplicas = &kibanaoapi.StreamIngestSettingValue{
			Value: float64(m.IndexNumberOfReplicas.ValueInt64()),
		}
	}
	if typeutils.IsKnown(m.IndexRefreshInterval) {
		ingest.Settings.IndexRefreshInterval = &kibanaoapi.StreamIngestSettingValue{
			Value: m.IndexRefreshInterval.ValueString(),
		}
	}

	return ingest
}
