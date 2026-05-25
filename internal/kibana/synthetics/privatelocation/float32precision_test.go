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
	"context"
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestFloat32PrecisionValue_Float64SemanticEquals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		left      Float32PrecisionValue
		right     basetypes.Float64Valuable
		wantEqual bool
		wantError bool
	}{
		{
			name:      "null vs null",
			left:      NewFloat32PrecisionNull(),
			right:     NewFloat32PrecisionNull(),
			wantEqual: true,
		},
		{
			name:      "unknown vs unknown",
			left:      NewFloat32PrecisionUnknown(),
			right:     NewFloat32PrecisionUnknown(),
			wantEqual: true,
		},
		{
			name:      "float32 degradation",
			left:      NewFloat32PrecisionValue(42.42),
			right:     NewFloat32PrecisionValue(42.41999816894531),
			wantEqual: true,
		},
		{
			name:      "different values",
			left:      NewFloat32PrecisionValue(42.42),
			right:     NewFloat32PrecisionValue(42.43),
			wantEqual: false,
		},
		{
			name:      "value vs null",
			left:      NewFloat32PrecisionValue(42.42),
			right:     NewFloat32PrecisionNull(),
			wantEqual: false,
		},
		{
			name:      "mismatched value type",
			left:      NewFloat32PrecisionValue(42.42),
			right:     basetypes.NewFloat64Value(42.42),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotEqual, diags := tt.left.Float64SemanticEquals(ctx, tt.right)
			if tt.wantError {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError())
			require.Equal(t, tt.wantEqual, gotEqual)
		})
	}

	t.Run("NaN vs value", func(t *testing.T) {
		t.Parallel()
		// basetypes cannot represent NaN on Go 1.26+; Float64SemanticEquals delegates
		// the known-value comparison to float64SemanticallyEqualUnderFloat32.
		require.False(t, float64SemanticallyEqualUnderFloat32(math.NaN(), 42.42))
	})
}

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
