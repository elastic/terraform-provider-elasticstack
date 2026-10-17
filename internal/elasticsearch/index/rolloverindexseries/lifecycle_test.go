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
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCreate_computedIDStaysClusterAndAliasScoped(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	require.Equal(t, testClusterUUID+"/orders-write", state.ID.ValueString())
	require.Equal(t, "orders-000001", state.WriteIndex.ValueString())

	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.addMember("orders-write", "orders-000001", new(false), false)
	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.True(t, found)
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000001", "orders-000002")
}

func TestCreate_bootstrapsMissingAliasWithAliasOnlyIndex(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)

	require.Equal(t, "orders-000001", state.WriteIndex.ValueString())
	requireIndices(t, state.Indices, "orders-000001")
	require.Len(t, mock.putBodies, 1)
	body := mock.putBodies[0]
	require.NotContains(t, body, "settings")
	require.NotContains(t, body, "mappings")
	var aliases map[string]struct {
		IsWriteIndex bool `json:"is_write_index"`
	}
	require.NoError(t, json.Unmarshal(body["aliases"], &aliases))
	require.Equal(t, map[string]struct {
		IsWriteIndex bool `json:"is_write_index"`
	}{
		"orders-write": {IsWriteIndex: true},
	}, aliases)
	require.NotContains(t, strings.Join(mock.requestPaths(), "\n"), "*")
}

func TestCreate_bootstrapCollisionDoesNotMutate(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addIndex("orders-000001")

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, text, "orders-000001")
	require.Contains(t, strings.ToLower(text), "already")
	require.Zero(t, mock.aliasMutations)
	require.Empty(t, mock.membership["orders-write"])
	requireNoSurvivorProbe(t, mock, "orders-000002")
}

func TestCreate_unrelatedAlreadyExistsIsNotABootstrapCollision(t *testing.T) {
	var unrelated diag.Diagnostics
	unrelated.AddError("Unable to create index", "alias [orders-write] already exists")
	require.False(t, bootstrapNameCollision(unrelated))

	var collision diag.Diagnostics
	collision.AddError("Unable to create index", "type: resource_already_exists_exception, reason: index [orders-000001] already exists")
	require.True(t, bootstrapNameCollision(collision))
}

func TestCreate_aliasOccupiedBootstrapNameDoesNotMutate(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addIndex("other-index")
	mock.addAliasObject("orders-000001")
	mock.addMember("orders-000001", "other-index", nil, false)

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), "orders-000001")
	require.Zero(t, mock.aliasMutations)
	_, attached := mock.membership["orders-write"]
	require.False(t, attached)
	require.Contains(t, mock.indices, "other-index")
}

func TestCreate_detachedLaterGenerationDoesNotBlockBootstrap(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addIndex("orders-000002")

	state := applySeries(t, client)
	require.Equal(t, "orders-000001", state.WriteIndex.ValueString())
	requireIndices(t, state.Indices, "orders-000001")
	require.Contains(t, mock.indices, "orders-000002")
	require.NotContains(t, mock.membership["orders-write"], "orders-000002")
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "orders-000002")
		require.NotContains(t, request, "*")
	}
}

func TestCreate_interruptedBootstrapIsRediscovered(t *testing.T) {
	mock, client := newClusterMock(t)
	first := applySeries(t, client)
	require.Len(t, mock.putBodies, 1)

	second := applySeries(t, client)
	require.Len(t, mock.putBodies, 1)
	require.Equal(t, first.ID.ValueString(), second.ID.ValueString())
	require.Equal(t, "orders-000001", second.WriteIndex.ValueString())
	requireIndices(t, second.Indices, "orders-000001")
}

func TestCreate_adoptsHealthyNumericWriterWithoutMutation(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.addMember("orders-write", "orders-000002", new(false), false)

	state := applySeries(t, client)
	require.Empty(t, mock.putBodies)
	require.Zero(t, mock.aliasMutations)
	require.Equal(t, testClusterUUID+"/orders-write", state.ID.ValueString())
	require.Equal(t, "orders-000003", state.WriteIndex.ValueString())
	requireIndices(t, state.Indices, "orders-000002", "orders-000003")
	require.Equal(t, "orders", state.Prefix.ValueString())
}

