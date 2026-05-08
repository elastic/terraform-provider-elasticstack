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

package logstash

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func writeLogstashPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, pipelineID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, sdkDiags := client.ID(ctx, pipelineID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	pipeline := models.LogstashPipeline{
		PipelineID:   pipelineID,
		LastModified: typeutils.FormatStrictDateTime(time.Now().UTC()),
		Pipeline:     data.Pipeline.ValueString(),
	}

	if typeutils.IsKnown(data.Description) {
		pipeline.Description = data.Description.ValueString()
	}

	if typeutils.IsKnown(data.Username) {
		pipeline.Username = data.Username.ValueString()
	}

	// Parse pipeline_metadata from JSON string to map.
	metaStr := "{}"
	if typeutils.IsKnown(data.PipelineMetadata) && data.PipelineMetadata.ValueString() != "" {
		metaStr = data.PipelineMetadata.ValueString()
	}
	var pipelineMetadata map[string]any
	if err := json.Unmarshal([]byte(metaStr), &pipelineMetadata); err != nil {
		diags.AddError("Error parsing pipeline_metadata", err.Error())
		var zero Data
		return zero, diags
	}
	pipeline.PipelineMetadata = pipelineMetadata

	// Expand typed settings fields to flat API map.
	pipeline.PipelineSettings = expandSettings(data)

	sdkDiags = elasticsearch.PutLogstashPipeline(ctx, client, &pipeline)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	data.ID = types.StringValue(id.String())
	return data, diags
}
