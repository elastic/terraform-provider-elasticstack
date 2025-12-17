package datafeed_state

import (
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MLDatafeedStateData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	DatafeedId              types.String         `tfsdk:"datafeed_id"`
	State                   types.String         `tfsdk:"state"`
	Force                   types.Bool           `tfsdk:"force"`
	Timeout                 customtypes.Duration `tfsdk:"datafeed_timeout"`
	Start                   timetypes.RFC3339    `tfsdk:"start"`
	End                     timetypes.RFC3339    `tfsdk:"end"`
	Timeouts                timeouts.Value       `tfsdk:"timeouts"`
}

func timeInSameLocation(ms int64, source timetypes.RFC3339) (time.Time, diag.Diagnostics) {
	t := time.UnixMilli(ms)
	if !utils.IsKnown(source) {
		return t, nil
	}

	sourceTime, diags := source.ValueRFC3339Time()
	if diags.HasError() {
		return t, diags
	}

	t = t.In(sourceTime.Location())
	return t, nil
}

func (d *MLDatafeedStateData) SetStartAndEndFromAPI(datafeedStats *models.DatafeedStats) diag.Diagnostics {
	var diags diag.Diagnostics

	if datafeed.State(datafeedStats.State) == datafeed.StateStarted {
		if datafeedStats.RunningState == nil {
			diags.AddWarning("Running state was empty for a started datafeed", "The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring start and end response values.")
			return diags
		}

		if datafeedStats.RunningState.SearchInterval != nil {
			start, diags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.StartMS, d.Start)
			if diags.HasError() {
				return diags
			}

			end, diags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.EndMS, d.End)
			if diags.HasError() {
				return diags
			}

			d.Start = timetypes.NewRFC3339TimeValue(start)
			d.End = timetypes.NewRFC3339TimeValue(end)
		}

		if datafeedStats.RunningState.RealTimeConfigured {
			d.End = timetypes.NewRFC3339Null()
		}
	}

	if d.Start.IsUnknown() {
		d.Start = timetypes.NewRFC3339Null()
	}

	if d.End.IsUnknown() {
		d.End = timetypes.NewRFC3339Null()
	}

	return diags
}
