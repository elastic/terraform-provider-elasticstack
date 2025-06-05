package anomaly_detector

// apiCreateAnomalyDetectorRequest represents the request body for creating an ML job
type apiAnalysisLimits struct {
	ModelMemoryLimit            *string `json:"model_memory_limit,omitempty"`
	CategorizationExamplesLimit *int64  `json:"categorization_examples_limit,omitempty"`
}

type apiCreateAnomalyDetectorRequest struct {
	Description                          *string                `json:"description,omitempty"`
	Groups                               []string               `json:"groups,omitempty"`
	AnalysisConfig                       *apiAnalysisConfig     `json:"analysis_config"`
	DataDescription                      *apiDataDescription    `json:"data_description"`
	ModelPlotConfig                      *apiModelPlotConfig    `json:"model_plot_config,omitempty"`
	ResultsIndexName                     *string                `json:"results_index_name,omitempty"`
	AnalysisLimits                       *apiAnalysisLimits     `json:"analysis_limits,omitempty"`
	ModelSnapshotRetentionDays           *int64                 `json:"model_snapshot_retention_days,omitempty"`
	ResultsRetentionDays                 *int64                 `json:"results_retention_days,omitempty"`
	AllowLazyOpen                        *bool                  `json:"allow_lazy_open,omitempty"`
	CategorizationExamplesLimit          *int64                 `json:"categorization_examples_limit,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                 `json:"daily_model_snapshot_retention_after_days,omitempty"`
	CustomSettings                       map[string]interface{} `json:"custom_settings,omitempty"`
}

type apiAnalysisConfig struct {
	BucketSpan              string        `json:"bucket_span"`
	Detectors               []apiDetector `json:"detectors"`
	Influencers             []string      `json:"influencers,omitempty"`
	CategorizationFieldName *string       `json:"categorization_field_name,omitempty"`
	SummaryCountFieldName   *string       `json:"summary_count_field_name,omitempty"`
	Latency                 *string       `json:"latency,omitempty"`
}

type apiDetector struct {
	Function            string          `json:"function"`
	FieldName           *string         `json:"field_name,omitempty"`
	ByFieldName         *string         `json:"by_field_name,omitempty"`
	PartitionFieldName  *string         `json:"partition_field_name,omitempty"`
	DetectorDescription *string         `json:"detector_description,omitempty"`
	UseNull             *bool           `json:"use_null,omitempty"`
	ExcludeFrequent     *string         `json:"exclude_frequent,omitempty"`
	CustomRules         []apiCustomRule `json:"custom_rules,omitempty"`
}

type apiDataDescription struct {
	TimeField  string  `json:"time_field"`
	TimeFormat *string `json:"time_format,omitempty"`
}

type apiRuleCondition struct {
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type apiCustomRule struct {
	Actions    []string           `json:"actions"`
	Scope      *string            `json:"scope,omitempty"` // Or applies_to, whatever the API expects. Making it a pointer for omitempty.
	Conditions []apiRuleCondition `json:"conditions,omitempty"`
}

type apiModelPlotConfig struct {
	Enabled bool `json:"enabled"`
}

// apiGetAnomalyDetectorResponse represents the top-level response from GET /_ml/anomaly_detectors
type apiGetAnomalyDetectorResponse struct {
	Count int64    `json:"count"`
	Jobs  []apiJob `json:"jobs"`
}

// apiJob represents a single ML job definition from the API response for GET
type apiJob struct {
	JobID                                string                 `json:"job_id"`
	Description                          *string                `json:"description,omitempty"`
	Groups                               []string               `json:"groups,omitempty"`
	AnalysisConfig                       *apiAnalysisConfig     `json:"analysis_config"`
	DataDescription                      *apiDataDescription    `json:"data_description"`
	ModelPlotConfig                      *apiModelPlotConfig    `json:"model_plot_config,omitempty"`
	ResultsIndexName                     *string                `json:"results_index_name,omitempty"`
	AnalysisLimits                       *apiAnalysisLimits     `json:"analysis_limits,omitempty"`
	ModelSnapshotRetentionDays           *int64                 `json:"model_snapshot_retention_days,omitempty"`
	ResultsRetentionDays                 *int64                 `json:"results_retention_days,omitempty"`
	AllowLazyOpen                        *bool                  `json:"allow_lazy_open,omitempty"`
	CategorizationExamplesLimit          *int64                 `json:"categorization_examples_limit,omitempty"`
	DailyModelSnapshotRetentionAfterDays *int64                 `json:"daily_model_snapshot_retention_after_days,omitempty"`
	CustomSettings                       map[string]interface{} `json:"custom_settings,omitempty"`
}
