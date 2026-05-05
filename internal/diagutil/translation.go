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
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func FrameworkDiagsFromSDK(sdkDiags sdkdiag.Diagnostics) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	for _, sdkDiag := range sdkDiags {
		var fwDiag fwdiag.Diagnostic

		if sdkDiag.Severity == sdkdiag.Error {
			fwDiag = fwdiag.NewErrorDiagnostic(sdkDiag.Summary, sdkDiag.Detail)
		} else {
			fwDiag = fwdiag.NewWarningDiagnostic(sdkDiag.Summary, sdkDiag.Detail)
		}

		diags.Append(fwDiag)
	}

	return diags
}

func SDKDiagsFromFramework(fwDiags fwdiag.Diagnostics) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics

	for _, fwDiag := range fwDiags {
		var sdkDiag sdkdiag.Diagnostic

		if fwDiag.Severity() == fwdiag.SeverityError {
			sdkDiag = sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  fwDiag.Summary(),
				Detail:   fwDiag.Detail(),
			}
		} else {
			sdkDiag = sdkdiag.Diagnostic{
				Severity: sdkdiag.Warning,
				Summary:  fwDiag.Summary(),
				Detail:   fwDiag.Detail(),
			}
		}

		diags = append(diags, sdkDiag)
	}

	return diags
}

func FrameworkDiagFromError(err error) fwdiag.Diagnostics {
	if err == nil {
		return nil
	}
	return fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(err.Error(), ""),
	}
}

func SdkDiagsAsError(diags sdkdiag.Diagnostics) error {
	for _, diag := range diags {
		if diag.Severity == sdkdiag.Error {
			return fmt.Errorf("%s: %s", diag.Summary, diag.Detail)
		}
	}
	return nil
}

func FwDiagsAsError(diags fwdiag.Diagnostics) error {
	for _, diag := range diags {
		if diag.Severity() == fwdiag.SeverityError {
			return fmt.Errorf("%s: %s", diag.Summary(), diag.Detail())
		}
	}
	return nil
}
