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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type fakePrivate struct {
	data map[string][]byte
}

var _ entitycore.PrivateStateStorage = (*fakePrivate)(nil)

func newFakePrivate() *fakePrivate { return &fakePrivate{data: map[string][]byte{}} }

func (f *fakePrivate) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return f.data[key], nil
}

func (f *fakePrivate) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if value == nil {
		delete(f.data, key)
		return nil
	}
	f.data[key] = value
	return nil
}

func (f *fakePrivate) storedHash(t *testing.T, attrPath string) []byte {
	t.Helper()
	hash, err := decodeSecretHashFromPrivateState(f.data[secretHashKey(attrPath)])
	require.NoError(t, err)
	return hash
}

func TestSecretHashKeys(t *testing.T) {
	t.Parallel()
	require.Equal(t, "secret_hash:ssl.key_wo", secretHashKey(sslKeyWOPath))
	require.Equal(t, "secret_hash:auth.password_wo", secretHashKey(authPasswordWOPath))
	require.Equal(t, "secret_hash:auth.api_key_wo", secretHashKey(authAPIKeyWOPath))
}

func TestSecretHashEncodingRoundTrips(t *testing.T) {
	t.Parallel()
	hash, err := secretHasher.Compute("value")
	require.NoError(t, err)
	encoded, err := encodeSecretHashForPrivateState(hash)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(encoded), `"`), "bcrypt bytes contain $ and must be JSON wrapped")
	decoded, err := decodeSecretHashFromPrivateState(encoded)
	require.NoError(t, err)
	require.Equal(t, hash, decoded)

	empty, err := decodeSecretHashFromPrivateState(nil)
	require.NoError(t, err)
	require.Nil(t, empty)
}

func TestHashHelpersHandlePEMSizedValues(t *testing.T) {
	t.Parallel()
	pem := "-----BEGIN PRIVATE KEY-----\n" + strings.Repeat("A", 1600) + "\n-----END PRIVATE KEY-----"
	hash, err := secretHasher.Compute(pem)
	require.NoError(t, err)
	require.True(t, secretHasher.Matches(pem, hash))
	require.False(t, secretHasher.Matches(pem+"x", hash))
}

func TestLoadStoredHashes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	got, diags := loadStoredHashes(ctx, nil)
	require.False(t, diags.HasError())
	require.Empty(t, got)

	private := newFakePrivate()
	var persistDiags diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k"}, nil, &persistDiags)
	require.False(t, persistDiags.HasError())

	got, diags = loadStoredHashes(ctx, private)
	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	require.True(t, secretHasher.Matches("k", got[sslKeyWOPath]))
}

func TestPersistSecretHashes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("stores a hash for each configured value", func(t *testing.T) {
		t.Parallel()
		private := newFakePrivate()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k", authPasswordWOPath: "p", authAPIKeyWOPath: "a"}, nil, &diags)
		require.False(t, diags.HasError())
		require.True(t, secretHasher.Matches("k", private.storedHash(t, sslKeyWOPath)))
		require.True(t, secretHasher.Matches("p", private.storedHash(t, authPasswordWOPath)))
		require.True(t, secretHasher.Matches("a", private.storedHash(t, authAPIKeyWOPath)))
	})

	t.Run("keeps an existing matching hash", func(t *testing.T) {
		t.Parallel()
		private := newFakePrivate()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k"}, nil, &diags)
		existing := private.data[secretHashKey(sslKeyWOPath)]

		stored, loadDiags := loadStoredHashes(ctx, private)
		require.False(t, loadDiags.HasError())
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k"}, stored, &diags)
		require.False(t, diags.HasError())
		require.Equal(t, existing, private.data[secretHashKey(sslKeyWOPath)], "an unchanged value must not be rehashed")
	})

	t.Run("replaces the hash when the value changed", func(t *testing.T) {
		t.Parallel()
		private := newFakePrivate()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "old"}, nil, &diags)
		stored, _ := loadStoredHashes(ctx, private)
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "new"}, stored, &diags)
		require.False(t, diags.HasError())
		require.True(t, secretHasher.Matches("new", private.storedHash(t, sslKeyWOPath)))
	})

	t.Run("clears the hash of a removed value", func(t *testing.T) {
		t.Parallel()
		private := newFakePrivate()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k", authPasswordWOPath: "p"}, nil, &diags)
		stored, _ := loadStoredHashes(ctx, private)

		persistSecretHashes(ctx, private, map[string]string{authPasswordWOPath: "p"}, stored, &diags)
		require.False(t, diags.HasError())
		require.NotContains(t, private.data, secretHashKey(sslKeyWOPath))
		require.Contains(t, private.data, secretHashKey(authPasswordWOPath))
	})

	t.Run("nil private state with a configured value is an error", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, nil, map[string]string{sslKeyWOPath: "k"}, nil, &diags)
		require.True(t, diags.HasError())
	})

	t.Run("nil private state with nothing configured is fine", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		persistSecretHashes(ctx, nil, nil, nil, &diags)
		require.False(t, diags.HasError())
	})
}

