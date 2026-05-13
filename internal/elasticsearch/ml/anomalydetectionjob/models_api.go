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

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/putjob"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/appliesto"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/conditionoperator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/excludefrequent"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/filtertype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/ruleaction"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// APIModel represents the API model for ML anomaly detection jobs
type APIModel struct {
	JobID                                string                   `json:"job_id"`
	Description                          string                   `json:"description,omitempty"`
	Groups                               []string                 `json:"groups,omitempty"`
	AnalysisConfig                       AnalysisConfigAPIModel   `json:"analysis_config"`
	AnalysisLimits                       *AnalysisLimitsAPIModel  `json:"analysis_limits,omitempty"`
	DataDescription                      DataDescriptionAPIModel  `json:"data_description"`
	ModelPlotConfig                      *ModelPlotConfigAPIModel `json:"model_plot_config,omitempty"`
	AllowLazyOpen                        *bool                    `json:"allow_lazy_open,omitempty"`
	BackgroundPersistInterval            string                   `json:"background_persist_interval,omitempty"`
	CustomSettings                       map[string]any           `json:"custom_settings,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                   `json:"daily_model_snapshot_retention_after_days,omitempty"`
	ModelSnapshotRetentionDays           *int64                   `json:"model_snapshot_retention_days,omitempty"`
	RenormalizationWindowDays            *int64                   `json:"renormalization_window_days,omitempty"`
	ResultsIndexName                     string                   `json:"results_index_name,omitempty"`
	ResultsRetentionDays                 *int64                   `json:"results_retention_days,omitempty"`

	// Read-only fields
	CreateTime      any    `json:"create_time,omitempty"`
	JobType         string `json:"job_type,omitempty"`
	JobVersion      string `json:"job_version,omitempty"`
	ModelSnapshotID string `json:"model_snapshot_id,omitempty"`
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

// FilterScopeAPIModel is one entry under a detection rule's scope (field name -> ML filter).
type FilterScopeAPIModel struct {
	FilterID   string `json:"filter_id"`
	FilterType string `json:"filter_type,omitempty"`
}

// CustomRuleAPIModel represents a custom rule in API format
type CustomRuleAPIModel struct {
	Actions    []any                          `json:"actions,omitempty"`
	Conditions []RuleConditionAPIModel        `json:"conditions,omitempty"`
	Scope      map[string]FilterScopeAPIModel `json:"scope,omitempty"`
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

// UpdateAPIModel represents the API model for updating ML anomaly detection jobs
// This includes only the fields that can be updated after job creation
type UpdateAPIModel struct {
	Description                          *string                  `json:"description,omitempty"`
	Groups                               []string                 `json:"groups,omitempty"`
	AnalysisLimits                       *AnalysisLimitsAPIModel  `json:"analysis_limits,omitempty"`
	ModelPlotConfig                      *ModelPlotConfigAPIModel `json:"model_plot_config,omitempty"`
	AllowLazyOpen                        *bool                    `json:"allow_lazy_open,omitempty"`
	BackgroundPersistInterval            *string                  `json:"background_persist_interval,omitempty"`
	CustomSettings                       map[string]any           `json:"custom_settings,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                   `json:"daily_model_snapshot_retention_after_days,omitempty"`
	ModelSnapshotRetentionDays           *int64                   `json:"model_snapshot_retention_days,omitempty"`
	RenormalizationWindowDays            *int64                   `json:"renormalization_window_days,omitempty"`
	ResultsRetentionDays                 *int64                   `json:"results_retention_days,omitempty"`
}

