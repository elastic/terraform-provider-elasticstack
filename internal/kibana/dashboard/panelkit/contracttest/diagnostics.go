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

package contracttest

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func summarizeDiags(diags diag.Diagnostics) string {
	if diags == nil {
		return ""
	}
	var b strings.Builder
	for _, d := range diags {
		if d.Severity() == diag.SeverityError || d.Severity() == diag.SeverityWarning {
			b.WriteString(d.Severity().String())
			b.WriteString(": ")
			b.WriteString(d.Summary())
			if dt := d.Detail(); dt != "" {
				b.WriteString(" — ")
				b.WriteString(dt)
			}
			b.WriteString("\n")
		}
	}
	s := strings.TrimSuffix(b.String(), "\n")
	if s == "" {
		return "(no diagnostics text)"
	}
	return s
}
