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

package schemautil_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
)

func TestDiffIndexTemplateSuppress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		old   string
		new   string
		equal bool
	}{
		{
			`{"key1.key2": 2, "index.key2.key1": "3"}`,
			`{"index": {"key1.key2": "2", "key2.key1": "3"}}`,
			true,
		},
		{
			`{"key1": "2", "key2": "3"}`,
			`{"index": {"key1": "2", "key2": "3"}}`,
			true,
		},
		{
			`{"index":{"key1": "2", "key2": "3"}}`,
			`{"index": {"key1": "2", "key2": "3"}}`,
			true,
		},
		{
			`{"key1": "2", "key2": "3"}`,
			`{"index.key1": "2", "index.key2": "3"}`,
			true,
		},
		{
			`{"key1": 1, "key2": 2}`,
			`{"key1": "2", "index.key2": "3"}`,
			false,
		},
	}

	for _, tc := range tests {
		if sup := tfsdkutils.DiffIndexSettingSuppress("", tc.old, tc.new, nil); sup != tc.equal {
			t.Errorf("Failed for test case: %+v", tc)
		}
	}
}
