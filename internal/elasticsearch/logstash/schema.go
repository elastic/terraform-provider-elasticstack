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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const defaultPipelineMetadata = `{"type":"logstash_pipeline","version":1}`

// GetSchema returns the PF schema for the logstash pipeline resource.
// The elasticsearch_connection block is injected by the envelope.
func GetSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manage Logstash Pipelines via Centralized Pipeline Management. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/logstash-apis.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
			},
			"pipeline_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the pipeline.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the pipeline.",
				Optional:            true,
				Computed:            true,
			},
			"last_modified": schema.StringAttribute{
				MarkdownDescription: "Date the pipeline was last updated.",
				Computed:            true,
			},
			"pipeline": schema.StringAttribute{
				MarkdownDescription: "Configuration for the pipeline.",
				Required:            true,
			},
			"pipeline_metadata": schema.StringAttribute{
				MarkdownDescription: "Optional JSON metadata about the pipeline.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultPipelineMetadata),
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "User who last updated the pipeline.",
				Optional:            true,
				Computed:            true,
			},
			// Pipeline settings
			"pipeline_batch_delay": schema.Int64Attribute{
				MarkdownDescription: "Time in milliseconds to wait for each event before sending an undersized batch to pipeline workers.",
				Optional:            true,
				Computed:            true,
			},
			"pipeline_batch_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of events an individual worker thread collects before executing filters and outputs.",
				Optional:            true,
				Computed:            true,
			},
			"pipeline_ecs_compatibility": schema.StringAttribute{
				MarkdownDescription: "Sets the pipeline default value for ecs_compatibility, " +
					"a setting that is available to plugins that implement an ECS compatibility " +
					"mode for use with the Elastic Common Schema.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "v1", "v8"),
				},
			},
			"pipeline_ordered": schema.StringAttribute{
				MarkdownDescription: "Set the pipeline event ordering.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "true", "false"),
				},
			},
			"pipeline_plugin_classloaders": schema.BoolAttribute{
				MarkdownDescription: "(Beta) Load Java plugins in independent classloaders to isolate their dependencies.",
				Optional:            true,
				Computed:            true,
			},
			"pipeline_unsafe_shutdown": schema.BoolAttribute{
				MarkdownDescription: "Forces Logstash to exit during shutdown even if there are still inflight events in memory.",
				Optional:            true,
				Computed:            true,
			},
			"pipeline_workers": schema.Int64Attribute{
				MarkdownDescription: "The number of parallel workers used to run the filter and output stages of the pipeline.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"queue_checkpoint_acks": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of ACKed events before forcing a checkpoint when persistent queues are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"queue_checkpoint_retry": schema.BoolAttribute{
				MarkdownDescription: "When enabled, Logstash will retry four times per attempted checkpoint write for any checkpoint writes that fail. Any subsequent errors are not retried.",
				Optional:            true,
				Computed:            true,
			},
			"queue_checkpoint_writes": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of written events before forcing a checkpoint when persistent queues are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"queue_drain": schema.BoolAttribute{
				MarkdownDescription: "When enabled, Logstash waits until the persistent queue is drained before shutting down.",
				Optional:            true,
				Computed:            true,
			},
			"queue_max_bytes": schema.StringAttribute{
				MarkdownDescription: "Units for the total capacity of the queue when persistent queues are enabled.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(queueMaxBytesRegexp, "must be valid size unit"),
				},
			},
			"queue_max_events": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of unread events in the queue when persistent queues are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"queue_page_capacity": schema.StringAttribute{
				MarkdownDescription: "The size of the page data files used when persistent queues are enabled. The queue data consists of append-only data files separated into pages.",
				Optional:            true,
				Computed:            true,
			},
			"queue_type": schema.StringAttribute{
				MarkdownDescription: "The internal queueing model for event buffering. Options are memory for in-memory queueing, or persisted for disk-based acknowledged queueing.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("memory", "persisted"),
				},
			},
		},
	}
}
