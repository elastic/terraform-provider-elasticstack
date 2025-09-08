package anomaly_detector

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AnomalyDetectorJobTFModel represents the Terraform resource model for ML anomaly detection jobs
type AnomalyDetectorJobTFModel struct {
	ID                                   types.String `tfsdk:"id"`
	ElasticsearchConnection              types.List   `tfsdk:"elasticsearch_connection"`
	JobID                                types.String `tfsdk:"job_id"`
	Description                          types.String `tfsdk:"description"`
	Groups                               types.Set    `tfsdk:"groups"`
	AnalysisConfig                       types.Object `tfsdk:"analysis_config"`
	AnalysisLimits                       types.Object `tfsdk:"analysis_limits"`
	DataDescription                      types.Object `tfsdk:"data_description"`
	DatafeedConfig                       types.Object `tfsdk:"datafeed_config"`
	ModelPlotConfig                      types.Object `tfsdk:"model_plot_config"`
	AllowLazyOpen                        types.Bool   `tfsdk:"allow_lazy_open"`
	BackgroundPersistInterval            types.String `tfsdk:"background_persist_interval"`
	CustomSettings                       types.String `tfsdk:"custom_settings"`
	DailyModelSnapshotRetentionAfterDays types.Int64  `tfsdk:"daily_model_snapshot_retention_after_days"`
	ModelSnapshotRetentionDays           types.Int64  `tfsdk:"model_snapshot_retention_days"`
	RenormalizationWindowDays            types.Int64  `tfsdk:"renormalization_window_days"`
	ResultsIndexName                     types.String `tfsdk:"results_index_name"`
	ResultsRetentionDays                 types.Int64  `tfsdk:"results_retention_days"`

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
	FieldDelimiter types.String `tfsdk:"field_delimiter"`
	Format         types.String `tfsdk:"format"`
	QuoteCharacter types.String `tfsdk:"quote_character"`
	TimeField      types.String `tfsdk:"time_field"`
	TimeFormat     types.String `tfsdk:"time_format"`
}

// DatafeedConfigTFModel represents datafeed configuration
type DatafeedConfigTFModel struct {
	AggregationsConfig     types.String `tfsdk:"aggregations"`
	ChunkingConfig         types.Object `tfsdk:"chunking_config"`
	DatafeedID             types.String `tfsdk:"datafeed_id"`
	DelayedDataCheckConfig types.Object `tfsdk:"delayed_data_check_config"`
	Frequency              types.String `tfsdk:"frequency"`
	Indices                types.List   `tfsdk:"indices"`
	IndicesOptions         types.Object `tfsdk:"indices_options"`
	MaxEmptySearches       types.Int64  `tfsdk:"max_empty_searches"`
	Query                  types.String `tfsdk:"query"`
	QueryDelay             types.String `tfsdk:"query_delay"`
	RuntimeMappings        types.String `tfsdk:"runtime_mappings"`
	ScriptFields           types.String `tfsdk:"script_fields"`
	ScrollSize             types.Int64  `tfsdk:"scroll_size"`
}

// ChunkingConfigTFModel represents chunking configuration
type ChunkingConfigTFModel struct {
	Mode     types.String `tfsdk:"mode"`
	TimeSpan types.String `tfsdk:"time_span"`
}

// DelayedDataCheckConfigTFModel represents delayed data check configuration
type DelayedDataCheckConfigTFModel struct {
	CheckWindow types.String `tfsdk:"check_window"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// IndicesOptionsTFModel represents indices options configuration
type IndicesOptionsTFModel struct {
	ExpandWildcards   types.List `tfsdk:"expand_wildcards"`
	IgnoreUnavailable types.Bool `tfsdk:"ignore_unavailable"`
	AllowNoIndices    types.Bool `tfsdk:"allow_no_indices"`
	IgnoreThrottled   types.Bool `tfsdk:"ignore_throttled"`
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

// AnomalyDetectorJobAPIModel represents the API model for ML anomaly detection jobs
type AnomalyDetectorJobAPIModel struct {
	JobID                                string                   `json:"job_id"`
	Description                          string                   `json:"description,omitempty"`
	Groups                               []string                 `json:"groups,omitempty"`
	AnalysisConfig                       AnalysisConfigAPIModel   `json:"analysis_config"`
	AnalysisLimits                       *AnalysisLimitsAPIModel  `json:"analysis_limits,omitempty"`
	DataDescription                      DataDescriptionAPIModel  `json:"data_description"`
	DatafeedConfig                       *DatafeedConfigAPIModel  `json:"datafeed_config,omitempty"`
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
	FieldDelimiter string `json:"field_delimiter,omitempty"`
	Format         string `json:"format,omitempty"`
	QuoteCharacter string `json:"quote_character,omitempty"`
	TimeField      string `json:"time_field,omitempty"`
	TimeFormat     string `json:"time_format,omitempty"`
}

// DatafeedConfigAPIModel represents datafeed configuration in API format
type DatafeedConfigAPIModel struct {
	Aggregations           map[string]interface{}          `json:"aggregations,omitempty"`
	ChunkingConfig         *ChunkingConfigAPIModel         `json:"chunking_config,omitempty"`
	DatafeedID             string                          `json:"datafeed_id,omitempty"`
	DelayedDataCheckConfig *DelayedDataCheckConfigAPIModel `json:"delayed_data_check_config,omitempty"`
	Frequency              string                          `json:"frequency,omitempty"`
	Indices                []string                        `json:"indices,omitempty"`
	IndicesOptions         *IndicesOptionsAPIModel         `json:"indices_options,omitempty"`
	JobID                  string                          `json:"job_id,omitempty"`
	MaxEmptySearches       *int64                          `json:"max_empty_searches,omitempty"`
	Query                  map[string]interface{}          `json:"query,omitempty"`
	QueryDelay             string                          `json:"query_delay,omitempty"`
	RuntimeMappings        map[string]interface{}          `json:"runtime_mappings,omitempty"`
	ScriptFields           map[string]interface{}          `json:"script_fields,omitempty"`
	ScrollSize             *int64                          `json:"scroll_size,omitempty"`
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
