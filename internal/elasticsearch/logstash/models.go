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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data holds the Terraform state for the logstash pipeline resource.
type Data struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`

	PipelineID       types.String         `tfsdk:"pipeline_id"`
	Description      types.String         `tfsdk:"description"`
	LastModified     types.String         `tfsdk:"last_modified"`
	Pipeline         types.String         `tfsdk:"pipeline"`
	PipelineMetadata jsontypes.Normalized `tfsdk:"pipeline_metadata"`
	Username         types.String         `tfsdk:"username"`

	// Pipeline settings
	PipelineBatchDelay         types.Int64  `tfsdk:"pipeline_batch_delay"`
	PipelineBatchSize          types.Int64  `tfsdk:"pipeline_batch_size"`
	PipelineEcsCompatibility   types.String `tfsdk:"pipeline_ecs_compatibility"`
	PipelineOrdered            types.String `tfsdk:"pipeline_ordered"`
	PipelinePluginClassloaders types.Bool   `tfsdk:"pipeline_plugin_classloaders"`
	PipelineUnsafeShutdown     types.Bool   `tfsdk:"pipeline_unsafe_shutdown"`
	PipelineWorkers            types.Int64  `tfsdk:"pipeline_workers"`
	QueueCheckpointAcks        types.Int64  `tfsdk:"queue_checkpoint_acks"`
	QueueCheckpointRetry       types.Bool   `tfsdk:"queue_checkpoint_retry"`
	QueueCheckpointWrites      types.Int64  `tfsdk:"queue_checkpoint_writes"`
	QueueDrain                 types.Bool   `tfsdk:"queue_drain"`
	QueueMaxBytes              types.String `tfsdk:"queue_max_bytes"`
	QueueMaxEvents             types.Int64  `tfsdk:"queue_max_events"`
	QueuePageCapacity          types.String `tfsdk:"queue_page_capacity"`
	QueueType                  types.String `tfsdk:"queue_type"`
}

func (d Data) GetID() types.String                    { return d.ID }
func (d Data) GetResourceID() types.String            { return d.PipelineID }
func (d Data) GetElasticsearchConnection() types.List { return d.ElasticsearchConnection }
