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

package tag

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckManagedTag(t *testing.T) {
	t.Parallel()

	t.Run("nil detail is allowed", func(t *testing.T) {
		diags := checkManagedTag(nil)
		assert.False(t, diags.HasError())
	})

	t.Run("nil managed is allowed", func(t *testing.T) {
		diags := checkManagedTag(&kibanaoapi.TagDetail{})
		assert.False(t, diags.HasError())
	})

	t.Run("managed false is allowed", func(t *testing.T) {
		managed := false
		diags := checkManagedTag(&kibanaoapi.TagDetail{Managed: &managed})
		assert.False(t, diags.HasError())
	})

	t.Run("managed true returns expected diagnostic", func(t *testing.T) {
		managed := true
		diags := checkManagedTag(&kibanaoapi.TagDetail{ID: "tag-123", Managed: &managed})
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Managed Kibana tag", diags[0].Summary())
		assert.Contains(t, diags[0].Detail(), "tag-123")
	})
}
