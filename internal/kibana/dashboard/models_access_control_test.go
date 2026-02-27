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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessControlValue_toCreateAPI(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var m *AccessControlValue
		apiModel := m.toCreateAPI()
		assert.Nil(t, apiModel)
	})

	t.Run("empty values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringNull(),
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("private"),
			Owner:      types.StringValue("user123"),
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		mode := kbapi.PostDashboardsJSONBodyAccessControlAccessMode("private")
		owner := "user123"
		assert.Equal(t, &mode, apiModel.AccessMode)
		assert.Equal(t, &owner, apiModel.Owner)
	})

	t.Run("partial values - access_mode", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("private"),
			Owner:      types.StringNull(),
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		mode := kbapi.PostDashboardsJSONBodyAccessControlAccessMode("private")
		assert.Equal(t, &mode, apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})

	t.Run("partial values - owner", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringValue("user123"),
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		user123 := "user123"
		assert.Equal(t, &user123, apiModel.Owner)
	})
}

func TestAccessControlValue_toUpdateAPI(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var m *AccessControlValue
		apiModel := m.toUpdateAPI()
		assert.Nil(t, apiModel)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("public"),
			Owner:      types.StringValue("admin"),
		}
		apiModel := m.toUpdateAPI()
		assert.NotNil(t, apiModel)
		mode := kbapi.PutDashboardsIdJSONBodyAccessControlAccessMode("public")
		admin := "admin"
		assert.Equal(t, &mode, apiModel.AccessMode)
		assert.Equal(t, &admin, apiModel.Owner)
	})

	t.Run("empty values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringNull(),
		}
		apiModel := m.toUpdateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})
}

func TestNewAccessControlFromAPI(t *testing.T) {
	t.Run("nil inputs", func(t *testing.T) {
		val := newAccessControlFromAPI(nil, nil)
		assert.Nil(t, val)
	})

	t.Run("filled inputs", func(t *testing.T) {
		accessMode := "private"
		owner := "user1"
		val := newAccessControlFromAPI(&accessMode, &owner)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("private"), val.AccessMode)
		assert.Equal(t, types.StringValue("user1"), val.Owner)
	})

	t.Run("partial inputs - access_mode", func(t *testing.T) {
		accessMode := "private"
		val := newAccessControlFromAPI(&accessMode, nil)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("private"), val.AccessMode)
		assert.Equal(t, types.StringNull(), val.Owner)
	})

	t.Run("partial inputs - owner", func(t *testing.T) {
		owner := "user1"
		val := newAccessControlFromAPI(nil, &owner)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringNull(), val.AccessMode)
		assert.Equal(t, types.StringValue("user1"), val.Owner)
	})
}
