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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = newLogstashPipelineResource()
	_ resource.ResourceWithConfigure   = newLogstashPipelineResource()
	_ resource.ResourceWithImportState = newLogstashPipelineResource()
)

type logstashPipelineResource struct {
	*entitycore.ElasticsearchResource[Data]
}

func newLogstashPipelineResource() *logstashPipelineResource {
	return &logstashPipelineResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[Data]("logstash_pipeline", entitycore.ElasticsearchResourceOptions[Data]{
			Schema: GetSchema,
			Read:   readLogstashPipeline,
			Delete: deleteLogstashPipeline,
			Create: writeLogstashPipeline,
			Update: writeLogstashPipeline,
		}),
	}
}

// NewLogstashPipelineResource returns the PF resource constructor for
// elasticstack_elasticsearch_logstash_pipeline.
func NewLogstashPipelineResource() resource.Resource {
	return newLogstashPipelineResource()
}

func (r *logstashPipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
