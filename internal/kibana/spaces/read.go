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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataSourceModel

	// Call client API
	spaces, err := d.client.List()
	if err != nil {
		resp.Diagnostics.AddError("unable to list spaces", err.Error())
		return
	}

	// Map response body to model
	for _, space := range spaces {
		spaceState := model{
			ID:          types.StringValue(space.ID),
			Name:        types.StringValue(space.Name),
			Description: types.StringValue(space.Description),
			Initials:    types.StringValue(space.Initials),
			Color:       types.StringValue(space.Color),
			ImageURL:    types.StringValue(space.ImageURL),
			Solution:    types.StringValue(space.Solution),
		}

		disabledFeatures, diags := types.ListValueFrom(ctx, types.StringType, space.DisabledFeatures)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		spaceState.DisabledFeatures = disabledFeatures

		state.Spaces = append(state.Spaces, spaceState)
	}

	state.ID = types.StringValue("spaces")

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
