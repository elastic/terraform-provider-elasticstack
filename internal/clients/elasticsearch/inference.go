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
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// InferenceEndpoint is the model for the inference endpoint API.
type InferenceEndpoint struct {
	InferenceID      string         `json:"inference_id"`
	TaskType         string         `json:"task_type"`
	Service          string         `json:"service"`
	ServiceSettings  map[string]any `json:"service_settings"`
	TaskSettings     map[string]any `json:"task_settings,omitempty"`
	ChunkingSettings map[string]any `json:"chunking_settings,omitempty"`
}

// inferenceEndpointCreateRequest is the request body for create (PUT /_inference/...).
type inferenceEndpointCreateRequest struct {
	Service          string         `json:"service"`
	ServiceSettings  map[string]any `json:"service_settings"`
	TaskSettings     map[string]any `json:"task_settings,omitempty"`
	ChunkingSettings map[string]any `json:"chunking_settings,omitempty"`
}

// InferenceEndpointUpdate holds the fields we send to the update API
// (PUT /_inference/{task_type}/{inference_id}/_update).
// See https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-inference-update
type InferenceEndpointUpdate struct {
	InferenceID      string         `json:"inference_id"`
	TaskType         string         `json:"task_type"`
	ServiceSettings  map[string]any `json:"service_settings,omitempty"`
	TaskSettings     map[string]any `json:"task_settings,omitempty"`
	ChunkingSettings map[string]any `json:"chunking_settings,omitempty"`
}

// inferenceEndpointUpdateRequest is the JSON body for the update API.
type inferenceEndpointUpdateRequest struct {
	ServiceSettings  map[string]any `json:"service_settings,omitempty"`
	TaskSettings     map[string]any `json:"task_settings,omitempty"`
	ChunkingSettings map[string]any `json:"chunking_settings,omitempty"`
}

// inferenceGetResponse is the top-level GET response body.
type inferenceGetResponse struct {
	Endpoints []InferenceEndpoint `json:"endpoints"`
}

func PutInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, endpoint *InferenceEndpoint) fwdiag.Diagnostics {
	reqBody := inferenceEndpointCreateRequest{
		Service:          endpoint.Service,
		ServiceSettings:  endpoint.ServiceSettings,
		TaskSettings:     endpoint.TaskSettings,
		ChunkingSettings: endpoint.ChunkingSettings,
	}
	return doFWWrite(apiClient, reqBody,
		"Unable to marshal inference endpoint",
		"Unable to create inference endpoint",
		"Unable to create or update inference endpoint",
		func(esClient *elasticsearch.Client, body io.Reader) (*esapi.Response, error) {
			opts := []func(*esapi.InferencePutRequest){
				esClient.InferencePut.WithBody(body),
				esClient.InferencePut.WithContext(ctx),
			}
			if endpoint.TaskType != "" {
				opts = append(opts, esClient.InferencePut.WithTaskType(endpoint.TaskType))
			}
			return esClient.InferencePut(endpoint.InferenceID, opts...)
		},
	)
}

func GetInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID string) (*InferenceEndpoint, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := esClient.InferenceGet(
		esClient.InferenceGet.WithInferenceID(inferenceID),
		esClient.InferenceGet.WithContext(ctx),
	)
	if err != nil {
		diags.AddError("Unable to get inference endpoint", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckErrorFromFW(res, "Unable to get inference endpoint"); d.HasError() {
		return nil, d
	}

	var response inferenceGetResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Unable to decode inference endpoint response", err.Error())
		return nil, diags
	}

	if len(response.Endpoints) == 0 {
		return nil, nil
	}

	return &response.Endpoints[0], nil
}

func UpdateInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, update *InferenceEndpointUpdate) fwdiag.Diagnostics {
	reqBody := inferenceEndpointUpdateRequest{
		ServiceSettings:  update.ServiceSettings,
		TaskSettings:     update.TaskSettings,
		ChunkingSettings: update.ChunkingSettings,
	}
	return doFWWrite(apiClient, reqBody,
		"Unable to marshal inference endpoint update",
		"Unable to update inference endpoint",
		"Unable to update inference endpoint",
		func(esClient *elasticsearch.Client, body io.Reader) (*esapi.Response, error) {
			opts := []func(*esapi.InferenceUpdateRequest){
				esClient.InferenceUpdate.WithBody(body),
				esClient.InferenceUpdate.WithContext(ctx),
			}
			if update.TaskType != "" {
				opts = append(opts, esClient.InferenceUpdate.WithTaskType(update.TaskType))
			}
			return esClient.InferenceUpdate(update.InferenceID, opts...)
		},
	)
}

func DeleteInferenceEndpoint(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, inferenceID string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	res, err := esClient.InferenceDelete(
		inferenceID,
		esClient.InferenceDelete.WithContext(ctx),
	)
	if err != nil {
		diags.AddError("Unable to delete inference endpoint", err.Error())
		return diags
	}
	defer res.Body.Close()

	return diagutil.CheckErrorFromFW(res, "Unable to delete inference endpoint")
}
