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

package ccr

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowInt64ToInt(t *testing.T) {
	t.Parallel()

	got, diags := NarrowInt64ToInt("field", 42)
	require.False(t, diags.HasError())
	assert.Equal(t, 42, got)
}

func TestNarrowInt64ToInt_overflow(t *testing.T) {
	t.Parallel()

	if math.MaxInt == math.MaxInt64 {
		t.Skip("int is 64-bit; overflow against MaxInt is not practical on this platform")
	}

	_, diags := NarrowInt64ToInt("field", math.MaxInt64)
	require.True(t, diags.HasError())
}
