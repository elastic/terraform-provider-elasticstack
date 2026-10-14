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

package casefielddefinition

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReadTestClient(t *testing.T, status int, body string) *clients.KibanaScopedClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	api, err := kbapi.NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	return clients.NewKibanaScopedClientForTest(&kibanaoapi.Client{URL: srv.URL, HTTP: srv.Client(), API: api})
}

// TestReadCallbackFailureRetainsState verifies the resource's read callback
// reports a server error as an error diagnostic with found=false and the prior
// model intact. The entitycore envelope only removes state when found=false
// is returned without errors (see TestNewKibanaResource_Read_shortCircuitReadFuncError),
// so a transient failure is never treated as a deletion.
func TestReadCallbackFailureRetainsState(t *testing.T) {
	read := entitycore.SimpleKibanaRead[fieldDefinitionModel, kbapi.CasesFieldDefinitionResponse](
		kibanaoapi.GetCaseFieldDefinition, (*fieldDefinitionModel).populateFromAPI,
	)
	prior := fieldDefinitionModel{FieldDefinitionID: types.StringValue("fd-1"), Name: types.StringValue("prior")}

	t.Run("500 is an error and prior is kept", func(t *testing.T) {
		client := newReadTestClient(t, http.StatusInternalServerError, `{"message":"boom"}`)
		got, found, diags := read(context.Background(), client, "fd-1", "default", prior)
		assert.True(t, diags.HasError())
		assert.False(t, found)
		assert.Equal(t, "prior", got.Name.ValueString())
	})

	t.Run("404 is not found without error", func(t *testing.T) {
		client := newReadTestClient(t, http.StatusNotFound, `{"message":"nope"}`)
		_, found, diags := read(context.Background(), client, "fd-1", "default", prior)
		assert.False(t, diags.HasError())
		assert.False(t, found)
	})
}
