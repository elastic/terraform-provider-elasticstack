package tool

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/export_ab"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oapiClient, err := d.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	spaceId := "default"
	if !config.SpaceID.IsNull() && !config.SpaceID.IsUnknown() {
		spaceId = config.SpaceID.ValueString()
	}

	toolId := config.ID.ValueString()

	tool := export_ab.FetchTool(ctx, oapiClient.API, toolId, &resp.Diagnostics)
	if tool == nil {
		resp.Diagnostics.AddError("Tool not found", fmt.Sprintf("Unable to fetch tool with ID %s", toolId))
		return
	}

	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: toolId}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceId)
	state.ToolID = tool.ID
	state.Type = tool.Type
	state.Description = tool.Description
	state.Tags = tool.Tags
	state.ReadOnly = tool.ReadOnly
	state.Configuration = tool.Configuration

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
