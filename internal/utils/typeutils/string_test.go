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

type testStringEnum string

func TestTrimmedStringishValue(t *testing.T) {
	t.Parallel()

	t.Run("empty value returns null", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, types.StringNull(), typeutils.TrimmedStringishValue(""))
	})

	t.Run("blank value returns null", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, types.StringNull(), typeutils.TrimmedStringishValue("   "))
	})

	t.Run("non-blank value returns value unchanged", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, types.StringValue("hello"), typeutils.TrimmedStringishValue("hello"))
	})

	t.Run("string-like type supported", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, types.StringValue("v1"), typeutils.TrimmedStringishValue(testStringEnum("v1")))
		require.Equal(t, types.StringNull(), typeutils.TrimmedStringishValue(testStringEnum("")))
	})
}

func TestTrimmedStringishPointerValue(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer returns null", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, types.StringNull(), typeutils.TrimmedStringishPointerValue[string](nil))
	})

	t.Run("pointer to blank value returns null", func(t *testing.T) {
		t.Parallel()
		v := "   "
		require.Equal(t, types.StringNull(), typeutils.TrimmedStringishPointerValue(&v))
	})

	t.Run("pointer to non-blank value returns value unchanged", func(t *testing.T) {
		t.Parallel()
		v := "hello"
		require.Equal(t, types.StringValue("hello"), typeutils.TrimmedStringishPointerValue(&v))
	})
}
