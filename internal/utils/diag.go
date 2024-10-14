package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ConvertSDKDiagnosticsToFramework(sdkDiags sdkdiag.Diagnostics) fwdiag.Diagnostics {
	var fwDiags fwdiag.Diagnostics

	for _, sdkDiag := range sdkDiags {
		if sdkDiag.Severity == sdkdiag.Error {
			fwDiags.AddError(sdkDiag.Summary, sdkDiag.Detail)
		} else {
			fwDiags.AddWarning(sdkDiag.Summary, sdkDiag.Detail)
		}
	}

	return fwDiags
}

func CheckError(res *esapi.Response, errMsg string) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics

	if res.IsError() {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  errMsg,
			Detail:   fmt.Sprintf("Failed with: %s", body),
		})
		return diags
	}
	return diags
}

func CheckHttpError(res *http.Response, errMsg string) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics

	if res.StatusCode >= 400 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  errMsg,
			Detail:   fmt.Sprintf("Failed with: %s", body),
		})
		return diags
	}
	return diags
}

func CheckHttpErrorFromFW(res *http.Response, errMsg string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if res.StatusCode >= 400 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			diags.AddError(errMsg, err.Error())
			return diags
		}
		diags.AddError(errMsg, fmt.Sprintf("Failed with: %s", body))
		return diags
	}
	return diags
}

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

func FrameworkDiagFromError(err error) fwdiag.Diagnostics {
	if err == nil {
		return nil
	}
	return fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(err.Error(), ""),
	}
}
