package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func (r *datafeedResource) read(ctx context.Context, model *Datafeed) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if !r.resourceReady(&diags) {
		return false, diags
	}

	datafeedID := model.DatafeedID.ValueString()
	if datafeedID == "" {
		diags.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return false, diags
	}

	// Get the datafeed from Elasticsearch
	apiModel, getDiags := elasticsearch.GetDatafeed(ctx, r.client, datafeedID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}

	if apiModel == nil {
		// Datafeed not found
		return false, diags
	}

	// Convert API model to TF model
	convertDiags := model.FromAPIModel(ctx, apiModel)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return false, diags
	}

	return true, diags
}
