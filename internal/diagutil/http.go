package diagutil

import (
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

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

func CheckErrorFromFW(res *esapi.Response, errMsg string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if res.IsError() {
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
