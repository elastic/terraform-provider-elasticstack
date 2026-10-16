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
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	method string
	body   map[string]any
}

// updateRecorder serves GET (with the given secrets) and PUT requests for one
// source and records the sequence of calls.
func updateRecorder(t *testing.T, getSecrets map[string]any) (http.Handler, func() []recordedRequest) {
	t.Helper()
	var mu sync.Mutex
	var calls []recordedRequest
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordedRequest{method: r.Method}
		if r.Method == http.MethodPut {
			raw, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			rec.body = map[string]any{}
			assert.NoError(t, json.Unmarshal(raw, &rec.body))
		}
		mu.Lock()
		calls = append(calls, rec)
		mu.Unlock()

		item := map[string]any{"id": "src", "name": "n", "host": "https://h", "is_default": false}
		if r.Method == http.MethodGet && getSecrets != nil {
			item["secrets"] = getSecrets
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"item": item}))
	})
	return h, func() []recordedRequest {
		mu.Lock()
		defer mu.Unlock()
		return append([]recordedRequest(nil), calls...)
	}
}

func updatePlanWithKeyWO(ctx context.Context, keyWO string) model {
	return model{
		SourceID: types.StringValue("src"),
		Name:     types.StringValue("n"),
		Host:     types.StringValue("https://h"),
		Default:  types.BoolValue(false),
		SpaceIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate":             types.StringValue("cert"),
			"certificate_authorities": types.ListNull(types.StringType),
			"key":                     types.StringNull(),
			"key_wo":                  types.StringValue(keyWO),
		}),
	}
}

func TestUpdateReferencesUnchangedSecretByIDAfterGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hash, err := secretHasher.Compute("wo-key")
	require.NoError(t, err)

	h, calls := updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	client := newTestFleetClient(t, h)
	plan := updatePlanWithKeyWO(ctx, "wo-key")

	_, diags := updateAgentDownloadSource(ctx, client, plan, plan, map[string][]byte{sslKeyWOPath: hash})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	got := calls()
	require.GreaterOrEqual(t, len(got), 2)
	require.Equal(t, http.MethodGet, got[0].method, "the secret id lookup must precede the PUT")
	require.Equal(t, http.MethodPut, got[1].method)
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}}, got[1].body["secrets"])
}

func TestUpdateSendsLiteralWhenValueChanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hash, err := secretHasher.Compute("old-key")
	require.NoError(t, err)

	h, calls := updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	client := newTestFleetClient(t, h)
	plan := updatePlanWithKeyWO(ctx, "new-key")

	_, diags := updateAgentDownloadSource(ctx, client, plan, plan, map[string][]byte{sslKeyWOPath: hash})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	var put recordedRequest
	for _, c := range calls() {
		if c.method == http.MethodPut {
			put = c
		}
	}
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "new-key"}}, put.body["secrets"])
}

func TestUpdateSendsLiteralWhenFleetHoldsNoSecretID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hash, err := secretHasher.Compute("wo-key")
	require.NoError(t, err)

	h, calls := updateRecorder(t, nil)
	client := newTestFleetClient(t, h)
	plan := updatePlanWithKeyWO(ctx, "wo-key")

	_, diags := updateAgentDownloadSource(ctx, client, plan, plan, map[string][]byte{sslKeyWOPath: hash})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	var put recordedRequest
	for _, c := range calls() {
		if c.method == http.MethodPut {
			put = c
		}
	}
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "wo-key"}}, put.body["secrets"])
}

func TestUpdateWithoutStoredHashSkipsSecretLookup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	h, calls := updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	client := newTestFleetClient(t, h)
	plan := updatePlanWithKeyWO(ctx, "wo-key")

	_, diags := updateAgentDownloadSource(ctx, client, plan, plan, nil)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	got := calls()
	require.Equal(t, http.MethodPut, got[0].method, "no GET before the PUT when no hash is stored")
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "wo-key"}}, got[0].body["secrets"])
}

func TestUpdateRemovedWriteOnlyValueSendsSslWithoutSecret(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hash, err := secretHasher.Compute("wo-key")
	require.NoError(t, err)

	h, calls := updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	client := newTestFleetClient(t, h)
	plan := updatePlanWithKeyWO(ctx, "ignored")
	plan.Ssl = types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
		"certificate":             types.StringValue("cert"),
		"certificate_authorities": types.ListNull(types.StringType),
		"key":                     types.StringNull(),
		"key_wo":                  types.StringNull(),
	})

	_, diags := updateAgentDownloadSource(ctx, client, plan, plan, map[string][]byte{sslKeyWOPath: hash})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	got := calls()
	require.Equal(t, http.MethodPut, got[0].method, "a removed value is not referenced, so no lookup is needed")
	require.NotContains(t, got[0].body, "secrets")
	require.Contains(t, got[0].body, "ssl")
}

func TestWriteOnlyValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := updatePlanWithKeyWO(ctx, "k")
	m.Auth = &authModel{
		Username: types.StringValue("u"), Password: types.StringNull(),
		PasswordWO: types.StringValue("p"), APIKey: types.StringNull(), APIKeyWO: types.StringUnknown(),
		Headers: types.ListNull(authHeadersElementType(ctx)),
	}
	got, diags := m.writeOnlyValues(ctx)
	require.False(t, diags.HasError())
	require.Equal(t, map[string]string{sslKeyWOPath: "k", authPasswordWOPath: "p"}, got)

	got, diags = model{}.writeOnlyValues(ctx)
	require.False(t, diags.HasError())
	require.Empty(t, got)
}
