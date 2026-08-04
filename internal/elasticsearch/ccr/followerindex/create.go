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
	"errors"
	"fmt"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
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

// followerIndexGetter fetches a follower index document for polling.
type followerIndexGetter func(ctx context.Context, indexName string) (*estypes.FollowerIndex, diag.Diagnostics)

// errFollowerGetFailed is a sentinel returned by the state checker to bail out
// of the shared poll loop while preserving the original framework diagnostics
// in a closure-captured variable.
var errFollowerGetFailed = errors.New("ccr follower index get failed")

// waitForFollowerActive polls GET /_ccr/info until the follower reports an active
// status with readable parameters, or the timeout elapses. It returns the most
// recent follower observed so callers can still map known fields on timeout.
func waitForFollowerActive(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
) (*estypes.FollowerIndex, diag.Diagnostics) {
	return waitForFollowerActiveWithInterval(ctx, indexName, followerActivePollInterval, func(ctx context.Context, indexName string) (*estypes.FollowerIndex, diag.Diagnostics) {
		return elasticsearch.GetFollowerIndex(ctx, client, indexName)
	})
}

// waitForFollowerActiveWithInterval delegates the poll loop to the shared
// [asyncutils.WaitForStateTransition] helper. The get failure is surfaced
// verbatim via a closure-captured variable, and ctx-deadline/cancellation
// errors are translated into the distinct diagnostics this function has
// always returned so existing behavior is preserved.
func waitForFollowerActiveWithInterval(
	ctx context.Context,
	indexName string,
	pollInterval time.Duration,
	get followerIndexGetter,
) (*estypes.FollowerIndex, diag.Diagnostics) {
	waitCtx, cancel := context.WithTimeout(ctx, followerActiveTimeout)
	defer cancel()

	var (
		last     *estypes.FollowerIndex
		getDiags diag.Diagnostics
	)

	stateChecker := func(checkCtx context.Context) (bool, error) {
		follower, diags := get(checkCtx, indexName)
		if diags.HasError() {
			getDiags = diags
			return false, errFollowerGetFailed
		}
		if follower != nil {
			last = follower
			if follower.Status.String() == statusActive && follower.Parameters != nil {
				return true, nil
			}
		}
		return false, nil
	}

	waitErr := asyncutils.WaitForStateTransition(waitCtx, "ccr follower index", indexName, stateChecker, asyncutils.WithPollInterval(pollInterval))

	var diags diag.Diagnostics
	switch {
	case errors.Is(waitErr, errFollowerGetFailed):
		diags.Append(getDiags...)
	case errors.Is(waitErr, context.DeadlineExceeded):
		diags.AddError(
			"Timed out waiting for CCR follower to start",
			fmt.Sprintf(
				"Follower index %q did not begin following within %s. The leader index may be unreachable or ineligible for replication.",
				indexName,
				followerActiveTimeout,
			),
		)
	case errors.Is(waitErr, context.Canceled):
		diags.AddError(
			"Context canceled while waiting for CCR follower to start",
			waitCtx.Err().Error(),
		)
	case waitErr != nil:
		diags.AddError("Failed waiting for CCR follower to start", waitErr.Error())
	}

	return last, diags
}
