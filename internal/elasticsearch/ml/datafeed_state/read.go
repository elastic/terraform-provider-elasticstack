package datafeed_state

import (
	"context"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *mlDatafeedStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MLDatafeedStateData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readData, diags := r.read(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

func (r *mlDatafeedStateResource) read(ctx context.Context, data MLDatafeedStateData) (*MLDatafeedStateData, diag.Diagnostics) {
	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	if diags.HasError() {
		return nil, diags
	}

	datafeedId := data.DatafeedId.ValueString()
	// Check if the datafeed exists by getting its stats
	datafeedStats, getDiags := elasticsearch.GetDatafeedStats(ctx, client, datafeedId)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if datafeedStats == nil {
		return nil, diags
	}

	// Update the data with current information
	data.State = types.StringValue(datafeedStats.State)

	if datafeedStats.State == "started" {
		if datafeedStats.RunningState == nil {
			diags.AddWarning("Running state was empty for a started datafeed", "The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring start and end response values.")
		}

		data.Start = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.StartMS))
		if datafeedStats.RunningState.RealTimeConfigured {
			data.End = timetypes.NewRFC3339Null()
		} else {
			data.End = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.EndMS))
		}
	}

	return &data, diags
}
