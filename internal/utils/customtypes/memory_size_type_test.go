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

package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/require"
)

func TestMemorySizeType_String(t *testing.T) {
	require.Equal(t, "customtypes.MemorySizeType", MemorySizeType{}.String())
}

func TestMemorySizeType_ValueType(t *testing.T) {
	require.Equal(t, MemorySize{}, MemorySizeType{}.ValueType(context.Background()))
}

func TestMemorySizeType_Equal(t *testing.T) {
	tests := []struct {
		name     string
		typ      MemorySizeType
		other    attr.Type
		expected bool
	}{
		{
			name:     "equal to another MemorySizeType",
			typ:      MemorySizeType{},
			other:    MemorySizeType{},
			expected: true,
		},
		{
			name:     "not equal to different type",
			typ:      MemorySizeType{},
			other:    DurationType{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.typ.Equal(tt.other))
		})
	}
}
