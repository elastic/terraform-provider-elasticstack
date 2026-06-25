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
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed descriptions/ilm_set_priority_action.md
var setPriorityActionDescription string

func singleNestedBlock(desc string, nested schema.NestedBlockObject, validators ...validator.Object) schema.SingleNestedBlock {
	b := schema.SingleNestedBlock{
		MarkdownDescription: desc,
		Attributes:          nested.Attributes,
		Blocks:              nested.Blocks,
	}
	if len(validators) > 0 {
		b.Validators = validators
	}
	return b
}

func blockAllocate() schema.SingleNestedBlock {
	return singleNestedBlock("Updates the index settings to change which nodes are allowed to host the index shards and change the number of replicas.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrNumberOfReplicas: schema.Int64Attribute{
				Description: "Number of replicas to assign to the index.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			attrTotalShardsPerNode: schema.Int64Attribute{
				Description: "The maximum number of shards for the index on a single Elasticsearch node. When omitted, the existing index setting is left unchanged.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			attrInclude: schema.StringAttribute{
				Description: "Assigns an index to nodes that have at least one of the specified custom attributes. Must be valid JSON document.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{validators.StringIsJSONObject{}},
			},
			attrExclude: schema.StringAttribute{
				Description: "Assigns an index to nodes that have none of the specified custom attributes. Must be valid JSON document.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{validators.StringIsJSONObject{}},
			},
			attrRequire: schema.StringAttribute{
				Description: "Assigns an index to nodes that have all of the specified custom attributes. Must be valid JSON document.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{validators.StringIsJSONObject{}},
			},
		},
	})
}

func blockDeleteAction() schema.SingleNestedBlock {
	return singleNestedBlock("Permanently removes the index.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrDeleteSearchableSnapshot: schema.BoolAttribute{
				Description: "Deletes the searchable snapshot created in a previous phase.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	})
}

func blockForcemerge() schema.SingleNestedBlock {
	return singleNestedBlock("Force merges the index into the specified maximum number of segments. This action makes the index read-only.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"max_num_segments": schema.Int64Attribute{
				Description: "Number of segments to merge to. To fully merge the index, set to 1. Required when the `forcemerge` action is configured.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"index_codec": schema.StringAttribute{
				Description: "Codec used to compress the document store.",
				Optional:    true,
			},
		},
	}, objectvalidator.AlsoRequires(path.MatchRelative().AtName("max_num_segments")))
}

func blockFreeze() schema.SingleNestedBlock {
	return singleNestedBlock("Freeze the index to minimize its memory footprint.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				Description: "Controls whether ILM freezes the index.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	})
}

func blockMigrate() schema.SingleNestedBlock {
	return singleNestedBlock(
		`Moves the index to the data tier that corresponds to the current phase by updating `+
			`the "index.routing.allocation.include._tier_preference" index setting.`,
		schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrEnabled: schema.BoolAttribute{
					Description: "Controls whether ILM automatically migrates the index during this phase.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	)
}

func blockReadonly() schema.SingleNestedBlock {
	return singleNestedBlock("Makes the index read-only.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				Description: "Controls whether ILM makes the index read-only.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	})
}

func blockRollover() schema.SingleNestedBlock {
	return singleNestedBlock("Rolls over a target to a new index when the existing index meets one or more of the rollover conditions.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrMaxAge: schema.StringAttribute{
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
			attrMaxPrimaryShardDocs: schema.Int64Attribute{
				Description: "Triggers rollover when the largest primary shard in the index reaches a certain number of documents. Supported from Elasticsearch version **8.2**",
				Optional:    true,
			},
			attrMaxPrimaryShardSize: schema.StringAttribute{
				Description: "Triggers rollover when the largest primary shard in the index reaches a certain size.",
				Optional:    true,
			},
			attrMinAge: schema.StringAttribute{
				Description: "Prevents rollover until after the minimum elapsed time from index creation is reached. Supported from Elasticsearch version **8.4**",
				Optional:    true,
			},
			attrMinDocs: schema.Int64Attribute{
				Description: "Prevents rollover until after the specified minimum number of documents is reached. Supported from Elasticsearch version **8.4**",
				Optional:    true,
			},
			attrMinSize: schema.StringAttribute{
				Description: "Prevents rollover until the index reaches a certain size.",
				Optional:    true,
			},
			attrMinPrimaryShardDocs: schema.Int64Attribute{
				Description: "Prevents rollover until the largest primary shard in the index reaches a certain number of documents. Supported from Elasticsearch version **8.4**",
				Optional:    true,
			},
			attrMinPrimaryShardSize: schema.StringAttribute{
				Description: "Prevents rollover until the largest primary shard in the index reaches a certain size. Supported from Elasticsearch version **8.4**",
				Optional:    true,
			},
		},
	})
}

