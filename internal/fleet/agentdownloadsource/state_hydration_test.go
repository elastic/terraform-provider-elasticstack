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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
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

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

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
	require.Nil(t, state.Auth)
}

func TestReadAndHydrateStatePreservesRedactedPassword(t *testing.T) {
	t.Parallel()

	sourceID := "source-basic-auth"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Username: types.StringValue("stale-user"),
		Password: types.StringValue("secret"),
		APIKey:   types.StringNull(),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
	require.Equal(t, "secret", state.Auth.Password.ValueString())
	require.True(t, state.Auth.APIKey.IsNull())
}

func TestReadAndHydrateStateUsesAPISecretWhenPresent(t *testing.T) {
	t.Parallel()

	sourceID := "source-echoed-secret"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Password: types.StringValue("configured-password"),
		APIKey:   types.StringValue("configured-key"),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
					"password": "echoed-password",
					"api_key":  "echoed-key",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
	require.Equal(t, "echoed-password", state.Auth.Password.ValueString())
	require.Equal(t, "echoed-key", state.Auth.APIKey.ValueString())
}

func TestReadAndHydrateStatePreservesRedactedAPIKeyIndependently(t *testing.T) {
	t.Parallel()

	sourceID := "source-api-key-auth"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Password: types.StringNull(),
		APIKey:   types.StringValue("secret-key"),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.True(t, state.Auth.Username.IsNull())
	require.True(t, state.Auth.Password.IsNull())
	require.Equal(t, "secret-key", state.Auth.APIKey.ValueString())
}

func TestReadAndHydrateStateDecodesAbsentHeadersAsNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sourceID := "source-absent-headers"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Headers: mustAuthHeaders(t, authHeaderModel{Key: types.StringValue("X-A"), Value: types.StringValue("1")}),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(ctx, client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.True(t, state.Auth.Headers.IsNull())
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
}

func TestReadAndHydrateStateUsesAPIHeadersNotFallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sourceID := "source-api-headers"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Username: types.StringValue("stale-user"),
		Headers:  mustAuthHeaders(t, authHeaderModel{Key: types.StringValue("X-A"), Value: types.StringValue("1")}),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
					"headers": []any{
						map[string]any{"key": "X-B", "value": "2"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(ctx, client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
	require.True(t, state.Auth.Headers.Equal(mustAuthHeaders(t, authHeaderModel{Key: types.StringValue("X-B"), Value: types.StringValue("2")})))
}

func TestReadAndHydrateStateLeavesSecretsNullWithoutFallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sourceID := "source-import-redacted"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
					"headers": []any{
						map[string]any{"key": "X-A", "value": "1"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(ctx, client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
	require.True(t, state.Auth.Password.IsNull())
	require.True(t, state.Auth.APIKey.IsNull())
	require.True(t, state.Auth.Headers.Equal(mustAuthHeaders(t, authHeaderModel{Key: types.StringValue("X-A"), Value: types.StringValue("1")})))
}

func TestReadAndHydrateStateDecodesEmptyHeadersList(t *testing.T) {
	t.Parallel()

	sourceID := "source-empty-headers"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"headers": []any{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.False(t, state.Auth.Headers.IsNull())
	require.Empty(t, state.Auth.Headers.Elements())
}

func TestReadAndHydrateStateTreatsEmptyProxyIDAsNull(t *testing.T) {
	t.Parallel()

	sourceID := "source-cleared-proxy"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"proxy_id":   "",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.True(t, state.ProxyID.IsNull())
}

func TestReadAndHydrateStateHydratesSslFromReadPayload(t *testing.T) {
	t.Parallel()

	sourceID := "source-ssl-read"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"proxy_id":   "proxy-from-read",
				"ssl": map[string]any{
					"certificate":             "cert-from-read",
					"certificate_authorities": []string{"ca-from-read"},
					"key":                     "key-from-read",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.False(t, state.Ssl.IsNull(), "expected a populated ssl object in state")

	var readDiags diag.Diagnostics
	sslModel := typeutils.ObjectTypeAs[sslModel](context.Background(), state.Ssl, fwpath.Root("ssl"), &readDiags)
	require.False(t, readDiags.HasError(), "unexpected diagnostics: %#v", readDiags)
	require.Equal(t, "cert-from-read", sslModel.Certificate.ValueString())
	var caDiags diag.Diagnostics
	require.Equal(t, []string{"ca-from-read"}, typeutils.ListTypeToSliceString(context.Background(), sslModel.CertificateAuthorities, fwpath.Root("certificate_authorities"), &caDiags))
	require.Equal(t, "key-from-read", sslModel.Key.ValueString())
}

func TestReadAndHydrateStateNullsAbsentAndEmptySsl(t *testing.T) {
	t.Parallel()

	for name, sslPayload := range map[string]any{
		"ssl absent": nil,
		"ssl present but empty": map[string]any{
			"certificate":             "",
			"certificate_authorities": []string{},
			"key":                     "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sourceID := "source-ssl-" + strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			spaceID := "space-a"
			preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
			preservedKibanaConnection := providerschema.KibanaConnectionNullList()

			item := map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"proxy_id":   "proxy-from-read",
			}
			if sslPayload != nil {
				item["ssl"] = sslPayload
			}

			client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]any{"item": item}); err != nil {
					t.Errorf("failed to encode response: %v", err)
				}
			}))

			state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, nil, nullSslObject())

			require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
			require.True(t, found)
			require.True(t, state.Ssl.IsNull(), "expected a null ssl object in state")
		})
	}
}

