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

package securityexceptionlist

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestExceptionListModel_GetResourceID(t *testing.T) {
	t.Parallel()

	t.Run("parses auto-gen id from composite state id", func(t *testing.T) {
		t.Parallel()
		m := ExceptionListModel{ID: types.StringValue("default/auto-gen-uuid")}
		require.Equal(t, "auto-gen-uuid", m.GetResourceID().ValueString())
	})

	t.Run("returns empty when id is not composite", func(t *testing.T) {
		t.Parallel()
		m := ExceptionListModel{ID: types.StringNull()}
		require.Empty(t, m.GetResourceID().ValueString())
	})
}

func TestExceptionListModel_GetSpaceID(t *testing.T) {
	t.Parallel()

	m := ExceptionListModel{SpaceID: types.StringValue("custom")}
	require.Equal(t, "custom", m.GetSpaceID().ValueString())
}
