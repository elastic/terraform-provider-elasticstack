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

import (
	"context"
	"fmt"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// followerActiveTimeout bounds how long create waits for a freshly created
	// follower to begin following. PUT /_ccr/follow returns before shard follow
	// tasks start (index_following_started=false), during which GET /_ccr/info
	// reports status "paused" and omits parameters.
	followerActiveTimeout      = 2 * time.Minute
	followerActivePollInterval = 2 * time.Second
)

func createFollowerIndex(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[Model],
) (entitycore.WriteResult[Model], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	indexName := req.WriteID

	// data_stream_name is only accepted by the CCR follow API on Elasticsearch
	// 8.4.0+. Reject it early on older clusters with a clear message instead of
	// surfacing the raw "unknown field [data_stream_name]" parse error.
	diags.Append(enforceDataStreamNameSupported(ctx, client, plan)...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	// Preserve the configured/desired status. GET /_ccr/info reports a transient
	// "paused" status immediately after creation, so the pause decision must be
	// driven by the plan rather than the value read back from Elasticsearch.
	desiredStatus := plan.Status.ValueString()

	followReq, buildDiags := buildFollowRequest(plan)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	diags.Append(elasticsearch.CreateFollowerIndex(ctx, client, indexName, followReq)...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	// Wait until following has actually started. Only then does GET /_ccr/info
	// report the shard-level parameters needed to populate Computed attributes,
	// and only then are shard follow tasks present (required before pausing).
	follower, waitDiags := waitForFollowerActive(ctx, client, indexName)
	diags.Append(waitDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}
	if follower != nil {
		plan = mapFollowerIndexToModel(follower, plan)
	}
	plan.Status = types.StringValue(desiredStatus)

	if desiredStatus == statusPaused {
		diags.Append(elasticsearch.PauseFollowerIndex(ctx, client, indexName)...)
		if diags.HasError() {
			return entitycore.WriteResult[Model]{Model: plan}, diags
		}
	}

	id, idDiags := client.ID(ctx, indexName)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())

	return entitycore.WriteResult[Model]{Model: plan}, diags
}

// enforceDataStreamNameSupported rejects data_stream_name on Elasticsearch
// versions that predate its support on the CCR follow API (added in 8.4.0).
func enforceDataStreamNameSupported(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	plan Model,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(plan.DataStreamName) || plan.DataStreamName.ValueString() == "" {
		return diags
	}

	supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionDataStreamName)
	diags.Append(versionDiags...)
	if diags.HasError() {
		return diags
	}
	if !supported {
		diags.AddAttributeError(
			path.Root("data_stream_name"),
			"data_stream_name requires a newer Elasticsearch version",
			fmt.Sprintf(
				"The data_stream_name attribute is only supported on Elasticsearch %s and later. Remove data_stream_name or upgrade the cluster.",
				MinVersionDataStreamName,
			),
		)
	}
	return diags
}

// waitForFollowerActive polls GET /_ccr/info until the follower reports an active
// status with readable parameters, or the timeout elapses. It returns the most
// recent follower observed so callers can still map known fields on timeout.
func waitForFollowerActive(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
) (*estypes.FollowerIndex, diag.Diagnostics) {
	var diags diag.Diagnostics
	deadline := time.Now().Add(followerActiveTimeout)

	var last *estypes.FollowerIndex
	for {
		follower, getDiags := elasticsearch.GetFollowerIndex(ctx, client, indexName)
		diags.Append(getDiags...)
		if diags.HasError() {
			return last, diags
		}
		if follower != nil {
			last = follower
			if follower.Status.String() == statusActive && follower.Parameters != nil {
				return follower, diags
			}
		}

		if !time.Now().Before(deadline) {
			diags.AddError(
				"Timed out waiting for CCR follower to start",
				fmt.Sprintf(
					"Follower index %q did not begin following within %s. The leader index may be unreachable or ineligible for replication.",
					indexName,
					followerActiveTimeout,
				),
			)
			return last, diags
		}

		select {
		case <-ctx.Done():
			diags.AddError(
				"Context canceled while waiting for CCR follower to start",
				ctx.Err().Error(),
			)
			return last, diags
		case <-time.After(followerActivePollInterval):
		}
	}
}
