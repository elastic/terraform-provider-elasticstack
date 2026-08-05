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

package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessControlValue_toCreateAPI(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var m *models.AccessControlValue
		apiModel := accessControlValueToCreateAPI(m)
		assert.Nil(t, apiModel.AccessMode)
	})

	t.Run("empty values", func(t *testing.T) {
		m := &models.AccessControlValue{
			AccessMode: types.StringNull(),
		}
		apiModel := accessControlValueToCreateAPI(m)
		assert.Nil(t, apiModel.AccessMode)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &models.AccessControlValue{
			AccessMode: types.StringValue("write_restricted"),
		}
		apiModel := accessControlValueToCreateAPI(m)
		require.NotNil(t, apiModel.AccessMode)
		mode, err := apiModel.AccessMode.AsKibanaHTTPAPIsKbnDashboardAccessControlAccessMode0()
		require.NoError(t, err)
		assert.Equal(t, kbapi.WriteRestricted, mode)
	})

	t.Run("default uses its union branch", func(t *testing.T) {
		m := &models.AccessControlValue{
			AccessMode: types.StringValue("default"),
		}
		apiModel := accessControlValueToCreateAPI(m)
		require.NotNil(t, apiModel.AccessMode)
		mode, err := apiModel.AccessMode.AsKibanaHTTPAPIsKbnDashboardAccessControlAccessMode1()
		require.NoError(t, err)
		assert.Equal(t, kbapi.KibanaHTTPAPIsKbnDashboardAccessControlAccessMode1Default, mode)
	})
}

func TestNewAccessControlFromAPI(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		val := newAccessControlFromAPI(nil)
		assert.Nil(t, val)
	})

	t.Run("filled input", func(t *testing.T) {
		accessMode := "write_restricted"
		val := newAccessControlFromAPI(&accessMode)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("write_restricted"), val.AccessMode)
	})
}

func TestDashboardModel_populateFromAPI_mapsAccessControlUnion(t *testing.T) {
	tests := []struct {
		name string
		set  func(*kbapi.KibanaHTTPAPIsKbnDashboardAccessControl_AccessMode) error
		want string
	}{
		{
			name: "write restricted",
			set: func(mode *kbapi.KibanaHTTPAPIsKbnDashboardAccessControl_AccessMode) error {
				return mode.FromKibanaHTTPAPIsKbnDashboardAccessControlAccessMode0(kbapi.WriteRestricted)
			},
			want: "write_restricted",
		},
		{
			name: "default",
			set: func(mode *kbapi.KibanaHTTPAPIsKbnDashboardAccessControl_AccessMode) error {
				return mode.FromKibanaHTTPAPIsKbnDashboardAccessControlAccessMode1(kbapi.KibanaHTTPAPIsKbnDashboardAccessControlAccessMode1Default)
			},
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var accessMode kbapi.KibanaHTTPAPIsKbnDashboardAccessControl_AccessMode
			require.NoError(t, tt.set(&accessMode))

			model := &models.DashboardModel{}
			resp := &kbapi.GetDashboardsIdResponse{
				JSON200: &struct {
					Data     kbapi.KibanaHTTPAPIsKbnDashboardData                   `json:"data"`
					Id       string                                                 `json:"id"` //nolint:revive // var-naming: API struct field
					Meta     kbapi.KibanaHTTPAPIsKbnAsCodeMeta                      `json:"meta"`
					Warnings *[]kbapi.KibanaHTTPAPIsKbnDashboardDroppedPanelWarning `json:"warnings,omitempty"`
				}{
					Data: kbapi.KibanaHTTPAPIsKbnDashboardData{
						Title:         "test dashboard",
						AccessControl: &kbapi.KibanaHTTPAPIsKbnDashboardAccessControl{AccessMode: &accessMode},
					},
					Id: "dashboard-id",
				},
			}

			diags := dashboardPopulateFromAPI(context.Background(), model, resp, "dashboard-id", "default")
			require.False(t, diags.HasError())
			require.NotNil(t, model.AccessControl)
			assert.Equal(t, types.StringValue(tt.want), model.AccessControl.AccessMode)
		})
	}
}

func TestDashboardModel_populateFromAPI_clearsAccessControlWhenAccessModeMissing(t *testing.T) {
	model := &models.DashboardModel{
		AccessControl: &models.AccessControlValue{
			AccessMode: types.StringValue("write_restricted"),
		},
	}

	resp := &kbapi.GetDashboardsIdResponse{
		JSON200: &struct {
			Data     kbapi.KibanaHTTPAPIsKbnDashboardData                   `json:"data"`
			Id       string                                                 `json:"id"` //nolint:revive // var-naming: API struct field
			Meta     kbapi.KibanaHTTPAPIsKbnAsCodeMeta                      `json:"meta"`
			Warnings *[]kbapi.KibanaHTTPAPIsKbnDashboardDroppedPanelWarning `json:"warnings,omitempty"`
		}{
			Data: kbapi.KibanaHTTPAPIsKbnDashboardData{
				Title: "test dashboard",
				Query: &kbapi.KibanaHTTPAPIsKbnAsCodeQuery{},
				TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
					From: "now-15m",
					To:   "now",
				},
				RefreshInterval: &kbapi.KibanaHTTPAPIsKbnDataServiceServerRefreshIntervalSchema{
					Pause: true,
					Value: 0,
				},
			},
			Id: "dashboard-id",
		},
	}

	diags := dashboardPopulateFromAPI(context.Background(), model, resp, "dashboard-id", "default")
	assert.False(t, diags.HasError())
	assert.Nil(t, model.AccessControl)
}
