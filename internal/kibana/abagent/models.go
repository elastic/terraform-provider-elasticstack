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

package abagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type agentModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AvatarColor  types.String `tfsdk:"avatar_color"`
	AvatarSymbol types.String `tfsdk:"avatar_symbol"`
	Labels       types.List   `tfsdk:"labels"` // > string
	Tools        types.List   `tfsdk:"tools"`  // > string
	Instructions types.String `tfsdk:"instructions"`
}

func (model *agentModel) populateFromAPI(ctx context.Context, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue(data.ID)
	model.Name = types.StringValue(data.Name)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	if data.AvatarColor != nil && *data.AvatarColor != "" {
		model.AvatarColor = types.StringValue(*data.AvatarColor)
	} else {
		model.AvatarColor = types.StringNull()
	}

	if data.AvatarSymbol != nil && *data.AvatarSymbol != "" {
		model.AvatarSymbol = types.StringValue(*data.AvatarSymbol)
	} else {
		model.AvatarSymbol = types.StringNull()
	}

	// Extract instructions from nested configuration
	if data.Configuration.Instructions != nil && *data.Configuration.Instructions != "" {
		model.Instructions = types.StringValue(*data.Configuration.Instructions)
	} else {
		model.Instructions = types.StringNull()
	}

	// Handle labels - keep as null if not present or empty from API
	if len(data.Labels) > 0 {
		labels, d := types.ListValueFrom(ctx, types.StringType, data.Labels)
		diags.Append(d...)
		model.Labels = labels
	} else {
		// Keep as null if not returned by API or empty
		model.Labels = types.ListNull(types.StringType)
	}

	// Extract tool IDs from nested configuration
	if len(data.Configuration.Tools) > 0 && len(data.Configuration.Tools[0].ToolIDs) > 0 {
		tools, d := types.ListValueFrom(ctx, types.StringType, data.Configuration.Tools[0].ToolIDs)
		diags.Append(d...)
		model.Tools = tools
	} else {
		// Keep as null if not returned by API or empty
		model.Tools = types.ListNull(types.StringType)
	}

	return diags
}

func (model agentModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderAgentsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostAgentBuilderAgentsJSONRequestBody{
		Id:          model.ID.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
	}

	if !model.AvatarColor.IsNull() {
		color := model.AvatarColor.ValueString()
		body.AvatarColor = &color
	}

	if !model.AvatarSymbol.IsNull() {
		symbol := model.AvatarSymbol.ValueString()
		body.AvatarSymbol = &symbol
	}

	if !model.Instructions.IsNull() {
		instructions := model.Instructions.ValueString()
		body.Configuration.Instructions = &instructions
	}

	if !model.Tools.IsNull() {
		var toolIDs []string
		d := model.Tools.ElementsAs(ctx, &toolIDs, false)
		diags.Append(d...)
		body.Configuration.Tools = []struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive // matches generated API type
		}{
			{ToolIds: toolIDs},
		}
	} else {
		body.Configuration.Tools = []struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive // matches generated API type
		}{
			{ToolIds: []string{}},
		}
	}

	if !model.Labels.IsNull() {
		var labels []string
		d := model.Labels.ElementsAs(ctx, &labels, false)
		diags.Append(d...)
		if len(labels) > 0 {
			body.Labels = &labels
		}
	}

	return body, diags
}

func (model agentModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderAgentsIdJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := model.Name.ValueString()
	body := kbapi.PutAgentBuilderAgentsIdJSONRequestBody{
		Name: &name,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	if !model.AvatarColor.IsNull() {
		color := model.AvatarColor.ValueString()
		body.AvatarColor = &color
	}

	if !model.AvatarSymbol.IsNull() {
		symbol := model.AvatarSymbol.ValueString()
		body.AvatarSymbol = &symbol
	}

	config := &struct {
		Instructions *string `json:"instructions,omitempty"`
		Tools        *[]struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive // matches generated API type
		} `json:"tools,omitempty"`
		WorkflowIds *[]string `json:"workflow_ids,omitempty"` //nolint:revive // matches generated API type
	}{}

	if !model.Instructions.IsNull() {
		instructions := model.Instructions.ValueString()
		config.Instructions = &instructions
	}

	if !model.Tools.IsNull() {
		var toolIDs []string
		d := model.Tools.ElementsAs(ctx, &toolIDs, false)
		diags.Append(d...)
		tools := []struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive // matches generated API type
		}{
			{ToolIds: toolIDs},
		}
		config.Tools = &tools
	} else {
		tools := []struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive // matches generated API type
		}{
			{ToolIds: []string{}},
		}
		config.Tools = &tools
	}

	body.Configuration = config

	if !model.Labels.IsNull() {
		var labels []string
		d := model.Labels.ElementsAs(ctx, &labels, false)
		diags.Append(d...)
		if len(labels) > 0 {
			body.Labels = &labels
		}
	}

	return body, diags
}
