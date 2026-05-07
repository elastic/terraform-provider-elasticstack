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
	"github.com/stretchr/testify/require"
)

func TestJSONBytesEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    []byte
		want    bool
		wantErr bool
	}{
		{
			name: "identical JSON",
			a:    []byte(`{"a":1,"b":2}`),
			b:    []byte(`{"a":1,"b":2}`),
			want: true,
		},
		{
			name: "semantically equivalent with different key order",
			a:    []byte(`{"a":1,"b":2}`),
			b:    []byte(`{"b":2,"a":1}`),
			want: true,
		},
		{
			name: "different values",
			a:    []byte(`{"a":1}`),
			b:    []byte(`{"a":2}`),
			want: false,
		},
		{
			name:    "invalid JSON in a",
			a:       []byte(`not json`),
			b:       []byte(`{}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON in b",
			a:       []byte(`{}`),
			b:       []byte(`not json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := typeutils.JSONBytesEqual(tt.a, tt.b)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
