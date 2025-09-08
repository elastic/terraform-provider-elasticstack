package anomaly_detector

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

// AnomalyDetectorJobUpdateAPIModel represents the API model for updating ML anomaly detection jobs
// This includes only the fields that can be updated after job creation
type AnomalyDetectorJobUpdateAPIModel struct {
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
