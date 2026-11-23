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

package securityexceptionitem

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func testResourceSchema(t *testing.T) rschema.Schema {
	t.Helper()
	ctx := context.Background()
	var resp resource.SchemaResponse
	newExceptionItemResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	return resp.Schema
}

func runMigrateV0ToV1(t *testing.T, raw map[string]any) (*resource.UpgradeStateResponse, map[string]any) {
	t.Helper()

	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateExceptionItemV0ToV1(context.Background(), req, resp)

	var got map[string]any
	if !resp.Diagnostics.HasError() {
		require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	}
	return resp, got
}

// requireUpgradedStateDecodes proves the upgraded state decodes cleanly under
// the current (v1) resource schema and its Go model, exercising the same
// consumption path terraform-plugin-framework uses after UpgradeState returns.
func requireUpgradedStateDecodes(t *testing.T, resp *resource.UpgradeStateResponse) {
	t.Helper()
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)

	ctx := context.Background()
	sch := testResourceSchema(t)
	raw, err := resp.DynamicValue.Unmarshal(sch.Type().TerraformType(ctx))
	require.NoError(t, err)

	state := tfsdk.State{Schema: sch, Raw: raw}
	var model ExceptionItemModel
	diags := state.Get(ctx, &model)
	require.False(t, diags.HasError(), "upgraded state must decode into the resource's Go model: %s", diags)
}

func TestMigrateExceptionItemV0ToV1_Comments(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"id":      "default/item-id",
		"list_id": "list-id",
		"name":    "name",
		"item_id": "item-id",
		"comments": []any{
			map[string]any{"id": "comment-id-1", "comment": "alpha"},
			map[string]any{"id": "comment-id-2", "comment": "beta"},
		},
	}

	resp, got := runMigrateV0ToV1(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	comments := got["comments"].([]any)
	require.Len(t, comments, 2)
	require.Equal(t, map[string]any{
		"id":         "comment-id-1",
		"comment":    "alpha",
		"created_at": nil,
		"created_by": nil,
	}, comments[0])
	require.Equal(t, map[string]any{
		"id":         "comment-id-2",
		"comment":    "beta",
		"created_at": nil,
		"created_by": nil,
	}, comments[1])

	require.NotContains(t, got, "required_comments", "required_comments must stay null so nothing is inferred from prior comment state")
	requireUpgradedStateDecodes(t, resp)
}

func TestMigrateExceptionItemV0ToV1_EmptyCommentLog(t *testing.T) {
	t.Parallel()

	resp, got := runMigrateV0ToV1(t, map[string]any{
		"id":      "default/item-id",
		"list_id": "list-id",
		"name":    "name",
		"item_id": "item-id",
	})
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotContains(t, got, "comments")
	require.NotContains(t, got, "required_comments")
	requireUpgradedStateDecodes(t, resp)
}
