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
			"set_priority":        setPriorityObjectType(),
			"unfollow":            unfollowObjectType(),
			"rollover":            rolloverObjectType(),
			"readonly":            readonlyObjectType(),
			"shrink":              shrinkObjectType(),
			"forcemerge":          forcemergeObjectType(),
			"searchable_snapshot": searchableSnapshotObjectType(),
			"downsample":          downsampleObjectType(),
		},
	}
}

func warmPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":      types.StringType,
			"set_priority": setPriorityObjectType(),
			"unfollow":     unfollowObjectType(),
			"readonly":     readonlyObjectType(),
			"allocate":     allocateObjectType(),
			"migrate":      migrateObjectType(),
			"shrink":       shrinkObjectType(),
			"forcemerge":   forcemergeObjectType(),
			"downsample":   downsampleObjectType(),
		},
	}
}

func coldPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"set_priority":        setPriorityObjectType(),
			"unfollow":            unfollowObjectType(),
			"readonly":            readonlyObjectType(),
			"searchable_snapshot": searchableSnapshotObjectType(),
			"allocate":            allocateObjectType(),
			"migrate":             migrateObjectType(),
			"freeze":              freezeObjectType(),
			"downsample":          downsampleObjectType(),
		},
	}
}

func frozenPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"searchable_snapshot": searchableSnapshotObjectType(),
		},
	}
}

func deletePhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":           types.StringType,
			"wait_for_snapshot": waitForSnapshotObjectType(),
			ilmPhaseDelete:      deleteActionObjectType(),
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

func phaseObjectNull(phaseName string) types.Object {
	ot := phaseObjectType(phaseName)
	return types.ObjectNull(ot.AttrTypes)
}
