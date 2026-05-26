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

package diagutil

import (
	"fmt"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func FrameworkDiagFromError(err error) fwdiag.Diagnostics {
	if err == nil {
		return nil
	}
	return fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(err.Error(), ""),
	}
}

// ErrDiag returns a Diagnostics containing a single error with the given summary
// and the error's message as the detail. It is the context-rich counterpart to
// FrameworkDiagFromError for cases where a human-readable summary differs from
// the underlying error text.
func ErrDiag(summary string, err error) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	diags.AddError(summary, err.Error())
	return diags
}

func FwDiagsAsError(diags fwdiag.Diagnostics) error {
	for _, diag := range diags {
		if diag.Severity() == fwdiag.SeverityError {
			return fmt.Errorf("%s: %s", diag.Summary(), diag.Detail())
		}
	}
	return nil
}
