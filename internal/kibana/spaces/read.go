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

package spaces

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, diags := d.Client().GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oapiClient, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana OpenAPI client", err.Error())
		return
	}

	// Call client API
	spaces, fwDiags := kibanaoapi.ListSpaces(ctx, oapiClient)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to model
	for _, space := range spaces {
		spaceState := model{
			ID:   types.StringValue(space.Id),
			Name: types.StringValue(space.Name),
		}

		if space.Description != nil {
			spaceState.Description = types.StringValue(*space.Description)
		} else {
			spaceState.Description = types.StringValue("")
		}

		if space.Initials != nil {
			spaceState.Initials = types.StringValue(*space.Initials)
		} else {
			spaceState.Initials = types.StringValue("")
		}

		if space.Color != nil {
			spaceState.Color = types.StringValue(*space.Color)
		} else {
			spaceState.Color = types.StringValue("")
		}

		if space.ImageUrl != nil {
			spaceState.ImageURL = types.StringValue(*space.ImageUrl)
		} else {
			spaceState.ImageURL = types.StringValue("")
		}

		if space.Solution != nil {
			spaceState.Solution = types.StringValue(*space.Solution)
		} else {
			spaceState.Solution = types.StringValue("")
		}

		rawFeatures := []string{}
		if space.DisabledFeatures != nil {
			rawFeatures = *space.DisabledFeatures
		}
		disabledFeatures, diags := types.ListValueFrom(ctx, types.StringType, rawFeatures)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		spaceState.DisabledFeatures = disabledFeatures

		state.Spaces = append(state.Spaces, spaceState)
	}

	state.ID = types.StringValue("spaces")

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
