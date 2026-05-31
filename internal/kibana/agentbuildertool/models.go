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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type toolModel struct {
	ID               types.String         `tfsdk:"id"`
	KibanaConnection types.List           `tfsdk:"kibana_connection"`
	ToolID           types.String         `tfsdk:"tool_id"`
	SpaceID          types.String         `tfsdk:"space_id"`
	Type             types.String         `tfsdk:"type"`
	Description      types.String         `tfsdk:"description"`
	Tags             types.Set            `tfsdk:"tags"`
	Configuration    jsontypes.Normalized `tfsdk:"configuration"`
}

type toolDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID                        types.String                    `tfsdk:"id"`
	SpaceID                   types.String                    `tfsdk:"space_id"`
	ToolID                    types.String                    `tfsdk:"tool_id"`
	Type                      types.String                    `tfsdk:"type"`
	Description               types.String                    `tfsdk:"description"`
	Tags                      types.Set                       `tfsdk:"tags"`
	ReadOnly                  types.Bool                      `tfsdk:"readonly"`
	Configuration             types.String                    `tfsdk:"configuration"`
	IncludeWorkflow           types.Bool                      `tfsdk:"include_workflow"`
	WorkflowID                types.String                    `tfsdk:"workflow_id"`
	WorkflowConfigurationYaml customtypes.NormalizedYamlValue `tfsdk:"workflow_configuration_yaml"`
}

func (m toolDataSourceModel) GetID() types.String         { return m.ID }
func (m toolDataSourceModel) GetResourceID() types.String { return m.ToolID }
func (m toolDataSourceModel) GetSpaceID() types.String    { return m.SpaceID }
func (toolDataSourceModel) UsesCompositeResourceID() bool { return true }

func (model toolModel) GetID() types.String             { return model.ID }
func (model toolModel) GetResourceID() types.String     { return model.ToolID }
func (model toolModel) GetSpaceID() types.String        { return model.SpaceID }
func (toolModel) UsesCompositeResourceID() bool         { return true }
func (model toolModel) GetKibanaConnection() types.List { return model.KibanaConnection }

var _ entitycore.KibanaResourceModel = toolModel{}
var _ entitycore.WithVersionRequirements = toolModel{}

func (model toolModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder tools require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
		},
	}, nil
}

// toolBaseData holds fields shared between toolDataSourceModel and toolModel
// populated from the API response.
type toolBaseData struct {
	ID          types.String
	ToolID      types.String
	SpaceID     types.String
	Type        types.String
	Description types.String
	Tags        types.Set
}

// populateToolBaseFromAPI extracts the fields common to both toolDataSourceModel
// and toolModel from an API response, eliminating duplicated population logic.
func populateToolBaseFromAPI(ctx context.Context, data *models.Tool, spaceID string) (toolBaseData, diag.Diagnostics) {
	var diags diag.Diagnostics

	base := toolBaseData{
		ID:      types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String()),
		ToolID:  types.StringValue(data.ID),
		SpaceID: types.StringValue(spaceID),
		Type:    types.StringValue(data.Type),
	}

	if data.Description != nil && *data.Description != "" {
		base.Description = types.StringValue(*data.Description)
	} else {
		base.Description = types.StringNull()
	}

	diags.Append(agentbuilder.PopulateSet(ctx, data.Tags, &base.Tags)...)
	return base, diags
}

func (m *toolDataSourceModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	spaceID := m.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	base, diags := populateToolBaseFromAPI(ctx, data, spaceID)
	m.ID = base.ID
	m.ToolID = base.ToolID
	m.SpaceID = base.SpaceID
	m.Type = base.Type
	m.Description = base.Description
	m.Tags = base.Tags

	m.ReadOnly = types.BoolValue(data.ReadOnly)

	if data.Configuration != nil {
		configJSON, err := json.Marshal(data.Configuration)
		if err != nil {
			diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return diags
		}
		m.Configuration = types.StringValue(string(configJSON))
	} else {
		m.Configuration = types.StringNull()
	}

	return diags
}

func (model *toolModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	spaceID := model.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	base, diags := populateToolBaseFromAPI(ctx, data, spaceID)
	model.ID = base.ID
	model.ToolID = base.ToolID
	model.SpaceID = base.SpaceID
	model.Type = base.Type
	model.Description = base.Description
	model.Tags = base.Tags

	if data.Configuration != nil {
		configJSON, err := json.Marshal(data.Configuration)
		if err != nil {
			diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return diags
		}
		model.Configuration = jsontypes.NewNormalizedValue(string(configJSON))
	} else {
		model.Configuration = jsontypes.NewNormalizedNull()
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
		Configuration: typeutils.PointerInterfaceMapFromAnyMap(configuration),
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

	apiConfiguration := typeutils.PointerInterfaceMapFromAnyMap(configuration)
	body := kbapi.PutAgentBuilderToolsToolidJSONRequestBody{
		Configuration: &apiConfiguration,
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
