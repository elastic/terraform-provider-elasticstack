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
	descID                            = "Internal identifier of the resource in the format `<cluster_uuid>/<name>`."
	descName                          = "Name of the follower index."
	descRemoteCluster                 = "Remote cluster alias containing the leader index."
	descLeaderIndex                   = "Name of the leader index on the remote cluster."
	descDataStreamName                = "Local data stream name when following a data stream leader. Write-only; not returned by the CCR info API."
	descSettingsRaw                   = "JSON-encoded index settings to override from the leader index. Write-only; not returned by the CCR info API."
	descMaxOutstandingReadRequests    = "Maximum number of outstanding read requests from the remote cluster."
	descMaxOutstandingWriteRequests   = "Maximum number of outstanding write requests on the follower."
	descMaxReadRequestOperationCount  = "Maximum number of operations to pull per read from the remote cluster."
	descMaxReadRequestSize            = "Maximum size in bytes per read batch from the remote cluster (e.g. `\"100mb\"`)."
	descMaxRetryDelay                 = "Maximum time to wait before retrying a failed operation (e.g. `\"10s\"`)."
	descMaxWriteBufferCount           = "Maximum number of operations queued for writing."
	descMaxWriteBufferSize            = "Maximum total bytes of operations queued for writing (e.g. `\"100mb\"`)."
	descMaxWriteRequestOperationCount = "Maximum number of operations per bulk write request on the follower."
	descMaxWriteRequestSize           = "Maximum total bytes per bulk write request on the follower (e.g. `\"100mb\"`)."
	descReadPollTimeout               = "Maximum time to wait for new operations on the remote cluster when synchronized (e.g. `\"10m\"`)."
	descDeleteIndexOnDestroy          = "When true, destroy deletes the underlying index after unfollowing. When false (default), the index is opened as a regular index."
	descStatus                        = "Desired replication status: `active` or `paused`."
)
