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
