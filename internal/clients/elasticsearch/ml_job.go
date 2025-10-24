package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

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

// OpenMLJob opens a machine learning job
func OpenMLJob(ctx context.Context, apiClient *clients.ApiClient, jobId string) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	res, err := esClient.ML.OpenJob(jobId, esClient.ML.OpenJob.WithContext(ctx))
	if err != nil {
		diags.AddError("Failed to open ML job", err.Error())
		return diags
	}
	defer res.Body.Close()
	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to open ML job: %s", jobId))
	diags.Append(fwDiags...)

	return diags
}

// PutDatafeed creates a machine learning datafeed
func PutDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string, createRequest models.DatafeedCreateRequest) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	// Send create request to Elasticsearch using helper function
	body, err := json.Marshal(createRequest)
	if err != nil {
		diags.AddError("Error marshaling request", err.Error())
		return diags
	}

	res, err := esClient.ML.PutDatafeed(bytes.NewReader(body), datafeedId, esClient.ML.PutDatafeed.WithContext(ctx))
	if err != nil {
		diags.AddError("Failed to create ML datafeed", err.Error())
		return diags
	}
	defer res.Body.Close()

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to create ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)

	return diags
}

// CloseMLJob closes a machine learning job
func CloseMLJob(ctx context.Context, apiClient *clients.ApiClient, jobId string, force bool, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	options := []func(*esapi.MLCloseJobRequest){
		esClient.ML.CloseJob.WithContext(ctx),
		esClient.ML.CloseJob.WithForce(force),
		esClient.ML.CloseJob.WithAllowNoMatch(true),
	}

	if timeout > 0 {
		options = append(options, esClient.ML.CloseJob.WithTimeout(timeout))
	}

	res, err := esClient.ML.CloseJob(jobId, options...)
	if err != nil {
		diags.AddError("Failed to close ML job", err.Error())
		return diags
	}
	defer res.Body.Close()

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to close ML job: %s", jobId))
	diags.Append(fwDiags...)

	return diags
}

// GetMLJobStats retrieves the stats for a specific machine learning job
func GetMLJobStats(ctx context.Context, apiClient *clients.ApiClient, jobId string) (*MLJob, diag.Diagnostics) {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}
	options := []func(*esapi.MLGetJobStatsRequest){
		esClient.ML.GetJobStats.WithContext(ctx),
		esClient.ML.GetJobStats.WithJobID(jobId),
		esClient.ML.GetJobStats.WithAllowNoMatch(true),
	}

	res, err := esClient.ML.GetJobStats(options...)
	if err != nil {
		diags.AddError("Failed to get ML job stats", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, diags
	}
	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML job stats: %s", jobId))...)
	if diags.HasError() {
		return nil, diags
	}

	var jobStats MLJobStats
	if err := json.NewDecoder(res.Body).Decode(&jobStats); err != nil {
		diags.AddError("Failed to decode ML job stats response", err.Error())
		return nil, diags
	}

	// Find the specific job in the response
	for _, job := range jobStats.Jobs {
		if job.JobId == jobId {
			return &job, diags
		}
	}

	// Job not found in response
	return nil, diags
}