func TestCreate_adoptionIncludesOutOfPrefixMembers(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000003", new(true), false)
	mock.addMember("orders-write", "unrelated-index-1", nil, false)

	state := applySeries(t, client)
	require.Empty(t, mock.putBodies)
	require.Equal(t, "orders-000003", state.WriteIndex.ValueString())
	requireIndices(t, state.Indices, "orders-000003", "unrelated-index-1")

	protected := state
	protected.DeletionProtection = types.BoolValue(true)
	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", protected)
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), "deletion_protection")
	require.Contains(t, mock.indices, "unrelated-index-1")

	state.DeletionProtection = types.BoolValue(false)
	diags = deleteRolloverIndexSeries(t.Context(), client, "orders-write", state)
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.NotContains(t, mock.indices, "unrelated-index-1")
	require.NotContains(t, mock.indices, "orders-000003")
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "DELETE /orders-write")
		require.NotContains(t, request, "*")
	}
}

func TestCreate_missingWriterBlocksAdoption(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000001", nil, false)
	mock.addMember("orders-write", "orders-000002", nil, false)

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, strings.ToLower(text), "effective writer")
	require.Contains(t, text, "orders-000001")
	require.Contains(t, text, "orders-000002")
	require.Empty(t, mock.putBodies)
	require.Zero(t, mock.aliasMutations)
}

func TestCreate_nonNumericWriterBlocksAdoption(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-current", new(true), false)

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, text, "orders-current")
	require.Contains(t, strings.ToLower(text), "successor")
	require.Empty(t, mock.putBodies)
	require.Zero(t, mock.aliasMutations)
	require.Contains(t, mock.indices, "orders-current")
}

func TestCreate_dataStreamAliasRejected(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-stream", new(true), true)

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, strings.ToLower(text), "ordinary index aliases")
	require.Contains(t, text, "orders-stream")
	require.Empty(t, mock.putBodies)
	require.Zero(t, mock.aliasMutations)
	require.Contains(t, mock.dataStreams, "orders-stream")
}

func TestCreate_numericDataStreamTargetRejected(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-000001", new(true), true)

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, strings.ToLower(text), "ordinary index aliases")
	require.Contains(t, text, "orders-000001")
	require.NotContains(t, strings.ToLower(text), "successor")
	require.Empty(t, mock.putBodies)
	require.Zero(t, mock.aliasMutations)

	_, found, readDiags := readRolloverIndexSeries(t.Context(), client, "orders-write", planModel("orders", "orders-write"))
	require.False(t, found)
	require.True(t, readDiags.HasError())
	require.Contains(t, strings.ToLower(diagnosticText(readDiags)), "ordinary index aliases")

	deleteDiags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.True(t, deleteDiags.HasError())
	require.Contains(t, strings.ToLower(diagnosticText(deleteDiags)), "ordinary index aliases")
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "DELETE ")
	}
}

func TestCreate_aliasLookupErrorIsNotAbsence(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.failAliasCode = 401

	_, diags := createSeries(t, client)
	require.True(t, diags.HasError())
	require.Empty(t, mock.putBodies)
}

func TestRead_healthyAliasRefreshesWithoutWarning(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.setWriteFlag("orders-write", "orders-000001", new(false))
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.Empty(t, diags)
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000001", "orders-000002")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, state.Prefix.ValueString(), refreshed.Prefix.ValueString())
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.Zero(t, mock.aliasMutations)
}

func TestRead_externalRolloverDoesNotRestoreWriter(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	require.Equal(t, "orders-000001", state.WriteIndex.ValueString())
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.setWriteFlag("orders-write", "orders-000001", new(false))
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.Empty(t, diags)
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000001", "orders-000002")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	write := mock.membership["orders-write"]["orders-000001"].write
	require.NotNil(t, write)
	require.False(t, *write)
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.Zero(t, mock.aliasMutations)
	require.Empty(t, mock.putBodies[1:])
}

func TestRead_deletedBootstrapIsNotRecreated(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.dropIndex("orders-000001")
	puts := len(mock.putBodies)

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.Empty(t, diags)
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000002")
	require.NotContains(t, mock.indices, "orders-000001")
	require.Len(t, mock.putBodies, puts)
	require.Zero(t, mock.aliasMutations)
}

func TestRead_removedMemberIsNotReattached(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.removeMember("orders-write", "orders-000001")

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.Empty(t, diags)
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000002")
	require.Contains(t, mock.indices, "orders-000001")
	_, attached := mock.membership["orders-write"]["orders-000001"]
	require.False(t, attached)
	require.Zero(t, mock.aliasMutations)
	require.Len(t, mock.putBodies, 1)
}

