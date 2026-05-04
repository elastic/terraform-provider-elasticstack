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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// OpenMLJob opens a machine learning job
func OpenMLJob(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Ml.OpenJob(jobID).Do(ctx)
	if err != nil {
		diags.AddError("Failed to open ML job", fmt.Sprintf("Unable to open ML job: %s — %s", jobID, err.Error()))
		return diags
	}

	return diags
}

// DatafeedRequest is the request body for creating or updating an ML datafeed.
//
// Query, Aggregations, ScriptFields, and RuntimeMappings are json.RawMessage so
// they are embedded as-is without round-tripping through typed structs.  This is
// required for Query in particular: types.Query.UnmarshalJSON normalises term
// shorthand ({"term":{"f":"v"}}) to the verbose form ({"term":{"f":{"value":"v"}}}),
// which would produce a permanent diff in Terraform state.
type DatafeedRequest struct {
	JobID                  string                          `json:"job_id,omitempty"`
	Indices                []string                        `json:"indices"`
	Query                  json.RawMessage                 `json:"query,omitempty"`
	Aggregations           json.RawMessage                 `json:"aggregations,omitempty"`
	ScriptFields           json.RawMessage                 `json:"script_fields,omitempty"`
	RuntimeMappings        json.RawMessage                 `json:"runtime_mappings,omitempty"`
	ScrollSize             *int                            `json:"scroll_size,omitempty"`
	Frequency              string                          `json:"frequency,omitempty"`
	QueryDelay             string                          `json:"query_delay,omitempty"`
	MaxEmptySearches       *int                            `json:"max_empty_searches,omitempty"`
	ChunkingConfig         *DatafeedChunkingConfig         `json:"chunking_config,omitempty"`
	DelayedDataCheckConfig *DatafeedDelayedDataCheckConfig `json:"delayed_data_check_config,omitempty"`
	IndicesOptions         *DatafeedIndicesOptions         `json:"indices_options,omitempty"`
}

// DatafeedChunkingConfig is the chunking configuration within a DatafeedRequest.
type DatafeedChunkingConfig struct {
	Mode     string `json:"mode"`
	TimeSpan string `json:"time_span,omitempty"`
}

// DatafeedDelayedDataCheckConfig is the delayed-data-check configuration within a DatafeedRequest.
type DatafeedDelayedDataCheckConfig struct {
	Enabled     bool   `json:"enabled"`
	CheckWindow string `json:"check_window,omitempty"`
}

// DatafeedIndicesOptions controls how a datafeed accesses its indices.
type DatafeedIndicesOptions struct {
	ExpandWildcards   []string `json:"expand_wildcards,omitempty"`
	IgnoreUnavailable *bool    `json:"ignore_unavailable,omitempty"`
	AllowNoIndices    *bool    `json:"allow_no_indices,omitempty"`
	IgnoreThrottled   *bool    `json:"ignore_throttled,omitempty"`
}

// PutDatafeed creates a machine learning datafeed.
//
// We use .Raw() to send the marshalled DatafeedRequest as-is so that Query is
// preserved exactly as the user wrote it (see DatafeedRequest for details).
func PutDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string, req DatafeedRequest) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	body, err := json.Marshal(req)
	if err != nil {
		diags.AddError("Failed to marshal datafeed request", err.Error())
		return diags
	}

	_, err = typedClient.Ml.PutDatafeed(datafeedID).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML datafeed", fmt.Sprintf("Unable to create ML datafeed: %s — %s", datafeedID, err.Error()))
		return diags
	}

	return diags
}

// CloseMLJob closes a machine learning job
func CloseMLJob(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string, force bool, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Ml.CloseJob(jobID).
		Force(force).
		AllowNoMatch(true)

	if timeout > 0 {
		req.Timeout(durationToMsString(timeout))
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to close ML job", fmt.Sprintf("Unable to close ML job: %s — %s", jobID, err.Error()))
		return diags
	}

	return diags
}

// GetMLJobStats retrieves the stats for a specific machine learning job
func GetMLJobStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string) (*types.JobStats, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Ml.GetJobStats().JobId(jobID).AllowNoMatch(true).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Failed to get ML job stats", fmt.Sprintf("Unable to get ML job stats: %s — %s", jobID, err.Error()))
		return nil, diags
	}

	for i := range res.Jobs {
		if res.Jobs[i].JobId == jobID {
			return &res.Jobs[i], diags
		}
	}

	return nil, diags
}

// MLDatafeedResponse wraps the types.MLDatafeed with a raw JSON query that
// preserves the original form returned by Elasticsearch without re-normalisation
// through the typed Query struct (e.g. term shorthand vs verbose value form).
type MLDatafeedResponse struct {
	*types.MLDatafeed
	QueryRaw json.RawMessage
}

// rawDatafeedListResponse is a minimal struct used to decode the get-datafeeds
// response while preserving the raw query bytes for each datafeed.
type rawDatafeedListResponse struct {
	Count     int                   `json:"count"`
	Datafeeds []rawDatafeedDocument `json:"datafeeds"`
}

