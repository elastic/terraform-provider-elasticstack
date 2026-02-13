package workflow

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

	workflowId := config.ID.ValueString()

	workflow := export_ab.FetchWorkflow(ctx, oapiClient.API, workflowId, &resp.Diagnostics)
	if workflow == nil {
		resp.Diagnostics.AddError("Workflow not found", fmt.Sprintf("Unable to fetch workflow with ID %s", workflowId))
		return
	}

	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: workflowId}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceId)
	state.WorkflowID = types.StringValue(workflow.ID.ValueString())
	state.Yaml = types.StringValue(workflow.Yaml.ValueString())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
