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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func validateRequiredComments(t *testing.T, list types.List) *validator.ListResponse {
	t.Helper()
	resp := &validator.ListResponse{}
	for _, v := range requiredCommentValidators() {
		v.ValidateList(context.Background(), validator.ListRequest{
			Path:        path.Root("required_comments"),
			ConfigValue: list,
		}, resp)
	}
	return resp
}

func TestRequiredCommentsValidator_RejectsNullElement(t *testing.T) {
	t.Parallel()

	list := types.ListValueMust(types.StringType, []attr.Value{types.StringNull()})

	resp := validateRequiredComments(t, list)
	require.NotEmpty(t, resp.Diagnostics.Errors(), "a null element must be rejected: %s", resp.Diagnostics)
}

func TestRequiredCommentsValidator_RejectsWhitespaceOnlyText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		texts  []string
		errors int
	}{
		{"empty string", []string{"alpha", ""}, 1},
		{"only spaces", []string{"   "}, 1},
		{"only tabs", []string{"\t"}, 1},
		{"only newlines", []string{"\n"}, 1},
		{"leading and trailing space kept", []string{"  padded  "}, 0},
		{"valid", []string{"alpha", "beta"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			list, diags := types.ListValueFrom(context.Background(), types.StringType, tc.texts)
			require.False(t, diags.HasError())

			resp := validateRequiredComments(t, list)
			require.Len(t, resp.Diagnostics.Errors(), tc.errors, "%s", resp.Diagnostics)
		})
	}
}

func TestRequiredCommentsValidator_RejectsDuplicateText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		texts  []string
		errors int
	}{
		{"exact duplicate", []string{"alpha", "beta", "alpha"}, 1},
		{"case sensitive so not a duplicate", []string{"Alpha", "alpha"}, 0},
		{"surrounding whitespace so not a duplicate", []string{"alpha", " alpha "}, 0},
		{"multiple duplicates", []string{"a", "a", "b", "b"}, 2},
		{"unique", []string{"alpha", "beta", "gamma"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			list, diags := types.ListValueFrom(context.Background(), types.StringType, tc.texts)
			require.False(t, diags.HasError())

			resp := validateRequiredComments(t, list)
			require.Len(t, resp.Diagnostics.Errors(), tc.errors, "%s", resp.Diagnostics)
		})
	}
}

func TestRequiredCommentsValidator_SkipsUnknownElements(t *testing.T) {
	t.Parallel()

	// Unknown elements are resolved later (the comments plan modifier keeps
	// them unknown while required_comments is unknown); they must not be
	// compared as duplicate empty strings.
	list := types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown(), types.StringUnknown()})

	resp := validateRequiredComments(t, list)
	require.Empty(t, resp.Diagnostics.Errors(), "unknown elements must not be reported as duplicates: %s", resp.Diagnostics)
}

func TestMissingCommentTexts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		required []string
		stored   []string
		missing  []string
	}{
		{"all required texts present", []string{"alpha", "beta"}, []string{"alpha", "beta"}, nil},
		{"analyst-written text satisfies the requirement", []string{"alpha"}, []string{"alpha"}, nil},
		{"missing texts returned in configuration order", []string{"alpha", "gamma", "delta"}, []string{"alpha", "beta"}, []string{"gamma", "delta"}},
		{"missing nothing because text removed from required", []string{"alpha"}, []string{"alpha", "beta"}, nil},
		{"exact case-sensitive match", []string{"Alpha"}, []string{"alpha"}, []string{"Alpha"}},
		{"no trimming before comparison", []string{"alpha"}, []string{" alpha"}, []string{"alpha"}},
		{"empty stored log", []string{"alpha"}, nil, []string{"alpha"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.missing, missingCommentTexts(tc.required, tc.stored))
		})
	}
}

func TestBuildAppendComments(t *testing.T) {
	t.Parallel()

	stored := []kbapi.SecurityExceptionsAPIExceptionListItemComment{
		{Id: "id-alpha", Comment: "alpha"},
		{Id: "id-beta", Comment: "beta"},
	}

	t.Run("server before 8.6 echoes stored comments with ids in stored order before appending", func(t *testing.T) {
		t.Parallel()
		body := buildAppendComments(stored, []string{"gamma"}, false)
		require.Equal(t, kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray{
			{Id: new(kbapi.SecurityExceptionsAPINonEmptyString("id-alpha")), Comment: "alpha"},
			{Id: new(kbapi.SecurityExceptionsAPINonEmptyString("id-beta")), Comment: "beta"},
			{Comment: "gamma"},
		}, body)
	})

	t.Run("server 8.6 or later appends only the missing texts", func(t *testing.T) {
		t.Parallel()
		body := buildAppendComments(stored, []string{"gamma", "delta"}, true)
		require.Equal(t, kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray{
			{Comment: "gamma"},
			{Comment: "delta"},
		}, body)
	})

	t.Run("appends multiple missing texts in configuration order", func(t *testing.T) {
		t.Parallel()
		body := buildAppendComments(nil, []string{"gamma", "delta"}, true)
		require.Equal(t, kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray{
			{Comment: "gamma"},
			{Comment: "delta"},
		}, body)
	})
}

func TestSupportsMissingOnlyCommentBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		version  string
		expected bool
	}{
		{"8.5.0 does not support the missing-only body", "8.5.0", false},
		{"8.5.9 does not support the missing-only body", "8.5.9", false},
		{"8.6.0 supports the missing-only body", "8.6.0", true},
		{"8.11.2 supports the missing-only body", "8.11.2", true},
		{"9.4.0 supports the missing-only body", "9.4.0", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, supportsMissingOnlyCommentBody(version.Must(version.NewVersion(tc.version))))
		})
	}
}

func TestCommentReconciler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	storedComments := func(texts ...string) (func(context.Context) ([]kbapi.SecurityExceptionsAPIExceptionListItemComment, bool, diag.Diagnostics), func(context.Context) (bool, diag.Diagnostics)) {
		stored := make([]kbapi.SecurityExceptionsAPIExceptionListItemComment, len(texts))
		for i, text := range texts {
			stored[i] = kbapi.SecurityExceptionsAPIExceptionListItemComment{Id: "id-" + text, Comment: text, CreatedBy: "analyst"}
		}
		read := func(context.Context) ([]kbapi.SecurityExceptionsAPIExceptionListItemComment, bool, diag.Diagnostics) {
			return stored, true, nil
		}
		return read, func(context.Context) (bool, diag.Diagnostics) { return true, nil }
	}

	t.Run("failed fresh read aborts the update", func(t *testing.T) {
		t.Parallel()
		read := func(context.Context) ([]kbapi.SecurityExceptionsAPIExceptionListItemComment, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("Failed to read exception item", "boom")
			return nil, false, diags
		}
		r := commentReconciler{readStoredComments: read}
		body, diags := r.reconcile(ctx, []string{"alpha"})
		require.Nil(t, body)
		require.True(t, diags.HasError())
	})

	t.Run("item not found aborts the update", func(t *testing.T) {
		t.Parallel()
		read := func(context.Context) ([]kbapi.SecurityExceptionsAPIExceptionListItemComment, bool, diag.Diagnostics) {
			return nil, false, nil
		}
		r := commentReconciler{readStoredComments: read}
		body, diags := r.reconcile(ctx, []string{"alpha"})
		require.Nil(t, body)
		require.True(t, diags.HasError())
	})

	t.Run("nothing missing omits comments even when other attributes changed", func(t *testing.T) {
		t.Parallel()
		read, version := storedComments("alpha", "beta")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: version}
		body, diags := r.reconcile(ctx, []string{"alpha"})
		require.False(t, diags.HasError(), "%s", diags)
		require.Nil(t, body)
	})

	t.Run("removed required text stays stored and sends nothing", func(t *testing.T) {
		t.Parallel()
		read, version := storedComments("alpha", "beta")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: version}
		body, diags := r.reconcile(ctx, []string{"beta"})
		require.False(t, diags.HasError(), "%s", diags)
		require.Nil(t, body)
	})

	t.Run("analyst-written text satisfies the requirement", func(t *testing.T) {
		t.Parallel()
		read, version := storedComments("alpha")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: version}
		body, diags := r.reconcile(ctx, []string{"alpha"})
		require.False(t, diags.HasError(), "%s", diags)
		require.Nil(t, body)
	})

	t.Run("server 8.6 or later appends only the missing texts", func(t *testing.T) {
		t.Parallel()
		read, version := storedComments("alpha", "beta")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: version}
		body, diags := r.reconcile(ctx, []string{"alpha", "gamma", "delta"})
		require.False(t, diags.HasError(), "%s", diags)
		require.Equal(t, kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray{
			{Comment: "gamma"},
			{Comment: "delta"},
		}, *body)
	})

	t.Run("server before 8.6 echoes stored comments with ids then appends", func(t *testing.T) {
		t.Parallel()
		read, _ := storedComments("alpha", "beta")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: func(context.Context) (bool, diag.Diagnostics) {
			return false, nil
		}}
		body, diags := r.reconcile(ctx, []string{"alpha", "gamma"})
		require.False(t, diags.HasError(), "%s", diags)
		require.Equal(t, kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray{
			{Id: new(kbapi.SecurityExceptionsAPINonEmptyString("id-alpha")), Comment: "alpha"},
			{Id: new(kbapi.SecurityExceptionsAPINonEmptyString("id-beta")), Comment: "beta"},
			{Comment: "gamma"},
		}, *body)
	})

	t.Run("version discovery failure aborts the update", func(t *testing.T) {
		t.Parallel()
		read, _ := storedComments("alpha")
		r := commentReconciler{readStoredComments: read, supportsMissingOnlyBody: func(context.Context) (bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("Failed to get Kibana version", "boom")
			return false, diags
		}}
		body, diags := r.reconcile(ctx, []string{"alpha", "gamma"})
		require.Nil(t, body)
		require.True(t, diags.HasError())
	})
}
