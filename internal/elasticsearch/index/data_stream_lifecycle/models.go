package data_stream_lifecycle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	DataRetention           types.String `tfsdk:"data_retention"`
	ExpandWildcards         types.String `tfsdk:"expand_wildcards"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	Downsampling            types.List   `tfsdk:"downsampling"`
}

type downsamplingTfModel struct {
	After         types.String `tfsdk:"after"`
	FixedInterval types.String `tfsdk:"fixed_interval"`
}

func (model tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, utils.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compId, nil
}

func (model tfModel) toAPIModel(ctx context.Context) (models.LifecycleSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiModel := models.LifecycleSettings{
		DataRetention: model.DataRetention.ValueString(),
		Enabled:       model.Enabled.ValueBool(),
	}

	if !model.Downsampling.IsNull() && !model.Downsampling.IsUnknown() && len(model.Downsampling.Elements()) > 0 {

		downsampling := make([]downsamplingTfModel, len(model.Downsampling.Elements()))
		if diags := model.Downsampling.ElementsAs(ctx, &downsampling, true); diags.HasError() {
			return models.LifecycleSettings{}, diags
		}

		apiModel.Downsampling = make([]models.Downsampling, len(model.Downsampling.Elements()))
		for i, ds := range downsampling {
			apiModel.Downsampling[i] = models.Downsampling{
				After:         ds.After.ValueString(),
				FixedInterval: ds.FixedInterval.ValueString(),
			}
		}
	}

	return apiModel, diags
}

func (model *tfModel) populateFromAPI(ctx context.Context, ds []models.DataStreamLifecycle) diag.Diagnostics {
	actualRetention := model.DataRetention.ValueString()
	actualDownsampling := make([]downsamplingTfModel, len(model.Downsampling.Elements()))
	if diags := model.Downsampling.ElementsAs(ctx, &actualDownsampling, true); diags.HasError() {
		return nil
	}

	for _, lf := range ds {
		if lf.Lifecycle.DataRetention != actualRetention {
			model.DataRetention = types.StringValue(lf.Lifecycle.DataRetention)
		}
		var updateDownsampling bool
		if len(lf.Lifecycle.Downsampling) != len(actualDownsampling) {
			updateDownsampling = true
		} else {
			for i, ds := range actualDownsampling {
				if ds.After.ValueString() != lf.Lifecycle.Downsampling[i].After || ds.FixedInterval.ValueString() != lf.Lifecycle.Downsampling[i].FixedInterval {
					updateDownsampling = true
					break
				}
			}
		}
		if updateDownsampling {
			listValue, diags := convertDownsamplingToModel(ctx, lf.Lifecycle.Downsampling)
			diags.Append(diags...)
			if diags.HasError() {
				return diags
			}
			model.Downsampling = listValue
		}
	}
	return nil
}

func convertDownsamplingToModel(ctx context.Context, apiDownsamplings []models.Downsampling) (types.List, diag.Diagnostics) {
	var downsamplings []downsamplingTfModel

	for _, apiDs := range apiDownsamplings {
		downsamplings = append(downsamplings, downsamplingTfModel{
			After:         types.StringValue(apiDs.After),
			FixedInterval: types.StringValue(apiDs.FixedInterval),
		})
	}

	listValue, diags := types.ListValueFrom(ctx, downsamplingElementType(), downsamplings)

	return listValue, diags
}
