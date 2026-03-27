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

package agentbuildertool

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type toolModel struct {
	ID            types.String `tfsdk:"id"`
	ToolID        types.String `tfsdk:"tool_id"`
	SpaceID       types.String `tfsdk:"space_id"`
	Type          types.String `tfsdk:"type"`
	Description   types.String `tfsdk:"description"`
	Tags          types.List   `tfsdk:"tags"`
	Configuration types.String `tfsdk:"configuration"`
}

func (model *toolModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	spaceID := model.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = "default"
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.ToolID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.Type = types.StringValue(data.Type)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(data.Tags) > 0 {
		tags, d := types.ListValueFrom(ctx, types.StringType, data.Tags)
		diags.Append(d...)
		model.Tags = tags
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	if data.Configuration != nil {
		configJSON, err := json.Marshal(data.Configuration)
		if err != nil {
			diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return diags
		}
		model.Configuration = types.StringValue(string(configJSON))
	} else {
		model.Configuration = types.StringNull()
	}

	return diags
}

func (model toolModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderToolsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration := make(map[string]any)

	if !model.Configuration.IsNull() && model.Configuration.ValueString() != "" {
		if err := json.Unmarshal([]byte(model.Configuration.ValueString()), &configuration); err != nil {
			diags.AddError("Configuration Error", "Failed to parse configuration JSON: "+err.Error())
			return kbapi.PostAgentBuilderToolsJSONRequestBody{}, diags
		}
	}

	body := kbapi.PostAgentBuilderToolsJSONRequestBody{
		Id:            model.ToolID.ValueString(),
		Type:          kbapi.PostAgentBuilderToolsJSONBodyType(model.Type.ValueString()),
		Configuration: configuration,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	if !model.Tags.IsNull() {
		var tags []string
		d := model.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(d...)
		if len(tags) > 0 {
			body.Tags = &tags
		}
	}

	return body, diags
}

func (model toolModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderToolsToolidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration := make(map[string]any)

	if !model.Configuration.IsNull() && model.Configuration.ValueString() != "" {
		if err := json.Unmarshal([]byte(model.Configuration.ValueString()), &configuration); err != nil {
			diags.AddError("Configuration Error", "Failed to parse configuration JSON: "+err.Error())
			return kbapi.PutAgentBuilderToolsToolidJSONRequestBody{}, diags
		}
	}

	body := kbapi.PutAgentBuilderToolsToolidJSONRequestBody{
		Configuration: &configuration,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	if !model.Tags.IsNull() {
		var tags []string
		d := model.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(d...)
		if len(tags) > 0 {
			body.Tags = &tags
		}
	}

	return body, diags
}