func TestClearAllSecretHashes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	private := newFakePrivate()
	var diags diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "k", authPasswordWOPath: "p", authAPIKeyWOPath: "a"}, nil, &diags)
	require.Len(t, private.data, 3)

	clearAllSecretHashes(ctx, private, &diags)
	require.False(t, diags.HasError())
	require.Empty(t, private.data)

	clearAllSecretHashes(ctx, nil, &diags)
	require.False(t, diags.HasError(), "nil private state is a no-op")
}

// writeRequest builds a write request whose plan has null write-only values and
// whose config carries them, as the framework presents them.
func writeRequestForKeyWO(t *testing.T, configured string, prior *model, private entitycore.PrivateStateStorage) entitycore.KibanaWriteRequest[model] {
	t.Helper()
	ctx := context.Background()
	plan := updatePlanWithKeyWO(ctx, "x")
	plan.Ssl = testSslObject(ctx, types.StringNull(), types.StringNull())
	config := updatePlanWithKeyWO(ctx, configured)
	if configured == "" {
		config.Ssl = testSslObject(ctx, types.StringNull(), types.StringNull())
	}
	return entitycore.KibanaWriteRequest[model]{Plan: plan, Config: config, Prior: prior, Private: private}
}

func TestWriteReadsWriteOnlyValuesFromConfigAndPersistsHashOnCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var postBody string
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			raw, _ := io.ReadAll(r.Body)
			postBody = string(raw)
		}
		itemServer(t, map[string]any{"id": "src", "name": "n", "host": "https://h", "is_default": false}).ServeHTTP(w, r)
	}))

	private := newFakePrivate()
	req := writeRequestForKeyWO(t, "wo-key", nil, private)
	_, diags := writeWithClient(ctx, client, req)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	require.Contains(t, postBody, `"secrets":{"ssl":{"key":"wo-key"}}`, "the literal comes from config, not the null-write-only plan")
	require.True(t, secretHasher.Matches("wo-key", private.storedHash(t, sslKeyWOPath)))
}

func TestWriteDoesNotPersistHashWhenRequestFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))

	private := newFakePrivate()
	_, diags := writeWithClient(ctx, client, writeRequestForKeyWO(t, "wo-key", nil, private))
	require.True(t, diags.HasError())
	require.Empty(t, private.data)
}

func TestWriteUpdateKeepsHashUntilClearSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	private := newFakePrivate()
	var seed diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "wo-key"}, nil, &seed)

	// A failing clearing update leaves the hash in place so the next plan retries.
	failing := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	prior := updatePlanWithKeyWO(ctx, "x")
	_, diags := writeWithClient(ctx, failing, writeRequestForKeyWO(t, "", &prior, private))
	require.True(t, diags.HasError())
	require.Contains(t, private.data, secretHashKey(sslKeyWOPath), "hash must survive a failed clearing update")

	h, calls := updateRecorder(t, nil)
	ok := newTestFleetClient(t, h)
	_, diags = writeWithClient(ctx, ok, writeRequestForKeyWO(t, "", &prior, private))
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotContains(t, private.data, secretHashKey(sslKeyWOPath), "hash is removed after the clearing update succeeds")
	for _, c := range calls() {
		if c.method == http.MethodPut {
			require.NotContains(t, c.body, "secrets")
		}
	}

	// Re-adding after a successful removal has no stored hash, so a literal is sent.
	h, calls = updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	readd := newTestFleetClient(t, h)
	_, diags = writeWithClient(ctx, readd, writeRequestForKeyWO(t, "wo-key", &prior, private))
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	readdCalls := calls()
	require.Equal(t, http.MethodPut, readdCalls[0].method, "no lookup before the PUT without a stored hash")
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "wo-key"}}, readdCalls[0].body["secrets"])
	require.True(t, secretHasher.Matches("wo-key", private.storedHash(t, sslKeyWOPath)))
}

func TestWriteUpdateReferencesUnchangedSecretByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	private := newFakePrivate()
	var seed diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "wo-key"}, nil, &seed)
	before := private.data[secretHashKey(sslKeyWOPath)]

	h, calls := updateRecorder(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}})
	client := newTestFleetClient(t, h)
	prior := updatePlanWithKeyWO(ctx, "x")
	_, diags := writeWithClient(ctx, client, writeRequestForKeyWO(t, "wo-key", &prior, private))
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

	for _, c := range calls() {
		if c.method == http.MethodPut {
			require.Equal(t, map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}}, c.body["secrets"])
		}
	}
	require.Equal(t, before, private.data[secretHashKey(sslKeyWOPath)])
}
