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

package inferenceendpoint

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Data struct {
	ID               types.String         `tfsdk:"id"`
	InferenceID      types.String         `tfsdk:"inference_id"`
	TaskType         types.String         `tfsdk:"task_type"`
	Service          types.String         `tfsdk:"service"`
	ServiceSettings  jsontypes.Normalized `tfsdk:"service_settings"`
	TaskSettings     jsontypes.Normalized `tfsdk:"task_settings"`
	ChunkingSettings jsontypes.Normalized `tfsdk:"chunking_settings"`
}

func (data *Data) toAPIModel(_ context.Context) (*elasticsearch.InferenceEndpoint, diag.Diagnostics) {
	var diags diag.Diagnostics

	endpoint := &elasticsearch.InferenceEndpoint{
		InferenceID: data.InferenceID.ValueString(),
		TaskType:    data.TaskType.ValueString(),
		Service:     data.Service.ValueString(),
	}

	if !data.ServiceSettings.IsNull() && !data.ServiceSettings.IsUnknown() {
		var ss map[string]any
		if err := json.Unmarshal([]byte(data.ServiceSettings.ValueString()), &ss); err != nil {
			diags.AddError("Invalid service_settings JSON", fmt.Sprintf("Error parsing service_settings: %s", err))
			return nil, diags
		}
		endpoint.ServiceSettings = ss
	}

	if !data.TaskSettings.IsNull() && !data.TaskSettings.IsUnknown() {
		var ts map[string]any
		if err := json.Unmarshal([]byte(data.TaskSettings.ValueString()), &ts); err != nil {
			diags.AddError("Invalid task_settings JSON", fmt.Sprintf("Error parsing task_settings: %s", err))
			return nil, diags
		}
		endpoint.TaskSettings = ts
	}

	if !data.ChunkingSettings.IsNull() && !data.ChunkingSettings.IsUnknown() {
		var cs map[string]any
		if err := json.Unmarshal([]byte(data.ChunkingSettings.ValueString()), &cs); err != nil {
			diags.AddError("Invalid chunking_settings JSON", fmt.Sprintf("Error parsing chunking_settings: %s", err))
			return nil, diags
		}
		endpoint.ChunkingSettings = cs
	}

	return endpoint, diags
}

func (data *Data) toUpdateModel(_ context.Context) (*elasticsearch.InferenceEndpointUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics

	update := &elasticsearch.InferenceEndpointUpdate{
		InferenceID: data.InferenceID.ValueString(),
		TaskType:    data.TaskType.ValueString(),
	}

	if !data.ServiceSettings.IsNull() && !data.ServiceSettings.IsUnknown() {
		var ss map[string]any
		if err := json.Unmarshal([]byte(data.ServiceSettings.ValueString()), &ss); err != nil {
			diags.AddError("Invalid service_settings JSON", fmt.Sprintf("Error parsing service_settings: %s", err))
			return nil, diags
		}
		update.ServiceSettings = ss
	}

	if !data.TaskSettings.IsNull() && !data.TaskSettings.IsUnknown() {
		var ts map[string]any
		if err := json.Unmarshal([]byte(data.TaskSettings.ValueString()), &ts); err != nil {
			diags.AddError("Invalid task_settings JSON", fmt.Sprintf("Error parsing task_settings: %s", err))
			return nil, diags
		}
		update.TaskSettings = ts
	}

	return update, diags
}

func (data *Data) fromAPIModel(_ context.Context, endpoint *elasticsearch.InferenceEndpoint) diag.Diagnostics {
	var diags diag.Diagnostics

	data.InferenceID = types.StringValue(endpoint.InferenceID)
	data.TaskType = types.StringValue(endpoint.TaskType)
	data.Service = types.StringValue(endpoint.Service)

	// service_settings: preserve plan value since the API may omit sensitive fields (e.g. api_key)
	// We only update if the field was previously null/unknown (first read after import).
	if data.ServiceSettings.IsNull() || data.ServiceSettings.IsUnknown() {
		if endpoint.ServiceSettings != nil {
			b, err := json.Marshal(endpoint.ServiceSettings)
			if err != nil {
				diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling service_settings: %s", err))
				return diags
			}
			data.ServiceSettings = jsontypes.NewNormalizedValue(string(b))
		}
	}

	// task_settings and chunking_settings: only populate from the API if the
	// user explicitly configured them. ES returns defaults for these fields
	// even when not set by the user, which would cause a persistent diff.
	if !data.TaskSettings.IsNull() && !data.TaskSettings.IsUnknown() {
		if endpoint.TaskSettings != nil {
			b, err := json.Marshal(endpoint.TaskSettings)
			if err != nil {
				diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling task_settings: %s", err))
				return diags
			}
			data.TaskSettings = jsontypes.NewNormalizedValue(string(b))
		} else {
			data.TaskSettings = jsontypes.NewNormalizedNull()
		}
	}

	if !data.ChunkingSettings.IsNull() && !data.ChunkingSettings.IsUnknown() {
		if endpoint.ChunkingSettings != nil {
			b, err := json.Marshal(endpoint.ChunkingSettings)
			if err != nil {
				diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling chunking_settings: %s", err))
				return diags
			}
			data.ChunkingSettings = jsontypes.NewNormalizedValue(string(b))
		} else {
			data.ChunkingSettings = jsontypes.NewNormalizedNull()
		}
	}

	return diags
}
