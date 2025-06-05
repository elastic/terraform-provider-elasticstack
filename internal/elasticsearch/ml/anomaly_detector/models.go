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
		"detector_description": types.StringType,
		"use_null":             types.BoolType,
		"exclude_frequent":     types.StringType,
		"custom_rules":         types.ListType{ElemType: types.ObjectType{AttrTypes: customRuleModel{}.attrTypes(ctx, diags)}},
	}
}

func (m analysisConfigModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	detectorAttrTypes := detectorModel{}.attrTypes(ctx, diags)
	if diags.HasError() {
		return nil
	}
	return map[string]attr.Type{
		"bucket_span":               types.StringType,
		"detectors":                 types.ListType{ElemType: types.ObjectType{AttrTypes: detectorAttrTypes}},
		"influencers":               types.ListType{ElemType: types.StringType},
		"categorization_field_name": types.StringType,
		"summary_count_field_name":  types.StringType,
		"latency":                   types.StringType,
	}
}

func (m dataDescriptionModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"time_field":  types.StringType,
		"time_format": types.StringType,
	}
}

func (m modelPlotConfigModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

type analysisLimitsModel struct {
	ModelMemoryLimit          types.String `tfsdk:"model_memory_limit"`
	CategorizationExamplesLimit types.Int64  `tfsdk:"categorization_examples_limit"`
}

func (m analysisLimitsModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"model_memory_limit":            types.StringType,
		"categorization_examples_limit": types.Int64Type,
	}
}

type anomalyDetectorResourceModel struct {
	ID                                   types.String `tfsdk:"id"`
	JobId                                types.String `tfsdk:"job_id"`
	Description                          types.String `tfsdk:"description"`
	Groups                               types.List   `tfsdk:"groups"`
	AnalysisConfig                       types.Object `tfsdk:"analysis_config"`
	DataDescription                      types.Object `tfsdk:"data_description"`
	ModelPlotConfig                      types.Object `tfsdk:"model_plot_config"`
	ResultsIndexName                     types.String `tfsdk:"results_index_name"`
	AnalysisLimits                       types.Object `tfsdk:"analysis_limits"`
	ModelSnapshotRetentionDays           types.Int64  `tfsdk:"model_snapshot_retention_days"`
	ResultsRetentionDays                 types.Int64  `tfsdk:"results_retention_days"`
	AllowLazyOpen                        types.Bool   `tfsdk:"allow_lazy_open"`
	DailyModelSnapshotRetentionAfterDays types.Int64  `tfsdk:"daily_model_snapshot_retention_after_days"`
	CustomSettings                       types.Map    `tfsdk:"custom_settings"`
}

type analysisConfigModel struct {
	BucketSpan              types.String `tfsdk:"bucket_span"`
	Detectors               types.List   `tfsdk:"detectors"`
	Influencers             types.List   `tfsdk:"influencers"`
	CategorizationFieldName types.String `tfsdk:"categorization_field_name"`
	SummaryCountFieldName   types.String `tfsdk:"summary_count_field_name"`
	Latency                 types.String `tfsdk:"latency"`
}

type detectorModel struct {
	Function            types.String `tfsdk:"function"`
	FieldName           types.String `tfsdk:"field_name"`
	ByFieldName         types.String `tfsdk:"by_field_name"`
	PartitionFieldName  types.String `tfsdk:"partition_field_name"`
	DetectorDescription types.String `tfsdk:"detector_description"`
	UseNull             types.Bool   `tfsdk:"use_null"`
	ExcludeFrequent     types.String `tfsdk:"exclude_frequent"`
	CustomRules         types.List   `tfsdk:"custom_rules"`
}

type dataDescriptionModel struct {
	TimeField  types.String `tfsdk:"time_field"`
	TimeFormat types.String `tfsdk:"time_format"`
}

type ruleConditionModel struct {
	Operator types.String  `tfsdk:"operator"`
	Value    types.Float64 `tfsdk:"value"`
}

func (m ruleConditionModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	return map[string]attr.Type{
		"operator": types.StringType,
		"value":    types.Float64Type,
	}
}

type customRuleModel struct {
	Actions    types.List   `tfsdk:"actions"`
	Scope      types.String `tfsdk:"scope"`
	Conditions types.List   `tfsdk:"conditions"`
}

func (m customRuleModel) attrTypes(ctx context.Context, diags diag.Diagnostics) map[string]attr.Type {
	conditionAttrTypes := ruleConditionModel{}.attrTypes(ctx, diags)
	if diags.HasError() {
		return nil
	}
	return map[string]attr.Type{
		"actions":    types.ListType{ElemType: types.StringType},
		"scope":      types.StringType,
		"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: conditionAttrTypes}},
	}
}

type modelPlotConfigModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
