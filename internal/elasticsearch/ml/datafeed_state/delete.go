package datafeed_state

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

	client, fwDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedId := data.DatafeedId.ValueString()
	currentState, fwDiags := datafeed.GetDatafeedState(ctx, client, datafeedId)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState == nil {
		// Datafeed already doesn't exist, nothing to do
		tflog.Info(ctx, fmt.Sprintf("ML datafeed %s not found during delete", datafeedId))
		return
	}

	// If the datafeed is started, stop it when deleting the resource
	if *currentState == "started" {
		tflog.Info(ctx, fmt.Sprintf("Stopping ML datafeed %s during delete", datafeedId))

		// Parse timeout duration
		timeout, parseErrs := data.Timeout.Parse()
		resp.Diagnostics.Append(parseErrs...)
		if resp.Diagnostics.HasError() {
			return
		}

		force := data.Force.ValueBool()
		fwDiags = elasticsearch.StopDatafeed(ctx, client, datafeedId, force, timeout)
		resp.Diagnostics.Append(fwDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Wait for the datafeed to stop
		_, diags := datafeed.WaitForDatafeedState(ctx, client, datafeedId, "stopped")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Info(ctx, fmt.Sprintf("ML datafeed %s successfully stopped during delete", datafeedId))
	}
}
