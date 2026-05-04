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

package agentbuilderagent

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type agentModel struct {
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	AgentID          types.String `tfsdk:"agent_id"`
	SpaceID          types.String `tfsdk:"space_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AvatarColor      types.String `tfsdk:"avatar_color"`
	AvatarSymbol     types.String `tfsdk:"avatar_symbol"`
	Labels           types.Set    `tfsdk:"labels"` // []string
	Tools            types.Set    `tfsdk:"tools"`  // []string
	Instructions     types.String `tfsdk:"instructions"`
}

type agentDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID                  types.String `tfsdk:"id"`
	AgentID             types.String `tfsdk:"agent_id"`
	SpaceID             types.String `tfsdk:"space_id"`
	IncludeDependencies types.Bool   `tfsdk:"include_dependencies"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	AvatarColor         types.String `tfsdk:"avatar_color"`
	AvatarSymbol        types.String `tfsdk:"avatar_symbol"`
	Labels              types.Set    `tfsdk:"labels"`
	Tools               []toolModel  `tfsdk:"tools"`
	Instructions        types.String `tfsdk:"instructions"`
}

// GetVersionRequirements returns the static minimum Kibana version requirements
// for the Agent Builder agent data source. This satisfies the optional
// entitycore.KibanaDataSourceWithVersionRequirements interface, allowing the
// generic Kibana data source envelope to enforce the requirement before invoking
// the entity read callback.
func (model agentDataSourceModel) GetVersionRequirements() ([]entitycore.DataSourceVersionRequirement, diag.Diagnostics) {
	return []entitycore.DataSourceVersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder agents require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
		},
	}, nil
}

type toolModel struct {
	ID                        types.String                    `tfsdk:"id"`
	SpaceID                   types.String                    `tfsdk:"space_id"`
	ToolID                    types.String                    `tfsdk:"tool_id"`
	Type                      types.String                    `tfsdk:"type"`
	Description               types.String                    `tfsdk:"description"`
	Tags                      types.Set                       `tfsdk:"tags"`
	ReadOnly                  types.Bool                      `tfsdk:"readonly"`
	Configuration             types.String                    `tfsdk:"configuration"`
	WorkflowID                types.String                    `tfsdk:"workflow_id"`
	WorkflowConfigurationYaml customtypes.NormalizedYamlValue `tfsdk:"workflow_configuration_yaml"`
}

func (model *agentDataSourceModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.AgentID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
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

	cfg := data.Configuration

	if cfg.Instructions != nil && *cfg.Instructions != "" {
		model.Instructions = types.StringValue(*cfg.Instructions)
	} else {
		model.Instructions = types.StringNull()
	}

	if len(data.Labels) > 0 {
		labels, d := types.SetValueFrom(ctx, types.StringType, data.Labels)
		diags.Append(d...)
		model.Labels = labels
	} else {
		model.Labels = types.SetNull(types.StringType)
	}

	return diags
}

func (model *agentModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.AgentID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
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

	cfg := data.Configuration

	if cfg.Instructions != nil && *cfg.Instructions != "" {
		model.Instructions = types.StringValue(*cfg.Instructions)
	} else {
		model.Instructions = types.StringNull()
	}

	diags.Append(populateSet(ctx, data.Labels, &model.Labels)...)

	var toolIDs []string
	if len(cfg.Tools) > 0 {
		toolIDs = cfg.Tools[0].ToolIDs
	}
	diags.Append(populateSet(ctx, toolIDs, &model.Tools)...)

	return diags
}

func populateSet(ctx context.Context, src []string, dst *types.Set) diag.Diagnostics {
	if len(src) > 0 {
		v, d := types.SetValueFrom(ctx, types.StringType, src)
		*dst = v
		return d
	}
	*dst = types.SetNull(types.StringType)
	return nil
}

func (model agentModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderAgentsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostAgentBuilderAgentsJSONRequestBody{
		Id:          model.AgentID.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
	}

	if !model.AvatarColor.IsNull() && !model.AvatarColor.IsUnknown() {
		body.AvatarColor = model.AvatarColor.ValueStringPointer()
	}
	if !model.AvatarSymbol.IsNull() && !model.AvatarSymbol.IsUnknown() {
		body.AvatarSymbol = model.AvatarSymbol.ValueStringPointer()
	}
	if !model.Instructions.IsNull() && !model.Instructions.IsUnknown() {
		body.Configuration.Instructions = model.Instructions.ValueStringPointer()
	}

	toolIDs, d := setToStrings(ctx, model.Tools)
	diags.Append(d...)
	body.Configuration.Tools = []struct {
		ToolIds []string `json:"tool_ids"` //nolint:revive
	}{{ToolIds: toolIDs}}

	labels, d := setToStrings(ctx, model.Labels)
	diags.Append(d...)
	if len(labels) > 0 {
		body.Labels = &labels
	}

	return body, diags
}

func (model agentModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderAgentsIdJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := model.Name.ValueString()
	body := kbapi.PutAgentBuilderAgentsIdJSONRequestBody{
		Name: &name,
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		body.Description = model.Description.ValueStringPointer()
	}
	if !model.AvatarColor.IsNull() && !model.AvatarColor.IsUnknown() {
		body.AvatarColor = model.AvatarColor.ValueStringPointer()
	}
	if !model.AvatarSymbol.IsNull() && !model.AvatarSymbol.IsUnknown() {
		body.AvatarSymbol = model.AvatarSymbol.ValueStringPointer()
	}

	toolIDs, d := setToStrings(ctx, model.Tools)
	diags.Append(d...)
	tools := []struct {
		ToolIds []string `json:"tool_ids"` //nolint:revive
	}{{ToolIds: toolIDs}}

	var instructions *string
	if !model.Instructions.IsNull() && !model.Instructions.IsUnknown() {
		instructions = model.Instructions.ValueStringPointer()
	}

	body.Configuration = &struct {
		EnableElasticCapabilities *bool     `json:"enable_elastic_capabilities,omitempty"`
		Instructions              *string   `json:"instructions,omitempty"`
		PluginIds                 *[]string `json:"plugin_ids,omitempty"` //nolint:revive
		SkillIds                  *[]string `json:"skill_ids,omitempty"`  //nolint:revive
		Tools                     *[]struct {
			ToolIds []string `json:"tool_ids"` //nolint:revive
		} `json:"tools,omitempty"`
		WorkflowIds *[]string `json:"workflow_ids,omitempty"` //nolint:revive
	}{
		Instructions: instructions,
		Tools:        &tools,
	}

	labels, d := setToStrings(ctx, model.Labels)
	diags.Append(d...)
	if len(labels) > 0 {
		body.Labels = &labels
	}

	return body, diags
}

func setToStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return []string{}, nil
	}
	var out []string
	d := set.ElementsAs(ctx, &out, false)
	return out, d
}
