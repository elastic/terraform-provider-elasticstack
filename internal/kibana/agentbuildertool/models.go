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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ entitycore.WithVersionRequirements = toolDataSourceModel{}

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

func (model toolModel) GetID() types.String             { return model.ID }
func (model toolModel) GetResourceID() types.String     { return model.ToolID }
func (model toolModel) GetSpaceID() types.String        { return model.SpaceID }
func (model toolModel) GetKibanaConnection() types.List { return model.KibanaConnection }

var _ entitycore.KibanaResourceModel = toolModel{}
var _ entitycore.WithVersionRequirements = toolModel{}

func (model toolModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return toolVersionRequirements(), nil
}

func (model toolDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return toolVersionRequirements(), nil
}

func toolVersionRequirements() []entitycore.VersionRequirement {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder tools require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
		},
	}
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
	var d diag.Diagnostics

	base := toolBaseData{
		ID:      types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String()),
		ToolID:  types.StringValue(data.ID),
		SpaceID: types.StringValue(spaceID),
		Type:    types.StringValue(data.Type),
	}

	base.Description = typeutils.NonEmptyStringOrNull(data.Description)

	base.Tags, d = typeutils.StringSetOrNull(ctx, data.Tags)
	diags.Append(d...)
	return base, diags
}

func (model *toolDataSourceModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
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

	model.ReadOnly = types.BoolValue(data.ReadOnly)

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

func toolConfigurationFromModel(config jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	configuration := make(map[string]any)
	if config.IsNull() || config.ValueString() == "" {
		return configuration, diags
	}
	if err := json.Unmarshal([]byte(config.ValueString()), &configuration); err != nil {
		diags.AddError("Configuration Error", "Failed to parse configuration JSON: "+err.Error())
		return nil, diags
	}
	return configuration, diags
}

func (model toolModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderToolsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration, d := toolConfigurationFromModel(model.Configuration)
	diags.Append(d...)
	if diags.HasError() {
		return kbapi.PostAgentBuilderToolsJSONRequestBody{}, diags
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

	tags, d := optionalTagsFromSet(ctx, model.Tags)
	diags.Append(d...)
	if len(tags) > 0 {
		body.Tags = &tags
	}

	return body, diags
}

func (model toolModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderToolsToolidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration, d := toolConfigurationFromModel(model.Configuration)
	diags.Append(d...)
	if diags.HasError() {
		return kbapi.PutAgentBuilderToolsToolidJSONRequestBody{}, diags
	}

	apiConfiguration := typeutils.PointerInterfaceMapFromAnyMap(configuration)
	body := kbapi.PutAgentBuilderToolsToolidJSONRequestBody{
		Configuration: &apiConfiguration,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	tags, d := optionalTagsFromSet(ctx, model.Tags)
	diags.Append(d...)
	if len(tags) > 0 {
		body.Tags = &tags
	}

	return body, diags
}

func optionalTagsFromSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var diags diag.Diagnostics
	tags := typeutils.SetTypeAs[string](ctx, set, path.Empty(), &diags)
	return tags, diags
}
