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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutLogstashPipeline(ctx context.Context, apiClient *clients.APIClient, logstashPipeline *models.LogstashPipeline) diag.Diagnostics {
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
	if diags := diagutil.CheckError(res, "Unable to create or update logstash pipeline"); diags.HasError() {
		return diags
	}

	return diags
}

func GetLogstashPipeline(ctx context.Context, apiClient *clients.APIClient, pipelineID string) (*models.LogstashPipeline, diag.Diagnostics) {
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
	if diags := diagutil.CheckError(res, "Unable to find logstash pipeline on cluster."); diags.HasError() {
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

func DeleteLogstashPipeline(ctx context.Context, apiClient *clients.APIClient, pipelineID string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.LogstashDeletePipeline(pipelineID, esClient.LogstashDeletePipeline.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to delete logstash pipeline"); diags.HasError() {
		return diags
	}
	return diags
}
