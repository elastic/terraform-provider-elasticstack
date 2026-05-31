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

func (model agentModel) GetID() types.String         { return model.ID }
func (model agentModel) GetResourceID() types.String { return model.AgentID }
func (model agentModel) GetSpaceID() types.String    { return model.SpaceID }
func (agentModel) UsesCompositeResourceID() bool     { return true }

var _ entitycore.KibanaResourceModel = agentModel{}
var _ entitycore.WithVersionRequirements = agentModel{}
var _ entitycore.WithVersionRequirements = agentDataSourceModel{}

// agentVersionGate is a zero-size embedded struct that satisfies
// entitycore.WithVersionRequirements for both agentModel and agentDataSourceModel,
// eliminating the duplicate method bodies.
type agentVersionGate struct{}

func (agentVersionGate) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder agents require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
		},
	}, nil
}

// agentBaseModel holds every field shared by the resource model (agentModel)
// and the data source model (agentDataSourceModel). Both embed it so the common
// schema fields, the kibana_connection block (via KibanaConnectionField), the
// version gate, and the population logic are declared exactly once.
//
// The `tools` field is intentionally NOT part of this base: the resource
// represents it as a set of tool IDs (types.Set) while the data source returns
// rich, nested tool objects ([]toolModel). Both use the `tfsdk:"tools"` tag, so
// declaring it in the base (or embedding one model into the other directly)
// would trip the Plugin Framework's duplicate-tag detection.
type agentBaseModel struct {
	entitycore.KibanaConnectionField
	agentVersionGate
	ID           types.String `tfsdk:"id"`
	AgentID      types.String `tfsdk:"agent_id"`
	SpaceID      types.String `tfsdk:"space_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AvatarColor  types.String `tfsdk:"avatar_color"`
	AvatarSymbol types.String `tfsdk:"avatar_symbol"`
	Labels       types.Set    `tfsdk:"labels"`    // []string
	SkillIDs     types.Set    `tfsdk:"skill_ids"` // []string
	Instructions types.String `tfsdk:"instructions"`
}

type agentModel struct {
	agentBaseModel
	Tools types.Set `tfsdk:"tools"` // []string
}

type agentDataSourceModel struct {
	agentBaseModel
	IncludeDependencies types.Bool  `tfsdk:"include_dependencies"`
	Tools               []toolModel `tfsdk:"tools"`
}

func (m agentDataSourceModel) GetID() types.String         { return m.ID }
func (m agentDataSourceModel) GetResourceID() types.String { return m.AgentID }
func (m agentDataSourceModel) GetSpaceID() types.String    { return m.SpaceID }
func (agentDataSourceModel) UsesCompositeResourceID() bool { return true }

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

// populateFromAPI maps the fields shared by the resource and data source models
// from an API response. Entity-specific fields (the differing `tools`
// representations and the data source's include_dependencies) are populated by
// the callers.
func (model *agentBaseModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	var diags diag.Diagnostics
	if data == nil {
		return diags
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.AgentID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.Name = types.StringValue(data.Name)
	model.AvatarColor = types.StringPointerValue(data.AvatarColor)
	model.AvatarSymbol = types.StringPointerValue(data.AvatarSymbol)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	if cfg := data.Configuration; cfg.Instructions != nil && *cfg.Instructions != "" {
		model.Instructions = types.StringValue(*cfg.Instructions)
	} else {
		model.Instructions = types.StringNull()
	}

	diags.Append(agentbuilder.PopulateSet(ctx, data.Labels, &model.Labels)...)
	diags.Append(agentbuilder.PopulateSet(ctx, data.Configuration.SkillIDs, &model.SkillIDs)...)
	return diags
}

func (model *agentModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Agent) diag.Diagnostics {
	if data == nil {
		return nil
	}
	diags := model.agentBaseModel.populateFromAPI(ctx, spaceID, data)
	var toolIDs []string
	if len(data.Configuration.Tools) > 0 {
		toolIDs = data.Configuration.Tools[0].ToolIDs
	}
	diags.Append(agentbuilder.PopulateSet(ctx, toolIDs, &model.Tools)...)
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