func TestRead_unhealthyAliasWarnsAndKeepsNullWriteIndex(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(false), false)
	mock.setWriteFlag("orders-write", "orders-000001", new(false))
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.False(t, diags.HasError(), diagnosticText(diags))
	requireWarning(t, diags, "orders-write", "will not repair")
	require.True(t, refreshed.WriteIndex.IsNull())
	requireIndices(t, refreshed.Indices, "orders-000001", "orders-000002")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.True(t, refreshed.DeletionProtection.ValueBool())
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.Zero(t, mock.aliasMutations)
}

func TestRead_nonNumericWriterWarnsAndExposesName(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-current", new(true), false)
	mock.removeMember("orders-write", "orders-000001")
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.True(t, found)
	require.False(t, diags.HasError(), diagnosticText(diags))
	requireWarning(t, diags, "orders-current", "will not repair")
	require.Equal(t, "orders-current", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-current")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.Contains(t, mock.indices, "orders-current")
	require.Zero(t, mock.aliasMutations)
}

func TestRead_dataStreamTargetRetainsState(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.removeMember("orders-write", "orders-000001")
	mock.addMember("orders-write", "orders-stream", new(true), true)
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, strings.ToLower(diagnosticText(diags)), "ordinary index aliases")
	require.Contains(t, diagnosticText(diags), "orders-stream")
	require.Empty(t, warningText(diags))
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, state.Prefix.ValueString(), refreshed.Prefix.ValueString())
	require.Equal(t, state.WriteIndex.ValueString(), refreshed.WriteIndex.ValueString())
	require.True(t, refreshed.DeletionProtection.ValueBool())
	require.Contains(t, mock.dataStreams, "orders-stream")
	for _, request := range mock.requestPaths()[before:] {
		require.NotContains(t, request, "DELETE ")
		require.NotContains(t, request, "PUT ")
		require.NotContains(t, request, "POST ")
	}
}

func TestRead_authorizationErrorIsNotAbsence(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.failAliasCode = 401

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Empty(t, warningText(diags))
	require.NotContains(t, strings.ToLower(diagnosticText(diags)), "no effective writer")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, "orders-000001", refreshed.WriteIndex.ValueString())
	require.Contains(t, mock.indices, "orders-000001")
	require.NotEmpty(t, mock.membership["orders-write"])
}

func TestRead_transportErrorIsNotAbsence(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.failTransport = true

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Empty(t, warningText(diags))
	require.NotContains(t, strings.ToLower(diagnosticText(diags)), "no effective writer")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, "orders-000001", refreshed.WriteIndex.ValueString())
	require.Contains(t, mock.indices, "orders-000001")
}

func TestRead_decodingErrorIsNotAbsence(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.aliasBody = "not-json"

	refreshed, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", state)

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Empty(t, warningText(diags))
	require.NotContains(t, strings.ToLower(diagnosticText(diags)), "no effective writer")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, "orders-000001", refreshed.WriteIndex.ValueString())
	require.Contains(t, mock.indices, "orders-000001")
}

func TestRead_timeoutErrorIsNotAbsence(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)

	refreshed, found, diags := readRolloverIndexSeries(ctx, client, "orders-write", state)

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Empty(t, warningText(diags))
	require.NotContains(t, strings.ToLower(diagnosticText(diags)), "no effective writer")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, "orders-000001", refreshed.WriteIndex.ValueString())
	require.Contains(t, mock.indices, "orders-000001")
}

func TestRead_aliasAbsenceRebootstrapsWhenNameIsFree(t *testing.T) {
	mock, client := newClusterMock(t)
	_ = applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.dropIndex("orders-000001")
	mock.removeMember("orders-write", "orders-000002")
	before := len(mock.requestPaths())

	_, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", planModel("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.False(t, found)

	state := applySeries(t, client)
	require.Equal(t, "orders-000001", state.WriteIndex.ValueString())
	requireIndices(t, state.Indices, "orders-000001")
	require.Contains(t, mock.indices, "orders-000002")
	_, attached := mock.membership["orders-write"]["orders-000002"]
	require.False(t, attached)
	for _, request := range mock.requestPaths()[before:] {
		require.NotContains(t, request, "orders-000002")
		require.NotContains(t, request, "*")
	}
}

