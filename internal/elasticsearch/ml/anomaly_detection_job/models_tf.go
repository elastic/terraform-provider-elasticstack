package anomaly_detection_job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// AnomalyDetectionJobTFModel represents the Terraform resource model for ML anomaly detection jobs
type AnomalyDetectionJobTFModel struct {
	ID                                   types.String         `tfsdk:"id"`
	ElasticsearchConnection              types.List           `tfsdk:"elasticsearch_connection"`
	JobID                                types.String         `tfsdk:"job_id"`
	Description                          types.String         `tfsdk:"description"`
	Groups                               types.Set            `tfsdk:"groups"`
	AnalysisConfig                       types.Object         `tfsdk:"analysis_config"`
	AnalysisLimits                       types.Object         `tfsdk:"analysis_limits"`
	DataDescription                      types.Object         `tfsdk:"data_description"`
	ModelPlotConfig                      types.Object         `tfsdk:"model_plot_config"`
	AllowLazyOpen                        types.Bool           `tfsdk:"allow_lazy_open"`
	BackgroundPersistInterval            types.String         `tfsdk:"background_persist_interval"`
	CustomSettings                       jsontypes.Normalized `tfsdk:"custom_settings"`
	DailyModelSnapshotRetentionAfterDays types.Int64          `tfsdk:"daily_model_snapshot_retention_after_days"`
	ModelSnapshotRetentionDays           types.Int64          `tfsdk:"model_snapshot_retention_days"`
	RenormalizationWindowDays            types.Int64          `tfsdk:"renormalization_window_days"`
	ResultsIndexName                     types.String         `tfsdk:"results_index_name"`
	ResultsRetentionDays                 types.Int64          `tfsdk:"results_retention_days"`

	// Read-only computed fields
	CreateTime      types.String `tfsdk:"create_time"`
	JobType         types.String `tfsdk:"job_type"`
	JobVersion      types.String `tfsdk:"job_version"`
	ModelSnapshotID types.String `tfsdk:"model_snapshot_id"`
}

// AnalysisConfigTFModel represents the analysis configuration
type AnalysisConfigTFModel struct {
	BucketSpan                 types.String `tfsdk:"bucket_span"`
	CategorizationFieldName    types.String `tfsdk:"categorization_field_name"`
	CategorizationFilters      types.List   `tfsdk:"categorization_filters"`
	Detectors                  types.List   `tfsdk:"detectors"`
	Influencers                types.List   `tfsdk:"influencers"`
	Latency                    types.String `tfsdk:"latency"`
	ModelPruneWindow           types.String `tfsdk:"model_prune_window"`
	MultivariateByFields       types.Bool   `tfsdk:"multivariate_by_fields"`
	PerPartitionCategorization types.Object `tfsdk:"per_partition_categorization"`
	SummaryCountFieldName      types.String `tfsdk:"summary_count_field_name"`
}

// DetectorTFModel represents a detector configuration
type DetectorTFModel struct {
	ByFieldName         types.String `tfsdk:"by_field_name"`
	DetectorDescription types.String `tfsdk:"detector_description"`
	ExcludeFrequent     types.String `tfsdk:"exclude_frequent"`
	FieldName           types.String `tfsdk:"field_name"`
	Function            types.String `tfsdk:"function"`
	OverFieldName       types.String `tfsdk:"over_field_name"`
	PartitionFieldName  types.String `tfsdk:"partition_field_name"`
	UseNull             types.Bool   `tfsdk:"use_null"`
	CustomRules         types.List   `tfsdk:"custom_rules"`
}

// CustomRuleTFModel represents a custom rule configuration
type CustomRuleTFModel struct {
	Actions    types.List `tfsdk:"actions"`
	Conditions types.List `tfsdk:"conditions"`
}

// RuleConditionTFModel represents a rule condition
type RuleConditionTFModel struct {
	AppliesTo types.String  `tfsdk:"applies_to"`
	Operator  types.String  `tfsdk:"operator"`
	Value     types.Float64 `tfsdk:"value"`
}

// AnalysisLimitsTFModel represents analysis limits configuration
type AnalysisLimitsTFModel struct {
	CategorizationExamplesLimit types.Int64  `tfsdk:"categorization_examples_limit"`
	ModelMemoryLimit            types.String `tfsdk:"model_memory_limit"`
}

// DataDescriptionTFModel represents data description configuration
type DataDescriptionTFModel struct {
	TimeField  types.String `tfsdk:"time_field"`
	TimeFormat types.String `tfsdk:"time_format"`
}

