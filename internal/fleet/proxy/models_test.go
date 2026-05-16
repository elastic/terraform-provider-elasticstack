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

package proxy

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateFromAPI_ProxyHeaders_StringBoolNumber(t *testing.T) {
	headers := map[string]kbapi.FleetProxyHeaderValue{}

	var sv kbapi.FleetProxyHeaderValue
	require.NoError(t, sv.FromFleetProxyHeaderValueString("hello"))
	headers["string-header"] = sv

	var bv kbapi.FleetProxyHeaderValue
	require.NoError(t, bv.FromFleetProxyHeaderValueBoolean(true))
	headers["bool-header"] = bv

	var nv kbapi.FleetProxyHeaderValue
	require.NoError(t, nv.FromFleetProxyHeaderValueNumber(42))
	headers["number-header"] = nv

	item := kbapi.FleetProxyItem{
		Id:           "proxy-1",
		Name:         "test",
		Url:          "https://proxy.example.com",
		ProxyHeaders: &headers,
	}

	var model proxyModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, "proxy-1", model.ProxyID.ValueString())
	assert.Equal(t, "default", model.SpaceID.ValueString())

	require.False(t, model.ProxyHeaders.IsNull())
	elems := model.ProxyHeaders.Elements()
	require.Len(t, elems, 3)
	assert.Equal(t, "hello", elems["string-header"].(types.String).ValueString())
	assert.Equal(t, "true", elems["bool-header"].(types.String).ValueString())
	assert.Equal(t, "42", elems["number-header"].(types.String).ValueString())
}

func TestPopulateFromAPI_ProxyHeaders_NilMap(t *testing.T) {
	item := kbapi.FleetProxyItem{
		Id:           "proxy-1",
		Name:         "test",
		Url:          "https://proxy.example.com",
		ProxyHeaders: nil,
	}

	var model proxyModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError())
	assert.True(t, model.ProxyHeaders.IsNull())
}

func TestProxyHeadersFromModel_EncodesStrings(t *testing.T) {
	in, mapDiags := types.MapValue(types.StringType, map[string]attr.Value{
		"X-Custom":  types.StringValue("my-value"),
		"X-Another": types.StringValue("another"),
	})
	require.False(t, mapDiags.HasError())

	out, diags := proxyHeadersFromModel(in)
	require.False(t, diags.HasError())
	require.NotNil(t, out)
	require.Len(t, *out, 2)

	for k, v := range *out {
		raw, err := json.Marshal(v)
		require.NoError(t, err, "header %q failed to marshal", k)

		var s string
		require.NoError(t, json.Unmarshal(raw, &s), "header %q did not encode as a JSON string", k)

		switch k {
		case "X-Custom":
			assert.Equal(t, "my-value", s)
		case "X-Another":
			assert.Equal(t, "another", s)
		default:
			t.Fatalf("unexpected key %q", k)
		}
	}
}

func TestProxyHeadersFromModel_NullOrEmpty(t *testing.T) {
	out, diags := proxyHeadersFromModel(types.MapNull(types.StringType))
	require.False(t, diags.HasError())
	assert.Nil(t, out)

	empty, mapDiags := types.MapValue(types.StringType, map[string]attr.Value{})
	require.False(t, mapDiags.HasError())
	out, diags = proxyHeadersFromModel(empty)
	require.False(t, diags.HasError())
	assert.Nil(t, out)
}
