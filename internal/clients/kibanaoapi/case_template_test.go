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

package kibanaoapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const caseTemplateBody = `{"templateId":"tpl-1","templateVersion":2,"latestVersion":2,"isLatest":true,` +
	`"name":"t","owner":"cases","definition":{"name":"t"},"definitionString":"name: t\n","deletedAt":null}`

func newCaseTemplateTestClient(t *testing.T, status int, body string, seen *[]string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.Method+" "+r.URL.Path)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	api, err := kbapi.NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	return &Client{URL: srv.URL, HTTP: srv.Client(), API: api}
}

func TestCaseTemplateRouting(t *testing.T) {
	ctx := context.Background()
	req := kbapi.CasesTemplateWriteRequest{Owner: kbapi.CasesOwnerCases, Definition: "name: t\n"}

	t.Run("create/read/update use default-space routes", func(t *testing.T) {
		var seen []string
		c := newCaseTemplateTestClient(t, http.StatusOK, caseTemplateBody, &seen)

		created, diags := CreateCaseTemplate(ctx, c, "default", req)
		require.False(t, diags.HasError())
		assert.Equal(t, "tpl-1", created.TemplateId)

		got, diags := GetCaseTemplate(ctx, c, "default", "tpl-1")
		require.False(t, diags.HasError())
		assert.Equal(t, 2, got.TemplateVersion)

		updated, diags := UpdateCaseTemplate(ctx, c, "default", "tpl-1", req)
		require.False(t, diags.HasError())
		assert.Equal(t, "tpl-1", updated.TemplateId)

		assert.Equal(t, []string{
			"POST /api/cases/templates",
			"GET /api/cases/templates/tpl-1",
			"PUT /api/cases/templates/tpl-1",
		}, seen)
	})

	t.Run("named space routes are prefixed", func(t *testing.T) {
		var seen []string
		c := newCaseTemplateTestClient(t, http.StatusOK, caseTemplateBody, &seen)

		_, diags := CreateCaseTemplate(ctx, c, "ops", req)
		require.False(t, diags.HasError())
		_, diags = GetCaseTemplate(ctx, c, "ops", "tpl-1")
		require.False(t, diags.HasError())

		assert.Equal(t, []string{
			"POST /s/ops/api/cases/templates",
			"GET /s/ops/api/cases/templates/tpl-1",
		}, seen)
	})

	t.Run("read 404 returns nil without error", func(t *testing.T) {
		var seen []string
		c := newCaseTemplateTestClient(t, http.StatusNotFound, `{"message":"nope"}`, &seen)
		got, diags := GetCaseTemplate(ctx, c, "default", "tpl-1")
		assert.Nil(t, got)
		assert.False(t, diags.HasError())
	})

	t.Run("read 500 returns an error diagnostic", func(t *testing.T) {
		var seen []string
		c := newCaseTemplateTestClient(t, http.StatusInternalServerError, `{"message":"boom"}`, &seen)
		got, diags := GetCaseTemplate(ctx, c, "default", "tpl-1")
		assert.Nil(t, got)
		assert.True(t, diags.HasError())
	})
}

func TestDeleteCaseTemplateStatuses(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"204 is success", http.StatusNoContent, false},
		{"404 is already deleted", http.StatusNotFound, false},
		{"409 is an error", http.StatusConflict, true},
		{"500 is an error", http.StatusInternalServerError, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var seen []string
			body := ""
			if tc.status >= 400 {
				body = `{"message":"boom"}`
			}
			c := newCaseTemplateTestClient(t, tc.status, body, &seen)
			diags := DeleteCaseTemplate(context.Background(), c, "ops", "tpl-1")
			assert.Equal(t, tc.wantErr, diags.HasError())
			assert.Equal(t, []string{"DELETE /s/ops/api/cases/templates/tpl-1"}, seen)
		})
	}
}

func TestCaseTemplateMutateErrors(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusConflict} {
		var seen []string
		c := newCaseTemplateTestClient(t, status, `{"message":"rejected"}`, &seen)
		req := kbapi.CasesTemplateWriteRequest{}

		got, diags := CreateCaseTemplate(context.Background(), c, "default", req)
		assert.Nil(t, got)
		assert.True(t, diags.HasError(), "create status %d", status)

		got, diags = UpdateCaseTemplate(context.Background(), c, "default", "tpl-1", req)
		assert.Nil(t, got)
		assert.True(t, diags.HasError(), "update status %d", status)
	}
}
