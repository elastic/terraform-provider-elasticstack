package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutLogstashPipeline(ctx context.Context, apiClient *clients.ApiClient, logstashPipeline *models.LogstashPipeline) diag.Diagnostics {
	var diags diag.Diagnostics
	logstashPipelineBytes, err := json.Marshal(logstashPipeline)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.LogstashPutPipeline(logstashPipeline.PipelineID, bytes.NewReader(logstashPipelineBytes), esClient.LogstashPutPipeline.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update logstash pipeline"); diags.HasError() {
		return diags
	}

	return diags
}

func GetLogstashPipeline(ctx context.Context, apiClient *clients.ApiClient, pipelineID string) (*models.LogstashPipeline, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	res, err := esClient.LogstashGetPipeline(esClient.LogstashGetPipeline.WithDocumentID(pipelineID), esClient.LogstashGetPipeline.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to find logstash pipeline on cluster."); diags.HasError() {
		return nil, diags
	}

	logstashPipeline := make(map[string]models.LogstashPipeline)
	if err := json.NewDecoder(res.Body).Decode(&logstashPipeline); err != nil {
		return nil, diag.FromErr(err)
	}

	if logstashPipeline, ok := logstashPipeline[pipelineID]; ok {
		logstashPipeline.PipelineID = pipelineID
		return &logstashPipeline, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find logstash pipeline in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" logstash pipeline in the cluster`, pipelineID),
	})
	return nil, diags
}

func DeleteLogstashPipeline(ctx context.Context, apiClient *clients.ApiClient, pipeline_id string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.LogstashDeletePipeline(pipeline_id, esClient.LogstashDeletePipeline.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete logstash pipeline"); diags.HasError() {
		return diags
	}
	return diags
}
