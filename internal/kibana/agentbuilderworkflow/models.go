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

package agentbuilderworkflow

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// workflowDataSourceModel is the core model shared between the workflow data
// source and resource. The data source exposes exactly these fields, while the
// resource model (workflowModel) embeds it and adds the computed attributes
// derived from the YAML definition. Defining the version gate and the
// kibana_connection block (via KibanaConnectionField) here means they are
// declared exactly once rather than duplicated across both models.
type workflowDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID                types.String                    `tfsdk:"id"`
	SpaceID           types.String                    `tfsdk:"space_id"`
	WorkflowID        types.String                    `tfsdk:"workflow_id"`
	ConfigurationYaml customtypes.NormalizedYamlValue `tfsdk:"configuration_yaml"`
}

func (workflowDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *minKibanaAgentBuilderAPIVersion,
		ErrorMessage: fmt.Sprintf("Agent Builder workflows require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion),
	}}, nil
}

var _ entitycore.WithVersionRequirements = workflowDataSourceModel{}

type workflowModel struct {
	workflowDataSourceModel
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Valid       types.Bool   `tfsdk:"valid"`
}

func (model workflowModel) GetID() types.String         { return model.ID }
func (model workflowModel) GetResourceID() types.String { return model.WorkflowID }
func (model workflowModel) GetSpaceID() types.String    { return model.SpaceID }

var _ entitycore.KibanaResourceModel = workflowModel{}
var _ entitycore.WithVersionRequirements = workflowModel{}

func (model *workflowModel) populateFromAPI(data *models.Workflow) {
	if data == nil {
		return
	}

	spaceID := model.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.WorkflowID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.ConfigurationYaml = customtypes.NewNormalizedYamlValue(data.Yaml)
	model.Name = types.StringValue(data.Name)

	model.Description = typeutils.NonEmptyStringOrNull(data.Description)

	model.Enabled = types.BoolValue(data.Enabled)
	model.Valid = types.BoolValue(data.Valid)
}

func (model workflowModel) toAPICreateModel() kbapi.PostWorkflowsWorkflowJSONRequestBody {
	body := kbapi.PostWorkflowsWorkflowJSONRequestBody{
		Yaml: model.ConfigurationYaml.ValueString(),
	}

	if typeutils.IsKnown(model.WorkflowID) {
		ID := model.WorkflowID.ValueString()
		body.Id = &ID
	}

	return body
}

func (model workflowModel) toAPIUpdateModel() kbapi.PutWorkflowsWorkflowIdJSONRequestBody {
	yaml := model.ConfigurationYaml.ValueString()
	return kbapi.PutWorkflowsWorkflowIdJSONRequestBody{
		Yaml: &yaml,
	}
}
