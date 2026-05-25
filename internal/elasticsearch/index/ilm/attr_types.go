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
			attrPriority: types.Int64Type,
		},
	}
}

func unfollowObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrEnabled: types.BoolType,
		},
	}
}

func rolloverObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrMaxAge:              types.StringType,
			"max_docs":              types.Int64Type,
			"max_size":              types.StringType,
			attrMaxPrimaryShardDocs: types.Int64Type,
			attrMaxPrimaryShardSize: types.StringType,
			attrMinAge:              types.StringType,
			attrMinDocs:             types.Int64Type,
			attrMinSize:             types.StringType,
			attrMinPrimaryShardDocs: types.Int64Type,
			attrMinPrimaryShardSize: types.StringType,
		},
	}
}

func readonlyObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrEnabled: types.BoolType,
		},
	}
}

func shrinkObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number_of_shards":        types.Int64Type,
			attrMaxPrimaryShardSize:   types.StringType,
			attrAllowWriteAfterShrink: types.BoolType,
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
			attrSnapshotRepository: types.StringType,
			attrForceMergeIndex:    types.BoolType,
		},
	}
}

func downsampleObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrFixedInterval: types.StringType,
			attrWaitTimeout:   types.StringType,
		},
	}
}

func allocateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrNumberOfReplicas:   types.Int64Type,
			attrTotalShardsPerNode: types.Int64Type,
			attrInclude:            jsontypes.NormalizedType{},
			attrExclude:            jsontypes.NormalizedType{},
			attrRequire:            jsontypes.NormalizedType{},
		},
	}
}

func migrateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrEnabled: types.BoolType,
		},
	}
}

func freezeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrEnabled: types.BoolType,
		},
	}
}

func deleteActionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrDeleteSearchableSnapshot: types.BoolType,
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
			attrMinAge:                  types.StringType,
			ilmActionSetPriority:        setPriorityObjectType(),
			ilmActionUnfollow:           unfollowObjectType(),
			ilmActionRollover:           rolloverObjectType(),
			ilmActionReadonly:           readonlyObjectType(),
			ilmActionShrink:             shrinkObjectType(),
			ilmActionForcemerge:         forcemergeObjectType(),
			ilmActionSearchableSnapshot: searchableSnapshotObjectType(),
			ilmActionDownsample:         downsampleObjectType(),
		},
	}
}

func warmPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrMinAge:           types.StringType,
			ilmActionSetPriority: setPriorityObjectType(),
			ilmActionUnfollow:    unfollowObjectType(),
			ilmActionReadonly:    readonlyObjectType(),
			ilmActionAllocate:    allocateObjectType(),
			ilmActionMigrate:     migrateObjectType(),
			ilmActionShrink:      shrinkObjectType(),
			ilmActionForcemerge:  forcemergeObjectType(),
			ilmActionDownsample:  downsampleObjectType(),
		},
	}
}

func coldPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrMinAge:                  types.StringType,
			ilmActionSetPriority:        setPriorityObjectType(),
			ilmActionUnfollow:           unfollowObjectType(),
			ilmActionReadonly:           readonlyObjectType(),
			ilmActionSearchableSnapshot: searchableSnapshotObjectType(),
			ilmActionAllocate:           allocateObjectType(),
			ilmActionMigrate:            migrateObjectType(),
			ilmActionFreeze:             freezeObjectType(),
			ilmActionDownsample:         downsampleObjectType(),
		},
	}
}

func frozenPhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrMinAge:                  types.StringType,
			ilmActionSearchableSnapshot: searchableSnapshotObjectType(),
		},
	}
}

func deletePhaseObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrMinAge:               types.StringType,
			ilmActionWaitForSnapshot: waitForSnapshotObjectType(),
			ilmPhaseDelete:           deleteActionObjectType(),
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
