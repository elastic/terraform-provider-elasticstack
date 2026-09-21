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

package privatelocation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloat64SemanticallyEqualUnderFloat32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		a, b      float64
		wantEqual bool
	}{
		{name: "NaN vs value", a: math.NaN(), b: 42.42, wantEqual: false},
		{name: "value vs NaN", a: 42.42, b: math.NaN(), wantEqual: false},
		{name: "NaN vs NaN", a: math.NaN(), b: math.NaN(), wantEqual: false},
		{name: "float32 degradation", a: 42.42, b: 42.41999816894531, wantEqual: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantEqual, float64SemanticallyEqualUnderFloat32(tt.a, tt.b))
		})
	}
}
