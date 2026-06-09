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

package ccr

// Shared tuning-parameter descriptions used by both CCR resources.
const (
	DescMaxOutstandingReadRequests    = "Maximum number of outstanding read requests from the remote cluster."
	DescMaxOutstandingWriteRequests   = "Maximum number of outstanding write requests on the follower."
	DescMaxReadRequestOperationCount  = "Maximum number of operations to pull per read from the remote cluster."
	DescMaxReadRequestSize            = "Maximum size in bytes per read batch from the remote cluster (e.g. `\"100mb\"`)."
	DescMaxRetryDelay                 = "Maximum time to wait before retrying a failed operation (e.g. `\"10s\"`)."
	DescMaxWriteBufferCount           = "Maximum number of operations queued for writing."
	DescMaxWriteBufferSize            = "Maximum total bytes of operations queued for writing (e.g. `\"100mb\"`)."
	DescMaxWriteRequestOperationCount = "Maximum number of operations per bulk write request on the follower."
	DescMaxWriteRequestSize           = "Maximum total bytes per bulk write request on the follower (e.g. `\"100mb\"`)."
	DescReadPollTimeout               = "Maximum time to wait for new operations on the remote cluster when synchronized (e.g. `\"10m\"`)."
)
