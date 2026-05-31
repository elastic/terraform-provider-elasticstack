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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description: "Reads an Agent Builder workflow by ID. See https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The workflow ID to look up.",
				Required:    true,
			},
			"space_id": dsschema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
			},
			"workflow_id": dsschema.StringAttribute{
				Description: "The ID of the workflow.",
				Computed:    true,
			},
			"configuration_yaml": dsschema.StringAttribute{
				Description: "The workflow definition in YAML format.",
				Computed:    true,
				CustomType:  customtypes.NormalizedYamlType{},
			},
		},
	}
}

func readWorkflowDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID, spaceID string,
	config workflowDataSourceModel,
) (workflowDataSourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !agentbuilder.EnforceVersion(ctx, client, minKibanaAgentBuilderAPIVersion, "workflows", &diags) {
		return config, false, diags
	}

	if spaceID == "" {
		spaceID = "default"
	}

	oapiClient := client.GetKibanaOapiClient()

	workflow, d := kibanaoapi.GetWorkflow(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return config, false, diags
	}
	if workflow == nil {
		return config, false, diags
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: workflow.ID}

	config.ID = types.StringValue(compositeID.String())
	config.SpaceID = types.StringValue(spaceID)
	config.WorkflowID = types.StringValue(workflow.ID)
	config.ConfigurationYaml = customtypes.NewNormalizedYamlValue(workflow.Yaml)

	return config, true, diags
}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return entitycore.NewKibanaDataSource[workflowDataSourceModel](
		entitycore.ComponentKibana,
		"agentbuilder_workflow",
		entitycore.KibanaDataSourceOptions[workflowDataSourceModel]{
			Schema: getDataSourceSchema,
			Read:   readWorkflowDataSource,
		},
	)
}
