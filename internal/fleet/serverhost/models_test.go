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

package serverhost

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAPICreateModel_HostID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  types.String
		wantID *string
	}{
		{
			name:  "null host_id",
			value: types.StringNull(),
		},
		{
			name:  "unknown host_id",
			value: types.StringUnknown(),
		},
		{
			name:   "explicit host_id",
			value:  types.StringValue("my-host-id"),
			wantID: new("my-host-id"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, diags := serverHostModel{
				HostID: tc.value,
				Name:   types.StringValue("test-host"),
				Hosts: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("https://fleet-server:8220"),
				}),
			}.toAPICreateModel(t.Context())

			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantID, body.Id)
		})
	}
}

func TestToAPICreateModel_ProxyID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantProxy *string
	}{
		{name: "null proxy_id", value: types.StringNull()},
		{name: "explicit proxy_id", value: types.StringValue("my-proxy"), wantProxy: new("my-proxy")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, diags := serverHostModel{
				Name:    types.StringValue("test-host"),
				ProxyID: tc.value,
				Hosts: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("https://fleet-server:8220"),
				}),
			}.toAPICreateModel(t.Context())

			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantProxy, body.ProxyId)
		})
	}
}

func TestToAPIUpdateModel_ProxyID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plan      types.String
		prior     types.String
		wantProxy *string
	}{
		{name: "set to set unchanged", plan: types.StringValue("proxy-a"), prior: types.StringValue("proxy-a"), wantProxy: new("proxy-a")},
		{name: "set to different value", plan: types.StringValue("proxy-b"), prior: types.StringValue("proxy-a"), wantProxy: new("proxy-b")},
		{name: "unset to set", plan: types.StringValue("proxy-a"), prior: types.StringNull(), wantProxy: new("proxy-a")},
		{name: "set to null clears with empty string", plan: types.StringNull(), prior: types.StringValue("proxy-a"), wantProxy: new("")},
		{name: "set to empty clears with empty string", plan: types.StringValue(""), prior: types.StringValue("proxy-a"), wantProxy: new("")},
		{name: "never set stays omitted", plan: types.StringNull(), prior: types.StringNull()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hosts := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://fleet-server:8220")})
			body, diags := serverHostModel{
				Name:    types.StringValue("test-host"),
				Hosts:   hosts,
				ProxyID: tc.plan,
			}.toAPIUpdateModel(t.Context(), serverHostModel{ProxyID: tc.prior})

			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantProxy, body.ProxyId)
		})
	}
}

func TestPopulateFromAPI_ProxyID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		api  *string
		want types.String
	}{
		{name: "nil maps to null", api: nil, want: types.StringNull()},
		{name: "empty maps to null", api: new(""), want: types.StringNull()},
		{name: "value preserved", api: new("my-proxy"), want: types.StringValue("my-proxy")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := serverHostModel{SpaceIDs: types.SetNull(types.StringType)}
			diags := m.populateFromAPI(t.Context(), &kbapi.ServerHost{
				Id:       "host-1",
				Name:     "host",
				HostUrls: []string{"https://fleet-server:8220"},
				ProxyId:  tc.api,
			})

			require.False(t, diags.HasError())
			assert.Equal(t, tc.want, m.ProxyID)
		})
	}
}