type rawDatafeedDocument struct {
	DatafeedID string          `json:"datafeed_id"`
	Query      json.RawMessage `json:"query"`
}

// GetDatafeed retrieves a machine learning datafeed.
//
// We use .Perform() (raw *http.Response) rather than .Do() so that the query
// is preserved exactly as returned by Elasticsearch.  The typed types.Query
// struct normalises term shorthand ({"term":{"f":"v"}}) to the verbose form
// ({"term":{"f":{"value":"v"}}}) on marshal, which would produce a permanent
// diff in Terraform state.
func GetDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string) (*MLDatafeedResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Ml.GetDatafeeds().DatafeedId(datafeedID).AllowNoMatch(true).Perform(ctx)
	if err != nil {
		diags.AddError("Failed to get ML datafeed", fmt.Sprintf("Unable to get ML datafeed: %s — %s", datafeedID, err.Error()))
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, diags
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, fmt.Sprintf("Unable to get ML datafeed: %s", datafeedID)); d.HasError() {
		return nil, d
	}

	// Decode the typed response for all fields except query.
	var typedResponse struct {
		Count     int                `json:"count"`
		Datafeeds []types.MLDatafeed `json:"datafeeds"`
	}
	// We need both typed and raw simultaneously, so decode the body twice via a
	// buffered copy.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		diags.AddError("Failed to read ML datafeed response", err.Error())
		return nil, diags
	}

	if err := json.Unmarshal(body, &typedResponse); err != nil {
		diags.AddError("Failed to decode ML datafeed response", err.Error())
		return nil, diags
	}

	var rawResponse rawDatafeedListResponse
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		diags.AddError("Failed to decode ML datafeed raw response", err.Error())
		return nil, diags
	}

	// Both slices are decoded from the same JSON body and share the same ordering.
	for i := range typedResponse.Datafeeds {
		if typedResponse.Datafeeds[i].DatafeedId == datafeedID {
			resp := &MLDatafeedResponse{MLDatafeed: &typedResponse.Datafeeds[i]}
			if i < len(rawResponse.Datafeeds) {
				resp.QueryRaw = rawResponse.Datafeeds[i].Query
			}
			return resp, diags
		}
	}

	return nil, diags
}

// UpdateDatafeed updates a machine learning datafeed.
//
// We use .Raw() to send the marshalled DatafeedRequest as-is (see PutDatafeed).
// The caller must leave JobID empty — update_datafeed does not accept job_id.
func UpdateDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string, req DatafeedRequest) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	body, err := json.Marshal(req)
	if err != nil {
		diags.AddError("Failed to marshal datafeed request", err.Error())
		return diags
	}

	_, err = typedClient.Ml.UpdateDatafeed(datafeedID).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		diags.AddError("Failed to update ML datafeed", fmt.Sprintf("Unable to update ML datafeed: %s — %s", datafeedID, err.Error()))
		return diags
	}

	return diags
}

// DeleteDatafeed deletes a machine learning datafeed
func DeleteDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string, force bool) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Ml.DeleteDatafeed(datafeedID).Force(force).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Failed to delete ML datafeed", fmt.Sprintf("Unable to delete ML datafeed: %s — %s", datafeedID, err.Error()))
		return diags
	}

	return diags
}

// StopDatafeed stops a machine learning datafeed
func StopDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string, force bool, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Ml.StopDatafeed(datafeedID).
		Force(force).
		AllowNoMatch(true)

	if timeout > 0 {
		req.Timeout(durationToMsString(timeout))
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to stop ML datafeed", fmt.Sprintf("Unable to stop ML datafeed: %s — %s", datafeedID, err.Error()))
		return diags
	}

	return diags
}

// StartDatafeed starts a machine learning datafeed
func StartDatafeed(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string, start string, end string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Ml.StartDatafeed(datafeedID)

	if start != "" {
		req.Start(start)
	}

	if end != "" {
		req.End(end)
	}

	if timeout > 0 {
		req.Timeout(durationToMsString(timeout))
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to start ML datafeed", fmt.Sprintf("Unable to start ML datafeed: %s — %s", datafeedID, err.Error()))
		return diags
	}

	return diags
}

// GetDatafeedStats retrieves the statistics for a machine learning datafeed
func GetDatafeedStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, datafeedID string) (*types.DatafeedStats, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Ml.GetDatafeedStats().DatafeedId(datafeedID).AllowNoMatch(true).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Failed to get ML datafeed stats", fmt.Sprintf("Unable to get ML datafeed stats: %s — %s", datafeedID, err.Error()))
		return nil, diags
	}

	// Since we're requesting stats for a specific datafeed ID, we expect exactly one result
	if len(res.Datafeeds) == 0 {
		return nil, diags
	}

	if len(res.Datafeeds) > 1 {
		diags.AddError("Unexpected response", fmt.Sprintf("Expected single datafeed stats for ID %s, got %d", datafeedID, len(res.Datafeeds)))
		return nil, diags
	}

	return &res.Datafeeds[0], diags
}
