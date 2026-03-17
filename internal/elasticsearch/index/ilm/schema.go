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

package ilm

import (
	"context"
	_ "embed"
	"maps"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed resource-description.md
var resourceDescription string

const ilmSetPriorityActionDescription = "" +
	"Sets the priority of the index as soon as the policy enters the hot, warm, or cold phase. " +
	"Higher priority indices are recovered before indices with lower priorities following a node restart. " +
	"Default priority is 1."

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(false),
			"hot": schemaPhaseBlock(
				"The index is actively being updated and queried.",
				"set_priority", "unfollow", "rollover", "readonly", "shrink", "forcemerge", "searchable_snapshot", "downsample",
			),
			"warm": schemaPhaseBlock(
				"The index is no longer being updated but is still being queried.",
				"set_priority", "unfollow", "readonly", "allocate", "migrate", "shrink", "forcemerge", "downsample",
			),
			"cold": schemaPhaseBlock(
				"The index is no longer being updated and is queried infrequently. The information still needs to be searchable, but it is okay if those queries are slower.",
				"set_priority", "unfollow", "readonly", "searchable_snapshot", "allocate", "migrate", "freeze", "downsample",
			),
			"frozen": schemaPhaseBlock(
				"The index is no longer being updated and is queried rarely. The information still needs to be searchable, but it is okay if those queries are extremely slow.",
				"searchable_snapshot",
			),
			"delete": schemaPhaseBlock("The index is no longer needed and can safely be removed.", "wait_for_snapshot", "delete"),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Identifier for the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Optional user metadata about the ilm policy. Must be valid JSON document.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"modified_date": schema.StringAttribute{
				Description: "The DateTime of the last modification.",
				Computed:    true,
			},
		},
	}
}

func schemaPhaseBlock(description string, _ ...string) schema.ListNestedBlock {
	// Include all action blocks so phaseModel struct can unmarshal (framework omits
	// blocks not in schema, causing "struct defines fields not found in object").
	// Only the actions in the actions list are documented/expected per phase.
	blocks := make(map[string]schema.Block)
	maps.Copy(blocks, supportedActionBlocks)
	return schema.ListNestedBlock{
		Description: description,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"min_age": schema.StringAttribute{
					Description: "ILM moves indices through the lifecycle according to their age. To control the timing of these transitions, you set a minimum age for each phase.",
					Optional:    true,
					Computed:    true,
				},
			},
			Blocks: blocks,
		},
	}
}

