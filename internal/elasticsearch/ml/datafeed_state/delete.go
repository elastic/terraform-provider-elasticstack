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

package datafeedstate

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlDatafeedStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MLDatafeedStateData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, fwDiags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := data.DatafeedID.ValueString()
	currentState, fwDiags := datafeed.GetDatafeedState(ctx, client, datafeedID)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState == nil {
		// Datafeed already doesn't exist, nothing to do
		tflog.Info(ctx, fmt.Sprintf("ML datafeed %s not found during delete", datafeedID))
		return
	}

	// If the datafeed is started, stop it when deleting the resource
	if *currentState == datafeed.StateStarted {
		tflog.Info(ctx, fmt.Sprintf("Stopping ML datafeed %s during delete", datafeedID))

		// Parse timeout duration
		timeout, parseErrs := data.Timeout.Parse()
		resp.Diagnostics.Append(parseErrs...)
		if resp.Diagnostics.HasError() {
			return
		}

		force := data.Force.ValueBool()
		fwDiags = elasticsearch.StopDatafeed(ctx, client, datafeedID, force, timeout)
		resp.Diagnostics.Append(fwDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Wait for the datafeed to stop
		_, diags := datafeed.WaitForDatafeedState(ctx, client, datafeedID, datafeed.StateStopped)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Info(ctx, fmt.Sprintf("ML datafeed %s successfully stopped during delete", datafeedID))
	}
}
