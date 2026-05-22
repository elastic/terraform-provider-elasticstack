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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func assertCloseStatePanic(t *testing.T, fn func(), wantSubstrings ...string) {
	t.Helper()

	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "expected panic")
		msg, ok := recovered.(string)
		require.True(t, ok, "panic message must be string, got %T: %v", recovered, recovered)
		require.True(t, strings.HasPrefix(msg, "entitycore: ephemeral close state "))
		require.Contains(t, msg, "Close state must be plain Go types only")
		for _, sub := range wantSubstrings {
			require.Contains(t, msg, sub)
		}
	}()

	fn()
}
