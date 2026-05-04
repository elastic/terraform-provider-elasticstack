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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readDataSource is the envelope read callback for the spaces data source.
func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, err := kbClient.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("unable to get Kibana OpenAPI client", err.Error())
		return config, diags
	}

	// Call client API
	spaces, fwDiags := kibanaoapi.ListSpaces(ctx, oapiClient)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return config, diags
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
		disabledFeatures, d := types.ListValueFrom(ctx, types.StringType, rawFeatures)
		if d.HasError() {
			diags.Append(d...)
			return config, diags
		}

		spaceState.DisabledFeatures = disabledFeatures

		config.Spaces = append(config.Spaces, spaceState)
	}

	config.ID = types.StringValue("spaces")

	return config, diags
}
