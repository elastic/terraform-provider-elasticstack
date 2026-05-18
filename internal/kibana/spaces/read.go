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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
		spaceState := SpaceModel{
			ID:          types.StringValue(space.Id),
			Name:        types.StringValue(space.Name),
			Description: types.StringPointerValue(space.Description),
			Initials:    types.StringPointerValue(space.Initials),
			Color:       types.StringPointerValue(space.Color),
			ImageURL:    types.StringPointerValue(space.ImageUrl),
			Solution:    types.StringPointerValue(space.Solution),
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

// fetchSpace loads a single space by ID. It returns (nil, false, nil) when the space is not found (HTTP 404).
func fetchSpace(ctx context.Context, oapiClient *kibanaoapi.Client, spaceID string) (*kbapi.SpaceResponse, bool, diag.Diagnostics) {
	space, fwDiags := kibanaoapi.GetSpace(ctx, oapiClient, spaceID)
	if fwDiags.HasError() {
		return nil, false, fwDiags
	}
	if space == nil {
		return nil, false, nil
	}
	return space, true, nil
}

// readSpaceResource is the resource read callback for a single Kibana space.
func readSpaceResource(ctx context.Context, client *clients.KibanaScopedClient, resourceID, _ string, model resourceModel) (resourceModel, bool, diag.Diagnostics) {
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("unable to get Kibana OpenAPI client", err.Error())
		return model, false, diags
	}

	space, found, fwDiags := fetchSpace(ctx, oapiClient, resourceID)
	if fwDiags.HasError() {
		return model, false, fwDiags
	}
	if !found {
		return model, false, nil
	}

	updated, mapDiags := mapSpaceResponseToResourceModel(ctx, space, resourceID)
	if mapDiags.HasError() {
		return model, false, mapDiags
	}
	updated.KibanaConnectionField = model.KibanaConnectionField
	updated.ImageURL = model.ImageURL
	return updated, true, mapDiags
}

func mapSpaceResponseToResourceModel(ctx context.Context, space *kbapi.SpaceResponse, resourceID string) (resourceModel, diag.Diagnostics) {
	var m resourceModel
	var diags diag.Diagnostics

	m.ID = types.StringValue(resourceID)
	m.SpaceID = types.StringValue(resourceID)
	m.Name = types.StringValue(space.Name)

	m.Description = types.StringPointerValue(space.Description)
	m.Initials = types.StringPointerValue(space.Initials)
	m.Color = types.StringPointerValue(space.Color)
	m.Solution = types.StringPointerValue(space.Solution)

	rawFeatures := []string{}
	if space.DisabledFeatures != nil {
		rawFeatures = *space.DisabledFeatures
	}
	setVal, d := types.SetValueFrom(ctx, types.StringType, rawFeatures)
	diags.Append(d...)
	if d.HasError() {
		return m, diags
	}
	m.DisabledFeatures = setVal

	return m, diags
}

func finalizeResourceModelFromAPIResponse(ctx context.Context, plan resourceModel, space *kbapi.SpaceResponse) (resourceModel, diag.Diagnostics) {
	out, diags := mapSpaceResponseToResourceModel(ctx, space, plan.SpaceID.ValueString())
	if diags.HasError() {
		return out, diags
	}
	out.KibanaConnectionField = plan.KibanaConnectionField
	out.ImageURL = plan.ImageURL
	return out, diags
}
