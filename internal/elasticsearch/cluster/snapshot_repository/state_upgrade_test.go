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

package snapshot_repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func runUpgrade(t *testing.T, raw map[string]any) *resource.UpgradeStateResponse {
	t.Helper()
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	r := newSnapshotRepositoryResource()
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[0]
	require.True(t, ok)

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: rawJSON},
	}
	resp := &resource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), req, resp)
	return resp
}

func baseSnapshotRepositoryState() map[string]any {
	return map[string]any{
		"id":     "cluster-uuid/my-repo",
		"name":   "my-repo",
		"verify": true,
	}
}

func TestSnapshotRepositoryUpgradeState_fs_singleton_list_to_object(t *testing.T) {
	t.Parallel()

	raw := baseSnapshotRepositoryState()
	raw["fs"] = []any{
		map[string]any{
			"location":                   "/tmp",
			"compress":                   true,
			"chunk_size":                 "1gb",
			"max_snapshot_bytes_per_sec": "40mb",
			"max_restore_bytes_per_sec":  "20mb",
			"readonly":                   false,
			"max_number_of_snapshots":    500,
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	fs, ok := got["fs"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "/tmp", fs["location"])
	require.Equal(t, true, fs["compress"])
	require.Equal(t, "1gb", fs["chunk_size"])
}

func TestSnapshotRepositoryUpgradeState_url_singleton_list_to_object(t *testing.T) {
	t.Parallel()

	raw := baseSnapshotRepositoryState()
	raw["url"] = []any{
		map[string]any{
			"url":                     "file:///tmp",
			"compress":                true,
			"http_max_retries":        5,
			"http_socket_timeout":     "50s",
			"max_number_of_snapshots": 500,
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	u, ok := got["url"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "file:///tmp", u["url"])
	require.InEpsilon(t, 5.0, u["http_max_retries"], 0.0001)
}

func TestSnapshotRepositoryUpgradeState_multiple_blocks_error(t *testing.T) {
	t.Parallel()

	raw := baseSnapshotRepositoryState()
	raw["fs"] = []any{
		map[string]any{"location": "/a"},
		map[string]any{"location": "/b"},
	}

	resp := runUpgrade(t, raw)
	require.True(t, resp.Diagnostics.HasError())
}

func TestSnapshotRepositoryUpgradeState_absent_blocks_preserved(t *testing.T) {
	t.Parallel()

	raw := baseSnapshotRepositoryState()
	// Only fs set, others absent
	raw["fs"] = []any{map[string]any{"location": "/tmp"}}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	_, ok := got["url"]
	require.False(t, ok, "url should be absent")
}

func TestSnapshotRepositoryUpgradeState_already_object(t *testing.T) {
	t.Parallel()

	raw := baseSnapshotRepositoryState()
	raw["fs"] = map[string]any{"location": "/tmp"}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	fs, ok := got["fs"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "/tmp", fs["location"])
}