// GetDatafeed retrieves a machine learning datafeed
func GetDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string) (*models.Datafeed, diag.Diagnostics) {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}

	options := []func(*esapi.MLGetDatafeedsRequest){
		esClient.ML.GetDatafeeds.WithContext(ctx),
		esClient.ML.GetDatafeeds.WithDatafeedID(datafeedId),
		esClient.ML.GetDatafeeds.WithAllowNoMatch(true),
	}

	res, err := esClient.ML.GetDatafeeds(options...)
	if err != nil {
		diags.AddError("Failed to get ML datafeed", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, diags
	}

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML datafeed: %s", datafeedId))...)
	if diags.HasError() {
		return nil, diags
	}

	var response struct {
		Datafeeds []models.Datafeed `json:"datafeeds"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Failed to decode ML datafeed response", err.Error())
		return nil, diags
	}

	// Find the specific datafeed in the response
	for _, df := range response.Datafeeds {
		if df.DatafeedId == datafeedId {
			return &df, diags
		}
	}

	return nil, diags
}

// UpdateDatafeed updates a machine learning datafeed
func UpdateDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string, request models.DatafeedUpdateRequest) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	// Marshal the update request
	body, err := json.Marshal(request)
	if err != nil {
		diags.AddError("Error marshaling update request", err.Error())
		return diags
	}

	res, err := esClient.ML.UpdateDatafeed(bytes.NewReader(body), datafeedId, esClient.ML.UpdateDatafeed.WithContext(ctx))
	if err != nil {
		diags.AddError("Failed to update ML datafeed", err.Error())
		return diags
	}
	defer res.Body.Close()

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to update ML datafeed: %s", datafeedId))...)

	return diags
}

// DeleteDatafeed deletes a machine learning datafeed
func DeleteDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string, force bool) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	options := []func(*esapi.MLDeleteDatafeedRequest){
		esClient.ML.DeleteDatafeed.WithContext(ctx),
		esClient.ML.DeleteDatafeed.WithForce(force),
	}

	res, err := esClient.ML.DeleteDatafeed(datafeedId, options...)
	if err != nil {
		diags.AddError("Failed to delete ML datafeed", err.Error())
		return diags
	}
	defer res.Body.Close()

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to delete ML datafeed: %s", datafeedId))...)

	return diags
}

// StopDatafeed stops a machine learning datafeed
func StopDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string, force bool, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	options := []func(*esapi.MLStopDatafeedRequest){
		esClient.ML.StopDatafeed.WithContext(ctx),
		esClient.ML.StopDatafeed.WithForce(force),
		esClient.ML.StopDatafeed.WithAllowNoMatch(true),
	}

	if timeout > 0 {
		options = append(options, esClient.ML.StopDatafeed.WithTimeout(timeout))
	}

	res, err := esClient.ML.StopDatafeed(datafeedId, options...)
	if err != nil {
		diags.AddError("Failed to stop ML datafeed", err.Error())
		return diags
	}
	defer res.Body.Close()

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to stop ML datafeed: %s", datafeedId))...)

	return diags
}

// StartDatafeed starts a machine learning datafeed
func StartDatafeed(ctx context.Context, apiClient *clients.ApiClient, datafeedId string, start string, end string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
	}

	options := []func(*esapi.MLStartDatafeedRequest){
		esClient.ML.StartDatafeed.WithContext(ctx),
	}

	if start != "" {
		options = append(options, esClient.ML.StartDatafeed.WithStart(start))
	}

	if end != "" {
		options = append(options, esClient.ML.StartDatafeed.WithEnd(end))
	}

	if timeout > 0 {
		options = append(options, esClient.ML.StartDatafeed.WithTimeout(timeout))
	}

	res, err := esClient.ML.StartDatafeed(datafeedId, options...)
	if err != nil {
		diags.AddError("Failed to start ML datafeed", err.Error())
		return diags
	}
	defer res.Body.Close()

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to start ML datafeed: %s", datafeedId))...)

	return diags
}

// GetDatafeedStats retrieves the statistics for a machine learning datafeed
func GetDatafeedStats(ctx context.Context, apiClient *clients.ApiClient, datafeedId string) (*models.DatafeedStats, diag.Diagnostics) {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return nil, diags
	}

	options := []func(*esapi.MLGetDatafeedStatsRequest){
		esClient.ML.GetDatafeedStats.WithContext(ctx),
		esClient.ML.GetDatafeedStats.WithDatafeedID(datafeedId),
		esClient.ML.GetDatafeedStats.WithAllowNoMatch(true),
	}

	res, err := esClient.ML.GetDatafeedStats(options...)
	if err != nil {
		diags.AddError("Failed to get ML datafeed stats", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, diags
	}

	diags.Append(diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML datafeed stats: %s", datafeedId))...)
	if diags.HasError() {
		return nil, diags
	}

	var response models.DatafeedStatsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Failed to decode ML datafeed stats response", err.Error())
		return nil, diags
	}

	// Since we're requesting stats for a specific datafeed ID, we expect exactly one result
	if len(response.Datafeeds) == 0 {
		return nil, diags // Datafeed not found, return nil without error
	}

	if len(response.Datafeeds) > 1 {
		diags.AddError("Unexpected response", fmt.Sprintf("Expected single datafeed stats for ID %s, got %d", datafeedId, len(response.Datafeeds)))
		return nil, diags
	}

	return &response.Datafeeds[0], diags
}
