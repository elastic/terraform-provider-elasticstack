package kibana_oapi

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func reportUnknownError(statusCode int, body []byte) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			string(body),
		),
	}
}

func reportUnknownErrorSDK(statusCode int, body []byte) sdkdiags.Diagnostics {
	return sdkdiags.Diagnostics{
		sdkdiags.Diagnostic{
			Severity: sdkdiags.Error,
			Summary:  fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			Detail:   string(body),
		},
	}
}