func TestUpdateAgentDownloadSourceReadBackKeepsRotatedPassword(t *testing.T) {
	t.Parallel()

	sourceID := "source-rotate-password"
	spaceID := "space-a"
	spaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"auth": map[string]any{
					"username": "svc-account",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	plan := model{
		SourceID: types.StringValue(sourceID),
		Name:     types.StringValue("name-from-read"),
		Host:     types.StringValue("https://read.example.com"),
		Default:  types.BoolValue(false),
		SpaceIDs: spaceIDs,
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("new-secret"),
		},
	}
	prior := model{
		SourceID: types.StringValue(sourceID),
		SpaceIDs: spaceIDs,
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("old-secret"),
		},
	}

	result, diags := updateAgentDownloadSource(context.Background(), client, plan, prior, nil)

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, result.Model.Auth)
	require.Equal(t, "svc-account", result.Model.Auth.Username.ValueString())
	require.Equal(t, "new-secret", result.Model.Auth.Password.ValueString())
}

func TestCreateAgentDownloadSourceReportsHeadersConversionFailure(t *testing.T) {
	t.Parallel()

	var postBody string
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			postBody = string(raw)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"item": map[string]any{
				"id":         "source-bad-headers",
				"name":       "download-source",
				"host":       "https://artifacts.example.com/elastic-agent",
				"is_default": false,
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	plan := model{
		Name:     types.StringValue("download-source"),
		Host:     types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default:  types.BoolValue(false),
		SourceID: types.StringValue("source-bad-headers"),
		Auth: &authModel{
			Headers: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("not-a-header"),
			}),
		},
	}

	_, diags := createAgentDownloadSource(context.Background(), client, plan)

	require.True(t, diags.HasError(), "headers conversion failure must surface a diagnostic; request body: %s", postBody)
	require.Empty(t, postBody, "headers conversion failure must not send a request that omits headers")
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

func TestReadPathPassesPriorAuthAndWritePathPassesPlanAuth(t *testing.T) {
	t.Parallel()

	requireAuthFallbackArg(t, "read.go", "readAgentDownloadSource", "prior")
	requireAuthFallbackArg(t, "write.go", "finalizeWrite", "plan")
}

func requireAuthFallbackArg(t *testing.T, filename, funcName, receiver string) {
	t.Helper()

	funcDecl := findFuncDecl(t, filename, funcName)

	var matched bool
	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		if matched {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "readAndHydrateState" || len(call.Args) < 7 {
			return true
		}
		sel, ok := call.Args[6].(*ast.SelectorExpr)
		if !ok || sel.Sel == nil || sel.Sel.Name != "Auth" {
			return true
		}
		recv, ok := sel.X.(*ast.Ident)
		if ok && recv.Name == receiver {
			matched = true
			return false
		}
		return true
	})

	require.True(t, matched, "%s should pass %s.Auth to readAndHydrateState", funcName, receiver)
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

func TestReadAndHydrateStateHydratesAuthAndSSLTogether(t *testing.T) {
	t.Parallel()

	sourceID := "source-auth-ssl-read"
	spaceID := "space-a"
	preservedSpaceIDs := types.SetValueMust(types.StringType, []attr.Value{types.StringValue(spaceID)})
	preservedKibanaConnection := providerschema.KibanaConnectionNullList()
	fallback := &authModel{
		Username: types.StringValue("svc-account"),
		Password: types.StringValue("configured-secret"),
	}

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"item": map[string]any{
				"id":         sourceID,
				"name":       "name-from-read",
				"host":       "https://read.example.com",
				"is_default": false,
				"proxy_id":   "proxy-from-read",
				"auth": map[string]any{
					"username": "svc-account",
					"headers":  []map[string]any{{"key": "X-A", "value": "1"}},
				},
				"ssl": map[string]any{
					"certificate":             "cert-from-read",
					"certificate_authorities": []string{"ca-from-read"},
					"key":                     "key-from-read",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))

	state, found, diags := readAndHydrateState(context.Background(), client, sourceID, spaceID, preservedSpaceIDs, preservedKibanaConnection, fallback, nullSslObject())

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	require.NotNil(t, state.Auth)
	require.Equal(t, "svc-account", state.Auth.Username.ValueString())
	require.Equal(t, "configured-secret", state.Auth.Password.ValueString(), "redacted password must be preserved from the fallback")
	require.False(t, state.Ssl.IsNull(), "expected a populated ssl object in state")

	var readDiags diag.Diagnostics
	sslModel := typeutils.ObjectTypeAs[sslModel](context.Background(), state.Ssl, fwpath.Root("ssl"), &readDiags)
	require.False(t, readDiags.HasError(), "unexpected diagnostics: %#v", readDiags)
	require.Equal(t, "cert-from-read", sslModel.Certificate.ValueString())
}
