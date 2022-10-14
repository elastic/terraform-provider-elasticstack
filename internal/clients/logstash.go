package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutLogstashPipeline(ctx context.Context, logstashPipeline *models.LogstashPipeline) diag.Diagnostics {
	var diags diag.Diagnostics
	logstashPipelineBytes, err := json.Marshal(logstashPipeline)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.LogstashPutPipeline(logstashPipeline.PipelineID, bytes.NewReader(logstashPipelineBytes), a.es.LogstashPutPipeline.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update logstash pipeline"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetLogstashPipeline(ctx context.Context, pipelineID string) (*models.LogstashPipeline, diag.Diagnostics) {
	var diags diag.Diagnostics
	res, err := a.es.LogstashGetPipeline(pipelineID, a.es.LogstashGetPipeline.WithContext(ctx))
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

func (a *ApiClient) DeleteLogstashPipeline(ctx context.Context, pipeline_id string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.LogstashDeletePipeline(pipeline_id, a.es.LogstashDeletePipeline.WithContext(ctx))

	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete logstash pipeline"); diags.HasError() {
		return diags
	}
	return diags
}