// BuildFromPlan populates the UpdateAPIModel from the plan and state models
func (u *UpdateAPIModel) BuildFromPlan(ctx context.Context, plan, state *TFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	hasChanges := false

	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() {
		u.Description = new(plan.Description.ValueString())
		hasChanges = true
	}

	if !plan.Groups.Equal(state.Groups) {
		var groups []string
		d := plan.Groups.ElementsAs(ctx, &groups, false)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
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
			AnnotationsEnabled: new(modelPlotConfig.AnnotationsEnabled.ValueBool()),
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
			apiAnalysisLimits.CategorizationExamplesLimit = new(analysisLimits.CategorizationExamplesLimit.ValueInt64())
		}
		u.AnalysisLimits = apiAnalysisLimits
		hasChanges = true
	}

	if !plan.AllowLazyOpen.Equal(state.AllowLazyOpen) {
		u.AllowLazyOpen = new(plan.AllowLazyOpen.ValueBool())
		hasChanges = true
	}

	if !plan.BackgroundPersistInterval.Equal(state.BackgroundPersistInterval) && !plan.BackgroundPersistInterval.IsNull() {
		u.BackgroundPersistInterval = new(plan.BackgroundPersistInterval.ValueString())
		hasChanges = true
	}

	if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
		var customSettings map[string]any
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			diags.AddError("Failed to parse custom_settings", err.Error())
			return false, diags
		}
		u.CustomSettings = customSettings
		hasChanges = true
	}

	if !plan.DailyModelSnapshotRetentionAfterDays.Equal(state.DailyModelSnapshotRetentionAfterDays) && !plan.DailyModelSnapshotRetentionAfterDays.IsNull() {
		u.DailyModelSnapshotRetentionAfterDays = new(plan.DailyModelSnapshotRetentionAfterDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ModelSnapshotRetentionDays.Equal(state.ModelSnapshotRetentionDays) && !plan.ModelSnapshotRetentionDays.IsNull() {
		u.ModelSnapshotRetentionDays = new(plan.ModelSnapshotRetentionDays.ValueInt64())
		hasChanges = true
	}

	if !plan.RenormalizationWindowDays.Equal(state.RenormalizationWindowDays) && !plan.RenormalizationWindowDays.IsNull() {
		u.RenormalizationWindowDays = new(plan.RenormalizationWindowDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ResultsRetentionDays.Equal(state.ResultsRetentionDays) && !plan.ResultsRetentionDays.IsNull() {
		u.ResultsRetentionDays = new(plan.ResultsRetentionDays.ValueInt64())
		hasChanges = true
	}

	return hasChanges, diags
}

// toPutJobRequest converts an APIModel to a putjob.Request for the typed API client.
func (a *APIModel) toPutJobRequest() putjob.Request {
	req := putjob.Request{
		AnalysisConfig:  a.toTypedAnalysisConfig(),
		DataDescription: a.toTypedDataDescription(),
	}

	if a.Description != "" {
		req.Description = &a.Description
	}
	if len(a.Groups) > 0 {
		req.Groups = a.Groups
	}
	if a.AnalysisLimits != nil {
		req.AnalysisLimits = &types.AnalysisLimits{
			CategorizationExamplesLimit: a.AnalysisLimits.CategorizationExamplesLimit,
			ModelMemoryLimit:            a.AnalysisLimits.ModelMemoryLimit,
		}
	}
	if a.ModelPlotConfig != nil {
		req.ModelPlotConfig = &types.ModelPlotConfig{
			AnnotationsEnabled: a.ModelPlotConfig.AnnotationsEnabled,
			Enabled:            &a.ModelPlotConfig.Enabled,
			Terms:              typeutils.NonEmptyStringPtr(a.ModelPlotConfig.Terms),
		}
	}
	req.AllowLazyOpen = a.AllowLazyOpen
	if a.BackgroundPersistInterval != "" {
		req.BackgroundPersistInterval = types.Duration(a.BackgroundPersistInterval)
	}
	if a.CustomSettings != nil {
		raw, err := json.Marshal(a.CustomSettings)
		if err == nil {
			req.CustomSettings = json.RawMessage(raw)
		}
	}
	req.DailyModelSnapshotRetentionAfterDays = a.DailyModelSnapshotRetentionAfterDays
	req.ModelSnapshotRetentionDays = a.ModelSnapshotRetentionDays
	req.RenormalizationWindowDays = a.RenormalizationWindowDays
	if a.ResultsIndexName != "" {
		req.ResultsIndexName = &a.ResultsIndexName
	}
	req.ResultsRetentionDays = a.ResultsRetentionDays

	return req
}

// fromTypedJob converts a types.Job to an APIModel for use with fromAPIModel.
func fromTypedJob(job *types.Job) *APIModel {
	m := &APIModel{
		JobID:                                job.JobId,
		Groups:                               job.Groups,
		AllowLazyOpen:                        &job.AllowLazyOpen,
		CustomSettings:                       customSettingsFromRaw(job.CustomSettings),
		DailyModelSnapshotRetentionAfterDays: job.DailyModelSnapshotRetentionAfterDays,
		ModelSnapshotRetentionDays:           &job.ModelSnapshotRetentionDays,
		RenormalizationWindowDays:            job.RenormalizationWindowDays,
		ResultsRetentionDays:                 job.ResultsRetentionDays,
	}
	if job.Description != nil {
		m.Description = *job.Description
	}
	if job.JobType != nil {
		m.JobType = *job.JobType
	}
	if job.JobVersion != nil {
		m.JobVersion = *job.JobVersion
	}
	if job.ModelSnapshotId != nil {
		m.ModelSnapshotID = *job.ModelSnapshotId
	}
	m.CreateTime = job.CreateTime
	m.ResultsIndexName = job.ResultsIndexName

	// BackgroundPersistInterval
	if job.BackgroundPersistInterval != nil {
		m.BackgroundPersistInterval = durationToString(job.BackgroundPersistInterval)
	}

	// AnalysisConfig
	m.AnalysisConfig = typedAnalysisConfigToAPIModel(&job.AnalysisConfig)

	// AnalysisLimits
	if job.AnalysisLimits != nil {
		m.AnalysisLimits = &AnalysisLimitsAPIModel{
			CategorizationExamplesLimit: job.AnalysisLimits.CategorizationExamplesLimit,
		}
		if job.AnalysisLimits.ModelMemoryLimit != nil {
			m.AnalysisLimits.ModelMemoryLimit = bytesSizeToString(job.AnalysisLimits.ModelMemoryLimit)
		}
	}

	// DataDescription
	m.DataDescription = DataDescriptionAPIModel{}
	if job.DataDescription.TimeField != nil {
		m.DataDescription.TimeField = *job.DataDescription.TimeField
	}
	if job.DataDescription.TimeFormat != nil {
		m.DataDescription.TimeFormat = *job.DataDescription.TimeFormat
	}

	// ModelPlotConfig
	if job.ModelPlotConfig != nil {
		m.ModelPlotConfig = &ModelPlotConfigAPIModel{
			AnnotationsEnabled: job.ModelPlotConfig.AnnotationsEnabled,
		}
		if job.ModelPlotConfig.Enabled != nil {
			m.ModelPlotConfig.Enabled = *job.ModelPlotConfig.Enabled
		}
		if job.ModelPlotConfig.Terms != nil {
			m.ModelPlotConfig.Terms = *job.ModelPlotConfig.Terms
		}
	}

	return m
}

// toTypedAnalysisConfig converts an AnalysisConfigAPIModel to types.AnalysisConfig.
func (a *APIModel) toTypedAnalysisConfig() types.AnalysisConfig {
	cfg := types.AnalysisConfig{
		BucketSpan: types.Duration(a.AnalysisConfig.BucketSpan),
		Detectors:  make([]types.Detector, len(a.AnalysisConfig.Detectors)),
	}
	if a.AnalysisConfig.CategorizationFieldName != "" {
		cfg.CategorizationFieldName = &a.AnalysisConfig.CategorizationFieldName
	}
	if len(a.AnalysisConfig.CategorizationFilters) > 0 {
		cfg.CategorizationFilters = a.AnalysisConfig.CategorizationFilters
	}
	if len(a.AnalysisConfig.Influencers) > 0 {
		cfg.Influencers = a.AnalysisConfig.Influencers
	}
	if a.AnalysisConfig.Latency != "" {
		cfg.Latency = types.Duration(a.AnalysisConfig.Latency)
	}
	if a.AnalysisConfig.ModelPruneWindow != "" {
		cfg.ModelPruneWindow = types.Duration(a.AnalysisConfig.ModelPruneWindow)
	}
	cfg.MultivariateByFields = a.AnalysisConfig.MultivariateByFields
	if a.AnalysisConfig.SummaryCountFieldName != "" {
		cfg.SummaryCountFieldName = &a.AnalysisConfig.SummaryCountFieldName
	}
	if a.AnalysisConfig.PerPartitionCategorization != nil {
		cfg.PerPartitionCategorization = &types.PerPartitionCategorization{
			Enabled:    &a.AnalysisConfig.PerPartitionCategorization.Enabled,
			StopOnWarn: a.AnalysisConfig.PerPartitionCategorization.StopOnWarn,
		}
	}
	for i, d := range a.AnalysisConfig.Detectors {
		cfg.Detectors[i] = detectorAPIModelToTyped(&d)
	}
	return cfg
}

// toTypedDataDescription converts a DataDescriptionAPIModel to types.DataDescription.
func (a *APIModel) toTypedDataDescription() types.DataDescription {
	dd := types.DataDescription{}
	if a.DataDescription.TimeField != "" {
		dd.TimeField = &a.DataDescription.TimeField
	}
	if a.DataDescription.TimeFormat != "" {
		dd.TimeFormat = &a.DataDescription.TimeFormat
	}
	return dd
}

// detectorAPIModelToTyped converts a DetectorAPIModel to types.Detector.
func detectorAPIModelToTyped(d *DetectorAPIModel) types.Detector {
	det := types.Detector{}
	if d.Function != "" {
		det.Function = &d.Function
	}
	if d.FieldName != "" {
		det.FieldName = &d.FieldName
	}
	if d.ByFieldName != "" {
		det.ByFieldName = &d.ByFieldName
	}
	if d.OverFieldName != "" {
		det.OverFieldName = &d.OverFieldName
	}
	if d.PartitionFieldName != "" {
		det.PartitionFieldName = &d.PartitionFieldName
	}
	if d.DetectorDescription != "" {
		det.DetectorDescription = &d.DetectorDescription
	}
	if d.ExcludeFrequent != "" {
		ef := excludefrequent.ExcludeFrequent{Name: d.ExcludeFrequent}
		det.ExcludeFrequent = &ef
	}
	det.UseNull = d.UseNull
	if len(d.CustomRules) > 0 {
		det.CustomRules = make([]types.DetectionRule, len(d.CustomRules))
		for i, cr := range d.CustomRules {
			det.CustomRules[i] = detectionRuleAPIModelToTyped(&cr)
		}
	}
	return det
}

// detectionRuleAPIModelToTyped converts a CustomRuleAPIModel to types.DetectionRule.
func detectionRuleAPIModelToTyped(cr *CustomRuleAPIModel) types.DetectionRule {
	rule := types.DetectionRule{}
	if len(cr.Actions) > 0 {
		rule.Actions = make([]ruleaction.RuleAction, 0, len(cr.Actions))
		for _, a := range cr.Actions {
			if s, ok := a.(string); ok {
				rule.Actions = append(rule.Actions, ruleaction.RuleAction{Name: s})
			}
		}
	}
	if len(cr.Conditions) > 0 {
		rule.Conditions = make([]types.RuleCondition, len(cr.Conditions))
		for i, c := range cr.Conditions {
			rule.Conditions[i] = types.RuleCondition{
				AppliesTo: appliesto.AppliesTo{Name: c.AppliesTo},
				Operator:  conditionoperator.ConditionOperator{Name: c.Operator},
				Value:     types.Float64(c.Value),
			}
		}
	}
	if len(cr.Scope) > 0 {
		rule.Scope = make(map[string]types.FilterRef, len(cr.Scope))
		for k, v := range cr.Scope {
			fr := types.FilterRef{FilterId: v.FilterID}
			if v.FilterType != "" {
				ft := filtertype.FilterType{Name: v.FilterType}
				fr.FilterType = &ft
			}
			rule.Scope[k] = fr
		}
	}
	return rule
}

// typedAnalysisConfigToAPIModel converts a types.AnalysisConfig to AnalysisConfigAPIModel.
func typedAnalysisConfigToAPIModel(cfg *types.AnalysisConfig) AnalysisConfigAPIModel {
	a := AnalysisConfigAPIModel{
		BucketSpan: durationToString(cfg.BucketSpan),
		Detectors:  make([]DetectorAPIModel, len(cfg.Detectors)),
	}
	if cfg.CategorizationFieldName != nil {
		a.CategorizationFieldName = *cfg.CategorizationFieldName
	}
	a.CategorizationFilters = cfg.CategorizationFilters
	a.Influencers = cfg.Influencers
	a.Latency = durationToString(cfg.Latency)
	a.ModelPruneWindow = durationToString(cfg.ModelPruneWindow)
	a.MultivariateByFields = cfg.MultivariateByFields
	if cfg.SummaryCountFieldName != nil {
		a.SummaryCountFieldName = *cfg.SummaryCountFieldName
	}
	if cfg.PerPartitionCategorization != nil {
		ppc := &PerPartitionCategorizationAPIModel{}
		if cfg.PerPartitionCategorization.Enabled != nil {
			ppc.Enabled = *cfg.PerPartitionCategorization.Enabled
		}
		ppc.StopOnWarn = cfg.PerPartitionCategorization.StopOnWarn
		a.PerPartitionCategorization = ppc
	}
	for i, d := range cfg.Detectors {
		a.Detectors[i] = typedDetectorToAPIModel(&d)
	}
	return a
}

// typedDetectorToAPIModel converts a types.Detector to DetectorAPIModel.
func typedDetectorToAPIModel(d *types.Detector) DetectorAPIModel {
	det := DetectorAPIModel{}
	if d.Function != nil {
		det.Function = *d.Function
	}
	if d.FieldName != nil {
		det.FieldName = *d.FieldName
	}
	if d.ByFieldName != nil {
		det.ByFieldName = *d.ByFieldName
	}
	if d.OverFieldName != nil {
		det.OverFieldName = *d.OverFieldName
	}
	if d.PartitionFieldName != nil {
		det.PartitionFieldName = *d.PartitionFieldName
	}
	if d.DetectorDescription != nil {
		det.DetectorDescription = *d.DetectorDescription
	}
	if d.ExcludeFrequent != nil {
		det.ExcludeFrequent = d.ExcludeFrequent.Name
	}
	det.UseNull = d.UseNull
	if len(d.CustomRules) > 0 {
		det.CustomRules = make([]CustomRuleAPIModel, len(d.CustomRules))
		for i, cr := range d.CustomRules {
			det.CustomRules[i] = typedDetectionRuleToAPIModel(&cr)
		}
	}
	return det
}

// typedDetectionRuleToAPIModel converts a types.DetectionRule to CustomRuleAPIModel.
func typedDetectionRuleToAPIModel(cr *types.DetectionRule) CustomRuleAPIModel {
	rule := CustomRuleAPIModel{}
	if len(cr.Actions) > 0 {
		rule.Actions = make([]any, len(cr.Actions))
		for i, a := range cr.Actions {
			rule.Actions[i] = a.Name
		}
	}
	if len(cr.Conditions) > 0 {
		rule.Conditions = make([]RuleConditionAPIModel, len(cr.Conditions))
		for i, c := range cr.Conditions {
			rule.Conditions[i] = RuleConditionAPIModel{
				AppliesTo: c.AppliesTo.Name,
				Operator:  c.Operator.Name,
				Value:     float64(c.Value),
			}
		}
	}
	if len(cr.Scope) > 0 {
		rule.Scope = make(map[string]FilterScopeAPIModel, len(cr.Scope))
		for k, v := range cr.Scope {
			fs := FilterScopeAPIModel{FilterID: v.FilterId}
			if v.FilterType != nil {
				fs.FilterType = v.FilterType.Name
			}
			rule.Scope[k] = fs
		}
	}
	return rule
}

// customSettingsFromRaw converts json.RawMessage to map[string]any.
func customSettingsFromRaw(raw json.RawMessage) map[string]any {
	if raw == nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

// durationToString converts a types.Duration (any) to string.
func durationToString(d types.Duration) string {
	if d == nil {
		return ""
	}
	switch v := d.(type) {
	case string:
		return v
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		s := string(raw)
		// Remove surrounding quotes if present
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}

// bytesSizeToString converts a types.ByteSize (any) to string.
func bytesSizeToString(b types.ByteSize) string {
	if b == nil {
		return ""
	}
	switch v := b.(type) {
	case string:
		return v
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		s := string(raw)
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}
