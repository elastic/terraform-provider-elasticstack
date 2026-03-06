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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestDurationType_ValueType(t *testing.T) {
	require.Equal(t, Duration{}, DurationType{}.ValueType(context.Background()))
}

func TestDurationType_ValueFromString(t *testing.T) {
	stringValue := basetypes.NewStringValue("duration")
	expectedResult := Duration{StringValue: stringValue}
	durationValue, diags := DurationType{}.ValueFromString(context.Background(), stringValue)

	require.Nil(t, diags)
	require.Equal(t, expectedResult, durationValue)
}

func TestDurationType_ValueFromTerraform(t *testing.T) {
	tests := []struct {
		name          string
		tfValue       tftypes.Value
		expectedValue attr.Value
		expectedError string
	}{
		{
			name:          "should return an error if the tf value is not a string",
			tfValue:       tftypes.NewValue(tftypes.Bool, true),
			expectedValue: nil,
			expectedError: "expected string",
		},
		{
			name:          "should return a new duration value if the tf value is a string",
			tfValue:       tftypes.NewValue(tftypes.String, "3h"),
			expectedValue: NewDurationValue("3h"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := DurationType{}.ValueFromTerraform(context.Background(), tt.tfValue)

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedValue, val)
		})
	}
}
