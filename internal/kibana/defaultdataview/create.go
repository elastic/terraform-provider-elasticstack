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

package defaultdataview

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.setDefaultDataView(ctx, req.Plan, &resp.State)...)
}

// setDefaultDataView is a helper method that contains the core logic for setting the default data view.
func (r *Resource) setDefaultDataView(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model defaultDataViewModel
	diags := plan.Get(ctx, &model)
	if diags.HasError() {
		return diags
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("unable to get kibana client", err.Error())
		return diags
	}

	dataViewID := model.DataViewID.ValueStringPointer()
	force := model.Force.ValueBool()
	spaceID := model.SpaceID.ValueString()
	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		DataViewId: dataViewID,
		Force:      &force,
	}

	apiDiags := kibanaoapi.SetDefaultDataView(ctx, client, spaceID, setReq)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return diags
	}

	model, readDiags := r.read(ctx, client, model)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	diags = state.Set(ctx, model)
	return diags
}
