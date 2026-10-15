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

package rolloverindexseries

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDelete_rolloverAfterEnumerationDeletesNewMember(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	var introduced bool
	mock.onDelete = func(name string) {
		if introduced || name == "orders-000004" {
			return
		}
		introduced = true
		mock.indices["orders-000004"] = struct{}{}
		mock.putMember("orders-write", "orders-000004", aliasTarget{write: new(true)})
	}

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000002")
	require.NotContains(t, mock.indices, "orders-000003")
	require.NotContains(t, mock.indices, "orders-000004")
	require.ElementsMatch(t, []string{
		"DELETE /orders-000002",
		"DELETE /orders-000003",
		"DELETE /orders-000004",
	}, deleteRequests(mock))
	requireNoBroadCleanup(t, mock.requestPaths())
}

func TestDelete_memberDisappearsBeforeTypeResolutionIsRediscovered(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.onResolve = func(name string) {
		if name != "orders-000002" {
			return
		}
		delete(mock.indices, name)
		delete(mock.membership["orders-write"], name)
	}

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000003")
	require.NotContains(t, deleteRequests(mock), "DELETE /orders-000002")
	require.Contains(t, deleteRequests(mock), "DELETE /orders-000003")
}

func TestDelete_memberStillUnresolvableAfterRediscoveryFails(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.mu.Lock()
	delete(mock.indices, "orders-000002")
	mock.mu.Unlock()

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), "could not be resolved")
	require.Contains(t, mock.indices, "orders-000003")
	require.Empty(t, deleteRequests(mock))
}

func TestDelete_memberDisappearsBeforeDeletionIsTolerated(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.onDelete = func(name string) {
		if name != "orders-000002" {
			return
		}
		delete(mock.indices, name)
		delete(mock.membership["orders-write"], name)
	}

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000002")
	require.NotContains(t, mock.indices, "orders-000003")
	require.Contains(t, deleteRequests(mock), "DELETE /orders-000002")
	require.Contains(t, deleteRequests(mock), "DELETE /orders-000003")
}

func TestDelete_permissionFailureKeepsRemainingMemberForRetry(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.onDelete = func(name string) {
		if len(mock.deleteStatus) > 0 {
			return
		}
		other := "orders-000003"
		if name == other {
			other = "orders-000002"
		}
		mock.deleteStatus = map[string]int{other: http.StatusForbidden}
	}

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, diags.HasError())
	text := strings.ToLower(diagnosticText(diags))
	require.NotContains(t, text, "not found")
	require.NotContains(t, text, "absence")
	require.Contains(t, text, "unable to delete index")
	require.Len(t, mock.indices, 1)

	mock.onDelete = nil
	mock.deleteStatus = nil
	retry := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, retry.HasError(), diagnosticText(retry))
	require.Empty(t, mock.indices)
}

func TestDelete_documentsCleanupBoundariesAndPriorProtection(t *testing.T) {
	guidance := resourceDescription + "\n" + descDeletionProtection
	for _, needle := range []string{
		"quiesce producers",
		"in-flight",
		"not yet attached",
		"searchable snapshot",
		"separate cleanup",
		"detached",
		"block",
		"policy",
		"prior, separate apply",
		"same operation",
		"previously applied",
		"action.destructive_requires_name",
	} {
		require.Contains(t, strings.ToLower(guidance), strings.ToLower(needle))
	}
	lower := strings.ToLower(guidance)
	require.NotContains(t, lower, "the provider deletes snapshots")
	require.NotContains(t, lower, "stop ilm globally")
}

func TestDelete_deletesLifecycleRenamedMemberNotHistoricalAlias(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "shrink-id-orders-000001", new(false), false)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.addAliasObject("orders-000001")

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "shrink-id-orders-000001")
	require.NotContains(t, mock.indices, "orders-000002")
	require.Contains(t, mock.aliasObjects, "orders-000001")
	require.ElementsMatch(t, []string{
		"DELETE /shrink-id-orders-000001",
		"DELETE /orders-000002",
	}, deleteRequests(mock))
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "DELETE /orders-000001")
	}
}

// Mock alias transfer for searchable snapshots, source-backed by the ILM
// searchable_snapshot action: mounted indices use a restored- prefix in hot or
// cold and a partial- prefix in frozen, and aliases move to that mount.
// This is not a live searchable-snapshot execution. Live shrink evidence is
// TestAccRolloverIndexSeries_ilmShrinkRetainsAliasAndDeletesReplacement.
func TestDelete_searchableSnapshotAliasTransferRemovesMountedIndex(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "restored-id-orders-000001", new(false), false)
	mock.addMember("orders-write", "partial-id-orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.addAliasObject("orders-000001")
	mock.addAliasObject("orders-000002")

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "restored-id-orders-000001")
	require.NotContains(t, mock.indices, "partial-id-orders-000002")
	require.NotContains(t, mock.indices, "orders-000003")
	require.Contains(t, mock.aliasObjects, "orders-000001")
	require.Contains(t, mock.aliasObjects, "orders-000002")
	require.ElementsMatch(t, []string{
		"DELETE /restored-id-orders-000001",
		"DELETE /partial-id-orders-000002",
		"DELETE /orders-000003",
	}, deleteRequests(mock))
	requireNoBroadCleanup(t, mock.requestPaths())
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "DELETE /orders-000001")
		require.NotContains(t, request, "DELETE /orders-000002")
	}
}

