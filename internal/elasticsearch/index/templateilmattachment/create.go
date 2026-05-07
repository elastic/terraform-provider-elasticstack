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

package templateilmattachment

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tfModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, fwDiags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	if serverVersion.LessThan(MinVersion) {
		resp.Diagnostics.AddError(
			"Unsupported Elasticsearch Version",
			fmt.Sprintf(
				"This resource requires Elasticsearch %s or later (current: %s). "+
					"This resource is not supported on this Elasticsearch version.",
				MinVersion, serverVersion,
			),
		)
		return
	}

	componentTemplateName := plan.getComponentTemplateName()
	updatedPlan, writeDiags := writeILMAttachment(ctx, client, componentTemplateName, plan, true)
	resp.Diagnostics.Append(writeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updatedPlan)...)
}
