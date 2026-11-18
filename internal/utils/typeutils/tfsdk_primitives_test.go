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

func TestBoolDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.Bool
		def   bool
		want  bool
	}{
		{name: "null", input: types.BoolNull(), def: true, want: true},
		{name: "unknown", input: types.BoolUnknown(), def: false, want: false},
		{name: "known true stays true even with false default", input: types.BoolValue(true), def: false, want: true},
		{name: "known false stays false even with true default", input: types.BoolValue(false), def: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.input
			got := typeutils.BoolDefault(&v, tt.def)
			require.Equal(t, tt.want, got)
			require.Equal(t, types.BoolValue(tt.want), v)
		})
	}
}

func TestStringDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.String
		def   string
		want  string
	}{
		{name: "null", input: types.StringNull(), def: "fallback", want: "fallback"},
		{name: "unknown", input: types.StringUnknown(), def: "fallback", want: "fallback"},
		{name: "known value stays put", input: types.StringValue("value"), def: "fallback", want: "value"},
		{name: "known empty string stays empty", input: types.StringValue(""), def: "fallback", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.input
			got := typeutils.StringDefault(&v, tt.def)
			require.Equal(t, tt.want, got)
			require.Equal(t, types.StringValue(tt.want), v)
		})
	}
}
