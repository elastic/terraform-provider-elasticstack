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

package typeutils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPreserveRedactedSecret(t *testing.T) {
	t.Parallel()

	t.Run("non-empty api value wins over fallback", func(t *testing.T) {
		t.Parallel()
		apiValue := "api-secret"
		result := typeutils.PreserveRedactedSecret(&apiValue, types.StringValue("configured-secret"))
		require.Equal(t, types.StringValue("api-secret"), result)
	})

	t.Run("nil api value falls back to configured value", func(t *testing.T) {
		t.Parallel()
		result := typeutils.PreserveRedactedSecret(nil, types.StringValue("configured-secret"))
		require.Equal(t, types.StringValue("configured-secret"), result)
	})

	t.Run("empty-string api value is treated as redacted and falls back", func(t *testing.T) {
		t.Parallel()
		apiValue := ""
		result := typeutils.PreserveRedactedSecret(&apiValue, types.StringValue("configured-secret"))
		require.Equal(t, types.StringValue("configured-secret"), result)
	})

	t.Run("nil api value with unknown fallback returns null", func(t *testing.T) {
		t.Parallel()
		result := typeutils.PreserveRedactedSecret(nil, types.StringUnknown())
		require.Equal(t, types.StringNull(), result)
	})

	t.Run("nil api value with null fallback returns null", func(t *testing.T) {
		t.Parallel()
		result := typeutils.PreserveRedactedSecret(nil, types.StringNull())
		require.Equal(t, types.StringNull(), result)
	})
}
