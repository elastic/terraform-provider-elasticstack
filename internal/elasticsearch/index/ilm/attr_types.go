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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func setPriorityObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"priority": types.Int64Type,
		},
	}
}

func unfollowObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func rolloverObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"max_age":                types.StringType,
			"max_docs":               types.Int64Type,
			"max_size":               types.StringType,
			"max_primary_shard_docs": types.Int64Type,
			"max_primary_shard_size": types.StringType,
			"min_age":                types.StringType,
			"min_docs":               types.Int64Type,
			"min_size":               types.StringType,
			"min_primary_shard_docs": types.Int64Type,
			"min_primary_shard_size": types.StringType,
		},
	}
}

func readonlyObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func shrinkObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number_of_shards":         types.Int64Type,
			"max_primary_shard_size":   types.StringType,
			"allow_write_after_shrink": types.BoolType,
		},
	}
}

func forcemergeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"max_num_segments": types.Int64Type,
			"index_codec":      types.StringType,
		},
	}
}

func searchableSnapshotObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"snapshot_repository": types.StringType,
			"force_merge_index":   types.BoolType,
		},
	}
}

func downsampleObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"fixed_interval": types.StringType,
			"wait_timeout":   types.StringType,
		},
	}
}

func allocateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number_of_replicas":    types.Int64Type,
			"total_shards_per_node": types.Int64Type,
			"include":               jsontypes.NormalizedType{},
			"exclude":               jsontypes.NormalizedType{},
			"require":               jsontypes.NormalizedType{},
		},
	}
}

func migrateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func freezeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func deleteActionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"delete_searchable_snapshot": types.BoolType,
		},
	}
}

func waitForSnapshotObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"policy": types.StringType,
		},
	}
}

func hotPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"set_priority":        types.ListType{ElemType: setPriorityObjectType()},
			"unfollow":            types.ListType{ElemType: unfollowObjectType()},
			"rollover":            types.ListType{ElemType: rolloverObjectType()},
			"readonly":            types.ListType{ElemType: readonlyObjectType()},
			"shrink":              types.ListType{ElemType: shrinkObjectType()},
			"forcemerge":          types.ListType{ElemType: forcemergeObjectType()},
			"searchable_snapshot": types.ListType{ElemType: searchableSnapshotObjectType()},
			"downsample":          types.ListType{ElemType: downsampleObjectType()},
		},
	}
}

func warmPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":      types.StringType,
			"set_priority": types.ListType{ElemType: setPriorityObjectType()},
			"unfollow":     types.ListType{ElemType: unfollowObjectType()},
			"readonly":     types.ListType{ElemType: readonlyObjectType()},
			"allocate":     types.ListType{ElemType: allocateObjectType()},
			"migrate":      types.ListType{ElemType: migrateObjectType()},
			"shrink":       types.ListType{ElemType: shrinkObjectType()},
			"forcemerge":   types.ListType{ElemType: forcemergeObjectType()},
			"downsample":   types.ListType{ElemType: downsampleObjectType()},
		},
	}
}

func coldPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"set_priority":        types.ListType{ElemType: setPriorityObjectType()},
			"unfollow":            types.ListType{ElemType: unfollowObjectType()},
			"readonly":            types.ListType{ElemType: readonlyObjectType()},
			"searchable_snapshot": types.ListType{ElemType: searchableSnapshotObjectType()},
			"allocate":            types.ListType{ElemType: allocateObjectType()},
			"migrate":             types.ListType{ElemType: migrateObjectType()},
			"freeze":              types.ListType{ElemType: freezeObjectType()},
			"downsample":          types.ListType{ElemType: downsampleObjectType()},
		},
	}
}

func frozenPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"searchable_snapshot": types.ListType{ElemType: searchableSnapshotObjectType()},
		},
	}
}

func deletePhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":           types.StringType,
			"wait_for_snapshot": types.ListType{ElemType: waitForSnapshotObjectType()},
			ilmPhaseDelete:      types.ListType{ElemType: deleteActionObjectType()},
		},
	}
}

func phaseObjectType(phaseName string) types.ObjectType {
	switch phaseName {
	case ilmPhaseHot:
		return hotPhaseObjectType()
	case ilmPhaseWarm:
		return warmPhaseObjectType()
	case ilmPhaseCold:
		return coldPhaseObjectType()
	case ilmPhaseFrozen:
		return frozenPhaseObjectType()
	case ilmPhaseDelete:
		return deletePhaseObjectType()
	default:
		return types.ObjectType{AttrTypes: map[string]attr.Type{}}
	}
}

func phaseListType(phaseName string) types.ListType {
	return types.ListType{ElemType: phaseObjectType(phaseName)}
}
