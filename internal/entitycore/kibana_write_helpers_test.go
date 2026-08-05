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

package entitycore

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestRequireNonNilKibanaWriteResponse(t *testing.T) {
	t.Run("non-nil response reports false and appends no diagnostics", func(t *testing.T) {
		var diags diag.Diagnostics
		resp := "created"

		nilResp := RequireNonNilKibanaWriteResponse(&diags, &resp, "create", "security list")

		require.False(t, nilResp)
		require.False(t, diags.HasError())
	})

	t.Run("nil response reports true and appends the standard error", func(t *testing.T) {
		var diags diag.Diagnostics
		var resp *string

		nilResp := RequireNonNilKibanaWriteResponse(&diags, resp, "update", "exception item")

		require.True(t, nilResp)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Failed to update exception item", diags[0].Summary())
		require.Equal(t, "API returned empty response", diags[0].Detail())
	})
}

func TestKibanaResourceID(t *testing.T) {
	got := KibanaResourceID("default", "abc-123")
	require.Equal(t, types.StringValue("default/abc-123"), got)
}
