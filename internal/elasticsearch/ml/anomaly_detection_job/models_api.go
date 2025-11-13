package anomaly_detection_job

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// AnomalyDetectionJobAPIModel represents the API model for ML anomaly detection jobs
type AnomalyDetectionJobAPIModel struct {
	JobID                                string                   `json:"job_id"`
	Description                          string                   `json:"description,omitempty"`
	Groups                               []string                 `json:"groups,omitempty"`
	AnalysisConfig                       AnalysisConfigAPIModel   `json:"analysis_config"`
	AnalysisLimits                       *AnalysisLimitsAPIModel  `json:"analysis_limits,omitempty"`
	DataDescription                      DataDescriptionAPIModel  `json:"data_description"`
	ModelPlotConfig                      *ModelPlotConfigAPIModel `json:"model_plot_config,omitempty"`
	AllowLazyOpen                        *bool                    `json:"allow_lazy_open,omitempty"`
	BackgroundPersistInterval            string                   `json:"background_persist_interval,omitempty"`
	CustomSettings                       map[string]interface{}   `json:"custom_settings,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                   `json:"daily_model_snapshot_retention_after_days,omitempty"`
	ModelSnapshotRetentionDays           *int64                   `json:"model_snapshot_retention_days,omitempty"`
	RenormalizationWindowDays            *int64                   `json:"renormalization_window_days,omitempty"`
	ResultsIndexName                     string                   `json:"results_index_name,omitempty"`
	ResultsRetentionDays                 *int64                   `json:"results_retention_days,omitempty"`

	// Read-only fields
	CreateTime      interface{} `json:"create_time,omitempty"`
	JobType         string      `json:"job_type,omitempty"`
	JobVersion      string      `json:"job_version,omitempty"`
	ModelSnapshotID string      `json:"model_snapshot_id,omitempty"`
}

// AnalysisConfigAPIModel represents the analysis configuration in API format
type AnalysisConfigAPIModel struct {
	BucketSpan                 string                              `json:"bucket_span"`
	CategorizationFieldName    string                              `json:"categorization_field_name,omitempty"`
	CategorizationFilters      []string                            `json:"categorization_filters,omitempty"`
	Detectors                  []DetectorAPIModel                  `json:"detectors"`
	Influencers                []string                            `json:"influencers,omitempty"`
	Latency                    string                              `json:"latency,omitempty"`
	ModelPruneWindow           string                              `json:"model_prune_window,omitempty"`
	MultivariateByFields       *bool                               `json:"multivariate_by_fields,omitempty"`
	PerPartitionCategorization *PerPartitionCategorizationAPIModel `json:"per_partition_categorization,omitempty"`
	SummaryCountFieldName      string                              `json:"summary_count_field_name,omitempty"`
}

// DetectorAPIModel represents a detector configuration in API format
type DetectorAPIModel struct {
	ByFieldName         string               `json:"by_field_name,omitempty"`
	DetectorDescription string               `json:"detector_description,omitempty"`
	ExcludeFrequent     string               `json:"exclude_frequent,omitempty"`
	FieldName           string               `json:"field_name,omitempty"`
	Function            string               `json:"function"`
	OverFieldName       string               `json:"over_field_name,omitempty"`
	PartitionFieldName  string               `json:"partition_field_name,omitempty"`
	UseNull             *bool                `json:"use_null,omitempty"`
	CustomRules         []CustomRuleAPIModel `json:"custom_rules,omitempty"`
}

// CustomRuleAPIModel represents a custom rule in API format
type CustomRuleAPIModel struct {
	Actions    []interface{}           `json:"actions,omitempty"`
	Conditions []RuleConditionAPIModel `json:"conditions,omitempty"`
}

// RuleConditionAPIModel represents a rule condition in API format
type RuleConditionAPIModel struct {
	AppliesTo string  `json:"applies_to"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
}

// AnalysisLimitsAPIModel represents analysis limits in API format
type AnalysisLimitsAPIModel struct {
	CategorizationExamplesLimit *int64 `json:"categorization_examples_limit,omitempty"`
	ModelMemoryLimit            string `json:"model_memory_limit,omitempty"`
}

// DataDescriptionAPIModel represents data description in API format
type DataDescriptionAPIModel struct {
	TimeField  string `json:"time_field,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`
}

// ChunkingConfigAPIModel represents chunking configuration in API format
type ChunkingConfigAPIModel struct {
	Mode     string `json:"mode"`
	TimeSpan string `json:"time_span,omitempty"`
}

