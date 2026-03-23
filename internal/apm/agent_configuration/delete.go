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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *resourceAgentConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AgentConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibana, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	idParts := strings.Split(state.ID.ValueString(), ":")
	serviceName := idParts[0]
	var serviceEnv *string
	if len(idParts) > 1 {
		serviceEnv = &idParts[1]
	}

	deleteReqBody := kbapi.APMUIDeleteServiceObject{
		Service: kbapi.APMUIServiceObject{
			Name:        &serviceName,
			Environment: serviceEnv,
		},
	}
	apiResp, err := kibana.API.DeleteAgentConfiguration(
		ctx,
		&kbapi.DeleteAgentConfigurationParams{
			ElasticApiVersion: elasticAPIVersion,
		},
		deleteReqBody,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete APM agent configuration", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if diags := diagutil.CheckHTTPErrorFromFW(apiResp, "Failed to delete APM agent configuration"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted APM agent configuration with ID: %s", state.ID.ValueString()))
}
