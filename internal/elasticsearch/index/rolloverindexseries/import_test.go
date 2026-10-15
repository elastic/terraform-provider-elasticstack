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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestImportState_PopulatesCanonicalIdentityAndProtection(t *testing.T) {
	t.Parallel()

	res := importable(t)

	ctx := context.Background()
	const rawID = "abc123/orders-write/orders"
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(ctx, resource.ImportStateRequest{ID: rawID}, &imported)

	require.False(t, imported.Diagnostics.HasError(), diagnosticText(imported.Diagnostics))

	var model Model
	require.False(t, imported.State.Get(ctx, &model).HasError())
	require.Equal(t, "orders-write", model.RolloverAlias.ValueString())
	require.Equal(t, "orders", model.Prefix.ValueString())
	require.Equal(t, "abc123/orders-write", model.ID.ValueString())
	require.NotEqual(t, rawID, model.ID.ValueString())
	require.True(t, model.DeletionProtection.ValueBool())
}

func TestImportState_RejectsMissingSegment(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123/orders-write"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "missing")
}

func TestImportState_RejectsIdentifierWithoutSeparator(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
}

func TestImportState_RejectsEmptyPrefix(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123/orders-write/"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "empty")
}

func TestImportState_RejectsEmptyAlias(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123//orders"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "empty")
}

func TestImportState_RejectsEmptyClusterUUID(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "/orders-write/orders"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "empty")
}

func TestImportState_RejectsExtraSegment(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123/orders-write/orders/extra"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "extra")
}

func TestImportState_RejectsInvalidNaming(t *testing.T) {
	t.Parallel()

	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	res.ImportState(context.Background(), resource.ImportStateRequest{ID: "abc123/orders-write/orders-*"}, &imported)

	require.True(t, imported.Diagnostics.HasError())
	text := diagnosticText(imported.Diagnostics)
	require.Contains(t, text, "<cluster_uuid>/<rollover_alias>/<prefix>")
	require.Contains(t, strings.ToLower(text), "naming")
	require.Contains(t, strings.ToLower(text), "wildcard")
}

func TestImportState_ReadbackPreservesImportedPrefixAndProtection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := importedModel(t)
	mock, client := newClusterMock(t)
	const writer = "legacy-000007"
	mock.addMember("orders-write", writer, new(true), false)
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(ctx, client, state.RolloverAlias.ValueString(), state)

	require.False(t, diags.HasError(), diagnosticText(diags))
	require.True(t, found)
	require.Equal(t, "orders", refreshed.Prefix.ValueString())
	require.Equal(t, "orders-write", refreshed.RolloverAlias.ValueString())
	require.Equal(t, testClusterUUID+"/orders-write", refreshed.ID.ValueString())
	require.NotContains(t, refreshed.ID.ValueString(), writer)
	require.True(t, refreshed.DeletionProtection.ValueBool())
	require.Equal(t, writer, refreshed.WriteIndex.ValueString())
	requireReadOnlyRequests(t, mock.requestPaths()[before:])
}

func TestImportState_ReadbackPreservesAppliedProtection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := importedModel(t)
	state.DeletionProtection = types.BoolValue(false)
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "legacy-000007", new(true), false)

	refreshed, found, diags := readRolloverIndexSeries(ctx, client, state.RolloverAlias.ValueString(), state)

	require.False(t, diags.HasError(), diagnosticText(diags))
	require.True(t, found)
	require.False(t, refreshed.DeletionProtection.ValueBool())
	require.Equal(t, "orders", refreshed.Prefix.ValueString())
	require.Equal(t, testClusterUUID+"/orders-write", refreshed.ID.ValueString())
}

func TestImportState_ImmediateDeleteStaysProtected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := importedModel(t)
	mock, client := newClusterMock(t)
	const writer = "legacy-000007"
	mock.addMember("orders-write", writer, new(true), false)

	diags := deleteRolloverIndexSeries(ctx, client, state.RolloverAlias.ValueString(), state)

	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), "deletion_protection=false")
	require.Contains(t, mock.indices, writer)
	require.Empty(t, deleteRequests(mock))
}

func TestImportState_DataStreamHydrationRejectsWithoutMutation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := importedModel(t)
	mock, client := newClusterMock(t)
	mock.addMember("orders-write", "orders-stream", new(true), true)
	before := len(mock.requestPaths())

	refreshed, found, diags := readRolloverIndexSeries(ctx, client, state.RolloverAlias.ValueString(), state)

	require.False(t, found)
	require.True(t, diags.HasError())
	text := diagnosticText(diags)
	require.Contains(t, strings.ToLower(text), "data stream")
	require.Contains(t, text, "orders-stream")
	require.Equal(t, "orders", refreshed.Prefix.ValueString())
	require.Equal(t, testClusterUUID+"/orders-write", refreshed.ID.ValueString())
	require.True(t, refreshed.DeletionProtection.ValueBool())
	require.Contains(t, mock.dataStreams, "orders-stream")
	for _, request := range mock.requestPaths()[before:] {
		require.NotContains(t, request, "DELETE ")
		require.NotContains(t, request, "PUT ")
		require.NotContains(t, request, "POST ")
	}
}

func TestDocs_importInstructionsUseThreePartIdentity(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	script, err := os.ReadFile(filepath.Join(root, "examples/resources/elasticstack_elasticsearch_rollover_index_series/import.sh"))
	require.NoError(t, err)
	doc, err := os.ReadFile(filepath.Join(root, "docs/resources/elasticsearch_rollover_index_series.md"))
	require.NoError(t, err)
	example, err := os.ReadFile(filepath.Join(root, "examples/resources/elasticstack_elasticsearch_rollover_index_series/resource.tf"))
	require.NoError(t, err)

	command := "terraform import elasticstack_elasticsearch_rollover_index_series.orders '<cluster_uuid>/orders-write/orders'"
	for _, body := range []string{string(script), string(doc)} {
		require.Contains(t, body, command)
		require.Contains(t, body, "<cluster_uuid>/<rollover_alias>")
		require.Contains(t, strings.ToLower(body), "deletion_protection")
		require.Contains(t, strings.ToLower(body), "true")
	}
	require.NotContains(t, string(example), "adopt =")
	require.NotContains(t, strings.ToLower(string(doc)), "adoption flag")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}

func importedModel(t *testing.T) Model {
	t.Helper()
	res := importable(t)
	imported := resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, res)}
	ctx := context.Background()
	res.ImportState(ctx, resource.ImportStateRequest{ID: testClusterUUID + "/orders-write/orders"}, &imported)
	require.False(t, imported.Diagnostics.HasError(), diagnosticText(imported.Diagnostics))
	var model Model
	require.False(t, imported.State.Get(ctx, &model).HasError())
	return model
}

func importable(t *testing.T) resource.ResourceWithImportState {
	t.Helper()
	res, ok := NewRolloverIndexSeriesResource().(resource.ResourceWithImportState)
	require.True(t, ok, "the resource must implement ResourceWithImportState")
	return res
}
