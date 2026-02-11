package output

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *outputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config outputsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	requestSpaceID := ""
	if utils.IsKnown(config.SpaceID) {
		requestSpaceID = config.SpaceID.ValueString()
	}

	outputs, diags := fleet.GetOutputs(ctx, client, requestSpaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := outputsDataSourceModel{
		OutputID:            config.OutputID,
		Type:                config.Type,
		DefaultIntegrations: config.DefaultIntegrations,
		DefaultMonitoring:   config.DefaultMonitoring,
	}

	if utils.IsKnown(config.SpaceID) {
		state.SpaceID = config.SpaceID
	} else {
		state.SpaceID = types.StringNull()
	}

	resp.Diagnostics.Append(state.populateFromAPI(ctx, outputs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filterKey := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s",
		client.URL,
		stringFilterValue(config.OutputID),
		stringFilterValue(config.Type),
		boolFilterValue(config.DefaultIntegrations),
		boolFilterValue(config.DefaultMonitoring),
		requestSpaceID,
	)
	stateID, err := utils.StringToHash(filterKey)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
	state.ID = types.StringPointerValue(stateID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func stringFilterValue(value types.String) string {
	if !utils.IsKnown(value) {
		return ""
	}

	return value.ValueString()
}

func boolFilterValue(value types.Bool) string {
	if !utils.IsKnown(value) {
		return ""
	}

	return strconv.FormatBool(value.ValueBool())
}