// DelayedDataCheckConfigAPIModel represents delayed data check configuration in API format
type DelayedDataCheckConfigAPIModel struct {
	CheckWindow string `json:"check_window,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// IndicesOptionsAPIModel represents indices options in API format
type IndicesOptionsAPIModel struct {
	ExpandWildcards   []string `json:"expand_wildcards,omitempty"`
	IgnoreUnavailable *bool    `json:"ignore_unavailable,omitempty"`
	AllowNoIndices    *bool    `json:"allow_no_indices,omitempty"`
	IgnoreThrottled   *bool    `json:"ignore_throttled,omitempty"`
}

// ModelPlotConfigAPIModel represents model plot configuration in API format
type ModelPlotConfigAPIModel struct {
	AnnotationsEnabled *bool  `json:"annotations_enabled,omitempty"`
	Enabled            bool   `json:"enabled"`
	Terms              string `json:"terms,omitempty"`
}

// PerPartitionCategorizationAPIModel represents per-partition categorization in API format
type PerPartitionCategorizationAPIModel struct {
	Enabled    bool  `json:"enabled"`
	StopOnWarn *bool `json:"stop_on_warn,omitempty"`
}

// AnomalyDetectionJobUpdateAPIModel represents the API model for updating ML anomaly detection jobs
// This includes only the fields that can be updated after job creation
type AnomalyDetectionJobUpdateAPIModel struct {
	Description                          *string                  `json:"description,omitempty"`
	Groups                               []string                 `json:"groups,omitempty"`
	AnalysisLimits                       *AnalysisLimitsAPIModel  `json:"analysis_limits,omitempty"`
	ModelPlotConfig                      *ModelPlotConfigAPIModel `json:"model_plot_config,omitempty"`
	AllowLazyOpen                        *bool                    `json:"allow_lazy_open,omitempty"`
	BackgroundPersistInterval            *string                  `json:"background_persist_interval,omitempty"`
	CustomSettings                       map[string]interface{}   `json:"custom_settings,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                   `json:"daily_model_snapshot_retention_after_days,omitempty"`
	ModelSnapshotRetentionDays           *int64                   `json:"model_snapshot_retention_days,omitempty"`
	RenormalizationWindowDays            *int64                   `json:"renormalization_window_days,omitempty"`
	ResultsRetentionDays                 *int64                   `json:"results_retention_days,omitempty"`
}

// BuildFromPlan populates the AnomalyDetectionJobUpdateAPIModel from the plan and state models
func (u *AnomalyDetectionJobUpdateAPIModel) BuildFromPlan(ctx context.Context, plan, state *AnomalyDetectionJobTFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	hasChanges := false

	if !plan.Description.Equal(state.Description) {
		u.Description = utils.Pointer(plan.Description.ValueString())
		hasChanges = true
	}

	if !plan.Groups.Equal(state.Groups) {
		var groups []string
		d := plan.Groups.ElementsAs(ctx, &groups, false)
		diags.Append(d...)
		u.Groups = groups
		hasChanges = true
	}

	if !plan.ModelPlotConfig.Equal(state.ModelPlotConfig) {
		var modelPlotConfig ModelPlotConfigTFModel
		d := plan.ModelPlotConfig.As(ctx, &modelPlotConfig, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		apiModelPlotConfig := &ModelPlotConfigAPIModel{
			Enabled:            modelPlotConfig.Enabled.ValueBool(),
			AnnotationsEnabled: utils.Pointer(modelPlotConfig.AnnotationsEnabled.ValueBool()),
			Terms:              modelPlotConfig.Terms.ValueString(),
		}
		u.ModelPlotConfig = apiModelPlotConfig
		hasChanges = true
	}

	if !plan.AnalysisLimits.Equal(state.AnalysisLimits) {
		var analysisLimits AnalysisLimitsTFModel
		d := plan.AnalysisLimits.As(ctx, &analysisLimits, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		apiAnalysisLimits := &AnalysisLimitsAPIModel{
			ModelMemoryLimit: analysisLimits.ModelMemoryLimit.ValueString(),
		}
		if !analysisLimits.CategorizationExamplesLimit.IsNull() {
			apiAnalysisLimits.CategorizationExamplesLimit = utils.Pointer(analysisLimits.CategorizationExamplesLimit.ValueInt64())
		}
		u.AnalysisLimits = apiAnalysisLimits
		hasChanges = true
	}

	if !plan.AllowLazyOpen.Equal(state.AllowLazyOpen) {
		u.AllowLazyOpen = utils.Pointer(plan.AllowLazyOpen.ValueBool())
		hasChanges = true
	}

	if !plan.BackgroundPersistInterval.Equal(state.BackgroundPersistInterval) && !plan.BackgroundPersistInterval.IsNull() {
		u.BackgroundPersistInterval = utils.Pointer(plan.BackgroundPersistInterval.ValueString())
		hasChanges = true
	}

	if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
		var customSettings map[string]interface{}
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			diags.AddError("Failed to parse custom_settings", err.Error())
			return false, diags
		}
		u.CustomSettings = customSettings
		hasChanges = true
	}

	if !plan.DailyModelSnapshotRetentionAfterDays.Equal(state.DailyModelSnapshotRetentionAfterDays) && !plan.DailyModelSnapshotRetentionAfterDays.IsNull() {
		u.DailyModelSnapshotRetentionAfterDays = utils.Pointer(plan.DailyModelSnapshotRetentionAfterDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ModelSnapshotRetentionDays.Equal(state.ModelSnapshotRetentionDays) && !plan.ModelSnapshotRetentionDays.IsNull() {
		u.ModelSnapshotRetentionDays = utils.Pointer(plan.ModelSnapshotRetentionDays.ValueInt64())
		hasChanges = true
	}

	if !plan.RenormalizationWindowDays.Equal(state.RenormalizationWindowDays) && !plan.RenormalizationWindowDays.IsNull() {
		u.RenormalizationWindowDays = utils.Pointer(plan.RenormalizationWindowDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ResultsRetentionDays.Equal(state.ResultsRetentionDays) && !plan.ResultsRetentionDays.IsNull() {
		u.ResultsRetentionDays = utils.Pointer(plan.ResultsRetentionDays.ValueInt64())
		hasChanges = true
	}

	return hasChanges, diags
}
