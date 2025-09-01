package spaces

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			ImageUrl:    types.StringValue(space.ImageURL),
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