// ModelPlotConfigTFModel represents model plot configuration
type ModelPlotConfigTFModel struct {
	AnnotationsEnabled types.Bool   `tfsdk:"annotations_enabled"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Terms              types.String `tfsdk:"terms"`
}

// PerPartitionCategorizationTFModel represents per-partition categorization configuration
type PerPartitionCategorizationTFModel struct {
	Enabled    types.Bool `tfsdk:"enabled"`
	StopOnWarn types.Bool `tfsdk:"stop_on_warn"`
}

// ToAPIModel converts TF model to AnomalyDetectionJobAPIModel
func (plan *AnomalyDetectionJobTFModel) toAPIModel(ctx context.Context) (*AnomalyDetectionJobAPIModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	apiModel := &AnomalyDetectionJobAPIModel{
		JobID:       plan.JobID.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Convert groups
	if utils.IsKnown(plan.Groups) {
		var groups []string
		d := plan.Groups.ElementsAs(ctx, &groups, false)
		diags.Append(d...)
		apiModel.Groups = groups
	}

	// Convert analysis_config
	var analysisConfig AnalysisConfigTFModel
	d := plan.AnalysisConfig.As(ctx, &analysisConfig, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Convert detectors
	var detectors []DetectorTFModel
	d = analysisConfig.Detectors.ElementsAs(ctx, &detectors, false)
	diags.Append(d...)

	apiDetectors := make([]DetectorAPIModel, len(detectors))
	for i, detector := range detectors {
		apiDetectors[i] = DetectorAPIModel{
			Function:            detector.Function.ValueString(),
			FieldName:           detector.FieldName.ValueString(),
			ByFieldName:         detector.ByFieldName.ValueString(),
			OverFieldName:       detector.OverFieldName.ValueString(),
			PartitionFieldName:  detector.PartitionFieldName.ValueString(),
			DetectorDescription: detector.DetectorDescription.ValueString(),
			ExcludeFrequent:     detector.ExcludeFrequent.ValueString(),
		}
		if utils.IsKnown(detector.UseNull) {
			apiDetectors[i].UseNull = utils.Pointer(detector.UseNull.ValueBool())
		}
	}

	// Convert influencers
	var influencers []string
	if utils.IsKnown(analysisConfig.Influencers) {
		d = analysisConfig.Influencers.ElementsAs(ctx, &influencers, false)
		diags.Append(d...)
	}

	apiModel.AnalysisConfig = AnalysisConfigAPIModel{
		BucketSpan:              analysisConfig.BucketSpan.ValueString(),
		CategorizationFieldName: analysisConfig.CategorizationFieldName.ValueString(),
		Detectors:               apiDetectors,
		Influencers:             influencers,
		Latency:                 analysisConfig.Latency.ValueString(),
		ModelPruneWindow:        analysisConfig.ModelPruneWindow.ValueString(),
		SummaryCountFieldName:   analysisConfig.SummaryCountFieldName.ValueString(),
	}

	if utils.IsKnown(analysisConfig.MultivariateByFields) {
		apiModel.AnalysisConfig.MultivariateByFields = utils.Pointer(analysisConfig.MultivariateByFields.ValueBool())
	}

	// Convert categorization filters
	if utils.IsKnown(analysisConfig.CategorizationFilters) {
		var categorizationFilters []string
		d = analysisConfig.CategorizationFilters.ElementsAs(ctx, &categorizationFilters, false)
		diags.Append(d...)
		apiModel.AnalysisConfig.CategorizationFilters = categorizationFilters
	}

	// Convert per_partition_categorization
	if utils.IsKnown(analysisConfig.PerPartitionCategorization) {
		var perPartitionCategorization PerPartitionCategorizationTFModel
		d = analysisConfig.PerPartitionCategorization.As(ctx, &perPartitionCategorization, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		apiModel.AnalysisConfig.PerPartitionCategorization = &PerPartitionCategorizationAPIModel{
			Enabled: perPartitionCategorization.Enabled.ValueBool(),
		}
		if utils.IsKnown(perPartitionCategorization.StopOnWarn) {
			apiModel.AnalysisConfig.PerPartitionCategorization.StopOnWarn = utils.Pointer(perPartitionCategorization.StopOnWarn.ValueBool())
		}
	}

	// Convert analysis_limits
	if utils.IsKnown(plan.AnalysisLimits) {
		var analysisLimits AnalysisLimitsTFModel
		d = plan.AnalysisLimits.As(ctx, &analysisLimits, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		apiModel.AnalysisLimits = &AnalysisLimitsAPIModel{
			ModelMemoryLimit: analysisLimits.ModelMemoryLimit.ValueString(),
		}
		if utils.IsKnown(analysisLimits.CategorizationExamplesLimit) {
			apiModel.AnalysisLimits.CategorizationExamplesLimit = utils.Pointer(analysisLimits.CategorizationExamplesLimit.ValueInt64())
		}
	}

	// Convert data_description
	var dataDescription DataDescriptionTFModel
	d = plan.DataDescription.As(ctx, &dataDescription, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	apiModel.DataDescription = DataDescriptionAPIModel{
		TimeField:  dataDescription.TimeField.ValueString(),
		TimeFormat: dataDescription.TimeFormat.ValueString(),
	}

	// Convert optional fields
	if utils.IsKnown(plan.AllowLazyOpen) {
		apiModel.AllowLazyOpen = utils.Pointer(plan.AllowLazyOpen.ValueBool())
	}

	if utils.IsKnown(plan.BackgroundPersistInterval) {
		apiModel.BackgroundPersistInterval = plan.BackgroundPersistInterval.ValueString()
	}

	if utils.IsKnown(plan.CustomSettings) {
		var customSettings map[string]interface{}
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			diags.AddError("Failed to parse custom_settings", err.Error())
			return nil, diags
		}
		apiModel.CustomSettings = customSettings
	}

	if utils.IsKnown(plan.DailyModelSnapshotRetentionAfterDays) {
		apiModel.DailyModelSnapshotRetentionAfterDays = utils.Pointer(plan.DailyModelSnapshotRetentionAfterDays.ValueInt64())
	}

	if utils.IsKnown(plan.ModelSnapshotRetentionDays) {
		apiModel.ModelSnapshotRetentionDays = utils.Pointer(plan.ModelSnapshotRetentionDays.ValueInt64())
	}

	if utils.IsKnown(plan.RenormalizationWindowDays) {
		apiModel.RenormalizationWindowDays = utils.Pointer(plan.RenormalizationWindowDays.ValueInt64())
	}

	if utils.IsKnown(plan.ResultsIndexName) {
		apiModel.ResultsIndexName = plan.ResultsIndexName.ValueString()
	}

	if utils.IsKnown(plan.ResultsRetentionDays) {
		apiModel.ResultsRetentionDays = utils.Pointer(plan.ResultsRetentionDays.ValueInt64())
	}

	// Convert model_plot_config
	if utils.IsKnown(plan.ModelPlotConfig) {
		var modelPlotConfig ModelPlotConfigTFModel
		d = plan.ModelPlotConfig.As(ctx, &modelPlotConfig, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		apiModel.ModelPlotConfig = &ModelPlotConfigAPIModel{
			Enabled: modelPlotConfig.Enabled.ValueBool(),
			Terms:   modelPlotConfig.Terms.ValueString(),
		}
		if utils.IsKnown(modelPlotConfig.AnnotationsEnabled) {
			apiModel.ModelPlotConfig.AnnotationsEnabled = utils.Pointer(modelPlotConfig.AnnotationsEnabled.ValueBool())
		}
	}

	return apiModel, diags
}

// FromAPIModel populates the model from an API response.
func (tfModel *AnomalyDetectionJobTFModel) fromAPIModel(ctx context.Context, apiModel *AnomalyDetectionJobAPIModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	tfModel.JobID = types.StringValue(apiModel.JobID)
	tfModel.Description = types.StringValue(apiModel.Description)
	tfModel.JobType = types.StringValue(apiModel.JobType)
	tfModel.JobVersion = types.StringValue(apiModel.JobVersion)

	// Convert create_time
	if apiModel.CreateTime != nil {
		tfModel.CreateTime = types.StringValue(fmt.Sprintf("%v", apiModel.CreateTime))
	} else {
		tfModel.CreateTime = types.StringNull()
	}

	// Convert model_snapshot_id
	tfModel.ModelSnapshotID = types.StringValue(apiModel.ModelSnapshotID)

	// Convert groups
	if len(apiModel.Groups) > 0 {
		groupsSet, d := types.SetValueFrom(ctx, types.StringType, apiModel.Groups)
		diags.Append(d...)
		tfModel.Groups = groupsSet
	} else {
		tfModel.Groups = types.SetNull(types.StringType)
	}

	// Convert optional fields
	tfModel.AllowLazyOpen = types.BoolPointerValue(apiModel.AllowLazyOpen)

	if apiModel.BackgroundPersistInterval != "" {
		tfModel.BackgroundPersistInterval = types.StringValue(apiModel.BackgroundPersistInterval)
	}

	if apiModel.CustomSettings != nil {
		customSettingsJSON, err := json.Marshal(apiModel.CustomSettings)
		if err != nil {
			diags.AddError("Failed to marshal custom_settings", err.Error())
			return diags
		}
		tfModel.CustomSettings = jsontypes.NewNormalizedValue(string(customSettingsJSON))
	} else {
		tfModel.CustomSettings = jsontypes.NewNormalizedNull()
	}

	tfModel.DailyModelSnapshotRetentionAfterDays = types.Int64PointerValue(apiModel.DailyModelSnapshotRetentionAfterDays)

	tfModel.ModelSnapshotRetentionDays = types.Int64PointerValue(apiModel.ModelSnapshotRetentionDays)

	if apiModel.RenormalizationWindowDays != nil {
		tfModel.RenormalizationWindowDays = types.Int64Value(*apiModel.RenormalizationWindowDays)
	}

	if apiModel.ResultsIndexName != "" {
		tfModel.ResultsIndexName = types.StringValue(apiModel.ResultsIndexName)
	}

	tfModel.ResultsRetentionDays = types.Int64PointerValue(apiModel.ResultsRetentionDays)

	// Convert analysis_config
	tfModel.AnalysisConfig = tfModel.convertAnalysisConfigFromAPI(ctx, &apiModel.AnalysisConfig, &diags)

	// Convert analysis_limits
	tfModel.AnalysisLimits = tfModel.convertAnalysisLimitsFromAPI(ctx, apiModel.AnalysisLimits, &diags)

	// Convert data_description
	tfModel.DataDescription = tfModel.convertDataDescriptionFromAPI(ctx, &apiModel.DataDescription, &diags)

	// Convert model_plot_config
	tfModel.ModelPlotConfig = tfModel.convertModelPlotConfigFromAPI(ctx, apiModel.ModelPlotConfig, &diags)

	// Convert analysis_limits
	tfModel.AnalysisLimits = tfModel.convertAnalysisLimitsFromAPI(ctx, apiModel.AnalysisLimits, &diags)

	// Convert model_plot_config
	tfModel.ModelPlotConfig = tfModel.convertModelPlotConfigFromAPI(ctx, apiModel.ModelPlotConfig, &diags)

	return diags
}

// Helper functions for schema attribute types
// Conversion helper methods
func (tfModel *AnomalyDetectionJobTFModel) convertAnalysisConfigFromAPI(ctx context.Context, apiConfig *AnalysisConfigAPIModel, diags *fwdiags.Diagnostics) types.Object {
	if apiConfig == nil || apiConfig.BucketSpan == "" {
		return types.ObjectNull(getAnalysisConfigAttrTypes())
	}

	analysisConfigTF := AnalysisConfigTFModel{
		BucketSpan: types.StringValue(apiConfig.BucketSpan),
	}

	// Convert optional string fields
	if apiConfig.CategorizationFieldName != "" {
		analysisConfigTF.CategorizationFieldName = types.StringValue(apiConfig.CategorizationFieldName)
	} else {
		analysisConfigTF.CategorizationFieldName = types.StringNull()
	}

	if apiConfig.Latency != "" {
		analysisConfigTF.Latency = types.StringValue(apiConfig.Latency)
	} else {
		analysisConfigTF.Latency = types.StringNull()
	}

	if apiConfig.ModelPruneWindow != "" {
		analysisConfigTF.ModelPruneWindow = types.StringValue(apiConfig.ModelPruneWindow)
	} else {
		analysisConfigTF.ModelPruneWindow = types.StringNull()
	}

	if apiConfig.SummaryCountFieldName != "" {
		analysisConfigTF.SummaryCountFieldName = types.StringValue(apiConfig.SummaryCountFieldName)
	} else {
		analysisConfigTF.SummaryCountFieldName = types.StringNull()
	}

	// Convert boolean fields
	analysisConfigTF.MultivariateByFields = types.BoolPointerValue(apiConfig.MultivariateByFields)

	// Convert categorization filters
	if len(apiConfig.CategorizationFilters) > 0 {
		categorizationFiltersListValue, d := types.ListValueFrom(ctx, types.StringType, apiConfig.CategorizationFilters)
		diags.Append(d...)
		analysisConfigTF.CategorizationFilters = categorizationFiltersListValue
	} else {
		analysisConfigTF.CategorizationFilters = types.ListNull(types.StringType)
	}

	// Convert influencers
	if len(apiConfig.Influencers) > 0 {
		influencersListValue, d := types.ListValueFrom(ctx, types.StringType, apiConfig.Influencers)
		diags.Append(d...)
		analysisConfigTF.Influencers = influencersListValue
	} else {
		analysisConfigTF.Influencers = types.ListNull(types.StringType)
	}

	// Convert detectors
	if len(apiConfig.Detectors) > 0 {
		detectorsTF := make([]DetectorTFModel, len(apiConfig.Detectors))
		for i, detector := range apiConfig.Detectors {
			detectorsTF[i] = DetectorTFModel{
				Function: types.StringValue(detector.Function),
			}

			// Convert optional string fields
			if detector.FieldName != "" {
				detectorsTF[i].FieldName = types.StringValue(detector.FieldName)
			} else {
				detectorsTF[i].FieldName = types.StringNull()
			}

			if detector.ByFieldName != "" {
				detectorsTF[i].ByFieldName = types.StringValue(detector.ByFieldName)
			} else {
				detectorsTF[i].ByFieldName = types.StringNull()
			}

			if detector.OverFieldName != "" {
				detectorsTF[i].OverFieldName = types.StringValue(detector.OverFieldName)
			} else {
				detectorsTF[i].OverFieldName = types.StringNull()
			}

			if detector.PartitionFieldName != "" {
				detectorsTF[i].PartitionFieldName = types.StringValue(detector.PartitionFieldName)
			} else {
				detectorsTF[i].PartitionFieldName = types.StringNull()
			}

			if detector.DetectorDescription != "" {
				detectorsTF[i].DetectorDescription = types.StringValue(detector.DetectorDescription)
			} else {
				detectorsTF[i].DetectorDescription = types.StringNull()
			}

			if detector.ExcludeFrequent != "" {
				detectorsTF[i].ExcludeFrequent = types.StringValue(detector.ExcludeFrequent)
			} else {
				detectorsTF[i].ExcludeFrequent = types.StringNull()
			}

			// Convert boolean field
			detectorsTF[i].UseNull = types.BoolPointerValue(detector.UseNull)

			// Convert custom rules
			if len(detector.CustomRules) > 0 {
				customRulesTF := make([]CustomRuleTFModel, len(detector.CustomRules))
				for j, rule := range detector.CustomRules {
					// Convert actions
					if len(rule.Actions) > 0 {
						// Convert interface{} actions to strings
						actions := make([]string, len(rule.Actions))
						for k, action := range rule.Actions {
							if actionStr, ok := action.(string); ok {
								actions[k] = actionStr
							}
						}
						actionsListValue, d := types.ListValueFrom(ctx, types.StringType, actions)
						diags.Append(d...)
						customRulesTF[j].Actions = actionsListValue
					} else {
						customRulesTF[j].Actions = types.ListNull(types.StringType)
					}

					// Convert conditions
					if len(rule.Conditions) > 0 {
						conditionsTF := make([]RuleConditionTFModel, len(rule.Conditions))
						for k, condition := range rule.Conditions {
							conditionsTF[k] = RuleConditionTFModel{
								AppliesTo: types.StringValue(condition.AppliesTo),
								Operator:  types.StringValue(condition.Operator),
								Value:     types.Float64Value(condition.Value),
							}
						}
						conditionsListValue, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getRuleConditionAttrTypes()}, conditionsTF)
						diags.Append(d...)
						customRulesTF[j].Conditions = conditionsListValue
					} else {
						customRulesTF[j].Conditions = types.ListNull(types.ObjectType{AttrTypes: getRuleConditionAttrTypes()})
					}
				}
				customRulesListValue, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getCustomRuleAttrTypes()}, customRulesTF)
				diags.Append(d...)
				detectorsTF[i].CustomRules = customRulesListValue
			} else {
				detectorsTF[i].CustomRules = types.ListNull(types.ObjectType{AttrTypes: getCustomRuleAttrTypes()})
			}
		}
		detectorsListValue, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDetectorAttrTypes()}, detectorsTF)
		diags.Append(d...)
		analysisConfigTF.Detectors = detectorsListValue
	} else {
		analysisConfigTF.Detectors = types.ListNull(types.ObjectType{AttrTypes: getDetectorAttrTypes()})
	}

	// Convert per_partition_categorization
	if apiConfig.PerPartitionCategorization != nil {
		perPartitionCategorizationTF := PerPartitionCategorizationTFModel{
			Enabled: types.BoolValue(apiConfig.PerPartitionCategorization.Enabled),
		}
		perPartitionCategorizationTF.StopOnWarn = types.BoolPointerValue(apiConfig.PerPartitionCategorization.StopOnWarn)
		perPartitionCategorizationObjectValue, d := types.ObjectValueFrom(ctx, getPerPartitionCategorizationAttrTypes(), perPartitionCategorizationTF)
		diags.Append(d...)
		analysisConfigTF.PerPartitionCategorization = perPartitionCategorizationObjectValue
	} else {
		analysisConfigTF.PerPartitionCategorization = types.ObjectNull(getPerPartitionCategorizationAttrTypes())
	}

	analysisConfigObjectValue, d := types.ObjectValueFrom(ctx, getAnalysisConfigAttrTypes(), analysisConfigTF)
	diags.Append(d...)
	return analysisConfigObjectValue
}

func (tfModel *AnomalyDetectionJobTFModel) convertDataDescriptionFromAPI(ctx context.Context, apiDataDescription *DataDescriptionAPIModel, diags *fwdiags.Diagnostics) types.Object {
	if apiDataDescription == nil {
		return types.ObjectNull(getDataDescriptionAttrTypes())
	}

	dataDescriptionTF := DataDescriptionTFModel{}

	if apiDataDescription.TimeField != "" {
		dataDescriptionTF.TimeField = types.StringValue(apiDataDescription.TimeField)
	} else {
		dataDescriptionTF.TimeField = types.StringNull()
	}

	if apiDataDescription.TimeFormat != "" {
		dataDescriptionTF.TimeFormat = types.StringValue(apiDataDescription.TimeFormat)
	} else {
		dataDescriptionTF.TimeFormat = types.StringNull()
	}

	dataDescriptionObjectValue, d := types.ObjectValueFrom(ctx, getDataDescriptionAttrTypes(), dataDescriptionTF)
	diags.Append(d...)
	return dataDescriptionObjectValue
}

func (tfModel *AnomalyDetectionJobTFModel) convertAnalysisLimitsFromAPI(ctx context.Context, apiLimits *AnalysisLimitsAPIModel, diags *fwdiags.Diagnostics) types.Object {
	if apiLimits == nil {
		return types.ObjectNull(getAnalysisLimitsAttrTypes())
	}

	analysisLimitsTF := AnalysisLimitsTFModel{
		CategorizationExamplesLimit: types.Int64PointerValue(apiLimits.CategorizationExamplesLimit),
	}

	if apiLimits.ModelMemoryLimit != "" {
		analysisLimitsTF.ModelMemoryLimit = types.StringValue(apiLimits.ModelMemoryLimit)
	} else {
		analysisLimitsTF.ModelMemoryLimit = types.StringNull()
	}

	analysisLimitsObjectValue, d := types.ObjectValueFrom(ctx, getAnalysisLimitsAttrTypes(), analysisLimitsTF)
	diags.Append(d...)
	return analysisLimitsObjectValue
}

func (tfModel *AnomalyDetectionJobTFModel) convertModelPlotConfigFromAPI(ctx context.Context, apiModelPlotConfig *ModelPlotConfigAPIModel, diags *fwdiags.Diagnostics) types.Object {
	if apiModelPlotConfig == nil {
		return types.ObjectNull(getModelPlotConfigAttrTypes())
	}

	modelPlotConfigTF := ModelPlotConfigTFModel{
		Enabled: types.BoolValue(apiModelPlotConfig.Enabled),
	}

	if apiModelPlotConfig.Terms != "" {
		modelPlotConfigTF.Terms = types.StringValue(apiModelPlotConfig.Terms)
	} else {
		modelPlotConfigTF.Terms = types.StringNull()
	}

	modelPlotConfigTF.AnnotationsEnabled = types.BoolPointerValue(apiModelPlotConfig.AnnotationsEnabled)

	modelPlotConfigObjectValue, d := types.ObjectValueFrom(ctx, getModelPlotConfigAttrTypes(), modelPlotConfigTF)
	diags.Append(d...)
	return modelPlotConfigObjectValue
}
