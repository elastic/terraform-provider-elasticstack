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
	ProcessingSteps       types.List           `tfsdk:"processing_steps"`
	FieldOverridesJSON    jsontypes.Normalized `tfsdk:"field_overrides_json"`
	LifecycleJSON         jsontypes.Normalized `tfsdk:"lifecycle_json"`
	FailureStoreJSON      jsontypes.Normalized `tfsdk:"failure_store_json"`
	IndexNumberOfShards   types.Int64          `tfsdk:"index_number_of_shards"`
	IndexNumberOfReplicas types.Int64          `tfsdk:"index_number_of_replicas"`
	IndexRefreshInterval  types.String         `tfsdk:"index_refresh_interval"`
}

// populateFromAPI populates the classic config model from an API ingest response.
func (m *classicConfigModel) populateFromAPI(_ context.Context, ingest *kibanaoapi.StreamIngest) diag.Diagnostics {
	var diags diag.Diagnostics
	if ingest == nil {
		return diags
	}

	// Processing steps — each step is a JSON-encoded streamlang object.
	steps, stepsDiags := populateProcessingStepsFromAPI(ingest)
	diags.Append(stepsDiags...)
	if diags.HasError() {
		return diags
	}
	m.ProcessingSteps = steps

	// Classic-specific field overrides
	if ingest.Classic != nil && len(ingest.Classic.FieldOverrides) > 0 {
		m.FieldOverridesJSON = jsontypes.NewNormalizedValue(string(ingest.Classic.FieldOverrides))
	} else {
		m.FieldOverridesJSON = jsontypes.NewNormalizedNull()
	}

	m.LifecycleJSON, m.FailureStoreJSON = populateLifecycleAndFailureStoreFromAPI(ingest)
	m.IndexNumberOfShards, m.IndexNumberOfReplicas, m.IndexRefreshInterval = populateIndexSettingsFromAPI(ingest)

	return diags
}

// toAPIIngest converts the classic config model to an API ingest object.
func (m *classicConfigModel) toAPIIngest(diags *diag.Diagnostics) *kibanaoapi.StreamIngest {
	ingest := &kibanaoapi.StreamIngest{}

	// Processing steps — the API requires the array to be present (even empty).
	stepsJSON := processingStepsToAPI(m.ProcessingSteps, diags)
	if diags.HasError() {
		return ingest
	}
	ingest.Processing.Steps = stepsJSON

	// Classic field overrides — the API always requires the classic block.
	ingest.Classic = &kibanaoapi.StreamIngestClassic{}
	if typeutils.IsKnown(m.FieldOverridesJSON) {
		ingest.Classic.FieldOverrides = json.RawMessage(m.FieldOverridesJSON.ValueString())
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

	applyIndexSettingsToAPI(m.IndexNumberOfShards, m.IndexNumberOfReplicas, m.IndexRefreshInterval, ingest)

	return ingest
}
