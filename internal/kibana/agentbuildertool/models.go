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

// toolModel is the core model for the Agent Builder tool resource. It also
// serves as the shared base for toolDataSourceModel, which embeds it. Defining
// the version gate and the kibana_connection block (via KibanaConnectionField)
// here means they are declared exactly once rather than duplicated across the
// resource and data source models.
type toolModel struct {
	entitycore.KibanaConnectionField
	ID            types.String         `tfsdk:"id"`
	ToolID        types.String         `tfsdk:"tool_id"`
	SpaceID       types.String         `tfsdk:"space_id"`
	Type          types.String         `tfsdk:"type"`
	Description   types.String         `tfsdk:"description"`
	Tags          types.Set            `tfsdk:"tags"`
	Configuration jsontypes.Normalized `tfsdk:"configuration"`
}

func (toolModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *minKibanaAgentBuilderAPIVersion,
		ErrorMessage: fmt.Sprintf("Agent Builder tools require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
	}}, nil
}

func (model toolModel) GetID() types.String         { return model.ID }
func (model toolModel) GetResourceID() types.String { return model.ToolID }
func (model toolModel) GetSpaceID() types.String    { return model.SpaceID }

var _ entitycore.KibanaResourceModel = toolModel{}
var _ entitycore.WithVersionRequirements = toolModel{}

// toolDataSourceModel embeds toolModel to inherit the shared fields, the
// kibana_connection block, and the version gate, adding only the attributes
// that are unique to the data source.
type toolDataSourceModel struct {
	toolModel
	ReadOnly                  types.Bool                      `tfsdk:"readonly"`
	IncludeWorkflow           types.Bool                      `tfsdk:"include_workflow"`
	WorkflowID                types.String                    `tfsdk:"workflow_id"`
	WorkflowConfigurationYaml customtypes.NormalizedYamlValue `tfsdk:"workflow_configuration_yaml"`
}

var _ entitycore.WithVersionRequirements = toolDataSourceModel{}

func (model *toolModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics
	var d diag.Diagnostics

	spaceID := model.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.ToolID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.Type = types.StringValue(data.Type)
	model.Description = typeutils.NonEmptyStringOrNull(data.Description)

	model.Tags, d = typeutils.StringSetOrNull(ctx, data.Tags)
	diags.Append(d...)

	jsonStr, ok, d := marshalToolConfigurationJSON(data.Configuration)
	diags.Append(d...)
	if ok {
		model.Configuration = jsontypes.NewNormalizedValue(jsonStr)
	} else {
		model.Configuration = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (model *toolDataSourceModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	diags := model.toolModel.populateFromAPI(ctx, data)
	model.ReadOnly = types.BoolValue(data.ReadOnly)
	return diags
}

// marshalToolConfigurationJSON marshals a Tool's configuration map to a JSON
// string. Returns ("", false, nil) when config is nil.
func marshalToolConfigurationJSON(config any) (string, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if config == nil {
		return "", false, diags
	}
	b, err := json.Marshal(config)
	if err != nil {
		diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
		return "", false, diags
	}
	return string(b), true, diags
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
