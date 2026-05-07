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

package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure pipelineResource satisfies framework interfaces.
var (
	_ resource.Resource                = newPipelineResource()
	_ resource.ResourceWithConfigure   = newPipelineResource()
	_ resource.ResourceWithImportState = newPipelineResource()
)

// Data is the Plugin Framework model for the ingest pipeline resource.
type Data struct {
	entitycore.ElasticsearchConnectionField
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Processors  types.List           `tfsdk:"processors"`
	OnFailure   types.List           `tfsdk:"on_failure"`
	Metadata    jsontypes.Normalized `tfsdk:"metadata"`
}

func (d Data) GetID() types.String         { return d.ID }
func (d Data) GetResourceID() types.String { return d.Name }

// GetSchema returns the Plugin Framework schema for the ingest pipeline resource
// without the elasticsearch_connection block (injected by the envelope).
func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages tasks and resources related to ingest pipelines and processors. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the ingest pipeline.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the ingest pipeline.",
				Optional:            true,
			},
			"processors": schema.ListAttribute{
				MarkdownDescription: ingestPipelineProcessorsDescription,
				Required:            true,
				ElementType:         jsontypes.NormalizedType{},
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"on_failure": schema.ListAttribute{
				MarkdownDescription: ingestPipelineOnFailureDescription,
				Optional:            true,
				ElementType:         jsontypes.NormalizedType{},
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional user metadata about the ingest pipeline.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
		},
	}
}

// pipelineResource is the concrete Plugin Framework resource type.
type pipelineResource struct {
	*entitycore.ElasticsearchResource[Data]
}

func newPipelineResource() *pipelineResource {
	return &pipelineResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[Data](
			entitycore.ComponentElasticsearch,
			"ingest_pipeline",
			GetSchema,
			readIngestPipeline,
			deleteIngestPipeline,
			createIngestPipeline,
			updateIngestPipeline,
		),
	}
}

// NewIngestPipelineResource returns a new Plugin Framework resource for the ingest pipeline.
func NewIngestPipelineResource() resource.Resource {
	return newPipelineResource()
}

// ImportState implements resource.ResourceWithImportState.
func (r *pipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readIngestPipeline is the readFunc callback for the envelope.
func readIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	pipeline, apiDiags := elasticsearch.GetIngestPipeline(ctx, client, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(apiDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	if pipeline == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Ingest pipeline "%s" not found, removing from state`, resourceID))
		return state, false, nil
	}

	compID, compDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(compDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	data := Data{}
	data.ElasticsearchConnection = state.ElasticsearchConnection
	data.ID = types.StringValue(compID.String())
	data.Name = types.StringValue(resourceID)

	if pipeline.Description != nil {
		data.Description = types.StringValue(*pipeline.Description)
	} else {
		data.Description = types.StringNull()
	}

	// Serialize processors
	procs := make([]jsontypes.Normalized, len(pipeline.Processors))
	for i, v := range pipeline.Processors {
		b, err := json.Marshal(v)
		if err != nil {
			diags.AddError("Failed to serialize processor", err.Error())
			return state, false, diags
		}
		procs[i] = jsontypes.NewNormalizedValue(string(b))
	}
	processorsList, listDiags := types.ListValueFrom(ctx, jsontypes.NormalizedType{}, procs)
	diags.Append(listDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	data.Processors = processorsList

	// Serialize on_failure
	if len(pipeline.OnFailure) > 0 {
		failureProcs := make([]jsontypes.Normalized, len(pipeline.OnFailure))
		for i, v := range pipeline.OnFailure {
			b, err := json.Marshal(v)
			if err != nil {
				diags.AddError("Failed to serialize on_failure processor", err.Error())
				return state, false, diags
			}
			failureProcs[i] = jsontypes.NewNormalizedValue(string(b))
		}
		onFailureList, listDiags := types.ListValueFrom(ctx, jsontypes.NormalizedType{}, failureProcs)
		diags.Append(listDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.OnFailure = onFailureList
	} else {
		data.OnFailure = types.ListNull(jsontypes.NormalizedType{})
	}

	// Serialize metadata
	if pipeline.Meta_ != nil {
		b, err := json.Marshal(pipeline.Meta_)
		if err != nil {
			diags.AddError("Failed to serialize metadata", err.Error())
			return state, false, diags
		}
		data.Metadata = jsontypes.NewNormalizedValue(string(b))
	} else {
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	return data, true, diags
}

// deleteIngestPipeline is the deleteFunc callback for the envelope.
func deleteIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ Data) diag.Diagnostics {
	return diagutil.FrameworkDiagsFromSDK(elasticsearch.DeleteIngestPipeline(ctx, client, resourceID))
}

// createIngestPipeline is the createFunc callback for the envelope.
func createIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	body, buildDiags := buildPipelineBody(ctx, data)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return data, diags
	}

	apiDiags := elasticsearch.PutIngestPipeline(ctx, client, resourceID, body)
	diags.Append(diagutil.FrameworkDiagsFromSDK(apiDiags)...)
	if diags.HasError() {
		return data, diags
	}

	compID, compDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(compDiags)...)
	if diags.HasError() {
		return data, diags
	}
	data.ID = types.StringValue(compID.String())

	return data, diags
}

// updateIngestPipeline is the updateFunc callback; identical to createIngestPipeline for PUT-based resources.
func updateIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	return createIngestPipeline(ctx, client, resourceID, data)
}

// buildPipelineBody constructs the JSON body map for a PutIngestPipeline call.
func buildPipelineBody(ctx context.Context, data Data) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := map[string]any{}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body["description"] = data.Description.ValueString()
	}

	// Decode processors list
	if data.Processors.IsNull() || data.Processors.IsUnknown() {
		diags.AddError("Missing required processors", "processors must contain at least one element")
		return nil, diags
	}

	if !data.Processors.IsNull() && !data.Processors.IsUnknown() {
		var procValues []jsontypes.Normalized
		diags.Append(data.Processors.ElementsAs(ctx, &procValues, false)...)
		if diags.HasError() {
			return nil, diags
		}
		procs := make([]map[string]any, len(procValues))
		for i, v := range procValues {
			item := map[string]any{}
			if err := json.Unmarshal([]byte(v.ValueString()), &item); err != nil {
				diags.AddError("Failed to decode processor JSON", err.Error())
				return nil, diags
			}
			procs[i] = item
		}
		body["processors"] = procs
	}

	// Decode on_failure list
	if !data.OnFailure.IsNull() && !data.OnFailure.IsUnknown() {
		var failureValues []jsontypes.Normalized
		diags.Append(data.OnFailure.ElementsAs(ctx, &failureValues, false)...)
		if diags.HasError() {
			return nil, diags
		}
		failureProcs := make([]map[string]any, len(failureValues))
		for i, v := range failureValues {
			item := map[string]any{}
			if err := json.Unmarshal([]byte(v.ValueString()), &item); err != nil {
				diags.AddError("Failed to decode on_failure processor JSON", err.Error())
				return nil, diags
			}
			failureProcs[i] = item
		}
		body["on_failure"] = failureProcs
	}

	// Decode metadata
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
			diags.AddError("Failed to decode metadata JSON", err.Error())
			return nil, diags
		}
		body["_meta"] = metadata
	}

	return body, diags
}
