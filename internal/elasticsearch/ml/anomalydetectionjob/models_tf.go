// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package anomalydetectionjob

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TFModel represents the Terraform resource model for ML anomaly detection jobs
type TFModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	JobID                   types.String `tfsdk:"job_id"`
	Description             types.String `tfsdk:"description"`
	Groups                  types.Set    `tfsdk:"groups"`
	// AnalysisConfig is required in configuration, but can be null in state during import.
	AnalysisConfig                       *AnalysisConfigTFModel `tfsdk:"analysis_config"`
	AnalysisLimits                       types.Object           `tfsdk:"analysis_limits"`
	DataDescription                      types.Object           `tfsdk:"data_description"`
	ModelPlotConfig                      types.Object           `tfsdk:"model_plot_config"`
	AllowLazyOpen                        types.Bool             `tfsdk:"allow_lazy_open"`
	BackgroundPersistInterval            types.String           `tfsdk:"background_persist_interval"`
	CustomSettings                       jsontypes.Normalized   `tfsdk:"custom_settings"`
	DailyModelSnapshotRetentionAfterDays types.Int64            `tfsdk:"daily_model_snapshot_retention_after_days"`
	ModelSnapshotRetentionDays           types.Int64            `tfsdk:"model_snapshot_retention_days"`
	RenormalizationWindowDays            types.Int64            `tfsdk:"renormalization_window_days"`
	ResultsIndexName                     types.String           `tfsdk:"results_index_name"`
	ResultsRetentionDays                 types.Int64            `tfsdk:"results_retention_days"`

	// Read-only computed fields
	CreateTime      types.String `tfsdk:"create_time"`
	JobType         types.String `tfsdk:"job_type"`
	JobVersion      types.String `tfsdk:"job_version"`
	ModelSnapshotID types.String `tfsdk:"model_snapshot_id"`
}

