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

package followerindex

const resourceDescription = "Manages a Cross-Cluster Replication (CCR) follower index. CCR requires a Platinum or Enterprise license."

const (
	descID                   = "Internal identifier of the resource in the format `<cluster_uuid>/<name>`."
	descName                 = "Name of the follower index."
	descRemoteCluster        = "Remote cluster alias containing the leader index."
	descLeaderIndex          = "Name of the leader index on the remote cluster."
	descDataStreamName       = "Local data stream name when following a data stream leader. Requires Elasticsearch 8.4.0 or later. Write-only; not returned by the CCR info API."
	descSettingsRaw          = "JSON-encoded index settings to override from the leader index. Write-only; not returned by the CCR info API."
	descDeleteIndexOnDestroy = "When true, destroy deletes the underlying index after unfollowing. When false (default), the index is opened as a regular index."
	descStatus               = "Desired replication status: `active` or `paused`."
)
