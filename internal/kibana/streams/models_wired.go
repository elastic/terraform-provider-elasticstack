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

// wiredConfigModel is the Terraform model for a wired stream's configuration.
type wiredConfigModel struct {
	ProcessingSteps       types.List           `tfsdk:"processing_steps"`
	FieldsJSON            jsontypes.Normalized `tfsdk:"fields_json"`
	RoutingJSON           jsontypes.Normalized `tfsdk:"routing_json"`
	LifecycleJSON         jsontypes.Normalized `tfsdk:"lifecycle_json"`
	FailureStoreJSON      jsontypes.Normalized `tfsdk:"failure_store_json"`
	IndexNumberOfShards   types.Int64          `tfsdk:"index_number_of_shards"`
	IndexNumberOfReplicas types.Int64          `tfsdk:"index_number_of_replicas"`
	IndexRefreshInterval  types.String         `tfsdk:"index_refresh_interval"`
}

// populateFromAPI populates the wired config model from an API ingest response.
func (m *wiredConfigModel) populateFromAPI(_ context.Context, ingest *kibanaoapi.StreamIngest) diag.Diagnostics {
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

	// Wired-specific fields and routing.
	// An empty object {} from the API means no fields are configured → null.
	if ingest.Wired != nil {
		if len(ingest.Wired.Fields) > 0 && string(ingest.Wired.Fields) != "{}" {
			m.FieldsJSON = jsontypes.NewNormalizedValue(string(ingest.Wired.Fields))
		} else {
			m.FieldsJSON = jsontypes.NewNormalizedNull()
		}

		if len(ingest.Wired.Routing) > 0 {
			routingBytes, err := json.Marshal(ingest.Wired.Routing)
			if err != nil {
				diags.AddError("Failed to marshal routing", err.Error())
				return diags
			}
			m.RoutingJSON = jsontypes.NewNormalizedValue(string(routingBytes))
		} else {
			m.RoutingJSON = jsontypes.NewNormalizedNull()
		}
	} else {
		m.FieldsJSON = jsontypes.NewNormalizedNull()
		m.RoutingJSON = jsontypes.NewNormalizedNull()
	}

	m.LifecycleJSON, m.FailureStoreJSON = populateLifecycleAndFailureStoreFromAPI(ingest)
	m.IndexNumberOfShards, m.IndexNumberOfReplicas, m.IndexRefreshInterval = populateIndexSettingsFromAPI(ingest)

	return diags
}

// toAPIIngest converts the wired config model to an API ingest object.
func (m *wiredConfigModel) toAPIIngest(diags *diag.Diagnostics) *kibanaoapi.StreamIngest {
	ingest := &kibanaoapi.StreamIngest{}

	// Processing steps — the API requires the array to be present (even empty).
	stepsJSON := processingStepsToAPI(m.ProcessingSteps, diags)
	if diags.HasError() {
		return ingest
	}
	ingest.Processing.Steps = stepsJSON

	// Wired fields and routing — the API requires the wired block to be present.
	ingest.Wired = &kibanaoapi.StreamIngestWired{
		Fields:  json.RawMessage(`{}`),
		Routing: []kibanaoapi.StreamRoutingRule{},
	}
	if typeutils.IsKnown(m.FieldsJSON) {
		ingest.Wired.Fields = json.RawMessage(m.FieldsJSON.ValueString())
	}
	if typeutils.IsKnown(m.RoutingJSON) {
		var routing []kibanaoapi.StreamRoutingRule
		if err := json.Unmarshal([]byte(m.RoutingJSON.ValueString()), &routing); err != nil {
			diags.AddError("Failed to parse routing_json", err.Error())
			return ingest
		}
		ingest.Wired.Routing = routing
	}

	// Lifecycle — required by API; default to inherit when not configured.
	if typeutils.IsKnown(m.LifecycleJSON) {
		ingest.Lifecycle = json.RawMessage(m.LifecycleJSON.ValueString())
	} else {
		ingest.Lifecycle = json.RawMessage(`{"inherit":{}}`)
	}

	// Failure store — required by API; default to disabled when not configured.
	// {inherit:{}} would cause "all ancestors have inherit configuration" errors
	// when the parent stream has not been explicitly configured.
	if typeutils.IsKnown(m.FailureStoreJSON) {
		ingest.FailureStore = json.RawMessage(m.FailureStoreJSON.ValueString())
	} else {
		ingest.FailureStore = json.RawMessage(`{"disabled":{}}`)
	}

	applyIndexSettingsToAPI(m.IndexNumberOfShards, m.IndexNumberOfReplicas, m.IndexRefreshInterval, ingest)

	return ingest
}
