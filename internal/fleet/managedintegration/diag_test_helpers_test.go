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

package managedintegration

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func requireDiagnosticAtPath(t *testing.T, diags diag.Diagnostics, want path.Path, summaryContains string) {
	t.Helper()
	for _, d := range diags.Errors() {
		if summaryContains != "" && !strings.Contains(d.Summary(), summaryContains) {
			continue
		}
		dwp, ok := d.(diag.DiagnosticWithPath)
		if !ok {
			continue
		}
		if dwp.Path().Equal(want) {
			return
		}
	}
	t.Fatalf("expected attribute diagnostic at %v (summary contains %q), got %v", want, summaryContains, diags)
}
