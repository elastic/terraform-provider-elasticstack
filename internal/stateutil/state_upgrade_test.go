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

package stateutil_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalStateMap_nil_raw_state(t *testing.T) {
	t.Parallel()
	req := resource.UpgradeStateRequest{RawState: nil}
	resp := &resource.UpgradeStateResponse{}
	m := stateutil.UnmarshalStateMap(req, resp)
	require.Nil(t, m)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Summary(), "Invalid raw state")
}

func TestUnmarshalStateMap_nil_json(t *testing.T) {
	t.Parallel()
	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: nil}}
	resp := &resource.UpgradeStateResponse{}
	m := stateutil.UnmarshalStateMap(req, resp)
	require.Nil(t, m)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Summary(), "Invalid raw state")
}

func TestUnmarshalStateMap_invalid_json(t *testing.T) {
	t.Parallel()
	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: []byte("not-json")}}
	resp := &resource.UpgradeStateResponse{}
	m := stateutil.UnmarshalStateMap(req, resp)
	require.Nil(t, m)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Summary(), "State upgrade error")
}

func TestUnmarshalStateMap_success(t *testing.T) {
	t.Parallel()
	data := map[string]any{"key": "value", "num": float64(42)}
	raw, err := json.Marshal(data)
	require.NoError(t, err)
	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: raw}}
	resp := &resource.UpgradeStateResponse{}
	m := stateutil.UnmarshalStateMap(req, resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "value", m["key"])
	require.InEpsilon(t, float64(42), m["num"], 0.0001)
}

func TestMarshalStateMap_success(t *testing.T) {
	t.Parallel()
	m := map[string]any{"foo": "bar"}
	resp := &resource.UpgradeStateResponse{}
	stateutil.MarshalStateMap(m, resp)
	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.DynamicValue)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	require.Equal(t, "bar", got["foo"])
}

func TestMarshalStateMap_error(t *testing.T) {
	t.Parallel()
	m := map[string]any{"bad": func() {}}
	resp := &resource.UpgradeStateResponse{}
	stateutil.MarshalStateMap(m, resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Summary(), "State upgrade error")
	require.Nil(t, resp.DynamicValue)
}
