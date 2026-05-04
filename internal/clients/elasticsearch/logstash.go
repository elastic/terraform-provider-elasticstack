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

func PutLogstashPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, logstashPipeline *models.LogstashPipeline) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// The typed client's types.LogstashPipeline.PipelineSettings only supports
	// a subset of the settings keys the provider uses (6 of ~17). Using the typed
	// struct would silently drop unsupported settings on both PUT and GET. We
	// therefore marshal the provider's model to JSON and send it via Raw() to
	// preserve all settings while still using the typed client for transport.
	b, err := json.Marshal(logstashPipeline)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := typedClient.Logstash.PutPipeline(logstashPipeline.PipelineID).Raw(bytes.NewReader(b)).Perform(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()

	if d := diagutil.CheckHTTPError(res, "Unable to create or update logstash pipeline"); d.HasError() {
		return d
	}

	return diags
}

func GetLogstashPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, pipelineID string) (*models.LogstashPipeline, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// We use .Perform() instead of .Do() because types.LogstashPipeline only
	// supports a subset of pipeline settings. Decoding into the provider's own
	// model via json.NewDecoder preserves all settings sent by Elasticsearch.
	// See the comment in PutLogstashPipeline for the same rationale.
	res, err := typedClient.Logstash.GetPipeline().Id(pipelineID).Perform(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckHTTPError(res, "Unable to find logstash pipeline on cluster."); d.HasError() {
		return nil, d
	}

	logstashPipeline := make(map[string]models.LogstashPipeline)
	if err := json.NewDecoder(res.Body).Decode(&logstashPipeline); err != nil {
		return nil, diag.FromErr(err)
	}

	if pipeline, ok := logstashPipeline[pipelineID]; ok {
		pipeline.PipelineID = pipelineID
		return &pipeline, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find logstash pipeline in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" logstash pipeline in the cluster`, pipelineID),
	})
	return nil, diags
}

func DeleteLogstashPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, pipelineID string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Use .Perform() for explicit 404 handling, consistent with GetLogstashPipeline.
	// The typed client's .Do() handles 404 internally, but making the status check
	// explicit in provider code makes the contract visible to maintainers.
	res, err := typedClient.Logstash.DeletePipeline(pipelineID).Perform(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete logstash pipeline",
			Detail:   fmt.Sprintf("Failed with: %s", err.Error()),
		})
		return diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return diags
	}

	if d := diagutil.CheckHTTPError(res, "Unable to delete logstash pipeline"); d.HasError() {
		return d
	}

	return diags
}
