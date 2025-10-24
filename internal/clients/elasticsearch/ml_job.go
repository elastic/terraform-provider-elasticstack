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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)
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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to update ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)

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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to delete ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)

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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to stop ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)

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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to start ML datafeed: %s", datafeedId))
	diags.Append(fwDiags...)

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

	fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML datafeed stats: %s", datafeedId))
	diags.Append(fwDiags...)
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
