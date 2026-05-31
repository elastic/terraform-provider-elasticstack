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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model agentModel) GetID() types.String             { return model.ID }
func (model agentModel) GetResourceID() types.String     { return model.AgentID }
func (model agentModel) GetSpaceID() types.String        { return model.SpaceID }
func (model agentModel) GetKibanaConnection() types.List { return model.KibanaConnection }

var _ entitycore.KibanaResourceModel = agentModel{}
var _ entitycore.WithVersionRequirements = agentModel{}

func (model agentModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return agentVersionRequirements(), nil
}

type agentModel struct {
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	AgentID          types.String `tfsdk:"agent_id"`
	SpaceID          types.String `tfsdk:"space_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AvatarColor      types.String `tfsdk:"avatar_color"`
	AvatarSymbol     types.String `tfsdk:"avatar_symbol"`
	Labels           types.Set    `tfsdk:"labels"`    // []string
	Tools            types.Set    `tfsdk:"tools"`     // []string
	SkillIDs         types.Set    `tfsdk:"skill_ids"` // []string
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
	SkillIDs            types.Set    `tfsdk:"skill_ids"`
	Instructions        types.String `tfsdk:"instructions"`
}

func (model agentDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return agentVersionRequirements(), nil
}

func agentVersionRequirements() []entitycore.VersionRequirement {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder agents require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
		},
	}
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

// agentBaseData holds fields shared between agentDataSourceModel and agentModel
// populated from the API response.
type agentBaseData struct {
	ID           types.String
	AgentID      types.String
	SpaceID      types.String
	Name         types.String
	Description  types.String
	AvatarColor  types.String
	AvatarSymbol types.String
	Instructions types.String
	Labels       types.Set
}

// populateAgentBaseFromAPI extracts the fields common to both agentDataSourceModel
// and agentModel from an API response, eliminating duplicated population logic.
func populateAgentBaseFromAPI(ctx context.Context, spaceID string, data *models.Agent) (agentBaseData, diag.Diagnostics) {
	var diags diag.Diagnostics

	base := agentBaseData{
		ID:           types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String()),
		AgentID:      types.StringValue(data.ID),
		SpaceID:      types.StringValue(spaceID),
		Name:         types.StringValue(data.Name),
		AvatarColor:  types.StringPointerValue(data.AvatarColor),
		AvatarSymbol: types.StringPointerValue(data.AvatarSymbol),
	}

	if data.Description != nil && *data.Description != "" {
		base.Description = types.StringValue(*data.Description)
	} else {
		base.Description = types.StringNull()
	}

	cfg := data.Configuration
	if cfg.Instructions != nil && *cfg.Instructions != "" {
		base.Instructions = types.StringValue(*cfg.Instructions)
	} else {
		base.Instructions = types.StringNull()
	}

	diags.Append(agentbuilder.PopulateSet(ctx, data.Labels, &base.Labels)...)
	return base, diags
}

func (model *agentDataSourceModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}
	base, diags := populateAgentBaseFromAPI(ctx, spaceID, data)
	model.ID = base.ID
	model.AgentID = base.AgentID
	model.SpaceID = base.SpaceID
	model.Name = base.Name
	model.Description = base.Description
	model.AvatarColor = base.AvatarColor
	model.AvatarSymbol = base.AvatarSymbol
	model.Instructions = base.Instructions
	model.Labels = base.Labels
	diags.Append(agentbuilder.PopulateSet(ctx, data.Configuration.SkillIDs, &model.SkillIDs)...)
	return diags
}

func (model *agentModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}
	base, diags := populateAgentBaseFromAPI(ctx, spaceID, data)
	model.ID = base.ID
	model.AgentID = base.AgentID
	model.SpaceID = base.SpaceID
	model.Name = base.Name
	model.Description = base.Description
	model.AvatarColor = base.AvatarColor
	model.AvatarSymbol = base.AvatarSymbol
	model.Instructions = base.Instructions
	model.Labels = base.Labels
	var toolIDs []string
	if len(data.Configuration.Tools) > 0 {
		toolIDs = data.Configuration.Tools[0].ToolIDs
	}
	diags.Append(agentbuilder.PopulateSet(ctx, toolIDs, &model.Tools)...)
	diags.Append(agentbuilder.PopulateSet(ctx, data.Configuration.SkillIDs, &model.SkillIDs)...)
	return diags
}

func (model agentModel) toAPICreateModel(ctx context.Context, supportsSkillIDs bool) (kbapi.PostAgentBuilderAgentsJSONRequestBody, diag.Diagnostics) {
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

	toolIDs, d := agentbuilder.SetToStrings(ctx, model.Tools)
	diags.Append(d...)
	body.Configuration.Tools = []struct {
		ToolIds []string `json:"tool_ids"` //nolint:revive
	}{{ToolIds: toolIDs}}

	if supportsSkillIDs {
		skillIDs, d := agentbuilder.SetToStrings(ctx, model.SkillIDs)
		diags.Append(d...)
		if len(skillIDs) > 0 {
			body.Configuration.SkillIds = &skillIDs
		}
	}

	labels, d := agentbuilder.SetToStrings(ctx, model.Labels)
	diags.Append(d...)
	if len(labels) > 0 {
		body.Labels = &labels
	}

	return body, diags
}

func (model agentModel) toAPIUpdateModel(ctx context.Context, supportsSkillIDs bool) (kbapi.PutAgentBuilderAgentsIdJSONRequestBody, diag.Diagnostics) {
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

	toolIDs, d := agentbuilder.SetToStrings(ctx, model.Tools)
	diags.Append(d...)
	tools := []struct {
		ToolIds []string `json:"tool_ids"` //nolint:revive
	}{{ToolIds: toolIDs}}

	// Always send skill_ids on update (including empty) when the server
	// supports it so cleared values are propagated to Kibana. On 9.3.x the
	// field is unknown and must be omitted entirely (nil pointer → omitempty
	// drops it). A non-nil pointer to an empty slice serialises as [] and
	// clears the value on 9.4+.
	var skillIDsPtr *[]string
	if supportsSkillIDs {
		skillIDs, d := agentbuilder.SetToStrings(ctx, model.SkillIDs)
		diags.Append(d...)
		skillIDsPtr = &skillIDs
	}

	var instructions *string
	if !model.Instructions.IsNull() && !model.Instructions.IsUnknown() {
		instructions = model.Instructions.ValueStringPointer()
	}

	body.Configuration = &struct {
		ConnectorIds              *[]string `json:"connector_ids,omitempty"` //nolint:revive
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
		SkillIds:     skillIDsPtr,
	}

	labels, d := agentbuilder.SetToStrings(ctx, model.Labels)
	diags.Append(d...)
	if len(labels) > 0 {
		body.Labels = &labels
	}

	return body, diags
}
