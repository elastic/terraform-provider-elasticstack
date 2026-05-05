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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID, taskType string, endpoint *types.InferenceEndpoint) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Inference.Put(inferenceID).Request(endpoint)
	if taskType != "" {
		req.TaskType(taskType)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to create or update inference endpoint", err.Error())
		return diags
	}

	return diags
}

func GetInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID string) (*types.InferenceEndpointInfo, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Inference.Get().InferenceId(inferenceID).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil, nil
		}
		diags.AddError("Unable to get inference endpoint", err.Error())
		return nil, diags
	}

	if len(res.Endpoints) == 0 {
		return nil, nil
	}

	return &res.Endpoints[0], nil
}

func UpdateInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID, taskType string, update *types.InferenceEndpoint) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	// Build the update body manually, omitting Service which the API rejects as
	// an immutable field. The typed client's InferenceEndpoint always serializes
	// Service because the struct tag lacks omitempty.
	body := make(map[string]any)
	if len(update.ServiceSettings) > 0 {
		var ss map[string]any
		if err := json.Unmarshal(update.ServiceSettings, &ss); err != nil {
			diags.AddError("Unable to unmarshal service_settings", err.Error())
			return diags
		}
		body["service_settings"] = ss
	}
	if len(update.TaskSettings) > 0 {
		var ts map[string]any
		if err := json.Unmarshal(update.TaskSettings, &ts); err != nil {
			diags.AddError("Unable to unmarshal task_settings", err.Error())
			return diags
		}
		body["task_settings"] = ts
	}
	if update.ChunkingSettings != nil {
		b, err := json.Marshal(update.ChunkingSettings)
		if err != nil {
			diags.AddError("Unable to marshal chunking_settings", err.Error())
			return diags
		}
		var cs map[string]any
		if err := json.Unmarshal(b, &cs); err != nil {
			diags.AddError("Unable to unmarshal chunking_settings", err.Error())
			return diags
		}
		body["chunking_settings"] = cs
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		diags.AddError("Unable to marshal update body", err.Error())
		return diags
	}

	req := typedClient.Inference.Update(inferenceID).Raw(bytes.NewReader(jsonBody))
	if taskType != "" {
		req.TaskType(taskType)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to update inference endpoint", err.Error())
		return diags
	}

	return diags
}

func DeleteInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Inference.Delete(inferenceID).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete inference endpoint", err.Error())
		return diags
	}

	return diags
}
