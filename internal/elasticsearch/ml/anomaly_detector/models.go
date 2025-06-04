package anomaly_detector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m detectorModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"function":             types.StringType,
		"field_name":           types.StringType,
		"by_field_name":        types.StringType,
		"partition_field_name": types.StringType,
	}
}

func (m analysisConfigModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	detectorAttrTypes := detectorModel{}.attrTypes(ctx, diags)
	if diags.HasError() {
		return nil
	}
	return map[string]attr.Type{
		"bucket_span": types.StringType,
		"detectors":   types.ListType{ElemType: types.ObjectType{AttrTypes: detectorAttrTypes}},
		"influencers": types.ListType{ElemType: types.StringType},
	}
}

func (m dataDescriptionModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"time_field": types.StringType,
	}
}

func (m modelPlotConfigModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

type anomalyDetectorResourceModel struct {
	ID               types.String `tfsdk:"id"`
	JobId            types.String `tfsdk:"job_id"`
	Description      types.String `tfsdk:"description"`
	Groups           types.List   `tfsdk:"groups"`
	AnalysisConfig   types.Object `tfsdk:"analysis_config"`
	DataDescription  types.Object `tfsdk:"data_description"`
	ModelPlotConfig  types.Object `tfsdk:"model_plot_config"`
	ResultsIndexName types.String `tfsdk:"results_index_name"`
}

type analysisConfigModel struct {
	BucketSpan  types.String `tfsdk:"bucket_span"`
	Detectors   types.List   `tfsdk:"detectors"`
	Influencers types.List   `tfsdk:"influencers"`
}

type detectorModel struct {
	Function           types.String `tfsdk:"function"`
	FieldName          types.String `tfsdk:"field_name"`
	ByFieldName        types.String `tfsdk:"by_field_name"`
	PartitionFieldName types.String `tfsdk:"partition_field_name"`
}

type dataDescriptionModel struct {
	TimeField types.String `tfsdk:"time_field"`
}

type modelPlotConfigModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