// AnalysisConfigTFModel represents the analysis configuration
type AnalysisConfigTFModel struct {
	BucketSpan                 types.String                       `tfsdk:"bucket_span"`
	CategorizationFieldName    types.String                       `tfsdk:"categorization_field_name"`
	CategorizationFilters      types.List                         `tfsdk:"categorization_filters"`
	Detectors                  []DetectorTFModel                  `tfsdk:"detectors"`
	Influencers                types.List                         `tfsdk:"influencers"`
	Latency                    types.String                       `tfsdk:"latency"`
	ModelPruneWindow           types.String                       `tfsdk:"model_prune_window"`
	MultivariateByFields       types.Bool                         `tfsdk:"multivariate_by_fields"`
	PerPartitionCategorization *PerPartitionCategorizationTFModel `tfsdk:"per_partition_categorization"`
	SummaryCountFieldName      types.String                       `tfsdk:"summary_count_field_name"`
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
	CategorizationExamplesLimit types.Int64            `tfsdk:"categorization_examples_limit"`
	ModelMemoryLimit            customtypes.MemorySize `tfsdk:"model_memory_limit"`
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

// ToAPIModel converts TF model to APIModel
func (plan *TFModel) toAPIModel(ctx context.Context) (*APIModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiModel := &APIModel{
		JobID:       plan.JobID.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Convert groups
	if typeutils.IsKnown(plan.Groups) {
		var groups []string
		d := plan.Groups.ElementsAs(ctx, &groups, false)
		diags.Append(d...)
		apiModel.Groups = groups
	}

	if plan.AnalysisConfig == nil {
		diags.AddError("Missing analysis_config", "analysis_config is required")
		return nil, diags
	}
	analysisConfig := plan.AnalysisConfig

	// Convert detectors
	apiDetectors := make([]DetectorAPIModel, len(analysisConfig.Detectors))
	for i, detector := range analysisConfig.Detectors {
		apiDetectors[i] = DetectorAPIModel{
			Function:            detector.Function.ValueString(),
			FieldName:           detector.FieldName.ValueString(),
			ByFieldName:         detector.ByFieldName.ValueString(),
			OverFieldName:       detector.OverFieldName.ValueString(),
			PartitionFieldName:  detector.PartitionFieldName.ValueString(),
			DetectorDescription: detector.DetectorDescription.ValueString(),
			ExcludeFrequent:     detector.ExcludeFrequent.ValueString(),
		}
		if typeutils.IsKnown(detector.UseNull) {
			apiDetectors[i].UseNull = schemautil.Pointer(detector.UseNull.ValueBool())
		}
	}

	// Convert influencers
	var influencers []string
	if typeutils.IsKnown(analysisConfig.Influencers) {
		d := analysisConfig.Influencers.ElementsAs(ctx, &influencers, false)
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

	if typeutils.IsKnown(analysisConfig.MultivariateByFields) {
		apiModel.AnalysisConfig.MultivariateByFields = schemautil.Pointer(analysisConfig.MultivariateByFields.ValueBool())
	}

	// Convert categorization filters
	if typeutils.IsKnown(analysisConfig.CategorizationFilters) {
		var categorizationFilters []string
		d := analysisConfig.CategorizationFilters.ElementsAs(ctx, &categorizationFilters, false)
		diags.Append(d...)
		apiModel.AnalysisConfig.CategorizationFilters = categorizationFilters
	}

	// Convert per_partition_categorization
	if analysisConfig.PerPartitionCategorization != nil {
		apiModel.AnalysisConfig.PerPartitionCategorization = &PerPartitionCategorizationAPIModel{
			Enabled: analysisConfig.PerPartitionCategorization.Enabled.ValueBool(),
		}
		if typeutils.IsKnown(analysisConfig.PerPartitionCategorization.StopOnWarn) {
			apiModel.AnalysisConfig.PerPartitionCategorization.StopOnWarn = schemautil.Pointer(analysisConfig.PerPartitionCategorization.StopOnWarn.ValueBool())
		}
	}

	// Convert analysis_limits
	if typeutils.IsKnown(plan.AnalysisLimits) {
		var analysisLimits AnalysisLimitsTFModel
		d := plan.AnalysisLimits.As(ctx, &analysisLimits, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		apiModel.AnalysisLimits = &AnalysisLimitsAPIModel{
			ModelMemoryLimit: analysisLimits.ModelMemoryLimit.ValueString(),
		}
		if typeutils.IsKnown(analysisLimits.CategorizationExamplesLimit) {
			apiModel.AnalysisLimits.CategorizationExamplesLimit = schemautil.Pointer(analysisLimits.CategorizationExamplesLimit.ValueInt64())
		}
	}

	// Convert data_description
	var dataDescription DataDescriptionTFModel
	d := plan.DataDescription.As(ctx, &dataDescription, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	apiModel.DataDescription = DataDescriptionAPIModel{
		TimeField:  dataDescription.TimeField.ValueString(),
		TimeFormat: dataDescription.TimeFormat.ValueString(),
	}

	// Convert optional fields
	if typeutils.IsKnown(plan.AllowLazyOpen) {
		apiModel.AllowLazyOpen = schemautil.Pointer(plan.AllowLazyOpen.ValueBool())
	}

	if typeutils.IsKnown(plan.BackgroundPersistInterval) {
		apiModel.BackgroundPersistInterval = plan.BackgroundPersistInterval.ValueString()
	}

	if typeutils.IsKnown(plan.CustomSettings) {
		var customSettings map[string]any
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			diags.AddError("Failed to parse custom_settings", err.Error())
			return nil, diags
		}
		apiModel.CustomSettings = customSettings
	}

	if typeutils.IsKnown(plan.DailyModelSnapshotRetentionAfterDays) {
		apiModel.DailyModelSnapshotRetentionAfterDays = schemautil.Pointer(plan.DailyModelSnapshotRetentionAfterDays.ValueInt64())
	}

	if typeutils.IsKnown(plan.ModelSnapshotRetentionDays) {
		apiModel.ModelSnapshotRetentionDays = schemautil.Pointer(plan.ModelSnapshotRetentionDays.ValueInt64())
	}

	if typeutils.IsKnown(plan.RenormalizationWindowDays) {
		apiModel.RenormalizationWindowDays = schemautil.Pointer(plan.RenormalizationWindowDays.ValueInt64())
	}

	if typeutils.IsKnown(plan.ResultsIndexName) {
		apiModel.ResultsIndexName = plan.ResultsIndexName.ValueString()
	}

	if typeutils.IsKnown(plan.ResultsRetentionDays) {
		apiModel.ResultsRetentionDays = schemautil.Pointer(plan.ResultsRetentionDays.ValueInt64())
	}

	// Convert model_plot_config
	if typeutils.IsKnown(plan.ModelPlotConfig) {
		var modelPlotConfig ModelPlotConfigTFModel
		d = plan.ModelPlotConfig.As(ctx, &modelPlotConfig, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		apiModel.ModelPlotConfig = &ModelPlotConfigAPIModel{
			Enabled: modelPlotConfig.Enabled.ValueBool(),
			Terms:   modelPlotConfig.Terms.ValueString(),
		}
		if typeutils.IsKnown(modelPlotConfig.AnnotationsEnabled) {
			apiModel.ModelPlotConfig.AnnotationsEnabled = schemautil.Pointer(modelPlotConfig.AnnotationsEnabled.ValueBool())
		}
	}

	return apiModel, diags
}

// FromAPIModel populates the model from an API response.
func (plan *TFModel) fromAPIModel(ctx context.Context, apiModel *APIModel) diag.Diagnostics {
	var diags diag.Diagnostics

	plan.JobID = types.StringValue(apiModel.JobID)
	plan.Description = typeutils.NonEmptyStringishValue(apiModel.Description)
	plan.JobType = types.StringValue(apiModel.JobType)
	plan.JobVersion = types.StringValue(apiModel.JobVersion)

	// Convert create_time
	if apiModel.CreateTime != nil {
		plan.CreateTime = types.StringValue(fmt.Sprintf("%v", apiModel.CreateTime))
	} else {
		plan.CreateTime = types.StringNull()
	}

	// Convert model_snapshot_id
	plan.ModelSnapshotID = types.StringValue(apiModel.ModelSnapshotID)

	// Convert groups
	var groupDiags diag.Diagnostics
	plan.Groups, groupDiags = typeutils.NonEmptySetOrDefault(ctx, plan.Groups, types.StringType, apiModel.Groups)
	diags.Append(groupDiags...)

	// Convert optional fields
	plan.AllowLazyOpen = types.BoolPointerValue(apiModel.AllowLazyOpen)
	plan.BackgroundPersistInterval = typeutils.NonEmptyStringishValue(apiModel.BackgroundPersistInterval)

	if apiModel.CustomSettings != nil {
		customSettingsJSON, err := json.Marshal(apiModel.CustomSettings)
		if err != nil {
			diags.AddError("Failed to marshal custom_settings", err.Error())
			return diags
		}
		plan.CustomSettings = jsontypes.NewNormalizedValue(string(customSettingsJSON))
	} else {
		plan.CustomSettings = jsontypes.NewNormalizedNull()
	}

	plan.DailyModelSnapshotRetentionAfterDays = types.Int64PointerValue(apiModel.DailyModelSnapshotRetentionAfterDays)

	plan.ModelSnapshotRetentionDays = types.Int64PointerValue(apiModel.ModelSnapshotRetentionDays)

	if apiModel.RenormalizationWindowDays != nil {
		plan.RenormalizationWindowDays = types.Int64Value(*apiModel.RenormalizationWindowDays)
	}

	resultsIndexName := strings.TrimPrefix(apiModel.ResultsIndexName, "custom-")
	plan.ResultsIndexName = typeutils.NonEmptyStringishValue(resultsIndexName)
	plan.ResultsRetentionDays = types.Int64PointerValue(apiModel.ResultsRetentionDays)

	// Convert analysis_config
	plan.AnalysisConfig = plan.convertAnalysisConfigFromAPI(ctx, &apiModel.AnalysisConfig, &diags)

	// Convert analysis_limits
	plan.AnalysisLimits = plan.convertAnalysisLimitsFromAPI(ctx, apiModel.AnalysisLimits, &diags)

	// Convert data_description
	plan.DataDescription = plan.convertDataDescriptionFromAPI(ctx, &apiModel.DataDescription, &diags)

	// Convert model_plot_config
	plan.ModelPlotConfig = plan.convertModelPlotConfigFromAPI(ctx, apiModel.ModelPlotConfig, &diags)

	// Convert analysis_limits
	plan.AnalysisLimits = plan.convertAnalysisLimitsFromAPI(ctx, apiModel.AnalysisLimits, &diags)

	// Convert model_plot_config
	plan.ModelPlotConfig = plan.convertModelPlotConfigFromAPI(ctx, apiModel.ModelPlotConfig, &diags)

	return diags
}

// Helper functions for schema attribute types
// Conversion helper methods
func (plan *TFModel) convertAnalysisConfigFromAPI(ctx context.Context, apiConfig *AnalysisConfigAPIModel, diags *diag.Diagnostics) *AnalysisConfigTFModel {
	if apiConfig == nil || apiConfig.BucketSpan == "" {
		return nil
	}

	var analysisConfigTF AnalysisConfigTFModel
	if plan.AnalysisConfig != nil {
		analysisConfigTF = *plan.AnalysisConfig
	}
	analysisConfigTF.BucketSpan = types.StringValue(apiConfig.BucketSpan)

	// Convert optional string fields
	analysisConfigTF.CategorizationFieldName = typeutils.NonEmptyStringishValue(apiConfig.CategorizationFieldName)
	analysisConfigTF.Latency = typeutils.NonEmptyStringishValue(apiConfig.Latency)
	analysisConfigTF.ModelPruneWindow = typeutils.NonEmptyStringishValue(apiConfig.ModelPruneWindow)
	analysisConfigTF.SummaryCountFieldName = typeutils.NonEmptyStringishValue(apiConfig.SummaryCountFieldName)

	// Convert boolean fields
	analysisConfigTF.MultivariateByFields = types.BoolPointerValue(apiConfig.MultivariateByFields)

	// Convert categorization filters
	var categorizationFiltersDiags diag.Diagnostics
	analysisConfigTF.CategorizationFilters, categorizationFiltersDiags = typeutils.NonEmptyListOrDefault(ctx, analysisConfigTF.CategorizationFilters, types.StringType, apiConfig.CategorizationFilters)
	diags.Append(categorizationFiltersDiags...)
	// Ensure the list is properly typed (handles untyped zero-value lists from import)
	analysisConfigTF.CategorizationFilters = typeutils.EnsureTypedList(ctx, analysisConfigTF.CategorizationFilters, types.StringType)

	// Convert influencers
	var influencersDiags diag.Diagnostics
	analysisConfigTF.Influencers, influencersDiags = typeutils.NonEmptyListOrDefault(ctx, analysisConfigTF.Influencers, types.StringType, apiConfig.Influencers)
	diags.Append(influencersDiags...)
	// Ensure the list is properly typed (handles untyped zero-value lists from import)
	analysisConfigTF.Influencers = typeutils.EnsureTypedList(ctx, analysisConfigTF.Influencers, types.StringType)

	// Convert detectors
	if len(apiConfig.Detectors) > 0 {
		detectorsTF := make([]DetectorTFModel, len(apiConfig.Detectors))
		for i, detector := range apiConfig.Detectors {
			var originalDetector DetectorTFModel
			if len(analysisConfigTF.Detectors) > i {
				originalDetector = analysisConfigTF.Detectors[i]
			}

			detectorsTF[i] = DetectorTFModel{
				Function: types.StringValue(detector.Function),
			}

			// Convert optional string fields
			detectorsTF[i].FieldName = typeutils.NonEmptyStringishValue(detector.FieldName)
			detectorsTF[i].ByFieldName = typeutils.NonEmptyStringishValue(detector.ByFieldName)
			detectorsTF[i].OverFieldName = typeutils.NonEmptyStringishValue(detector.OverFieldName)
			detectorsTF[i].PartitionFieldName = typeutils.NonEmptyStringishValue(detector.PartitionFieldName)
			detectorsTF[i].DetectorDescription = typeutils.NonEmptyStringishValue(detector.DetectorDescription)
			detectorsTF[i].ExcludeFrequent = typeutils.NonEmptyStringishValue(detector.ExcludeFrequent)

			// Convert boolean field
			if detector.UseNull != nil {
				detectorsTF[i].UseNull = types.BoolValue(*detector.UseNull)
			} else {
				detectorsTF[i].UseNull = types.BoolValue(false)
			}

			// Convert custom rules

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

			var customRulesDiags diag.Diagnostics
			detectorsTF[i].CustomRules, customRulesDiags = typeutils.NonEmptyListOrDefault(
				ctx,
				originalDetector.CustomRules,
				types.ObjectType{AttrTypes: getCustomRuleAttrTypes()},
				apiConfig.Detectors[i].CustomRules,
			)
			diags.Append(customRulesDiags...)
			// Ensure the list is properly typed (handles untyped zero-value lists from import)
			detectorsTF[i].CustomRules = typeutils.EnsureTypedList(ctx, detectorsTF[i].CustomRules, types.ObjectType{AttrTypes: getCustomRuleAttrTypes()})
		}
		analysisConfigTF.Detectors = detectorsTF
	}

	// Convert per_partition_categorization
	if apiConfig.PerPartitionCategorization != nil {
		perPartitionCategorizationTF := PerPartitionCategorizationTFModel{
			Enabled: types.BoolValue(apiConfig.PerPartitionCategorization.Enabled),
		}
		perPartitionCategorizationTF.StopOnWarn = types.BoolPointerValue(apiConfig.PerPartitionCategorization.StopOnWarn)
		analysisConfigTF.PerPartitionCategorization = &perPartitionCategorizationTF
	}

	return &analysisConfigTF
}

func (plan *TFModel) convertDataDescriptionFromAPI(ctx context.Context, apiDataDescription *DataDescriptionAPIModel, diags *diag.Diagnostics) types.Object {
	if apiDataDescription == nil {
		return types.ObjectNull(getDataDescriptionAttrTypes())
	}

	dataDescriptionTF := DataDescriptionTFModel{
		TimeField:  typeutils.NonEmptyStringishValue(apiDataDescription.TimeField),
		TimeFormat: typeutils.NonEmptyStringishValue(apiDataDescription.TimeFormat),
	}

	dataDescriptionObjectValue, d := types.ObjectValueFrom(ctx, getDataDescriptionAttrTypes(), dataDescriptionTF)
	diags.Append(d...)
	return dataDescriptionObjectValue
}

func (plan *TFModel) convertAnalysisLimitsFromAPI(ctx context.Context, apiLimits *AnalysisLimitsAPIModel, diags *diag.Diagnostics) types.Object {
	if apiLimits == nil {
		return types.ObjectNull(getAnalysisLimitsAttrTypes())
	}

	analysisLimitsTF := AnalysisLimitsTFModel{
		CategorizationExamplesLimit: types.Int64PointerValue(apiLimits.CategorizationExamplesLimit),
	}

	if apiLimits.ModelMemoryLimit != "" {
		analysisLimitsTF.ModelMemoryLimit = customtypes.NewMemorySizeValue(apiLimits.ModelMemoryLimit)
	} else {
		analysisLimitsTF.ModelMemoryLimit = customtypes.NewMemorySizeNull()
	}

	analysisLimitsObjectValue, d := types.ObjectValueFrom(ctx, getAnalysisLimitsAttrTypes(), analysisLimitsTF)
	diags.Append(d...)
	return analysisLimitsObjectValue
}

func (plan *TFModel) convertModelPlotConfigFromAPI(ctx context.Context, apiModelPlotConfig *ModelPlotConfigAPIModel, diags *diag.Diagnostics) types.Object {
	if apiModelPlotConfig == nil {
		return types.ObjectNull(getModelPlotConfigAttrTypes())
	}

	modelPlotConfigTF := ModelPlotConfigTFModel{
		Enabled: types.BoolValue(apiModelPlotConfig.Enabled),
		Terms:   typeutils.NonEmptyStringishValue(apiModelPlotConfig.Terms),
	}

	modelPlotConfigTF.AnnotationsEnabled = types.BoolPointerValue(apiModelPlotConfig.AnnotationsEnabled)

	modelPlotConfigObjectValue, d := types.ObjectValueFrom(ctx, getModelPlotConfigAttrTypes(), modelPlotConfigTF)
	diags.Append(d...)
	return modelPlotConfigObjectValue
}
