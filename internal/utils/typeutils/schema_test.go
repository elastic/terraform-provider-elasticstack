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
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestExpandStringSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  *schema.Set
		want []string
	}{
		{
			name: "returns empty",
			set:  schema.NewSet(schema.HashString, []any{}),
			want: nil,
		},
		{
			name: "converts to string array",
			set:  schema.NewSet(schema.HashString, []any{"a", "b", "c"}),
			want: []string{"c", "b", "a"}, // reordered by hash
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeutils.ExpandStringSet(tt.set); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExpandStringSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
