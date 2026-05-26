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

package securitylistdatastreams

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModel_GetResourceID(t *testing.T) {
	t.Parallel()

	t.Run("prefers space_id when known", func(t *testing.T) {
		t.Parallel()
		m := Model{
			ID:      types.StringValue("default"),
			SpaceID: types.StringValue("custom"),
		}
		require.Equal(t, "custom", m.GetResourceID().ValueString())
	})

	t.Run("falls back to id on import", func(t *testing.T) {
		t.Parallel()
		m := Model{ID: types.StringValue("default")}
		require.Equal(t, "default", m.GetResourceID().ValueString())
	})
}

func TestModel_GetSpaceID(t *testing.T) {
	t.Parallel()

	t.Run("prefers space_id when known", func(t *testing.T) {
		t.Parallel()
		m := Model{
			ID:      types.StringValue("default"),
			SpaceID: types.StringValue("custom"),
		}
		require.Equal(t, "custom", m.GetSpaceID().ValueString())
	})

	t.Run("falls back to id on import", func(t *testing.T) {
		t.Parallel()
		m := Model{ID: types.StringValue("default")}
		require.Equal(t, "default", m.GetSpaceID().ValueString())
	})
}
