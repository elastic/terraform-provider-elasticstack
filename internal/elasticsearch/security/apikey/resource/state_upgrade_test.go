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

package resource

import (
	"context"
	"encoding/json"
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func baseAPIKeyState() map[string]any {
	return map[string]any{
		"id":     "test-cluster-uuid/my-key",
		"key_id": "abc123",
		"name":   "my-key",
	}
}

func runUpgrade(t *testing.T, priorVersion int64, raw map[string]any) *fwresource.UpgradeStateResponse {
	t.Helper()
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	r := newResource()
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[priorVersion]
	require.True(t, ok, "expected upgrader for schema version %d", priorVersion)

	req := fwresource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: rawJSON},
	}
	resp := &fwresource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), req, resp)
	return resp
}

func requireUpgradedJSON(t *testing.T, resp *fwresource.UpgradeStateResponse) map[string]any {
	t.Helper()
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)
	require.NotNil(t, resp.DynamicValue.JSON)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	return got
}

func requireNullifiedKey(t *testing.T, got map[string]any, key string) {
	t.Helper()
	v, ok := got[key]
	require.True(t, ok, "expected %q to be present after nullification", key)
	require.Nil(t, v, "expected %q to be JSON null", key)
}

func TestUpgradeStateV0ToV1(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		patch  map[string]any
		assert func(t *testing.T, got map[string]any)
	}{
		{
			name: "metadata_empty_string",
			patch: map[string]any{
				attrMetadata: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrMetadata)
			},
		},
		{
			name: "role_descriptors_empty_string",
			patch: map[string]any{
				attrRoleDescriptors: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrRoleDescriptors)
			},
		},
		{
			name: "metadata_and_role_descriptors_empty_string",
			patch: map[string]any{
				attrMetadata:        "",
				attrRoleDescriptors: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrMetadata)
				requireNullifiedKey(t, got, attrRoleDescriptors)
			},
		},
		{
			name: "metadata_valid_json_preserved",
			patch: map[string]any{
				attrMetadata: `{}`,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				metadata, ok := got[attrMetadata].(string)
				require.True(t, ok)
				require.JSONEq(t, `{}`, metadata)
			},
		},
		{
			name: "role_descriptors_valid_json_preserved",
			patch: map[string]any{
				attrRoleDescriptors: `{"role-a":{"cluster":["all"]}}`,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				roleDescriptors, ok := got[attrRoleDescriptors].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"role-a":{"cluster":["all"]}}`, roleDescriptors)
			},
		},
		{
			name: "metadata_null",
			patch: map[string]any{
				attrMetadata: nil,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Nil(t, got[attrMetadata])
			},
		},
		{
			name:  "metadata_absent",
			patch: map[string]any{},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				_, ok := got[attrMetadata]
				require.False(t, ok)
			},
		},
		{
			name: "expiration_empty_string",
			patch: map[string]any{
				attrExpiration: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrExpiration)
			},
		},
		{
			name: "expiration_non_empty_preserved",
			patch: map[string]any{
				attrExpiration: "7d",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Equal(t, "7d", got[attrExpiration])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseAPIKeyState()
			maps.Copy(raw, tc.patch)

			got := requireUpgradedJSON(t, runUpgrade(t, 0, raw))
			tc.assert(t, got)
		})
	}
}

func TestUpgradeStateV1ToV2(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		patch  map[string]any
		assert func(t *testing.T, got map[string]any)
	}{
		{
			name: "metadata_empty_string",
			patch: map[string]any{
				attrMetadata: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrMetadata)
			},
		},
		{
			name: "role_descriptors_empty_string",
			patch: map[string]any{
				attrRoleDescriptors: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrRoleDescriptors)
			},
		},
		{
			name: "metadata_and_role_descriptors_empty_string",
			patch: map[string]any{
				attrMetadata:        "",
				attrRoleDescriptors: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				requireNullifiedKey(t, got, attrMetadata)
				requireNullifiedKey(t, got, attrRoleDescriptors)
			},
		},
		{
			name: "metadata_valid_json_preserved",
			patch: map[string]any{
				attrMetadata: `{"env":"prod"}`,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				metadata, ok := got[attrMetadata].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"env":"prod"}`, metadata)
			},
		},
		{
			name: "role_descriptors_valid_json_preserved",
			patch: map[string]any{
				attrRoleDescriptors: `{"role-a":{"cluster":["all"]}}`,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				roleDescriptors, ok := got[attrRoleDescriptors].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"role-a":{"cluster":["all"]}}`, roleDescriptors)
			},
		},
		{
			name:  "type_absent_defaults_to_rest",
			patch: map[string]any{},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Equal(t, apikey.DefaultAPIKeyType, got[attrType])
			},
		},
		{
			name: "type_null_defaults_to_rest",
			patch: map[string]any{
				attrType: nil,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Equal(t, apikey.DefaultAPIKeyType, got[attrType])
			},
		},
		{
			name: "type_empty_string_defaults_to_rest",
			patch: map[string]any{
				attrType: "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Equal(t, apikey.DefaultAPIKeyType, got[attrType])
			},
		},
		{
			name: "type_cross_cluster_preserved",
			patch: map[string]any{
				attrType: apikey.CrossClusterAPIKeyType,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Equal(t, apikey.CrossClusterAPIKeyType, got[attrType])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseAPIKeyState()
			maps.Copy(raw, tc.patch)

			got := requireUpgradedJSON(t, runUpgrade(t, 1, raw))
			tc.assert(t, got)
		})
	}
}
