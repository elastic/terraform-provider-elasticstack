package anomaly_detector

// apiCreateAnomalyDetectorRequest represents the request body for creating an ML job
type apiCreateAnomalyDetectorRequest struct {
	Description      *string             `json:"description,omitempty"`
	Groups           []string            `json:"groups,omitempty"`
	AnalysisConfig   *apiAnalysisConfig  `json:"analysis_config"`
	DataDescription  *apiDataDescription `json:"data_description"`
	ModelPlotConfig  *apiModelPlotConfig `json:"model_plot_config,omitempty"`
	ResultsIndexName *string             `json:"results_index_name,omitempty"`
}

type apiAnalysisConfig struct {
	BucketSpan  string        `json:"bucket_span"`
	Detectors   []apiDetector `json:"detectors"`
	Influencers []string      `json:"influencers,omitempty"`
}

type apiDetector struct {
	Function           string  `json:"function"`
	FieldName          *string `json:"field_name,omitempty"`
	ByFieldName        *string `json:"by_field_name,omitempty"`
	PartitionFieldName *string `json:"partition_field_name,omitempty"`
}

type apiDataDescription struct {
	TimeField string `json:"time_field"`
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
	JobID            string              `json:"job_id"`
	Description      *string             `json:"description,omitempty"`
	Groups           []string            `json:"groups,omitempty"`
	AnalysisConfig   *apiAnalysisConfig  `json:"analysis_config"`
	DataDescription  *apiDataDescription `json:"data_description"`
	ModelPlotConfig  *apiModelPlotConfig `json:"model_plot_config,omitempty"`
	ResultsIndexName *string             `json:"results_index_name,omitempty"`
}
