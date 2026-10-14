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

const caseFieldDefinitionBody = `{"fieldDefinitionId":"fd-1","owner":"cases","name":"f","definition":"name: f\ntype: keyword","displayOrder":2}`

func newCaseFieldDefinitionTestClient(t *testing.T, status int, body string, seen *[]string) *Client {
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

func TestCaseFieldDefinitionRouting(t *testing.T) {
	ctx := context.Background()

	t.Run("create/read/update use default-space routes", func(t *testing.T) {
		var seen []string
		c := newCaseFieldDefinitionTestClient(t, http.StatusOK, caseFieldDefinitionBody, &seen)

		created, diags := CreateCaseFieldDefinition(ctx, c, "default", kbapi.CasesFieldDefinitionWriteRequest{Owner: "cases", Definition: "x"})
		require.False(t, diags.HasError())
		assert.Equal(t, "fd-1", created.FieldDefinitionId)

		got, diags := GetCaseFieldDefinition(ctx, c, "default", "fd-1")
		require.False(t, diags.HasError())
		assert.Equal(t, "f", got.Name)

		_, diags = UpdateCaseFieldDefinition(ctx, c, "default", "fd-1", kbapi.CasesFieldDefinitionPutRequest{Owner: "cases", Definition: "x"})
		require.False(t, diags.HasError())

		assert.Equal(t, []string{
			"POST /api/cases/field_definitions",
			"GET /api/cases/field_definitions/fd-1",
			"PUT /api/cases/field_definitions/fd-1",
		}, seen)
	})

	t.Run("named space routes are prefixed", func(t *testing.T) {
		var seen []string
		c := newCaseFieldDefinitionTestClient(t, http.StatusOK, caseFieldDefinitionBody, &seen)

		_, diags := CreateCaseFieldDefinition(ctx, c, "ops", kbapi.CasesFieldDefinitionWriteRequest{Owner: "cases", Definition: "x"})
		require.False(t, diags.HasError())
		_, diags = GetCaseFieldDefinition(ctx, c, "ops", "fd-1")
		require.False(t, diags.HasError())

		assert.Equal(t, []string{
			"POST /s/ops/api/cases/field_definitions",
			"GET /s/ops/api/cases/field_definitions/fd-1",
		}, seen)
	})

	t.Run("read 404 returns nil without error", func(t *testing.T) {
		var seen []string
		c := newCaseFieldDefinitionTestClient(t, http.StatusNotFound, `{"message":"nope"}`, &seen)
		got, diags := GetCaseFieldDefinition(ctx, c, "default", "fd-1")
		assert.Nil(t, got)
		assert.False(t, diags.HasError())
	})
}

func TestDeleteCaseFieldDefinitionStatuses(t *testing.T) {
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
			c := newCaseFieldDefinitionTestClient(t, tc.status, body, &seen)
			diags := DeleteCaseFieldDefinition(context.Background(), c, "ops", "fd-1")
			assert.Equal(t, tc.wantErr, diags.HasError())
			assert.Equal(t, []string{"DELETE /s/ops/api/cases/field_definitions/fd-1"}, seen)
		})
	}
}

func TestCaseFieldDefinitionMutateErrors(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusConflict} {
		var seen []string
		c := newCaseFieldDefinitionTestClient(t, status, `{"message":"rejected"}`, &seen)
		got, diags := UpdateCaseFieldDefinition(context.Background(), c, "default", "fd-1", kbapi.CasesFieldDefinitionPutRequest{})
		assert.Nil(t, got)
		assert.True(t, diags.HasError(), "status %d", status)
	}
}