func TestDelete_deletesOutOfPrefixMemberByConcreteName(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.addMember("orders-write", "unrelated-index-1", new(false), false)

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000003")
	require.NotContains(t, mock.indices, "unrelated-index-1")
	require.ElementsMatch(t, []string{"DELETE /orders-000003", "DELETE /unrelated-index-1"}, deleteRequests(mock))
	requireNoBroadCleanup(t, mock.requestPaths())
}

func TestDelete_leavesDetachedPrefixIndices(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.addIndex("orders-000001")
	mock.addIndex("orders-000002")

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000003")
	require.Contains(t, mock.indices, "orders-000001")
	require.Contains(t, mock.indices, "orders-000002")
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "orders-000001")
		require.NotContains(t, request, "orders-000002")
		require.NotContains(t, request, "*")
	}
}

func TestDelete_unhealthyWriterStillDeletesMembers(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.addMember("orders-write", "orders-000003", new(false), false)

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-000002")
	require.NotContains(t, mock.indices, "orders-000003")
	require.ElementsMatch(t, []string{"DELETE /orders-000002", "DELETE /orders-000003"}, deleteRequests(mock))
	requireNoBroadCleanup(t, mock.requestPaths())
}

func TestDelete_enabledProtectionDeletesNothing(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000001", new(true), false)

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", planModel("orders", "orders-write"))
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, text, "deletion_protection=false")
	require.Contains(t, strings.ToLower(text), "terraform apply")
	require.Empty(t, deleteRequests(mock))
	require.Contains(t, mock.indices, "orders-000001")
	require.Zero(t, mock.aliasMutations)
}

func TestDelete_dataStreamIntroducedAfterEnumerationIsNotDeleted(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.onDelete = func(name string) {
		if name != "orders-000003" {
			return
		}
		mock.dataStreams["orders-stream"] = struct{}{}
		mock.putMember("orders-write", "orders-stream", aliasTarget{write: new(true), dataStream: true})
	}

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, text, "orders-stream")
	require.Contains(t, strings.ToLower(text), "data stream")
	require.NotContains(t, deleteRequests(mock), "DELETE /orders-stream")
	require.Contains(t, mock.dataStreams, "orders-stream")
	require.NotContains(t, mock.indices, "orders-000003")
}

func TestDelete_aliasLookupErrorStopsBeforeDeletion(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.failAliasCode = http.StatusForbidden

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, diags.HasError())
	text := strings.ToLower(diagnosticText(diags))
	require.NotContains(t, text, "not found")
	require.Contains(t, text, "unable to get index alias")
	require.Empty(t, deleteRequests(mock))
	require.Contains(t, mock.indices, "orders-000003")
}

func TestDelete_dataStreamTargetBlocksBeforeMutation(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000001", new(true), true)
	mock.addMember("orders-write", "orders-000002", new(false), false)

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, text, "orders-000001")
	require.Contains(t, strings.ToLower(text), "data stream")
	require.Empty(t, deleteRequests(mock))
	require.Contains(t, mock.dataStreams, "orders-000001")
	require.Contains(t, mock.indices, "orders-000002")
}

func TestDelete_discoversHiddenAndClosedMembers(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-closed", new(false), false)
	mock.addMember("orders-write", ".orders-hidden", new(true), false)

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "orders-closed")
	require.NotContains(t, mock.indices, ".orders-hidden")

	var sawExpandedAlias bool
	for _, request := range mock.requestPaths() {
		if !strings.Contains(request, "GET /_alias/orders-write") {
			continue
		}
		sawExpandedAlias = true
		require.Contains(t, request, "expand_wildcards=")
		require.Contains(t, request, "open")
		require.Contains(t, request, "closed")
		require.Contains(t, request, "hidden")
	}
	require.True(t, sawExpandedAlias)
	require.ElementsMatch(t, []string{"DELETE /orders-closed", "DELETE /.orders-hidden"}, deleteRequests(mock))
}

func TestDelete_continualRolloversStopAtDeleteTimeout(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000001", new(true), false)
	next := 2
	mock.onDelete = func(string) {
		successor := fmt.Sprintf("orders-%06d", next)
		next++
		mock.indices[successor] = struct{}{}
		mock.putMember("orders-write", successor, aliasTarget{write: new(true)})
	}

	previousDelay := deleteRediscoveryDelay
	deleteRediscoveryDelay = 5 * time.Second
	t.Cleanup(func() { deleteRediscoveryDelay = previousDelay })

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	started := time.Now()
	diags := deleteRolloverIndexSeries(ctx, client, "orders-write", unprotectedPlan("orders", "orders-write"))
	elapsed := time.Since(started)

	require.True(t, diags.HasError(), diagnosticText(diags))
	require.Contains(t, strings.ToLower(diagnosticText(diags)), "timeout")
	require.NotEmpty(t, mock.indices)
	require.Less(t, elapsed, time.Second)
}

func deleteRequests(mock *clusterMock) []string {
	var deletes []string
	for _, request := range mock.requestPaths() {
		if strings.HasPrefix(request, "DELETE ") {
			deletes = append(deletes, request)
		}
	}
	return deletes
}

func requireNoBroadCleanup(t *testing.T, requests []string) {
	t.Helper()
	for _, request := range requests {
		require.NotContains(t, request, "*")
		require.NotContains(t, request, "/_ilm/")
		require.NotContains(t, request, "/_cluster/settings")
		require.NotContains(t, request, "DELETE /orders-write")
		require.NotRegexp(t, request, `DELETE /orders-\*`)
	}
}
