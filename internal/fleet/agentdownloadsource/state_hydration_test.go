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

package agentdownloadsource

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestReadAndHydrateStateUsesReadPayload(t *testing.T) {
	t.Parallel()

	sourceID := "source-from-mutation"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: got %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/s/"+spaceID+"/api/fleet/agent_download_sources/"+sourceID {
			t.Errorf("unexpected path: got %q", r.URL.Path)
		}

		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"proxy_id":   "proxy-from-read",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection)

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.Equal(t, sourceID, state.ID.ValueString())
	require.Equal(t, sourceID, state.SourceID.ValueString())
	require.Equal(t, "name-from-read", state.Name.ValueString())
	require.Equal(t, "https://read.example.com", state.Host.ValueString())
	require.False(t, state.Default.ValueBool())
	require.Equal(t, "proxy-from-read", state.ProxyID.ValueString())
	require.Equal(t, preservedSpaceIDs, state.SpaceIDs)
	require.Equal(t, preservedKibanaConnection, state.KibanaConnection)
}

func TestCreateAndUpdateFinalizeStateViaReadHydration(t *testing.T) {
	t.Parallel()

	assertFuncCallsFinalizeWrite(t, "create.go", "createAgentDownloadSource")
	assertFuncCallsFinalizeWrite(t, "update.go", "updateAgentDownloadSource")
	assertFinalizeWriteUsesReadHydration(t)
}

// assertFuncCallsFinalizeWrite verifies that the named write-callback helper
// delegates to the shared finalizeWrite helper (write.go) to build its result,
// rather than trusting the raw create/update API response directly.
func assertFuncCallsFinalizeWrite(t *testing.T, filename string, funcName string) {
	t.Helper()

	funcDecl := findFuncDecl(t, filename, funcName)

	var callsFinalizeWrite bool
	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		if callsFinalizeWrite {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "finalizeWrite" {
				callsFinalizeWrite = true
				return false
			}
		}
		return true
	})

	require.True(t, callsFinalizeWrite, "%s should call finalizeWrite to build its result from a read-back", funcName)
}

// assertFinalizeWriteUsesReadHydration verifies that finalizeWrite (write.go)
// calls readAndHydrateState and returns its result as the write result's
// Model, rather than trusting the raw create/update API response directly.
func assertFinalizeWriteUsesReadHydration(t *testing.T) {
	t.Helper()

	funcDecl := findFuncDecl(t, "write.go", "finalizeWrite")

	var (
		hasReadAndHydrateCall bool
		hasResultFromRead     bool
	)

	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		if hasReadAndHydrateCall && hasResultFromRead {
			return false
		}

		if call, ok := node.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "readAndHydrateState" {
				hasReadAndHydrateCall = true
			}
		}

		if composite, ok := node.(*ast.CompositeLit); ok {
			for _, elt := range composite.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, keyOk := kv.Key.(*ast.Ident)
				value, valueOk := kv.Value.(*ast.Ident)
				if keyOk && valueOk && key.Name == "Model" && value.Name == "readState" {
					hasResultFromRead = true
				}
			}
		}

		return true
	})

	require.True(t, hasReadAndHydrateCall, "finalizeWrite should call readAndHydrateState")
	require.True(t, hasResultFromRead, "finalizeWrite should return a write result built from readState")
}

// findFuncDecl parses filename and returns the *ast.FuncDecl named funcName.
func findFuncDecl(t *testing.T, filename, funcName string) *ast.FuncDecl {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, 0)
	require.NoError(t, err)

	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if ok && fd.Name != nil && fd.Name.Name == funcName {
			return fd
		}
	}

	require.FailNow(t, "function not found", "%s not found in %s", funcName, filename)
	return nil
}

func newTestFleetClient(t *testing.T, handler http.Handler) *fleet.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	api, err := kbapi.NewClientWithResponses(server.URL+"/", kbapi.WithHTTPClient(server.Client()))
	require.NoError(t, err)

	return &fleet.Client{
		URL:  server.URL,
		HTTP: server.Client(),
		API:  api,
	}
}