func TestRead_aliasAbsenceBootstrapCollisionDoesNotRepair(t *testing.T) {
	mock, client := newClusterMock(t)
	_ = applySeries(t, client)
	mock.removeMember("orders-write", "orders-000001")
	mock.addIndex("orders-000002")
	before := len(mock.requestPaths())

	_, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", planModel("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.False(t, found)

	_, createDiags := createSeries(t, client)
	require.True(t, createDiags.HasError())
	text := diagnosticText(createDiags)
	require.Contains(t, text, "orders-000001")
	require.Contains(t, strings.ToLower(text), "already")
	require.Contains(t, mock.indices, "orders-000001")
	require.Contains(t, mock.indices, "orders-000002")
	_, attached := mock.membership["orders-write"]
	require.False(t, attached)
	require.Zero(t, mock.aliasMutations)
	for _, request := range mock.requestPaths()[before:] {
		require.NotContains(t, request, "orders-000002")
		require.NotContains(t, request, "*")
	}
}

func TestUpdate_disablingDeletionProtectionReadsBackWithoutMutation(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.addMember("orders-write", "orders-000002", new(true), false)
	mock.setWriteFlag("orders-write", "orders-000001", new(false))
	before := len(mock.requestPaths())

	plan := state
	plan.DeletionProtection = types.BoolValue(false)
	refreshed := applyUpdate(t, client, plan)

	require.False(t, refreshed.DeletionProtection.ValueBool())
	require.Equal(t, "orders-000002", refreshed.WriteIndex.ValueString())
	requireIndices(t, refreshed.Indices, "orders-000001", "orders-000002")
	require.Equal(t, state.ID.ValueString(), refreshed.ID.ValueString())
	require.Equal(t, state.Prefix.ValueString(), refreshed.Prefix.ValueString())
	require.Equal(t, state.RolloverAlias.ValueString(), refreshed.RolloverAlias.ValueString())
	write := mock.membership["orders-write"]["orders-000001"].write
	require.NotNil(t, write)
	require.False(t, *write)
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.Zero(t, mock.aliasMutations)
	require.Len(t, mock.putBodies, 1)
	requireAliasDiscovery(t, mock.requestPaths()[before:])
}

func TestUpdate_writerWarningPreservesTimeouts(t *testing.T) {
	mock, client := newClusterMock(t)
	state := applySeries(t, client)
	mock.setWriteFlag("orders-write", "orders-000001", new(false))
	before := len(mock.requestPaths())

	plan := state
	plan.DeletionProtection = types.BoolValue(false)
	plan.Timeouts = operationTimeouts("4m")
	written, diags := updateRolloverIndexSeries(t.Context(), client, entitycore.WriteRequest[Model]{
		Plan:    plan,
		WriteID: "orders-write",
	})
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.Empty(t, warningText(diags))

	refreshed, found, readDiags := readRolloverIndexSeries(t.Context(), client, "orders-write", written.Model)
	require.True(t, found)
	require.False(t, readDiags.HasError(), diagnosticText(readDiags))
	requireWarning(t, readDiags, "will not repair")
	require.True(t, refreshed.WriteIndex.IsNull())
	require.False(t, refreshed.DeletionProtection.ValueBool())
	require.True(t, refreshed.Timeouts.Equal(plan.Timeouts))
	requireIndices(t, refreshed.Indices, "orders-000001")
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
	require.NotContains(t, strings.Join(mock.requestPaths()[before:], "\n"), "_validate/query")
	require.Zero(t, mock.aliasMutations)
	require.Len(t, mock.putBodies, 1)
}

func TestResourceDescription_documentsAliasExistenceAndReplacementOrdering(t *testing.T) {
	for _, needle := range []string{
		"exact `rollover_alias`",
		"removes this resource from state",
		"`<prefix>-000001`",
		"does not look up former members",
		"detached",
		"destroy before create",
		"Terraform Core",
		"`create_before_destroy`",
		"propagated through dependencies",
		"does not enforce",
	} {
		require.Contains(t, resourceDescription, needle)
	}
	lower := strings.ToLower(resourceDescription)
	require.NotContains(t, lower, "the provider enforces")
	require.NotContains(t, lower, "prevents create_before_destroy")
}

func TestRead_missingAliasIsAbsenceWithoutSurvivorProbe(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addIndex("orders-000002")

	_, found, diags := readRolloverIndexSeries(t.Context(), client, "orders-write", planModel("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.False(t, found)
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "orders-000002")
		require.NotContains(t, request, "*")
	}
}

