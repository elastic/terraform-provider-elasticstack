package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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

	if fwDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML job stats: %s", jobId)); fwDiags.HasError() {
		diags.Append(fwDiags...)
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
