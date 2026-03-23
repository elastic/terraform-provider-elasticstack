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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workflowModel struct {
	ID                types.String                    `tfsdk:"id"`
	WorkflowID        types.String                    `tfsdk:"workflow_id"`
	SpaceID           types.String                    `tfsdk:"space_id"`
	ConfigurationYaml customtypes.NormalizedYamlValue `tfsdk:"configuration_yaml"`
	Name              types.String                    `tfsdk:"name"`
	Description       types.String                    `tfsdk:"description"`
	Enabled           types.Bool                      `tfsdk:"enabled"`
	Valid             types.Bool                      `tfsdk:"valid"`
}

func (model *workflowModel) populateFromAPI(data *kbapi.WorkflowDetailDto) {
	if data == nil {
		return
	}

	spaceID := model.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = "default"
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.Id}).String())
	model.WorkflowID = types.StringValue(data.Id)
	model.SpaceID = types.StringValue(spaceID)
	model.ConfigurationYaml = customtypes.NewNormalizedYamlValue(data.Yaml)
	model.Name = types.StringValue(data.Name)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	model.Enabled = types.BoolValue(data.Enabled)
	model.Valid = types.BoolValue(data.Valid)
}

func (model workflowModel) toAPICreateModel() kbapi.CreateWorkflowCommand {
	body := kbapi.CreateWorkflowCommand{
		Yaml: model.ConfigurationYaml.ValueString(),
	}

	if typeutils.IsKnown(model.WorkflowID) {
		id := model.WorkflowID.ValueString()
		body.Id = &id
	}

	return body
}

func (model workflowModel) toAPIUpdateModel() kbapi.UpdateWorkflowCommand {
	yaml := model.ConfigurationYaml.ValueString()
	return kbapi.UpdateWorkflowCommand{
		Yaml: &yaml,
	}
}
