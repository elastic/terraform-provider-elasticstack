package models

import "time"

// Datafeed represents the complete datafeed as returned by the Elasticsearch API
type Datafeed struct {
	DatafeedId             string                  `json:"datafeed_id"`
	JobId                  string                  `json:"job_id"`
	Indices                []string                `json:"indices"`
	Query                  map[string]interface{}  `json:"query,omitempty"`
	Aggregations           map[string]interface{}  `json:"aggregations,omitempty"`
	ScriptFields           map[string]interface{}  `json:"script_fields,omitempty"`
	RuntimeMappings        map[string]interface{}  `json:"runtime_mappings,omitempty"`
	ScrollSize             *int                    `json:"scroll_size,omitempty"`
	ChunkingConfig         *ChunkingConfig         `json:"chunking_config,omitempty"`
	Frequency              *string                 `json:"frequency,omitempty"`
	QueryDelay             *string                 `json:"query_delay,omitempty"`
	DelayedDataCheckConfig *DelayedDataCheckConfig `json:"delayed_data_check_config,omitempty"`
	MaxEmptySearches       *int                    `json:"max_empty_searches,omitempty"`
	IndicesOptions         *IndicesOptions         `json:"indices_options,omitempty"`
	Authorization          *Authorization          `json:"authorization,omitempty"`
}

// ChunkingConfig represents the chunking configuration for datafeeds
type ChunkingConfig struct {
	Mode     string `json:"mode"`                // "auto", "manual", "off"
	TimeSpan string `json:"time_span,omitempty"` // Only for manual mode
}

// DelayedDataCheckConfig represents the delayed data check configuration
type DelayedDataCheckConfig struct {
	Enabled     *bool   `json:"enabled,omitempty"`
	CheckWindow *string `json:"check_window,omitempty"`
}

// IndicesOptions represents the indices options for search
type IndicesOptions struct {
	ExpandWildcards   []string `json:"expand_wildcards,omitempty"`
	IgnoreUnavailable *bool    `json:"ignore_unavailable,omitempty"`
	AllowNoIndices    *bool    `json:"allow_no_indices,omitempty"`
	IgnoreThrottled   *bool    `json:"ignore_throttled,omitempty"`
}

// Authorization represents authorization headers stored with the datafeed
type Authorization struct {
	Roles []string `json:"roles,omitempty"`
}

// DatafeedCreateRequest represents the request body for creating a datafeed
type DatafeedCreateRequest struct {
	JobId                  string                  `json:"job_id"`
	Indices                []string                `json:"indices"`
	Query                  map[string]interface{}  `json:"query,omitempty"`
	Aggregations           map[string]interface{}  `json:"aggregations,omitempty"`
	ScriptFields           map[string]interface{}  `json:"script_fields,omitempty"`
	RuntimeMappings        map[string]interface{}  `json:"runtime_mappings,omitempty"`
	ScrollSize             *int                    `json:"scroll_size,omitempty"`
	ChunkingConfig         *ChunkingConfig         `json:"chunking_config,omitempty"`
	Frequency              *string                 `json:"frequency,omitempty"`
	QueryDelay             *string                 `json:"query_delay,omitempty"`
	DelayedDataCheckConfig *DelayedDataCheckConfig `json:"delayed_data_check_config,omitempty"`
	MaxEmptySearches       *int                    `json:"max_empty_searches,omitempty"`
	IndicesOptions         *IndicesOptions         `json:"indices_options,omitempty"`
}

// DatafeedUpdateRequest represents the request body for updating a datafeed
type DatafeedUpdateRequest struct {
	JobId                  *string                 `json:"job_id,omitempty"`
	Indices                []string                `json:"indices,omitempty"`
	Query                  map[string]interface{}  `json:"query,omitempty"`
	Aggregations           map[string]interface{}  `json:"aggregations,omitempty"`
	ScriptFields           map[string]interface{}  `json:"script_fields,omitempty"`
	RuntimeMappings        map[string]interface{}  `json:"runtime_mappings,omitempty"`
	ScrollSize             *int                    `json:"scroll_size,omitempty"`
	ChunkingConfig         *ChunkingConfig         `json:"chunking_config,omitempty"`
	Frequency              *string                 `json:"frequency,omitempty"`
	QueryDelay             *string                 `json:"query_delay,omitempty"`
	DelayedDataCheckConfig *DelayedDataCheckConfig `json:"delayed_data_check_config,omitempty"`
	MaxEmptySearches       *int                    `json:"max_empty_searches,omitempty"`
	IndicesOptions         *IndicesOptions         `json:"indices_options,omitempty"`
}

// DatafeedResponse represents the response from the datafeed API
type DatafeedResponse struct {
	DatafeedId             string                  `json:"datafeed_id"`
	JobId                  string                  `json:"job_id"`
	Indices                []string                `json:"indices"`
	Query                  map[string]interface{}  `json:"query"`
	Aggregations           map[string]interface{}  `json:"aggregations,omitempty"`
	ScriptFields           map[string]interface{}  `json:"script_fields,omitempty"`
	RuntimeMappings        map[string]interface{}  `json:"runtime_mappings,omitempty"`
	ScrollSize             int                     `json:"scroll_size"`
	ChunkingConfig         ChunkingConfig          `json:"chunking_config"`
	Frequency              string                  `json:"frequency"`
	QueryDelay             string                  `json:"query_delay"`
	DelayedDataCheckConfig *DelayedDataCheckConfig `json:"delayed_data_check_config,omitempty"`
	MaxEmptySearches       *int                    `json:"max_empty_searches,omitempty"`
	IndicesOptions         *IndicesOptions         `json:"indices_options,omitempty"`
	Authorization          *Authorization          `json:"authorization,omitempty"`
}

// DatafeedStatsResponse represents the response from the datafeed stats API
type DatafeedStatsResponse struct {
	Datafeeds []DatafeedStats `json:"datafeeds"`
}

// DatafeedStats represents the statistics for a single datafeed
type DatafeedStats struct {
	DatafeedId            string           `json:"datafeed_id"`
	State                 string           `json:"state"`
	Node                  *DatafeedNode    `json:"node,omitempty"`
	AssignmentExplanation *string          `json:"assignment_explanation,omitempty"`
	RunningState          *DatafeedRunning `json:"running_state,omitempty"`
}

// DatafeedNode represents the node information for a datafeed
type DatafeedNode struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	EphemeralId      string            `json:"ephemeral_id"`
	TransportAddress string            `json:"transport_address"`
	Attributes       map[string]string `json:"attributes"`
}

// DatafeedRunning represents the running state of a datafeed
type DatafeedRunning struct {

// MLJobStats represents the statistics structure for an ML job
type MLJobStats struct {
	Jobs []MLJob `json:"jobs"`
}

// MLJob represents a single ML job in the stats response
type MLJob struct {
	JobId string     `json:"job_id"`
	State string     `json:"state"`
	Node  *MLJobNode `json:"node,omitempty"`
}

// MLJobNode represents the node information for an ML job
type MLJobNode struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}
