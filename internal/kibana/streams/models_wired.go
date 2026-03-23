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
	ProcessingStepsJSON   jsontypes.Normalized `tfsdk:"processing_steps_json"`
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

	// Processing steps
	if len(ingest.Processing.Steps) > 0 {
		m.ProcessingStepsJSON = jsontypes.NewNormalizedValue(string(ingest.Processing.Steps))
	} else {
		m.ProcessingStepsJSON = jsontypes.NewNormalizedNull()
	}

	// Wired-specific fields and routing
	if ingest.Wired != nil {
		if len(ingest.Wired.Fields) > 0 {
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

	return diags
}

// toAPIIngest converts the wired config model to an API ingest object.
func (m *wiredConfigModel) toAPIIngest(diags *diag.Diagnostics) *kibanaoapi.StreamIngest {
	ingest := &kibanaoapi.StreamIngest{}

	// Processing steps
	if typeutils.IsKnown(m.ProcessingStepsJSON) {
		ingest.Processing.Steps = json.RawMessage(m.ProcessingStepsJSON.ValueString())
	}

	// Wired fields and routing
	ingest.Wired = &kibanaoapi.StreamIngestWired{}
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

	// Lifecycle
	if typeutils.IsKnown(m.LifecycleJSON) {
		ingest.Lifecycle = json.RawMessage(m.LifecycleJSON.ValueString())
	}

	// Failure store
	if typeutils.IsKnown(m.FailureStoreJSON) {
		ingest.FailureStore = json.RawMessage(m.FailureStoreJSON.ValueString())
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
