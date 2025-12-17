package datafeed_state

import (
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
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

func (d *MLDatafeedStateData) SetStartAndEndFromAPI(datafeedStats *models.DatafeedStats) diag.Diagnostics {
	var diags diag.Diagnostics

	if datafeed.State(datafeedStats.State) == datafeed.StateStarted {
		if datafeedStats.RunningState == nil {
			diags.AddWarning("Running state was empty for a started datafeed", "The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring start and end response values.")
			return diags
		}

		if datafeedStats.RunningState.SearchInterval != nil {
			d.Start = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.StartMS))
			d.End = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.EndMS))
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