func TestDelete_missingAliasSucceedsWithoutDeletingDetachedIndex(t *testing.T) {
	mock, client := newClusterMock(t)
	mock.addIndex("orders-000002")

	diags := deleteRolloverIndexSeries(t.Context(), client, "orders-write", unprotectedPlan("orders", "orders-write"))
	require.False(t, diags.HasError(), diagnosticText(diags))
	require.Contains(t, mock.indices, "orders-000002")
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, "DELETE ")
	}
}

func applyUpdate(t *testing.T, client *clients.ElasticsearchScopedClient, plan Model) Model {
	t.Helper()
	written, diags := updateRolloverIndexSeries(t.Context(), client, entitycore.WriteRequest[Model]{
		Plan:    plan,
		WriteID: plan.RolloverAlias.ValueString(),
	})
	require.False(t, diags.HasError(), diagnosticText(diags))
	state, found, readDiags := readRolloverIndexSeries(t.Context(), client, plan.RolloverAlias.ValueString(), written.Model)
	require.False(t, readDiags.HasError(), diagnosticText(readDiags))
	require.Empty(t, warningText(readDiags))
	require.True(t, found)
	return state
}

func applySeries(t *testing.T, client *clients.ElasticsearchScopedClient) Model {
	t.Helper()
	written, diags := createSeries(t, client)
	require.False(t, diags.HasError(), diagnosticText(diags))
	state, found, readDiags := readRolloverIndexSeries(t.Context(), client, "orders-write", written.Model)
	require.False(t, readDiags.HasError(), diagnosticText(readDiags))
	require.True(t, found)
	return state
}

func createSeries(t *testing.T, client *clients.ElasticsearchScopedClient) (entitycore.WriteResult[Model], diag.Diagnostics) {
	t.Helper()
	return createRolloverIndexSeries(t.Context(), client, entitycore.WriteRequest[Model]{
		Plan:    planModel("orders", "orders-write"),
		WriteID: "orders-write",
	})
}

func operationTimeouts(update string) timeouts.Value {
	null := types.StringNull()
	return timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		},
		map[string]attr.Value{
			"create": null,
			"read":   null,
			"update": types.StringValue(update),
			"delete": null,
		},
	)}
}

func planModel(prefix, alias string) Model {
	return Model{
		Prefix:             types.StringValue(prefix),
		RolloverAlias:      types.StringValue(alias),
		DeletionProtection: types.BoolValue(true),
	}
}

//nolint:unparam // delete tests share one prefix; callers still pass the series name explicitly
func unprotectedPlan(prefix, alias string) Model {
	model := planModel(prefix, alias)
	model.DeletionProtection = types.BoolValue(false)
	return model
}

func requireIndices(t *testing.T, got types.Set, names ...string) {
	t.Helper()
	elements := make([]attr.Value, len(names))
	for i, name := range names {
		elements[i] = types.StringValue(name)
	}
	want, diags := types.SetValue(types.StringType, elements)
	require.False(t, diags.HasError())
	require.Truef(t, got.Equal(want), "indices = %s, want %s", got, want)
}

func warningText(diags diag.Diagnostics) string {
	var warnings []string
	for _, d := range diags {
		if d.Severity() != diag.SeverityWarning {
			continue
		}
		warnings = append(warnings, d.Summary()+" "+d.Detail())
	}
	return strings.Join(warnings, "; ")
}

func requireWarning(t *testing.T, diags diag.Diagnostics, needles ...string) {
	t.Helper()
	text := warningText(diags)
	require.NotEmpty(t, text)
	for _, needle := range needles {
		require.Contains(t, text, needle)
	}
}

func requireAliasDiscovery(t *testing.T, requests []string) {
	t.Helper()
	var gets int
	for _, request := range requests {
		if strings.HasPrefix(request, "GET /_alias/") {
			gets++
		}
	}
	require.Positive(t, gets)
}

func requireReadOnlyRequests(t *testing.T, requests []string) {
	t.Helper()
	for _, request := range requests {
		require.Truef(t, strings.HasPrefix(request, "GET "), "mutating request %s", request)
		require.NotContains(t, request, "_validate/query")
	}
}

func requireNoSurvivorProbe(t *testing.T, mock *clusterMock, detached string) {
	t.Helper()
	for _, request := range mock.requestPaths() {
		require.NotContains(t, request, detached)
		require.NotContains(t, request, "*")
	}
}