func blockSearchableSnapshot() schema.SingleNestedBlock {
	return singleNestedBlock("Takes a snapshot of the managed index in the configured repository and mounts it as a searchable snapshot.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrSnapshotRepository: schema.StringAttribute{
				Description: "Repository used to store the snapshot. Required when the `searchable_snapshot` action is configured.",
				Optional:    true,
			},
			attrForceMergeIndex: schema.BoolAttribute{
				Description: "Force merges the managed index to one segment.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}, objectvalidator.AlsoRequires(path.MatchRelative().AtName(attrSnapshotRepository)))
}

// blockSearchableSnapshotInFrozenPhase is the frozen-phase-only action; Elasticsearch requires this action for the frozen phase.
func blockSearchableSnapshotInFrozenPhase() schema.SingleNestedBlock {
	return singleNestedBlock(
		"Required in the `frozen` phase. Takes a snapshot of the managed index in the configured repository and mounts it as a searchable snapshot.",
		schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrSnapshotRepository: schema.StringAttribute{
					Description: "Repository used to store the snapshot. Required when the `searchable_snapshot` action is configured.",
					Optional:    true,
				},
				attrForceMergeIndex: schema.BoolAttribute{
					Description: "Force merges the managed index to one segment.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
		objectvalidator.AlsoRequires(path.MatchRelative().AtName(attrSnapshotRepository)),
	)
}

func blockSetPriority() schema.SingleNestedBlock {
	return singleNestedBlock(setPriorityActionDescription, schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrPriority: schema.Int64Attribute{
				Description: "The priority for the index. Must be 0 or greater. Required when the `set_priority` action is configured.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(0)},
			},
		},
	}, objectvalidator.AlsoRequires(path.MatchRelative().AtName(attrPriority)))
}

func blockShrink() schema.SingleNestedBlock {
	return singleNestedBlock("Sets a source index to read-only and shrinks it into a new index with fewer primary shards.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"number_of_shards": schema.Int64Attribute{
				Description: "Number of shards to shrink to.",
				Optional:    true,
			},
			attrMaxPrimaryShardSize: schema.StringAttribute{
				Description: "The max primary shard size for the target index.",
				Optional:    true,
			},
			attrAllowWriteAfterShrink: schema.BoolAttribute{
				Description: "If true, the shrunken index is made writable by removing the write block.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	})
}

func blockUnfollow() schema.SingleNestedBlock {
	return singleNestedBlock("Convert a follower index to a regular index. Performed automatically before a rollover, shrink, or searchable snapshot action.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				Description: "Controls whether ILM makes the follower index a regular one.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	})
}

func blockWaitForSnapshot() schema.SingleNestedBlock {
	return singleNestedBlock("Waits for the specified SLM policy to be executed before removing the index. This ensures that a snapshot of the deleted index is available.", schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"policy": schema.StringAttribute{
				Description: "Name of the SLM policy that the delete action should wait for. Required when the `wait_for_snapshot` action is configured.",
				Optional:    true,
			},
		},
	}, objectvalidator.AlsoRequires(path.MatchRelative().AtName("policy")))
}

func blockDownsample() schema.SingleNestedBlock {
	return singleNestedBlock(
		"Roll up documents within a fixed interval to a single summary document. "+
			"Reduces the index footprint by storing time series data at reduced granularity.",
		schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrFixedInterval: schema.StringAttribute{
					Description: "Downsampling interval. Required when the `downsample` action is configured.",
					Optional:    true,
				},
				attrWaitTimeout: schema.StringAttribute{
					Description: "Maximum time to wait for the downsample operation to complete before timing out.",
					Optional:    true,
					Computed:    true,
				},
			},
		},
		objectvalidator.AlsoRequires(path.MatchRelative().AtName(attrFixedInterval)),
	)
}
