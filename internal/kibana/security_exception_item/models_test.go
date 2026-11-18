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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestExceptionItemModel_fromAPI_EmptyCommentsLogIsNull(t *testing.T) {
	t.Parallel()

	m := ExceptionItemModel{SpaceID: types.StringValue("default")}
	diags := m.fromAPI(context.Background(), &kbapi.SecurityExceptionsAPIExceptionListItem{})
	require.False(t, diags.HasError(), "%s", diags)

	require.True(t, m.Comments.IsNull(), "an empty API comment log must be null in state, not an empty list")
}

func TestExceptionItemModel_fromAPI_PopulatesFullCommentLog(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 10, 12, 9, 30, 0, 0, time.UTC)
	apiResp := &kbapi.SecurityExceptionsAPIExceptionListItem{
		Comments: kbapi.SecurityExceptionsAPIExceptionListItemCommentArray{
			{Id: "comment-id-1", Comment: "alpha", CreatedBy: "analyst", CreatedAt: createdAt},
			{Id: "comment-id-2", Comment: "out-of-band", CreatedBy: "somebody-else", CreatedAt: createdAt},
		},
	}

	m := ExceptionItemModel{SpaceID: types.StringValue("default")}
	diags := m.fromAPI(context.Background(), apiResp)
	require.False(t, diags.HasError(), "%s", diags)

	var comments []CommentModel
	diags = m.Comments.ElementsAs(context.Background(), &comments, false)
	require.False(t, diags.HasError(), "%s", diags)
	require.Len(t, comments, 2)

	require.Equal(t, types.StringValue("comment-id-1"), comments[0].ID)
	require.Equal(t, types.StringValue("alpha"), comments[0].Comment)
	require.Equal(t, types.StringValue("2026-10-12T09:30:00.000Z"), comments[0].CreatedAt)
	require.Equal(t, types.StringValue("analyst"), comments[0].CreatedBy)

	require.Equal(t, types.StringValue("out-of-band"), comments[1].Comment)
	require.Equal(t, types.StringValue("somebody-else"), comments[1].CreatedBy)
}

func testExceptionItemModelWithOneEntry() ExceptionItemModel {
	return ExceptionItemModel{
		ListID:      types.StringValue("list-id"),
		Name:        types.StringValue("name"),
		Description: types.StringValue("description"),
		Type:        types.StringValue("simple"),
		Entries: types.ListValueMust(types.ObjectType{AttrTypes: getEntryAttrTypes()}, []attr.Value{
			types.ObjectValueMust(getEntryAttrTypes(), map[string]attr.Value{
				attrType:      types.StringValue("match"),
				attrField:     types.StringValue("process.name"),
				attrOperator:  types.StringNull(),
				attrValue:     types.StringValue("test-process"),
				attrValues:    types.ListNull(types.StringType),
				entryTypeList: types.ObjectNull(getListAttrTypes()),
				attrEntries:   types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()}),
			}),
		}),
	}
}

func TestExceptionItemModel_toCreateRequest_RequiredComments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("includes required texts in configuration order with no id", func(t *testing.T) {
		t.Parallel()
		m := testExceptionItemModelWithOneEntry()
		m.RequiredComments = types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("alpha"),
			types.StringValue("beta"),
		})

		req, diags := m.toCreateRequest(ctx)
		require.False(t, diags.HasError(), "%s", diags)

		raw, err := req.MarshalJSON()
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		comments, ok := body["comments"].([]any)
		require.True(t, ok, "create body must include comments: %s", raw)
		require.Len(t, comments, 2)
		require.Equal(t, map[string]any{"comment": "alpha"}, comments[0])
		require.Equal(t, map[string]any{"comment": "beta"}, comments[1])
	})

	t.Run("omits comments when required_comments is null or empty", func(t *testing.T) {
		t.Parallel()
		for _, list := range []types.List{
			types.ListNull(types.StringType),
			types.ListValueMust(types.StringType, nil),
		} {
			m := testExceptionItemModelWithOneEntry()
			m.RequiredComments = list

			req, diags := m.toCreateRequest(ctx)
			require.False(t, diags.HasError(), "%s", diags)

			raw, err := req.MarshalJSON()
			require.NoError(t, err)
			var body map[string]any
			require.NoError(t, json.Unmarshal(raw, &body))
			_, ok := body["comments"]
			require.False(t, ok, "create body must omit comments: %s", raw)
		}
	})
}

func TestExceptionItemModel_toUpdateRequest_Comments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := testExceptionItemModelWithOneEntry()
	m.RequiredComments = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("alpha")})

	marshalBody := func(t *testing.T, req *kbapi.UpdateExceptionListItemJSONRequestBody) map[string]any {
		t.Helper()
		raw, err := req.MarshalJSON()
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		return body
	}

	t.Run("omits comments when nothing is missing", func(t *testing.T) {
		t.Parallel()
		req, diags := m.toUpdateRequest(ctx, "item-id", nil)
		require.False(t, diags.HasError(), "%s", diags)
		_, ok := marshalBody(t, req)["comments"]
		require.False(t, ok, "update body must omit comments when nothing is missing")
	})

	t.Run("appends only the reconciled missing texts", func(t *testing.T) {
		t.Parallel()
		comments := buildAppendComments(
			[]kbapi.SecurityExceptionsAPIExceptionListItemComment{{Id: "id-alpha", Comment: "alpha"}},
			[]string{"gamma"}, true,
		)
		req, diags := m.toUpdateRequest(ctx, "item-id", &comments)
		require.False(t, diags.HasError(), "%s", diags)
		require.Equal(t, []any{map[string]any{"comment": "gamma"}}, marshalBody(t, req)["comments"])
	})
}

func TestExceptionItemModel_GetResourceID(t *testing.T) {
	t.Parallel()

	t.Run("parses resource id from composite state id", func(t *testing.T) {
		t.Parallel()
		m := ExceptionItemModel{ID: types.StringValue("default/auto-gen-uuid")}
		require.Equal(t, "auto-gen-uuid", m.GetResourceID().ValueString())
	})

	t.Run("falls back to item_id when id is not composite", func(t *testing.T) {
		t.Parallel()
		m := ExceptionItemModel{ID: types.StringNull(), ItemID: types.StringValue("configured-item-id")}
		require.Equal(t, "configured-item-id", m.GetResourceID().ValueString())
	})
}
