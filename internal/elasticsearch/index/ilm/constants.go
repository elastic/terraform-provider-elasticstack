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

// ILM phase action names. Referenced from the schema, attr types, expand, and
// flatten layers so the strings stay in sync across the whole package.
const (
	ilmActionAllocate           = "allocate"
	ilmActionDownsample         = "downsample"
	ilmActionForcemerge         = "forcemerge"
	ilmActionFreeze             = "freeze"
	ilmActionMigrate            = "migrate"
	ilmActionReadonly           = "readonly"
	ilmActionRollover           = "rollover"
	ilmActionSearchableSnapshot = "searchable_snapshot"
	ilmActionSetPriority        = "set_priority"
	ilmActionShrink             = "shrink"
	ilmActionUnfollow           = "unfollow"
	ilmActionWaitForSnapshot    = "wait_for_snapshot"
)

// Terraform schema attribute keys for ILM phase blocks and nested action
// settings. Reused across schema, attr-types helpers, expand, and flatten.
const (
	attrAllowWriteAfterShrink    = "allow_write_after_shrink"
	attrDeleteSearchableSnapshot = "delete_searchable_snapshot"
	attrEnabled                  = "enabled"
	attrExclude                  = "exclude"
	attrFixedInterval            = "fixed_interval"
	attrForceMergeIndex          = "force_merge_index"
	attrInclude                  = "include"
	attrMaxAge                   = "max_age"
	attrMaxPrimaryShardDocs      = "max_primary_shard_docs"
	attrMaxPrimaryShardSize      = "max_primary_shard_size"
	attrMetadata                 = "metadata"
	attrMinAge                   = "min_age"
	attrMinDocs                  = "min_docs"
	attrMinPrimaryShardDocs      = "min_primary_shard_docs"
	attrMinPrimaryShardSize      = "min_primary_shard_size"
	attrMinSize                  = "min_size"
	attrNumberOfReplicas         = "number_of_replicas"
	attrPriority                 = "priority"
	attrRequire                  = "require"
	attrSnapshotRepository       = "snapshot_repository"
	attrTotalShardsPerNode       = "total_shards_per_node"
	attrWaitTimeout              = "wait_timeout"
)
