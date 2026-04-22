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

	resource := &Resource{}
	state, found, diags := resource.readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection)

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

	assertMethodUsesReadHydration(t, "create.go", "Create")
	assertMethodUsesReadHydration(t, "update.go", "Update")
}

func assertMethodUsesReadHydration(t *testing.T, filename string, methodName string) {
	t.Helper()

	path := filename
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	require.NoError(t, err)

	var methodDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != methodName {
			continue
		}
		methodDecl = funcDecl
		break
	}

	require.NotNil(t, methodDecl, "method %s not found in %s", methodName, path)

	var (
		hasReadAndHydrateCall bool
		hasStateSetFromRead   bool
	)

	ast.Inspect(methodDecl.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && selector.Sel != nil && selector.Sel.Name == "readAndHydrateState" {
			hasReadAndHydrateCall = true
		}

		if ok && selector.Sel != nil && selector.Sel.Name == "Set" && len(call.Args) >= 2 {
			if ident, identOK := call.Args[1].(*ast.Ident); identOK && ident.Name == "readState" {
				hasStateSetFromRead = true
			}
		}

		return true
	})

	require.True(t, hasReadAndHydrateCall, "%s should call readAndHydrateState", methodName)
	require.True(t, hasStateSetFromRead, "%s should set final state from readState", methodName)
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