var supportedActionBlocks = map[string]schema.Block{
	"allocate": schema.ListNestedBlock{
		Description: "Updates the index settings to change which nodes are allowed to host the index shards and change the number of replicas.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"number_of_replicas": schema.Int64Attribute{
					Description: "Number of replicas to assign to the index. Default: `0`",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(0),
				},
				"total_shards_per_node": schema.Int64Attribute{
					Description: "The maximum number of shards for the index on a single Elasticsearch node. Defaults to `-1` (unlimited). Supported from Elasticsearch version **7.16**",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(-1),
				},
				"include": schema.StringAttribute{
					Description: "Assigns an index to nodes that have at least one of the specified custom attributes. Must be valid JSON document.",
					Optional:    true,
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
				},
				"exclude": schema.StringAttribute{
					Description: "Assigns an index to nodes that have none of the specified custom attributes. Must be valid JSON document.",
					Optional:    true,
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
				},
				"require": schema.StringAttribute{
					Description: "Assigns an index to nodes that have all of the specified custom attributes. Must be valid JSON document.",
					Optional:    true,
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
				},
			},
		},
	},
	"delete": schema.ListNestedBlock{
		Description: "Permanently removes the index.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"delete_searchable_snapshot": schema.BoolAttribute{
					Description: "Deletes the searchable snapshot created in a previous phase.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	},
	"forcemerge": schema.ListNestedBlock{
		Description: "Force merges the index into the specified maximum number of segments. This action makes the index read-only.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_num_segments": schema.Int64Attribute{
					Description: "Number of segments to merge to. To fully merge the index, set to 1.",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"index_codec": schema.StringAttribute{
					Description: "Codec used to compress the document store.",
					Optional:    true,
				},
			},
		},
	},
	"freeze":   schemaEnabledActionBlock("Freeze the index to minimize its memory footprint."),
	"migrate":  schemaEnabledActionBlock(`Moves the index to the data tier that corresponds to the current phase by updating the "index.routing.allocation.include._tier_preference" index setting.`),
	"readonly": schemaEnabledActionBlock("Makes the index read-only."),
	"rollover": schema.ListNestedBlock{
		Description: "Rolls over a target to a new index when the existing index meets one or more of the rollover conditions.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_age": schema.StringAttribute{
					Description: "Triggers rollover after the maximum elapsed time from index creation is reached.",
					Optional:    true,
				},
				"max_docs": schema.Int64Attribute{
					Description: "Triggers rollover after the specified maximum number of documents is reached.",
					Optional:    true,
				},
				"max_size": schema.StringAttribute{
					Description: "Triggers rollover when the index reaches a certain size.",
					Optional:    true,
				},
				"max_primary_shard_docs": schema.Int64Attribute{
					Description: "Triggers rollover when the largest primary shard in the index reaches a certain number of documents. Supported from Elasticsearch version **8.2**",
					Optional:    true,
				},
				"max_primary_shard_size": schema.StringAttribute{
					Description: "Triggers rollover when the largest primary shard in the index reaches a certain size.",
					Optional:    true,
				},
				"min_age": schema.StringAttribute{
					Description: "Prevents rollover until after the minimum elapsed time from index creation is reached. Supported from Elasticsearch version **8.4**",
					Optional:    true,
				},
				"min_docs": schema.Int64Attribute{
					Description: "Prevents rollover until after the specified minimum number of documents is reached. Supported from Elasticsearch version **8.4**",
					Optional:    true,
				},
				"min_size": schema.StringAttribute{
					Description: "Prevents rollover until the index reaches a certain size.",
					Optional:    true,
				},
				"min_primary_shard_docs": schema.Int64Attribute{
					Description: "Prevents rollover until the largest primary shard in the index reaches a certain number of documents. Supported from Elasticsearch version **8.4**",
					Optional:    true,
				},
				"min_primary_shard_size": schema.StringAttribute{
					Description: "Prevents rollover until the largest primary shard in the index reaches a certain size. Supported from Elasticsearch version **8.4**",
					Optional:    true,
				},
			},
		},
	},
	"searchable_snapshot": schema.ListNestedBlock{
		Description: "Takes a snapshot of the managed index in the configured repository and mounts it as a searchable snapshot.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"snapshot_repository": schema.StringAttribute{
					Description: "Repository used to store the snapshot.",
					Required:    true,
				},
				"force_merge_index": schema.BoolAttribute{
					Description: "Force merges the managed index to one segment.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	},
	"set_priority": schema.ListNestedBlock{
		Description: ilmSetPriorityActionDescription,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"priority": schema.Int64Attribute{
					Description: "The priority for the index. Must be 0 or greater.",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
				},
			},
		},
	},
	"shrink": schema.ListNestedBlock{
		Description: "Sets a source index to read-only and shrinks it into a new index with fewer primary shards.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"number_of_shards": schema.Int64Attribute{
					Description: "Number of shards to shrink to.",
					Optional:    true,
				},
				"max_primary_shard_size": schema.StringAttribute{
					Description: "The max primary shard size for the target index.",
					Optional:    true,
				},
				"allow_write_after_shrink": schema.BoolAttribute{
					Description: "If true, the shrunken index is made writable by removing the write block.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
				},
			},
		},
	},
	"unfollow": schemaEnabledActionBlock("Convert a follower index to a regular index. Performed automatically before a rollover, shrink, or searchable snapshot action."),
	"wait_for_snapshot": schema.ListNestedBlock{
		Description: "Waits for the specified SLM policy to be executed before removing the index. This ensures that a snapshot of the deleted index is available.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"policy": schema.StringAttribute{
					Description: "Name of the SLM policy that the delete action should wait for.",
					Required:    true,
				},
			},
		},
	},
	"downsample": schema.ListNestedBlock{
		Description: "Roll up documents within a fixed interval to a single summary document. Reduces the index footprint by storing time series data at reduced granularity.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"fixed_interval": schema.StringAttribute{
					Description: "Downsampling interval",
					Required:    true,
				},
				"wait_timeout": schema.StringAttribute{
					Description: "Downsampling interval",
					Optional:    true,
					Computed:    true,
				},
			},
		},
	},
}

func schemaEnabledActionBlock(description string) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: description,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Description: "Controls whether ILM executes this action during the phase.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	}
}
