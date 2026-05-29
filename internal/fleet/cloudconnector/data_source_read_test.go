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

package cloudconnector

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestListParamsFromModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        cloudConnectorsDataSourceModel
		wantKuery     *string
		wantPage      *string
		wantPerPage   *string
		wantAllNilMsg string
	}{
		{
			name: "default config sends no query params",
			config: cloudConnectorsDataSourceModel{
				Kuery:   types.StringNull(),
				Page:    types.Int64Null(),
				PerPage: types.Int64Null(),
			},
			wantAllNilMsg: "Default Inputs scenario: no kuery / page / perPage sent to the API",
		},
		{
			name: "kuery passed verbatim",
			config: cloudConnectorsDataSourceModel{
				Kuery:   types.StringValue(`cloud_provider:aws`),
				Page:    types.Int64Null(),
				PerPage: types.Int64Null(),
			},
			wantKuery: new(`cloud_provider:aws`),
		},
		{
			name: "empty kuery value is omitted",
			config: cloudConnectorsDataSourceModel{
				Kuery: types.StringValue(""),
			},
		},
		{
			name: "pagination passed as stringified integers",
			config: cloudConnectorsDataSourceModel{
				Kuery:   types.StringNull(),
				Page:    types.Int64Value(2),
				PerPage: types.Int64Value(50),
			},
			wantPage:    new("2"),
			wantPerPage: new("50"),
		},
		{
			name: "kuery + pagination combined",
			config: cloudConnectorsDataSourceModel{
				Kuery:   types.StringValue(`cloud_provider:azure`),
				Page:    types.Int64Value(1),
				PerPage: types.Int64Value(10),
			},
			wantKuery:   new(`cloud_provider:azure`),
			wantPage:    new("1"),
			wantPerPage: new("10"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := listParamsFromModel(tt.config)

			if tt.wantAllNilMsg != "" {
				assert.Nil(t, got.Kuery, tt.wantAllNilMsg)
				assert.Nil(t, got.Page, tt.wantAllNilMsg)
				assert.Nil(t, got.PerPage, tt.wantAllNilMsg)
				return
			}

			if tt.wantKuery == nil {
				assert.Nil(t, got.Kuery)
			} else if assert.NotNil(t, got.Kuery) {
				assert.Equal(t, *tt.wantKuery, *got.Kuery)
			}

			if tt.wantPage == nil {
				assert.Nil(t, got.Page)
			} else if assert.NotNil(t, got.Page) {
				assert.Equal(t, *tt.wantPage, *got.Page)
			}

			if tt.wantPerPage == nil {
				assert.Nil(t, got.PerPage)
			} else if assert.NotNil(t, got.PerPage) {
				assert.Equal(t, *tt.wantPerPage, *got.PerPage)
			}
		})
	}
}
