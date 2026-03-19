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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *inferenceEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates or updates an inference endpoint. See the [inference endpoint API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-inference-put-1) for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"inference_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the inference endpoint.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"task_type": schema.StringAttribute{
				MarkdownDescription: "The task type of the inference endpoint. One of `sparse_embedding`, `text_embedding`, `rerank`, `completion`, `chat_completion`. When omitted, the task type is inferred from the service.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "The service type for the inference endpoint (e.g. `openai`, `cohere`, `elasticsearch`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_settings": schema.StringAttribute{
				MarkdownDescription: "Settings specific to the service provider, as a JSON object. May include credentials and model identifiers.",
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
				Sensitive:           true,
			},
			"task_settings": schema.StringAttribute{
				MarkdownDescription: "Task-specific settings, as a JSON object. Optional and service-dependent.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"chunking_settings": schema.StringAttribute{
				MarkdownDescription: "Configuration for chunking input text, as a JSON object. Applicable only for embedding task types.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
		},
	}
}
