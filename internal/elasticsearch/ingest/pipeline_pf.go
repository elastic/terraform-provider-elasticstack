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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

var (
	_ resource.Resource                = newPipelineResource()
	_ resource.ResourceWithConfigure   = newPipelineResource()
	_ resource.ResourceWithImportState = newPipelineResource()
)

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

// GetSchema returns the Plugin Framework schema without the
// elasticsearch_connection block, which is injected by the envelope.
func GetSchema(_ context.Context) schema.Schema {
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
				ElementType:         ProcessorJSONType{},
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"on_failure": schema.ListAttribute{
				MarkdownDescription: ingestPipelineOnFailureDescription,
				Optional:            true,
				ElementType:         ProcessorJSONType{},
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
			writeIngestPipeline,
			writeIngestPipeline,
		),
	}
}

func NewIngestPipelineResource() resource.Resource {
	return newPipelineResource()
}

func (r *pipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

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

	data := Data{
		ElasticsearchConnectionField: entitycore.ElasticsearchConnectionField{ElasticsearchConnection: state.ElasticsearchConnection},
		ID:                           types.StringValue(compID.String()),
		Name:                         types.StringValue(resourceID),
		Description:                  types.StringPointerValue(pipeline.Description),
	}

	processorsList, listDiags := jsonListFromSlice(ctx, pipeline.Processors, "processor")
	diags.Append(listDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	data.Processors = processorsList

	onFailureList, listDiags := jsonListFromSlice(ctx, pipeline.OnFailure, "on_failure processor")
	diags.Append(listDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	data.OnFailure = onFailureList

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

func deleteIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ Data) diag.Diagnostics {
	return diagutil.FrameworkDiagsFromSDK(elasticsearch.DeleteIngestPipeline(ctx, client, resourceID))
}

func writeIngestPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
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

func buildPipelineBody(ctx context.Context, data Data) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := map[string]any{}

	if typeutils.IsKnown(data.Description) {
		body["description"] = data.Description.ValueString()
	}

	processors, procDiags := decodeJSONList(ctx, data.Processors, "processor")
	diags.Append(procDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if processors == nil {
		// Required + MinItems=1 enforced by schema; envelope should not call us with a missing list.
		diags.AddError("Missing required processors", "processors must contain at least one element")
		return nil, diags
	}
	body["processors"] = processors

	onFailure, ofDiags := decodeJSONList(ctx, data.OnFailure, "on_failure processor")
	diags.Append(ofDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if onFailure != nil {
		body["on_failure"] = onFailure
	}

	if typeutils.IsKnown(data.Metadata) {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
			diags.AddError("Failed to decode metadata JSON", err.Error())
			return nil, diags
		}
		body["_meta"] = metadata
	}

	return body, diags
}

// jsonListFromSlice marshals each element of items to JSON and returns a
// types.List of jsontypes.Normalized values; an empty/nil input yields a null
// list so Terraform does not record an empty Optional list as set.
func jsonListFromSlice[T any](ctx context.Context, items []T, label string) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(ProcessorJSONType{}), nil
	}
	values := make([]ProcessorJSONValue, len(items))
	for i, v := range items {
		b, err := json.Marshal(v)
		if err != nil {
			var diags diag.Diagnostics
			diags.AddError(fmt.Sprintf("Failed to serialize %s", label), err.Error())
			return types.ListNull(ProcessorJSONType{}), diags
		}
		values[i] = NewProcessorJSONValue(string(b))
	}
	return types.ListValueFrom(ctx, ProcessorJSONType{}, values)
}

// decodeJSONList unmarshals each Normalized element of list into a JSON object.
// Returns (nil, nil) when the list is null or unknown.
func decodeJSONList(ctx context.Context, list types.List, label string) ([]map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(list) {
		return nil, diags
	}
	var values []ProcessorJSONValue
	diags.Append(list.ElementsAs(ctx, &values, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]map[string]any, len(values))
	for i, v := range values {
		item := map[string]any{}
		if err := json.Unmarshal([]byte(v.ValueString()), &item); err != nil {
			diags.AddError(fmt.Sprintf("Failed to decode %s JSON", label), err.Error())
			return nil, diags
		}
		out[i] = item
	}
	return out, diags
}
