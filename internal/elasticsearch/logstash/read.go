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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readLogstashPipeline(ctx context.Context, client *clients.ElasticsearchScopedClient, pipelineID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	pipeline, pipelineDiags := elasticsearch.GetLogstashPipeline(ctx, client, pipelineID)
	diags.Append(pipelineDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if pipeline == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Logstash pipeline "%s" not found, removing from state`, pipelineID))
		return state, false, diags
	}

	var data Data
	data.ID = state.ID
	data.ElasticsearchConnection = state.ElasticsearchConnection

	data.PipelineID = types.StringValue(pipeline.PipelineID)
	data.Description = types.StringValue(pipeline.Description)
	data.LastModified = types.StringValue(pipeline.LastModified)
	data.Pipeline = types.StringValue(pipeline.Pipeline)
	data.Username = types.StringValue(pipeline.Username)

	if pipeline.PipelineMetadata != nil {
		data.PipelineMetadata = typeutils.MarshalToNormalized(pipeline.PipelineMetadata, path.Root("pipeline_metadata"), &diags)
		if diags.HasError() {
			return state, false, diags
		}
	} else {
		data.PipelineMetadata = jsontypes.NewNormalizedValue("{}")
	}

	// Flatten settings from API response into typed fields.
	if pipeline.PipelineSettings != nil {
		flattenSettings(pipeline.PipelineSettings, &data)
	}

	return data, true, diags
}
