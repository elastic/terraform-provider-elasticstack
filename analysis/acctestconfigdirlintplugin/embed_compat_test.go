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

package acctestconfigdirlint

import "testing"

func TestIsTestdataTFEmbedPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"testdata/main.tf", true},
		{"testdata/case/main.tf", true},
		{"testdata/case/compat.tf", true},
		{"testdata/a/b/c/main.tf", true},
		{"testdata/a/b/c/compatibility.tf", true},
		{"testdata/../evil/main.tf", false},
		{"testdata/a/../b/main.tf", false},
		{"testdata//x/main.tf", false},
		{"testdata/./x/main.tf", false},
		{"../testdata/x/main.tf", false},
		{"testdata/x/main.tf.extra", false},
		{"other/testdata/x/main.tf", false},
		{"testdata/x/notmain.tf", true},
		{"testdata/x/not_tf.txt", false},
		{`testdata\x\main.tf`, false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isTestdataTFEmbedPath(tt.path); got != tt.want {
			t.Errorf("isTestdataTFEmbedPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
