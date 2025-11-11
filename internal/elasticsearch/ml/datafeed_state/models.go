package datafeed_state

import (
	"strconv"
	"time"

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

func (d *MLDatafeedStateData) GetStartAsString() (string, diag.Diagnostics) {
	return d.getTimeAttributeAsString(d.Start)
}

func (d *MLDatafeedStateData) GetEndAsString() (string, diag.Diagnostics) {
	return d.getTimeAttributeAsString(d.End)
}

func (d *MLDatafeedStateData) getTimeAttributeAsString(val timetypes.RFC3339) (string, diag.Diagnostics) {
	if !utils.IsKnown(val) {
		return "", nil
	}

	valTime, diags := val.ValueRFC3339Time()
	if diags.HasError() {
		return "", diags
	}
	return strconv.FormatInt(valTime.Unix(), 10), nil
}

func (d *MLDatafeedStateData) SetStartAndEndFromAPI(datafeedStats *models.DatafeedStats) diag.Diagnostics {
	var diags diag.Diagnostics
	if datafeedStats.RunningState == nil {
		diags.AddWarning("Running state was empty for a started datafeed", "The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring start and end response values.")
		return diags
	}

	d.Start = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.StartMS))
	if datafeedStats.RunningState.RealTimeConfigured {
		d.End = timetypes.NewRFC3339Null()
	} else {
		d.End = timetypes.NewRFC3339TimeValue(time.UnixMilli(datafeedStats.RunningState.SearchInterval.EndMS))
	}

	return diags
}
