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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, sdkDiags := d.client.EnforceMinVersion(ctx, minKibanaAgentBuilderAPIVersion)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !supported {
		resp.Diagnostics.AddError("Unsupported server version",
			fmt.Sprintf("Agent Builder workflows require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion))
		return
	}

	oapiClient, err := d.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	spaceID := "default"
	if typeutils.IsKnown(config.SpaceID) {
		spaceID = config.SpaceID.ValueString()
	}

	workflowID := config.ID.ValueString()
	if compID, diags := clients.CompositeIDFromStrFw(workflowID); !diags.HasError() {
		workflowID = compID.ResourceID
		if !typeutils.IsKnown(config.SpaceID) {
			spaceID = compID.ClusterID
		}
	}

	workflow, diags := kibanaoapi.GetWorkflow(ctx, oapiClient, spaceID, workflowID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if workflow == nil {
		resp.Diagnostics.AddError("Workflow not found", fmt.Sprintf("Unable to fetch workflow with ID %s", workflowID))
		return
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: workflowID}

	var state workflowDataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceID)
	state.WorkflowID = types.StringValue(workflow.ID)
	state.ConfigurationYaml = customtypes.NewNormalizedYamlValue(workflow.Yaml)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
