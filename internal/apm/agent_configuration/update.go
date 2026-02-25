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

package agentconfiguration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *resourceAgentConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AgentConfiguration
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibana, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	settings := make(map[string]string)
	resp.Diagnostics.Append(plan.Settings.ElementsAs(ctx, &settings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agentConfig := kbapi.CreateUpdateAgentConfigurationJSONRequestBody{
		AgentName: plan.AgentName.ValueStringPointer(),
		Service: kbapi.APMUIServiceObject{
			Name:        plan.ServiceName.ValueStringPointer(),
			Environment: plan.ServiceEnvironment.ValueStringPointer(),
		},
		Settings: settings,
	}

	overwrite := true
	params := &kbapi.CreateUpdateAgentConfigurationParams{
		Overwrite:         &overwrite,
		ElasticApiVersion: elasticAPIVersion,
	}

	apiResp, err := kibana.API.CreateUpdateAgentConfiguration(ctx, params, agentConfig)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update APM agent configuration", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if diags := diagutil.CheckHTTPErrorFromFW(apiResp, "Failed to update APM agent configuration"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updatedState, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Updated APM agent configuration with ID: %s", updatedState.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
}
