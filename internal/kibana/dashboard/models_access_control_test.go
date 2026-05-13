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
		mode := kbapi.KbnDashboardAccessControlAccessMode("write_restricted")
		assert.Equal(t, &mode, apiModel.AccessMode)
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

func TestDashboardModel_populateFromAPI_clearsAccessControlWhenAccessModeMissing(t *testing.T) {
	model := &models.DashboardModel{
		AccessControl: &models.AccessControlValue{
			AccessMode: types.StringValue("write_restricted"),
		},
	}

	resp := &kbapi.GetDashboardsIdResponse{
		JSON200: &struct {
			Data     kbapi.KbnDashboardData                   `json:"data"`
			Id       string                                   `json:"id"` //nolint:revive // var-naming: API struct field
			Meta     kbapi.KbnAsCodeMeta                      `json:"meta"`
			Warnings *[]kbapi.KbnDashboardDroppedPanelWarning `json:"warnings,omitempty"`
		}{
			Data: kbapi.KbnDashboardData{
				Title: "test dashboard",
				Query: kbapi.KbnAsCodeQuery{},
				TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{
					From: "now-15m",
					To:   "now",
				},
				RefreshInterval: kbapi.KbnDataServiceServerRefreshIntervalSchema{
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
