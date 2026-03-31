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
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("private"),
		}
		apiModel := m.toCreateAPI()
		assert.NotNil(t, apiModel)
		mode := kbapi.PostDashboardsIdJSONBodyAccessControlAccessMode("private")
		assert.Equal(t, &mode, apiModel.AccessMode)
	})
}

func TestNewAccessControlFromAPI(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		val := newAccessControlFromAPI(nil)
		assert.Nil(t, val)
	})

	t.Run("filled input", func(t *testing.T) {
		accessMode := "private"
		val := newAccessControlFromAPI(&accessMode)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("private"), val.AccessMode)
	})
}
