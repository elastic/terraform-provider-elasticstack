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

package workflow

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/exportagentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := d.client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if serverVersion.LessThan(minKibanaAgentBuilderAPIVersion) {
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

	workflow := exportagentbuilder.FetchWorkflow(ctx, oapiClient.API, workflowID, &resp.Diagnostics)
	if workflow == nil {
		resp.Diagnostics.AddError("Workflow not found", fmt.Sprintf("Unable to fetch workflow with ID %s", workflowID))
		return
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: workflowID}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceID)
	state.WorkflowID = types.StringValue(workflow.ID.ValueString())
	state.Yaml = types.StringValue(workflow.Yaml.ValueString())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
